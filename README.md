# Golang AI Agent 入门项目

最开始的prompt：“我想创建一个golang的简单项目，帮我学习如何编写AI Agent，把原理和实现都做详细讲解”。

这个项目用 Go 从零实现一个最小但完整的 AI Agent。默认使用规则驱动的 `mock` 模型来模拟 LLM 的工具调用能力，不需要 API Key 就能运行；同时也支持 `anthropic` 和 `deepseek` provider，用真实大模型演示 tool use。

## 你会学到什么

- AI Agent 和普通聊天机器人的区别
- Agent Loop 是什么
- 模型、工具、观察结果、终止条件分别负责什么
- 如何用 Go 组织一个可扩展的 Agent 项目
- 如何添加一个新工具
- 如何接入 Claude/Anthropic、DeepSeek、OpenAI 或本地模型

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

CLI 默认会把函数级 trace log 打印到 stderr，方便观察程序运行轨迹和关键参数。每条日志都是 JSON：

```text
[trace] {"function":"agent.Run.step","step":1,"messages":1,"ts":"..."}
[trace] {"function":"tools.Weather.Execute.done","city":"北京","result":"北京：晴，25°C","ts":"..."}
```

如果只想看最终输出，可以关闭函数级日志：

```bash
go run ./cmd/agent -trace-log=false "查询北京天气"
```

注意：trace log 会打印用户输入和工具参数；API Key、Authorization、secret、token 等敏感字段会被掩码。

使用真实 Claude provider：

```bash
export ANTHROPIC_API_KEY="你的 API Key"
go run ./cmd/agent -provider anthropic "帮我计算 12 * 23，然后查询北京天气"
```

可选参数：

```bash
go run ./cmd/agent \
  -provider anthropic \
  -anthropic-model claude-opus-4-7 \
  -anthropic-max-tokens 4096 \
  "查询北京天气"
```

`anthropic` provider 使用官方 Go SDK：

- 默认模型是 `claude-opus-4-7`。
- API Key 从 `ANTHROPIC_API_KEY` 环境变量读取，不要写进代码或提交到 Git。
- Agent Loop 仍然由本项目控制；Claude 只负责根据上下文返回 `tool_use` 或最终文本。
- 工具仍在本地执行，所以你可以在 `internal/tools` 中控制真实副作用和安全边界。

使用真实 DeepSeek provider：

```bash
export DEEPSEEK_API_KEY="你的 DeepSeek API Key"
go run ./cmd/agent -provider deepseek "帮我计算 12 * 23，然后查询北京天气"
```

可选参数：

```bash
go run ./cmd/agent \
  -provider deepseek \
  -deepseek-model deepseek-chat \
  -deepseek-max-tokens 4096 \
  "查询北京天气"
```

`deepseek` provider 使用 OpenAI-compatible Chat Completions 协议：

- 默认模型是 `deepseek-chat`。
- API Key 从 `DEEPSEEK_API_KEY` 环境变量读取。
- 如果你使用代理或兼容网关，可以用 `DEEPSEEK_BASE_URL` 覆盖默认地址。
- DeepSeek 返回 `tool_calls` 后，Agent 仍然在本地执行工具，并把 `tool_call_id` 对应的 observation 继续发回模型。

记忆默认只通过 `memory` 工具按需读取。如果希望每次请求都自动把本地记忆注入模型上下文，可以开启：

```bash
go run ./cmd/agent \
  -memory-path memory/demo.json \
  -memory-in-context \
  "请根据我的偏好回答"
```

开启后，程序会在请求前读取记忆文件，把内容追加到 system prompt 的 `<memory>` 块中。记忆越多，消耗的 token 越多；不要把密码、API Key、token 等敏感信息写入记忆。

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
├── internal/model/anthropic.go    # 真实 Claude/Anthropic tool use provider
├── internal/model/deepseek.go     # 真实 DeepSeek tool calling provider
├── internal/memory/               # 本地持久化记忆存储
├── internal/skills/               # 本地 Markdown skill 加载与选择
├── internal/tools/tool.go         # 工具接口和注册表
├── internal/tools/calculator.go   # 计算器工具
├── internal/tools/clock.go        # 当前时间工具
├── internal/tools/echo.go         # 回显工具
├── internal/tools/weather.go      # mock 天气工具
├── internal/tools/memory.go       # 本地记忆工具
├── memory/memory.json             # 默认记忆文件，运行时自动创建
└── skills/                        # 示例 skill 文件
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
- `memory` 负责读写本地持久化记忆。

模型只决定“要调用什么工具、传什么参数”；真正执行动作的是工具。

### 3. 记忆 Memory 如何持久化上下文

记忆功能解决的问题是：普通 Agent 的 `messages` 只在一次运行中存在，程序结束后就丢失了。如果希望 Agent 跨运行记住用户偏好或项目事实，就需要把信息保存到外部存储。

本项目实现的是一个最小本地记忆系统：

```text
用户说“记住 language=Go”
  ↓
模型或 mock provider 决定调用 memory 工具
  ↓
memory 工具执行 set 操作
  ↓
internal/memory.Store 写入 memory/memory.json
  ↓
下次运行时，memory 工具可以 get/list/delete 这些记录
```

它由两部分组成：

```text
internal/memory/Store  = 负责 JSON 文件读写
internal/tools/Memory  = 暴露给 Agent/模型调用的工具
```

默认记忆文件是：

```text
memory/memory.json
```

可以用 `-memory-path` 修改：

```bash
go run ./cmd/agent -memory-path memory/demo.json "记住 language=Go"
go run ./cmd/agent -memory-path memory/demo.json "列出记忆"
```

`memory` 工具支持四个动作：

```json
{"action":"set","key":"language","value":"Go"}
{"action":"get","key":"language"}
{"action":"list"}
{"action":"delete","key":"language"}
```

保存后的文件结构类似：

```json
{
  "language": {
    "key": "language",
    "value": "Go",
    "updated_at": "2026-06-08T10:00:00+08:00"
  }
}
```

默认情况下，记忆不会自动进入模型上下文，模型只知道可以调用 `memory` 工具读取它。开启 `-memory-in-context` 后，程序会在每次请求前读取记忆文件，并把内容注入 system prompt：

```text
<memory>
- language = Go
</memory>
```

这种方式的优点是模型一开始就能看到已保存记忆；缺点是记忆越多，token 消耗越多，也可能暴露不相关上下文。

注意：不要把密码、API Key、token 等敏感信息写入记忆。记忆文件是普通本地 JSON 文件，如果项目公开提交到 GitHub，应把 `memory/` 加入 `.gitignore`。

### 4. Agent Loop

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

### 5. Skill 如何按需增强模型

Skill 是给模型看的专项说明书，不直接执行代码。它和工具的区别是：

```text
Tool  = Go 程序真正执行的能力，例如 calculator/weather
Skill = 模型回答时参考的任务策略、领域知识、输出格式说明
```

本项目支持本地 Markdown skill。默认从 `skills/` 目录加载：

```text
skills/
├── agent-teacher.md
├── math-tutor.md
└── travel-planner.md
```

每个 skill 文件包含 frontmatter 和正文：

```md
---
name: agent-teacher
description: 当用户想学习 AI Agent 原理、tool_calls、provider、trace log、模型交互流程时使用
---

你是一个 AI Agent 教学助手。
回答时遵循：
1. 先解释原理
2. 再对应到本项目代码
3. 最后给一个运行例子
```

运行时流程是：

```text
用户 goal
  ↓
internal/skills.LoadDir() 加载 skills/*.md
  ↓
internal/skills.Select() 根据关键词选择相关 skill
  ↓
Agent 把 selected skills 放进 model.Request
  ↓
DeepSeek/Anthropic provider 把 skill 内容拼到 system prompt
  ↓
模型按专项说明回答
```

默认 selector 是简单关键词匹配，例如：

- `agent`、`tool_calls`、`provider`、`trace`、`deepseek` -> `agent-teacher`
- `计算`、`数学`、`讲解`、`*`、`/` -> `math-tutor`
- `旅行`、`行程`、`预算`、`景点` -> `travel-planner`

可以用 `-skill-dir` 指定其他 skill 目录：

```bash
go run ./cmd/agent -skill-dir skills "请讲解 DeepSeek 的 tool_calls"
```

如果你不想加载本地 skills，可以指向一个不存在或空目录：

```bash
go run ./cmd/agent -skill-dir /tmp/empty-skills "查询北京天气"
```

### 6. Trace 为什么重要

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
func (Weather) InputSchema() string {
    return `{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`
}
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

当前项目已经实现了 `anthropic` 和 `deepseek` provider。后续如果要增加 OpenAI 或 Ollama，可以继续新增：

```text
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
4. 如果模型返回 tool call，转成项目里的 `DecisionToolCall`，并保留 `tool_use_id` 或 `tool_call_id`。
5. Agent 执行本地工具，把 observation 和对应 ID 一起放回上下文。
6. 如果模型返回文本答案，转成 `DecisionFinal`。

这就是为什么项目先抽象出 `Provider`：Agent Loop 和具体模型 API 解耦。

## DeepSeek 第一次请求与 tool_calls 格式

以 DeepSeek provider 为例，第一次请求发生在 `internal/model/deepseek.go` 的 `DeepSeekProvider.Next()` 中。它的目的不是直接拿最终答案，而是把当前 Agent 状态发给 DeepSeek，让模型判断下一步应该直接回答，还是调用某个工具。

当你运行：

```bash
go run ./cmd/agent -provider deepseek "查询北京天气"
```

Agent Loop 会先构造内部消息：

```go
[]Message{
    {Role: RoleUser, Content: "查询北京天气"},
}
```

然后把两类信息传给 DeepSeek provider：

1. `Messages`：当前对话历史。第一次请求时通常只有用户输入。
2. `Tools`：当前 Agent 支持的工具，例如 `calculator`、`clock`、`weather`、`echo`。

DeepSeek provider 会把这些内部结构转换成 OpenAI-compatible Chat Completions 请求。

### 第一次请求体

概念上，请求 JSON 类似这样：

```json
{
  "model": "deepseek-chat",
  "messages": [
    {
      "role": "system",
      "content": "你是这个 Go 学习项目里的 DeepSeek provider..."
    },
    {
      "role": "user",
      "content": "查询北京天气"
    }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "weather",
        "description": "查询城市天气，当前返回 mock 数据",
        "parameters": {
          "type": "object",
          "properties": {
            "city": {
              "type": "string",
              "description": "城市名称，例如 北京、上海、深圳、杭州"
            }
          },
          "required": ["city"],
          "additionalProperties": false
        }
      }
    }
  ],
  "tool_choice": "auto",
  "max_tokens": 4096
}
```

`messages` 给模型提供上下文；`tools` 告诉模型有哪些函数可以调用，以及每个函数需要什么参数；`tool_choice: "auto"` 表示让模型自己决定是否调用工具。

### DeepSeek 直接回答的返回格式

如果用户问的是普通问题，例如：

```text
你好
```

DeepSeek 可能直接返回：

```json
{
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "你好！有什么我可以帮你的吗？"
      }
    }
  ]
}
```

这种情况下，provider 会把它转成：

```go
Decision{
    Type:   DecisionFinal,
    Answer: answer,
}
```

### DeepSeek 调用工具的返回格式

如果用户问：

```text
查询北京天气
```

DeepSeek 应该返回 `tool_calls`，格式类似：

```json
{
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "",
        "tool_calls": [
          {
            "id": "call_1",
            "type": "function",
            "function": {
              "name": "weather",
              "arguments": "{\"city\":\"北京\"}"
            }
          }
        ]
      }
    }
  ]
}
```

`tool_calls` 是一个数组，每个元素表示一次工具调用。

字段含义：

- `id`：本次工具调用的唯一 ID。后续把工具结果发回 DeepSeek 时，必须带上这个 ID。
- `type`：当前是 `function`，表示函数工具调用。
- `function.name`：模型想调用的工具名，例如 `weather`。
- `function.arguments`：工具参数。注意它是一个 JSON 字符串，不是直接的 JSON object。例如 `"{\"city\":\"北京\"}"` 的实际含义是 `{"city":"北京"}`。

代码会把这个结果转成内部决策：

```go
Decision{
    Type:      DecisionToolCall,
    ToolUseID: "call_1",
    ToolName:  "weather",
    Arguments: json.RawMessage(`{"city":"北京"}`),
}
```

然后 Agent 根据 `ToolName` 找到本地工具，执行：

```go
observation, err := tool.Execute(ctx, decision.Arguments)
```

对于 `weather` 工具，输入是：

```json
{"city":"北京"}
```

输出是：

```text
北京：晴，25°C
```

### 第二次请求如何把工具结果发回 DeepSeek

工具执行后，内部 messages 会变成类似：

```go
[]Message{
    {
        Role:    RoleUser,
        Content: "查询北京天气",
    },
    {
        Role:      RoleAssistant,
        ToolUseID: "call_1",
        ToolName:  "weather",
        ToolInput: `{"city":"北京"}`,
    },
    {
        Role:      RoleTool,
        ToolUseID: "call_1",
        ToolName:  "weather",
        Content:   "北京：晴，25°C",
    },
}
```

再转换成 DeepSeek API 格式时，会变成：

```json
[
  {
    "role": "system",
    "content": "你是这个 Go 学习项目里的 DeepSeek provider..."
  },
  {
    "role": "user",
    "content": "查询北京天气"
  },
  {
    "role": "assistant",
    "content": "",
    "tool_calls": [
      {
        "id": "call_1",
        "type": "function",
        "function": {
          "name": "weather",
          "arguments": "{\"city\":\"北京\"}"
        }
      }
    ]
  },
  {
    "role": "tool",
    "tool_call_id": "call_1",
    "content": "北京：晴，25°C"
  }
]
```

最后这条最关键：

```json
{
  "role": "tool",
  "tool_call_id": "call_1",
  "content": "北京：晴，25°C"
}
```

它告诉 DeepSeek：你刚才请求的 `call_1` 工具调用已经执行完了，结果是 `北京：晴，25°C`。

然后 DeepSeek 会基于这个 observation 生成最终回答，例如：

```json
{
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "北京当前天气：晴，25°C。"
      }
    }
  ]
}
```

### DeepSeek tool calling 的完整链路

```text
用户输入
  ↓
Agent 构造 messages + tools
  ↓
第一次请求 DeepSeek
  ↓
DeepSeek 返回 tool_calls: weather({"city":"北京"})
  ↓
Agent 执行本地 Weather 工具
  ↓
得到 observation: 北京：晴，25°C
  ↓
第二次请求 DeepSeek，带上 tool_call_id 和工具结果
  ↓
DeepSeek 返回最终自然语言答案
  ↓
Agent 结束
```

所以第一个请求的核心目的不是“拿最终答案”，而是：

> 让 DeepSeek 根据用户目标和工具列表，决定下一步是否需要调用工具，以及调用哪个工具、传什么参数。

## 为什么 DeepSeek 能精准返回函数名称

DeepSeek 能精准返回 `calculator`、`weather` 这些函数名，核心原因不是它“读懂了你的 Go 代码”，而是每次请求时，程序都会把可用工具列表、工具名称、工具描述、参数 schema 一起发给 DeepSeek。

也就是说，DeepSeek 不是凭空猜函数名，而是在请求里的 `tools` 列表中做选择。

### 工具名称来自请求里的 function.name

请求中会包含类似这样的工具定义：

```json
{
  "type": "function",
  "function": {
    "name": "calculator",
    "description": "计算一个简单的二元四则运算表达式，例如 12 * 23",
    "parameters": {
      "type": "object",
      "properties": {
        "expression": {
          "type": "string",
          "description": "简单二元四则运算表达式，例如 12 * 23"
        }
      },
      "required": ["expression"]
    }
  }
}
```

以及：

```json
{
  "type": "function",
  "function": {
    "name": "weather",
    "description": "查询城市天气，当前返回 mock 数据",
    "parameters": {
      "type": "object",
      "properties": {
        "city": {
          "type": "string",
          "description": "城市名称，例如 北京、上海、深圳、杭州"
        }
      },
      "required": ["city"]
    }
  }
}
```

所以用户输入：

```text
帮我计算 12 * 23，然后查询北京天气，然后计算 10 * 20
```

模型可以把任务拆成：

```text
12 * 23        -> calculator({"expression":"12 * 23"})
查询北京天气    -> weather({"city":"北京"})
10 * 20        -> calculator({"expression":"10 * 20"})
```

DeepSeek 返回的 `tool_calls` 里，`function.name` 正是从这些工具定义里选择出来的：

```json
{
  "tool_calls": [
    {
      "id": "call_1",
      "type": "function",
      "function": {
        "name": "calculator",
        "arguments": "{\"expression\":\"12 * 23\"}"
      }
    }
  ]
}
```

这里的 `calculator` 必须和本地工具注册表里的名字一致。比如本地工具实现中：

```go
func (Calculator) Name() string {
    return "calculator"
}
```

Agent 才能通过下面的逻辑找到并执行对应工具：

```go
tool, ok := a.tools[decision.ToolName]
observation, err := tool.Execute(ctx, decision.Arguments)
```

所以精准匹配的关键是：

```text
DeepSeek 返回的 function.name == 本地 Tool.Name()
```

### System prompt 也会强化工具语义

请求里还会带上 system prompt：

```text
你可以使用工具完成任务：calculator 负责精确计算，clock 负责获取当前时间，weather 返回 mock 天气，echo 原样返回文本。
当用户的问题需要外部实时信息、确定性计算或项目提供的工具能力时，请优先调用合适的工具；工具返回 observation 后，再给出简洁的中文最终答案。
```

这段提示进一步告诉 DeepSeek：

- 计算任务应该使用 `calculator`
- 天气任务应该使用 `weather`
- 工具返回后再总结

因此 DeepSeek 不只是看到工具列表，还被明确告知这些工具适合什么场景。

### DeepSeek 返回的是调用意图，不是直接执行函数

DeepSeek 并不会直接执行 Go 函数。它只返回结构化的“调用意图”：

```json
{
  "name": "calculator",
  "arguments": "{\"expression\":\"12 * 23\"}"
}
```

真正执行函数的是本地 Agent：

```text
DeepSeek 返回 tool_calls
  ↓
Agent 读取 function.name
  ↓
Agent 在本地工具注册表里查找同名工具
  ↓
Agent 执行 Go 函数
```

### 为什么多轮请求还能继续精准选择

每一轮请求都会带上完整上下文，包括之前的工具调用和工具结果。

例如第二轮请求里会包含：

```json
{
  "role": "assistant",
  "tool_calls": [
    {
      "id": "call_1",
      "type": "function",
      "function": {
        "name": "calculator",
        "arguments": "{\"expression\":\"12 * 23\"}"
      }
    }
  ]
},
{
  "role": "tool",
  "content": "276",
  "tool_call_id": "call_1"
}
```

这告诉 DeepSeek：

```text
第一个计算任务已经完成，结果是 276。
```

于是模型会继续判断剩余任务，例如：

```text
查询北京天气
计算 10 * 20
```

并继续选择合适的工具。

### 当前实现一次只执行一个 tool_call

需要注意的是，DeepSeek 一次可能返回多个 `tool_calls`。例如它可能一次性返回：

```json
"tool_calls": [
  {
    "function": {
      "name": "calculator",
      "arguments": "{\"expression\": \"12 * 23\"}"
    }
  },
  {
    "function": {
      "name": "calculator",
      "arguments": "{\"expression\": \"10 * 20\"}"
    }
  },
  {
    "function": {
      "name": "weather",
      "arguments": "{\"city\": \"北京\"}"
    }
  }
]
```

这说明模型已经理解了全部任务。但当前 `DeepSeekProvider.Next()` 只取第一个 tool call：

```go
toolCall := message.ToolCalls[0]
```

所以程序实际一轮只执行一个工具。剩下的任务会在后续轮次中，由 DeepSeek 根据上下文重新规划并继续返回。

因此如果你看到 4 次 DeepSeek 交互，通常是：

```text
第 1 次：DeepSeek 返回一个或多个 tool_calls，程序执行第一个
第 2 次：DeepSeek 看到第一个工具结果后，继续返回下一个工具调用
第 3 次：继续执行剩余工具调用
第 4 次：所有任务完成，DeepSeek 返回最终答案
```

### 是否总能精准

不一定。虽然工具定义清楚时通常会很准，但模型仍然可能：

- 返回不存在的工具名
- 参数格式不符合 schema
- 选择了不合适的工具
- 一次返回多个 `tool_calls`，但当前 Agent 只处理第一个
- 把 `arguments` 生成成非法 JSON 字符串

所以 Agent 端必须校验。当前代码已经做了未知工具检查：

```go
tool, ok := a.tools[decision.ToolName]
if !ok {
    return Result{}, fmt.Errorf("unknown tool %q", decision.ToolName)
}
```

一句话总结：

> DeepSeek 不是凭空知道 Go 函数名，而是根据每次请求里提供的 `tools` 列表、函数描述、参数 schema 和历史 observation，生成符合 tool calling 协议的 `tool_calls`。

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

1. 添加一个 `file_search` 工具，搜索当前目录里的文件名。
2. 给 Agent 增加多轮对话模式。
3. 实现 `ollama` provider，接入本地模型。
4. 实现 `openai` provider，对比 function calling、Claude tool use 和 DeepSeek tool calling。
5. 给有副作用的工具增加人工确认机制。

## 当前版本的边界

- calculator 只支持 `number operator number` 格式。
- weather 是 mock 数据，不调用真实天气 API。
- 没有长期记忆。
- 没有真实网络工具。
- 没有复杂 JSON Schema 校验。

这些边界是有意保留的：第一版的重点是理解 Agent 的最小闭环。