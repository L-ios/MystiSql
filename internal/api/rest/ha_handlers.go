package rest

import (
	"fmt"
	"net/http"
	"time"

	"MystiSql/internal/service/health"
	"MystiSql/internal/service/topology"
	"MystiSql/pkg/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HAHandlers struct {
	healthChecker   *health.EnhancedHealthChecker
	topologyManager *topology.TopologyManager
	logger          *zap.Logger
}

func NewHAHandlers(healthChecker *health.EnhancedHealthChecker, topologyManager *topology.TopologyManager, logger *zap.Logger) *HAHandlers {
	return &HAHandlers{
		healthChecker:   healthChecker,
		topologyManager: topologyManager,
		logger:          logger,
	}
}

// GetInstanceHealthStatus 获取单个实例健康状态
// GET /api/v1/instances/:name/health
func (h *HAHandlers) GetInstanceHealthStatus(c *gin.Context) {
	instanceName := c.Param("name")

	healthStatus, err := h.healthChecker.GetHealth(instanceName)
	if err != nil {
		h.logger.Error("Failed to get instance health",
			zap.String("instance", instanceName),
			zap.Error(err))
		c.JSON(http.StatusNotFound, NewErrorResponse(
			"INSTANCE_NOT_FOUND",
			fmt.Sprintf("Instance %s not found in health status", instanceName)))
		return
	}

	c.JSON(http.StatusOK, InstanceHealthStatusResponse{
		Name:             healthStatus.Name,
		Status:           string(healthStatus.Status),
		ConsecutiveFails: healthStatus.ConsecutiveFails,
		ConsecutiveOKs:   healthStatus.ConsecutiveOKs,
		LastCheck:        healthStatus.LastCheck,
		LastError:        healthStatus.LastError,
		ResponseTime:     healthStatus.ResponseTime.String(),
	})
}

// GetAllInstancesHealth 获取所有实例健康状态
// GET /api/v1/instances/health
func (h *HAHandlers) GetAllInstancesHealth(c *gin.Context) {
	allHealth := h.healthChecker.GetAllHealth()
	stats := h.healthChecker.GetStats()

	details := make([]InstanceHealthDetailExt, 0, len(allHealth))
	for name, health := range allHealth {
		details = append(details, InstanceHealthDetailExt{
			Name:         name,
			Status:       string(health.Status),
			ResponseTime: health.ResponseTime.String(),
			LastError:    health.LastError,
			LastCheck:    health.LastCheck,
		})
	}

	c.JSON(http.StatusOK, AllInstancesHealthResponse{
		Total:     stats.TotalInstances,
		Healthy:   stats.HealthyInstances,
		Unhealthy: stats.UnhealthyInstances,
		Checking:  stats.CheckingInstances,
		Details:   details,
		Timestamp: time.Now(),
	})
}

// ForceHealthCheck 强制执行健康检查
// POST /api/v1/instances/:name/health/check
func (h *HAHandlers) ForceHealthCheck(c *gin.Context) {
	instanceName := c.Param("name")

	healthStatus, err := h.healthChecker.ForceCheck(instanceName)
	if err != nil {
		h.logger.Error("Failed to force health check",
			zap.String("instance", instanceName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"HEALTH_CHECK_FAILED",
			fmt.Sprintf("Failed to force health check: %v", err)))
		return
	}

	c.JSON(http.StatusOK, InstanceHealthStatusResponse{
		Name:             healthStatus.Name,
		Status:           string(healthStatus.Status),
		ConsecutiveFails: healthStatus.ConsecutiveFails,
		ConsecutiveOKs:   healthStatus.ConsecutiveOKs,
		LastCheck:        healthStatus.LastCheck,
		LastError:        healthStatus.LastError,
		ResponseTime:     healthStatus.ResponseTime.String(),
	})
}

// GetHealthyInstances 获取健康实例列表
// GET /api/v1/instances/healthy
func (h *HAHandlers) GetHealthyInstances(c *gin.Context) {
	healthy := h.healthChecker.GetHealthyInstances()

	c.JSON(http.StatusOK, gin.H{
		"instances": healthy,
		"count":     len(healthy),
	})
}

// GetUnhealthyInstances 获取不健康实例列表
// GET /api/v1/instances/unhealthy
func (h *HAHandlers) GetUnhealthyInstances(c *gin.Context) {
	unhealthy := h.healthChecker.GetUnhealthyInstances()

	c.JSON(http.StatusOK, gin.H{
		"instances": unhealthy,
		"count":     len(unhealthy),
	})
}

// GetHAStatus 获取高可用整体状态
// GET /api/v1/ha/status
func (h *HAHandlers) GetHAStatus(c *gin.Context) {
	stats := h.healthChecker.GetStats()
	topologyStatus := h.topologyManager.GetAllTopologyStatus()

	c.JSON(http.StatusOK, HAStatusResponse{
		Enabled:    true,
		Health:     stats,
		Topologies: topologyStatus,
		Timestamp:  time.Now(),
	})
}

// GetTopology 获取单个拓扑状态
// GET /api/v1/ha/topology/:name
func (h *HAHandlers) GetTopology(c *gin.Context) {
	topologyName := c.Param("name")

	status, err := h.topologyManager.GetTopologyStatus(topologyName)
	if err != nil {
		h.logger.Error("Failed to get topology status",
			zap.String("topology", topologyName),
			zap.Error(err))
		c.JSON(http.StatusNotFound, NewErrorResponse(
			"TOPOLOGY_NOT_FOUND",
			fmt.Sprintf("Topology %s not found", topologyName)))
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetAllTopologies 获取所有拓扑状态
// GET /api/v1/ha/topology
func (h *HAHandlers) GetAllTopologies(c *gin.Context) {
	status := h.topologyManager.GetAllTopologyStatus()

	c.JSON(http.StatusOK, gin.H{
		"topologies": status,
		"count":      len(status),
	})
}

// FailoverRequest 故障转移请求
type FailoverRequest struct {
	SlaveName string `json:"slaveName" binding:"required"`
	Force     bool   `json:"force"`
}

// ManualFailover 手动故障转移
// POST /api/v1/ha/failover
func (h *HAHandlers) ManualFailover(c *gin.Context) {
	var req FailoverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			"INVALID_REQUEST",
			fmt.Sprintf("Invalid request body: %v", err)))
		return
	}

	topologyName := c.Query("topology")
	if topologyName == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			"MISSING_TOPOLOGY",
			"topology query parameter is required"))
		return
	}

	err := h.topologyManager.PromoteSlave(topologyName, req.SlaveName, req.Force)
	if err != nil {
		h.logger.Error("Manual failover failed",
			zap.String("topology", topologyName),
			zap.String("slave", req.SlaveName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"FAILOVER_FAILED",
			fmt.Sprintf("Failed to promote slave: %v", err)))
		return
	}

	h.logger.Info("Manual failover completed",
		zap.String("topology", topologyName),
		zap.String("new_master", req.SlaveName))

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   fmt.Sprintf("Slave %s promoted to master", req.SlaveName),
		"topology":  topologyName,
		"newMaster": req.SlaveName,
		"timestamp": time.Now(),
	})
}

// AutoFailover 自动故障转移
// POST /api/v1/ha/failover/auto
func (h *HAHandlers) AutoFailover(c *gin.Context) {
	topologyName := c.Query("topology")
	if topologyName == "" {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			"MISSING_TOPOLOGY",
			"topology query parameter is required"))
		return
	}

	err := h.topologyManager.ForceFailover(topologyName)
	if err != nil {
		h.logger.Error("Auto failover failed",
			zap.String("topology", topologyName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"AUTO_FAILOVER_FAILED",
			fmt.Sprintf("Auto failover failed: %v", err)))
		return
	}

	h.logger.Info("Auto failover completed",
		zap.String("topology", topologyName))

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "Auto failover completed",
		"topology":  topologyName,
		"timestamp": time.Now(),
	})
}

// AddSlaveRequest 添加从库请求
type AddSlaveRequest struct {
	Name     string            `json:"name" binding:"required"`
	Host     string            `json:"host" binding:"required"`
	Port     int               `json:"port" binding:"required"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	Database string            `json:"database"`
	Weight   int               `json:"weight"`
	Labels   map[string]string `json:"labels"`
}

// AddSlave 添加从库
// POST /api/v1/ha/topology/:name/slaves
func (h *HAHandlers) AddSlave(c *gin.Context) {
	topologyName := c.Param("name")

	var req AddSlaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(
			"INVALID_REQUEST",
			fmt.Sprintf("Invalid request body: %v", err)))
		return
	}

	topology, err := h.topologyManager.GetTopology(topologyName)
	if err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse(
			"TOPOLOGY_NOT_FOUND",
			fmt.Sprintf("Topology %s not found", topologyName)))
		return
	}

	slave := &types.DatabaseInstance{
		Name:      req.Name,
		Type:      topology.Master.Type,
		Host:      req.Host,
		Port:      req.Port,
		Username:  req.Username,
		Password:  req.Password,
		Database:  req.Database,
		Weight:    req.Weight,
		Labels:    req.Labels,
		Role:      types.InstanceRoleSlave,
		Master:    topology.Master.Name,
		Status:    types.InstanceStatusUnknown,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = h.topologyManager.AddSlave(topologyName, slave)
	if err != nil {
		h.logger.Error("Failed to add slave",
			zap.String("topology", topologyName),
			zap.String("slave", req.Name),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"ADD_SLAVE_FAILED",
			fmt.Sprintf("Failed to add slave: %v", err)))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": fmt.Sprintf("Slave %s added to topology %s", req.Name, topologyName),
		"slave":   slave,
	})
}

// RemoveSlave 移除从库
// DELETE /api/v1/ha/topology/:name/slaves/:slaveName
func (h *HAHandlers) RemoveSlave(c *gin.Context) {
	topologyName := c.Param("name")
	slaveName := c.Param("slaveName")

	err := h.topologyManager.RemoveSlave(topologyName, slaveName)
	if err != nil {
		h.logger.Error("Failed to remove slave",
			zap.String("topology", topologyName),
			zap.String("slave", slaveName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"REMOVE_SLAVE_FAILED",
			fmt.Sprintf("Failed to remove slave: %v", err)))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Slave %s removed from topology %s", slaveName, topologyName),
	})
}

type InstanceHealthStatusResponse struct {
	Name             string    `json:"name"`
	Status           string    `json:"status"`
	ConsecutiveFails int       `json:"consecutiveFails"`
	ConsecutiveOKs   int       `json:"consecutiveOKs"`
	LastCheck        time.Time `json:"lastCheck"`
	LastError        string    `json:"lastError,omitempty"`
	ResponseTime     string    `json:"responseTime"`
}

type AllInstancesHealthResponse struct {
	Total     int                       `json:"total"`
	Healthy   int                       `json:"healthy"`
	Unhealthy int                       `json:"unhealthy"`
	Checking  int                       `json:"checking"`
	Details   []InstanceHealthDetailExt `json:"details"`
	Timestamp time.Time                 `json:"timestamp"`
}

type InstanceHealthDetailExt struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	ResponseTime string    `json:"responseTime,omitempty"`
	LastError    string    `json:"lastError,omitempty"`
	LastCheck    time.Time `json:"lastCheck"`
}

type HAStatusResponse struct {
	Enabled    bool                                    `json:"enabled"`
	Health     health.HealthCheckStats                 `json:"health"`
	Topologies map[string]*topology.TopologyStatusInfo `json:"topologies"`
	Timestamp  time.Time                               `json:"timestamp"`
}

// RegisterHARoutes 注册高可用相关路由
func RegisterHARoutes(router *gin.RouterGroup, haHandlers *HAHandlers) {
	instances := router.Group("/instances")
	{
		instances.GET("/health", haHandlers.GetAllInstancesHealth)
		instances.GET("/healthy", haHandlers.GetHealthyInstances)
		instances.GET("/unhealthy", haHandlers.GetUnhealthyInstances)
		instances.GET("/:name/health", haHandlers.GetInstanceHealthStatus)
		instances.POST("/:name/health/check", haHandlers.ForceHealthCheck)
	}

	ha := router.Group("/ha")
	{
		ha.GET("/status", haHandlers.GetHAStatus)
		ha.GET("/topology", haHandlers.GetAllTopologies)
		ha.GET("/topology/:name", haHandlers.GetTopology)
		ha.POST("/failover", haHandlers.ManualFailover)
		ha.POST("/failover/auto", haHandlers.AutoFailover)
		ha.POST("/topology/:name/slaves", haHandlers.AddSlave)
		ha.DELETE("/topology/:name/slaves/:slaveName", haHandlers.RemoveSlave)
	}
}
