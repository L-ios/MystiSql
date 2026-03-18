package rest

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CORSMiddleware 创建 CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
}

// LoggerMiddleware 创建日志中间件
// 记录每个请求的详细信息，包括请求方法、路径、状态码、响应时间等
func LoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 请求完成后记录日志
		latency := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		errors := c.Errors.ByType(gin.ErrorTypePrivate)

		// 构建日志字段
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
		}

		// 根据状态码选择日志级别
		if status >= 500 {
			logger.Error("Server error", fields...)
		} else if status >= 400 {
			logger.Warn("Client error", fields...)
		} else {
			logger.Info("Request", fields...)
		}

		// 记录错误
		if len(errors) > 0 {
			for _, err := range errors {
				logger.Error("Request error",
					zap.String("error", err.Error()),
					zap.Any("meta", err.Meta),
				)
			}
		}
	}
}

// RecoveryMiddleware 创建错误恢复中间件
// 捕获 panic 并返回 500 错误，避免服务崩溃
func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录堆栈信息
				stack := string(debug.Stack())
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", stack),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
				)

				// 返回 500 错误
				c.JSON(http.StatusInternalServerError, NewErrorResponse(
					"INTERNAL_SERVER_ERROR",
					fmt.Sprintf("Internal server error: %v", err),
				))
				c.Abort()
			}
		}()

		c.Next()
	}
}
