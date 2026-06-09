package model

import "strings"

func withMemory(base string, memoryContext string) string {
	memoryContext = strings.TrimSpace(memoryContext)
	if memoryContext == "" {
		return base
	}

	var builder strings.Builder
	builder.WriteString(strings.TrimSpace(base))
	builder.WriteString("\n\n下面是本次请求自动注入的持久化记忆。请把它当作可能相关的历史上下文；如果用户问题与这些记忆无关，不要强行提及。\n")
	builder.WriteString("\n<memory>\n")
	builder.WriteString(memoryContext)
	builder.WriteString("\n</memory>\n")
	return builder.String()
}
