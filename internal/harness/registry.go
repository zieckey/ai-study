package harness

import (
	"context"
	"fmt"

	"github.com/zieckey/ai-study/internal/model"
	"github.com/zieckey/ai-study/internal/tools"
	"github.com/zieckey/ai-study/internal/trace"
)

type Registry struct {
	tools map[string]tools.Tool
}

func NewRegistry(ctx context.Context, toolList ...tools.Tool) (*Registry, error) {
	trace.Log(ctx, "harness.NewRegistry", map[string]any{"tools": len(toolList)})
	registry := make(map[string]tools.Tool, len(toolList))
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
	return &Registry{tools: registry}, nil
}

func (r *Registry) Get(name string) (tools.Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *Registry) Len() int {
	if r == nil {
		return 0
	}
	return len(r.tools)
}

func (r *Registry) Specs(ctx context.Context) []model.ToolSpec {
	trace.Log(ctx, "harness.Registry.Specs", map[string]any{"tools": r.Len()})
	specs := make([]model.ToolSpec, 0, len(r.tools))
	for _, tool := range r.tools {
		specs = append(specs, model.ToolSpec{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.InputSchema(),
		})
	}
	return specs
}
