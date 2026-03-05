package config

import (
	"os"
	"path/filepath"
	"testing"

	"MystiSql/pkg/types"
)

func TestLoadFromPath(t *testing.T) {
	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  host: 127.0.0.1
  port: 9090
  mode: debug

discovery:
  type: static

instances:
  - name: test-mysql
    type: mysql
    host: localhost
    port: 3306
    username: root
    password: root
    database: test_db
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("写入配置文件失败: %v", err)
	}

	// 加载配置
	cfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}

	// 验证服务器配置
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("期望 Server.Host = '127.0.0.1', 实际 = '%s'", cfg.Server.Host)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("期望 Server.Port = 9090, 实际 = %d", cfg.Server.Port)
	}

	if cfg.Server.Mode != "debug" {
		t.Errorf("期望 Server.Mode = 'debug', 实际 = '%s'", cfg.Server.Mode)
	}

	// 验证发现配置
	if cfg.Discovery.Type != "static" {
		t.Errorf("期望 Discovery.Type = 'static', 实际 = '%s'", cfg.Discovery.Type)
	}

	// 验证实例配置
	if len(cfg.Instances) != 1 {
		t.Fatalf("期望 1 个实例, 实际 = %d", len(cfg.Instances))
	}

	instance := cfg.Instances[0]
	if instance.Name != "test-mysql" {
		t.Errorf("期望实例名称 = 'test-mysql', 实际 = '%s'", instance.Name)
	}

	if instance.Host != "localhost" {
		t.Errorf("期望 Host = 'localhost', 实际 = '%s'", instance.Host)
	}

	if instance.Port != 3306 {
		t.Errorf("期望 Port = 3306, 实际 = %d", instance.Port)
	}
}

func TestLoadFromPath_FileNotFound(t *testing.T) {
	_, err := LoadFromPath("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("期望返回错误，但加载成功了")
	}
}

func TestValidate_InvalidPort(t *testing.T) {
	cfg := types.NewConfig()
	cfg.Server.Port = 99999 // 无效端口

	err := Validate(cfg)
	if err == nil {
		t.Error("期望返回端口验证错误，但验证通过了")
	}
}

func TestValidate_InvalidMode(t *testing.T) {
	cfg := types.NewConfig()
	cfg.Server.Mode = "invalid" // 无效模式

	err := Validate(cfg)
	if err == nil {
		t.Error("期望返回模式验证错误，但验证通过了")
	}
}

func TestValidate_MissingInstanceName(t *testing.T) {
	cfg := types.NewConfig()
	cfg.Instances = []types.InstanceConfig{
		{
			Type: types.DatabaseTypeMySQL,
			Host: "localhost",
			Port: 3306,
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("期望返回缺少实例名称错误，但验证通过了")
	}
}

func TestValidate_DuplicateInstanceName(t *testing.T) {
	cfg := types.NewConfig()
	cfg.Instances = []types.InstanceConfig{
		{
			Name: "duplicate",
			Type: types.DatabaseTypeMySQL,
			Host: "localhost",
			Port: 3306,
		},
		{
			Name: "duplicate", // 重复名称
			Type: types.DatabaseTypeMySQL,
			Host: "localhost",
			Port: 3307,
		},
	}

	err := Validate(cfg)
	if err == nil {
		t.Error("期望返回重复名称错误，但验证通过了")
	}
}
