package model

import "testing"

func TestToAnthropicTools(t *testing.T) {
	tools, err := toAnthropicTools([]ToolSpec{{
		Name:        "weather",
		Description: "查询城市天气",
		InputSchema: `{"type":"object","properties":{"city":{"type":"string"}},"required":["city"],"additionalProperties":false}`,
	}})
	if err != nil {
		t.Fatalf("toAnthropicTools returned error: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("len(tools) = %d", len(tools))
	}
}

func TestToAnthropicMessagesRequiresToolUseID(t *testing.T) {
	_, err := toAnthropicMessages([]Message{{Role: RoleTool, ToolName: "weather", Content: "北京：晴，25°C"}})
	if err == nil {
		t.Fatal("toAnthropicMessages returned nil error")
	}
}

func TestToAnthropicMessagesWithToolResult(t *testing.T) {
	_, err := toAnthropicMessages([]Message{
		{Role: RoleUser, Content: "查询北京天气"},
		{Role: RoleAssistant, ToolUseID: "toolu_test", ToolName: "weather", ToolInput: `{"city":"北京"}`},
		{Role: RoleTool, ToolUseID: "toolu_test", ToolName: "weather", Content: "北京：晴，25°C"},
	})
	if err != nil {
		t.Fatalf("toAnthropicMessages returned error: %v", err)
	}
}
