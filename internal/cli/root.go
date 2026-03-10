package cli

import (
	"context"
	"fmt"

	"MystiSql/internal/config"
	"MystiSql/internal/discovery"
	"MystiSql/internal/discovery/static"
	"MystiSql/pkg/types"

	"github.com/spf13/cobra"
)

var (
	cfgFile    string
	verbose    bool
	cfg        *types.Config
	registry   discovery.InstanceRegistry
	rootCtx    context.Context
	rootCancel context.CancelFunc
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "mystisql",
	Short: "MystiSql - Kubernetes 数据库访问网关",
	Long: `MystiSql 是一个数据库访问网关，支持 MySQL、PostgreSQL、Oracle 和 Redis。

它提供了统一的访问接口，包括 CLI、WebUI、RESTful API、WebSocket 和 JDBC 驱动。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 初始化日志
		if err := InitLogger(verbose); err != nil {
			return fmt.Errorf("初始化日志失败: %w", err)
		}

		// 创建上下文
		rootCtx, rootCancel = context.WithCancel(context.Background())

		// 加载配置
		var err error
		if cfgFile != "" {
			cfg, err = config.LoadFromPath(cfgFile)
			if err != nil {
				return fmt.Errorf("加载配置失败: %w", err)
			}
			GetSugar().Debugf("从文件加载配置: %s", cfgFile)
		} else {
			cfg, err = config.LoadDefault()
			if err != nil {
				return fmt.Errorf("加载默认配置失败: %w", err)
			}
			GetSugar().Debug("使用默认配置")
		}

		// 初始化实例注册中心
		registry = discovery.NewRegistry()

		// 如果配置了静态实例，加载到注册中心
		if len(cfg.Instances) > 0 {
			discoverer := static.NewDiscoverer(cfg.Instances)
			instances, err := discoverer.Discover(rootCtx)
			if err != nil {
				return fmt.Errorf("发现实例失败: %w", err)
			}

			for _, instance := range instances {
				if err := registry.Register(instance); err != nil {
					GetSugar().Warnf("注册实例失败: %v", err)
				} else {
					GetSugar().Debugf("注册实例: %s", instance.Name)
				}
			}
		}

		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// 清理资源
		if rootCancel != nil {
			rootCancel()
		}
		// 忽略同步标准输出的错误
		_ = Sync()
	},
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细日志输出")
	rootCmd.PersistentFlags().StringVar(&tokenFlag, "token", "", "认证 Token")

	// 添加 TUI 命令
	rootCmd.AddCommand(NewTUICmd())
}

// GetConfig 获取配置
func GetConfig() *types.Config {
	return cfg
}

// GetRegistry 获取实例注册中心
func GetRegistry() discovery.InstanceRegistry {
	return registry
}

// GetContext 获取根上下文
func GetContext() context.Context {
	return rootCtx
}
