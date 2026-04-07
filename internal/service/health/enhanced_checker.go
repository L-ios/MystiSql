package health

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"MystiSql/pkg/types"

	"go.uber.org/zap"
)

type EnhancedHealthChecker struct {
	config    types.HealthCheckConfig
	registry  InstanceRegistry
	checkFunc func(ctx context.Context, instance *types.DatabaseInstance) error
	logger    *zap.Logger

	mu           sync.RWMutex
	healthStatus map[string]*types.InstanceHealth

	ctx     context.Context
	cancel  context.CancelFunc
	running int32

	eventCh chan HealthCheckEvent
}

type HealthCheckEvent struct {
	InstanceName string             `json:"instanceName"`
	OldStatus    types.HealthStatus `json:"oldStatus"`
	NewStatus    types.HealthStatus `json:"newStatus"`
	Error        string             `json:"error,omitempty"`
	Timestamp    time.Time          `json:"timestamp"`
}

type InstanceRegistry interface {
	GetInstance(name string) (*types.DatabaseInstance, error)
	ListInstances() ([]*types.DatabaseInstance, error)
}

func NewEnhancedHealthChecker(config types.HealthCheckConfig, registry InstanceRegistry, checkFunc func(ctx context.Context, instance *types.DatabaseInstance) error, logger *zap.Logger) *EnhancedHealthChecker {
	ctx, cancel := context.WithCancel(context.Background())

	return &EnhancedHealthChecker{
		config:       config,
		registry:     registry,
		checkFunc:    checkFunc,
		logger:       logger,
		healthStatus: make(map[string]*types.InstanceHealth),
		ctx:          ctx,
		cancel:       cancel,
		eventCh:      make(chan HealthCheckEvent, 200),
	}
}

func (h *EnhancedHealthChecker) Start() {
	if !atomic.CompareAndSwapInt32(&h.running, 0, 1) {
		return
	}

	go h.checkLoop()
	h.logger.Info("Enhanced health checker started",
		zap.Duration("interval", h.config.Interval),
		zap.Duration("timeout", h.config.Timeout),
		zap.Int("failure_threshold", h.config.FailureThreshold),
		zap.Int("recovery_threshold", h.config.RecoveryThreshold))
}

func (h *EnhancedHealthChecker) Stop() {
	if !atomic.CompareAndSwapInt32(&h.running, 1, 0) {
		return
	}

	h.cancel()
	close(h.eventCh)
	h.logger.Info("Enhanced health checker stopped")
}

func (h *EnhancedHealthChecker) checkLoop() {
	ticker := time.NewTicker(h.config.Interval)
	defer ticker.Stop()

	h.checkAllInstances()

	for {
		select {
		case <-ticker.C:
			h.checkAllInstances()
		case <-h.ctx.Done():
			return
		}
	}
}

func (h *EnhancedHealthChecker) checkAllInstances() {
	instances, err := h.registry.ListInstances()
	if err != nil {
		h.logger.Error("Failed to list instances for health check", zap.Error(err))
		return
	}

	var wg sync.WaitGroup
	for _, instance := range instances {
		wg.Add(1)
		go func(inst *types.DatabaseInstance) {
			defer wg.Done()
			h.checkSingleInstance(inst)
		}(instance)
	}
	wg.Wait()
}

func (h *EnhancedHealthChecker) checkSingleInstance(instance *types.DatabaseInstance) {
	ctx, cancel := context.WithTimeout(h.ctx, h.config.Timeout)
	defer cancel()

	startTime := time.Now()
	err := h.checkFunc(ctx, instance)
	responseTime := time.Since(startTime)

	h.mu.Lock()
	health, exists := h.healthStatus[instance.Name]
	if !exists {
		health = types.NewInstanceHealth(instance.Name)
		h.healthStatus[instance.Name] = health
	}
	oldStatus := health.Status

	if err != nil {
		health.ConsecutiveFails++
		health.ConsecutiveOKs = 0
		health.LastError = err.Error()
		health.ResponseTime = 0

		if health.ConsecutiveFails >= h.config.FailureThreshold {
			health.Status = types.HealthStatusUnhealthy
		}

		h.logger.Warn("Health check failed",
			zap.String("instance", instance.Name),
			zap.Int("consecutive_fails", health.ConsecutiveFails),
			zap.Error(err))
	} else {
		health.ConsecutiveOKs++
		health.ConsecutiveFails = 0
		health.LastError = ""
		health.ResponseTime = responseTime

		if health.Status == types.HealthStatusUnhealthy {
			if health.ConsecutiveOKs >= h.config.RecoveryThreshold {
				health.Status = types.HealthStatusHealthy
			}
		} else {
			health.Status = types.HealthStatusHealthy
		}

		h.logger.Debug("Health check passed",
			zap.String("instance", instance.Name),
			zap.Duration("response_time", responseTime))
	}

	health.LastCheck = time.Now()
	newStatus := health.Status
	h.mu.Unlock()

	if oldStatus != newStatus {
		h.emitEvent(instance.Name, oldStatus, newStatus, err)
	}
}

func (h *EnhancedHealthChecker) emitEvent(instanceName string, oldStatus, newStatus types.HealthStatus, err error) {
	event := HealthCheckEvent{
		InstanceName: instanceName,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		Timestamp:    time.Now(),
	}

	if err != nil {
		event.Error = err.Error()
	}

	select {
	case h.eventCh <- event:
		h.logger.Info("Health status changed",
			zap.String("instance", instanceName),
			zap.String("old_status", string(oldStatus)),
			zap.String("new_status", string(newStatus)))
	default:
		h.logger.Warn("Health event channel full, dropping event",
			zap.String("instance", instanceName))
	}
}

func (h *EnhancedHealthChecker) GetHealth(instanceName string) (*types.InstanceHealth, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	health, exists := h.healthStatus[instanceName]
	if !exists {
		return nil, fmt.Errorf("instance %s not found in health status", instanceName)
	}

	return health, nil
}

func (h *EnhancedHealthChecker) GetAllHealth() map[string]*types.InstanceHealth {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make(map[string]*types.InstanceHealth)
	for name, health := range h.healthStatus {
		healthCopy := *health
		result[name] = &healthCopy
	}

	return result
}

func (h *EnhancedHealthChecker) GetHealthyInstances() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var healthy []string
	for name, health := range h.healthStatus {
		if health.Status == types.HealthStatusHealthy {
			healthy = append(healthy, name)
		}
	}

	return healthy
}

func (h *EnhancedHealthChecker) GetUnhealthyInstances() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var unhealthy []string
	for name, health := range h.healthStatus {
		if health.Status == types.HealthStatusUnhealthy {
			unhealthy = append(unhealthy, name)
		}
	}

	return unhealthy
}

func (h *EnhancedHealthChecker) GetEventChannel() <-chan HealthCheckEvent {
	return h.eventCh
}

func (h *EnhancedHealthChecker) IsHealthy(instanceName string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	health, exists := h.healthStatus[instanceName]
	if !exists {
		return false
	}

	return health.Status == types.HealthStatusHealthy
}

func (h *EnhancedHealthChecker) ForceCheck(instanceName string) (*types.InstanceHealth, error) {
	instance, err := h.registry.GetInstance(instanceName)
	if err != nil {
		return nil, fmt.Errorf("instance %s not found: %w", instanceName, err)
	}

	h.checkSingleInstance(instance)

	return h.GetHealth(instanceName)
}

func (h *EnhancedHealthChecker) RemoveInstance(instanceName string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.healthStatus, instanceName)
	h.logger.Info("Removed instance from health checker", zap.String("instance", instanceName))
}

func (h *EnhancedHealthChecker) GetStats() HealthCheckStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := HealthCheckStats{
		TotalInstances: len(h.healthStatus),
	}

	for _, health := range h.healthStatus {
		switch health.Status {
		case types.HealthStatusHealthy:
			stats.HealthyInstances++
		case types.HealthStatusUnhealthy:
			stats.UnhealthyInstances++
		case types.HealthStatusChecking:
			stats.CheckingInstances++
		}
	}

	return stats
}

type HealthCheckStats struct {
	TotalInstances     int `json:"totalInstances"`
	HealthyInstances   int `json:"healthyInstances"`
	UnhealthyInstances int `json:"unhealthyInstances"`
	CheckingInstances  int `json:"checkingInstances"`
}
