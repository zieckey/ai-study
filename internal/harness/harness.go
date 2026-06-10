package harness

import (
	"context"
	"errors"
	"fmt"

	"github.com/zieckey/ai-study/internal/model"
	"github.com/zieckey/ai-study/internal/skills"
	"github.com/zieckey/ai-study/internal/tools"
	"github.com/zieckey/ai-study/internal/trace"
)

type Harness struct {
	provider model.Provider
	registry *Registry
	skills   []skills.Skill
	policy   Policy
	config   Config
}

func New(ctx context.Context, provider model.Provider, registeredTools []tools.Tool, config Config) (*Harness, error) {
	trace.Log(ctx, "harness.New", map[string]any{"tools": len(registeredTools), "max_steps": config.MaxSteps, "skill_dir": config.SkillDir, "memory_in_context": config.MemoryInContext, "memory_context_len": len(config.MemoryContext)})
	if provider == nil {
		return nil, errors.New("model provider is required")
	}
	if config.MaxSteps <= 0 {
		config.MaxSteps = 5
	}
	if config.Policy == nil {
		config.Policy = AllowAllPolicy{}
	}

	registry, err := NewRegistry(ctx, registeredTools...)
	if err != nil {
		return nil, err
	}

	loadedSkills, err := skills.LoadDir(ctx, config.SkillDir)
	if err != nil {
		return nil, err
	}

	return &Harness{
		provider: provider,
		registry: registry,
		skills:   loadedSkills,
		policy:   config.Policy,
		config:   config,
	}, nil
}

func (h *Harness) Run(ctx context.Context, goal string) (Result, error) {
	trace.Log(ctx, "harness.Run.start", map[string]any{"goal": goal, "max_steps": h.config.MaxSteps, "tools": h.registry.Len(), "memory_in_context": h.config.MemoryInContext, "memory_context_len": len(h.config.MemoryContext)})
	messages := []model.Message{{Role: model.RoleUser, Content: goal}}
	selectedSkills := skills.Select(ctx, goal, h.skills)
	result := Result{}

	for step := 1; step <= h.config.MaxSteps; step++ {
		trace.Log(ctx, "harness.Run.step", map[string]any{"step": step, "messages": messages})
		decision, err := h.provider.Next(ctx, model.Request{
			Messages:      messages,
			Tools:         h.registry.Specs(ctx),
			Skills:        skillSpecs(selectedSkills),
			MemoryContext: h.memoryContext(),
		})
		if err != nil {
			return Result{}, fmt.Errorf("model decision failed: %w", err)
		}

		switch decision.Type {
		case model.DecisionFinal:
			trace.Log(ctx, "harness.Run.decision.final", map[string]any{"step": step, "answer_len": len(decision.Answer)})
			result.Answer = decision.Answer
			result.Trace = append(result.Trace, TraceEvent{Step: step, Decision: "final"})
			return result, nil
		case model.DecisionToolCall:
			calls := decision.Calls()
			trace.Log(ctx, "harness.Run.decision.tool_call", map[string]any{"step": step, "tool_calls": calls})
			if len(calls) == 0 {
				return Result{}, fmt.Errorf("tool_call decision did not include any tool calls")
			}

			assistantMessages := make([]model.Message, 0, len(calls))
			toolMessages := make([]model.Message, 0, len(calls))
			for _, call := range calls {
				observation, err := h.executeTool(ctx, step, call)
				if err != nil {
					return Result{}, err
				}

				result.Trace = append(result.Trace, TraceEvent{
					Step:        step,
					Decision:    "tool_call",
					ToolUseID:   call.ToolUseID,
					ToolName:    call.ToolName,
					ToolInput:   call.Arguments,
					Observation: observation,
				})
				assistantMessages = append(assistantMessages, model.Message{Role: model.RoleAssistant, Content: decision.Answer, ToolUseID: call.ToolUseID, ToolName: call.ToolName, ToolInput: string(call.Arguments)})
				toolMessages = append(toolMessages, model.Message{Role: model.RoleTool, Content: observation, ToolUseID: call.ToolUseID, ToolName: call.ToolName})
			}
			messages = append(messages, assistantMessages...)
			messages = append(messages, toolMessages...)
		default:
			return Result{}, fmt.Errorf("unknown decision type %q", decision.Type)
		}
	}

	return Result{}, fmt.Errorf("harness stopped after %d steps without final answer", h.config.MaxSteps)
}

func (h *Harness) executeTool(ctx context.Context, step int, call model.ToolCall) (string, error) {
	if err := h.policy.AllowTool(ctx, call); err != nil {
		trace.Log(ctx, "harness.Run.policy.denied", map[string]any{"step": step, "tool_use_id": call.ToolUseID, "tool_name": call.ToolName, "error": err.Error()})
		return "", err
	}

	tool, ok := h.registry.Get(call.ToolName)
	if !ok {
		return "", fmt.Errorf("unknown tool %q", call.ToolName)
	}

	observation, err := tool.Execute(ctx, call.Arguments)
	if err != nil {
		trace.Log(ctx, "harness.Run.tool.error", map[string]any{"step": step, "tool_use_id": call.ToolUseID, "tool_name": call.ToolName, "error": err.Error()})
		return "", fmt.Errorf("tool %q failed: %w", call.ToolName, err)
	}
	trace.Log(ctx, "harness.Run.tool.success", map[string]any{"step": step, "tool_use_id": call.ToolUseID, "tool_name": call.ToolName, "observation": observation})
	return observation, nil
}

func (h *Harness) memoryContext() string {
	if !h.config.MemoryInContext {
		return ""
	}
	return h.config.MemoryContext
}

func skillSpecs(selected []skills.Skill) []model.SkillSpec {
	specs := make([]model.SkillSpec, 0, len(selected))
	for _, skill := range selected {
		specs = append(specs, model.SkillSpec{Name: skill.Name, Description: skill.Description, Content: skill.Content})
	}
	return specs
}
