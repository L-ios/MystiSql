package rest

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"MystiSql/internal/api/middleware"
	"MystiSql/internal/discovery"
	"MystiSql/internal/service/auth"
	"MystiSql/internal/service/query"
	"MystiSql/pkg/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server REST API 服务器
type Server struct {
	config       *types.ServerConfig
	registry     discovery.InstanceRegistry
	engine       *query.Engine
	authService  *auth.AuthService
	logger       *zap.Logger
	server       *http.Server
	router       *gin.Engine
	handlers     *Handlers
	authHandlers *AuthHandlers
	version      string
}

// NewServer 创建新的 REST API 服务器
func NewServer(config *types.ServerConfig, registry discovery.InstanceRegistry, engine *query.Engine, authService *auth.AuthService, logger *zap.Logger, version string) *Server {
	return &Server{
		config:      config,
		registry:    registry,
		engine:      engine,
		authService: authService,
		logger:      logger,
		version:     version,
	}
}

// Setup 初始化服务器
// 配置 Gin 模式、路由和中间件
func (s *Server) Setup() error {
	// 设置 Gin 模式
	switch s.config.Mode {
	case "release":
		gin.SetMode(gin.ReleaseMode)
	case "debug":
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由器
	s.router = gin.New()

	// 创建处理器
	s.handlers = NewHandlers(s.registry, s.engine, s.logger, s.version)
	if s.authService != nil {
		s.authHandlers = NewAuthHandlers(s.authService, s.logger)
	}

	// 添加中间件（顺序很重要）
	s.router.Use(RecoveryMiddleware(s.logger)) // 错误恢复（最外层）
	s.router.Use(LoggerMiddleware(s.logger))   // 日志记录
	s.router.Use(CORSMiddleware())             // CORS 支持

	// 设置路由
	s.setupRoutes()

	// 创建 HTTP 服务器
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("REST API server initialized",
		zap.String("address", addr),
		zap.String("mode", s.config.Mode),
	)

	return nil
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查端点（不需要版本前缀）
	s.router.GET("/health", s.handlers.Health)

	// 认证中间件（如果启用）
	var authMiddleware gin.HandlerFunc
	if s.authService != nil {
		authMiddleware = middleware.AuthMiddleware(s.authService, s.logger)
	}

	// API v1 路由组
	v1 := s.router.Group("/api/v1")
	{
		// 认证相关端点（不需要认证）
		if s.authHandlers != nil {
			auth := v1.Group("/auth")
			{
				auth.POST("/token", s.authHandlers.GenerateToken)
				auth.DELETE("/token", s.authHandlers.RevokeToken)
				auth.GET("/tokens", s.authHandlers.ListTokens)
				auth.GET("/token/info", s.authHandlers.GetTokenInfo)
			}
		}

		// 需要认证的端点
		if authMiddleware != nil {
			v1.Use(authMiddleware)
		}

		// 实例相关端点
		v1.GET("/instances", s.handlers.ListInstances)
		v1.GET("/instances/:name/health", s.handlers.GetInstanceHealth)
		v1.GET("/instances/:name/pool", s.handlers.GetPoolStats)

		// 查询端点
		v1.POST("/query", s.handlers.Query)
		v1.POST("/exec", s.handlers.Exec)
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	s.logger.Info("Starting REST API server",
		zap.String("address", s.server.Addr),
	)

	// 启动服务器（非阻塞）
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	s.logger.Info("REST API server started successfully")
	return nil
}

// Shutdown 优雅关闭服务器
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down REST API server...")

	// 给予 30 秒的优雅关闭时间
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Server shutdown error", zap.Error(err))
		return err
	}

	s.logger.Info("REST API server shutdown successfully")
	return nil
}

// Run 启动服务器并处理优雅关闭
// 这是一个阻塞方法，会一直运行直到收到终止信号
func (s *Server) Run() error {
	// 初始化服务器
	if err := s.Setup(); err != nil {
		return fmt.Errorf("failed to setup server: %w", err)
	}

	// 启动服务器
	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// 等待中断信号进行优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Received shutdown signal")

	// 优雅关闭
	if err := s.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}

// GetRouter 获取路由器（用于测试）
func (s *Server) GetRouter() *gin.Engine {
	return s.router
}
