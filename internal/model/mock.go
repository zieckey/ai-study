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
	trace.Log(ctx, "model.MockProvider.Next.start", map[string]any{"messages": len(req.Messages), "tools": len(req.Tools)})
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
	if expression := findExpression(ctx, goal); expression != "" {
		return toolCall(ctx, "calculator", map[string]string{"expression": expression})
	}
	if city := cityFromWeatherRequest(ctx, goal); city != "" {
		return toolCall(ctx, "weather", map[string]string{"city": city})
	}
	if asksForTime(ctx, goal) {
		return toolCall(ctx, "clock", map[string]string{"format": "2006-01-02 15:04:05"})
	}
	if text := textToRepeat(ctx, goal); text != "" {
		return toolCall(ctx, "echo", map[string]string{"text": text})
	}

	answer := "我还没有识别出需要调用的工具。你可以试试：帮我计算 12 * 23、现在几点、请重复 hello agent。"
	trace.Log(ctx, "model.MockProvider.Next.fallback", map[string]any{"goal": goal, "answer_len": len(answer)})
	return Decision{Type: DecisionFinal, Answer: answer}, nil
}

func toolCall(ctx context.Context, name string, args map[string]string) (Decision, error) {
	trace.Log(ctx, "model.toolCall", map[string]any{"tool_name": name, "arguments": args})
	raw, err := json.Marshal(args)
	if err != nil {
		return Decision{}, err
	}
	return Decision{Type: DecisionToolCall, ToolName: name, Arguments: raw}, nil
}

func finalFromObservation(ctx context.Context, messages []Message) Decision {
	trace.Log(ctx, "model.finalFromObservation", map[string]any{"messages": len(messages)})
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
		return Decision{Type: DecisionFinal, Answer: last.Content}
	default:
		return Decision{Type: DecisionFinal, Answer: last.Content}
	}
}

func firstUserMessage(ctx context.Context, messages []Message) string {
	trace.Log(ctx, "model.firstUserMessage", map[string]any{"messages": len(messages)})
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
			if text != "" {
				return text
			}
		}
	}
	return ""
}

func findExpression(ctx context.Context, goal string) string {
	re := regexp.MustCompile(`-?\d+(?:\.\d+)?\s*[+\-*/]\s*-?\d+(?:\.\d+)?`)
	expression := strings.TrimSpace(re.FindString(goal))
	trace.Log(ctx, "model.findExpression", map[string]any{"goal": goal, "expression": expression})
	return expression
}
