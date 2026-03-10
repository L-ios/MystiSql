package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"MystiSql/internal/service/auth"
)

func BenchmarkAuthMiddleware_TokenValidation(b *testing.B) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	authService, _ := auth.NewAuthService("benchmark-secret-key", 24*time.Hour)
	token, _ := authService.GenerateToken(context.Background(), "bench-user", "admin")

	router := gin.New()
	router.Use(AuthMiddleware(authService, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkAuthMiddleware_Latency(b *testing.B) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	authService, _ := auth.NewAuthService("benchmark-secret-key", 24*time.Hour)
	token, _ := authService.GenerateToken(context.Background(), "bench-user", "admin")

	router := gin.New()
	router.Use(AuthMiddleware(authService, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

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

	avgLatency := total / time.Duration(len(latencies))
	b.ReportMetric(float64(avgLatency.Nanoseconds()), "avg_ns")
	b.ReportMetric(float64(maxLatency.Nanoseconds()), "max_ns")

	if avgLatency > time.Millisecond {
		b.Logf("WARNING: Average latency %v exceeds 1ms requirement", avgLatency)
	}
}

func BenchmarkAuthMiddleware_WithoutToken(b *testing.B) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	authService, _ := auth.NewAuthService("benchmark-secret-key", 24*time.Hour)

	router := gin.New()
	router.Use(AuthMiddleware(authService, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkAuthMiddleware_Parallel(b *testing.B) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	authService, _ := auth.NewAuthService("benchmark-secret-key", 24*time.Hour)
	token, _ := authService.GenerateToken(context.Background(), "bench-user", "admin")

	router := gin.New()
	router.Use(AuthMiddleware(authService, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}
	})
}

func TestAuthMiddleware_LatencyRequirement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	authService, _ := auth.NewAuthService("test-secret-key", 24*time.Hour)
	token, _ := authService.GenerateToken(context.Background(), "test-user", "admin")

	router := gin.New()
	router.Use(AuthMiddleware(authService, logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	iterations := 1000
	latencies := make([]time.Duration, 0, iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		latency := time.Since(start)
		latencies = append(latencies, latency)
	}

	var total time.Duration
	var maxLatency time.Duration
	var countOver1ms int

	for _, l := range latencies {
		total += l
		if l > maxLatency {
			maxLatency = l
		}
		if l > time.Millisecond {
			countOver1ms++
		}
	}

	avgLatency := total / time.Duration(iterations)

	t.Logf("Authentication Middleware Performance:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Average latency: %v", avgLatency)
	t.Logf("  Max latency: %v", maxLatency)
	t.Logf("  Requests over 1ms: %d (%.2f%%)", countOver1ms, float64(countOver1ms)/float64(iterations)*100)

	if avgLatency > time.Millisecond {
		t.Errorf("Average latency %v exceeds 1ms requirement", avgLatency)
	}

	p99Latency := calculatePercentile(latencies, 99)
	t.Logf("  P99 latency: %v", p99Latency)

	if p99Latency > 2*time.Millisecond {
		t.Errorf("P99 latency %v exceeds 2ms", p99Latency)
	}
}

func calculatePercentile(latencies []time.Duration, percentile int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	index := (percentile * len(sorted)) / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}
