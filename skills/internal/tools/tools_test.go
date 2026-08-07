package tools

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestCalculatorExecute(t *testing.T) {
	tests := []struct {
		name       string
		expression string
		want       string
	}{
		{name: "add", expression: "1 + 2", want: "3"},
		{name: "multiply without spaces", expression: "12*23", want: "276"},
		{name: "divide", expression: "7 / 2", want: "3.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := json.Marshal(map[string]string{"expression": tt.expression})
			if err != nil {
				t.Fatal(err)
			}

			got, err := (Calculator{}).Execute(context.Background(), input)
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Execute() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCalculatorExecuteErrors(t *testing.T) {
	tests := []struct {
		name       string
		expression string
	}{
		{name: "division by zero", expression: "1 / 0"},
		{name: "invalid expression", expression: "1 + 2 + 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := json.Marshal(map[string]string{"expression": tt.expression})
			if err != nil {
				t.Fatal(err)
			}

			if _, err := (Calculator{}).Execute(context.Background(), input); err == nil {
				t.Fatal("Execute returned nil error")
			}
		})
	}
}

func TestClockExecute(t *testing.T) {
	clock := Clock{Now: func() time.Time {
		return time.Date(2026, 6, 3, 10, 11, 12, 0, time.UTC)
	}}
	input := json.RawMessage(`{"format":"2006-01-02 15:04:05"}`)

	got, err := clock.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got != "2026-06-03 10:11:12" {
		t.Fatalf("Execute() = %q", got)
	}
}

func TestEchoExecute(t *testing.T) {
	input := json.RawMessage(`{"text":"hello agent"}`)

	got, err := (Echo{}).Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got != "hello agent" {
		t.Fatalf("Execute() = %q", got)
	}
}

func TestWeatherExecute(t *testing.T) {
	input := json.RawMessage(`{"city":"北京"}`)

	got, err := (Weather{}).Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got != "北京：晴，25°C" {
		t.Fatalf("Execute() = %q", got)
	}
}

func TestWeatherExecuteRequiresCity(t *testing.T) {
	input := json.RawMessage(`{"city":""}`)

	if _, err := (Weather{}).Execute(context.Background(), input); err == nil {
		t.Fatal("Execute returned nil error")
	}
}

func TestRegistryRejectsDuplicateTools(t *testing.T) {
	_, err := Registry(context.Background(), Echo{}, Echo{})
	if err == nil {
		t.Fatal("Registry returned nil error")
	}
}
