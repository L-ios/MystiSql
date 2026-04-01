package websocket

import "encoding/json"

type MessageType string

const (
	MessageTypeQuery       MessageType = "query"
	MessageTypeQueryResult MessageType = "query_result"
	MessageTypeError       MessageType = "error"
	MessageTypePong        MessageType = "pong"
	MessageTypePing        MessageType = "ping"
)

type Message struct {
	RequestID string      `json:"requestId,omitempty"`
	Action    MessageType `json:"action"`
	Instance  string      `json:"instance,omitempty"`
	Query     string      `json:"query,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
}

func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

type QueryResultMessage struct {
	RequestID     string                   `json:"requestId"`
	Type          MessageType              `json:"type"`
	Columns       []map[string]interface{} `json:"columns"`
	Rows          []map[string]interface{} `json:"rows"`
	RowCount      int                      `json:"rowCount"`
	ExecutionTime int64                    `json:"executionTimeMs"`
}

type ErrorMessage struct {
	RequestID string      `json:"requestId,omitempty"`
	Type      MessageType `json:"type"`
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   string      `json:"details,omitempty"`
}

type PongMessage struct {
	Type      MessageType `json:"type"`
	Timestamp int64       `json:"timestamp"`
}
