package tools

import (
	"context"
	"encoding/json"
	"time"
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
	return `{"format":"string, optional Go time layout"}`
}

func (c Clock) Execute(_ context.Context, input json.RawMessage) (string, error) {
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
	return now().Format(args.Format), nil
}
