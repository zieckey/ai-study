# Golang AI Agent 入门项目

这个项目用 Go 从零实现一个最小但完整的 AI Agent。它不依赖真实大模型 API，默认使用一个规则驱动的 `mock` 模型来模拟 LLM 的工具调用能力，所以你可以直接运行、调试和修改代码，专注学习 Agent 的核心原理。

## 你会学到什么

- AI Agent 和普通聊天机器人的区别
- Agent Loop 是什么
- 模型、工具、观察结果、终止条件分别负责什么
- 如何用 Go 组织一个可扩展的 Agent 项目
- 如何添加一个新工具
- 后续如何接入 Claude、OpenAI 或本地模型

## 什么是 AI Agent

一个最小 AI Agent 可以理解为：

```text
LLM + Tools + Memory/State + Loop + Stop Condition
```

也就是：

1. 用户给出目标。
2. 模型判断下一步应该直接回答，还是调用工具。
3. 如果需要工具，Agent 执行工具。
4. 工具返回 observation，也就是外部世界的事实结果。
5. Agent 把 observation 放回上下文，再问模型下一步。
6. 模型最终给出答案，循环结束。

这个过程就是 Agent Loop：

```text
用户目标 -> 模型决策 -> 工具调用 -> 工具执行 -> observation -> 模型决策 -> 最终答案
```

普通聊天机器人通常是“一问一答”；Agent 则可以多走几步，通过工具获取信息或执行动作。

## 为什么先不用真实大模型

真实大模型的 tool calling 会涉及 API Key、网络、额度、SDK、模型输出格式和错误重试。对于入门学习来说，这些细节会掩盖 Agent 的核心机制。

所以本项目第一版使用 `mock` 模型：

- 输入里有 `12 * 23` 这种表达式时，模型决定调用 `calculator`。
- 输入里有 `时间` 或 `几点` 时，模型决定调用 `clock`。
- 输入里有 `天气` 时，模型决定调用 `weather`。
- 输入里有 `重复` 时，模型决定调用 `echo`。
- 工具返回 observation 后，模型输出 final answer。

这样你可以稳定看到完整流程。

## 运行项目

```bash
go run ./cmd/agent "帮我计算 12 * 23"
```

示例输出：

```text
Goal: 帮我计算 12 * 23

Step 1
Model decision: call tool calculator
Tool input: {"expression":"12 * 23"}
Observation: 276

Step 2
Model decision: final answer

Final Answer:
12 * 23 = 276
```

更多例子：

```bash
go run ./cmd/agent "现在几点？"
go run ./cmd/agent "查询北京天气"
go run ./cmd/agent "请重复 hello agent"
go run ./cmd/agent -max-steps 3 "帮我计算 1 + 2"
go run ./cmd/agent -json "查询北京天气"
```

`-json` 会输出结构化 trace，适合接日志、Web UI 或其他程序：

```json
{
  "goal": "查询北京天气",
  "answer": "天气查询结果：北京：晴，25°C",
  "trace": [
    {
      "step": 1,
      "decision": "tool_call",
      "tool_name": "weather",
      "tool_input": {"city":"北京"},
      "observation": "北京：晴，25°C"
    },
    {
      "step": 2,
      "decision": "final"
    }
  ]
}
```

运行测试：

```bash
go test ./...
```

## 项目结构

```text
.
├── cmd/agent/main.go              # CLI 入口
├── internal/agent/agent.go        # Agent Loop
├── internal/agent/types.go        # Agent 配置、结果、trace 类型
├── internal/model/provider.go     # 模型 Provider 抽象
├── internal/model/mock.go         # 规则驱动的 mock 模型
├── internal/tools/tool.go         # 工具接口和注册表
├── internal/tools/calculator.go   # 计算器工具
├── internal/tools/clock.go        # 当前时间工具
├── internal/tools/echo.go         # 回显工具
└── internal/tools/weather.go      # mock 天气工具
```

## 核心实现讲解

### 1. 模型 Provider

`internal/model/provider.go` 定义了模型接口：

```go
type Provider interface {
    Next(ctx context.Context, req Request) (Decision, error)
}
```

Agent 不关心底层是真实 LLM 还是 mock，只关心它能根据当前上下文返回一个 `Decision`。

`Decision` 有两种：

```text
tool_call: 模型要求调用某个工具
final:     模型给出最终答案
```

这和真实大模型的 tool use/function calling 思路是一致的，只是这里用 Go struct 表达，便于理解和测试。

### 2. 工具 Tool

`internal/tools/tool.go` 定义了工具接口：

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() string
    Execute(ctx context.Context, input json.RawMessage) (string, error)
}
```

工具是 Agent 影响外部世界或获取外部信息的方式。比如：

- `calculator` 负责计算。
- `clock` 负责读取当前时间。
- `weather` 负责返回 mock 天气。
- `echo` 负责返回文本。

模型只决定“要调用什么工具、传什么参数”；真正执行动作的是工具。

### 3. Agent Loop

`internal/agent/agent.go` 是核心。

简化后的逻辑是：

```text
messages = [user goal]

for step in 1..MaxSteps:
    decision = provider.Next(messages, tools)

    if decision is final:
        return answer

    if decision is tool_call:
        observation = tool.Execute(arguments)
        messages append tool_call
        messages append observation

return max steps error
```

这个循环体现了 Agent 的本质：Agent 不是某个单独的模型，而是一个控制器。它负责维护状态、询问模型、执行工具、处理结果、判断是否结束。

### 4. Trace 为什么重要

CLI 默认打印 trace：

```text
Step 1
Model decision: call tool calculator
Tool input: {"expression":"12 * 23"}
Observation: 276
```

Trace 可以帮助你看到模型“想做什么”、工具“实际做了什么”、Agent “如何把结果继续喂回模型”。学习 Agent 时，trace 比最终答案更重要。

## 如何添加一个新工具

本项目已经实现了 `weather` 工具，可以用它学习添加工具的完整路径。

第一步，实现 `tools.Tool` 接口：

```go
type Weather struct{}

func (Weather) Name() string { return "weather" }
func (Weather) Description() string { return "查询城市天气，当前返回 mock 数据" }
func (Weather) InputSchema() string { return `{"city":"string"}` }
func (Weather) Execute(ctx context.Context, input json.RawMessage) (string, error) {
    return "北京：晴，25°C", nil
}
```

第二步，在 `cmd/agent/main.go` 注册工具：

```go
[]tools.Tool{
    tools.Calculator{},
    tools.Clock{},
    tools.Echo{},
    tools.Weather{},
}
```

第三步，让 mock 模型在输入包含“天气”时返回：

```text
tool_call weather {"city":"北京"}
```

如果接入真实大模型，第三步通常由模型根据工具描述自动决定。你可以照这个模式继续添加 `currency`、`todo` 或 `file_search` 工具。

## 如何扩展到真实模型

当前项目只有 `mock` provider。后续可以新增：

```text
internal/model/anthropic.go
internal/model/openai.go
internal/model/ollama.go
```

它们都实现同一个接口：

```go
type Provider interface {
    Next(ctx context.Context, req Request) (Decision, error)
}
```

这样 Agent Loop 不需要改，只替换 provider 即可。

真实模型接入时，一般要做这些事：

1. 把 `messages` 转成模型 API 需要的格式。
2. 把 `tools.Tool` 的描述和参数 schema 转成模型的 tool schema。
3. 调用模型 API。
4. 如果模型返回 tool call，转成项目里的 `DecisionToolCall`。
5. 如果模型返回文本答案，转成 `DecisionFinal`。

这就是为什么项目先抽象出 `Provider`：Agent Loop 和具体模型 API 解耦。

## 常见问题

### Agent 一定需要 LLM 吗？

真实场景通常需要，但学习 Agent Loop 时不一定。这个项目用 mock 模型是为了把控制流程讲清楚。

### Agent 和工作流有什么区别？

工作流通常是固定步骤，比如 A -> B -> C。Agent 的下一步由模型根据上下文决定，因此路径可以变化。

### 为什么要限制 MaxSteps？

如果模型一直要求调用工具、不输出最终答案，Agent 可能陷入死循环。`MaxSteps` 是必要的安全边界。

### 为什么 calculator 不执行任意表达式？

工具是代码执行边界，必须谨慎。示例只支持简单二元四则运算，避免把用户输入当成代码执行。

## 练习

1. 添加一个 `weather` mock 工具。
2. 添加一个 `file_search` 工具，搜索当前目录里的文件名。
3. 给 CLI 增加 `-json` 参数，输出结构化 trace。
4. 给 Agent 增加多轮对话模式。
5. 实现 `anthropic` provider，接入 Claude 的 tool use。
6. 实现 `ollama` provider，接入本地模型。

## 当前版本的边界

- 只支持 mock provider。
- calculator 只支持 `number operator number` 格式。
- 没有长期记忆。
- 没有真实网络工具。
- 没有复杂 JSON Schema 校验。

这些边界是有意保留的：第一版的重点是理解 Agent 的最小闭环。