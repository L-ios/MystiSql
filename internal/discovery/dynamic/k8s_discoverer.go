package dynamic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"MystiSql/pkg/types"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

type DiscoveryEvent struct {
	Type      DiscoveryEventType      `json:"type"`
	Instance  *types.DatabaseInstance `json:"instance"`
	Timestamp time.Time               `json:"timestamp"`
}

type DiscoveryEventType string

const (
	DiscoveryEventAdd    DiscoveryEventType = "add"
	DiscoveryEventUpdate DiscoveryEventType = "update"
	DiscoveryEventDelete DiscoveryEventType = "delete"
)

type K8sDynamicDiscoverer struct {
	client        kubernetes.Interface
	namespace     string
	labelSelector labels.Selector
	portMapping   map[string]int
	logger        *zap.Logger

	mu        sync.RWMutex
	instances map[string]*types.DatabaseInstance
	eventCh   chan DiscoveryEvent
	ctx       context.Context
	cancel    context.CancelFunc
	running   bool
}

type K8sDynamicDiscovererOption func(*K8sDynamicDiscoverer)

func WithK8sLogger(logger *zap.Logger) K8sDynamicDiscovererOption {
	return func(d *K8sDynamicDiscoverer) {
		d.logger = logger
	}
}

func WithNamespace(namespace string) K8sDynamicDiscovererOption {
	return func(d *K8sDynamicDiscoverer) {
		d.namespace = namespace
	}
}

func WithLabelSelector(selector string) K8sDynamicDiscovererOption {
	return func(d *K8sDynamicDiscoverer) {
		if selector != "" {
			s, err := labels.Parse(selector)
			if err == nil {
				d.labelSelector = s
			}
		}
	}
}

func WithPortMapping(mapping map[string]int) K8sDynamicDiscovererOption {
	return func(d *K8sDynamicDiscoverer) {
		d.portMapping = mapping
	}
}

func NewK8sDynamicDiscoverer(client kubernetes.Interface, opts ...K8sDynamicDiscovererOption) *K8sDynamicDiscoverer {
	ctx, cancel := context.WithCancel(context.Background())

	logger := zap.NewNop()

	d := &K8sDynamicDiscoverer{
		client:        client,
		namespace:     "default",
		labelSelector: labels.Everything(),
		portMapping:   make(map[string]int),
		logger:        logger,
		instances:     make(map[string]*types.DatabaseInstance),
		eventCh:       make(chan DiscoveryEvent, 200),
		ctx:           ctx,
		cancel:        cancel,
	}

	for _, opt := range opts {
		opt(d)
	}

	return d
}

func (d *K8sDynamicDiscoverer) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.running {
		return nil
	}

	d.running = true

	go d.watchLoop()

	d.logger.Info("K8s dynamic discoverer started",
		zap.String("namespace", d.namespace),
		zap.String("selector", d.labelSelector.String()))

	return nil
}

func (d *K8sDynamicDiscoverer) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return nil
	}

	d.cancel()
	d.running = false
	close(d.eventCh)

	d.logger.Info("K8s dynamic discoverer stopped")
	return nil
}

func (d *K8sDynamicDiscoverer) watchLoop() {
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
			if err := d.doWatch(); err != nil {
				d.logger.Error("Watch failed, retrying", zap.Error(err))
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (d *K8sDynamicDiscoverer) doWatch() error {
	watcher, err := d.client.CoreV1().Services(d.namespace).Watch(
		d.ctx,
		metav1.ListOptions{
			LabelSelector: d.labelSelector.String(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to watch services: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return nil
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed")
			}

			d.handleEvent(event)
		}
	}
}

func (d *K8sDynamicDiscoverer) handleEvent(event watch.Event) {
	service, ok := event.Object.(*corev1.Service)
	if !ok {
		return
	}

	switch event.Type {
	case watch.Added:
		d.handleServiceAdded(service)
	case watch.Modified:
		d.handleServiceModified(service)
	case watch.Deleted:
		d.handleServiceDeleted(service)
	}
}

func (d *K8sDynamicDiscoverer) handleServiceAdded(service *corev1.Service) {
	instance := d.serviceToInstance(service)
	if instance == nil {
		return
	}

	d.mu.Lock()
	d.instances[instance.Name] = instance
	d.mu.Unlock()

	d.emitEvent(DiscoveryEventAdd, instance)

	d.logger.Info("Service added",
		zap.String("name", instance.Name),
		zap.String("type", string(instance.Type)))
}

func (d *K8sDynamicDiscoverer) handleServiceModified(service *corev1.Service) {
	instance := d.serviceToInstance(service)
	if instance == nil {
		return
	}

	d.mu.Lock()
	oldInstance, exists := d.instances[instance.Name]
	d.instances[instance.Name] = instance
	d.mu.Unlock()

	if !exists {
		d.emitEvent(DiscoveryEventAdd, instance)
	} else {
		if oldInstance.Host != instance.Host || oldInstance.Port != instance.Port {
			d.emitEvent(DiscoveryEventUpdate, instance)
		}
	}

	d.logger.Debug("Service modified",
		zap.String("name", instance.Name))
}

func (d *K8sDynamicDiscoverer) handleServiceDeleted(service *corev1.Service) {
	instance := d.serviceToInstance(service)
	if instance == nil {
		return
	}

	d.mu.Lock()
	delete(d.instances, instance.Name)
	d.mu.Unlock()

	d.emitEvent(DiscoveryEventDelete, instance)

	d.logger.Info("Service deleted",
		zap.String("name", instance.Name))
}

func (d *K8sDynamicDiscoverer) serviceToInstance(service *corev1.Service) *types.DatabaseInstance {
	dbType := d.getDatabaseType(service)
	if dbType == "" {
		return nil
	}

	port := d.getServicePort(service, dbType)

	instance := types.NewDatabaseInstance(
		service.Name,
		types.DatabaseType(dbType),
		service.Spec.ClusterIP,
		port,
	)

	if instance.Host == "" || instance.Host == "None" {
		instance.Host = fmt.Sprintf("%s.%s.svc.cluster.local", service.Name, service.Namespace)
	}

	if role, ok := service.Labels["role"]; ok {
		instance.Role = role
	} else {
		instance.Role = string(types.InstanceRoleMaster)
	}

	if master, ok := service.Labels["master"]; ok {
		instance.Master = master
		instance.Role = string(types.InstanceRoleSlave)
	}

	if weight, ok := service.Labels["weight"]; ok {
		var w int
		fmt.Sscanf(weight, "%d", &w)
		instance.Weight = w
	}

	instance.Labels = service.Labels
	instance.Annotations = service.Annotations

	return instance
}

func (d *K8sDynamicDiscoverer) getDatabaseType(service *corev1.Service) string {
	if dbType, ok := service.Labels["database-type"]; ok {
		return dbType
	}

	if dbType, ok := service.Labels["app"]; ok {
		if isValidDatabaseType(dbType) {
			return dbType
		}
	}

	for _, port := range service.Spec.Ports {
		switch port.Port {
		case 3306:
			return "mysql"
		case 5432:
			return "postgresql"
		case 1521:
			return "oracle"
		case 6379:
			return "redis"
		case 1433:
			return "mssql"
		case 27017:
			return "mongodb"
		case 9200:
			return "elasticsearch"
		case 9000:
			return "clickhouse"
		case 2379:
			return "etcd"
		}
	}

	return ""
}

func (d *K8sDynamicDiscoverer) getServicePort(service *corev1.Service, dbType string) int {
	if mappedPort, ok := d.portMapping[dbType]; ok {
		return mappedPort
	}

	for _, port := range service.Spec.Ports {
		if port.Port > 0 {
			return int(port.Port)
		}
	}

	return 0
}

func (d *K8sDynamicDiscoverer) emitEvent(eventType DiscoveryEventType, instance *types.DatabaseInstance) {
	event := DiscoveryEvent{
		Type:      eventType,
		Instance:  instance,
		Timestamp: time.Now(),
	}

	select {
	case d.eventCh <- event:
	default:
		d.logger.Warn("Event channel full, dropping event",
			zap.String("type", string(eventType)),
			zap.String("instance", instance.Name))
	}
}

func (d *K8sDynamicDiscoverer) GetEvents() <-chan DiscoveryEvent {
	return d.eventCh
}

func (d *K8sDynamicDiscoverer) GetInstances() []*types.DatabaseInstance {
	d.mu.RLock()
	defer d.mu.RUnlock()

	instances := make([]*types.DatabaseInstance, 0, len(d.instances))
	for _, instance := range d.instances {
		instances = append(instances, instance)
	}
	return instances
}

func (d *K8sDynamicDiscoverer) GetInstance(name string) (*types.DatabaseInstance, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	instance, exists := d.instances[name]
	if !exists {
		return nil, fmt.Errorf("instance %s not found", name)
	}
	return instance, nil
}

func (d *K8sDynamicDiscoverer) Refresh(ctx context.Context) error {
	services, err := d.client.CoreV1().Services(d.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: d.labelSelector.String(),
	})
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.instances = make(map[string]*types.DatabaseInstance)

	for _, service := range services.Items {
		instance := d.serviceToInstance(&service)
		if instance != nil {
			d.instances[instance.Name] = instance
		}
	}

	d.logger.Info("Refreshed instances", zap.Int("count", len(d.instances)))
	return nil
}

func isValidDatabaseType(dbType string) bool {
	validTypes := map[string]bool{
		"mysql":         true,
		"postgresql":    true,
		"oracle":        true,
		"redis":         true,
		"sqlite":        true,
		"mssql":         true,
		"mongodb":       true,
		"elasticsearch": true,
		"clickhouse":    true,
		"etcd":          true,
	}
	return validTypes[dbType]
}
