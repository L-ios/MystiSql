package cli

import (
	"context"
	"fmt"

	"MystiSql/internal/cli/repl"
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

var rootCmd = &cobra.Command{
	Use:   "mystisql",
	Short: "MystiSql - Kubernetes 数据库访问网关",
	Long: `MystiSql 是一个数据库访问网关，支持 MySQL、PostgreSQL、Oracle 和 Redis。

它提供了统一的访问接口，包括 CLI、WebUI、RESTful API、WebSocket 和 JDBC 驱动。

默认启动交互式 REPL 界面。使用子命令（如 query、serve）执行其他操作。`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := InitLogger(verbose); err != nil {
			return fmt.Errorf("初始化日志失败: %w", err)
		}

		rootCtx, rootCancel = context.WithCancel(context.Background())

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

		registry = discovery.NewRegistry()

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
	RunE: func(cmd *cobra.Command, args []string) error {
		r := repl.NewREPL(GetConfig(), GetRegistry())
		if err := r.Run(); err != nil {
			return fmt.Errorf("REPL error: %w", err)
		}
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if rootCancel != nil {
			rootCancel()
		}
		_ = Sync()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "配置文件路径")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细日志输出")
	rootCmd.PersistentFlags().StringVar(&tokenFlag, "token", "", "认证 Token")
}

func GetConfig() *types.Config {
	return cfg
}

func GetRegistry() discovery.InstanceRegistry {
	return registry
}

func GetContext() context.Context {
	return rootCtx
}
