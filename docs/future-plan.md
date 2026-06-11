总结当前项目有哪些能力，并指出下一步学习计划

当前项目已经从“简单 Go Agent 示例”演进成了一个比较完整的 AI Agent 学习型框架。

当前项目已有能力

1. Harness / Agent Runtime

核心运行时已经迁移到：

internal/harness/

它负责：

- 维护 messages 上下文
- 控制 MaxSteps
- 调用模型 provider
- 解析模型返回的 Decision
- 执行工具
- 支持一次处理多个 tool_calls
- 把 observation 回写到上下文
- 管理工具注册表
- 执行前经过 policy 检查
- 记录 trace

也就是说，现在项目的核心不再只是“一个 Agent Loop”，而是一个较清晰的 harness/runtime。

---
2. 多 Provider 支持

当前支持：

mock
anthropic / claude
deepseek

其中：

- mock 用于无 API Key 的本地教学和测试
- anthropic 使用 Claude tool use
- deepseek 使用 OpenAI-compatible tool calling 协议

Provider 抽象位于：

internal/model/provider.go

---
3. 工具系统

当前已有工具：

calculator
clock
weather
echo
memory
file_search
read_file

能力包括：

- 四则运算
- 当前时间
- mock 天气
- 文本回显
- 本地记忆读写
- 项目内文件搜索
- 项目内文件读取

file_search / read_file 已有安全边界：

- 只允许项目目录内路径
- 拒绝绝对路径
- 拒绝 .. 逃逸路径
- 跳过 .git、.claude、.cjadk 等目录

---
4. 多 tool_calls 支持

项目已经补齐了一个关键 Agent 能力：

模型一次返回多个 tool_calls
  ↓
harness 在同一轮依次执行多个工具
  ↓
把多个 observation 一起回写到上下文

这比“一次只处理一个工具调用”更接近真实 Agent。

---
5. Skill 系统

当前支持本地 Markdown skill：

skills/
├── agent-teacher.md
├── math-tutor.md
└── travel-planner.md

运行时会：

读取 skills/*.md
  ↓
根据用户 goal 做关键词匹配
  ↓
选中相关 skill
  ↓
注入 system prompt

Skill 负责提供专项任务策略，不直接执行动作。

---
6. Memory 系统

当前支持本地持久化记忆：

memory/memory.json

记忆有两种用法：

1. 工具式读取
  - 模型调用 memory 工具
  - 支持 set/get/list/delete
2. 自动注入上下文
  - 使用：
-memory-in-context
  - 会把记忆内容注入 system prompt 的 <memory> 块

---
7. Trace / 可观测性

当前有两类 trace：

1. Agent 执行 trace
  - -json
  - 输出 step、tool name、tool input、observation、final answer
2. 函数级 trace log
  - 默认开启
  - 输出：
{
  "location": "internal/harness/harness.go:56",
  "function": "harness.Run.step",
  ...
}

特点：

- location 放在最前面
- 只显示项目内相对路径
- API Key / token / secret 字段会掩码
- 方便从日志中提取 DeepSeek request / response

---
8. 日志提取能力

已有 docs/specs/log-extract.md 指示。

可以从 full-skill.log 中提取：

full-skill-request.json
full-skill-response.json

用于分析：

- 每次发给 DeepSeek 的请求
- 每次 DeepSeek 的响应
- system prompt
- skills
- memory context
- tools schema
- tool_calls

---
9. 文档和学习材料

README 已经包含大量教学内容：

- Agent Loop 原理
- Tool use / tool_calls
- DeepSeek 请求与响应格式
- 为什么模型能精准返回函数名
- Skill 原理
- Memory 原理
- Harness 原理
- 项目结构说明
- 示例命令

项目本身已经是一个可运行、可观察、可扩展的 Agent 教学工程。

---
建议下一步学习计划

我建议按下面路线继续。

---
第一步：让工具错误可恢复

当前工具失败时，大多是 harness 直接返回 error。

真实 Agent 更常见的方式是：

工具执行失败
  ↓
把错误作为 tool_result / observation 发回模型
  ↓
模型根据错误修正参数或换工具

例如：

read_file {"path":"../secret.txt"}

工具返回错误：

path must stay inside project root

模型下一轮可以改成：

file_search {"query":"secret"}

这是非常关键的 Agent 能力：自我修正。

建议下一步做：

工具错误不直接中断，而是作为 observation 发回模型

---
第二步：加入人工确认机制

现在 policy 已经有雏形：

type Policy interface {
    AllowTool(ctx context.Context, call model.ToolCall) error
}

下一步可以扩展：

read_file      allow
file_search    allow
calculator     allow
memory set     ask
write_file     ask
shell          ask / deny

这能帮助你理解：

模型提出动作 ≠ harness 必须执行

---
第三步：增加写文件工具 write_file

有了 policy 后，可以加入第一个有副作用工具：

write_file

但必须带安全限制：

- 只能写项目目录
- 拒绝覆盖敏感文件
- 默认需要确认
- 可选 dry-run

这是从“只读 Agent”走向“能改项目的 Agent”的关键一步。

---
第四步：实现 REPL 多轮对话

当前 CLI 基本是一轮命令。

下一步可以做：

go run ./cmd/agent -provider deepseek -repl

让用户连续输入：

> 记住我喜欢 Go
> 读取 README
> 根据 README 总结项目能力
> 接下来我该学什么？

这会帮助你学习：

- 会话级 messages 管理
- memory 和 messages 的区别
- 什么时候压缩上下文
- 多轮对话中的工具调用

---
第五步：做一个小型“项目分析 Agent”

目标：

给定一个 Go 项目，让 Agent 自动阅读 README、搜索关键文件、总结结构、提出下一步计划

它会用到：

- file_search
- read_file
- skills
- memory
- trace
- harness policy

这是当前项目最自然的综合练习。

---
我最推荐的下一步

建议你下一步做：

工具错误作为 observation 回传给模型，而不是直接中断

原因：

1. 改动不算太大
2. 非常贴近真实 Agent
3. 能马上通过 read_file 的错误路径验证
4. 是后续 write_file / shell / approval 的基础

推荐学习路线：

工具错误可恢复
  ↓
Policy 人工确认
  ↓
write_file 工具
  ↓
REPL 多轮对话
  ↓
项目分析 Agent