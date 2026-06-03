package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/zieckey/ai-study/internal/model"
	"github.com/zieckey/ai-study/internal/tools"
)

type Agent struct {
	provider model.Provider
	tools    map[string]tools.Tool
	config   Config
}

func New(provider model.Provider, registeredTools []tools.Tool, config Config) (*Agent, error) {
	if provider == nil {
		return nil, errors.New("model provider is required")
	}
	if config.MaxSteps <= 0 {
		config.MaxSteps = 5
	}

	toolMap, err := tools.Registry(registeredTools...)
	if err != nil {
		return nil, err
	}

	return &Agent{
		provider: provider,
		tools:    toolMap,
		config:   config,
	}, nil
}

func (a *Agent) Run(ctx context.Context, goal string) (Result, error) {
	messages := []model.Message{{Role: model.RoleUser, Content: goal}}
	result := Result{}

	for step := 1; step <= a.config.MaxSteps; step++ {
		decision, err := a.provider.Next(ctx, model.Request{
			Messages: messages,
			Tools:    a.toolSpecs(),
		})
		if err != nil {
			return Result{}, fmt.Errorf("model decision failed: %w", err)
		}

		switch decision.Type {
		case model.DecisionFinal:
			result.Answer = decision.Answer
			result.Trace = append(result.Trace, TraceEvent{Step: step, Decision: "final"})
			return result, nil
		case model.DecisionToolCall:
			tool, ok := a.tools[decision.ToolName]
			if !ok {
				return Result{}, fmt.Errorf("unknown tool %q", decision.ToolName)
			}

			observation, err := tool.Execute(ctx, decision.Arguments)
			if err != nil {
				return Result{}, fmt.Errorf("tool %q failed: %w", decision.ToolName, err)
			}

			result.Trace = append(result.Trace, TraceEvent{
				Step:        step,
				Decision:    "tool_call",
				ToolName:    decision.ToolName,
				ToolInput:   decision.Arguments,
				Observation: observation,
			})
			messages = append(messages,
				model.Message{Role: model.RoleAssistant, Content: "tool_call", ToolName: decision.ToolName, ToolInput: string(decision.Arguments)},
				model.Message{Role: model.RoleTool, Content: observation, ToolName: decision.ToolName},
			)
		default:
			return Result{}, fmt.Errorf("unknown decision type %q", decision.Type)
		}
	}

	return Result{}, fmt.Errorf("agent stopped after %d steps without final answer", a.config.MaxSteps)
}

func (a *Agent) toolSpecs() []model.ToolSpec {
	specs := make([]model.ToolSpec, 0, len(a.tools))
	for _, tool := range a.tools {
		specs = append(specs, model.ToolSpec{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.InputSchema(),
		})
	}
	return specs
}
