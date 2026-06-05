package skills

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/zieckey/ai-study/internal/trace"
)

func LoadDir(ctx context.Context, dir string) ([]Skill, error) {
	trace.Log(ctx, "skills.LoadDir", map[string]any{"dir": dir})
	if dir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			trace.Log(ctx, "skills.LoadDir.missing", map[string]any{"dir": dir})
			return nil, nil
		}
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	loaded := make([]Skill, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		skill, err := LoadFile(ctx, filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		loaded = append(loaded, skill)
	}

	trace.Log(ctx, "skills.LoadDir.done", map[string]any{"dir": dir, "skills": len(loaded)})
	return loaded, nil
}

func LoadFile(ctx context.Context, path string) (Skill, error) {
	trace.Log(ctx, "skills.LoadFile", map[string]any{"path": path})
	data, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, err
	}

	skill, err := ParseMarkdown(string(data))
	if err != nil {
		return Skill{}, fmt.Errorf("parse skill %s: %w", path, err)
	}
	if skill.Name == "" {
		skill.Name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	trace.Log(ctx, "skills.LoadFile.done", map[string]any{"path": path, "name": skill.Name, "content_len": len(skill.Content)})
	return skill, nil
}

func ParseMarkdown(markdown string) (Skill, error) {
	markdown = strings.ReplaceAll(markdown, "\r\n", "\n")
	if !strings.HasPrefix(markdown, "---\n") {
		return Skill{Content: strings.TrimSpace(markdown)}, nil
	}

	rest := strings.TrimPrefix(markdown, "---\n")
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		return Skill{}, fmt.Errorf("frontmatter is not closed")
	}

	frontmatter := rest[:end]
	content := strings.TrimSpace(rest[end+len("\n---\n"):])
	skill := Skill{Content: content}

	for _, line := range strings.Split(frontmatter, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return Skill{}, fmt.Errorf("invalid frontmatter line %q", line)
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch strings.TrimSpace(key) {
		case "name":
			skill.Name = value
		case "description":
			skill.Description = value
		}
	}

	return skill, nil
}
