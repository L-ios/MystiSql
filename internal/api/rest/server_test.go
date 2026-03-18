package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MystiSql/internal/discovery"
	"MystiSql/internal/service/auth"
	"MystiSql/internal/service/query"
	"MystiSql/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestServer(t testing.TB) (*Server, *discovery.Registry) {
	logger := zap.NewNop()
	registry := discovery.NewRegistry()
	engine := query.NewEngine(registry)
	authService, _ := auth.NewAuthService("test-secret-key", 24*time.Hour)

	config := &types.ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
		Mode: "debug",
	}

	server := NewServer(config, &types.WebUIConfig{Enabled: false, Mode: "embedded"}, registry, engine, authService, nil, nil, "", logger, "test-version")

	return server, registry
}

func TestServerSetup(t *testing.T) {
	server, _ := setupTestServer(t)

	if err := server.Setup(); err != nil {
		t.Fatalf("服务器初始化失败: %v", err)
	}

	if server.router == nil {
		t.Error("路由器未初始化")
	}

	if server.handlers == nil {
		t.Error("处理器未初始化")
	}

	if server.server == nil {
		t.Error("HTTP 服务器未初始化")
	}
}

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
			name: "默认模式(未知模式)",
			mode: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			registry := discovery.NewRegistry()
			engine := query.NewEngine(registry)
			authService, _ := auth.NewAuthService("test-secret-key", 24*time.Hour)

			config := &types.ServerConfig{
				Host: "0.0.0.0",
				Port: 8080,
				Mode: tt.mode,
			}

			server := NewServer(config, &types.WebUIConfig{Enabled: false, Mode: "embedded"}, registry, engine, authService, nil, nil, "", logger, "test-version")
			if err := server.Setup(); err != nil {
				t.Fatalf("设置服务器失败: %v", err)
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	server, _ := setupTestServer(t)

	if err := server.Setup(); err != nil {
		t.Fatalf("服务器失败: %v", err)
	}

	router := server.GetRouter()

	t.Run("CORS 中间件", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		req.Header.Set("Origin", "http://external-domain.com")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("日志中间件", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestServerShutdown(t *testing.T) {
	server, _ := setupTestServer(t)

	if err := server.Setup(); err != nil {
		t.Fatalf("服务器初始化失败: %v", err)
	}

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("服务器启动错误: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Errorf("服务器关闭失败: %v", err)
	}
}

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
			engine := query.NewEngine(registry)
			authService, _ := auth.NewAuthService("test-secret-key", 24*time.Hour)

			config := &types.ServerConfig{
				Host: tt.host,
				Port: tt.port,
				Mode: "debug",
			}

			server := NewServer(config, &types.WebUIConfig{Enabled: false, Mode: "embedded"}, registry, engine, authService, nil, nil, "", logger, "test-version")
			if err := server.Setup(); err != nil {
				t.Fatalf("设置服务器失败: %v", err)
			}

			if server.server.Addr != tt.want {
				t.Errorf("服务器地址错误 got %v, want %v", server.server.Addr, tt.want)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	logger := zap.NewNop()
	registry := discovery.NewRegistry()
	engine := query.NewEngine(registry)
	authService, _ := auth.NewAuthService("test-secret-key", 24*time.Hour)

	config := &types.ServerConfig{
		Host: "0.0.0.0",
		Port: 8080,
		Mode: "debug",
	}

	version := "v1.0.0-test"
	server := NewServer(config, &types.WebUIConfig{Enabled: false, Mode: "embedded"}, registry, engine, authService, nil, nil, "", logger, version)

	if server.version != version {
		t.Errorf("版本信息错误 got %v, want %v", server.version, version)
	}

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
		t.Errorf("版本信息错误 got %v, want %v", response.Version, version)
	}
}

func BenchmarkHealthEndpoint(b *testing.B) {
	server, _ := setupTestServer(b)

	if err := server.Setup(); err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	router := server.GetRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkListInstances(b *testing.B) {
	server, registry := setupTestServer(b)

	for i := 0; i < 100; i++ {
		instance := types.NewDatabaseInstance(
			"test-mysql-"+string(rune(i)),
			types.DatabaseTypeMySQL,
			"localhost",
			3306,
		)
		_ = registry.Register(instance)
	}

	if err := server.Setup(); err != nil {
		b.Fatalf("setup failed: %v", err)
	}

	router := server.GetRouter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/instances", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
