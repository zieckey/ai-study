package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zieckey/ai-study/internal/trace"
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

func (Weather) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	trace.Log(ctx, "tools.Weather.Execute.start", map[string]any{"input": input})
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
		trace.Log(ctx, "tools.Weather.Execute.done", map[string]any{"city": args.City, "result": weather, "mock_hit": true})
		return weather, nil
	}
	weather := fmt.Sprintf("%s：晴，25°C", args.City)
	trace.Log(ctx, "tools.Weather.Execute.done", map[string]any{"city": args.City, "result": weather, "mock_hit": false})
	return weather, nil
}
