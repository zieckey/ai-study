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
	calls := decision.Calls()
	if len(calls) != 1 {
		t.Fatalf("len(calls) = %d", len(calls))
	}
	if calls[0].ToolName != "calculator" {
		t.Fatalf("ToolName = %q", calls[0].ToolName)
	}
	if string(calls[0].Arguments) != `{"expression":"1 + 2"}` {
		t.Fatalf("Arguments = %s", string(calls[0].Arguments))
	}
}

func TestMockProviderMultipleToolCalls(t *testing.T) {
	decision, err := NewMockProvider().Next(context.Background(), Request{
		Messages: []Message{{Role: RoleUser, Content: "请重复 confirm ready，计算 5*6"}},
	})
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	calls := decision.Calls()
	if len(calls) != 2 {
		t.Fatalf("len(calls) = %d, decision = %+v", len(calls), decision)
	}
	if calls[0].ToolName != "calculator" {
		t.Fatalf("first call = %+v", calls[0])
	}
	if calls[1].ToolName != "echo" {
		t.Fatalf("second call = %+v", calls[1])
	}
}

func TestMockProviderClockDecision(t *testing.T) {
	decision, err := NewMockProvider().Next(context.Background(), Request{
		Messages: []Message{{Role: RoleUser, Content: "现在几点？"}},
	})
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	calls := decision.Calls()
	if decision.Type != DecisionToolCall || len(calls) != 1 || calls[0].ToolName != "clock" {
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
	calls := decision.Calls()
	if decision.Type != DecisionToolCall || len(calls) != 1 || calls[0].ToolName != "weather" {
		t.Fatalf("decision = %+v", decision)
	}
	if string(calls[0].Arguments) != `{"city":"上海"}` {
		t.Fatalf("Arguments = %s", string(calls[0].Arguments))
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
