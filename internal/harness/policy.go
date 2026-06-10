package harness

import (
	"context"
	"fmt"

	"github.com/zieckey/ai-study/internal/model"
	"github.com/zieckey/ai-study/internal/trace"
)

type Policy interface {
	AllowTool(ctx context.Context, call model.ToolCall) error
}

type AllowAllPolicy struct{}

func (AllowAllPolicy) AllowTool(ctx context.Context, call model.ToolCall) error {
	trace.Log(ctx, "harness.AllowAllPolicy.AllowTool", map[string]any{"tool_name": call.ToolName, "tool_use_id": call.ToolUseID})
	return nil
}

type StaticPolicy struct {
	Denied map[string]bool
}

func (p StaticPolicy) AllowTool(ctx context.Context, call model.ToolCall) error {
	trace.Log(ctx, "harness.StaticPolicy.AllowTool", map[string]any{"tool_name": call.ToolName, "tool_use_id": call.ToolUseID})
	if p.Denied != nil && p.Denied[call.ToolName] {
		return fmt.Errorf("tool %q is denied by harness policy", call.ToolName)
	}
	return nil
}
