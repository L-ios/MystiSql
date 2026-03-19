package rest

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"MystiSql/internal/service/audit"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuditHandlers struct {
	auditService *audit.AuditService
	logger       *zap.Logger
	logFilePath  string
}

func NewAuditHandlers(auditService *audit.AuditService, logFilePath string, logger *zap.Logger) *AuditHandlers {
	return &AuditHandlers{
		auditService: auditService,
		logFilePath:  logFilePath,
		logger:       logger,
	}
}

func (h *AuditHandlers) QueryLogs(c *gin.Context) {
	startTimeStr := c.Query("start")
	if startTimeStr == "" {
		startTimeStr = c.Query("start_time")
	}
	endTimeStr := c.Query("end")
	if endTimeStr == "" {
		endTimeStr = c.Query("end_time")
	}
	userID := c.Query("user_id")
	instance := c.Query("instance")
	queryType := c.Query("query_type")
	sensitiveStr := c.Query("sensitive")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "100")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 1000 {
		pageSize = 100
	}

	var startTime, endTime *time.Time
	if startTimeStr != "" {
		t, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			startTime = &t
		}
	}

	if endTimeStr != "" {
		t, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			endTime = &t
		}
	}

	var sensitive *bool
	if sensitiveStr != "" {
		s, err := strconv.ParseBool(sensitiveStr)
		if err == nil {
			sensitive = &s
		}
	}

	logs, total, err := h.readLogs(startTime, endTime, userID, instance, queryType, sensitive, page, pageSize)
	if err != nil {
		h.logger.Error("Failed to read audit logs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"QUERY_FAILED",
			fmt.Sprintf("Failed to query audit logs: %v", err),
		))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":       logs,
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": (total + pageSize - 1) / pageSize,
		},
	})
}

func (h *AuditHandlers) readLogs(startTime, endTime *time.Time, userID, instance, queryType string, sensitive *bool, page, pageSize int) ([]*audit.AuditLog, int, error) {
	file, err := os.Open(h.logFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*audit.AuditLog{}, 0, nil
		}
		return nil, 0, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	var allLogs []*audit.AuditLog
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var log audit.AuditLog
		if err := json.Unmarshal(line, &log); err != nil {
			continue
		}

		if startTime != nil && log.Timestamp.Before(*startTime) {
			continue
		}

		if endTime != nil && log.Timestamp.After(*endTime) {
			continue
		}

		if userID != "" && log.UserID != userID {
			continue
		}

		if instance != "" && log.Instance != instance {
			continue
		}

		if queryType != "" && log.QueryType != queryType {
			continue
		}

		if sensitive != nil && log.Sensitive != *sensitive {
			continue
		}

		allLogs = append(allLogs, &log)
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to scan log file: %w", err)
	}

	total := len(allLogs)

	start := (page - 1) * pageSize
	if start >= total {
		return []*audit.AuditLog{}, total, nil
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return allLogs[start:end], total, nil
}

// GetStats 获取审计日志统计信息
func (h *AuditHandlers) GetStats(c *gin.Context) {
	startTimeStr := c.Query("start")
	endTimeStr := c.Query("end")

	var startTime, endTime *time.Time
	if startTimeStr != "" {
		t, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			startTime = &t
		}
	}

	if endTimeStr != "" {
		t, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			endTime = &t
		}
	}

	// 读取所有日志进行统计（限制时间范围内）
	logs, _, err := h.readLogs(startTime, endTime, "", "", "", nil, 1, 10000)
	if err != nil {
		h.logger.Error("Failed to read audit logs for stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, NewErrorResponse(
			"STATS_FAILED",
			fmt.Sprintf("Failed to get audit stats: %v", err),
		))
		return
	}

	stats := h.calculateStats(logs)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

func (h *AuditHandlers) calculateStats(logs []*audit.AuditLog) *audit.AuditStats {
	stats := &audit.AuditStats{
		QueryTypeDist: make(map[string]int64),
		TopUsers:      []audit.UserStats{},
		TopInstances:  []audit.InstanceStats{},
	}

	if len(logs) == 0 {
		return stats
	}

	var totalExecTime int64
	userCounts := make(map[string]int64)
	instanceCounts := make(map[string]int64)
	sensitiveCount := int64(0)
	errorCount := int64(0)

	for _, log := range logs {
		stats.TotalQueries++
		totalExecTime += log.ExecutionTime

		if log.Status == "success" {
			stats.SuccessCount++
		} else if log.Status == "error" {
			errorCount++
		}

		if log.Sensitive {
			sensitiveCount++
		}

		if log.QueryType != "" {
			stats.QueryTypeDist[log.QueryType]++
		}

		if log.UserID != "" {
			userCounts[log.UserID]++
		}

		if log.Instance != "" {
			instanceCounts[log.Instance]++
		}
	}

	stats.ErrorCount = errorCount
	stats.SensitiveCount = sensitiveCount
	stats.AvgExecutionTime = totalExecTime / int64(len(logs))

	// 计算Top用户（最多10个）
	for userID, count := range userCounts {
		stats.TopUsers = append(stats.TopUsers, audit.UserStats{
			UserID: userID,
			Count:  count,
		})
	}
	// 按计数排序
	for i := 0; i < len(stats.TopUsers); i++ {
		for j := i + 1; j < len(stats.TopUsers); j++ {
			if stats.TopUsers[j].Count > stats.TopUsers[i].Count {
				stats.TopUsers[i], stats.TopUsers[j] = stats.TopUsers[j], stats.TopUsers[i]
			}
		}
	}
	if len(stats.TopUsers) > 10 {
		stats.TopUsers = stats.TopUsers[:10]
	}

	// 计算Top实例（最多10个）
	for instance, count := range instanceCounts {
		stats.TopInstances = append(stats.TopInstances, audit.InstanceStats{
			Instance: instance,
			Count:    count,
		})
	}
	// 按计数排序
	for i := 0; i < len(stats.TopInstances); i++ {
		for j := i + 1; j < len(stats.TopInstances); j++ {
			if stats.TopInstances[j].Count > stats.TopInstances[i].Count {
				stats.TopInstances[i], stats.TopInstances[j] = stats.TopInstances[j], stats.TopInstances[i]
			}
		}
	}
	if len(stats.TopInstances) > 10 {
		stats.TopInstances = stats.TopInstances[:10]
	}

	return stats
}
