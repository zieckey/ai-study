package model

import (
	"strings"
	"testing"
)

func TestWithMemory(t *testing.T) {
	prompt := withMemory("base", "- language = Go")
	if !strings.Contains(prompt, "base") {
		t.Fatalf("prompt missing base: %q", prompt)
	}
	if !strings.Contains(prompt, "<memory>") {
		t.Fatalf("prompt missing memory tag: %q", prompt)
	}
	if !strings.Contains(prompt, "language = Go") {
		t.Fatalf("prompt missing memory content: %q", prompt)
	}
}

func TestWithMemoryEmpty(t *testing.T) {
	prompt := withMemory("base", "")
	if prompt != "base" {
		t.Fatalf("prompt = %q", prompt)
	}
}
