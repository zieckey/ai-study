package harness

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	EventModelRequest  EventType = "model_request"
	EventModelResponse EventType = "model_response"
	EventToolCall      EventType = "tool_call"
	EventToolResult    EventType = "tool_result"
	EventFinalAnswer   EventType = "final_answer"
)

type Event struct {
	Type       EventType       `json:"type"`
	Step       int             `json:"step"`
	ToolUseID  string          `json:"tool_use_id,omitempty"`
	ToolName   string          `json:"tool_name,omitempty"`
	Input      json.RawMessage `json:"input,omitempty"`
	Output     string          `json:"output,omitempty"`
	Error      string          `json:"error,omitempty"`
	OccurredAt string          `json:"occurred_at"`
}

func newEvent(eventType EventType, step int) Event {
	return Event{Type: eventType, Step: step, OccurredAt: time.Now().Format(time.RFC3339)}
}
