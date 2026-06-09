package agent

import (
	"context"
	"errors"
	"fmt"

	"github.com/zieckey/ai-study/internal/model"
	"github.com/zieckey/ai-study/internal/skills"
	"github.com/zieckey/ai-study/internal/tools"
	"github.com/zieckey/ai-study/internal/trace"
)

type Agent struct {
	provider model.Provider
	tools    map[string]tools.Tool
	skills   []skills.Skill
	config   Config
}

func New(ctx context.Context, provider model.Provider, registeredTools []tools.Tool, config Config) (*Agent, error) {
	trace.Log(ctx, "agent.New", map[string]any{"tools": len(registeredTools), "max_steps": config.MaxSteps, "skill_dir": config.SkillDir, "memory_in_context": config.MemoryInContext, "memory_context_len": len(config.MemoryContext)})
	if provider == nil {
		return nil, errors.New("model provider is required")
	}
	if config.MaxSteps <= 0 {
		config.MaxSteps = 5
	}

	toolMap, err := tools.Registry(ctx, registeredTools...)
	if err != nil {
		return nil, err
	}

	loadedSkills, err := skills.LoadDir(ctx, config.SkillDir)
	if err != nil {
		return nil, err
	}

	return &Agent{
		provider: provider,
		tools:    toolMap,
		skills:   loadedSkills,
		config:   config,
	}, nil
}

func (a *Agent) Run(ctx context.Context, goal string) (Result, error) {
	trace.Log(ctx, "agent.Run.start", map[string]any{"goal": goal, "max_steps": a.config.MaxSteps, "tools": len(a.tools), "memory_in_context": a.config.MemoryInContext, "memory_context_len": len(a.config.MemoryContext)})
	messages := []model.Message{{Role: model.RoleUser, Content: goal}}
	selectedSkills := skills.Select(ctx, goal, a.skills)
	result := Result{}

	for step := 1; step <= a.config.MaxSteps; step++ {
		trace.Log(ctx, "\n\n\nagent.Run.step", map[string]any{"step": step, "messages": messages})
		decision, err := a.provider.Next(ctx, model.Request{
			Messages:      messages,
			Tools:         a.toolSpecs(ctx),
			Skills:        skillSpecs(selectedSkills),
			MemoryContext: a.memoryContext(),
		})
		if err != nil {
			return Result{}, fmt.Errorf("model decision failed: %w", err)
		}

		switch decision.Type {
		case model.DecisionFinal:
			trace.Log(ctx, "agent.Run.decision.final", map[string]any{"step": step, "answer_len": decision.Answer})
			result.Answer = decision.Answer
			result.Trace = append(result.Trace, TraceEvent{Step: step, Decision: "final"})
			return result, nil
		case model.DecisionToolCall:
			calls := decision.Calls()
			trace.Log(ctx, "agent.Run.decision.tool_call", map[string]any{"step": step, "tool_calls": calls})
			if len(calls) == 0 {
				return Result{}, fmt.Errorf("tool_call decision did not include any tool calls")
			}

			assistantMessages := make([]model.Message, 0, len(calls))
			toolMessages := make([]model.Message, 0, len(calls))
			for _, call := range calls {
				tool, ok := a.tools[call.ToolName]
				if !ok {
					return Result{}, fmt.Errorf("unknown tool %q", call.ToolName)
				}

				observation, err := tool.Execute(ctx, call.Arguments)
				if err != nil {
					trace.Log(ctx, "agent.Run.tool.error", map[string]any{"step": step, "tool_use_id": call.ToolUseID, "tool_name": call.ToolName, "error": err.Error()})
					return Result{}, fmt.Errorf("tool %q failed: %w", call.ToolName, err)
				}
				trace.Log(ctx, "agent.Run.tool.success", map[string]any{"step": step, "tool_use_id": call.ToolUseID, "tool_name": call.ToolName, "observation": observation})

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

	return Result{}, fmt.Errorf("agent stopped after %d steps without final answer", a.config.MaxSteps)
}

func (a *Agent) memoryContext() string {
	if !a.config.MemoryInContext {
		return ""
	}
	return a.config.MemoryContext
}

func skillSpecs(selected []skills.Skill) []model.SkillSpec {
	specs := make([]model.SkillSpec, 0, len(selected))
	for _, skill := range selected {
		specs = append(specs, model.SkillSpec{Name: skill.Name, Description: skill.Description, Content: skill.Content})
	}
	return specs
}

func (a *Agent) toolSpecs(ctx context.Context) []model.ToolSpec {
	trace.Log(ctx, "agent.toolSpecs", map[string]any{"tools": len(a.tools)})
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
