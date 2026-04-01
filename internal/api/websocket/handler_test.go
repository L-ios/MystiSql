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
		{
			name: "Query message with requestId",
			message: Message{
				RequestID: "req-001",
				Action:    "query",
				Instance:  "test-mysql",
				Query:     "SELECT 1",
			},
			json: `{"requestId":"req-001","action":"query","instance":"test-mysql","query":"SELECT 1"}`,
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

func TestQueryResultMessage_CamelCase(t *testing.T) {
	msg := QueryResultMessage{
		RequestID:     "req-001",
		Type:          MessageTypeQueryResult,
		RowCount:      42,
		ExecutionTime: 150,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// 验证 camelCase JSON keys
	expected := []string{`"requestId"`, `"rowCount"`, `"executionTimeMs"`, `"type"`}
	for _, key := range expected {
		if !contains(string(data), key) {
			t.Errorf("Expected JSON to contain %s, got: %s", key, string(data))
		}
	}

	// 验证 snake_case 不存在
	forbidden := []string{`"request_id"`, `"row_count"`, `"execution_time_ms"`}
	for _, key := range forbidden {
		if contains(string(data), key) {
			t.Errorf("JSON should NOT contain snake_case key %s, got: %s", key, string(data))
		}
	}
}

func TestErrorMessage_CamelCase(t *testing.T) {
	msg := ErrorMessage{
		RequestID: "req-002",
		Type:      MessageTypeError,
		Code:      "INVALID_SQL",
		Message:   "syntax error",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	if !contains(string(data), `"requestId"`) {
		t.Errorf("Expected camelCase 'requestId', got: %s", string(data))
	}
	if contains(string(data), `"request_id"`) {
		t.Errorf("Should NOT contain snake_case 'request_id', got: %s", string(data))
	}
}

func TestMessage_UnmarshalCamelCase(t *testing.T) {
	input := `{"requestId":"req-003","action":"query","instance":"my-db","query":"SELECT 1"}`
	var msg Message
	if err := json.Unmarshal([]byte(input), &msg); err != nil {
		t.Fatalf("Failed to unmarshal camelCase: %v", err)
	}
	if msg.RequestID != "req-003" {
		t.Errorf("Expected RequestID 'req-003', got '%s'", msg.RequestID)
	}
	if msg.Action != "query" {
		t.Errorf("Expected Action 'query', got '%s'", msg.Action)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && jsonContains(s, substr)
}

func jsonContains(s, key string) bool {
	for i := 0; i <= len(s)-len(key); i++ {
		if s[i:i+len(key)] == key {
			return true
		}
	}
	return false
}
