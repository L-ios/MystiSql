package multicluster

import (
	"context"
	"fmt"
	"sync"
	"time"

	"MystiSql/internal/discovery"
	"MystiSql/pkg/types"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type ClusterStatus string

const (
	ClusterStatusHealthy   ClusterStatus = "healthy"
	ClusterStatusUnhealthy ClusterStatus = "unhealthy"
	ClusterStatusUnknown   ClusterStatus = "unknown"
)

type ClusterInfo struct {
	Name          string        `json:"name"`
	Status        ClusterStatus `json:"status"`
	InstanceCount int           `json:"instanceCount"`
	LastChecked   time.Time     `json:"lastChecked"`
	Error         string        `json:"error,omitempty"`
}

type ClusterClient struct {
	Name       string
	Client     kubernetes.Interface
	Discoverer discovery.InstanceDiscoverer
	Status     ClusterStatus
	LastCheck  time.Time
}

type MultiClusterManager struct {
	mu       sync.RWMutex
	clusters map[string]*ClusterClient
	config   *types.MultiClusterConfig
	registry InstanceRegistry
	logger   *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	running  bool
}

type InstanceRegistry interface {
	RegisterInstance(instance *types.DatabaseInstance) error
	UnregisterInstance(name string) error
	GetInstance(name string) (*types.DatabaseInstance, error)
	ListInstances() ([]*types.DatabaseInstance, error)
}

type MultiClusterManagerOption func(*MultiClusterManager)

func WithMultiClusterLogger(logger *zap.Logger) MultiClusterManagerOption {
	return func(m *MultiClusterManager) {
		m.logger = logger
	}
}

func WithMultiClusterRegistry(registry InstanceRegistry) MultiClusterManagerOption {
	return func(m *MultiClusterManager) {
		m.registry = registry
	}
}

func NewMultiClusterManager(config *types.MultiClusterConfig, opts ...MultiClusterManagerOption) *MultiClusterManager {
	ctx, cancel := context.WithCancel(context.Background())

	logger := zap.NewNop()

	m := &MultiClusterManager{
		clusters: make(map[string]*ClusterClient),
		config:   config,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m *MultiClusterManager) Initialize() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, clusterConfig := range m.config.Clusters {
		client, err := m.createClusterClient(clusterConfig)
		if err != nil {
			m.logger.Error("Failed to create cluster client",
				zap.String("cluster", clusterConfig.Name),
				zap.Error(err))
			continue
		}

		m.clusters[clusterConfig.Name] = client
		m.logger.Info("Cluster client created",
			zap.String("cluster", clusterConfig.Name))
	}

	return nil
}

func (m *MultiClusterManager) createClusterClient(config types.ClusterConfig) (*ClusterClient, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = config.Kubeconfig

	configOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: "",
	}

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create rest config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &ClusterClient{
		Name:      config.Name,
		Client:    clientset,
		Status:    ClusterStatusUnknown,
		LastCheck: time.Now(),
	}, nil
}

func (m *MultiClusterManager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil
	}

	m.running = true

	go m.healthCheckLoop()

	m.logger.Info("Multi-cluster manager started",
		zap.Int("cluster_count", len(m.clusters)))

	return nil
}

func (m *MultiClusterManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.cancel()
	m.running = false

	m.logger.Info("Multi-cluster manager stopped")
	return nil
}

func (m *MultiClusterManager) healthCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	m.checkAllClusters()

	for {
		select {
		case <-ticker.C:
			m.checkAllClusters()
		case <-m.ctx.Done():
			return
		}
	}
}

func (m *MultiClusterManager) checkAllClusters() {
	m.mu.RLock()
	clusters := make([]*ClusterClient, 0, len(m.clusters))
	for _, c := range m.clusters {
		clusters = append(clusters, c)
	}
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, cluster := range clusters {
		wg.Add(1)
		go func(c *ClusterClient) {
			defer wg.Done()
			m.checkClusterHealth(c)
		}(cluster)
	}
	wg.Wait()
}

func (m *MultiClusterManager) checkClusterHealth(cluster *ClusterClient) {
	ctx, cancel := context.WithTimeout(m.ctx, 10*time.Second)
	defer cancel()

	_, err := cluster.Client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})

	m.mu.Lock()
	defer m.mu.Unlock()

	cluster.LastCheck = time.Now()

	if err != nil {
		cluster.Status = ClusterStatusUnhealthy
		m.logger.Warn("Cluster health check failed",
			zap.String("cluster", cluster.Name),
			zap.Error(err))
	} else {
		cluster.Status = ClusterStatusHealthy
	}
}

func (m *MultiClusterManager) GetCluster(name string) (*ClusterClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cluster, exists := m.clusters[name]
	if !exists {
		return nil, fmt.Errorf("cluster %s not found", name)
	}

	return cluster, nil
}

func (m *MultiClusterManager) GetAllClusters() map[string]*ClusterClient {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*ClusterClient)
	for k, v := range m.clusters {
		result[k] = v
	}
	return result
}

func (m *MultiClusterManager) GetClusterStatus(name string) (*ClusterInfo, error) {
	cluster, err := m.GetCluster(name)
	if err != nil {
		return nil, err
	}

	return &ClusterInfo{
		Name:        cluster.Name,
		Status:      cluster.Status,
		LastChecked: cluster.LastCheck,
	}, nil
}

func (m *MultiClusterManager) GetAllClusterStatus() map[string]*ClusterInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*ClusterInfo)
	for name, cluster := range m.clusters {
		result[name] = &ClusterInfo{
			Name:        cluster.Name,
			Status:      cluster.Status,
			LastChecked: cluster.LastCheck,
		}
	}
	return result
}

func (m *MultiClusterManager) AddCluster(config types.ClusterConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clusters[config.Name]; exists {
		return fmt.Errorf("cluster %s already exists", config.Name)
	}

	client, err := m.createClusterClient(config)
	if err != nil {
		return fmt.Errorf("failed to create cluster client: %w", err)
	}

	m.clusters[config.Name] = client

	m.logger.Info("Cluster added",
		zap.String("cluster", config.Name))

	return nil
}

func (m *MultiClusterManager) RemoveCluster(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clusters[name]; !exists {
		return fmt.Errorf("cluster %s not found", name)
	}

	delete(m.clusters, name)

	m.logger.Info("Cluster removed",
		zap.String("cluster", name))

	return nil
}

func (m *MultiClusterManager) DiscoverInstances(ctx context.Context, clusterName string) ([]*types.DatabaseInstance, error) {
	cluster, err := m.GetCluster(clusterName)
	if err != nil {
		return nil, err
	}

	if cluster.Discoverer == nil {
		return nil, fmt.Errorf("cluster %s has no discoverer configured", clusterName)
	}

	instances, err := cluster.Discoverer.Discover(ctx)
	if err != nil {
		return nil, fmt.Errorf("discovery failed for cluster %s: %w", clusterName, err)
	}

	for _, instance := range instances {
		instance.Name = types.FormatInstanceName(clusterName, instance.Name)
		if instance.Labels == nil {
			instance.Labels = make(map[string]string)
		}
		instance.Labels["cluster"] = clusterName
	}

	return instances, nil
}

func (m *MultiClusterManager) DiscoverAllInstances(ctx context.Context) (map[string][]*types.DatabaseInstance, error) {
	m.mu.RLock()
	clusters := make([]*ClusterClient, 0, len(m.clusters))
	for _, c := range m.clusters {
		clusters = append(clusters, c)
	}
	m.mu.RUnlock()

	result := make(map[string][]*types.DatabaseInstance)
	var mu sync.Mutex
	var wg sync.WaitGroup
	var firstError error

	for _, cluster := range clusters {
		wg.Add(1)
		go func(c *ClusterClient) {
			defer wg.Done()

			instances, err := m.DiscoverInstances(ctx, c.Name)
			if err != nil {
				m.logger.Warn("Failed to discover instances",
					zap.String("cluster", c.Name),
					zap.Error(err))
				return
			}

			mu.Lock()
			result[c.Name] = instances
			mu.Unlock()
		}(cluster)
	}

	wg.Wait()

	return result, firstError
}

func (m *MultiClusterManager) RegisterInstancesToRegistry(ctx context.Context) error {
	if m.registry == nil {
		return fmt.Errorf("registry not configured")
	}

	allInstances, err := m.DiscoverAllInstances(ctx)
	if err != nil {
		return err
	}

	for clusterName, instances := range allInstances {
		for _, instance := range instances {
			if err := m.registry.RegisterInstance(instance); err != nil {
				m.logger.Warn("Failed to register instance",
					zap.String("cluster", clusterName),
					zap.String("instance", instance.Name),
					zap.Error(err))
				continue
			}
		}
	}

	return nil
}

func (m *MultiClusterManager) GetStats() MultiClusterStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := MultiClusterStats{
		TotalClusters: len(m.clusters),
	}

	for _, cluster := range m.clusters {
		switch cluster.Status {
		case ClusterStatusHealthy:
			stats.HealthyClusters++
		case ClusterStatusUnhealthy:
			stats.UnhealthyClusters++
		case ClusterStatusUnknown:
			stats.UnknownClusters++
		}
	}

	return stats
}

type MultiClusterStats struct {
	TotalClusters     int `json:"totalClusters"`
	HealthyClusters   int `json:"healthyClusters"`
	UnhealthyClusters int `json:"unhealthyClusters"`
	UnknownClusters   int `json:"unknownClusters"`
}
