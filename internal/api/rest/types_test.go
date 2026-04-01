package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"MystiSql/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	resp := NewErrorResponse("INVALID_REQUEST", "Missing required field")

	assert.False(t, resp.Success)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, "INVALID_REQUEST", resp.Error.Code)
	assert.Equal(t, "Missing required field", resp.Error.Message)
}

func TestToInstanceResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	instance := &types.DatabaseInstance{
		Name:     "test-instance",
		Type:     types.DatabaseTypeMySQL,
		Host:     "localhost",
		Port:     3306,
		Database: "testdb",
		Username: "admin",
		Password: "secret-password",
		Status:   types.InstanceStatusHealthy,
		Labels:   map[string]string{"env": "test"},
	}

	resp := ToInstanceResponse(instance)

	assert.Equal(t, "test-instance", resp.Name)
	assert.Equal(t, types.DatabaseTypeMySQL, resp.Type)
	assert.Equal(t, "localhost", resp.Host)
	assert.Equal(t, 3306, resp.Port)
	assert.Equal(t, "testdb", resp.Database)
	assert.Equal(t, "admin", resp.Username)
	assert.Equal(t, string(types.InstanceStatusHealthy), resp.Status)
	assert.Equal(t, map[string]string{"env": "test"}, resp.Labels)
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORSMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	router := gin.New()
	router.Use(LoggerMiddleware(logger))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	router := gin.New()
	router.Use(RecoveryMiddleware(logger))
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/panic", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zap.NewNop()
	router := gin.New()
	router.Use(RecoveryMiddleware(logger))
	router.GET("/normal", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/normal", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
