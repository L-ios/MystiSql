package websocket

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestConnectionManager_NewConnectionManager(t *testing.T) {
	logger := zap.NewNop()
	cm := NewConnectionManager(100, 5*time.Minute, logger)

	if cm == nil {
		t.Fatal("NewConnectionManager returned nil")
	}

	if cm.connections == nil {
		t.Error("connections map is nil")
	}

	if cm.maxConnections != 100 {
		t.Errorf("expected maxConnections 100, got %d", cm.maxConnections)
	}

	if cm.idleTimeout != 5*time.Minute {
		t.Errorf("expected idleTimeout 5m, got %v", cm.idleTimeout)
	}

	cm.Close()
}

func TestConnectionManager_AddConnection(t *testing.T) {
	logger := zap.NewNop()
	cm := NewConnectionManager(2, 5*time.Minute, logger)

	conn1 := &ClientConnection{ID: "test-1", UserID: "user1"}
	conn2 := &ClientConnection{ID: "test-2", UserID: "user2"}

	if !cm.AddConnection(conn1) {
		t.Error("AddConnection should succeed for first connection")
	}

	if cm.GetConnectionCount() != 1 {
		t.Errorf("expected connection count 1, got %d", cm.GetConnectionCount())
	}

	if !cm.AddConnection(conn2) {
		t.Error("AddConnection should succeed for second connection")
	}

	if cm.GetConnectionCount() != 2 {
		t.Errorf("expected connection count 2, got %d", cm.GetConnectionCount())
	}

	conn3 := &ClientConnection{ID: "test-3", UserID: "user3"}
	if cm.AddConnection(conn3) {
		t.Error("AddConnection should fail when max connections reached")
	}

	if cm.GetConnectionCount() != 2 {
		t.Errorf("expected connection count still 2, got %d", cm.GetConnectionCount())
	}

	delete(cm.connections, "test-1")
	delete(cm.connections, "test-2")
	cm.Close()
}

func TestConnectionManager_RemoveConnection(t *testing.T) {
	logger := zap.NewNop()
	cm := NewConnectionManager(10, 5*time.Minute, logger)

	conn := &ClientConnection{ID: "test-1", UserID: "user1"}
	cm.AddConnection(conn)

	if cm.GetConnectionCount() != 1 {
		t.Errorf("expected connection count 1, got %d", cm.GetConnectionCount())
	}

	cm.RemoveConnection("test-1")

	if cm.GetConnectionCount() != 0 {
		t.Errorf("expected connection count 0, got %d", cm.GetConnectionCount())
	}

	cm.RemoveConnection("nonexistent")

	delete(cm.connections, "nonexistent")
	cm.Close()
}

func TestConnectionManager_GetConnection(t *testing.T) {
	logger := zap.NewNop()
	cm := NewConnectionManager(10, 5*time.Minute, logger)

	conn := &ClientConnection{ID: "test-1", UserID: "user1"}
	cm.AddConnection(conn)

	retrieved, err := cm.GetConnection("test-1")
	if err != nil {
		t.Errorf("GetConnection should not return error for existing connection: %v", err)
	}

	if retrieved.ID != "test-1" {
		t.Errorf("expected connection ID test-1, got %s", retrieved.ID)
	}

	_, err = cm.GetConnection("nonexistent")
	if err != ErrConnectionNotFound {
		t.Errorf("expected ErrConnectionNotFound, got %v", err)
	}

	delete(cm.connections, "test-1")
	cm.Close()
}

func TestConnectionManager_GetAllConnections(t *testing.T) {
	logger := zap.NewNop()
	cm := NewConnectionManager(10, 5*time.Minute, logger)

	conn1 := &ClientConnection{ID: "test-1", UserID: "user1"}
	conn2 := &ClientConnection{ID: "test-2", UserID: "user2"}

	cm.AddConnection(conn1)
	cm.AddConnection(conn2)

	all := cm.GetAllConnections()
	if len(all) != 2 {
		t.Errorf("expected 2 connections, got %d", len(all))
	}

	cm.RemoveConnection("test-1")
	all = cm.GetAllConnections()
	if len(all) != 1 {
		t.Errorf("expected 1 connection after removal, got %d", len(all))
	}

	delete(cm.connections, "test-2")
	cm.Close()
}

func TestConnectionManager_GetConnectionCount(t *testing.T) {
	logger := zap.NewNop()
	cm := NewConnectionManager(10, 5*time.Minute, logger)

	if cm.GetConnectionCount() != 0 {
		t.Errorf("expected 0 connections initially, got %d", cm.GetConnectionCount())
	}

	cm.AddConnection(&ClientConnection{ID: "test-1", UserID: "user1"})
	if cm.GetConnectionCount() != 1 {
		t.Errorf("expected 1 connection, got %d", cm.GetConnectionCount())
	}

	cm.AddConnection(&ClientConnection{ID: "test-2", UserID: "user2"})
	if cm.GetConnectionCount() != 2 {
		t.Errorf("expected 2 connections, got %d", cm.GetConnectionCount())
	}

	cm.RemoveConnection("test-1")
	if cm.GetConnectionCount() != 1 {
		t.Errorf("expected 1 connection after removal, got %d", cm.GetConnectionCount())
	}

	delete(cm.connections, "test-2")
	cm.Close()
}

func TestConnectionManager_Close(t *testing.T) {
	logger := zap.NewNop()
	cm := NewConnectionManager(10, 5*time.Minute, logger)

	cm.AddConnection(&ClientConnection{ID: "test-1", UserID: "user1"})
	cm.AddConnection(&ClientConnection{ID: "test-2", UserID: "user2"})

	if cm.GetConnectionCount() != 2 {
		t.Errorf("expected 2 connections before close, got %d", cm.GetConnectionCount())
	}

	delete(cm.connections, "test-1")
	delete(cm.connections, "test-2")
	cm.Close()

	if cm.GetConnectionCount() != 0 {
		t.Errorf("expected 0 connections after close, got %d", cm.GetConnectionCount())
	}
}

func TestMessage_ToJSON(t *testing.T) {
	msg := &Message{
		RequestID: "req-123",
		Action:    MessageTypeQuery,
		Instance:  "mysql-1",
		Query:     "SELECT 1",
		Timestamp: time.Now().Unix(),
	}

	data, err := msg.ToJSON()
	if err != nil {
		t.Errorf("ToJSON failed: %v", err)
	}

	parsed, err := ParseMessage(data)
	if err != nil {
		t.Errorf("ParseMessage failed: %v", err)
	}

	if parsed.RequestID != msg.RequestID {
		t.Errorf("expected RequestID %s, got %s", msg.RequestID, parsed.RequestID)
	}

	if parsed.Action != msg.Action {
		t.Errorf("expected Action %s, got %s", msg.Action, parsed.Action)
	}

	if parsed.Instance != msg.Instance {
		t.Errorf("expected Instance %s, got %s", msg.Instance, parsed.Instance)
	}

	if parsed.Query != msg.Query {
		t.Errorf("expected Query %s, got %s", msg.Query, parsed.Query)
	}
}

func TestMessage_ParseMessage(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantErr   bool
		checkFunc func(*Message) bool
	}{
		{
			name:    "valid query message",
			data:    []byte(`{"action":"query","instance":"mysql-1","query":"SELECT 1"}`),
			wantErr: false,
			checkFunc: func(m *Message) bool {
				return m.Action == MessageTypeQuery && m.Instance == "mysql-1" && m.Query == "SELECT 1"
			},
		},
		{
			name:    "invalid JSON",
			data:    []byte(`{"action":`),
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "with request ID",
			data:    []byte(`{"requestId":"req-123","action":"query","instance":"mysql-1"}`),
			wantErr: false,
			checkFunc: func(m *Message) bool {
				return m.RequestID == "req-123"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseMessage(tt.data)
			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tt.checkFunc != nil && msg != nil && !tt.checkFunc(msg) {
				t.Error("message content validation failed")
			}
		})
	}
}

func TestMessageTypeConstants(t *testing.T) {
	if MessageTypeQuery != "query" {
		t.Errorf("MessageTypeQuery should be 'query', got %s", MessageTypeQuery)
	}

	if MessageTypeQueryResult != "query_result" {
		t.Errorf("MessageTypeQueryResult should be 'query_result', got %s", MessageTypeQueryResult)
	}

	if MessageTypeError != "error" {
		t.Errorf("MessageTypeError should be 'error', got %s", MessageTypeError)
	}

	if MessageTypePong != "pong" {
		t.Errorf("MessageTypePong should be 'pong', got %s", MessageTypePong)
	}

	if MessageTypePing != "ping" {
		t.Errorf("MessageTypePing should be 'ping', got %s", MessageTypePing)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxConnections != 100 {
		t.Errorf("expected MaxConnections 100, got %d", config.MaxConnections)
	}

	if config.IdleTimeout != 5*time.Minute {
		t.Errorf("expected IdleTimeout 5m, got %v", config.IdleTimeout)
	}

	if config.ReadBufferSize != 1024 {
		t.Errorf("expected ReadBufferSize 1024, got %d", config.ReadBufferSize)
	}

	if config.WriteBufferSize != 1024 {
		t.Errorf("expected WriteBufferSize 1024, got %d", config.WriteBufferSize)
	}

	if config.EnableCompression != false {
		t.Errorf("expected EnableCompression false, got %v", config.EnableCompression)
	}
}
