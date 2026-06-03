package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type Echo struct{}

func (Echo) Name() string {
	return "echo"
}

func (Echo) Description() string {
	return "原样返回输入文本"
}

func (Echo) InputSchema() string {
	return `{"text":"string"}`
}

func (Echo) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("invalid echo input: %w", err)
	}
	return args.Text, nil
}
