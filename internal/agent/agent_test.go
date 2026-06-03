package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/zieckey/ai-study/internal/model"
	"github.com/zieckey/ai-study/internal/tools"
)

func TestAgentRunCalculator(t *testing.T) {
	a, err := New(context.Background(), model.NewMockProvider(), []tools.Tool{tools.Calculator{}, tools.Clock{}, tools.Echo{}, tools.Weather{}}, Config{MaxSteps: 5})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	result, err := a.Run(context.Background(), "帮我计算 12 * 23")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Answer != "12 * 23 = 276" {
		t.Fatalf("Answer = %q", result.Answer)
	}
	if len(result.Trace) != 2 {
		t.Fatalf("Trace length = %d", len(result.Trace))
	}
	if result.Trace[0].Decision != "tool_call" || result.Trace[0].ToolName != "calculator" {
		t.Fatalf("first trace = %+v", result.Trace[0])
	}
	if result.Trace[1].Decision != "final" {
		t.Fatalf("second trace = %+v", result.Trace[1])
	}
}

func TestAgentRunUnknownTool(t *testing.T) {
	a, err := New(context.Background(), staticProvider{decision: model.Decision{Type: model.DecisionToolCall, ToolName: "missing", Arguments: json.RawMessage(`{}`)}}, nil, Config{MaxSteps: 1})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if _, err := a.Run(context.Background(), "test"); err == nil {
		t.Fatal("Run returned nil error")
	}
}

func TestAgentRunMaxSteps(t *testing.T) {
	a, err := New(context.Background(), staticProvider{decision: model.Decision{Type: model.DecisionToolCall, ToolName: "echo", Arguments: json.RawMessage(`{"text":"again"}`)}}, []tools.Tool{tools.Echo{}}, Config{MaxSteps: 1})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if _, err := a.Run(context.Background(), "test"); err == nil {
		t.Fatal("Run returned nil error")
	}
}

func TestAgentRunToolFailure(t *testing.T) {
	a, err := New(context.Background(), staticProvider{decision: model.Decision{Type: model.DecisionToolCall, ToolName: "fail", Arguments: json.RawMessage(`{}`)}}, []tools.Tool{failingTool{}}, Config{MaxSteps: 1})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if _, err := a.Run(context.Background(), "test"); err == nil {
		t.Fatal("Run returned nil error")
	}
}

type staticProvider struct {
	decision model.Decision
}

func (p staticProvider) Next(context.Context, model.Request) (model.Decision, error) {
	return p.decision, nil
}

type failingTool struct{}

func (failingTool) Name() string {
	return "fail"
}

func (failingTool) Description() string {
	return "always fails"
}

func (failingTool) InputSchema() string {
	return `{}`
}

func (failingTool) Execute(context.Context, json.RawMessage) (string, error) {
	return "", errors.New("boom")
}
