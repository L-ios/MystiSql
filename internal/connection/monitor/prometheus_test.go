package monitor

import (
	"strings"
	"testing"

	"MystiSql/internal/connection"
)

func TestNewPrometheusExporter(t *testing.T) {
	collector := NewCollector()
	if collector == nil {
		t.Fatal("expected collector to be created")
	}

	exporter := NewPrometheusExporter(collector)
	if exporter == nil {
		t.Fatal("expected exporter to be created")
	}

	if exporter.collector != collector {
		t.Error("expected collector to be set")
	}
}

func TestFormatStats(t *testing.T) {
	stats := &connection.PoolStats{
		TotalConnections:   10,
		IdleConnections:    5,
		ActiveConnections:  3,
		MaxConnections:     20,
		MinConnections:     2,
		AcquireCount:       100,
		AcquireFailed:      5,
		AvgAcquireDuration: 1500000,
		WaitCount:          20,
		MaxWaitDuration:    3000000,
		QueryCount:         50,
		QueryFailed:        2,
		ExecCount:          30,
		ExecFailed:         1,
		HealthCheckCount:   100,
		HealthCheckFailed:  3,
		LastErrorMsg:       "connection refused",
	}

	result := FormatStats("test-instance", stats)

	if !strings.Contains(result, "test-instance") {
		t.Error("expected instance name in output")
	}

	if !strings.Contains(result, "total=10") {
		t.Error("expected total connections in output")
	}

	if !strings.Contains(result, "idle=5") {
		t.Error("expected idle connections in output")
	}

	if !strings.Contains(result, "active=3") {
		t.Error("expected active connections in output")
	}

	if !strings.Contains(result, "max=20") {
		t.Error("expected max connections in output")
	}

	if !strings.Contains(result, "min=2") {
		t.Error("expected min connections in output")
	}

	if !strings.Contains(result, "failed=5") {
		t.Error("expected failed acquires in output")
	}

	if !strings.Contains(result, "avg_duration=1ms") {
		t.Error("expected avg acquire duration in output")
	}

	if !strings.Contains(result, "Last Error: connection refused") {
		t.Error("expected last error in output")
	}
}

func TestFormatStatsEmptyError(t *testing.T) {
	stats := &connection.PoolStats{
		TotalConnections:  5,
		IdleConnections:   3,
		ActiveConnections: 2,
		MaxConnections:    10,
		MinConnections:    1,
	}

	result := FormatStats("test-instance", stats)

	if strings.Contains(result, "Last Error") {
		t.Error("should not contain Last Error when empty")
	}
}

func TestPrometheusExporterHandleEvent(t *testing.T) {
	collector := NewCollector()
	exporter := NewPrometheusExporter(collector)

	collector.RecordAcquire("test-instance", 1000000, true)
	collector.RecordAcquire("test-instance", 2000000, false)
	collector.RecordWait("test-instance", 500000)
	collector.RecordQuery("test-instance", 300000, true)
	collector.RecordQuery("test-instance", 400000, false)
	collector.RecordExec("test-instance", 250000, true)
	collector.RecordHealthCheck("test-instance", true)
	collector.RecordHealthCheck("test-instance", false)
	collector.RecordConnectionCreated("test-instance")
	collector.RecordConnectionClosed("test-instance")

	exporter.UpdateMetrics()

	metrics := collector.GetAllMetrics()
	if metrics == nil {
		t.Fatal("expected metrics")
	}

	instanceMetrics, exists := metrics["test-instance"]
	if !exists {
		t.Fatal("expected test-instance in metrics")
	}

	if instanceMetrics.AcquireCount != 2 {
		t.Errorf("expected AcquireCount=2, got %d", instanceMetrics.AcquireCount)
	}

	if instanceMetrics.AcquireFailed != 1 {
		t.Errorf("expected AcquireFailed=1, got %d", instanceMetrics.AcquireFailed)
	}

	if instanceMetrics.WaitCount != 1 {
		t.Errorf("expected WaitCount=1, got %d", instanceMetrics.WaitCount)
	}

	if instanceMetrics.QueryCount != 2 {
		t.Errorf("expected QueryCount=2, got %d", instanceMetrics.QueryCount)
	}

	if instanceMetrics.QueryFailed != 1 {
		t.Errorf("expected QueryFailed=1, got %d", instanceMetrics.QueryFailed)
	}

	if instanceMetrics.ExecCount != 1 {
		t.Errorf("expected ExecCount=1, got %d", instanceMetrics.ExecCount)
	}

	if instanceMetrics.HealthCheckCount != 2 {
		t.Errorf("expected HealthCheckCount=2, got %d", instanceMetrics.HealthCheckCount)
	}

	if instanceMetrics.HealthCheckFailed != 1 {
		t.Errorf("expected HealthCheckFailed=1, got %d", instanceMetrics.HealthCheckFailed)
	}

	if instanceMetrics.ConnectionsCreated != 1 {
		t.Errorf("expected ConnectionsCreated=1, got %d", instanceMetrics.ConnectionsCreated)
	}

	if instanceMetrics.ConnectionsClosed != 1 {
		t.Errorf("expected ConnectionsClosed=1, got %d", instanceMetrics.ConnectionsClosed)
	}
}

func TestPrometheusExporterUpdateMetrics(t *testing.T) {
	collector := NewCollector()
	exporter := NewPrometheusExporter(collector)

	poolStats := &connection.PoolStats{
		TotalConnections:  15,
		IdleConnections:   8,
		ActiveConnections: 5,
		MaxConnections:    25,
		MinConnections:    3,
	}

	collector.UpdatePoolStats("test-instance", poolStats)

	exporter.UpdateMetrics()

	metrics := collector.GetAllMetrics()
	instanceMetrics := metrics["test-instance"]

	if instanceMetrics.TotalConnections != 15 {
		t.Errorf("expected TotalConnections=15, got %d", instanceMetrics.TotalConnections)
	}

	if instanceMetrics.IdleConnections != 8 {
		t.Errorf("expected IdleConnections=8, got %d", instanceMetrics.IdleConnections)
	}

	if instanceMetrics.ActiveConnections != 5 {
		t.Errorf("expected ActiveConnections=5, got %d", instanceMetrics.ActiveConnections)
	}

	if instanceMetrics.MaxConnections != 25 {
		t.Errorf("expected MaxConnections=25, got %d", instanceMetrics.MaxConnections)
	}

	if instanceMetrics.MinConnections != 3 {
		t.Errorf("expected MinConnections=3, got %d", instanceMetrics.MinConnections)
	}
}
