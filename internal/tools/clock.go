package tools

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zieckey/ai-study/internal/trace"
)

type Clock struct {
	Now func() time.Time
}

func (Clock) Name() string {
	return "clock"
}

func (Clock) Description() string {
	return "返回当前本地时间"
}

func (Clock) InputSchema() string {
	return `{"type":"object","properties":{"format":{"type":"string","description":"可选的 Go time layout，例如 2006-01-02 15:04:05"}},"additionalProperties":false}`
}

func (c Clock) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	trace.Log(ctx, "tools.Clock.Execute.start", map[string]any{"input": input})
	var args struct {
		Format string `json:"format"`
	}
	if len(input) > 0 {
		if err := json.Unmarshal(input, &args); err != nil {
			return "", err
		}
	}
	if args.Format == "" {
		args.Format = time.RFC3339
	}

	now := time.Now
	if c.Now != nil {
		now = c.Now
	}
	formatted := now().Format(args.Format)
	trace.Log(ctx, "tools.Clock.Execute.done", map[string]any{"format": args.Format, "result": formatted})
	return formatted, nil
}
