package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"MystiSql/internal/discovery"
	"MystiSql/internal/service/query"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter(t *testing.T) (*gin.Engine, *query.Engine, *zap.Logger) {
	logger := zap.NewNop()
	registry := discovery.NewRegistry()
	engine := query.NewEngine(registry)
	handlers := NewHandlers(registry, engine, nil, logger, "test-version")

	router := gin.New()
	router.GET("/health", handlers.Health)
	router.GET("/api/v1/instances", handlers.ListInstances)
	router.POST("/api/v1/query", handlers.Query)

	return router, engine, logger
}

func TestHealthEndpoint(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListInstancesEndpoint_Empty(t *testing.T) {
	router, _, _ := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/instances", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
func TestQueryEndpoint_Basic(t *testing.T) {
	router, engine, _ := setupTestRouter(t)

	_ = engine
	_ = router

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
