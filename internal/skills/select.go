package skills

import (
	"context"
	"strings"

	"github.com/zieckey/ai-study/internal/trace"
)

type Rule struct {
	SkillName string
	Keywords  []string
}

var DefaultRules = []Rule{
	{SkillName: "agent-teacher", Keywords: []string{"agent", "智能体", "tool_calls", "tool call", "工具调用", "provider", "trace", "observation", "deepseek", "claude", "anthropic"}},
	{SkillName: "math-tutor", Keywords: []string{"计算", "数学", "讲解", "解释", "推导", "算", "+", "-", "*", "/"}},
	{SkillName: "travel-planner", Keywords: []string{"旅行", "旅游", "行程", "酒店", "景点", "预算", "路线"}},
}

func Select(ctx context.Context, goal string, allSkills []Skill) []Skill {
	trace.Log(ctx, "skills.Select", map[string]any{"goal": goal, "skills": len(allSkills)})
	if len(allSkills) == 0 {
		return nil
	}

	byName := make(map[string]Skill, len(allSkills))
	for _, skill := range allSkills {
		byName[skill.Name] = skill
	}

	lowerGoal := strings.ToLower(goal)
	seen := map[string]bool{}
	selected := make([]Skill, 0, len(allSkills))
	for _, rule := range DefaultRules {
		if seen[rule.SkillName] {
			continue
		}
		if matchesRule(lowerGoal, rule) {
			if skill, ok := byName[rule.SkillName]; ok {
				selected = append(selected, skill)
				seen[rule.SkillName] = true
			}
		}
	}

	trace.Log(ctx, "skills.Select.done", map[string]any{"goal": goal, "selected": skillNames(selected)})
	return selected
}

func matchesRule(lowerGoal string, rule Rule) bool {
	for _, keyword := range rule.Keywords {
		if strings.Contains(lowerGoal, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func skillNames(skills []Skill) []string {
	names := make([]string, 0, len(skills))
	for _, skill := range skills {
		names = append(names, skill.Name)
	}
	return names
}
