package model

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/zieckey/ai-study/internal/trace"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Next(ctx context.Context, req Request) (Decision, error) {
	trace.Log(ctx, "model.MockProvider.Next.start", map[string]any{"messages": req.Messages, "tools": req.Tools, "skills": req.Skills, "memory_context": req.MemoryContext})
	if len(req.Messages) == 0 {
		return Decision{}, fmt.Errorf("messages are required")
	}

	last := req.Messages[len(req.Messages)-1]
	if last.Role == RoleTool {
		decision := finalFromObservation(ctx, req.Messages)
		trace.Log(ctx, "model.MockProvider.Next.final_from_observation", map[string]any{"tool_name": last.ToolName, "answer_len": len(decision.Answer)})
		return decision, nil
	}

	goal := firstUserMessage(ctx, req.Messages)
	trace.Log(ctx, "model.MockProvider.Next.first_user_message", map[string]any{"goal": goal, "answer_len": len(goal)})
	if calls, err := toolCallsFromGoal(ctx, goal); err != nil {
		return Decision{}, err
	} else if len(calls) > 0 {
		return Decision{Type: DecisionToolCall, ToolCalls: calls}, nil
	}
	if len(req.Skills) > 0 {
		return Decision{Type: DecisionFinal, Answer: mockSkillAnswer(req.Skills)}, nil
	}

	answer := "我还没有识别出需要调用的工具。你可以试试：帮我计算 12 * 23、现在几点、请重复 hello agent。"
	trace.Log(ctx, "model.MockProvider.Next.fallback", map[string]any{"goal": goal, "answer_len": len(answer)})
	return Decision{Type: DecisionFinal, Answer: answer}, nil
}

func toolCallsFromGoal(ctx context.Context, goal string) ([]ToolCall, error) {
	trace.Log(ctx, "model.toolCallsFromGoal", map[string]any{"goal": goal})
	calls := []ToolCall{}
	for _, expression := range findExpressions(ctx, goal) {
		call, err := newToolCall(ctx, "calculator", map[string]string{"expression": expression})
		if err != nil {
			return nil, err
		}
		calls = append(calls, call)
	}
	if city := cityFromWeatherRequest(ctx, goal); city != "" {
		call, err := newToolCall(ctx, "weather", map[string]string{"city": city})
		if err != nil {
			return nil, err
		}
		calls = append(calls, call)
	}
	if asksForTime(ctx, goal) {
		call, err := newToolCall(ctx, "clock", map[string]string{"format": "2006-01-02 15:04:05"})
		if err != nil {
			return nil, err
		}
		calls = append(calls, call)
	}
	if text := textToRepeat(ctx, goal); text != "" {
		call, err := newToolCall(ctx, "echo", map[string]string{"text": text})
		if err != nil {
			return nil, err
		}
		calls = append(calls, call)
	}
	if args := memoryRequest(ctx, goal); args != nil {
		call, err := newToolCall(ctx, "memory", args)
		if err != nil {
			return nil, err
		}
		calls = append(calls, call)
	}
	return calls, nil
}

func newToolCall(ctx context.Context, name string, args map[string]string) (ToolCall, error) {
	trace.Log(ctx, "model.newToolCall", map[string]any{"tool_name": name, "arguments": args})
	raw, err := json.Marshal(args)
	if err != nil {
		return ToolCall{}, err
	}
	return ToolCall{ToolName: name, Arguments: raw}, nil
}

func toolCall(ctx context.Context, name string, args map[string]string) (Decision, error) {
	call, err := newToolCall(ctx, name, args)
	if err != nil {
		return Decision{}, err
	}
	return Decision{Type: DecisionToolCall, ToolCalls: []ToolCall{call}}, nil
}

func mockSkillAnswer(selected []SkillSpec) string {
	names := make([]string, 0, len(selected))
	for _, skill := range selected {
		names = append(names, skill.Name)
	}
	return fmt.Sprintf("mock provider 已选择相关 skills：%s。真实 provider 会把这些 skill 内容注入 system prompt，用来指导回答风格和任务策略。", strings.Join(names, ", "))
}

func finalFromObservation(ctx context.Context, messages []Message) Decision {
	trace.Log(ctx, "model.finalFromObservation", map[string]any{"messages": messages})
	last := messages[len(messages)-1]
	input := ""
	for i := len(messages) - 2; i >= 0; i-- {
		if messages[i].Role == RoleAssistant && messages[i].ToolName == last.ToolName {
			input = messages[i].ToolInput
			break
		}
	}

	switch last.ToolName {
	case "calculator":
		var args struct {
			Expression string `json:"expression"`
		}
		_ = json.Unmarshal([]byte(input), &args)
		if args.Expression != "" {
			return Decision{Type: DecisionFinal, Answer: fmt.Sprintf("%s = %s", args.Expression, last.Content)}
		}
		return Decision{Type: DecisionFinal, Answer: fmt.Sprintf("计算结果是 %s", last.Content)}
	case "clock":
		return Decision{Type: DecisionFinal, Answer: fmt.Sprintf("当前时间是 %s", last.Content)}
	case "weather":
		return Decision{Type: DecisionFinal, Answer: fmt.Sprintf("天气查询结果：%s", last.Content)}
	case "echo":
		if summary := summarizeMultipleToolResults(messages); summary != "" {
			return Decision{Type: DecisionFinal, Answer: summary}
		}
		return Decision{Type: DecisionFinal, Answer: last.Content}
	default:
		return Decision{Type: DecisionFinal, Answer: last.Content}
	}
}

func summarizeMultipleToolResults(messages []Message) string {
	toolResults := []string{}
	for _, message := range messages {
		if message.Role == RoleTool {
			switch message.ToolName {
			case "calculator":
				toolResults = append(toolResults, fmt.Sprintf("计算结果：%s", message.Content))
			case "weather":
				toolResults = append(toolResults, fmt.Sprintf("天气查询结果：%s", message.Content))
			case "clock":
				toolResults = append(toolResults, fmt.Sprintf("当前时间：%s", message.Content))
			case "echo":
				toolResults = append(toolResults, message.Content)
			}
		}
	}
	if len(toolResults) <= 1 {
		return ""
	}
	return strings.Join(toolResults, "\n")
}

func firstUserMessage(ctx context.Context, messages []Message) string {
	trace.Log(ctx, "model.firstUserMessage", map[string]any{"messages": messages})
	for _, message := range messages {
		if message.Role == RoleUser {
			return message.Content
		}
	}
	return ""
}

func cityFromWeatherRequest(ctx context.Context, goal string) string {
	trace.Log(ctx, "model.cityFromWeatherRequest", map[string]any{"goal": goal})
	if !strings.Contains(goal, "天气") && !strings.Contains(strings.ToLower(goal), "weather") {
		return ""
	}
	for _, city := range []string{"北京", "上海", "深圳", "杭州"} {
		if strings.Contains(goal, city) {
			return city
		}
	}
	return "北京"
}

func memoryRequest(ctx context.Context, goal string) map[string]string {
	trace.Log(ctx, "model.memoryRequest", map[string]any{"goal": goal})
	lower := strings.ToLower(goal)
	if strings.Contains(goal, "列出记忆") || strings.Contains(goal, "所有记忆") || strings.Contains(lower, "list memory") {
		return map[string]string{"action": "list"}
	}
	if strings.Contains(goal, "记住") || strings.Contains(goal, "保存记忆") || strings.Contains(lower, "remember") {
		text := strings.TrimSpace(goal)
		for _, marker := range []string{"请记住", "记住", "保存记忆", "remember"} {
			if idx := strings.Index(strings.ToLower(text), strings.ToLower(marker)); idx >= 0 {
				text = strings.TrimSpace(text[idx+len(marker):])
				break
			}
		}
		key, value := splitMemoryText(text)
		return map[string]string{"action": "set", "key": key, "value": value}
	}
	if strings.Contains(goal, "回忆") || strings.Contains(goal, "读取记忆") || strings.Contains(lower, "recall") {
		key := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(goal, "回忆"), "读取记忆"))
		if key == "" {
			return map[string]string{"action": "list"}
		}
		return map[string]string{"action": "get", "key": key}
	}
	if strings.Contains(goal, "删除记忆") || strings.Contains(lower, "delete memory") {
		key := strings.TrimSpace(strings.TrimPrefix(goal, "删除记忆"))
		return map[string]string{"action": "delete", "key": key}
	}
	return nil
}

func splitMemoryText(text string) (string, string) {
	text = strings.Trim(strings.TrimSpace(text), "：: ，,")
	for _, sep := range []string{"=", "：", ":", "是"} {
		if left, right, ok := strings.Cut(text, sep); ok {
			return strings.TrimSpace(left), strings.TrimSpace(right)
		}
	}
	if text == "" {
		return "note", ""
	}
	return text, text
}

func asksForTime(ctx context.Context, goal string) bool {
	matched := strings.Contains(goal, "时间") || strings.Contains(goal, "几点") || strings.Contains(strings.ToLower(goal), "time")
	trace.Log(ctx, "model.asksForTime", map[string]any{"goal": goal, "matched": matched})
	return matched
}

func textToRepeat(ctx context.Context, goal string) string {
	trace.Log(ctx, "model.textToRepeat", map[string]any{"goal": goal})
	for _, marker := range []string{"请重复", "重复", "echo"} {
		if idx := strings.Index(strings.ToLower(goal), strings.ToLower(marker)); idx >= 0 {
			text := strings.TrimSpace(goal[idx+len(marker):])
			text = strings.Trim(text, "：: ，,")
			text = trimFollowingTask(text)
			if text != "" {
				return text
			}
		}
	}
	return ""
}

func trimFollowingTask(text string) string {
	for _, marker := range []string{"，然后", ",然后", " 然后", "，计算", ",计算", "，查询", ",查询"} {
		if idx := strings.Index(text, marker); idx >= 0 {
			return strings.TrimSpace(strings.Trim(text[:idx], "：: ，,"))
		}
	}
	return text
}

func findExpression(ctx context.Context, goal string) string {
	expressions := findExpressions(ctx, goal)
	if len(expressions) == 0 {
		return ""
	}
	return expressions[0]
}

func findExpressions(ctx context.Context, goal string) []string {
	re := regexp.MustCompile(`-?\d+(?:\.\d+)?\s*[+\-*/]\s*-?\d+(?:\.\d+)?`)
	matches := re.FindAllString(goal, -1)
	expressions := make([]string, 0, len(matches))
	for _, match := range matches {
		expressions = append(expressions, strings.TrimSpace(match))
	}
	trace.Log(ctx, "model.findExpressions", map[string]any{"goal": goal, "expressions": expressions})
	return expressions
}
