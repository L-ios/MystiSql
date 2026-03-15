package monitor

import (
	"sync"
	"sync/atomic"
	"time"

	"MystiSql/internal/connection"
)

type Collector struct {
	instanceMetrics map[string]*instanceMetrics
	mu              sync.RWMutex
	eventHandlers   []func(connection.MetricsEvent)
}

type instanceMetrics struct {
	stats connection.PoolStats

	acquireCount       int64
	acquireFailed      int64
	acquireDuration    int64
	maxAcquireDuration int64

	waitCount       int64
	waitDuration    int64
	maxWaitDuration int64

	queryCount    int64
	queryFailed   int64
	queryDuration int64

	execCount    int64
	execFailed   int64
	execDuration int64

	healthCheckCount  int64
	healthCheckFailed int64

	connectionsCreated int64
	connectionsClosed  int64

	lastAcquireTime int64
	lastReleaseTime int64
	lastErrorTime   int64
	lastErrorMsg    string
}

func NewCollector() *Collector {
	return &Collector{
		instanceMetrics: make(map[string]*instanceMetrics),
		eventHandlers:   make([]func(connection.MetricsEvent), 0),
	}
}

func (c *Collector) RegisterEventHandler(handler func(connection.MetricsEvent)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.eventHandlers = append(c.eventHandlers, handler)
}

func (c *Collector) emitEvent(event connection.MetricsEvent) {
	c.mu.RLock()
	handlers := make([]func(connection.MetricsEvent), len(c.eventHandlers))
	copy(handlers, c.eventHandlers)
	c.mu.RUnlock()

	for _, handler := range handlers {
		go handler(event)
	}
}

func (c *Collector) getOrCreateMetrics(instance string) *instanceMetrics {
	c.mu.Lock()
	defer c.mu.Unlock()

	if m, exists := c.instanceMetrics[instance]; exists {
		return m
	}

	m := &instanceMetrics{}
	c.instanceMetrics[instance] = m
	return m
}

func (c *Collector) RecordAcquire(instance string, duration int64, success bool) {
	m := c.getOrCreateMetrics(instance)
	atomic.AddInt64(&m.acquireCount, 1)

	if !success {
		atomic.AddInt64(&m.acquireFailed, 1)
	}

	atomic.AddInt64(&m.acquireDuration, duration)

	for {
		old := atomic.LoadInt64(&m.maxAcquireDuration)
		if duration <= old || atomic.CompareAndSwapInt64(&m.maxAcquireDuration, old, duration) {
			break
		}
	}

	atomic.StoreInt64(&m.lastAcquireTime, time.Now().UnixNano())

	c.emitEvent(connection.MetricsEvent{
		Type:      "acquire",
		Instance:  instance,
		Timestamp: time.Now().UnixNano(),
		Duration:  duration,
		Success:   success,
	})
}

func (c *Collector) RecordRelease(instance string, duration int64) {
	m := c.getOrCreateMetrics(instance)
	atomic.StoreInt64(&m.lastReleaseTime, time.Now().UnixNano())

	c.emitEvent(connection.MetricsEvent{
		Type:      "release",
		Instance:  instance,
		Timestamp: time.Now().UnixNano(),
		Duration:  duration,
		Success:   true,
	})
}

func (c *Collector) RecordWait(instance string, duration int64) {
	m := c.getOrCreateMetrics(instance)
	atomic.AddInt64(&m.waitCount, 1)
	atomic.AddInt64(&m.waitDuration, duration)

	for {
		old := atomic.LoadInt64(&m.maxWaitDuration)
		if duration <= old || atomic.CompareAndSwapInt64(&m.maxWaitDuration, old, duration) {
			break
		}
	}

	c.emitEvent(connection.MetricsEvent{
		Type:      "wait",
		Instance:  instance,
		Timestamp: time.Now().UnixNano(),
		Duration:  duration,
		Success:   true,
	})
}

func (c *Collector) RecordQuery(instance string, duration int64, success bool) {
	m := c.getOrCreateMetrics(instance)
	atomic.AddInt64(&m.queryCount, 1)

	if !success {
		atomic.AddInt64(&m.queryFailed, 1)
	}

	atomic.AddInt64(&m.queryDuration, duration)

	c.emitEvent(connection.MetricsEvent{
		Type:      "query",
		Instance:  instance,
		Timestamp: time.Now().UnixNano(),
		Duration:  duration,
		Success:   success,
	})
}

func (c *Collector) RecordExec(instance string, duration int64, success bool) {
	m := c.getOrCreateMetrics(instance)
	atomic.AddInt64(&m.execCount, 1)

	if !success {
		atomic.AddInt64(&m.execFailed, 1)
	}

	atomic.AddInt64(&m.execDuration, duration)

	c.emitEvent(connection.MetricsEvent{
		Type:      "exec",
		Instance:  instance,
		Timestamp: time.Now().UnixNano(),
		Duration:  duration,
		Success:   success,
	})
}

func (c *Collector) RecordHealthCheck(instance string, success bool) {
	m := c.getOrCreateMetrics(instance)
	atomic.AddInt64(&m.healthCheckCount, 1)

	if !success {
		atomic.AddInt64(&m.healthCheckFailed, 1)
	}

	c.emitEvent(connection.MetricsEvent{
		Type:      "health_check",
		Instance:  instance,
		Timestamp: time.Now().UnixNano(),
		Success:   success,
	})
}

func (c *Collector) RecordConnectionCreated(instance string) {
	m := c.getOrCreateMetrics(instance)
	atomic.AddInt64(&m.connectionsCreated, 1)

	c.emitEvent(connection.MetricsEvent{
		Type:      "connection_created",
		Instance:  instance,
		Timestamp: time.Now().UnixNano(),
		Success:   true,
	})
}

func (c *Collector) RecordConnectionClosed(instance string) {
	m := c.getOrCreateMetrics(instance)
	atomic.AddInt64(&m.connectionsClosed, 1)

	c.emitEvent(connection.MetricsEvent{
		Type:      "connection_closed",
		Instance:  instance,
		Timestamp: time.Now().UnixNano(),
		Success:   true,
	})
}

func (c *Collector) UpdatePoolStats(instance string, stats *connection.PoolStats) {
	m := c.getOrCreateMetrics(instance)

	m.stats.TotalConnections = stats.TotalConnections
	m.stats.IdleConnections = stats.IdleConnections
	m.stats.ActiveConnections = stats.ActiveConnections
	m.stats.MaxConnections = stats.MaxConnections
	m.stats.MinConnections = stats.MinConnections
}

func (c *Collector) RecordError(instance string, errMsg string) {
	m := c.getOrCreateMetrics(instance)
	atomic.StoreInt64(&m.lastErrorTime, time.Now().UnixNano())
	m.lastErrorMsg = errMsg
}

func (c *Collector) GetAllMetrics() map[string]*connection.PoolStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*connection.PoolStats)
	for instance, m := range c.instanceMetrics {
		stats := &connection.PoolStats{
			TotalConnections:   m.stats.TotalConnections,
			IdleConnections:    m.stats.IdleConnections,
			ActiveConnections:  m.stats.ActiveConnections,
			MaxConnections:     m.stats.MaxConnections,
			MinConnections:     m.stats.MinConnections,
			AcquireCount:       atomic.LoadInt64(&m.acquireCount),
			AcquireFailed:      atomic.LoadInt64(&m.acquireFailed),
			ReleaseCount:       atomic.LoadInt64(&m.acquireCount) - atomic.LoadInt64(&m.acquireFailed),
			WaitCount:          atomic.LoadInt64(&m.waitCount),
			WaitDuration:       atomic.LoadInt64(&m.waitDuration),
			MaxWaitDuration:    atomic.LoadInt64(&m.maxWaitDuration),
			AcquireDuration:    atomic.LoadInt64(&m.acquireDuration),
			MaxAcquireDuration: atomic.LoadInt64(&m.maxAcquireDuration),
			QueryCount:         atomic.LoadInt64(&m.queryCount),
			QueryFailed:        atomic.LoadInt64(&m.queryFailed),
			QueryDuration:      atomic.LoadInt64(&m.queryDuration),
			ExecCount:          atomic.LoadInt64(&m.execCount),
			ExecFailed:         atomic.LoadInt64(&m.execFailed),
			ExecDuration:       atomic.LoadInt64(&m.execDuration),
			HealthCheckCount:   atomic.LoadInt64(&m.healthCheckCount),
			HealthCheckFailed:  atomic.LoadInt64(&m.healthCheckFailed),
			ConnectionsCreated: atomic.LoadInt64(&m.connectionsCreated),
			ConnectionsClosed:  atomic.LoadInt64(&m.connectionsClosed),
			LastAcquireTime:    atomic.LoadInt64(&m.lastAcquireTime),
			LastReleaseTime:    atomic.LoadInt64(&m.lastReleaseTime),
			LastErrorTime:      atomic.LoadInt64(&m.lastErrorTime),
			LastErrorMsg:       m.lastErrorMsg,
		}

		if stats.AcquireCount > 0 {
			stats.AvgAcquireDuration = stats.AcquireDuration / stats.AcquireCount
		}
		if stats.WaitCount > 0 {
			stats.AvgWaitDuration = stats.WaitDuration / stats.WaitCount
		}

		result[instance] = stats
	}

	return result
}

func (c *Collector) GetInstanceMetrics(instance string) *connection.PoolStats {
	c.mu.RLock()
	m, exists := c.instanceMetrics[instance]
	c.mu.RUnlock()

	if !exists {
		return nil
	}

	stats := &connection.PoolStats{
		TotalConnections:   m.stats.TotalConnections,
		IdleConnections:    m.stats.IdleConnections,
		ActiveConnections:  m.stats.ActiveConnections,
		MaxConnections:     m.stats.MaxConnections,
		MinConnections:     m.stats.MinConnections,
		AcquireCount:       atomic.LoadInt64(&m.acquireCount),
		AcquireFailed:      atomic.LoadInt64(&m.acquireFailed),
		ReleaseCount:       atomic.LoadInt64(&m.acquireCount) - atomic.LoadInt64(&m.acquireFailed),
		WaitCount:          atomic.LoadInt64(&m.waitCount),
		WaitDuration:       atomic.LoadInt64(&m.waitDuration),
		MaxWaitDuration:    atomic.LoadInt64(&m.maxWaitDuration),
		AcquireDuration:    atomic.LoadInt64(&m.acquireDuration),
		MaxAcquireDuration: atomic.LoadInt64(&m.maxAcquireDuration),
		QueryCount:         atomic.LoadInt64(&m.queryCount),
		QueryFailed:        atomic.LoadInt64(&m.queryFailed),
		QueryDuration:      atomic.LoadInt64(&m.queryDuration),
		ExecCount:          atomic.LoadInt64(&m.execCount),
		ExecFailed:         atomic.LoadInt64(&m.execFailed),
		ExecDuration:       atomic.LoadInt64(&m.execDuration),
		HealthCheckCount:   atomic.LoadInt64(&m.healthCheckCount),
		HealthCheckFailed:  atomic.LoadInt64(&m.healthCheckFailed),
		ConnectionsCreated: atomic.LoadInt64(&m.connectionsCreated),
		ConnectionsClosed:  atomic.LoadInt64(&m.connectionsClosed),
		LastAcquireTime:    atomic.LoadInt64(&m.lastAcquireTime),
		LastReleaseTime:    atomic.LoadInt64(&m.lastReleaseTime),
		LastErrorTime:      atomic.LoadInt64(&m.lastErrorTime),
		LastErrorMsg:       m.lastErrorMsg,
	}

	if stats.AcquireCount > 0 {
		stats.AvgAcquireDuration = stats.AcquireDuration / stats.AcquireCount
	}
	if stats.WaitCount > 0 {
		stats.AvgWaitDuration = stats.WaitDuration / stats.WaitCount
	}

	return stats
}

func (c *Collector) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.instanceMetrics = make(map[string]*instanceMetrics)
}

var defaultCollector = NewCollector()

func DefaultCollector() *Collector {
	return defaultCollector
}
