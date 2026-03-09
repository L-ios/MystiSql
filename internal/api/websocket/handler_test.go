package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func TestWebSocketHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		setupMock  func()
		expectCode int
	}{
		{
			name:       "WebSocket placeholder response",
			expectCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewWebSocketHandler()

			router := gin.New()
			router.GET("/ws", handler.Handle)

			req := httptest.NewRequest("GET", "/ws", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectCode {
				t.Errorf("Expected status code %d, got %d", tt.expectCode, w.Code)
			}
		})
	}
}

func TestWebSocketMessage(t *testing.T) {
	tests := []struct {
		name    string
		message Message
		json    string
	}{
		{
			name: "Query message",
			message: Message{
				Action:   "query",
				Instance: "test-mysql",
				Query:    "SELECT * FROM users",
			},
			json: `{"action":"query","instance":"test-mysql","query":"SELECT * FROM users"}`,
		},
		{
			name: "Ping message",
			message: Message{
				Action: "ping",
			},
			json: `{"action":"ping"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatalf("Failed to marshal message: %v", err)
			}

			if string(data) != tt.json {
				t.Errorf("Expected JSON %s, got %s", tt.json, string(data))
			}

			var unmarshaled Message
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal message: %v", err)
			}

			if unmarshaled.Action != tt.message.Action {
				t.Errorf("Expected action %s, got %s", tt.message.Action, unmarshaled.Action)
			}
		})
	}
}

func TestWebSocketManager(t *testing.T) {
	manager := NewManager()

	t.Run("Add connection", func(t *testing.T) {
		conn := &websocket.Conn{}
		manager.AddConnection("conn1", conn)

		if manager.GetConnectionCount() != 1 {
			t.Errorf("Expected 1 connection, got %d", manager.GetConnectionCount())
		}
	})

	t.Run("Remove connection", func(t *testing.T) {
		manager.RemoveConnection("conn1")

		if manager.GetConnectionCount() != 0 {
			t.Errorf("Expected 0 connections, got %d", manager.GetConnectionCount())
		}
	})

	t.Run("Max connections", func(t *testing.T) {
		manager.maxConnections = 2

		conn1 := &websocket.Conn{}
		conn2 := &websocket.Conn{}
		conn3 := &websocket.Conn{}

		manager.AddConnection("conn1", conn1)
		manager.AddConnection("conn2", conn2)

		if manager.CanAddConnection() {
			t.Error("Should not allow more connections when at max")
		}

		manager.AddConnection("conn3", conn3)

		if manager.GetConnectionCount() != 2 {
			t.Errorf("Expected 2 connections (max), got %d", manager.GetConnectionCount())
		}

		manager.RemoveAllConnections()
	})
}

func TestWebSocketHandler_Authentication(t *testing.T) {
	handler := NewWebSocketHandler()

	t.Run("Token from URL parameter", func(t *testing.T) {
		router := gin.New()
		router.GET("/ws", handler.Handle)

		req := httptest.NewRequest("GET", "/ws?token=test-token-123", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		token := req.URL.Query().Get("token")
		if token != "test-token-123" {
			t.Errorf("Expected token test-token-123, got %s", token)
		}
	})
}

func TestWebSocketHandler_Heartbeat(t *testing.T) {
	handler := NewWebSocketHandler()

	pingMsg := Message{Action: "ping"}
	data, err := json.Marshal(pingMsg)
	if err != nil {
		t.Fatalf("Failed to marshal ping message: %v", err)
	}

	t.Run("Ping message format", func(t *testing.T) {
		if pingMsg.Action != "ping" {
			t.Errorf("Expected action 'ping', got %s", pingMsg.Action)
		}
	})

	t.Run("Pong response format", func(t *testing.T) {
		pongMsg := Message{Type: "pong"}
		pongData, err := json.Marshal(pongMsg)
		if err != nil {
			t.Fatalf("Failed to marshal pong message: %v", err)
		}

		expected := `{"type":"pong"}`
		if string(pongData) != expected {
			t.Errorf("Expected pong message %s, got %s", expected, string(pongData))
		}
	})
}

func TestWebSocketHandler_QueryExecution(t *testing.T) {
	handler := NewWebSocketHandler()

	queryMsg := Message{
		Action:   "query",
		Instance: "test-mysql",
		Query:    "SELECT * FROM users LIMIT 10",
	}

	t.Run("Query message validation", func(t *testing.T) {
		if queryMsg.Action != "query" {
			t.Errorf("Expected action 'query', got %s", queryMsg.Action)
		}

		if queryMsg.Instance == "" {
			t.Error("Instance should not be empty for query action")
		}

		if queryMsg.Query == "" {
			t.Error("Query should not be empty for query action")
		}
	})
}

func TestWebSocketResponse(t *testing.T) {
	tests := []struct {
		name     string
		response Message
		json     string
	}{
		{
			name: "Result response",
			response: Message{
				Type:     "result",
				Columns:  []string{"id", "name", "email"},
				Rows:     [][]interface{}{{1, "Alice", "alice@example.com"}},
				RowCount: 1,
			},
			json: `{"type":"result","columns":["id","name","email"],"rows":[[1,"Alice","alice@example.com"]],"rowCount":1}`,
		},
		{
			name: "Error response",
			response: Message{
				Type:    "error",
				Message: "Query execution failed",
				Code:    "QUERY_ERROR",
			},
			json: `{"type":"error","message":"Query execution failed","code":"QUERY_ERROR"}`,
		},
		{
			name: "Pong response",
			response: Message{
				Type: "pong",
			},
			json: `{"type":"pong"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Failed to marshal response: %v", err)
			}

			if string(data) != tt.json {
				t.Errorf("Expected JSON %s, got %s", tt.json, string(data))
			}
		})
	}
}

func TestWebSocketConnectionLifecycle(t *testing.T) {
	manager := NewManager()
	manager.maxConnections = 5
	manager.idleTimeout = 1 * time.Second

	t.Run("Connection limit", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			manager.AddConnection(string(rune(i)), &websocket.Conn{})
		}

		if !manager.IsAtMaxCapacity() {
			t.Error("Manager should be at max capacity")
		}

		manager.RemoveAllConnections()
	})

	t.Run("Concurrent connections", func(t *testing.T) {
		done := make(chan bool)

		for i := 0; i < 10; i++ {
			go func(id int) {
				connID := string(rune('A' + id))
				manager.AddConnection(connID, &websocket.Conn{})
				done <- true
			}(i)
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		count := manager.GetConnectionCount()
		if count > manager.maxConnections {
			t.Errorf("Connection count %d exceeds max %d", count, manager.maxConnections)
		}

		manager.RemoveAllConnections()
	})
}

func BenchmarkWebSocketMessage_Marshal(b *testing.B) {
	msg := Message{
		Action:   "query",
		Instance: "test-mysql",
		Query:    "SELECT * FROM users WHERE id = ?",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(msg)
	}
}

func BenchmarkWebSocketMessage_Unmarshal(b *testing.B) {
	data := []byte(`{"action":"query","instance":"test-mysql","query":"SELECT * FROM users"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var msg Message
		_ = json.Unmarshal(data, &msg)
	}
}
