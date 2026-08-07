package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zieckey/ai-study/internal/trace"
)

type Echo struct{}

func (Echo) Name() string {
	return "echo"
}

func (Echo) Description() string {
	return "原样返回输入文本"
}

func (Echo) InputSchema() string {
	return `{"type":"object","properties":{"text":{"type":"string","description":"需要原样返回的文本"}},"required":["text"],"additionalProperties":false}`
}

func (Echo) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	trace.Log(ctx, "tools.Echo.Execute.start", map[string]any{"input": input})
	var args struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("invalid echo input: %w", err)
	}
	trace.Log(ctx, "tools.Echo.Execute.done", map[string]any{"text": args.Text})
	return args.Text, nil
}
