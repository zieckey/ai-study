package model

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestToDeepSeekTools(t *testing.T) {
	tools, err := toDeepSeekTools([]ToolSpec{{
		Name:        "weather",
		Description: "查询城市天气",
		InputSchema: `{"type":"object","properties":{"city":{"type":"string"}},"required":["city"],"additionalProperties":false}`,
	}})
	if err != nil {
		t.Fatalf("toDeepSeekTools returned error: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("len(tools) = %d", len(tools))
	}
	if tools[0].Function.Name != "weather" {
		t.Fatalf("tool name = %q", tools[0].Function.Name)
	}
}

func TestToDeepSeekMessagesWithToolResult(t *testing.T) {
	messages, err := toDeepSeekMessages([]Message{
		{Role: RoleUser, Content: "查询北京天气"},
		{Role: RoleAssistant, ToolUseID: "call_1", ToolName: "weather", ToolInput: `{"city":"北京"}`},
		{Role: RoleTool, ToolUseID: "call_1", ToolName: "weather", Content: "北京：晴，25°C"},
	}, "system prompt")
	if err != nil {
		t.Fatalf("toDeepSeekMessages returned error: %v", err)
	}
	if len(messages) != 4 {
		t.Fatalf("len(messages) = %d", len(messages))
	}
	if messages[2].ToolCalls[0].ID != "call_1" {
		t.Fatalf("tool call id = %q", messages[2].ToolCalls[0].ID)
	}
	if messages[3].ToolCallID != "call_1" {
		t.Fatalf("tool result id = %q", messages[3].ToolCallID)
	}
}

func TestToDeepSeekMessagesRequiresToolUseID(t *testing.T) {
	_, err := toDeepSeekMessages([]Message{{Role: RoleTool, ToolName: "weather", Content: "北京：晴，25°C"}}, "system prompt")
	if err == nil {
		t.Fatal("toDeepSeekMessages returned nil error")
	}
}

func TestDeepSeekProviderNextToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}

		var req deepSeekChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Model != "deepseek-chat" {
			t.Fatalf("model = %q", req.Model)
		}
		if len(req.Tools) != 1 {
			t.Fatalf("len(tools) = %d", len(req.Tools))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","tool_calls":[{"id":"call_1","type":"function","function":{"name":"weather","arguments":"{\"city\":\"北京\"}"}}]}}]}`))
	}))
	defer server.Close()

	provider := NewDeepSeekProvider(DeepSeekConfig{APIKey: "test-key", BaseURL: server.URL})
	decision, err := provider.Next(context.Background(), Request{
		Messages: []Message{{Role: RoleUser, Content: "查询北京天气"}},
		Tools: []ToolSpec{{
			Name:        "weather",
			Description: "查询城市天气",
			InputSchema: `{"type":"object","properties":{"city":{"type":"string"}},"required":["city"],"additionalProperties":false}`,
		}},
	})
	if err != nil {
		t.Fatalf("Next returned error: %v", err)
	}
	if decision.Type != DecisionToolCall || decision.ToolUseID != "call_1" || decision.ToolName != "weather" {
		t.Fatalf("decision = %+v", decision)
	}
	if string(decision.Arguments) != `{"city":"北京"}` {
		t.Fatalf("arguments = %s", string(decision.Arguments))
	}
}
