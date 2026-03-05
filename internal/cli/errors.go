package cli

import (
	"errors"
)

var (
	// ErrConfigNotLoaded 配置未加载
	ErrConfigNotLoaded = errors.New("配置未加载")

	// ErrRegistryNotInitialized 注册中心未初始化
	ErrRegistryNotInitialized = errors.New("注册中心未初始化")
)

// GetConfigFile 获取配置文件路径
func GetConfigFile() string {
	return cfgFile
}
