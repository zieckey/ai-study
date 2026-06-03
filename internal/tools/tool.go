package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

type Tool interface {
	Name() string
	Description() string
	InputSchema() string
	Execute(ctx context.Context, input json.RawMessage) (string, error)
}

func Registry(toolList ...Tool) (map[string]Tool, error) {
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
