package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"MystiSql/internal/service/batch"
	"MystiSql/internal/service/rbac"
	"MystiSql/internal/service/validator"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupRBACTestRouter(t *testing.T) (*gin.Engine, *RBACHandlers) {
	logger := zap.NewNop()
	service := rbac.NewRBACService()
	handlers := NewRBACHandlers(service, logger)

	router := gin.New()
	router.GET("/roles", handlers.ListRoles)
	router.GET("/roles/:name", handlers.GetRole)
	router.POST("/roles", handlers.CreateRole)
	router.DELETE("/roles/:name", handlers.DeleteRole)
	router.GET("/users/:id/roles", handlers.ListUserRoles)
	router.POST("/users/:id/roles", handlers.AssignRoleToUser)

	return router, handlers
}

func TestNewRBACHandlers_NilService(t *testing.T) {
	logger := zap.NewNop()
	handlers := NewRBACHandlers(nil, logger)
	assert.Nil(t, handlers)
}

func TestListRoles_Success(t *testing.T) {
	router, _ := setupRBACTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	roles, ok := resp["roles"].([]interface{})
	require.True(t, ok, "roles should be an array")
	assert.GreaterOrEqual(t, len(roles), 0)
}

func TestGetRole_NotFound(t *testing.T) {
	router, _ := setupRBACTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/roles/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "role not found", resp["error"])
}

func TestGetRole_Success(t *testing.T) {
	router, _ := setupRBACTestRouter(t)

	createReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString(`{"name":"admin","permissions":["test:db:read"]}`))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	req := httptest.NewRequest(http.MethodGet, "/roles/admin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "response: "+w.Body.String())

	var role map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &role)
	require.NoError(t, err, "unmarshal error: "+w.Body.String())
	assert.Equal(t, "admin", role["Name"])
}

func TestCreateRole_InvalidJSON(t *testing.T) {
	router, _ := setupRBACTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRole_InvalidPermission(t *testing.T) {
	router, _ := setupRBACTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString(`{"name":"test","permissions":["invalid:perm"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"].(string), "invalid permission")
}

func TestCreateRole_Success(t *testing.T) {
	router, _ := setupRBACTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString(`{"name":"editor","permissions":["test:db:read","prod:db:write"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "response: "+w.Body.String())

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "role created successfully", resp["message"])
	roleData, ok := resp["role"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "editor", roleData["Name"])
}

func TestDeleteRole_Success(t *testing.T) {
	router, _ := setupRBACTestRouter(t)

	createReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString(`{"name":"deleteme","permissions":["test:db:read"]}`))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	req := httptest.NewRequest(http.MethodDelete, "/roles/deleteme", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "role deleted successfully", resp["message"])
}

func TestListUserRoles_Success(t *testing.T) {
	router, _ := setupRBACTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/users/user123/roles", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "user123", resp["user_id"])
	roles, ok := resp["roles"].([]interface{})
	require.True(t, ok)
	assert.Empty(t, roles)
}

func TestAssignRoleToUser_InvalidJSON(t *testing.T) {
	router, _ := setupRBACTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/users/user1/roles", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAssignRoleToUser_Success(t *testing.T) {
	router, _ := setupRBACTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/users/user1/roles", bytes.NewBufferString(`{"roles":["admin","editor"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "roles assigned successfully", resp["message"])
}

func setupValidatorTestRouter(t *testing.T) (*gin.Engine, *ValidatorHandlers) {
	logger := zap.NewNop()
	service := validator.NewValidatorService(logger)
	handlers := NewValidatorHandlers(service, logger)

	router := gin.New()
	router.POST("/validator/whitelist", handlers.UpdateWhitelist)
	router.GET("/validator/whitelist", handlers.GetWhitelist)
	router.POST("/validator/blacklist", handlers.UpdateBlacklist)
	router.GET("/validator/blacklist", handlers.GetBlacklist)
	router.POST("/validator/validate", handlers.ValidateQuery)

	return router, handlers
}

func TestUpdateWhitelist_InvalidJSON(t *testing.T) {
	router, _ := setupValidatorTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/validator/whitelist", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateWhitelist_Success(t *testing.T) {
	router, _ := setupValidatorTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/validator/whitelist", bytes.NewBufferString(`{"patterns":["SELECT.*","SHOW.*"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, float64(2), resp["count"])
}

func TestGetWhitelist_Success(t *testing.T) {
	router, _ := setupValidatorTestRouter(t)

	updateReq := httptest.NewRequest(http.MethodPost, "/validator/whitelist", bytes.NewBufferString(`{"patterns":["^SELECT"]}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateW := httptest.NewRecorder()
	router.ServeHTTP(updateW, updateReq)

	req := httptest.NewRequest(http.MethodGet, "/validator/whitelist", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	patterns, ok := resp["patterns"].([]interface{})
	require.True(t, ok)
	assert.Contains(t, patterns, "^SELECT")
}

func TestUpdateBlacklist_InvalidJSON(t *testing.T) {
	router, _ := setupValidatorTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/validator/blacklist", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateBlacklist_Success(t *testing.T) {
	router, _ := setupValidatorTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/validator/blacklist", bytes.NewBufferString(`{"patterns":["DROP.*","TRUNCATE.*"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, float64(2), resp["count"])
}

func TestGetBlacklist_Success(t *testing.T) {
	router, _ := setupValidatorTestRouter(t)

	updateReq := httptest.NewRequest(http.MethodPost, "/validator/blacklist", bytes.NewBufferString(`{"patterns":["^DROP"]}`))
	updateReq.Header.Set("Content-Type", "application/json")
	updateW := httptest.NewRecorder()
	router.ServeHTTP(updateW, updateReq)

	req := httptest.NewRequest(http.MethodGet, "/validator/blacklist", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	patterns, ok := resp["patterns"].([]interface{})
	require.True(t, ok)
	assert.Contains(t, patterns, "^DROP")
}

func TestValidateQuery_InvalidJSON(t *testing.T) {
	router, _ := setupValidatorTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/validator/validate", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestValidateQuery_MissingFields(t *testing.T) {
	router, _ := setupValidatorTestRouter(t)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "missing instance",
			body:       `{"query":"SELECT 1"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing query",
			body:       `{"instance":"test"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/validator/validate", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestValidateQuery_Allowed(t *testing.T) {
	router, _ := setupValidatorTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/validator/validate", bytes.NewBufferString(`{"instance":"test","query":"SELECT * FROM users"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, true, resp["allowed"])
}

func setupBatchTestRouter(t *testing.T) (*gin.Engine, *BatchHandlers) {
	logger := zap.NewNop()
	batchSvc := batch.NewBatchService(nil, batch.DefaultBatchConfig(), logger)
	handlers := NewBatchHandlers(batchSvc, logger)

	router := gin.New()
	router.POST("/batch", handlers.ExecuteBatch)

	return router, handlers
}

func TestNewBatchHandlers(t *testing.T) {
	logger := zap.NewNop()
	batchSvc := batch.NewBatchService(nil, batch.DefaultBatchConfig(), logger)
	handlers := NewBatchHandlers(batchSvc, logger)

	assert.NotNil(t, handlers)
	assert.Equal(t, batchSvc, handlers.batchService)
}

func TestExecuteBatch_InvalidJSON(t *testing.T) {
	router, _ := setupBatchTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "Invalid request")
}

func TestExecuteBatch_EmptyQueries(t *testing.T) {
	router, _ := setupBatchTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewBufferString(`{"instance":"test","queries":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Batch queries cannot be empty", resp["error"])
}

func TestExecuteBatch_TooManyQueries(t *testing.T) {
	router, _ := setupBatchTestRouter(t)

	queries := make([]string, 1001)
	for i := range queries {
		queries[i] = "SELECT 1"
	}
	body, _ := json.Marshal(map[string]interface{}{
		"instance": "test",
		"queries":  queries,
	})

	req := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "exceeds maximum limit")
}

func TestExecuteBatch_MissingInstance(t *testing.T) {
	router, _ := setupBatchTestRouter(t)

	// "instance" has binding:"required", so ShouldBindJSON fails before reaching the empty check
	req := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewBufferString(`{"queries":["SELECT 1"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "response: "+w.Body.String())

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "Invalid request")
}

func TestExecuteBatch_PartialFailure(t *testing.T) {
	router, _ := setupBatchTestRouter(t)

	// ExecuteBatch returns a response even when individual queries fail (partial success pattern)
	req := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewBufferString(`{"instance":"nonexistent","queries":["SELECT 1"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "response: "+w.Body.String())

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["message"], "successfully")
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "short token",
			token:    "short",
			expected: "****",
		},
		{
			name:     "exact 20 chars",
			token:    "12345678901234567890",
			expected: "****",
		},
		{
			name:     "long token",
			token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.testpayload.signature",
			expected: "eyJhbGciOi....signature",
		},
		{
			name:     "very long token",
			token:    "this_is_a_very_long_token_that_should_be_masked_in_the_middle",
			expected: "this_is_a_...the_middle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskToken(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}
