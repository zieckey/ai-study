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
	Messages []Message   `json:"messages"`
	Tools    []ToolSpec  `json:"tools"`
	Skills   []SkillSpec `json:"skills,omitempty"`
}

type DecisionType string

const (
	DecisionToolCall DecisionType = "tool_call"
	DecisionFinal    DecisionType = "final"
)

type Decision struct {
	Type      DecisionType    `json:"type"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	ToolName  string          `json:"tool_name,omitempty"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
	Answer    string          `json:"answer,omitempty"`
}

type Provider interface {
	Next(ctx context.Context, req Request) (Decision, error)
}
