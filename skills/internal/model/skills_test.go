package model

import (
	"strings"
	"testing"
)

func TestWithSkills(t *testing.T) {
	prompt := withSkills("base", []SkillSpec{{Name: "agent-teacher", Description: "teach", Content: "content"}})
	if !strings.Contains(prompt, "base") {
		t.Fatalf("prompt missing base: %q", prompt)
	}
	if !strings.Contains(prompt, `<skill name="agent-teacher">`) {
		t.Fatalf("prompt missing skill tag: %q", prompt)
	}
	if !strings.Contains(prompt, "content") {
		t.Fatalf("prompt missing content: %q", prompt)
	}
}
