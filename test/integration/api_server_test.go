package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"MystiSql/internal/api/rest"
	"MystiSql/internal/discovery"
	"MystiSql/pkg/types"
	"go.uber.org/zap"
)

func TestAPIServerLifecycle(t *testing.T) {
	cfg := &types.ServerConfig{
		Host: "127.0.0.1",
		Port: 18080,
		Mode: "debug",
	}

	registry := discovery.NewRegistry()
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	instance.SetCredentials("root", "root")
	instance.SetDatabase("test")
	if err := registry.Register(instance); err != nil {
		t.Fatalf("注册实例失败: %v", err)
	}

	logger := zap.NewNop()
	server := rest.NewServer(cfg, registry, logger, "test")

	if err := server.Setup(); err != nil {
		t.Fatalf("设置服务器失败: %v", err)
	}

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("服务器启动错误: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:18080/health")
	if err != nil {
		t.Fatalf("健康检查请求失败: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 200，实际为 %d", resp.StatusCode)
	}

	t.Logf("健康检查成功，状态码: %d", resp.StatusCode)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("关闭服务器失败: %v", err)
	}

	t.Logf("服务器关闭成功")
}

func TestAPIEndpoints(t *testing.T) {
	cfg := &types.ServerConfig{
		Host: "127.0.0.1",
		Port: 18081,
		Mode: "debug",
	}

	registry := discovery.NewRegistry()
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	instance.SetCredentials("root", "root")
	instance.SetDatabase("test")
	if err := registry.Register(instance); err != nil {
		t.Fatalf("注册实例失败: %v", err)
	}

	logger := zap.NewNop()
	server := rest.NewServer(cfg, registry, logger, "test")

	if err := server.Setup(); err != nil {
		t.Fatalf("设置服务器失败: %v", err)
	}

	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("服务器启动错误: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	t.Run("健康检查端点", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:18081/health")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码 200，实际为 %d", resp.StatusCode)
		}
	})

	t.Run("实例列表端点", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:18081/api/v1/instances")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码 200，实际为 %d", resp.StatusCode)
		}
	})

	t.Run("带健康检查的实例列表", func(t *testing.T) {
		resp, err := http.Get("http://127.0.0.1:18081/health?check-instances=true")
		if err != nil {
			t.Fatalf("请求失败: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("期望状态码 200，实际为 %d", resp.StatusCode)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Logf("关闭服务器失败: %v", err)
	}
}
