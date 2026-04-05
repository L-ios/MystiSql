package rest

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"MystiSql/internal/service/audit"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupAuditTestRouter(t *testing.T, logFilePath string) (*gin.Engine, *AuditHandlers) {
	logger := zap.NewNop()
	auditConfig := &audit.AuditConfig{
		Enabled: false,
	}
	auditService, err := audit.NewAuditService(auditConfig, logger)
	require.NoError(t, err)

	handlers := NewAuditHandlers(auditService, logFilePath, logger)
	require.NotNil(t, handlers)

	router := gin.New()
	router.GET("/audit/logs", handlers.QueryLogs)
	router.GET("/audit/stats", handlers.GetStats)

	return router, handlers
}

func writeAuditLogs(t *testing.T, filePath string, logs []*audit.AuditLog) {
	t.Helper()

	f, err := os.Create(filePath)
	require.NoError(t, err)
	defer f.Close()

	writer := bufio.NewWriter(f)
	for _, log := range logs {
		data, err := json.Marshal(log)
		require.NoError(t, err)
		fmt.Fprintln(writer, string(data))
	}
	require.NoError(t, writer.Flush())
}

func TestNewAuditHandlers(t *testing.T) {
	logger := zap.NewNop()
	auditConfig := &audit.AuditConfig{Enabled: false}
	auditService, err := audit.NewAuditService(auditConfig, logger)
	require.NoError(t, err)

	handlers := NewAuditHandlers(auditService, "/tmp/test.log", logger)
	assert.NotNil(t, handlers)
	assert.Equal(t, auditService, handlers.auditService)
	assert.Equal(t, "/tmp/test.log", handlers.logFilePath)
}

func TestAuditHandlers_QueryLogs_NoFile(t *testing.T) {
	router, _ := setupAuditTestRouter(t, "/nonexistent/path/audit.log")

	req := httptest.NewRequest(http.MethodGet, "/audit/logs", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	data := resp["data"].(map[string]interface{})
	logs := data["logs"].([]interface{})
	assert.Empty(t, logs)
	assert.Equal(t, float64(0), data["total"])
}

func TestAuditHandlers_QueryLogs_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFilePath := filepath.Join(tmpDir, "audit.log")

	now := time.Now()
	logs := []*audit.AuditLog{
		{
			Timestamp:     now.Add(-2 * time.Hour),
			UserID:        "user1",
			Instance:      "db-prod",
			Query:         "SELECT * FROM users",
			QueryType:     "SELECT",
			ExecutionTime: 50,
			Status:        "success",
		},
		{
			Timestamp:     now.Add(-1 * time.Hour),
			UserID:        "user2",
			Instance:      "db-staging",
			Query:         "DELETE FROM sessions WHERE expired = true",
			QueryType:     "DELETE",
			ExecutionTime: 120,
			Status:        "success",
			Sensitive:     true,
		},
		{
			Timestamp:     now.Add(-30 * time.Minute),
			UserID:        "user1",
			Instance:      "db-prod",
			Query:         "UPDATE users SET active = false",
			QueryType:     "UPDATE",
			ExecutionTime: 200,
			Status:        "error",
		},
	}
	writeAuditLogs(t, logFilePath, logs)

	router, _ := setupAuditTestRouter(t, logFilePath)

	t.Run("all logs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/audit/logs", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(3), data["total"])
	})

	t.Run("filter by time range excludes old", func(t *testing.T) {
		startTime := now.Add(-35 * time.Minute).Format(time.RFC3339)
		endTime := now.Add(1 * time.Minute).Format(time.RFC3339)
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/audit/logs?start=%s&end=%s", url.QueryEscape(startTime), url.QueryEscape(endTime)), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["total"])
	})

	t.Run("pagination page 2", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/audit/logs?page=2&page_size=2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(3), data["total"])
		assert.Equal(t, float64(2), data["page"])
		assert.Equal(t, float64(2), data["totalPages"])
	})

	t.Run("pagination beyond range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/audit/logs?page=100&page_size=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(3), data["total"])
		logs := data["logs"].([]interface{})
		assert.Empty(t, logs)
	})

	t.Run("invalid page params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/audit/logs?page=-1&page_size=abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["page"])
		assert.Equal(t, float64(100), data["pageSize"])
	})
}

func TestAuditHandlers_QueryLogs_TimeFilterAliases(t *testing.T) {
	tmpDir := t.TempDir()
	logFilePath := filepath.Join(tmpDir, "audit.log")

	now := time.Now()
	logs := []*audit.AuditLog{
		{Timestamp: now.Add(-1 * time.Hour), UserID: "u1", Query: "SELECT 1", Status: "success"},
	}
	writeAuditLogs(t, logFilePath, logs)

	router, _ := setupAuditTestRouter(t, logFilePath)

	startTime := now.Add(-2 * time.Hour).Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/audit/logs?start_time=%s", startTime), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["total"])
}

func TestAuditHandlers_QueryLogs_InvalidJSONLine(t *testing.T) {
	tmpDir := t.TempDir()
	logFilePath := filepath.Join(tmpDir, "audit.log")

	f, err := os.Create(logFilePath)
	require.NoError(t, err)
	fmt.Fprintln(f, "not a json line")
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, `{"timestamp":"2024-01-01T00:00:00Z","user_id":"u1","query":"SELECT 1","status":"success"}`)
	f.Close()

	router, _ := setupAuditTestRouter(t, logFilePath)

	req := httptest.NewRequest(http.MethodGet, "/audit/logs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["total"])
}

func TestAuditHandlers_GetStats_NoFile(t *testing.T) {
	router, _ := setupAuditTestRouter(t, "/nonexistent/path/audit.log")

	req := httptest.NewRequest(http.MethodGet, "/audit/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
}

func TestAuditHandlers_GetStats_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFilePath := filepath.Join(tmpDir, "audit.log")

	now := time.Now()
	logs := []*audit.AuditLog{
		{
			Timestamp:     now.Add(-2 * time.Hour),
			UserID:        "user1",
			Instance:      "db-prod",
			Query:         "SELECT * FROM users",
			QueryType:     "SELECT",
			ExecutionTime: 50,
			Status:        "success",
		},
		{
			Timestamp:     now.Add(-1 * time.Hour),
			UserID:        "user2",
			Instance:      "db-staging",
			Query:         "DELETE FROM sessions",
			QueryType:     "DELETE",
			ExecutionTime: 120,
			Status:        "success",
			Sensitive:     true,
		},
		{
			Timestamp:     now.Add(-30 * time.Minute),
			UserID:        "user1",
			Instance:      "db-prod",
			Query:         "UPDATE users SET active = false",
			QueryType:     "UPDATE",
			ExecutionTime: 200,
			Status:        "error",
		},
	}
	writeAuditLogs(t, logFilePath, logs)

	router, _ := setupAuditTestRouter(t, logFilePath)

	req := httptest.NewRequest(http.MethodGet, "/audit/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))

	stats := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(3), stats["totalQueries"])
	assert.Equal(t, float64(2), stats["successCount"])
	assert.Equal(t, float64(1), stats["errorCount"])
	assert.Equal(t, float64(1), stats["sensitiveCount"])
	assert.Equal(t, float64(123), stats["avgExecutionTimeMs"])

	topUsers := stats["topUsers"].([]interface{})
	assert.Len(t, topUsers, 2)

	queryTypeDist := stats["queryTypeDistribution"].(map[string]interface{})
	assert.Equal(t, float64(1), queryTypeDist["SELECT"])
	assert.Equal(t, float64(1), queryTypeDist["DELETE"])
	assert.Equal(t, float64(1), queryTypeDist["UPDATE"])
}

func TestAuditHandlers_GetStats_WithTimeFilter(t *testing.T) {
	tmpDir := t.TempDir()
	logFilePath := filepath.Join(tmpDir, "audit.log")

	now := time.Now()
	logs := []*audit.AuditLog{
		{
			Timestamp:     now.Add(-48 * time.Hour),
			UserID:        "user1",
			Query:         "SELECT 1",
			ExecutionTime: 10,
			Status:        "success",
		},
		{
			Timestamp:     now.Add(-1 * time.Hour),
			UserID:        "user2",
			Query:         "SELECT 2",
			ExecutionTime: 20,
			Status:        "success",
		},
	}
	writeAuditLogs(t, logFilePath, logs)

	router, _ := setupAuditTestRouter(t, logFilePath)

	startTime := now.Add(-50 * time.Hour).Format(time.RFC3339)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/audit/stats?start=%s", url.QueryEscape(startTime)), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	stats := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(2), stats["totalQueries"])
}

func TestAuditHandlers_GetStats_EmptyLogs(t *testing.T) {
	tmpDir := t.TempDir()
	logFilePath := filepath.Join(tmpDir, "audit.log")

	f, err := os.Create(logFilePath)
	require.NoError(t, err)
	f.Close()

	router, _ := setupAuditTestRouter(t, logFilePath)

	req := httptest.NewRequest(http.MethodGet, "/audit/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	stats := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(0), stats["totalQueries"])
}

func TestAuditHandlers_GetStats_ManyUsers_Top10(t *testing.T) {
	tmpDir := t.TempDir()
	logFilePath := filepath.Join(tmpDir, "audit.log")

	var logs []*audit.AuditLog
	for i := 0; i < 15; i++ {
		logs = append(logs, &audit.AuditLog{
			Timestamp:     time.Now(),
			UserID:        fmt.Sprintf("user-%d", i),
			Instance:      fmt.Sprintf("db-%d", i),
			Query:         "SELECT 1",
			QueryType:     "SELECT",
			ExecutionTime: int64(i * 10),
			Status:        "success",
		})
	}
	writeAuditLogs(t, logFilePath, logs)

	router, _ := setupAuditTestRouter(t, logFilePath)

	req := httptest.NewRequest(http.MethodGet, "/audit/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	stats := resp["data"].(map[string]interface{})
	topUsers := stats["topUsers"].([]interface{})
	assert.LessOrEqual(t, len(topUsers), 10)

	topInstances := stats["topInstances"].([]interface{})
	assert.LessOrEqual(t, len(topInstances), 10)
}
