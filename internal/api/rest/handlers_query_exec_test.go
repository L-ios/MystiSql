package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"MystiSql/internal/connection"
	"MystiSql/internal/discovery"
	"MystiSql/internal/service/query"
	"MystiSql/internal/service/validator"
	"MystiSql/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupQueryExecRouter(t *testing.T, validatorSvc *validator.ValidatorService) (*gin.Engine, *Handlers, *discovery.Registry) {
	logger := zap.NewNop()
	registry := discovery.NewRegistry()
	engine := query.NewEngine(registry, connection.GetRegistry())
	handlers := NewHandlers(registry, engine, validatorSvc, logger, "test-version")

	router := gin.New()
	router.POST("/api/v1/query", handlers.Query)
	router.POST("/api/v1/exec", handlers.Exec)
	router.GET("/api/v1/instances/:name/health", handlers.GetInstanceHealth)
	router.GET("/api/v1/instances/:name/pool", handlers.GetPoolStats)
	router.GET("/health", handlers.Health)

	return router, handlers, registry
}

// ==================== Query handler tests ====================

func TestQuery_InvalidJSON(t *testing.T) {
	router, _, _ := setupQueryExecRouter(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/query", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["success"].(bool))

	errDetail := resp["error"].(map[string]interface{})
	assert.Equal(t, "INVALID_REQUEST", errDetail["code"])
}

func TestQuery_MissingFields(t *testing.T) {
	router, _, _ := setupQueryExecRouter(t, nil)

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing instance",
			body: `{"sql":"SELECT 1"}`,
		},
		{
			name: "missing sql",
			body: `{"instance":"test-db"}`,
		},
		{
			name: "empty body",
			body: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/query", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestQuery_SQLBlocked(t *testing.T) {
	logger := zap.NewNop()
	validatorSvc := validator.NewValidatorService(logger)

	router, _, _ := setupQueryExecRouter(t, validatorSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/query", bytes.NewBufferString(`{"instance":"test","sql":"DROP TABLE users"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["success"].(bool))

	errDetail := resp["error"].(map[string]interface{})
	assert.Equal(t, "SQL_BLOCKED", errDetail["code"])
}

func TestQuery_EngineError(t *testing.T) {
	router, _, _ := setupQueryExecRouter(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/query", bytes.NewBufferString(`{"instance":"nonexistent","sql":"SELECT 1"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["success"].(bool))

	errDetail := resp["error"].(map[string]interface{})
	assert.Equal(t, "QUERY_FAILED", errDetail["code"])
}

func TestQuery_WithTimeout(t *testing.T) {
	router, _, _ := setupQueryExecRouter(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/query", bytes.NewBufferString(`{"instance":"nonexistent","sql":"SELECT 1","timeout":5}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestQuery_WithTransactionID(t *testing.T) {
	router, _, _ := setupQueryExecRouter(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/query", bytes.NewBufferString(`{"instance":"test","sql":"SELECT 1","transaction_id":"nonexistent-tx"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ==================== Exec handler tests ====================

func TestExec_InvalidJSON(t *testing.T) {
	router, _, _ := setupQueryExecRouter(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/exec", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["success"].(bool))

	errDetail := resp["error"].(map[string]interface{})
	assert.Equal(t, "INVALID_REQUEST", errDetail["code"])
}

func TestExec_MissingFields(t *testing.T) {
	router, _, _ := setupQueryExecRouter(t, nil)

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing instance",
			body: `{"sql":"DELETE FROM users WHERE id = 1"}`,
		},
		{
			name: "missing sql",
			body: `{"instance":"test-db"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/exec", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestExec_SQLBlocked(t *testing.T) {
	logger := zap.NewNop()
	validatorSvc := validator.NewValidatorService(logger)

	router, _, _ := setupQueryExecRouter(t, validatorSvc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/exec", bytes.NewBufferString(`{"instance":"test","sql":"DROP TABLE users"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["success"].(bool))

	errDetail := resp["error"].(map[string]interface{})
	assert.Equal(t, "SQL_BLOCKED", errDetail["code"])
}

func TestExec_EngineError(t *testing.T) {
	router, _, _ := setupQueryExecRouter(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/exec", bytes.NewBufferString(`{"instance":"nonexistent","sql":"INSERT INTO users (name) VALUES ('test')"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["success"].(bool))

	errDetail := resp["error"].(map[string]interface{})
	assert.Equal(t, "EXEC_FAILED", errDetail["code"])
}

func TestExec_WithTimeout(t *testing.T) {
	router, _, _ := setupQueryExecRouter(t, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/exec", bytes.NewBufferString(`{"instance":"nonexistent","sql":"UPDATE users SET name='x'","timeout":10}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ==================== GetInstanceHealth tests ====================

func TestGetInstanceHealth_NotFound(t *testing.T) {
	router, _, _ := setupQueryExecRouter(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/instances/nonexistent/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// ==================== GetPoolStats tests ====================

func TestGetPoolStats_NotFound(t *testing.T) {
	router, _, _ := setupQueryExecRouter(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/instances/nonexistent/pool", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["success"].(bool))
}

// ==================== Health with check-instances ====================

func TestHealth_WithCheckInstances(t *testing.T) {
	router, _, registry := setupQueryExecRouter(t, nil)

	// Register a healthy instance
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	instance.Status = types.InstanceStatusHealthy
	err := registry.Register(instance)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/health?check-instances=true", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "healthy", resp.Status)
	assert.NotNil(t, resp.Instances)
	assert.Equal(t, 1, resp.Instances.Total)
	assert.Equal(t, 1, resp.Instances.Healthy)
}

func TestHealth_WithCheckInstances_Unhealthy(t *testing.T) {
	router, _, registry := setupQueryExecRouter(t, nil)

	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	instance.Status = types.InstanceStatusUnhealthy
	err := registry.Register(instance)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/health?check-instances=true", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "unhealthy", resp.Status)
	assert.NotNil(t, resp.Instances)
	assert.Equal(t, 1, resp.Instances.Total)
	assert.Equal(t, 0, resp.Instances.Healthy)
	assert.Equal(t, 1, resp.Instances.Unhealthy)
}

func TestHealth_WithCheckInstances_UnknownStatus(t *testing.T) {
	router, _, registry := setupQueryExecRouter(t, nil)

	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	instance.Status = types.InstanceStatusUnknown
	err := registry.Register(instance)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/health?check-instances=true", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "unhealthy", resp.Status)
	assert.NotNil(t, resp.Instances)
	assert.Equal(t, 1, resp.Instances.Unhealthy)
}

func TestHealth_WithoutCheckInstances(t *testing.T) {
	router, _, _ := setupQueryExecRouter(t, nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "healthy", resp.Status)
	assert.Nil(t, resp.Instances)
}

// ==================== SetTransactionManager + Query with txManager ====================

func TestQuery_WithTxManager_TransactionNotFound(t *testing.T) {
	logger := zap.NewNop()
	registry := discovery.NewRegistry()
	engine := query.NewEngine(registry, connection.GetRegistry())
	handlers := NewHandlers(registry, engine, nil, logger, "test-version")

	router := gin.New()
	router.POST("/api/v1/query", handlers.Query)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/query", bytes.NewBufferString(`{"instance":"test","sql":"SELECT 1","transaction_id":"tx-123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestExec_WithTxManager_TransactionNotFound(t *testing.T) {
	logger := zap.NewNop()
	registry := discovery.NewRegistry()
	engine := query.NewEngine(registry, connection.GetRegistry())
	handlers := NewHandlers(registry, engine, nil, logger, "test-version")

	router := gin.New()
	router.POST("/api/v1/exec", handlers.Exec)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/exec", bytes.NewBufferString(`{"instance":"test","sql":"INSERT INTO t VALUES (1)","transaction_id":"tx-123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
