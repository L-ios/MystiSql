package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"MystiSql/internal/service/transaction"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTransactionTestRouter(t *testing.T) (*gin.Engine, *TransactionHandlers) {
	logger := zap.NewNop()
	txManager := transaction.NewTransactionManager(nil, logger, nil)
	handlers := NewTransactionHandlers(txManager, logger)

	router := gin.New()
	router.POST("/transaction/begin", handlers.BeginTransaction)
	router.POST("/transaction/commit", handlers.CommitTransaction)
	router.POST("/transaction/rollback", handlers.RollbackTransaction)
	router.GET("/transaction/:id", handlers.GetTransaction)
	router.GET("/transaction", handlers.ListTransactions)
	router.POST("/transaction/:id/extend", handlers.ExtendTransaction)

	return router, handlers
}

func TestNewTransactionHandlers(t *testing.T) {
	logger := zap.NewNop()
	txManager := transaction.NewTransactionManager(nil, logger, nil)
	handlers := NewTransactionHandlers(txManager, logger)

	assert.NotNil(t, handlers)
	assert.Equal(t, txManager, handlers.txManager)
}

func TestBeginTransaction_InvalidJSON(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/begin", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "Invalid request")
}

func TestBeginTransaction_MissingInstance(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/begin", bytes.NewBufferString(`{"isolation_level":"read_committed"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBeginTransaction_EngineError(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/begin", bytes.NewBufferString(`{"instance":"nonexistent"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	assert.Panics(t, func() {
		router.ServeHTTP(w, req)
	})
}

func TestBeginTransaction_WithUserID(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/begin", bytes.NewBufferString(`{"instance":"nonexistent"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	assert.Panics(t, func() {
		router.ServeHTTP(w, req)
	})
}

func TestBeginTransaction_WithTimeout(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/begin", bytes.NewBufferString(`{"instance":"nonexistent","timeout":"5m"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	assert.Panics(t, func() {
		router.ServeHTTP(w, req)
	})
}

func TestCommitTransaction_InvalidJSON(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/commit", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCommitTransaction_MissingTransactionID(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/commit", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCommitTransaction_NotFound(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/commit", bytes.NewBufferString(`{"transaction_id":"nonexistent-tx"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Transaction not found", resp["error"])
}

func TestRollbackTransaction_InvalidJSON(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/rollback", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRollbackTransaction_MissingTransactionID(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/rollback", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRollbackTransaction_NotFound(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/rollback", bytes.NewBufferString(`{"transaction_id":"nonexistent-tx"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Transaction not found", resp["error"])
}

func TestGetTransaction_NotFound(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/transaction/nonexistent-tx", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Transaction not found", resp["error"])
}

func TestListTransactions_Empty(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/transaction", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	transactions := resp["transactions"].([]interface{})
	assert.Empty(t, transactions)
	assert.Equal(t, float64(0), resp["count"])
}

func TestExtendTransaction_InvalidJSON(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/tx-123/extend", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExtendTransaction_MissingDuration(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/tx-123/extend", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestExtendTransaction_InvalidDuration(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/tx-123/extend", bytes.NewBufferString(`{"duration":"not-a-duration"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Invalid duration format", resp["error"])
}

func TestExtendTransaction_NotFound(t *testing.T) {
	router, _ := setupTransactionTestRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/transaction/tx-123/extend", bytes.NewBufferString(`{"duration":"5m"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "Transaction not found", resp["error"])
}
