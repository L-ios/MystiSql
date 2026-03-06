package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"MystiSql/internal/api/rest"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// serveCmd API 服务器命令
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "启动 REST API 服务器",
	Long: `启动 MystiSql REST API 服务器。

服务器将提供以下接口：
- GET  /health          - 健康检查
- GET  /api/v1/instances - 列出所有数据库实例
- POST /api/v1/query     - 执行 SQL 查询

示例：
  mystisql serve
  mystisql serve --port 9090
  mystisql serve --host 0.0.0.0 --port 8080`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 获取配置
		cfg := GetConfig()
		if cfg == nil {
			return fmt.Errorf("配置未初始化")
		}

		// 创建 logger
		logger, err := zap.NewProduction()
		if err != nil {
			return fmt.Errorf("初始化日志失败: %w", err)
		}
		defer func() {
			_ = logger.Sync()
		}()

		GetSugar().Info("MystiSql API 服务器启动中...")

		// 创建并初始化 API 服务器
		server := rest.NewServer(&cfg.Server, GetRegistry(), logger, Version)
		if err := server.Setup(); err != nil {
			return fmt.Errorf("初始化服务器失败: %w", err)
		}

		// 启动服务器（非阻塞）
		if err := server.Start(); err != nil {
			return fmt.Errorf("启动服务器失败: %w", err)
		}

		GetSugar().Infof("API 服务器已启动，监听地址: %s:%d", cfg.Server.Host, cfg.Server.Port)

		// 等待中断信号进行优雅关闭
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		GetSugar().Info("收到关闭信号，正在关闭服务器...")

		// 优雅关闭
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("服务器关闭失败: %w", err)
		}

		GetSugar().Info("服务器已关闭")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
