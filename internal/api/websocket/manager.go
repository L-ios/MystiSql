package websocket

import (
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	ErrConnectionNotFound = errors.New("connection not found")
	ErrMaxConnections     = errors.New("maximum connections reached")
)

type ConnectionManager struct {
	connections    map[string]*ClientConnection
	maxConnections int
	idleTimeout    time.Duration
	logger         *zap.Logger
	mu             sync.RWMutex
	stopCh         chan struct{}
}

func NewConnectionManager(maxConnections int, idleTimeout time.Duration, logger *zap.Logger) *ConnectionManager {
	cm := &ConnectionManager{
		connections:    make(map[string]*ClientConnection),
		maxConnections: maxConnections,
		idleTimeout:    idleTimeout,
		logger:         logger,
		stopCh:         make(chan struct{}),
	}

	go cm.cleanupIdleConnections()

	return cm
}

func (cm *ConnectionManager) AddConnection(conn *ClientConnection) bool {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(cm.connections) >= cm.maxConnections {
		cm.logger.Warn("Maximum WebSocket connections reached",
			zap.Int("current", len(cm.connections)),
			zap.Int("max", cm.maxConnections),
		)
		return false
	}

	cm.connections[conn.ID] = conn
	cm.logger.Info("WebSocket connection added",
		zap.String("connection_id", conn.ID),
		zap.Int("total_connections", len(cm.connections)),
	)

	return true
}

func (cm *ConnectionManager) RemoveConnection(connectionID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, exists := cm.connections[connectionID]; exists {
		delete(cm.connections, connectionID)
		cm.logger.Info("WebSocket connection removed",
			zap.String("connection_id", connectionID),
			zap.Int("total_connections", len(cm.connections)),
		)
	}
}

func (cm *ConnectionManager) GetConnection(connectionID string) (*ClientConnection, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, exists := cm.connections[connectionID]
	if !exists {
		return nil, ErrConnectionNotFound
	}

	return conn, nil
}

func (cm *ConnectionManager) GetAllConnections() []*ClientConnection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	connections := make([]*ClientConnection, 0, len(cm.connections))
	for _, conn := range cm.connections {
		connections = append(connections, conn)
	}

	return connections
}

func (cm *ConnectionManager) GetConnectionCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.connections)
}

func (cm *ConnectionManager) cleanupIdleConnections() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cm.checkIdleConnections()
		case <-cm.stopCh:
			return
		}
	}
}

func (cm *ConnectionManager) checkIdleConnections() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	for id, conn := range cm.connections {
		conn.mu.Lock()
		if now.Sub(conn.lastActivity) > cm.idleTimeout {
			cm.logger.Info("Closing idle WebSocket connection",
				zap.String("connection_id", id),
				zap.Duration("idle_time", now.Sub(conn.lastActivity)),
			)
			conn.Close()
			delete(cm.connections, id)
		}
		conn.mu.Unlock()
	}
}

func (cm *ConnectionManager) Close() error {
	close(cm.stopCh)

	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, conn := range cm.connections {
		conn.Close()
	}

	cm.connections = make(map[string]*ClientConnection)
	return nil
}

func (cm *ConnectionManager) BroadcastMessage(v interface{}) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var lastErr error
	for _, conn := range cm.connections {
		if err := conn.SendMessage(v); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

func (cm *ConnectionManager) BroadcastToUser(userID string, v interface{}) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var lastErr error
	for _, conn := range cm.connections {
		if conn.UserID == userID {
			if err := conn.SendMessage(v); err != nil {
				lastErr = err
			}
		}
	}

	return lastErr
}
