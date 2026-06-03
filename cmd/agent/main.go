package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zieckey/ai-study/internal/agent"
	"github.com/zieckey/ai-study/internal/model"
	"github.com/zieckey/ai-study/internal/tools"
)

func main() {
	maxSteps := flag.Int("max-steps", 5, "agent 最大循环步数")
	provider := flag.String("provider", "mock", "模型 provider，目前支持 mock")
	trace := flag.Bool("trace", true, "是否打印执行过程")
	flag.Parse()

	goal := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if goal == "" {
		fmt.Fprintln(os.Stderr, "用法: go run ./cmd/agent [选项] \"帮我计算 12 * 23\"")
		flag.PrintDefaults()
		os.Exit(2)
	}

	modelProvider, err := buildProvider(*provider)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	a, err := agent.New(modelProvider, []tools.Tool{
		tools.Calculator{},
		tools.Clock{},
		tools.Echo{},
	}, agent.Config{MaxSteps: *maxSteps})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	result, err := a.Run(context.Background(), goal)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Goal: %s\n\n", goal)
	if *trace {
		printTrace(result.Trace)
	}
	fmt.Printf("Final Answer:\n%s\n", result.Answer)
}

func buildProvider(name string) (model.Provider, error) {
	switch name {
	case "mock":
		return model.NewMockProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported provider %q; 当前示例只实现 mock", name)
	}
}

func printTrace(trace []agent.TraceEvent) {
	for _, event := range trace {
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
