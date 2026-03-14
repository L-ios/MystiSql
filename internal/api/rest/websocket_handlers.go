package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"MystiSql/internal/service/auth"
	"MystiSql/internal/service/query"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WebSocketConfig struct {
	Enabled              bool
	MaxConnections       int
	IdleTimeout          time.Duration
	MaxConcurrentQueries int
}

type WebSocketHandlers struct {
	authService *auth.AuthService
	engine      *query.Engine
	logger      *zap.Logger
	config      WebSocketConfig
	upgrader    *websocket.Upgrader

	connections      map[*websocket.Conn]bool
	connectionsMutex sync.RWMutex

	querySemaphores      map[*websocket.Conn]chan struct{}
	querySemaphoresMutex sync.RWMutex
}

func NewWebSocketHandlers(authService *auth.AuthService, engine *query.Engine, config WebSocketConfig, logger *zap.Logger) *WebSocketHandlers {
	if config.MaxConnections <= 0 {
		config.MaxConnections = 1000
	}
	if config.IdleTimeout <= 0 {
		config.IdleTimeout = 10 * time.Minute
	}
	if config.MaxConcurrentQueries <= 0 {
		config.MaxConcurrentQueries = 5
	}

	return &WebSocketHandlers{
		authService:     authService,
		engine:          engine,
		logger:          logger,
		config:          config,
		upgrader:        &websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		connections:     make(map[*websocket.Conn]bool),
		querySemaphores: make(map[*websocket.Conn]chan struct{}),
	}
}

func (h *WebSocketHandlers) HandleWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	_, err := h.authService.ValidateToken(context.Background(), token)
	if err != nil {
		h.logger.Warn("WebSocket authentication failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	h.connectionsMutex.Lock()
	if len(h.connections) >= h.config.MaxConnections {
		h.connectionsMutex.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Maximum connections reached"})
		return
	}
	h.connectionsMutex.Unlock()

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}

	h.connectionsMutex.Lock()
	h.connections[conn] = true
	h.connectionsMutex.Unlock()

	defer func() {
		h.connectionsMutex.Lock()
		delete(h.connections, conn)
		h.connectionsMutex.Unlock()

		h.querySemaphoresMutex.Lock()
		if sem, ok := h.querySemaphores[conn]; ok {
			close(sem)
			delete(h.querySemaphores, conn)
		}
		h.querySemaphoresMutex.Unlock()
		conn.Close()
	}()

	conn.SetReadDeadline(time.Now().Add(h.config.IdleTimeout))
	h.handleMessages(conn)
}

func (h *WebSocketHandlers) handleMessages(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				h.logger.Debug("WebSocket closed unexpectedly", zap.Error(err))
			} else {
				h.logger.Debug("WebSocket closed", zap.Error(err))
			}
			return
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			h.logger.Warn("Invalid message format", zap.Error(err))
			h.sendError(conn, "", "Invalid message format")
			continue
		}

		action, _ := msg["action"].(string)
		switch action {
		case "query":
			h.handleQuery(conn, msg)
		case "ping":
			h.sendPong(conn)
		default:
			h.sendError(conn, "", fmt.Sprintf("Unknown action: %s", action))
		}
	}
}

func (h *WebSocketHandlers) handleQuery(conn *websocket.Conn, msg map[string]interface{}) {
	requestID, _ := msg["requestId"].(string)
	instance, _ := msg["instance"].(string)
	queryStr, _ := msg["query"].(string)

	if instance == "" || queryStr == "" {
		h.sendError(conn, requestID, "Instance and query are required")
		return
	}

	h.querySemaphoresMutex.Lock()
	if _, exists := h.querySemaphores[conn]; exists {
		h.querySemaphoresMutex.Unlock()
		h.logger.Warn("Max concurrent queries reached")
		h.sendError(conn, requestID, "Too many concurrent queries")
		return
	}
	sem := make(chan struct{}, 1)
	sem <- struct{}{}
	h.querySemaphores[conn] = sem
	h.querySemaphoresMutex.Unlock()

	defer func() {
		h.querySemaphoresMutex.Lock()
		delete(h.querySemaphores, conn)
		h.querySemaphoresMutex.Unlock()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.engine.ExecuteQuery(ctx, instance, queryStr)
	if err != nil {
		h.sendError(conn, requestID, err.Error())
		return
	}

	response := map[string]interface{}{
		"requestId": requestID,
		"success":   true,
		"columns":   result.Columns,
		"rows":      result.Rows,
		"rowCount":  result.RowCount,
	}
	h.sendMessage(conn, response)
}

func (h *WebSocketHandlers) sendPong(conn *websocket.Conn) {
	response := map[string]interface{}{"action": "pong"}
	h.sendMessage(conn, response)
}

func (h *WebSocketHandlers) sendError(conn *websocket.Conn, requestID string, message string) {
	response := map[string]interface{}{
		"requestId": requestID,
		"success":   false,
		"error":     message,
	}
	h.sendMessage(conn, response)
}

func (h *WebSocketHandlers) sendMessage(conn *websocket.Conn, message map[string]interface{}) {
	if err := conn.WriteJSON(message); err != nil {
		h.logger.Error("Failed to send WebSocket message", zap.Error(err))
	}
}
