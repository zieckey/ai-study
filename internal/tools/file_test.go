package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileSearchExecute(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "internal", "harness.go"), []byte("package internal"), 0o644); err != nil {
		t.Fatal(err)
	}

	input, _ := json.Marshal(map[string]any{"query": "harness", "limit": 5})
	got, err := (FileSearch{Root: dir}).Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !strings.Contains(got, "internal/harness.go") {
		t.Fatalf("result = %q", got)
	}
}

func TestReadFileExecute(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello agent"), 0o644); err != nil {
		t.Fatal(err)
	}

	input, _ := json.Marshal(map[string]any{"path": "README.md"})
	got, err := (ReadFile{Root: dir}).Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got != "hello agent" {
		t.Fatalf("result = %q", got)
	}
}

func TestReadFileRejectsPathEscape(t *testing.T) {
	dir := t.TempDir()
	input, _ := json.Marshal(map[string]any{"path": "../secret.txt"})
	if _, err := (ReadFile{Root: dir}).Execute(context.Background(), input); err == nil {
		t.Fatal("Execute returned nil error")
	}
}
