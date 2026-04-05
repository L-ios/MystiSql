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
	wsapi "MystiSql/internal/api/websocket"
	"MystiSql/internal/connection"
	"MystiSql/internal/connection/mysql"
	"MystiSql/internal/connection/pool"
	"MystiSql/internal/connection/postgresql"
	"MystiSql/internal/discovery"
	"MystiSql/internal/service/audit"
	"MystiSql/internal/service/auth"
	"MystiSql/internal/service/batch"
	"MystiSql/internal/service/query"
	"MystiSql/internal/service/rbac"
	"MystiSql/internal/service/transaction"
	"MystiSql/internal/service/validator"
	"MystiSql/pkg/types"
	webui "MystiSql/web"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server REST API 服务器
type Server struct {
	config              *types.ServerConfig
	websocketConfig     *types.WebSocketConfig
	webuiConfig         *types.WebUIConfig
	corsOrigins         []string
	registry            discovery.InstanceRegistry
	engine              *query.Engine
	authService         *auth.AuthService
	validatorService    *validator.ValidatorService
	auditService        *audit.AuditService
	auditLogFile        string
	poolManager         *pool.ConnectionPoolManager
	txManager           *transaction.TransactionManager
	batchService        *batch.BatchService
	logger              *zap.Logger
	server              *http.Server
	router              *gin.Engine
	handlers            *Handlers
	authHandlers        *AuthHandlers
	transactionHandlers *TransactionHandlers
	batchHandlers       *BatchHandlers
	validatorHandlers   *ValidatorHandlers
	auditHandlers       *AuditHandlers
	rbacService         *rbac.RBACService
	rbacHandlers        *RBACHandlers
	version             string
	wsHandler           *wsapi.WebSocketHandler
	webuiHandler        *webui.Handler
}

// NewServer 创建新的 REST API 服务器
func NewServer(config *types.ServerConfig, websocketConfig *types.WebSocketConfig, webuiConfig *types.WebUIConfig, registry discovery.InstanceRegistry, engine *query.Engine, authService *auth.AuthService, validatorService *validator.ValidatorService, auditService *audit.AuditService, auditLogFile string, logger *zap.Logger, version string) *Server {
	return &Server{
		config:           config,
		websocketConfig:  websocketConfig,
		webuiConfig:      webuiConfig,
		registry:         registry,
		engine:           engine,
		authService:      authService,
		validatorService: validatorService,
		auditService:     auditService,
		auditLogFile:     auditLogFile,
		logger:           logger,
		version:          version,
	}
}

// SetCORSOrigins 设置允许的 CORS 来源列表
func (s *Server) SetCORSOrigins(origins []string) {
	s.corsOrigins = origins
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
	s.handlers = NewHandlers(s.registry, s.engine, s.validatorService, s.logger, s.version)
	if s.authService != nil {
		s.authHandlers = NewAuthHandlers(s.authService, s.logger)
	}

	// 初始化 WebSocket 处理器
	if s.wsHandler == nil && s.authService != nil && s.engine != nil {
		cfg := wsapi.DefaultConfig()
		if s.websocketConfig != nil {
			cfg.MaxConnections = s.websocketConfig.MaxConnections
			if idleTimeout, err := time.ParseDuration(s.websocketConfig.IdleTimeout); err == nil {
				cfg.IdleTimeout = idleTimeout
			}
			cfg.AllowedOrigins = s.websocketConfig.AllowedOrigins
			if s.websocketConfig.MaxMessageSize > 0 {
				cfg.MaxMessageSize = s.websocketConfig.MaxMessageSize
			}
		}
		s.wsHandler = wsapi.NewWebSocketHandler(s.engine, s.authService, s.logger, cfg)
	}

	// 初始化验证器处理器
	if s.validatorService != nil {
		s.validatorHandlers = NewValidatorHandlers(s.validatorService, s.logger)
	}

	// 初始化审计处理器
	if s.auditService != nil {
		s.auditHandlers = NewAuditHandlers(s.auditService, s.auditLogFile, s.logger)
	}

	// 初始化 WebUI 处理器
	if s.webuiConfig != nil && s.webuiConfig.Enabled && s.webuiConfig.Mode == "embedded" {
		webuiHandler, err := webui.NewHandler()
		if err != nil {
			s.logger.Warn("Failed to initialize WebUI handler", zap.Error(err))
		} else {
			s.webuiHandler = webuiHandler
			s.logger.Info("WebUI handler initialized")
		}
	}

	// 初始化 RBAC
	if s.rbacService == nil {
		s.rbacService = rbac.NewRBACService()
	}
	s.rbacHandlers = NewRBACHandlers(s.rbacService, s.logger)

	// 初始化 ConnectionPoolManager 和 TransactionManager
	if err := s.initTransactionManager(); err != nil {
		s.logger.Warn("Failed to initialize transaction manager", zap.Error(err))
	}

	// 设置 handlers 的事务管理器
	if s.txManager != nil && s.handlers != nil {
		s.handlers.SetTransactionManager(s.txManager)
	}

	// 添加中间件（顺序很重要）
	s.router.Use(RecoveryMiddleware(s.logger))  // 错误恢复（最外层）
	s.router.Use(SecurityHeadersMiddleware())   // 安全响应头
	s.router.Use(LoggerMiddleware(s.logger))    // 日志记录
	s.router.Use(CORSMiddleware(s.corsOrigins)) // CORS 支持

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
				auth.POST("/token/info", s.authHandlers.GetTokenInfo)
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

		// 事务管理端点
		if s.transactionHandlers != nil {
			transaction := v1.Group("/transaction")
			{
				transaction.POST("/begin", s.transactionHandlers.BeginTransaction)
				transaction.POST("/commit", s.transactionHandlers.CommitTransaction)
				transaction.POST("/rollback", s.transactionHandlers.RollbackTransaction)
				transaction.GET("/:id", s.transactionHandlers.GetTransaction)
				transaction.GET("", s.transactionHandlers.ListTransactions)
				transaction.POST("/:id/extend", s.transactionHandlers.ExtendTransaction)
			}
		}

		// 批量操作端点
		if s.batchHandlers != nil {
			v1.POST("/batch", s.batchHandlers.ExecuteBatch)
		}

		// 添加 validator 白名单/黑名单 API
		if s.validatorHandlers != nil {
			validatorGroup := v1.Group("/validator")
			{
				validatorGroup.GET("/whitelist", s.validatorHandlers.GetWhitelist)
				validatorGroup.POST("/whitelist", s.validatorHandlers.UpdateWhitelist)
				validatorGroup.GET("/blacklist", s.validatorHandlers.GetBlacklist)
				validatorGroup.POST("/blacklist", s.validatorHandlers.UpdateBlacklist)
			}
		}

		// 添加审计日志 API
		if s.auditHandlers != nil {
			auditGroup := v1.Group("/audit")
			{
				auditGroup.GET("/logs", s.auditHandlers.QueryLogs)
				auditGroup.GET("/stats", s.auditHandlers.GetStats)
			}
		}

		// RBAC 路由
		if s.rbacHandlers != nil {
			rbacGroup := v1.Group("/rbac")
			{
				rbacGroup.POST("/roles", s.rbacHandlers.CreateRole)
				rbacGroup.GET("/roles", s.rbacHandlers.ListRoles)
				rbacGroup.DELETE("/roles/:name", s.rbacHandlers.DeleteRole)
				rbacGroup.GET("/roles/:name", s.rbacHandlers.GetRole)
				rbacGroup.POST("/users/:id/roles", s.rbacHandlers.AssignRoleToUser)
				rbacGroup.GET("/users/:id/roles", s.rbacHandlers.ListUserRoles)
			}
		}
	}

	// WebSocket 端点（独立于 API v1）
	if s.wsHandler != nil {
		s.router.GET("/ws", s.wsHandler.Handle)
	}

	// WebUI 端点（必须在所有 API 路由之后）
	if s.webuiHandler != nil {
		s.router.NoRoute(gin.WrapH(s.webuiHandler))
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

	if s.wsHandler != nil {
		if err := s.wsHandler.Close(); err != nil {
			s.logger.Warn("WebSocket handler close error", zap.Error(err))
		}
	}

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

// initTransactionManager 初始化事务管理器
func (s *Server) initTransactionManager() error {
	// 创建多数据库类型工厂
	factory := newMultiDatabaseFactory()

	// 创建连接池管理器
	s.poolManager = pool.NewConnectionPoolManager(factory, nil)

	// 从注册中心加载所有实例并添加到连接池管理器
	instances, err := s.registry.ListInstances()
	if err != nil {
		s.logger.Error("Failed to list instances", zap.Error(err))
		return err
	}

	for _, instance := range instances {
		if err := s.poolManager.AddInstance(instance); err != nil {
			s.logger.Warn("Failed to add instance to pool manager",
				zap.String("instance", instance.Name),
				zap.Error(err),
			)
		} else {
			s.logger.Debug("Added instance to pool manager",
				zap.String("instance", instance.Name),
			)
		}
	}

	// 创建事务管理器
	s.txManager = transaction.NewTransactionManager(s.poolManager, s.logger, nil)

	// 创建批量操作服务
	s.batchService = batch.NewBatchService(s.txManager, nil, s.logger)

	// 创建事务处理器
	s.transactionHandlers = NewTransactionHandlers(s.txManager, s.logger)

	// 创建批量操作处理器
	s.batchHandlers = NewBatchHandlers(s.batchService, s.logger)

	s.logger.Info("Transaction manager initialized")
	return nil
}

// multiDatabaseFactory 多数据库类型工厂
type multiDatabaseFactory struct {
	mysqlFactory      connection.ConnectionFactory
	postgresqlFactory connection.ConnectionFactory
}

func newMultiDatabaseFactory() *multiDatabaseFactory {
	return &multiDatabaseFactory{
		mysqlFactory:      mysql.NewFactory(),
		postgresqlFactory: postgresql.NewFactory(),
	}
}

func (f *multiDatabaseFactory) CreateConnection(instance *types.DatabaseInstance) (connection.Connection, error) {
	switch instance.Type {
	case types.DatabaseTypeMySQL:
		return f.mysqlFactory.CreateConnection(instance)
	case types.DatabaseTypePostgreSQL:
		return f.postgresqlFactory.CreateConnection(instance)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", instance.Type)
	}
}
