package k8s

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"MystiSql/internal/discovery"
	"MystiSql/pkg/types"
)

// Discoverer represents a Kubernetes service discoverer
type Discoverer struct {
	client          *Client
	namespaces      []string
	selectors       []Selector
	portMapping     map[string]int
	informerFactory informers.SharedInformerFactory
	stopCh          chan struct{}
}

// Selector represents a label selector for Kubernetes resources
type Selector struct {
	LabelSelector string
	Type          string
}

// NewDiscoverer creates a new Kubernetes discoverer
func NewDiscoverer(client *Client, namespaces []string, selectors []Selector, portMapping map[string]int) *Discoverer {
	return &Discoverer{
		client:      client,
		namespaces:  namespaces,
		selectors:   selectors,
		portMapping: portMapping,
		stopCh:      make(chan struct{}),
	}
}

// Name returns the name of the discoverer
func (d *Discoverer) Name() string {
	return "k8s"
}

// Discover discovers database instances from Kubernetes
func (d *Discoverer) Discover(ctx context.Context) ([]*types.DatabaseInstance, error) {
	var instances []*types.DatabaseInstance

	for _, namespace := range d.namespaces {
		services, err := d.client.Clientset().CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list services in namespace %s: %w", namespace, err)
		}

		for _, service := range services.Items {
			instance := d.convertServiceToInstance(&service)
			if instance != nil {
				instances = append(instances, instance)
			}
		}

		pods, err := d.client.Clientset().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
		}

		for _, pod := range pods.Items {
			instance := d.convertPodToInstance(&pod)
			if instance != nil {
				instances = append(instances, instance)
			}
		}
	}

	return instances, nil
}

// Watch watches for changes in Kubernetes resources
func (d *Discoverer) Watch(ctx context.Context) (<-chan discovery.DiscoveryEvent, error) {
	eventCh := make(chan discovery.DiscoveryEvent)

	// Start informers for services and pods
	d.informerFactory = informers.NewSharedInformerFactory(d.client.Clientset(), 30*time.Second)

	// Watch services
	serviceInformer := d.informerFactory.Core().V1().Services().Informer()
	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			service := obj.(*corev1.Service)
			if instance := d.convertServiceToInstance(service); instance != nil {
				eventCh <- discovery.DiscoveryEvent{
					Type:     discovery.EventTypeAdd,
					Instance: instance,
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newService := newObj.(*corev1.Service)
			newInstance := d.convertServiceToInstance(newService)
			if newInstance != nil {
				eventCh <- discovery.DiscoveryEvent{
					Type:     discovery.EventTypeUpdate,
					Instance: newInstance,
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			service := obj.(*corev1.Service)
			if instance := d.convertServiceToInstance(service); instance != nil {
				eventCh <- discovery.DiscoveryEvent{
					Type:     discovery.EventTypeDelete,
					Instance: instance,
				}
			}
		},
	})

	// Watch pods
	podInformer := d.informerFactory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			if instance := d.convertPodToInstance(pod); instance != nil {
				eventCh <- discovery.DiscoveryEvent{
					Type:     discovery.EventTypeAdd,
					Instance: instance,
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPod := newObj.(*corev1.Pod)
			newInstance := d.convertPodToInstance(newPod)
			if newInstance != nil {
				eventCh <- discovery.DiscoveryEvent{
					Type:     discovery.EventTypeUpdate,
					Instance: newInstance,
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			if instance := d.convertPodToInstance(pod); instance != nil {
				eventCh <- discovery.DiscoveryEvent{
					Type:     discovery.EventTypeDelete,
					Instance: instance,
				}
			}
		},
	})

	// Start informers
	d.informerFactory.Start(d.stopCh)

	// Wait for caches to sync
	if !cache.WaitForCacheSync(d.stopCh, serviceInformer.HasSynced, podInformer.HasSynced) {
		return nil, fmt.Errorf("failed to sync informer caches")
	}

	return eventCh, nil
}

// Stop stops the discoverer
func (d *Discoverer) Stop() error {
	close(d.stopCh)
	return nil
}

// convertServiceToInstance converts a Kubernetes service to a database instance
func (d *Discoverer) convertServiceToInstance(service *corev1.Service) *types.DatabaseInstance {
	for _, selector := range d.selectors {
		labelSelector, err := labels.Parse(selector.LabelSelector)
		if err != nil {
			continue
		}

		if labelSelector.Matches(labels.Set(service.Labels)) {
			port := d.getPort(service, selector.Type)
			if port == 0 {
				return nil
			}

			return &types.DatabaseInstance{
				Name:        service.Name,
				Type:        types.DatabaseType(selector.Type),
				Host:        service.Spec.ClusterIP,
				Port:        port,
				Labels:      service.Labels,
				Annotations: service.Annotations,
				Status:      types.InstanceStatusUnknown,
			}
		}
	}

	return nil
}

// convertPodToInstance converts a Kubernetes pod to a database instance
func (d *Discoverer) convertPodToInstance(pod *corev1.Pod) *types.DatabaseInstance {
	for _, selector := range d.selectors {
		labelSelector, err := labels.Parse(selector.LabelSelector)
		if err != nil {
			continue
		}

		if labelSelector.Matches(labels.Set(pod.Labels)) {
			port := d.getPodPort(pod, selector.Type)
			if port == 0 {
				return nil
			}

			return &types.DatabaseInstance{
				Name:        pod.Name,
				Type:        types.DatabaseType(selector.Type),
				Host:        pod.Status.PodIP,
				Port:        port,
				Labels:      pod.Labels,
				Annotations: pod.Annotations,
				Status:      types.InstanceStatusUnknown,
			}
		}
	}

	return nil
}

// getPort gets the port for a service based on the database type
func (d *Discoverer) getPort(service *corev1.Service, dbType string) int {
	// Check port mapping first
	if port, ok := d.portMapping[dbType]; ok {
		return port
	}

	// Try to find the port in the service
	for _, port := range service.Spec.Ports {
		if port.Name == dbType || port.TargetPort.IntValue() > 0 {
			return int(port.Port)
		}
	}

	// Default ports
	switch dbType {
	case "mysql":
		return 3306
	case "postgresql":
		return 5432
	case "oracle":
		return 1521
	case "redis":
		return 6379
	default:
		return 0
	}
}

// getPodPort gets the port for a pod based on the database type
func (d *Discoverer) getPodPort(pod *corev1.Pod, dbType string) int {
	// Check port mapping first
	if port, ok := d.portMapping[dbType]; ok {
		return port
	}

	// Try to find the port in the pod
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			if port.Name == dbType || port.ContainerPort > 0 {
				return int(port.ContainerPort)
			}
		}
	}

	// Default ports
	switch dbType {
	case "mysql":
		return 3306
	case "postgresql":
		return 5432
	case "oracle":
		return 1521
	case "redis":
		return 6379
	default:
		return 0
	}
}
