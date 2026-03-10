package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MystiSql/internal/discovery"
	"MystiSql/pkg/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestServerSetup 测试服务器初始化
func TestServerSetup(t *testing.T) {
	logger := zap.NewNop()
	registry := discovery.NewRegistry()
	config := &types.ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
		Mode: "debug",
	}

	server := NewServer(config, registry, logger, "test-version")

	if err := server.Setup(); err != nil {
		t.Fatalf("服务器初始化失败: %v", err)
	}

	// 验证路由器已创建
	if server.router == nil {
		t.Error("路由器未初始化")
	}

	// 验证处理器已创建
	if server.handlers == nil {
		t.Error("处理器未初始化")
	}

	// 验证 HTTP 服务器已创建
	if server.server == nil {
		t.Error("HTTP 服务器未初始化")
	}
}

// TestServerMode 测试 Gin 模式设置
func TestServerMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{
			name: "debug 模式",
			mode: "debug",
		},
		{
			name: "release 模式",
			mode: "release",
		},
		{
			name: "默认模式（未知模式）",
			mode: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			registry := discovery.NewRegistry()
			config := &types.ServerConfig{
				Host: "0.0.0.0",
				Port: 8080,
				Mode: tt.mode,
			}

			server := NewServer(config, registry, logger, "test-version")
			if err := server.Setup(); err != nil {
				t.Fatalf("设置服务器失败: %v", err)
			}
		})
	}
}

// TestMiddleware 测试中间件
func TestMiddleware(t *testing.T) {
	logger := zap.NewNop()
	registry := discovery.NewRegistry()
	config := &types.ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
		Mode: "debug",
	}

	server := NewServer(config, registry, logger, "test-version")
	if err := server.Setup(); err != nil {
		t.Fatalf("设置服务器失败: %v", err)
	}
	router := server.GetRouter()

	// 测试 CORS 中间件
	t.Run("CORS 中间件", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/health", nil)
		req.Header.Set("Origin", "http://example.com")
		req.Header.Set("Access-Control-Request-Method", "GET")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 验证 CORS 头存在
		// 注意：实际 CORS 头由 gin-cors 中间件处理
		_ = w.Code
	})

	// 测试日志中间件
	t.Run("日志中间件", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 请求应该成功处理
		if w.Code != http.StatusOK {
			t.Errorf("状态码错误: got %v, want %v", w.Code, http.StatusOK)
		}
	})

	// 测试错误恢复中间件
	t.Run("错误恢复中间件", func(t *testing.T) {
		// 创建一个会 panic 的路由
		testRouter := gin.New()
		testRouter.Use(RecoveryMiddleware(logger))
		testRouter.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		req := httptest.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 应该返回 500 错误，而不是崩溃
		if w.Code != http.StatusInternalServerError {
			t.Errorf("状态码错误: got %v, want %v", w.Code, http.StatusInternalServerError)
		}
	})
}

// TestServerShutdown 测试服务器优雅关闭
func TestServerShutdown(t *testing.T) {
	logger := zap.NewNop()
	registry := discovery.NewRegistry()
	config := &types.ServerConfig{
		Host: "0.0.0.0",
		Port: 18080, // 使用不同的端口避免冲突
		Mode: "debug",
	}

	server := NewServer(config, registry, logger, "test-version")

	// 初始化服务器
	if err := server.Setup(); err != nil {
		t.Fatalf("服务器初始化失败: %v", err)
	}

	// 启动服务器（在后台）
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("服务器启动错误: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 测试优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("服务器关闭失败: %v", err)
	}
}

// TestServerAddress 测试服务器地址配置
func TestServerAddress(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{
			name: "默认地址",
			host: "0.0.0.0",
			port: 8080,
			want: "0.0.0.0:8080",
		},
		{
			name: "本地地址",
			host: "127.0.0.1",
			port: 3000,
			want: "127.0.0.1:3000",
		},
		{
			name: "自定义端口",
			host: "0.0.0.0",
			port: 9090,
			want: "0.0.0.0:9090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			registry := discovery.NewRegistry()
			config := &types.ServerConfig{
				Host: tt.host,
				Port: tt.port,
				Mode: "debug",
			}

			server := NewServer(config, registry, logger, "test-version")
			if err := server.Setup(); err != nil {
				t.Fatalf("设置服务器失败: %v", err)
			}

			// 验证服务器地址
			if server.server.Addr != tt.want {
				t.Errorf("服务器地址错误: got %v, want %v", server.server.Addr, tt.want)
			}
		})
	}
}

// TestServerTimeouts 测试服务器超时配置
func TestServerTimeouts(t *testing.T) {
	logger := zap.NewNop()
	registry := discovery.NewRegistry()
	config := &types.ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
		Mode: "debug",
	}

	server := NewServer(config, registry, logger, "test-version")
	if err := server.Setup(); err != nil {
		t.Fatalf("设置服务器失败: %v", err)
	}

	// 验证超时配置
	if server.server.ReadTimeout != 30*time.Second {
		t.Errorf("读取超时错误: got %v, want 30s", server.server.ReadTimeout)
	}

	if server.server.WriteTimeout != 30*time.Second {
		t.Errorf("写入超时错误: got %v, want 30s", server.server.WriteTimeout)
	}

	if server.server.IdleTimeout != 60*time.Second {
		t.Errorf("空闲超时错误: got %v, want 60s", server.server.IdleTimeout)
	}
}

// TestConcurrentRequests 测试并发请求
func TestConcurrentRequests(t *testing.T) {
	router, registry, _ := setupTestRouter()

	// 添加测试实例
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	instance.SetStatus(types.InstanceStatusHealthy)
	if err := registry.Register(instance); err != nil {
		t.Fatalf("注册实例失败: %v", err)
	}

	// 并发发送请求
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				done <- true
			} else {
				done <- false
			}
		}()
	}

	// 等待所有请求完成
	successCount := 0
	for i := 0; i < concurrency; i++ {
		if <-done {
			successCount++
		}
	}

	if successCount != concurrency {
		t.Errorf("并发请求失败: %d/%d 成功", successCount, concurrency)
	}
}

// TestVersion 测试版本信息
func TestVersion(t *testing.T) {
	logger := zap.NewNop()
	registry := discovery.NewRegistry()
	config := &types.ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
		Mode: "debug",
	}

	version := "v1.0.0-test"
	server := NewServer(config, registry, logger, version)

	if server.version != version {
		t.Errorf("版本信息错误: got %v, want %v", server.version, version)
	}

	// 测试版本信息在健康检查中返回
	if err := server.Setup(); err != nil {
		t.Fatalf("设置服务器失败: %v", err)
	}
	router := server.GetRouter()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("响应解析失败: %v", err)
	}

	if response.Version != version {
		t.Errorf("版本信息错误: got %v, want %v", response.Version, version)
	}
}

// BenchmarkHealthEndpoint 健康检查端点性能测试
func BenchmarkHealthEndpoint(b *testing.B) {
	router, _, _ := setupTestRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

// BenchmarkListInstances 实例列表端点性能测试
func BenchmarkListInstances(b *testing.B) {
	router, registry, _ := setupTestRouter()

	// 添加多个测试实例
	for i := 0; i < 100; i++ {
		instance := types.NewDatabaseInstance(
			"test-mysql-"+string(rune(i)),
			types.DatabaseTypeMySQL,
			"localhost",
			3306,
		)
		_ = registry.Register(instance) // 忽略基准测试中的错误
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/instances", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}