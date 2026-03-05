package config

import (
	"fmt"
	"strings"

	"MystiSql/pkg/errors"
	"MystiSql/pkg/types"
	"github.com/spf13/viper"
)

const (
	// 配置文件名称（不带扩展名）
	configName = "config"
	// 配置文件类型
	configType = "yaml"
	// 环境变量前缀
	envPrefix = "MYSTISQL"
)

// Loader 配置加载器
type Loader struct {
	viper *viper.Viper
}

// NewLoader 创建一个新的配置加载器
func NewLoader() *Loader {
	v := viper.New()

	// 设置配置文件名称和类型
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	// 设置环境变量支持
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 设置默认值
	setDefaults(v)

	return &Loader{viper: v}
}

// setDefaults 设置默认配置值
func setDefaults(v *viper.Viper) {
	// 服务器默认值
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")

	// 发现默认值
	v.SetDefault("discovery.type", "static")
}

// Load 从指定路径加载配置
func (l *Loader) Load(configPath string) (*types.Config, error) {
	if configPath != "" {
		// 使用指定的配置文件路径
		l.viper.SetConfigFile(configPath)
	} else {
		// 添加多个配置文件搜索路径（按优先级从高到低）
		l.viper.AddConfigPath(".")             // 当前目录（最高优先级）
		l.viper.AddConfigPath("./config")      // ./config 目录
		l.viper.AddConfigPath("/etc/mystisql") // 系统配置目录（最低优先级）
	}

	// 读取配置文件
	if err := l.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("%w: %s", errors.ErrConfigNotFound, configPath)
		}
		return nil, fmt.Errorf("%w: %v", errors.ErrConfigParseFailed, err)
	}

	// 解析配置到结构体
	var cfg types.Config
	if err := l.viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrConfigParseFailed, err)
	}

	// 验证配置
	if err := Validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadFromPath 从指定路径加载配置（便捷方法）
func LoadFromPath(configPath string) (*types.Config, error) {
	loader := NewLoader()
	return loader.Load(configPath)
}

// LoadDefault 使用默认路径加载配置
func LoadDefault() (*types.Config, error) {
	loader := NewLoader()
	return loader.Load("")
}

// Validate 验证配置的有效性
func Validate(cfg *types.Config) error {
	// 验证服务器配置
	if err := validateServerConfig(&cfg.Server); err != nil {
		return err
	}

	// 验证发现配置
	if err := validateDiscoveryConfig(&cfg.Discovery); err != nil {
		return err
	}

	// 验证实例配置
	if err := validateInstances(cfg.Instances); err != nil {
		return err
	}

	return nil
}

// validateServerConfig 验证服务器配置
func validateServerConfig(cfg *types.ServerConfig) error {
	// 验证端口范围
	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("%w: 端口号必须在 1-65535 范围内，当前值: %d",
			errors.ErrConfigInvalid, cfg.Port)
	}

	// 验证运行模式
	if cfg.Mode != "debug" && cfg.Mode != "release" {
		return fmt.Errorf("%w: server.mode 必须是 'debug' 或 'release'，当前值: %s",
			errors.ErrConfigInvalid, cfg.Mode)
	}

	return nil
}

// validateDiscoveryConfig 验证发现配置
func validateDiscoveryConfig(cfg *types.DiscoveryConfig) error {
	// Phase 1 只支持静态发现
	supportedTypes := []string{"static"}
	isValid := false
	for _, t := range supportedTypes {
		if cfg.Type == t {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("%w: discovery.type '%s' 不支持，支持的类型: %v",
			errors.ErrConfigInvalid, cfg.Type, supportedTypes)
	}

	return nil
}

// validateInstances 验证实例配置
func validateInstances(instances []types.InstanceConfig) error {
	// 检查实例名称是否重复
	nameSet := make(map[string]bool)

	for i, instance := range instances {
		// 验证必填字段
		if instance.Name == "" {
			return fmt.Errorf("%w: instances[%d].name 不能为空",
				errors.ErrMissingRequiredField, i)
		}

		if instance.Host == "" {
			return fmt.Errorf("%w: instances[%d].host 不能为空",
				errors.ErrMissingRequiredField, i)
		}

		if instance.Port < 1 || instance.Port > 65535 {
			return fmt.Errorf("%w: instances[%d].port 必须在 1-65535 范围内，当前值: %d",
				errors.ErrConfigInvalid, i, instance.Port)
		}

		// 验证数据库类型（Phase 1 只支持 MySQL）
		supportedTypes := []types.DatabaseType{types.DatabaseTypeMySQL}
		isValid := false
		for _, t := range supportedTypes {
			if instance.Type == t {
				isValid = true
				break
			}
		}

		if !isValid {
			return fmt.Errorf("%w: instances[%d].type '%s' 不支持，Phase 1 支持的类型: %v",
				errors.ErrConfigInvalid, i, instance.Type, supportedTypes)
		}

		// 检查名称重复
		if nameSet[instance.Name] {
			return fmt.Errorf("%w: 实例名称 '%s' 重复",
				errors.ErrConfigInvalid, instance.Name)
		}
		nameSet[instance.Name] = true
	}

	return nil
}
