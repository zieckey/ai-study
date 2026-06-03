package model

import (
	"context"
	"testing"
)

func TestMockProviderCalculatorDecision(t *testing.T) {
	decision, err := NewMockProvider().Next(context.Background(), Request{
		Messages: []Message{{Role: RoleUser, Content: "帮我计算 1 + 2"}},
	})
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if decision.Type != DecisionToolCall {
		t.Fatalf("Type = %q, want %q", decision.Type, DecisionToolCall)
	}
	if decision.ToolName != "calculator" {
		t.Fatalf("ToolName = %q", decision.ToolName)
	}
	if string(decision.Arguments) != `{"expression":"1 + 2"}` {
		t.Fatalf("Arguments = %s", string(decision.Arguments))
	}
}

func TestMockProviderClockDecision(t *testing.T) {
	decision, err := NewMockProvider().Next(context.Background(), Request{
		Messages: []Message{{Role: RoleUser, Content: "现在几点？"}},
	})
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if decision.Type != DecisionToolCall || decision.ToolName != "clock" {
		t.Fatalf("decision = %+v", decision)
	}
}

func TestMockProviderWeatherDecision(t *testing.T) {
	decision, err := NewMockProvider().Next(context.Background(), Request{
		Messages: []Message{{Role: RoleUser, Content: "查询上海天气"}},
	})
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if decision.Type != DecisionToolCall || decision.ToolName != "weather" {
		t.Fatalf("decision = %+v", decision)
	}
	if string(decision.Arguments) != `{"city":"上海"}` {
		t.Fatalf("Arguments = %s", string(decision.Arguments))
	}
}

func TestMockProviderFinalFromObservation(t *testing.T) {
	decision, err := NewMockProvider().Next(context.Background(), Request{
		Messages: []Message{
			{Role: RoleUser, Content: "帮我计算 1 + 2"},
			{Role: RoleAssistant, Content: "tool_call", ToolName: "calculator", ToolInput: `{"expression":"1 + 2"}`},
			{Role: RoleTool, ToolName: "calculator", Content: "3"},
		},
	})
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if decision.Type != DecisionFinal {
		t.Fatalf("Type = %q, want %q", decision.Type, DecisionFinal)
	}
	if decision.Answer != "1 + 2 = 3" {
		t.Fatalf("Answer = %q", decision.Answer)
	}
}
