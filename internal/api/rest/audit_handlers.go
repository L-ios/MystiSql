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

	logs, total, err := h.readLogs(startTime, endTime, userID, instance, page, pageSize)
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

func (h *AuditHandlers) readLogs(startTime, endTime *time.Time, userID, instance string, page, pageSize int) ([]*audit.AuditLog, int, error) {
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
