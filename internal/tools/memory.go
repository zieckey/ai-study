package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zieckey/ai-study/internal/memory"
	"github.com/zieckey/ai-study/internal/trace"
)

type Memory struct {
	Store *memory.Store
}

func (Memory) Name() string {
	return "memory"
}

func (Memory) Description() string {
	return "读写本地持久化记忆。支持 set/get/list/delete，用来跨运行保存用户偏好或项目事实。不要保存密码、API Key、token 等敏感信息"
}

func (Memory) InputSchema() string {
	return `{"type":"object","properties":{"action":{"type":"string","enum":["set","get","list","delete"],"description":"记忆操作类型"},"key":{"type":"string","description":"记忆键，set/get/delete 时需要"},"value":{"type":"string","description":"记忆值，set 时需要"}},"required":["action"],"additionalProperties":false}`
}

func (m Memory) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	trace.Log(ctx, "tools.Memory.Execute.start", map[string]any{"input": input})
	if m.Store == nil {
		return "", fmt.Errorf("memory store is not configured")
	}

	var args struct {
		Action string `json:"action"`
		Key    string `json:"key"`
		Value  string `json:"value"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("invalid memory input: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(args.Action)) {
	case "set":
		if err := m.Store.Set(ctx, args.Key, args.Value); err != nil {
			return "", err
		}
		result := fmt.Sprintf("已保存记忆：%s", strings.TrimSpace(args.Key))
		trace.Log(ctx, "tools.Memory.Execute.done", map[string]any{"action": "set", "key": args.Key})
		return result, nil
	case "get":
		entry, ok, err := m.Store.Get(ctx, args.Key)
		if err != nil {
			return "", err
		}
		if !ok {
			return fmt.Sprintf("未找到记忆：%s", strings.TrimSpace(args.Key)), nil
		}
		result := fmt.Sprintf("%s = %s", entry.Key, entry.Value)
		trace.Log(ctx, "tools.Memory.Execute.done", map[string]any{"action": "get", "key": args.Key, "found": true})
		return result, nil
	case "list":
		entries, err := m.Store.List(ctx)
		if err != nil {
			return "", err
		}
		if len(entries) == 0 {
			return "暂无记忆", nil
		}
		lines := make([]string, 0, len(entries))
		for _, entry := range entries {
			lines = append(lines, fmt.Sprintf("- %s = %s", entry.Key, entry.Value))
		}
		trace.Log(ctx, "tools.Memory.Execute.done", map[string]any{"action": "list", "entries": len(entries)})
		return strings.Join(lines, "\n"), nil
	case "delete":
		deleted, err := m.Store.Delete(ctx, args.Key)
		if err != nil {
			return "", err
		}
		if !deleted {
			return fmt.Sprintf("未找到记忆：%s", strings.TrimSpace(args.Key)), nil
		}
		trace.Log(ctx, "tools.Memory.Execute.done", map[string]any{"action": "delete", "key": args.Key})
		return fmt.Sprintf("已删除记忆：%s", strings.TrimSpace(args.Key)), nil
	default:
		return "", fmt.Errorf("unsupported memory action %q", args.Action)
	}
}
