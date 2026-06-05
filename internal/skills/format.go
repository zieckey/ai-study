package skills

import "strings"

func FormatForPrompt(selected []Skill) string {
	if len(selected) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("\n\n下面是当前任务相关的 skills。请遵循这些专项说明，但不要向用户生硬提及 skill 名称，除非用户询问。\n")
	for _, skill := range selected {
		builder.WriteString("\n<skill name=\"")
		builder.WriteString(skill.Name)
		builder.WriteString("\">\n")
		if skill.Description != "" {
			builder.WriteString("Description: ")
			builder.WriteString(skill.Description)
			builder.WriteString("\n\n")
		}
		builder.WriteString(strings.TrimSpace(skill.Content))
		builder.WriteString("\n</skill>\n")
	}
	return builder.String()
}
