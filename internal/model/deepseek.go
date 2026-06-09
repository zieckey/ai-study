package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zieckey/ai-study/internal/trace"
)

const (
	defaultDeepSeekBaseURL      = "https://api.deepseek.com"
	defaultDeepSeekModel        = "deepseek-chat"
	defaultDeepSeekSystemPrompt = `你是这个 Go 学习项目里的 DeepSeek provider。
你可以使用工具完成任务：calculator 负责精确计算，clock 负责获取当前时间，weather 返回 mock 天气，echo 原样返回文本，memory 负责读写本地持久化记忆。
当用户要求记住、保存、回忆、列出或删除偏好/事实时，请调用 memory 工具。不要把密码、API Key、token 等敏感信息写入记忆。
当用户的问题需要外部实时信息、确定性计算或项目提供的工具能力时，请优先调用合适的工具；工具返回 observation 后，再给出简洁的中文最终答案。`
)

type DeepSeekConfig struct {
	APIKey       string
	BaseURL      string
	Model        string
	MaxTokens    int64
	SystemPrompt string
	HTTPClient   *http.Client
}

type DeepSeekProvider struct {
	apiKey       string
	baseURL      string
	model        string
	maxTokens    int64
	systemPrompt string
	httpClient   *http.Client
}

func NewDeepSeekProvider(config DeepSeekConfig) *DeepSeekProvider {
	if config.APIKey == "" {
		config.APIKey = os.Getenv("DEEPSEEK_API_KEY")
	}
	if config.BaseURL == "" {
		config.BaseURL = os.Getenv("DEEPSEEK_BASE_URL")
	}
	if config.BaseURL == "" {
		config.BaseURL = defaultDeepSeekBaseURL
	}
	if config.Model == "" {
		config.Model = defaultDeepSeekModel
	}
	if config.MaxTokens <= 0 {
		config.MaxTokens = 4096
	}
	if config.SystemPrompt == "" {
		config.SystemPrompt = defaultDeepSeekSystemPrompt
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 60 * time.Second}
	}

	return &DeepSeekProvider{
		apiKey:       config.APIKey,
		baseURL:      strings.TrimRight(config.BaseURL, "/"),
		model:        config.Model,
		maxTokens:    config.MaxTokens,
		systemPrompt: config.SystemPrompt,
		httpClient:   config.HTTPClient,
	}
}

func (p *DeepSeekProvider) Next(ctx context.Context, req Request) (Decision, error) {
	trace.Log(ctx, "model.DeepSeekProvider.Next.start", map[string]any{"model": p.model, "base_url": p.baseURL, "max_tokens": p.maxTokens, "messages": req.Messages, "tools": len(req.Tools), "skills": len(req.Skills), "api_key_set": p.apiKey != ""})
	if p.apiKey == "" {
		return Decision{}, fmt.Errorf("DEEPSEEK_API_KEY is required for deepseek provider")
	}

	systemPrompt := withSkills(p.systemPrompt, req.Skills)
	trace.Log(ctx, "model.DeepSeekProvider.Next.prepareSkill", map[string]any{"systemPrompt": systemPrompt})
	messages, err := toDeepSeekMessages(ctx, req.Messages, systemPrompt)
	if err != nil {
		return Decision{}, err
	}
	deepSeekTools, err := toDeepSeekTools(ctx, req.Tools)
	if err != nil {
		return Decision{}, err
	}

	body := deepSeekChatRequest{
		Model:      p.model,
		Messages:   messages,
		Tools:      deepSeekTools,
		ToolChoice: "auto",
		MaxTokens:  p.maxTokens,
	}

	trace.Log(ctx, "model.DeepSeekProvider.Next.request", map[string]any{"model": body.Model, "messages": body.Messages, "tools": body.Tools, "skills": len(req.Skills), "tool_choice": body.ToolChoice, "max_tokens": body.MaxTokens})
	payload, err := json.Marshal(body)
	if err != nil {
		return Decision{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return Decision{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return Decision{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Decision{}, err
	}
	trace.Log(ctx, "model.DeepSeekProvider.Next.http_response", map[string]any{"status": resp.Status, "body_bytes": respBody})
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Decision{}, fmt.Errorf("deepseek API returned %s: %s", resp.Status, strings.TrimSpace(string(respBody)))
	}

	var chatResp deepSeekChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return Decision{}, err
	}
	if len(chatResp.Choices) == 0 {
		return Decision{}, fmt.Errorf("deepseek API returned no choices")
	}
	trace.Log(ctx, "model.DeepSeekProvider.Next.response", map[string]any{"choices": len(chatResp.Choices), "tool_calls": len(chatResp.Choices[0].Message.ToolCalls), "chatResp": chatResp})

	message := chatResp.Choices[0].Message
	if len(message.ToolCalls) > 0 {
		toolCall := message.ToolCalls[0]
		trace.Log(ctx, "model.DeepSeekProvider.Next.tool_call", map[string]any{"tool_call_id": toolCall.ID, "tool_name": toolCall.Function.Name, "arguments": toolCall.Function.Arguments})
		return Decision{
			Type:      DecisionToolCall,
			ToolUseID: toolCall.ID,
			ToolName:  toolCall.Function.Name,
			Arguments: json.RawMessage(toolCall.Function.Arguments),
			Answer:    message.Content,
		}, nil
	}

	answer := strings.TrimSpace(message.Content)
	if answer == "" {
		answer = "DeepSeek returned an empty answer."
	}
	trace.Log(ctx, "model.DeepSeekProvider.Next.final", map[string]any{"answer_len": len(answer)})
	return Decision{Type: DecisionFinal, Answer: answer}, nil
}

func toDeepSeekMessages(ctx context.Context, messages []Message, systemPrompt string) ([]deepSeekMessage, error) {
	trace.Log(ctx, "model.toDeepSeekMessages", map[string]any{"messages": messages, "system_prompt_len": len(systemPrompt)})
	result := []deepSeekMessage{{Role: "system", Content: systemPrompt}}
	for _, message := range messages {
		switch message.Role {
		case RoleUser:
			result = append(result, deepSeekMessage{Role: "user", Content: message.Content})
		case RoleAssistant:
			if message.ToolName != "" {
				if message.ToolUseID == "" {
					return nil, fmt.Errorf("assistant tool call for %q is missing tool_use_id", message.ToolName)
				}
				result = append(result, deepSeekMessage{
					Role:    "assistant",
					Content: message.Content,
					ToolCalls: []deepSeekToolCall{{
						ID:   message.ToolUseID,
						Type: "function",
						Function: deepSeekToolCallFunction{
							Name:      message.ToolName,
							Arguments: message.ToolInput,
						},
					}},
				})
				continue
			}
			result = append(result, deepSeekMessage{Role: "assistant", Content: message.Content})
		case RoleTool:
			if message.ToolUseID == "" {
				return nil, fmt.Errorf("tool result for %q is missing tool_use_id", message.ToolName)
			}
			result = append(result, deepSeekMessage{Role: "tool", Content: message.Content, ToolCallID: message.ToolUseID})
		default:
			return nil, fmt.Errorf("unsupported message role %q", message.Role)
		}
	}
	return result, nil
}

func toDeepSeekTools(ctx context.Context, toolSpecs []ToolSpec) ([]deepSeekTool, error) {
	trace.Log(ctx, "model.toDeepSeekTools", map[string]any{"tools": toolSpecs})
	tools := make([]deepSeekTool, 0, len(toolSpecs))
	for _, spec := range toolSpecs {
		var parameters map[string]any
		if err := json.Unmarshal([]byte(spec.InputSchema), &parameters); err != nil {
			return nil, fmt.Errorf("invalid input schema for tool %q: %w", spec.Name, err)
		}
		tools = append(tools, deepSeekTool{
			Type: "function",
			Function: deepSeekFunction{
				Name:        spec.Name,
				Description: spec.Description,
				Parameters:  parameters,
			},
		})
	}
	return tools, nil
}

type deepSeekChatRequest struct {
	Model      string            `json:"model"`
	Messages   []deepSeekMessage `json:"messages"`
	Tools      []deepSeekTool    `json:"tools,omitempty"`
	ToolChoice string            `json:"tool_choice,omitempty"`
	MaxTokens  int64             `json:"max_tokens,omitempty"`
}

type deepSeekMessage struct {
	Role       string             `json:"role"`
	Content    string             `json:"content,omitempty"`
	ToolCalls  []deepSeekToolCall `json:"tool_calls,omitempty"`
	ToolCallID string             `json:"tool_call_id,omitempty"`
}

type deepSeekTool struct {
	Type     string           `json:"type"`
	Function deepSeekFunction `json:"function"`
}

type deepSeekFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type deepSeekChatResponse struct {
	Choices []deepSeekChoice `json:"choices"`
}

type deepSeekChoice struct {
	Message deepSeekMessage `json:"message"`
}

type deepSeekToolCall struct {
	ID       string                   `json:"id"`
	Type     string                   `json:"type"`
	Function deepSeekToolCallFunction `json:"function"`
}

type deepSeekToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
