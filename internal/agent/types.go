package agent

import "encoding/json"

type Config struct {
	MaxSteps int
}

type Result struct {
	Answer string
	Trace  []TraceEvent
}

type TraceEvent struct {
	Step        int
	Decision    string
	ToolName    string
	ToolInput   json.RawMessage
	Observation string
}
