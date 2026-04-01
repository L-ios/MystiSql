package websocket

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"MystiSql/internal/service/auth"
	"MystiSql/internal/service/query"
	"MystiSql/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var (
	ErrConnectionLimitReached = errors.New("maximum connection limit reached")
	ErrUnauthorized           = errors.New("unauthorized")
)

type WebSocketHandler struct {
	queryEngine *query.Engine
	authService *auth.AuthService
	connManager *ConnectionManager
	logger      *zap.Logger
	upgrader    websocket.Upgrader
}

type Config struct {
	MaxConnections    int
	IdleTimeout       time.Duration
	ReadBufferSize    int
	WriteBufferSize   int
	EnableCompression bool
}

func DefaultConfig() *Config {
	return &Config{
		MaxConnections:    100,
		IdleTimeout:       5 * time.Minute,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: false,
	}
}

func NewWebSocketHandler(queryEngine *query.Engine, authService *auth.AuthService, logger *zap.Logger, config *Config) *WebSocketHandler {
	if config == nil {
		config = DefaultConfig()
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:    config.ReadBufferSize,
		WriteBufferSize:   config.WriteBufferSize,
		EnableCompression: config.EnableCompression,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return &WebSocketHandler{
		queryEngine: queryEngine,
		authService: authService,
		connManager: NewConnectionManager(config.MaxConnections, config.IdleTimeout, logger),
		logger:      logger,
		upgrader:    upgrader,
	}
}

func (h *WebSocketHandler) Handle(c *gin.Context) {
	token := h.extractToken(c)
	if token == "" {
		h.logger.Warn("WebSocket connection rejected: missing token",
			zap.String("client_ip", c.ClientIP()),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "MISSING_TOKEN",
				"message": "Authentication token is required",
			},
		})
		return
	}

	claims, err := h.authService.ValidateToken(c.Request.Context(), token)
	if err != nil {
		h.logger.Warn("WebSocket connection rejected: invalid token",
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_TOKEN",
				"message": "Invalid or expired token",
			},
		})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection",
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		return
	}

	clientConn := NewClientConnection(conn, claims.UserID, c.ClientIP(), h.logger)
	if !h.connManager.AddConnection(clientConn) {
		clientConn.Close()
		h.logger.Warn("WebSocket connection rejected: maximum limit reached",
			zap.String("client_ip", c.ClientIP()),
			zap.Int("max_connections", h.connManager.maxConnections),
		)
		return
	}

	h.logger.Info("WebSocket connection established",
		zap.String("connection_id", clientConn.ID),
		zap.String("user_id", claims.UserID),
		zap.String("client_ip", c.ClientIP()),
	)

	go h.handleConnection(c.Request.Context(), clientConn)
}

func (h *WebSocketHandler) extractToken(c *gin.Context) string {
	token := c.Query("token")
	if token != "" {
		return token
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return authHeader[7:]
		}
	}

	return ""
}

func (h *WebSocketHandler) handleConnection(ctx context.Context, conn *ClientConnection) {
	defer func() {
		h.connManager.RemoveConnection(conn.ID)
		conn.Close()
		h.logger.Info("WebSocket connection closed",
			zap.String("connection_id", conn.ID),
			zap.String("user_id", conn.UserID),
		)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					h.logger.Warn("WebSocket read error",
						zap.String("connection_id", conn.ID),
						zap.Error(err),
					)
				}
				return
			}

			if messageType == websocket.TextMessage {
				go h.handleMessage(ctx, conn, data)
			}
		}
	}
}

func (h *WebSocketHandler) handleMessage(ctx context.Context, conn *ClientConnection, data []byte) {
	msg, err := ParseMessage(data)
	if err != nil {
		conn.SendError("Invalid message format", err.Error())
		return
	}

	ctx = query.ContextWithUserID(ctx, conn.UserID)
	ctx = query.ContextWithClientIP(ctx, conn.ClientIP)

	switch msg.Action {
	case MessageTypeQuery:
		h.handleQuery(ctx, conn, msg)
	case MessageTypePing:
		conn.SendPong(time.Now().Unix())
	default:
		conn.SendError("Unknown action", fmt.Sprintf("Action '%s' is not supported", msg.Action))
	}
}

func (h *WebSocketHandler) handleQuery(ctx context.Context, conn *ClientConnection, msg *Message) {
	if msg.Instance == "" || msg.Query == "" {
		conn.SendError("Invalid request", "Instance and query are required")
		return
	}

	result, err := h.queryEngine.ExecuteQuery(ctx, msg.Instance, msg.Query)
	if err != nil {
		h.logger.Warn("Query execution failed via WebSocket",
			zap.String("connection_id", conn.ID),
			zap.String("instance", msg.Instance),
			zap.Error(err),
		)
		conn.SendError("Query execution failed", err.Error())
		return
	}

	conn.SendQueryResult(msg.RequestID, result)
}

func (h *WebSocketHandler) GetConnectionManager() *ConnectionManager {
	return h.connManager
}

func (h *WebSocketHandler) Close() error {
	return h.connManager.Close()
}

type ClientConnection struct {
	ID           string
	UserID       string
	ClientIP     string
	conn         *websocket.Conn
	logger       *zap.Logger
	mu           sync.Mutex
	closed       bool
	lastActivity time.Time
}

func NewClientConnection(conn *websocket.Conn, userID, clientIP string, logger *zap.Logger) *ClientConnection {
	return &ClientConnection{
		ID:           uuid.New().String(),
		UserID:       userID,
		ClientIP:     clientIP,
		conn:         conn,
		logger:       logger,
		lastActivity: time.Now(),
	}
}

func (c *ClientConnection) ReadMessage() (int, []byte, error) {
	messageType, data, err := c.conn.ReadMessage()
	if err == nil {
		c.mu.Lock()
		c.lastActivity = time.Now()
		c.mu.Unlock()
	}
	return messageType, data, err
}

func (c *ClientConnection) SendMessage(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return errors.New("connection closed")
	}

	if err := c.conn.WriteJSON(v); err != nil {
		return err
	}
	c.lastActivity = time.Now()
	return nil
}

func (c *ClientConnection) SendQueryResult(requestID string, result *types.QueryResult) error {
	msg := QueryResultMessage{
		RequestID:     requestID,
		Type:          MessageTypeQueryResult,
		Columns:       convertColumns(result.Columns),
		Rows:          convertRows(result.Columns, result.Rows),
		RowCount:      result.RowCount,
		ExecutionTime: result.ExecutionTime.Milliseconds(),
	}
	return c.SendMessage(msg)
}

func convertColumns(cols []types.ColumnInfo) []map[string]interface{} {
	result := make([]map[string]interface{}, len(cols))
	for i, col := range cols {
		result[i] = map[string]interface{}{
			"name": col.Name,
			"type": col.Type,
		}
	}
	return result
}

func convertRows(cols []types.ColumnInfo, rows []types.Row) []map[string]interface{} {
	result := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		rowMap := make(map[string]interface{})
		for j, col := range cols {
			if j < len(row) {
				rowMap[col.Name] = row[j]
			}
		}
		result[i] = rowMap
	}
	return result
}

func (c *ClientConnection) SendError(code, message string) error {
	msg := ErrorMessage{
		Type:    MessageTypeError,
		Code:    code,
		Message: message,
	}
	return c.SendMessage(msg)
}

func (c *ClientConnection) SendPong(timestamp int64) error {
	msg := PongMessage{
		Type:      MessageTypePong,
		Timestamp: timestamp,
	}
	return c.SendMessage(msg)
}

func (c *ClientConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	return c.conn.Close()
}
