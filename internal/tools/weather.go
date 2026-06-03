package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type Weather struct{}

func (Weather) Name() string {
	return "weather"
}

func (Weather) Description() string {
	return "查询城市天气，当前返回 mock 数据"
}

func (Weather) InputSchema() string {
	return `{"type":"object","properties":{"city":{"type":"string","description":"城市名称，例如 北京、上海、深圳、杭州"}},"required":["city"],"additionalProperties":false}`
}

func (Weather) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var args struct {
		City string `json:"city"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("invalid weather input: %w", err)
	}
	if args.City == "" {
		return "", fmt.Errorf("city is required")
	}

	weatherByCity := map[string]string{
		"北京": "北京：晴，25°C",
		"上海": "上海：多云，27°C",
		"深圳": "深圳：阵雨，29°C",
		"杭州": "杭州：小雨，24°C",
	}
	if weather, ok := weatherByCity[args.City]; ok {
		return weather, nil
	}
	return fmt.Sprintf("%s：晴，25°C", args.City), nil
}
