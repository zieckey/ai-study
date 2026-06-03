package model

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) Next(_ context.Context, req Request) (Decision, error) {
	if len(req.Messages) == 0 {
		return Decision{}, fmt.Errorf("messages are required")
	}

	last := req.Messages[len(req.Messages)-1]
	if last.Role == RoleTool {
		return finalFromObservation(req.Messages), nil
	}

	goal := firstUserMessage(req.Messages)
	if expression := findExpression(goal); expression != "" {
		return toolCall("calculator", map[string]string{"expression": expression})
	}
	if asksForTime(goal) {
		return toolCall("clock", map[string]string{"format": "2006-01-02 15:04:05"})
	}
	if text := textToRepeat(goal); text != "" {
		return toolCall("echo", map[string]string{"text": text})
	}

	return Decision{Type: DecisionFinal, Answer: "我还没有识别出需要调用的工具。你可以试试：帮我计算 12 * 23、现在几点、请重复 hello agent。"}, nil
}

func toolCall(name string, args map[string]string) (Decision, error) {
	raw, err := json.Marshal(args)
	if err != nil {
		return Decision{}, err
	}
	return Decision{Type: DecisionToolCall, ToolName: name, Arguments: raw}, nil
}

func finalFromObservation(messages []Message) Decision {
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
	case "echo":
		return Decision{Type: DecisionFinal, Answer: last.Content}
	default:
		return Decision{Type: DecisionFinal, Answer: last.Content}
	}
}

func firstUserMessage(messages []Message) string {
	for _, message := range messages {
		if message.Role == RoleUser {
			return message.Content
		}
	}
	return ""
}

func asksForTime(goal string) bool {
	return strings.Contains(goal, "时间") || strings.Contains(goal, "几点") || strings.Contains(strings.ToLower(goal), "time")
}

func textToRepeat(goal string) string {
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

func findExpression(goal string) string {
	re := regexp.MustCompile(`-?\d+(?:\.\d+)?\s*[+\-*/]\s*-?\d+(?:\.\d+)?`)
	return strings.TrimSpace(re.FindString(goal))
}
