package agent

import "encoding/json"

type Config struct {
	MaxSteps        int
	SkillDir        string
	MemoryInContext bool
	MemoryContext   string
}

type Result struct {
	Answer string       `json:"answer"`
	Trace  []TraceEvent `json:"trace"`
}

type TraceEvent struct {
	Step        int             `json:"step"`
	Decision    string          `json:"decision"`
	ToolName    string          `json:"tool_name,omitempty"`
	ToolInput   json.RawMessage `json:"tool_input,omitempty"`
	Observation string          `json:"observation,omitempty"`
}
