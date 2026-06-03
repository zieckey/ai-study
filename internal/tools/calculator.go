package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

type Calculator struct{}

func (Calculator) Name() string {
	return "calculator"
}

func (Calculator) Description() string {
	return "计算一个简单的二元四则运算表达式，例如 12 * 23"
}

func (Calculator) InputSchema() string {
	return `{"type":"object","properties":{"expression":{"type":"string","description":"简单二元四则运算表达式，例如 12 * 23"}},"required":["expression"],"additionalProperties":false}`
}

func (Calculator) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Expression string `json:"expression"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("invalid calculator input: %w", err)
	}

	left, operator, right, err := parseExpression(args.Expression)
	if err != nil {
		return "", err
	}

	var result float64
	switch operator {
	case "+":
		result = left + right
	case "-":
		result = left - right
	case "*":
		result = left * right
	case "/":
		if right == 0 {
			return "", fmt.Errorf("division by zero")
		}
		result = left / right
	default:
		return "", fmt.Errorf("unsupported operator %q", operator)
	}

	return formatNumber(result), nil
}

func parseExpression(expression string) (float64, string, float64, error) {
	re := regexp.MustCompile(`^\s*(-?\d+(?:\.\d+)?)\s*([+\-*/])\s*(-?\d+(?:\.\d+)?)\s*$`)
	matches := re.FindStringSubmatch(expression)
	if len(matches) != 4 {
		return 0, "", 0, fmt.Errorf("expression must look like: number operator number")
	}

	left, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, "", 0, fmt.Errorf("invalid left number %q", matches[1])
	}
	right, err := strconv.ParseFloat(matches[3], 64)
	if err != nil {
		return 0, "", 0, fmt.Errorf("invalid right number %q", matches[3])
	}

	return left, matches[2], right, nil
}

func formatNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
