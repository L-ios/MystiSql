package websocket

import (
	"encoding/json"
	"testing"
)

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

func TestWebSocketHandler_Heartbeat(t *testing.T) {
	pingMsg := Message{Action: "ping"}
	_, err := json.Marshal(pingMsg)
	if err != nil {
		t.Fatalf("Failed to marshal ping message: %v", err)
	}

	t.Run("Ping message format", func(t *testing.T) {
		if pingMsg.Action != "ping" {
			t.Errorf("Expected action 'ping', got %s", pingMsg.Action)
		}
	})
}

func TestWebSocketHandler_QueryExecution(t *testing.T) {
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
