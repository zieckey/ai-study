package harness

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

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

type ConfirmPolicy struct {
	Ask    map[string]bool
	Reader io.Reader
	Writer io.Writer
}

func (p ConfirmPolicy) AllowTool(ctx context.Context, call model.ToolCall) error {
	trace.Log(ctx, "harness.ConfirmPolicy.AllowTool", map[string]any{"tool_name": call.ToolName, "tool_use_id": call.ToolUseID, "ask": p.Ask})
	if p.Ask == nil || !p.Ask[call.ToolName] {
		return nil
	}

	reader := p.Reader
	if reader == nil {
		reader = os.Stdin
	}
	writer := p.Writer
	if writer == nil {
		writer = os.Stdout
	}

	fmt.Fprintf(writer, "\nApprove tool call? tool=%s input=%s [y/N]: ", call.ToolName, string(call.Arguments))
	line, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil && err != io.EOF {
		return err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	if answer == "y" || answer == "yes" {
		return nil
	}
	return fmt.Errorf("user denied tool %q", call.ToolName)
}
