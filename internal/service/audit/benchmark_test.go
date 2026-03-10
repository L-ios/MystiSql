package audit

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func BenchmarkAuditService_Log(b *testing.B) {
	logger := zap.NewNop()
	service, _ := NewAuditService(&AuditConfig{
		Enabled:       true,
		LogFile:       "/tmp/mystisql_bench_audit.log",
		BufferSize:    1000,
		RetentionDays: 1,
	}, logger)
	defer service.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log := NewAuditLog("bench-user", "127.0.0.1", "test-instance", "testdb", "SELECT * FROM users WHERE id = 1")
		log.SetQueryInfo("SELECT", 1, 5)
		log.SetSuccess()
		service.Log(context.Background(), log)
	}
}

func BenchmarkAuditService_LogParallel(b *testing.B) {
	logger := zap.NewNop()
	service, _ := NewAuditService(&AuditConfig{
		Enabled:       true,
		LogFile:       "/tmp/mystisql_bench_audit_parallel.log",
		BufferSize:    1000,
		RetentionDays: 1,
	}, logger)
	defer service.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log := NewAuditLog("bench-user", "127.0.0.1", "test-instance", "testdb", "SELECT * FROM users WHERE id = ?")
			log.SetQueryInfo("SELECT", 1, 5)
			log.SetSuccess()
			service.Log(context.Background(), log)
		}
	})
}

func TestAuditService_AsyncWriteDoesNotBlock(t *testing.T) {
	logger := zap.NewNop()
	service, _ := NewAuditService(&AuditConfig{
		Enabled:       true,
		LogFile:       "/tmp/mystisql_test_async_audit.log",
		BufferSize:    1000,
		RetentionDays: 1,
	}, logger)
	defer service.Close()

	iterations := 100
	latencies := make([]time.Duration, 0, iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()

		log := NewAuditLog("async-test-user", "127.0.0.1", "test-instance", "testdb", "SELECT * FROM users")
		log.SetQueryInfo("SELECT", 10, 50)
		log.SetSuccess()
		service.Log(context.Background(), log)

		latency := time.Since(start)
		latencies = append(latencies, latency)
	}

	var total time.Duration
	var maxLatency time.Duration
	for _, l := range latencies {
		total += l
		if l > maxLatency {
			maxLatency = l
		}
	}

	avgLatency := total / time.Duration(iterations)

	t.Logf("Audit Log Async Write Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Average Log() call latency: %v", avgLatency)
	t.Logf("  Max Log() call latency: %v", maxLatency)

	if avgLatency > time.Millisecond {
		t.Errorf("Average Log() latency %v exceeds 1ms (blocking detected)", avgLatency)
	}

	if maxLatency > 10*time.Millisecond {
		t.Errorf("Max Log() latency %v exceeds 10ms (potential blocking)", maxLatency)
	}
}

func TestAuditService_HighThroughput(t *testing.T) {
	logger := zap.NewNop()
	service, _ := NewAuditService(&AuditConfig{
		Enabled:       true,
		LogFile:       "/tmp/mystisql_test_throughput_audit.log",
		BufferSize:    1000,
		RetentionDays: 1,
	}, logger)
	defer service.Close()

	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		log := NewAuditLog("throughput-test-user", "127.0.0.1", "test-instance", "testdb", "SELECT * FROM users WHERE id = ?")
		log.SetQueryInfo("SELECT", 1, 5)
		log.SetSuccess()
		service.Log(context.Background(), log)
	}

	totalTime := time.Since(start)
	throughput := float64(iterations) / totalTime.Seconds()

	t.Logf("Audit Log Throughput:")
	t.Logf("  Total logs: %d", iterations)
	t.Logf("  Total time: %v", totalTime)
	t.Logf("  Throughput: %.2f logs/sec", throughput)

	if throughput < 10000 {
		t.Errorf("Throughput %.2f logs/sec is below 10000 logs/sec threshold", throughput)
	}
}
