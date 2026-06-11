package harness

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/zieckey/ai-study/internal/model"
	"github.com/zieckey/ai-study/internal/tools"
)

func TestHarnessRunCalculator(t *testing.T) {
	h, err := New(context.Background(), model.NewMockProvider(), []tools.Tool{tools.Calculator{}, tools.Clock{}, tools.Echo{}, tools.Weather{}}, Config{MaxSteps: 5})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	result, err := h.Run(context.Background(), "帮我计算 12 * 23")
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

func TestHarnessReturnsUnknownToolAsObservation(t *testing.T) {
	h, err := New(context.Background(), &sequenceProvider{decisions: []model.Decision{
		{Type: model.DecisionToolCall, ToolName: "missing", Arguments: json.RawMessage(`{}`)},
		{Type: model.DecisionFinal, Answer: "recovered"},
	}}, nil, Config{MaxSteps: 2})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	result, err := h.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Answer != "recovered" {
		t.Fatalf("Answer = %q", result.Answer)
	}
	if len(result.Trace) != 2 || !result.Trace[0].ToolError {
		t.Fatalf("Trace = %+v", result.Trace)
	}
}

func TestHarnessRunMaxSteps(t *testing.T) {
	h, err := New(context.Background(), staticProvider{decision: model.Decision{Type: model.DecisionToolCall, ToolName: "echo", Arguments: json.RawMessage(`{"text":"again"}`)}}, []tools.Tool{tools.Echo{}}, Config{MaxSteps: 1})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if _, err := h.Run(context.Background(), "test"); err == nil {
		t.Fatal("Run returned nil error")
	}
}

func TestHarnessRunMultipleToolCalls(t *testing.T) {
	h, err := New(context.Background(), &sequenceProvider{decisions: []model.Decision{
		{
			Type: model.DecisionToolCall,
			ToolCalls: []model.ToolCall{
				{ToolName: "echo", Arguments: json.RawMessage(`{"text":"one"}`)},
				{ToolName: "echo", Arguments: json.RawMessage(`{"text":"two"}`)},
			},
		},
		{Type: model.DecisionFinal, Answer: "done"},
	}}, []tools.Tool{tools.Echo{}}, Config{MaxSteps: 2})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	result, err := h.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Answer != "done" {
		t.Fatalf("Answer = %q", result.Answer)
	}
	if len(result.Trace) != 3 {
		t.Fatalf("Trace length = %d", len(result.Trace))
	}
}

func TestConfirmPolicyAllowAndDeny(t *testing.T) {
	call := model.ToolCall{ToolName: "echo", Arguments: json.RawMessage(`{"text":"ok"}`)}
	allow := ConfirmPolicy{Ask: map[string]bool{"echo": true}, Reader: strings.NewReader("y\n")}
	if err := allow.AllowTool(context.Background(), call); err != nil {
		t.Fatalf("AllowTool returned error: %v", err)
	}
	deny := ConfirmPolicy{Ask: map[string]bool{"echo": true}, Reader: strings.NewReader("n\n")}
	if err := deny.AllowTool(context.Background(), call); err == nil {
		t.Fatal("AllowTool returned nil error")
	}
}

func TestHarnessPolicyDenialAsObservation(t *testing.T) {
	h, err := New(context.Background(), &sequenceProvider{decisions: []model.Decision{
		{Type: model.DecisionToolCall, ToolName: "echo", Arguments: json.RawMessage(`{"text":"again"}`)},
		{Type: model.DecisionFinal, Answer: "policy recovered"},
	}}, []tools.Tool{tools.Echo{}}, Config{MaxSteps: 2, Policy: StaticPolicy{Denied: map[string]bool{"echo": true}}})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	result, err := h.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Answer != "policy recovered" {
		t.Fatalf("Answer = %q", result.Answer)
	}
	if !result.Trace[0].ToolError {
		t.Fatalf("Trace = %+v", result.Trace)
	}
}

func TestHarnessToolFailureAsObservation(t *testing.T) {
	h, err := New(context.Background(), &sequenceProvider{decisions: []model.Decision{
		{Type: model.DecisionToolCall, ToolName: "fail", Arguments: json.RawMessage(`{}`)},
		{Type: model.DecisionFinal, Answer: "tool recovered"},
	}}, []tools.Tool{failingTool{}}, Config{MaxSteps: 2})
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	result, err := h.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Answer != "tool recovered" {
		t.Fatalf("Answer = %q", result.Answer)
	}
	if !result.Trace[0].ToolError {
		t.Fatalf("Trace = %+v", result.Trace)
	}
}

type staticProvider struct {
	decision model.Decision
}

func (p staticProvider) Next(context.Context, model.Request) (model.Decision, error) {
	return p.decision, nil
}

type sequenceProvider struct {
	decisions []model.Decision
	index     int
}

func (p *sequenceProvider) Next(context.Context, model.Request) (model.Decision, error) {
	if p.index >= len(p.decisions) {
		return model.Decision{}, errors.New("no more decisions")
	}
	decision := p.decisions[p.index]
	p.index++
	return decision, nil
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
