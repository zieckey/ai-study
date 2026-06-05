package skills

import (
	"context"
	"testing"
)

func TestSelectAgentTeacher(t *testing.T) {
	selected := Select(context.Background(), "DeepSeek 的 tool_calls 是什么", []Skill{
		{Name: "agent-teacher"},
		{Name: "math-tutor"},
	})
	if len(selected) != 1 || selected[0].Name != "agent-teacher" {
		t.Fatalf("selected = %+v", selected)
	}
}

func TestSelectMathTutor(t *testing.T) {
	selected := Select(context.Background(), "请讲解 12 * 23 怎么计算", []Skill{
		{Name: "agent-teacher"},
		{Name: "math-tutor"},
	})
	if len(selected) != 1 || selected[0].Name != "math-tutor" {
		t.Fatalf("selected = %+v", selected)
	}
}

func TestFormatForPrompt(t *testing.T) {
	prompt := FormatForPrompt([]Skill{{Name: "agent-teacher", Description: "teach", Content: "content"}})
	if prompt == "" {
		t.Fatal("prompt is empty")
	}
}
