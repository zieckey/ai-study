package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zieckey/ai-study/internal/agent"
	"github.com/zieckey/ai-study/internal/model"
	"github.com/zieckey/ai-study/internal/tools"
	"github.com/zieckey/ai-study/internal/trace"
)

func main() {
	maxSteps := flag.Int("max-steps", 5, "agent 最大循环步数")
	provider := flag.String("provider", "mock", "模型 provider，支持 mock、anthropic 或 deepseek")
	anthropicModel := flag.String("anthropic-model", "claude-opus-4-7", "Anthropic provider 使用的 Claude 模型")
	anthropicMaxTokens := flag.Int64("anthropic-max-tokens", 4096, "Anthropic provider 单次模型响应的最大 token 数")
	deepSeekModel := flag.String("deepseek-model", "deepseek-chat", "DeepSeek provider 使用的模型")
	deepSeekMaxTokens := flag.Int64("deepseek-max-tokens", 4096, "DeepSeek provider 单次模型响应的最大 token 数")
	showTrace := flag.Bool("trace", true, "是否打印执行过程")
	traceLog := flag.Bool("trace-log", true, "是否打印函数级 trace log 到 stderr")
	jsonOutput := flag.Bool("json", false, "以 JSON 格式输出 goal、trace 和 answer")
	flag.Parse()

	goal := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if goal == "" {
		fmt.Fprintln(os.Stderr, "用法: go run ./cmd/agent [选项] \"帮我计算 12 * 23\"")
		flag.PrintDefaults()
		os.Exit(2)
	}

	logger := trace.NewLogger(os.Stdout, *traceLog)
	ctx := trace.WithLogger(context.Background(), logger)
	trace.Log(ctx, "main.main", map[string]any{
		"provider":                 *provider,
		"anthropic_model":          *anthropicModel,
		"anthropic_max_tokens":     *anthropicMaxTokens,
		"deepseek_model":           *deepSeekModel,
		"deepseek_max_tokens":      *deepSeekMaxTokens,
		"max_steps":                *maxSteps,
		"show_trace":               *showTrace,
		"json_output":              *jsonOutput,
		"goal":                     goal,
		"anthropic_api_key_set":    os.Getenv("ANTHROPIC_API_KEY") != "",
		"deepseek_api_key_set":     os.Getenv("DEEPSEEK_API_KEY") != "",
		"deepseek_base_url_custom": os.Getenv("DEEPSEEK_BASE_URL") != "",
	})

	modelProvider, err := buildProvider(ctx, *provider, *anthropicModel, *anthropicMaxTokens, *deepSeekModel, *deepSeekMaxTokens)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	a, err := agent.New(ctx, modelProvider, []tools.Tool{
		tools.Calculator{},
		tools.Clock{},
		tools.Echo{},
		tools.Weather{},
	}, agent.Config{MaxSteps: *maxSteps})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	result, err := a.Run(ctx, goal)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *jsonOutput {
		if err := printJSON(ctx, goal, result); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("Goal: %s\n\n", goal)
	if *showTrace {
		printTrace(ctx, result.Trace)
	}
	fmt.Printf("Final Answer:\n%s\n", result.Answer)
}

func buildProvider(ctx context.Context, name string, anthropicModel string, anthropicMaxTokens int64, deepSeekModel string, deepSeekMaxTokens int64) (model.Provider, error) {
	trace.Log(ctx, "main.buildProvider", map[string]any{
		"provider":             name,
		"anthropic_model":      anthropicModel,
		"anthropic_max_tokens": anthropicMaxTokens,
		"deepseek_model":       deepSeekModel,
		"deepseek_max_tokens":  deepSeekMaxTokens,
	})

	switch name {
	case "mock":
		return model.NewMockProvider(), nil
	case "anthropic", "claude":
		return model.NewAnthropicProvider(model.AnthropicConfig{
			Model:     anthropicModel,
			MaxTokens: anthropicMaxTokens,
		}), nil
	case "deepseek":
		return model.NewDeepSeekProvider(model.DeepSeekConfig{
			Model:     deepSeekModel,
			MaxTokens: deepSeekMaxTokens,
		}), nil
	default:
		return nil, fmt.Errorf("unsupported provider %q; 支持 mock、anthropic 或 deepseek", name)
	}
}

func printTrace(ctx context.Context, traceEvents []agent.TraceEvent) {
	trace.Log(ctx, "main.printTrace", map[string]any{"events": len(traceEvents)})
	for _, event := range traceEvents {
		fmt.Printf("Step %d\n", event.Step)
		switch event.Decision {
		case "tool_call":
			fmt.Printf("Model decision: call tool %s\n", event.ToolName)
			fmt.Printf("Tool input: %s\n", string(event.ToolInput))
			fmt.Printf("Observation: %s\n\n", event.Observation)
		case "final":
			fmt.Printf("Model decision: final answer\n\n")
		}
	}
}

func printJSON(ctx context.Context, goal string, result agent.Result) error {
	trace.Log(ctx, "main.printJSON", map[string]any{"goal": goal, "trace_events": len(result.Trace), "answer_len": len(result.Answer)})
	output := struct {
		Goal string `json:"goal"`
		agent.Result
	}{
		Goal:   goal,
		Result: result,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
