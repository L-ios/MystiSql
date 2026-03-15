package monitor

import (
	"testing"
	"time"

	"MystiSql/internal/connection"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	if c == nil {
		t.Fatal("expected collector to be created")
	}
}

func TestRecordAcquire(t *testing.T) {
	c := NewCollector()

	c.RecordAcquire("test-instance", 1000000, true)
	c.RecordAcquire("test-instance", 2000000, false)

	metrics := c.GetInstanceMetrics("test-instance")
	if metrics == nil {
		t.Fatal("expected metrics for instance")
	}

	if metrics.AcquireCount != 2 {
		t.Errorf("expected AcquireCount=2, got %d", metrics.AcquireCount)
	}

	if metrics.AcquireFailed != 1 {
		t.Errorf("expected AcquireFailed=1, got %d", metrics.AcquireFailed)
	}

	if metrics.MaxAcquireDuration != 2000000 {
		t.Errorf("expected MaxAcquireDuration=2000000, got %d", metrics.MaxAcquireDuration)
	}
}

func TestRecordWait(t *testing.T) {
	c := NewCollector()

	c.RecordWait("test-instance", 5000000)
	c.RecordWait("test-instance", 3000000)

	metrics := c.GetInstanceMetrics("test-instance")
	if metrics == nil {
		t.Fatal("expected metrics for instance")
	}

	if metrics.WaitCount != 2 {
		t.Errorf("expected WaitCount=2, got %d", metrics.WaitCount)
	}

	if metrics.MaxWaitDuration != 5000000 {
		t.Errorf("expected MaxWaitDuration=5000000, got %d", metrics.MaxWaitDuration)
	}

	if metrics.AvgWaitDuration != 4000000 {
		t.Errorf("expected AvgWaitDuration=4000000, got %d", metrics.AvgWaitDuration)
	}
}

func TestRecordQueryExec(t *testing.T) {
	c := NewCollector()

	c.RecordQuery("test-instance", 1000000, true)
	c.RecordQuery("test-instance", 2000000, false)
	c.RecordExec("test-instance", 500000, true)
	c.RecordExec("test-instance", 600000, false)

	metrics := c.GetInstanceMetrics("test-instance")
	if metrics == nil {
		t.Fatal("expected metrics for instance")
	}

	if metrics.QueryCount != 2 {
		t.Errorf("expected QueryCount=2, got %d", metrics.QueryCount)
	}

	if metrics.QueryFailed != 1 {
		t.Errorf("expected QueryFailed=1, got %d", metrics.QueryFailed)
	}

	if metrics.ExecCount != 2 {
		t.Errorf("expected ExecCount=2, got %d", metrics.ExecCount)
	}

	if metrics.ExecFailed != 1 {
		t.Errorf("expected ExecFailed=1, got %d", metrics.ExecFailed)
	}
}

func TestRecordHealthCheck(t *testing.T) {
	c := NewCollector()

	c.RecordHealthCheck("test-instance", true)
	c.RecordHealthCheck("test-instance", false)
	c.RecordHealthCheck("test-instance", true)

	metrics := c.GetInstanceMetrics("test-instance")
	if metrics == nil {
		t.Fatal("expected metrics for instance")
	}

	if metrics.HealthCheckCount != 3 {
		t.Errorf("expected HealthCheckCount=3, got %d", metrics.HealthCheckCount)
	}

	if metrics.HealthCheckFailed != 1 {
		t.Errorf("expected HealthCheckFailed=1, got %d", metrics.HealthCheckFailed)
	}
}

func TestRecordConnectionLifecycle(t *testing.T) {
	c := NewCollector()

	c.RecordConnectionCreated("test-instance")
	c.RecordConnectionCreated("test-instance")
	c.RecordConnectionClosed("test-instance")

	metrics := c.GetInstanceMetrics("test-instance")
	if metrics == nil {
		t.Fatal("expected metrics for instance")
	}

	if metrics.ConnectionsCreated != 2 {
		t.Errorf("expected ConnectionsCreated=2, got %d", metrics.ConnectionsCreated)
	}

	if metrics.ConnectionsClosed != 1 {
		t.Errorf("expected ConnectionsClosed=1, got %d", metrics.ConnectionsClosed)
	}
}

func TestUpdatePoolStats(t *testing.T) {
	c := NewCollector()

	stats := &connection.PoolStats{
		TotalConnections:  10,
		IdleConnections:   5,
		ActiveConnections: 5,
		MaxConnections:    20,
		MinConnections:    2,
	}

	c.UpdatePoolStats("test-instance", stats)

	metrics := c.GetInstanceMetrics("test-instance")
	if metrics == nil {
		t.Fatal("expected metrics for instance")
	}

	if metrics.TotalConnections != 10 {
		t.Errorf("expected TotalConnections=10, got %d", metrics.TotalConnections)
	}

	if metrics.IdleConnections != 5 {
		t.Errorf("expected IdleConnections=5, got %d", metrics.IdleConnections)
	}

	if metrics.ActiveConnections != 5 {
		t.Errorf("expected ActiveConnections=5, got %d", metrics.ActiveConnections)
	}
}

func TestGetAllMetrics(t *testing.T) {
	c := NewCollector()

	c.RecordAcquire("instance-1", 1000000, true)
	c.RecordAcquire("instance-2", 2000000, true)
	c.RecordQuery("instance-1", 500000, true)

	allMetrics := c.GetAllMetrics()
	if len(allMetrics) != 2 {
		t.Errorf("expected 2 instances, got %d", len(allMetrics))
	}

	if _, exists := allMetrics["instance-1"]; !exists {
		t.Error("expected instance-1 to exist in metrics")
	}

	if _, exists := allMetrics["instance-2"]; !exists {
		t.Error("expected instance-2 to exist in metrics")
	}
}

func TestEventHandler(t *testing.T) {
	c := NewCollector()

	var receivedEvent connection.MetricsEvent
	c.RegisterEventHandler(func(event connection.MetricsEvent) {
		receivedEvent = event
	})

	c.RecordAcquire("test-instance", 1000000, true)

	time.Sleep(10 * time.Millisecond)

	if receivedEvent.Type != "acquire" {
		t.Errorf("expected event type 'acquire', got %s", receivedEvent.Type)
	}

	if receivedEvent.Instance != "test-instance" {
		t.Errorf("expected instance 'test-instance', got %s", receivedEvent.Instance)
	}
}

func TestReset(t *testing.T) {
	c := NewCollector()

	c.RecordAcquire("test-instance", 1000000, true)
	c.RecordQuery("test-instance", 500000, true)

	c.Reset()

	metrics := c.GetInstanceMetrics("test-instance")
	if metrics != nil {
		t.Error("expected metrics to be nil after reset")
	}

	allMetrics := c.GetAllMetrics()
	if len(allMetrics) != 0 {
		t.Errorf("expected 0 instances after reset, got %d", len(allMetrics))
	}
}

func TestDefaultCollector(t *testing.T) {
	c1 := DefaultCollector()
	c2 := DefaultCollector()

	if c1 != c2 {
		t.Error("expected same default collector instance")
	}
}

func TestRecordError(t *testing.T) {
	c := NewCollector()

	c.RecordError("test-instance", "connection refused")

	metrics := c.GetInstanceMetrics("test-instance")
	if metrics == nil {
		t.Fatal("expected metrics for instance")
	}

	if metrics.LastErrorMsg != "connection refused" {
		t.Errorf("expected LastErrorMsg='connection refused', got %s", metrics.LastErrorMsg)
	}

	if metrics.LastErrorTime == 0 {
		t.Error("expected LastErrorTime to be set")
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := NewCollector()

	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				c.RecordAcquire("test-instance", int64(j*1000), true)
				c.RecordQuery("test-instance", int64(j*500), j%2 == 0)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	metrics := c.GetInstanceMetrics("test-instance")
	if metrics == nil {
		t.Fatal("expected metrics for instance")
	}

	if metrics.AcquireCount != 1000 {
		t.Errorf("expected AcquireCount=1000, got %d", metrics.AcquireCount)
	}

	if metrics.QueryCount != 1000 {
		t.Errorf("expected QueryCount=1000, got %d", metrics.QueryCount)
	}
}
