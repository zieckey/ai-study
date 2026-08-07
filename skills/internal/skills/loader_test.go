package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseMarkdown(t *testing.T) {
	skill, err := ParseMarkdown(`---
name: agent-teacher
description: teach agents
---

Explain agent loops.
`)
	if err != nil {
		t.Fatalf("ParseMarkdown returned error: %v", err)
	}
	if skill.Name != "agent-teacher" {
		t.Fatalf("Name = %q", skill.Name)
	}
	if skill.Description != "teach agents" {
		t.Fatalf("Description = %q", skill.Description)
	}
	if skill.Content != "Explain agent loops." {
		t.Fatalf("Content = %q", skill.Content)
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agent-teacher.md")
	if err := os.WriteFile(path, []byte("---\nname: agent-teacher\ndescription: teach agents\n---\n\ncontent"), 0o644); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadDir(context.Background(), dir)
	if err != nil {
		t.Fatalf("LoadDir returned error: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("len(loaded) = %d", len(loaded))
	}
	if loaded[0].Name != "agent-teacher" {
		t.Fatalf("Name = %q", loaded[0].Name)
	}
}
