package model

import (
	"context"
	"encoding/json"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type Message struct {
	Role      Role   `json:"role"`
	Content   string `json:"content"`
	ToolUseID string `json:"tool_use_id,omitempty"`
	ToolName  string `json:"tool_name,omitempty"`
	ToolInput string `json:"tool_input,omitempty"`
	ToolError bool   `json:"tool_error,omitempty"`
}

type ToolSpec struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema string `json:"input_schema"`
}

type SkillSpec struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

type Request struct {
	Messages      []Message   `json:"messages"`
	Tools         []ToolSpec  `json:"tools"`
	Skills        []SkillSpec `json:"skills,omitempty"`
	MemoryContext string      `json:"memory_context,omitempty"`
}

type DecisionType string

const (
	DecisionToolCall DecisionType = "tool_call"
	DecisionFinal    DecisionType = "final"
)

type ToolCall struct {
	ToolUseID string          `json:"tool_use_id,omitempty"`
	ToolName  string          `json:"tool_name"`
	Arguments json.RawMessage `json:"arguments"`
}

type Decision struct {
	Type      DecisionType `json:"type"`
	ToolCalls []ToolCall   `json:"tool_calls,omitempty"`
	Answer    string       `json:"answer,omitempty"`
}

func (d Decision) Calls() []ToolCall {
	return d.ToolCalls
}

type Provider interface {
	Next(ctx context.Context, req Request) (Decision, error)
}
