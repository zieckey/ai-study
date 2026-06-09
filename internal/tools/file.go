package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zieckey/ai-study/internal/trace"
)

type FileSearch struct {
	Root string
}

func (t FileSearch) Name() string {
	return "file_search"
}

func (t FileSearch) Description() string {
	return "在项目目录内按文件名或相对路径关键词搜索文件。只返回项目内相对路径，不读取文件内容"
}

func (t FileSearch) InputSchema() string {
	return `{"type":"object","properties":{"query":{"type":"string","description":"文件名或相对路径关键词，例如 README、deepseek、agent.go"},"limit":{"type":"integer","description":"最多返回多少条，默认 20"}},"required":["query"],"additionalProperties":false}`
}

func (t FileSearch) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	trace.Log(ctx, "tools.FileSearch.Execute.start", map[string]any{"input": input})
	var args struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("invalid file_search input: %w", err)
	}
	query := strings.ToLower(strings.TrimSpace(args.Query))
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	if args.Limit <= 0 {
		args.Limit = 20
	}

	root, err := projectRoot(t.Root)
	if err != nil {
		return "", err
	}

	matches := []string{}
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && shouldSkipDir(entry.Name()) && path != root {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.Contains(strings.ToLower(rel), query) {
			matches = append(matches, rel)
		}
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(matches)
	if len(matches) > args.Limit {
		matches = matches[:args.Limit]
	}
	if len(matches) == 0 {
		return "未找到匹配文件", nil
	}
	trace.Log(ctx, "tools.FileSearch.Execute.done", map[string]any{"query": args.Query, "matches": len(matches), "root": root})
	return strings.Join(matches, "\n"), nil
}

type ReadFile struct {
	Root string
}

func (t ReadFile) Name() string {
	return "read_file"
}

func (t ReadFile) Description() string {
	return "读取项目目录内指定文本文件内容。路径必须是项目内相对路径，不能读取项目外文件"
}

func (t ReadFile) InputSchema() string {
	return `{"type":"object","properties":{"path":{"type":"string","description":"项目内相对路径，例如 README.md 或 internal/agent/agent.go"},"max_bytes":{"type":"integer","description":"最多读取字节数，默认 20000"}},"required":["path"],"additionalProperties":false}`
}

func (t ReadFile) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	trace.Log(ctx, "tools.ReadFile.Execute.start", map[string]any{"input": input})
	var args struct {
		Path     string `json:"path"`
		MaxBytes int    `json:"max_bytes"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("invalid read_file input: %w", err)
	}
	if args.MaxBytes <= 0 {
		args.MaxBytes = 20000
	}
	if args.MaxBytes > 100000 {
		args.MaxBytes = 100000
	}

	root, err := projectRoot(t.Root)
	if err != nil {
		return "", err
	}
	path, err := safeProjectPath(root, args.Path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", args.Path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	truncated := false
	if len(data) > args.MaxBytes {
		data = data[:args.MaxBytes]
		truncated = true
	}
	rel, _ := filepath.Rel(root, path)
	rel = filepath.ToSlash(rel)
	trace.Log(ctx, "tools.ReadFile.Execute.done", map[string]any{"path": rel, "bytes": len(data), "truncated": truncated})
	if truncated {
		return fmt.Sprintf("%s\n\n[内容已截断到 %d bytes]", string(data), args.MaxBytes), nil
	}
	return string(data), nil
}

func projectRoot(root string) (string, error) {
	if root == "" {
		root = "."
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

func safeProjectPath(root string, relPath string) (string, error) {
	relPath = strings.TrimSpace(relPath)
	if relPath == "" {
		return "", fmt.Errorf("path is required")
	}
	if filepath.IsAbs(relPath) {
		return "", fmt.Errorf("path must be relative to project root")
	}
	clean := filepath.Clean(relPath)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("path must stay inside project root")
	}
	joined := filepath.Join(root, clean)
	resolved, err := filepath.EvalSymlinks(joined)
	if err != nil {
		return "", err
	}
	if resolved != root && !strings.HasPrefix(resolved, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes project root")
	}
	return resolved, nil
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", ".claude", ".claude-app", ".cjadk", "node_modules", "vendor":
		return true
	default:
		return false
	}
}
