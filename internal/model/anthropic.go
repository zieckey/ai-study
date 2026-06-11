package model

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/zieckey/ai-study/internal/trace"
)

const defaultAnthropicSystemPrompt = `你是这个 Go 学习项目里的真实 Claude provider。
你可以使用工具完成任务：calculator 负责精确计算，clock 负责获取当前时间，weather 返回 mock 天气，echo 原样返回文本，memory 负责读写本地持久化记忆，file_search 负责在项目内搜索文件，read_file 负责读取项目内文本文件。
当用户要求记住、保存、回忆、列出或删除偏好/事实时，请调用 memory 工具。不要把密码、API Key、token 等敏感信息写入记忆。
当用户的问题需要外部实时信息、确定性计算或项目提供的工具能力时，请优先调用合适的工具；工具返回 observation 后，再给出简洁的中文最终答案。`

type AnthropicConfig struct {
	Model        string
	MaxTokens    int64
	SystemPrompt string
	Effort       anthropic.OutputConfigEffort
}

type AnthropicProvider struct {
	client       anthropic.Client
	model        string
	maxTokens    int64
	systemPrompt string
	effort       anthropic.OutputConfigEffort
}

func NewAnthropicProvider(config AnthropicConfig) *AnthropicProvider {
	if config.Model == "" {
		config.Model = anthropic.ModelClaudeOpus4_7
	}
	if config.MaxTokens <= 0 {
		config.MaxTokens = 4096
	}
	if config.SystemPrompt == "" {
		config.SystemPrompt = defaultAnthropicSystemPrompt
	}
	if config.Effort == "" {
		config.Effort = anthropic.OutputConfigEffortHigh
	}

	return &AnthropicProvider{
		client:       anthropic.NewClient(),
		model:        config.Model,
		maxTokens:    config.MaxTokens,
		systemPrompt: config.SystemPrompt,
		effort:       config.Effort,
	}
}

func (p *AnthropicProvider) Next(ctx context.Context, req Request) (Decision, error) {
	trace.Log(ctx, "model.AnthropicProvider.Next.start", map[string]any{"model": p.model, "max_tokens": p.maxTokens, "effort": p.effort, "messages": len(req.Messages), "tools": len(req.Tools), "skills": len(req.Skills)})
	messages, err := toAnthropicMessages(ctx, req.Messages)
	if err != nil {
		return Decision{}, err
	}
	anthropicTools, err := toAnthropicTools(ctx, req.Tools)
	if err != nil {
		return Decision{}, err
	}

	adaptive := anthropic.ThinkingConfigAdaptiveParam{}
	systemPrompt := withMemory(withSkills(p.systemPrompt, req.Skills), req.MemoryContext)
	trace.Log(ctx, "model.AnthropicProvider.Next.request", map[string]any{"model": p.model, "messages": len(messages), "tools": len(anthropicTools), "skills": len(req.Skills), "memory_context_len": len(req.MemoryContext), "system_prompt_len": len(systemPrompt)})
	resp, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     p.model,
		MaxTokens: p.maxTokens,
		System: []anthropic.TextBlockParam{{
			Text:         systemPrompt,
			CacheControl: anthropic.NewCacheControlEphemeralParam(),
		}},
		Thinking: anthropic.ThinkingConfigParamUnion{OfAdaptive: &adaptive},
		OutputConfig: anthropic.OutputConfigParam{
			Effort: p.effort,
		},
		Messages: messages,
		Tools:    anthropicTools,
		ToolChoice: anthropic.ToolChoiceUnionParam{OfAuto: &anthropic.ToolChoiceAutoParam{
			DisableParallelToolUse: anthropic.Bool(true),
		}},
	})
	if err != nil {
		trace.Log(ctx, "model.AnthropicProvider.Next.error", map[string]any{"error": err.Error()})
		return Decision{}, err
	}
	trace.Log(ctx, "model.AnthropicProvider.Next.response", map[string]any{"content_blocks": len(resp.Content), "stop_reason": resp.StopReason})

	var textParts []string
	var calls []ToolCall
	for _, block := range resp.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.ToolUseBlock:
			trace.Log(ctx, "model.AnthropicProvider.Next.tool_use", map[string]any{"tool_use_id": variant.ID, "tool_name": variant.Name, "input": variant.Input})
			calls = append(calls, ToolCall{ToolUseID: variant.ID, ToolName: variant.Name, Arguments: variant.Input})
		case anthropic.TextBlock:
			if strings.TrimSpace(variant.Text) != "" {
				textParts = append(textParts, variant.Text)
			}
		}
	}
	if len(calls) > 0 {
		return Decision{Type: DecisionToolCall, ToolCalls: calls, Answer: strings.TrimSpace(strings.Join(textParts, "\n"))}, nil
	}

	answer := strings.TrimSpace(strings.Join(textParts, "\n"))
	if answer == "" {
		answer = fmt.Sprintf("Claude stopped with reason %q but did not return text or a tool call", resp.StopReason)
	}
	trace.Log(ctx, "model.AnthropicProvider.Next.final", map[string]any{"answer_len": len(answer)})
	return Decision{Type: DecisionFinal, Answer: answer}, nil
}

func toAnthropicMessages(ctx context.Context, messages []Message) ([]anthropic.MessageParam, error) {
	trace.Log(ctx, "model.toAnthropicMessages", map[string]any{"messages": len(messages)})
	result := make([]anthropic.MessageParam, 0, len(messages))
	for i := 0; i < len(messages); i++ {
		message := messages[i]
		switch message.Role {
		case RoleUser:
			result = append(result, anthropic.NewUserMessage(anthropic.NewTextBlock(message.Content)))
		case RoleAssistant:
			if message.ToolName != "" {
				blocks := []anthropic.ContentBlockParamUnion{}
				if strings.TrimSpace(message.Content) != "" {
					blocks = append(blocks, anthropic.NewTextBlock(message.Content))
				}
				for i < len(messages) && messages[i].Role == RoleAssistant && messages[i].ToolName != "" {
					block, err := anthropicToolUseBlock(messages[i])
					if err != nil {
						return nil, err
					}
					blocks = append(blocks, block)
					i++
				}
				i--
				result = append(result, anthropic.NewAssistantMessage(blocks...))
				continue
			}
			result = append(result, anthropic.NewAssistantMessage(anthropic.NewTextBlock(message.Content)))
		case RoleTool:
			blocks := []anthropic.ContentBlockParamUnion{}
			for i < len(messages) && messages[i].Role == RoleTool {
				if messages[i].ToolUseID == "" {
					return nil, fmt.Errorf("tool result for %q is missing tool_use_id", messages[i].ToolName)
				}
				blocks = append(blocks, anthropic.NewToolResultBlock(messages[i].ToolUseID, messages[i].Content, messages[i].ToolError))
				i++
			}
			i--
			result = append(result, anthropic.NewUserMessage(blocks...))
		default:
			return nil, fmt.Errorf("unsupported message role %q", message.Role)
		}
	}
	return result, nil
}

func anthropicToolUseBlock(message Message) (anthropic.ContentBlockParamUnion, error) {
	var input any = map[string]any{}
	if strings.TrimSpace(message.ToolInput) != "" {
		if err := json.Unmarshal([]byte(message.ToolInput), &input); err != nil {
			return anthropic.ContentBlockParamUnion{}, fmt.Errorf("invalid assistant tool input: %w", err)
		}
	}
	return anthropic.NewToolUseBlock(message.ToolUseID, input, message.ToolName), nil
}

func toAnthropicTools(ctx context.Context, toolSpecs []ToolSpec) ([]anthropic.ToolUnionParam, error) {
	trace.Log(ctx, "model.toAnthropicTools", map[string]any{"tools": len(toolSpecs)})
	tools := make([]anthropic.ToolUnionParam, 0, len(toolSpecs))
	for _, spec := range toolSpecs {
		var schema struct {
			Properties map[string]any `json:"properties"`
			Required   []string       `json:"required"`
		}
		if err := json.Unmarshal([]byte(spec.InputSchema), &schema); err != nil {
			return nil, fmt.Errorf("invalid input schema for tool %q: %w", spec.Name, err)
		}

		tool := anthropic.ToolParam{
			Name:        spec.Name,
			Description: anthropic.String(spec.Description),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: schema.Properties,
				Required:   schema.Required,
			},
		}
		tools = append(tools, anthropic.ToolUnionParam{OfTool: &tool})
	}
	return tools, nil
}
