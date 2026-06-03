package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zieckey/ai-study/internal/trace"
)

type Tool interface {
	Name() string
	Description() string
	InputSchema() string
	Execute(ctx context.Context, input json.RawMessage) (string, error)
}

func Registry(ctx context.Context, toolList ...Tool) (map[string]Tool, error) {
	trace.Log(ctx, "tools.Registry", map[string]any{"tools": len(toolList)})
	registry := make(map[string]Tool, len(toolList))
	for _, tool := range toolList {
		if tool == nil {
			return nil, fmt.Errorf("tool cannot be nil")
		}
		name := tool.Name()
		if name == "" {
			return nil, fmt.Errorf("tool name cannot be empty")
		}
		if _, exists := registry[name]; exists {
			return nil, fmt.Errorf("duplicate tool %q", name)
		}
		registry[name] = tool
	}
	return registry, nil
}
