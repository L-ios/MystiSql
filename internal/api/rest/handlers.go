package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"MystiSql/internal/discovery"
	"MystiSql/internal/service/query"
	"MystiSql/internal/service/transaction"
	"MystiSql/pkg/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handlers API 处理器
type Handlers struct {
	registry  discovery.InstanceRegistry
	engine    *query.Engine
	txManager *transaction.TransactionManager
	logger    *zap.Logger
	version   string
}

// NewHandlers 创建新的处理器
func NewHandlers(registry discovery.InstanceRegistry, engine *query.Engine, logger *zap.Logger, version string) *Handlers {
	return &Handlers{
		registry: registry,
		engine:   engine,
		logger:   logger,
		version:  version,
	}
}

// SetTransactionManager 设置事务管理器
func (h *Handlers) SetTransactionManager(txManager *transaction.TransactionManager) {
	h.txManager = txManager
}

// Health 健康检查端点
// GET /health
// 查询参数：
//   - check-instances: 是否检查实例健康状态（true/false）
func (h *Handlers) Health(c *gin.Context) {
	// 获取查询参数
	checkInstances := c.Query("check-instances") == "true"

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   h.version,
	}

	// 如果需要检查实例健康状态
	if checkInstances {
		instancesHealth := h.checkInstancesHealth(c.Request.Context())
		response.Instances = instancesHealth

		// 如果有实例且全部不健康，则整体状态为 unhealthy
		if instancesHealth.Total > 0 && instancesHealth.Healthy == 0 {
			response.Status = "unhealthy"
		}
	}

	c.JSON(http.StatusOK, response)
}

// checkInstancesHealth 检查所有实例的健康状态
func (h *Handlers) checkInstancesHealth(ctx context.Context) *InstancesHealth {
	instances, err := h.registry.ListInstances()
	if err != nil {
		h.logger.Error("Failed to list instances", zap.Error(err))
		return &InstancesHealth{
			Total:     0,
			Healthy:   0,
			Unhealthy: 0,
			Details:   []InstanceHealthDetail{},
		}
	}

	health := &InstancesHealth{
		Total:     len(instances),
		Healthy:   0,
		Unhealthy: 0,
		Details:   make([]InstanceHealthDetail, 0, len(instances)),
	}

	// 检查每个实例
	for _, instance := range instances {
		detail := InstanceHealthDetail{
			Name:   instance.Name,
			Type:   string(instance.Type),
			Status: string(instance.Status),
		}

		// 如果实例状态未知，尝试 ping 检查
		if instance.Status == types.InstanceStatusUnknown {
			status := h.pingInstance(ctx, instance)
			detail.Status = string(status)
		}

		health.Details = append(health.Details, detail)

		// 统计健康和不健康数量
		if detail.Status == string(types.InstanceStatusHealthy) {
			health.Healthy++
		} else {
			health.Unhealthy++
		}
	}

	return health
}

// pingInstance 通过 query engine 检查实例健康状态
func (h *Handlers) pingInstance(ctx context.Context, instance *types.DatabaseInstance) types.InstanceStatus {
	// 使用 query engine 进行健康检查
	status, err := h.engine.GetInstanceHealth(ctx, instance.Name)
	if err != nil {
		h.logger.Warn("Failed to check instance health",
			zap.String("instance", instance.Name),
			zap.Error(err),
		)
		return types.InstanceStatusUnhealthy
	}

	return status
}

// ListInstances 列出所有实例
// GET /api/v1/instances
func (h *Handlers) ListInstances(c *gin.Context) {
	// 获取所有实例
	instances, err := h.registry.ListInstances()
	if err != nil {
		h.logger.Error("Failed to list instances", zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"INTERNAL_ERROR",
			"Failed to list instances",
		))
		return
	}

	// 转换为响应格式（脱敏密码）
	instanceResponses := make([]InstanceResponse, 0, len(instances))
	for _, instance := range instances {
		instanceResponses = append(instanceResponses, ToInstanceResponse(instance))
	}

	response := InstancesListResponse{
		Total:     len(instanceResponses),
		Instances: instanceResponses,
	}

	c.JSON(http.StatusOK, response)
}

// Query 执行查询
// POST /api/v1/query
func (h *Handlers) Query(c *gin.Context) {
	// 解析请求
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			"INVALID_REQUEST",
			fmt.Sprintf("Invalid request: %v", err),
		))
		return
	}

	// 设置超时
	timeout := 30 * time.Second // 默认 30 秒
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	// 执行查询
	start := time.Now()
	var result *types.QueryResult
	var err error

	// 如果提供了事务 ID，使用事务连接执行查询
	if req.TransactionID != "" && h.txManager != nil {
		tx, txErr := h.txManager.GetTransaction(req.TransactionID)
		if txErr != nil {
			h.logger.Error("Failed to get transaction",
				zap.String("transaction_id", req.TransactionID),
				zap.Error(txErr),
			)
			c.JSON(http.StatusBadRequest, NewErrorResponse(
				"TRANSACTION_ERROR",
				fmt.Sprintf("Failed to get transaction: %v", txErr),
			))
			return
		}

		result, err = tx.Connection.Query(ctx, req.SQL)
	} else {
		result, err = h.engine.ExecuteQuery(ctx, req.Instance, req.SQL)
	}

	execTime := time.Since(start)

	if err != nil {
		h.logger.Error("Query failed",
			zap.String("instance", req.Instance),
			zap.String("sql", req.SQL),
			zap.String("transaction_id", req.TransactionID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, &QueryResponse{
			Success:       false,
			ExecutionTime: execTime,
			Error: &ErrorDetail{
				Code:    "QUERY_FAILED",
				Message: fmt.Sprintf("Query failed: %v", err),
			},
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, &QueryResponse{
		Success:       true,
		ExecutionTime: execTime,
		Data: &QueryResultData{
			Columns:  result.Columns,
			Rows:     result.Rows,
			RowCount: result.RowCount,
		},
	})
}

// Exec 执行非查询语句
// POST /api/v1/exec
func (h *Handlers) Exec(c *gin.Context) {
	// 解析请求
	var req ExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			"INVALID_REQUEST",
			fmt.Sprintf("Invalid request: %v", err),
		))
		return
	}

	// 设置超时
	timeout := 30 * time.Second // 默认 30 秒
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	// 执行更新
	start := time.Now()
	var result *types.ExecResult
	var err error

	// 如果提供了事务 ID，使用事务连接执行
	if req.TransactionID != "" && h.txManager != nil {
		tx, txErr := h.txManager.GetTransaction(req.TransactionID)
		if txErr != nil {
			h.logger.Error("Failed to get transaction",
				zap.String("transaction_id", req.TransactionID),
				zap.Error(txErr),
			)
			c.JSON(http.StatusBadRequest, NewErrorResponse(
				"TRANSACTION_ERROR",
				fmt.Sprintf("Failed to get transaction: %v", txErr),
			))
			return
		}

		result, err = tx.Connection.Exec(ctx, req.SQL)
	} else {
		result, err = h.engine.ExecuteExec(ctx, req.Instance, req.SQL)
	}

	execTime := time.Since(start)

	if err != nil {
		h.logger.Error("Exec failed",
			zap.String("instance", req.Instance),
			zap.String("sql", req.SQL),
			zap.String("transaction_id", req.TransactionID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, &ExecResponse{
			Success:       false,
			ExecutionTime: execTime,
			Error: &ErrorDetail{
				Code:    "EXEC_FAILED",
				Message: fmt.Sprintf("Exec failed: %v", err),
			},
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, &ExecResponse{
		Success:       true,
		ExecutionTime: execTime,
		Data: &ExecResultData{
			AffectedRows: result.RowsAffected,
			LastInsertID: result.LastInsertID,
		},
	})
}

// GetInstanceHealth 获取实例健康状态
// GET /api/v1/instances/:name/health
func (h *Handlers) GetInstanceHealth(c *gin.Context) {
	instanceName := c.Param("name")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// 获取实例健康状态
	status, err := h.engine.GetInstanceHealth(ctx, instanceName)
	if err != nil {
		h.logger.Error("Failed to get instance health",
			zap.String("instance", instanceName),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"HEALTH_CHECK_FAILED",
			fmt.Sprintf("Failed to get instance health: %v", err),
		))
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, &InstanceHealthResponse{
		Instance:  instanceName,
		Status:    string(status),
		Timestamp: time.Now(),
	})
}

// GetPoolStats 获取连接池统计信息
// GET /api/v1/instances/:name/pool
func (h *Handlers) GetPoolStats(c *gin.Context) {
	instanceName := c.Param("name")

	// 获取连接池统计信息
	stats, err := h.engine.GetPoolStats(instanceName)
	if err != nil {
		h.logger.Error("Failed to get pool stats",
			zap.String("instance", instanceName),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"POOL_STATS_FAILED",
			fmt.Sprintf("Failed to get pool stats: %v", err),
		))
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, &PoolStatsResponse{
		Instance:  instanceName,
		Stats:     *stats,
		Timestamp: time.Now(),
	})
}
