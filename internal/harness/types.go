package harness

import "encoding/json"

type Config struct {
	MaxSteps        int
	SkillDir        string
	MemoryInContext bool
	MemoryContext   string
	Policy          Policy
}

type Result struct {
	Answer string       `json:"answer"`
	Trace  []TraceEvent `json:"trace"`
}

type TraceEvent struct {
	Step        int             `json:"step"`
	Decision    string          `json:"decision"`
	ToolUseID   string          `json:"tool_use_id,omitempty"`
	ToolName    string          `json:"tool_name,omitempty"`
	ToolInput   json.RawMessage `json:"tool_input,omitempty"`
	Observation string          `json:"observation,omitempty"`
}
