package loadbalancer

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"

	"MystiSql/pkg/types"
)

var ErrNoAvailableInstances = errors.New("no available instances")

type LoadBalancer interface {
	Select(instances []*types.DatabaseInstance) (*types.DatabaseInstance, error)
	Name() string
	Reset()
}

type RoundRobinLB struct {
	counter uint64
}

func NewRoundRobinLB() *RoundRobinLB {
	return &RoundRobinLB{counter: 0}
}

func (lb *RoundRobinLB) Select(instances []*types.DatabaseInstance) (*types.DatabaseInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoAvailableInstances
	}

	idx := atomic.AddUint64(&lb.counter, 1) - 1
	selectedIdx := int(idx % uint64(len(instances)))

	return instances[selectedIdx], nil
}

func (lb *RoundRobinLB) Name() string {
	return "round-robin"
}

func (lb *RoundRobinLB) Reset() {
	atomic.StoreUint64(&lb.counter, 0)
}

type WeightedLB struct {
	mu      sync.RWMutex
	weights map[string]int
	total   int
}

func NewWeightedLB() *WeightedLB {
	return &WeightedLB{
		weights: make(map[string]int),
		total:   0,
	}
}

func (lb *WeightedLB) Select(instances []*types.DatabaseInstance) (*types.DatabaseInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoAvailableInstances
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.recalculateWeights(instances)

	if lb.total == 0 {
		return instances[rand.Intn(len(instances))], nil
	}

	r := rand.Intn(lb.total)
	for _, instance := range instances {
		weight := lb.weights[instance.Name]
		if weight <= 0 {
			continue
		}
		r -= weight
		if r < 0 {
			return instance, nil
		}
	}

	return instances[0], nil
}

func (lb *WeightedLB) recalculateWeights(instances []*types.DatabaseInstance) {
	lb.weights = make(map[string]int)
	lb.total = 0

	for _, instance := range instances {
		weight := instance.Weight
		if weight <= 0 {
			weight = 1
		}
		lb.weights[instance.Name] = weight
		lb.total += weight
	}
}

func (lb *WeightedLB) Name() string {
	return "weighted"
}

func (lb *WeightedLB) Reset() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.weights = make(map[string]int)
	lb.total = 0
}

func (lb *WeightedLB) SetWeight(instanceName string, weight int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.weights[instanceName] = weight
}

type LeastConnLB struct {
	mu          sync.RWMutex
	connections map[string]int64
}

func NewLeastConnLB() *LeastConnLB {
	return &LeastConnLB{
		connections: make(map[string]int64),
	}
}

func (lb *LeastConnLB) Select(instances []*types.DatabaseInstance) (*types.DatabaseInstance, error) {
	if len(instances) == 0 {
		return nil, ErrNoAvailableInstances
	}

	lb.mu.RLock()
	defer lb.mu.RUnlock()

	minConns := int64(-1)

	var candidates []*types.DatabaseInstance

	for _, instance := range instances {
		conns := lb.connections[instance.Name]
		if minConns == -1 || conns < minConns {
			minConns = conns
			candidates = []*types.DatabaseInstance{instance}
		} else if conns == minConns {
			candidates = append(candidates, instance)
		}
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	return candidates[rand.Intn(len(candidates))], nil
}

func (lb *LeastConnLB) Name() string {
	return "least-conn"
}

func (lb *LeastConnLB) Reset() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.connections = make(map[string]int64)
}

func (lb *LeastConnLB) Increment(instanceName string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.connections[instanceName]++
}

func (lb *LeastConnLB) Decrement(instanceName string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if lb.connections[instanceName] > 0 {
		lb.connections[instanceName]--
	}
}

func (lb *LeastConnLB) GetConnections(instanceName string) int64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	return lb.connections[instanceName]
}

func (lb *LeastConnLB) GetAllConnections() map[string]int64 {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	result := make(map[string]int64)
	for k, v := range lb.connections {
		result[k] = v
	}
	return result
}

type LoadBalancerFactory struct{}

func NewLoadBalancerFactory() *LoadBalancerFactory {
	return &LoadBalancerFactory{}
}

func (f *LoadBalancerFactory) Create(strategy types.ReadStrategy) LoadBalancer {
	switch strategy {
	case types.ReadStrategyRoundRobin:
		return NewRoundRobinLB()
	case types.ReadStrategyWeighted:
		return NewWeightedLB()
	case types.ReadStrategyLeastConn:
		return NewLeastConnLB()
	default:
		return NewRoundRobinLB()
	}
}
