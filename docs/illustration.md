# 我运行了一下程序，命令行参数如下：-provider deepseek "帮我计算 12 * 23，然后查询北京天气，然后计算 10 * 20" -json  。  该程序运行后，产生的日志为 full.log, 其中跟deepseek对话交互了4次，4次发往deepseek的请求分别记录到                   
  deepseek-request.json 文件中的四个key中，对应的deepseek返回的值存放到 deepseek-response.json 中。请帮忙解答我的疑问：deepseek每次回应，为什么能精准的返回需要Agent执行的函数名称呢？                                                       

⏺ 我先看一下这三个日志文件里的实际请求和响应，再结合代码解释 DeepSeek 为什么能稳定返回函数名。

  Read 3 files (ctrl+o to expand)                                                                                                                                                                                                            
  ⎿  PostToolUse:Read hook error                                                                                         ⎿  Failed with non-blocking status code: npm error code ENOTEMPTY
                                                                                                                                                                                                                                             
⏺ DeepSeek 能精准返回 calculator / weather 这些函数名，核心原因不是它“读懂了你的 Go 代码”，而是每次请求时，你都把可用工具列表、工具名称、工具描述、参数 schema 一起发给了 DeepSeek。                                                       
                                                 
  也就是说，DeepSeek 不是凭空猜函数名，而是在你给它的 tools 列表里做选择。                                                                                                                                                                   
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  1. 第一次请求里已经告诉了 DeepSeek 有哪些工具                                                                                                                                                                                              
                                                                                                                                                                                                                                             
  从你的 deepseek-request.json 第 1 个请求可以看到，发送给 DeepSeek 的请求里有：                                                                                                                                                           
                                                                                                                                                                                                                                             
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
          "required": ["city"]                                                                                                                                                                                                               
        }                                                                                                                                                                                                                                    
      }              
    },                                                                                                                                                                                                                                       
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
  ]                                                                                                                                                                                                                                          
                                                                                                                                                                                                                                           
  所以 DeepSeek 实际收到的信息是：                                                                                                                                                                                                           
                                         
  ▎ 你可以调用 calculator，它用于计算表达式，参数叫 expression。                                                                                                                                                                             
  ▎ 你可以调用 weather，它用于查询城市天气，参数叫 city。                                                                                                                                                                                  
                                                                                                                                                                                                                                             
  因此用户输入：                                                                                                                                                                                                                             
                                                                                                                                                                                                                                             
  帮我计算 12 * 23，然后查询北京天气，然后计算 10 * 20                                                                                                                                                                                       
                                                                                                                                                                                                                                             
  模型可以把任务拆成：                                                                                                                                                                                                                       
                                                                                                                                                                                                                                             
  12 * 23        -> calculator({"expression":"12 * 23"})                                                                                                                                                                                     
  查询北京天气    -> weather({"city":"北京"})
  10 * 20        -> calculator({"expression":"10 * 20"})                                                                                                                                                                                     
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  2. 工具名称来自 function.name                                                                                                                                                                                                              
                                                                                                                                                                                                                                             
  DeepSeek 返回的不是普通文本，而是 OpenAI-compatible tool calling 结构：
                                                                                                                                                                                                                                             
  "tool_calls": [                                                                                                                                                                                                                          
    {                                                                                                                                                                                                                                        
      "id": "call_00_abPRo7HYSkYRlKpBHsa59299",                                                                                                                                                                                              
      "type": "function",                                                                                                                                                                                                                    
      "function": {                                                                                                                                                                                                                          
        "name": "calculator",                                                                                                                                                                                                                
        "arguments": "{\"expression\": \"12 * 23\"}"                                                                                                                                                                                         
      }                                                                                                                                                                                                                                      
    }                                                                                                                                                                                                                                        
  ]                                                                                                                                                                                                                                          
                                                                                                                                                                                                                                           
  这里最关键的是：                                                                                                                                                                                                                           
                                         
  "function": {                                                                                                                                                                                                                              
    "name": "calculator"                                                                                                                                                                                                                     
  }                                                                                                                                                                                                                                          
                                                                                                                                                                                                                                             
  这个 calculator 正是你请求里给它的工具名：                                                                                                                                                                                                 
                                                                                                                                                                                                                                           
  "function": {                                                                                                                                                                                                                              
    "name": "calculator"                                                                                                                                                                                                                     
  }                  
                                                                                                                                                                                                                                             
  所以它能精准返回函数名，是因为：                                                                                                                                                                                                           
                     
  1. 请求里明确列出了合法工具名；                                                                                                                                                                                                            
  2. 每个工具有自然语言描述；                                                                                                                                                                                                              
  3. 每个工具有 JSON Schema 参数说明；                                                                                                                                                                                                       
  4. tool_choice: "auto" 允许模型自主选择工具；                                                                                                                                                                                              
  5. DeepSeek 的 tool calling 能力会倾向于从 tools 列表中选择合适工具，并按 schema 生成参数。                                                                                                                                                
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  3. System prompt 也在引导它使用工具                                                                                                                                                                                                        
                                                                                                                                                                                                                                             
  你的第一个请求里还有 system prompt：   
                                                                                                                                                                                                                                             
  你可以使用工具完成任务：calculator 负责精确计算，clock 负责获取当前时间，weather 返回 mock 天气，echo 原样返回文本。                                                                                                                       
  当用户的问题需要外部实时信息、确定性计算或项目提供的工具能力时，请优先调用合适的工具；工具返回 observation 后，再给出简洁的中文最终答案。                                                                                                  
                                                                                                                                                                                                                                             
  这段提示强化了工具语义：                                                                                                                                                                                                                   
                                                                                                                                                                                                                                             
  - 计算任务应该用 calculator                                                                                                                                                                                                                
  - 天气任务应该用 weather                                                                                                                                                                                                                 
  - 工具返回后再总结                                                                                                                                                                                                                         
                                         
  所以 DeepSeek 不只是看到工具列表，还被明确告知了这些工具各自适合什么场景。                                                                                                                                                                 
                                                                                                                                                                                                                                           
  ---                                                                                                                                                                                                                                        
  4. 它不是调用 Go 函数，而是返回“调用意图”                                                                                                                                                                                                
                                                                                                                                                                                                                                             
  DeepSeek 并不会直接执行你的 Go 函数。  
                                                                                                                                                                                                                                             
  它只返回：                                                                                                                                                                                                                                 
                                                                                                                                                                                                                                             
  {                                                                                                                                                                                                                                          
    "name": "calculator",                                                                                                                                                                                                                  
    "arguments": "{\"expression\":\"12 * 23\"}"                                                                                                                                                                                              
  }                                                                                                                                                                                                                                          
                                                                                                                                                                                                                                             
  这只是一个结构化的“调用意图”。                                                                                                                                                                                                             
                                                                                                                                                                                                                                           
  真正执行函数的是你的 Agent：                                                                                                                                                                                                               
                                                                                                                                                                                                                                           
  tool, ok := a.tools[decision.ToolName]                                                                                                                                                                                                     
  observation, err := tool.Execute(ctx, decision.Arguments)                                                                                                                                                                                  
                                                                                                                                                                                                                                             
  也就是：                                                                                                                                                                                                                                   
                                                                                                                                                                                                                                             
  DeepSeek 返回 tool_calls                                                                                                                                                                                                                 
    ↓                                                                                                                                                                                                                                        
  Agent 读取 function.name               
    ↓                                                                                                                                                                                                                                        
  Agent 在本地工具注册表里查找同名工具                                                                                                                                                                                                     
    ↓                                                                                                                                                                                                                                        
  Agent 执行 Go 函数                     
                                                                                                                                                                                                                                             
  所以精准匹配的关键是：DeepSeek 返回的 function.name 必须和本地工具注册表里的名字一致。                                                                                                                                                     
                     
  例如：                                                                                                                                                                                                                                     
                                                                                                                                                                                                                                           
  func (Calculator) Name() string {                                                                                                                                                                                                          
      return "calculator"                                                                                                                                                                                                                  
  }                                                                                                                                                                                                                                          
                                         
  DeepSeek 返回：                                                                                                                                                                                                                            
                                                                                                                                                                                                                                           
  "name": "calculator"                                                                                                                                                                                                                       
                                                                                                                                                                                                                                           
  Agent 才能找到这个工具。                                                                                                                                                                                                                   
                                         
  ---                                                                                                                                                                                                                                        
  5. 为什么第二、第三次还能继续精准选择？                                                                                                                                                                                                  
                                                                                                                                                                                                                                             
  因为每次请求都会带上完整上下文。       
                                                                                                                                                                                                                                             
  第二次请求里已经包含：                                                                                                                                                                                                                     
                                                                                                                                                                                                                                             
  {                                                                                                                                                                                                                                          
    "role": "assistant",                                                                                                                                                                                                                   
    "tool_calls": [                                                                                                                                                                                                                          
      {                                                                                                                                                                                                                                      
        "id": "call_00_abPRo7HYSkYRlKpBHsa59299",                                                                                                                                                                                            
        "type": "function",                                                                                                                                                                                                                  
        "function": {                                                                                                                                                                                                                      
          "name": "calculator",                                                                                                                                                                                                              
          "arguments": "{\"expression\": \"12 * 23\"}"                                                                                                                                                                                     
        }                                                                                                                                                                                                                                    
      }                                                                                                                                                                                                                                      
    ]                                                                                                                                                                                                                                        
  },                                                                                                                                                                                                                                         
  {                                                                                                                                                                                                                                        
    "role": "tool",                      
    "content": "276",                    
    "tool_call_id": "call_00_abPRo7HYSkYRlKpBHsa59299"                                                                                                                                                                                       
  }                                                                                                                                                                                                                                          
                                                                                                                                                                                                                                             
  这告诉 DeepSeek：                                                                                                                                                                                                                          
                                                                                                                                                                                                                                           
  ▎ 第一个计算任务已经完成，结果是 276。                                                                                                                                                                                                     
                                         
  于是它知道还有“查询北京天气”和“计算 10 * 20”没完成，所以继续返回：                                                                                                                                                                         
                                                                                                                                                                                                                                           
  "tool_calls": [                                                                                                                                                                                                                            
    {                                                                                                                                                                                                                                      
      "function": {  
        "name": "weather",               
        "arguments": "{\"city\": \"北京\"}"                                                                                                                                                                                                  
      }                                                                                                                                                                                                                                      
    }                                                                                                                                                                                                                                        
  ]                                                                                                                                                                                                                                          
                                                                                                                                                                                                                                           
  第三次请求又带上天气结果：                                                                                                                                                                                                                 
                                         
  {                                                                                                                                                                                                                                          
    "role": "tool",                                                                                                                                                                                                                        
    "content": "北京：晴，25°C",                                                                                                                                                                                                             
    "tool_call_id": "call_00_fESObNpRdpzDynfMOjxN5157"                                                                                                                                                                                       
  }                                                                                                                                                                                                                                          
                                                                                                                                                                                                                                             
  所以它知道天气也完成了，剩下：                                                                                                                                                                                                             
                                                                                                                                                                                                                                             
  计算 10 * 20                                                                                                                                                                                                                               
                                                                                                                                                                                                                                             
  于是返回：                                                                                                                                                                                                                                 
                                                                                                                                                                                                                                             
  "function": {                                                                                                                                                                                                                              
    "name": "calculator",                
    "arguments": "{\"expression\": \"10 * 20\"}"                                                                                                                                                                                             
  }                                                                                                                                                                                                                                          
                                                                                                                                                                                                                                             
  这就是多轮 Agent Loop 的作用：每一轮都把已经完成的工具调用和 observation 发回模型，模型据此决定下一步。                                                                                                                                    
                                                                                                                                                                                                                                           
  ---                                                                                                                                                                                                                                        
  6. 一个很重要的细节：DeepSeek 第一次其实返回了 3 个 tool_calls                                                                                                                                                                           
                                                                                                                                                                                                                                             
  从你的 deepseek-response.json 第 1 个响应可以看到，DeepSeek 第一次返回的是：
                                                                                                                                                                                                                                             
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
                                         
  也就是说，DeepSeek 一开始其实已经理解了全部三个任务。                                                                                                                                                                                      
  
  但当前程序在 DeepSeekProvider.Next() 里只取了第一个 tool call：                                                                                                                                                                            
                                                                                                                                                                                                                                           
  toolCall := message.ToolCalls[0]
                                                                                                                                                                                                                                             
  所以程序实际只执行了：
                                                                                                                                                                                                                                             
  calculator({"expression":"12 * 23"})                                                                                                                                                                                                     
                                                                                                                                                                                                                                             
  剩下两个第一次返回的 tool call 被当前实现忽略了。                                                                                                                                                                                          
                                                                                                                                                                                                                                             
  之后 DeepSeek 在第二、第三次请求中又根据上下文重新规划，继续返回 weather 和 calculator。                                                                                                                                                   
                                                                                                                                                                                                                                           
  所以你看到 4 次 DeepSeek 交互：                                                                                                                                                                                                            
                                                                                                                                                                                                                                           
  第 1 次：DeepSeek 其实返回了 3 个 tool_calls，但程序只执行第 1 个 calculator                                                                                                                                                               
  第 2 次：DeepSeek 看到 12*23 已完成，于是返回 weather                                                                                                                                                                                      
  第 3 次：DeepSeek 看到 weather 已完成，于是返回 10*20 的 calculator
  第 4 次：DeepSeek 看到所有任务完成，于是返回最终答案                                                                                                                                                                                       
                                                                                                                                                                                                                                             
  这说明 DeepSeek 的规划能力是对的，但当前 Agent 代码只支持“一轮执行一个工具调用”。                                                                                                                                                          
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  7. 是否一定总能精准？                                                                                                                                                                                                                    
                                                                                                                                                                                                                                             
  不一定，虽然通常会很准。               
                                                                                                                                                                                                                                             
  DeepSeek 能精准返回函数名，是因为工具定义清楚、名字语义明确、schema 简单。但模型仍然可能出现：                                                                                                                                             
                     
  - 返回不存在的工具名；                                                                                                                                                                                                                     
  - 参数格式不符合 schema；                                                                                                                                                                                                                
  - 选择了不合适的工具；                                                                                                                                                                                                                     
  - 一次返回多个 tool_calls，但 Agent 只处理第一个；
  - 把 arguments 生成成非法 JSON 字符串。                                                                                                                                                                                                    
                                                                                                                                                                                                                                             
  所以 Agent 端必须做校验。                                                                                                                                                                                                                  
                                                                                                                                                                                                                                             
  你现在的 Agent 已经有一层保护：                                                                                                                                                                                                          
                                         
  tool, ok := a.tools[decision.ToolName] 
  if !ok {                                                                                                                                                                                                                                   
      return Result{}, fmt.Errorf("unknown tool %q", decision.ToolName)
  }                                                                                                                                                                                                                                          
                                                                                                                                                                                                                                           
  也就是说，如果 DeepSeek 返回一个不存在的函数名，程序不会执行，而是报错。                                                                                                                                                                   
                                         
  ---                                                                                                                                                                                                                                        
  8. 总结                                                                                                                                                                                                                                  
                     
  DeepSeek 能精准返回需要 Agent 执行的函数名称，是因为：

  1. 请求里显式发送了工具列表                                                                                                                                                                                                                
    - calculator
    - weather                                                                                                                                                                                                                                
    - clock                                                                                                                                                                                                                                
    - echo           
  2. 每个工具都有清楚的自然语言描述      
    - calculator：计算表达式                                                                                                                                                                                                                 
    - weather：查询城市天气
  3. 每个工具都有 JSON Schema 参数定义                                                                                                                                                                                                       
    - calculator.expression                                                                                                                                                                                                                  
    - weather.city   
  4. system prompt 明确要求需要工具时优先调用工具                                                                                                                                                                                            
  5. tool_choice: "auto" 让模型在工具列表中自动选择                                                                                                                                                                                        
  6. 每一轮都会把工具调用和工具结果带回上下文
    - 模型知道哪些任务已经完成                                                                                                                                                                                                               
    - 因此能继续选择剩余任务需要的工具                                                                                                                                                                                                       
                                                                                                                                                                                                                                             
  一句话总结：                                                                                                                                                                                                                               
                                                                                                                                                                                                                                             
  ▎ DeepSeek 不是凭空知道 Go 函数名，而是根据你每次请求里提供的 tools 列表、函数描述、参数 schema 和历史 observation，生成符合 tool calling 协议的 tool_calls。   














  # skill 是什么？如何给这个项目增加skill功能的支持？
                                                   
⏺ Skill 可以理解为：给 Agent 按需加载的专项能力说明书。                                                                                                                                                                                      
                                                                                                                                                                                                                                           
  它不是工具，不直接执行代码；它更像一份“任务操作手册”或“领域知识包”。当用户问题触发某类任务时，Agent 把对应 skill 的内容放进模型上下文，让模型按这份说明来做事。                                                                            
                                         
  ---                                                                                                                                                                                                                                        
  1. Skill 和 Tool 的区别                                                                                                                                                                                                                  
                                                                                                                                                                                                                                             
  在你这个项目里，当前已经有：           
                                                                                                                                                                                                                                             
  Tool = 可执行函数                                                                                                                                                                                                                        
                                                                                                                                                                                                                                             
  例如：                                                                                                                                                                                                                                   
                                                                                                                                                                                                                                             
  calculator -> 真正执行计算                                                                                                                                                                                                                 
  weather    -> 真正返回天气 mock 数据                                                                                                                                                                                                       
  clock      -> 真正读取当前时间                                                                                                                                                                                                             
  echo       -> 真正回显文本                                                                                                                                                                                                                 
                                                                                                                                                                                                                                             
  而 skill 更像：                                                                                                                                                                                                                            
                                                                                                                                                                                                                                             
  Skill = 给模型看的专项说明                                                                                                                                                                                                                 
                                                                                                                                                                                                                                             
  例如：                                 
                                                                                                                                                                                                                                             
  math-tutor                                                                                                                                                                                                                                 
    说明：当用户想学习数学时，先解释思路，再给答案，不要只输出结果。                                                                                                                                                                         
                                                                                                                                                                                                                                             
  travel-planner                                                                                                                                                                                                                             
    说明：当用户让你规划旅行时，先确认目的地、天数、预算，再输出行程。                                                                                                                                                                       
                                                                                                                                                                                                                                             
  agent-teacher                                                                                                                                                                                                                              
    说明：当用户问 AI Agent 原理时，用“目标 -> 决策 -> 工具 -> 观察 -> 继续决策”的结构解释。                                                                                                                                                 
                                                                                                                                                                                                                                             
  所以：                                                                                                                                                                                                                                   
                                                                                                                                                                                                                                             
  ┌────────────┬────────────────────────────────────────┬─────────────┐                                                                                                                                                                      
  │    概念    │                  作用                  │   谁使用    │
  ├────────────┼────────────────────────────────────────┼─────────────┤                                                                                                                                                                      
  │ Tool       │ 执行动作 / 获取事实                    │ Go 程序执行 │                                                                                                                                                                    
  ├────────────┼────────────────────────────────────────┼─────────────┤                                                                                                                                                                      
  │ Skill      │ 提供任务策略 / 领域知识 / 输出格式指导 │ 大模型阅读  │                                                                                                                                                                      
  ├────────────┼────────────────────────────────────────┼─────────────┤                                                                                                                                                                      
  │ Provider   │ 连接具体大模型 API                     │ Go 程序调用 │                                                                                                                                                                      
  ├────────────┼────────────────────────────────────────┼─────────────┤                                                                                                                                                                      
  │ Agent Loop │ 控制循环、状态和工具执行               │ Go 程序控制 │                                                                                                                                                                    
  └────────────┴────────────────────────────────────────┴─────────────┘                                                                                                                                                                      
                                                                                                                                                                                                                                           
  ---                                                                                                                                                                                                                                        
  2. 为什么需要 Skill                                                                                                                                                                                                                      
                                                                                                                                                                                                                                             
  现在你的 Agent 每次只知道：            
                                                                                                                                                                                                                                             
  - system prompt                                                                                                                                                                                                                            
  - user message                                                                                                                                                                                                                             
  - tools                                                                                                                                                                                                                                    
  - 历史 tool result                                                                                                                                                                                                                       
                                                                                                                                                                                                                                             
  如果所有任务规则都塞进 system prompt，会变成：                                                                                                                                                                                             
                                                                                                                                                                                                                                             
  system prompt 越来越长                                                                                                                                                                                                                     
  每次请求都带所有规则                                                                                                                                                                                                                       
  不相关任务也消耗 token                                                                                                                                                                                                                     
  不同领域说明混在一起                                                                                                                                                                                                                       
                                                                                                                                                                                                                                             
  Skill 的作用是：                                                                                                                                                                                                                           
                                                                                                                                                                                                                                             
  只在相关任务发生时，把对应说明加入上下文                                                                                                                                                                                                 
                                                                                                                                                                                                                                             
  例如：                                                                                                                                                                                                                                     
                                                                                                                                                                                                                                             
  用户问：                                                                                                                                                                                                                                   
                                                                                                                                                                                                                                           
  请用教学方式讲解 12 * 23                                                                                                                                                                                                                   
                                                                                                                                                                                                                                             
  可以加载：                                                                                                                                                                                                                                 
                                                                                                                                                                                                                                             
  math-tutor skill                                                                                                                                                                                                                           
                                                                                                                                                                                                                                             
  用户问：                                                                                                                                                                                                                                   
                                                                                                                                                                                                                                             
  帮我理解 DeepSeek 的 tool_calls                                                                                                                                                                                                            
                                                                                                                                                                                                                                             
  可以加载：                                                                                                                                                                                                                                 
                                                                                                                                                                                                                                             
  agent-teacher skill                                                                                                                                                                                                                        
                                                                                                                                                                                                                                           
  这样模型得到的上下文更聚焦。                                                                                                                                                                                                               
                                         
  ---                                                                                                                                                                                                                                        
  3. 在这个项目里可以怎么设计 Skill                                                                                                                                                                                                        
                                                                                                                                                                                                                                             
  建议做一个简单本地版 skill 系统，不要一开始做复杂的插件系统。
                                                                                                                                                                                                                                             
  目录结构可以是：                                                                                                                                                                                                                           
                                                                                                                                                                                                                                             
  skills/                                                                                                                                                                                                                                    
  ├── agent-teacher.md                                                                                                                                                                                                                     
  ├── math-tutor.md                                                                                                                                                                                                                          
  └── travel-planner.md                  
                                                                                                                                                                                                                                             
  每个 skill 文件类似：                                                                                                                                                                                                                    
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  name: agent-teacher 
  description: 当用户想学习 AI Agent 原理、工具调用、模型 provider、trace log 时使用                                                                                                                                                         
  ---                                                                                                                                                                                                                                        
                      
  你是一个 AI Agent 教学助手。                                                                                                                                                                                                               
                                                                                                                                                                                                                                           
  回答时遵循：                                                                                                                                                                                                                               
  1. 先解释原理                                                                                                                                                                                                                            
  2. 再对应到本项目代码                                                                                                                                                                                                                      
  3. 最后给一个运行例子                                                                                                                                                                                                                      
  4. 尽量使用“用户目标 -> 模型决策 -> 工具调用 -> observation -> 最终答案”的结构                                                                                                                                                             
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  4. Skill 的核心数据结构                                                                                                                                                                                                                    
                                                                                                                                                                                                                                             
  可以新增：                             
                                                                                                                                                                                                                                             
  internal/skills/                                                                                                                                                                                                                           
  ├── skill.go        
  └── loader.go                                                                                                                                                                                                                              
                                                                                                                                                                                                                                             
  定义：                                                                                                                                                                                                                                     
                                                                                                                                                                                                                                             
  type Skill struct {                                                                                                                                                                                                                        
      Name        string                                                                                                                                                                                                                     
      Description string                 
      Content     string                                                                                                                                                                                                                     
  }                                                                                                                                                                                                                                        
                      
  然后实现：                                                                                                                                                                                                                                 
   
  func LoadDir(dir string) ([]Skill, error)                                                                                                                                                                                                  
                                                                                                                                                                                                                                             
  负责读取 skills/*.md。                                                                                                                                                                                                                     
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  5. Skill 怎么被选中                                                                                                                                                                                                                      
                                                                                                                                                                                                                                             
  第一版不要让大模型自己搜索 skill，可以先做简单规则匹配。
                                                                                                                                                                                                                                             
  例如：                                                                                                                                                                                                                                     
                                                                                                                                                                                                                                             
  func Select(goal string, skills []Skill) []Skill                                                                                                                                                                                           
                                                                                                                                                                                                                                             
  规则：                                 
                                                                                                                                                                                                                                             
  - 用户输入包含 agent、tool_calls、provider、trace，选 agent-teacher                                                                                                                                                                        
  - 用户输入包含 计算、数学、讲解，选 math-tutor
  - 用户输入包含 旅行、行程，选 travel-planner                                                                                                                                                                                               
                                                                                                                                                                                                                                             
  这相当于一个轻量 skill router。                                                                                                                                                                                                            
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  6. Skill 怎么进入模型上下文                                                                                                                                                                                                              
                                                                                                                                                                                                                                             
  当前 provider 的 system prompt 大概是：
                                                                                                                                                                                                                                             
  你是这个 Go 学习项目里的 DeepSeek provider。                                                                                                                                                                                             
  你可以使用工具完成任务...                                                                                                                                                                                                                  
                                                                                                                                                                                                                                           
  增加 skill 后，可以把选中的 skill 拼到 system prompt 后面：                                                                                                                                                                                
                                                                                                                                                                                                                                           
  你是这个 Go 学习项目里的 DeepSeek provider。                                                                                                                                                                                               
  你可以使用工具完成任务...                                                                                                                                                                                                                  
                                                                                                                                                                                                                                             
  下面是当前任务相关的 skills：                                                                                                                                                                                                              
                                                                                                                                                                                                                                           
  <skill name="agent-teacher">                                                                                                                                                                                                               
  你是一个 AI Agent 教学助手。                                                                                                                                                                                                             
  回答时遵循：                                                                                                                                                                                                                               
  1. 先解释原理                                                                                                                                                                                                                            
  2. 再对应到本项目代码                                                                                                                                                                                                                      
  3. 最后给一个运行例子                                                                                                                                                                                                                      
  </skill>                                                                                                                                                                                                                                   
                                                                                                                                                                                                                                             
  这样 DeepSeek / Claude 都能看到 skill 内容，并按它来回答。                                                                                                                                                                                 
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  7. 需要改哪些地方                                                                                                                                                                                                                        
                                                                                                                                                                                                                                             
  推荐实现路径：                         
                                                                                                                                                                                                                                             
  skills/*.md                                                                                                                                                                                                                                
          ↓           
  internal/skills.LoadDir()                                                                                                                                                                                                                  
          ↓                                                                                                                                                                                                                                
  internal/skills.Select(goal, allSkills)                                                                                                                                                                                                    
          ↓                                                                                                                                                                                                                                  
  Agent.Run() 根据 goal 选中 skills
          ↓                                                                                                                                                                                                                                  
  Provider.Next() 请求里带 selected skills                                                                                                                                                                                                 
          ↓                                                                                                                                                                                                                                  
  DeepSeek / Anthropic system prompt 拼入 skills                                                                                                                                                                                           
                                                                                                                                                                                                                                             
  具体文件：                                                                                                                                                                                                                                 
                                                                                                                                                                                                                                             
  skills/agent-teacher.md                                                                                                                                                                                                                    
  skills/math-tutor.md                                                                                                                                                                                                                       
                                                                                                                                                                                                                                             
  internal/skills/skill.go                                                                                                                                                                                                                   
  internal/skills/loader.go                                                                                                                                                                                                                  
  internal/skills/select.go                                                                                                                                                                                                                  
                                         
  internal/model/provider.go                                                                                                                                                                                                                 
  internal/model/anthropic.go                                                                                                                                                                                                              
  internal/model/deepseek.go                                                                                                                                                                                                                 
  internal/model/mock.go                 
  internal/agent/agent.go                                                                                                                                                                                                                    
  cmd/agent/main.go                                                                                                                                                                                                                        
  README.md                                                                                                                                                                                                                                  
                                                                                                                                                                                                                                           
  ---                                                                                                                                                                                                                                        
  8. 推荐的最小实现方案                                                                                                                                                                                                                    
                                                                                                                                                                                                                                             
  第一步：扩展 Request                   
                                                                                                                                                                                                                                             
  在 internal/model/provider.go 里把：                                                                                                                                                                                                       
                                                                                                                                                                                                                                             
  type Request struct {                                                                                                                                                                                                                      
      Messages []Message                                                                                                                                                                                                                     
      Tools    []ToolSpec                                                                                                                                                                                                                    
  }                                                                                                                                                                                                                                          
                                                                                                                                                                                                                                             
  改成：                                                                                                                                                                                                                                     
                                         
  type Request struct {                                                                                                                                                                                                                      
      Messages []Message                                                                                                                                                                                                                   
      Tools    []ToolSpec                                                                                                                                                                                                                    
      Skills   []SkillSpec               
  }                                                                                                                                                                                                                                          
                                                                                                                                                                                                                                           
  新增：                                                                                                                                                                                                                                     
                                                                                                                                                                                                                                           
  type SkillSpec struct {                                                                                                                                                                                                                    
      Name        string                                                                                                                                                                                                                     
      Description string
      Content     string                                                                                                                                                                                                                     
  }                                                                                                                                                                                                                                        
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                      
  第二步：实现本地 skill loader                                                                                                                                                                                                              
                                                                                                                                                                                                                                             
  读取 skills/*.md：  
                                                                                                                                                                                                                                             
  func LoadDir(dir string) ([]Skill, error)                                                                                                                                                                                                
                                                                                                                                                                                                                                             
  解析 frontmatter：                                                                                                                                                                                                                         
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  name: agent-teacher                                                                                                                                                                                                                      
  description: 当用户想学习 AI Agent 原理时使用                                                                                                                                                                                              
  ---                                                                                                                                                                                                                                        
  正文内容...                                                                                                                                                                                                                                
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  第三步：实现 skill selector                                                                                                                                                                                                              
                                                                                                                                                                                                                                             
  简单做关键词匹配：                                                                                                                                                                                                                       
                                                                                                                                                                                                                                             
  func Select(goal string, skills []Skill) []Skill
                                                                                                                                                                                                                                             
  第一版规则可以写死：                                                                                                                                                                                                                       
                                                                                                                                                                                                                                             
  agent/tool_calls/provider/trace -> agent-teacher                                                                                                                                                                                           
  计算/数学/讲解 -> math-tutor                                                                                                                                                                                                               
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  第四步：Agent 保存 skills                                                                                                                                                                                                                  
                                                                                                                                                                                                                                             
  agent.Config 增加：                    
                                                                                                                                                                                                                                             
  SkillDir string                                                                                                                                                                                                                          
                                                                                                                                                                                                                                             
  Agent 增加：                                                                                                                                                                                                                             
                                                                                                                                                                                                                                             
  skills []skills.Skill                                                                                                                                                                                                                      
                                                                                                                                                                                                                                             
  初始化时加载：                                                                                                                                                                                                                             
                                                                                                                                                                                                                                             
  loadedSkills, err := skills.LoadDir(config.SkillDir)                                                                                                                                                                                     
                                                                                                                                                                                                                                             
  每次 Run(goal) 时选择：                                                                                                                                                                                                                    
                                                                                                                                                                                                                                             
  selectedSkills := skills.Select(goal, a.skills)                                                                                                                                                                                            
                                                                                                                                                                                                                                             
  然后传给 provider：                                                                                                                                                                                                                        
                                                                                                                                                                                                                                             
  a.provider.Next(ctx, model.Request{                                                                                                                                                                                                        
      Messages: messages,                                                                                                                                                                                                                    
      Tools:    a.toolSpecs(ctx),        
      Skills:   selectedSkills,                                                                                                                                                                                                              
  })                                                                                                                                                                                                                                       
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  第五步：Provider 拼接 skill 到 system prompt
                                                                                                                                                                                                                                             
  DeepSeek 和 Anthropic 都可以共用一个格式：                                                                                                                                                                                               
                                                                                                                                                                                                                                             
  下面是当前任务相关的 skills：                                                                                                                                                                                                            
                                                                                                                                                                                                                                             
  <skill name="agent-teacher">                                                                                                                                                                                                               
  ...                                                                                                                                                                                                                                        
  </skill>                                                                                                                                                                                                                                   
                                                                                                                                                                                                                                           
  然后拼到 system prompt 末尾。                                                                                                                                                                                                              
                                         
  ---                                                                                                                                                                                                                                        
  9. 为什么这适合你的学习项目                                                                                                                                                                                                              
                                                                                                                                                                                                                                             
  这个设计保留了 Agent 的几个核心概念：  
                                                                                                                                                                                                                                             
  Tool 负责行动                                                                                                                                                                                                                              
  Skill 负责指导模型如何行动                                                                                                                                                                                                                 
  Provider 负责和模型通信                                                                                                                                                                                                                    
  Agent Loop 负责任务循环                                                                                                                                                                                                                    
                                                                                                                                                                                                                                             
  它也不会太复杂：                                                                                                                                                                                                                           
                                                                                                                                                                                                                                             
  - 不需要向量数据库                                                                                                                                                                                                                         
  - 不需要 MCP                                                                                                                                                                                                                             
  - 不需要远程插件系统                                                                                                                                                                                                                       
  - 不需要模型自己搜索文件                                                                                                                                                                                                                   
                                                                                                                                                                                                                                             
  但已经能展示真实 Agent 框架里的一个重要能力：                                                                                                                                                                                              
                                                                                                                                                                                                                                           
  按任务动态加载上下文                                                                                                                                                                                                                       
                                                                                                                                                                                                                                             
  ---                                                                                                                                                                                                                                        
  10. 一个例子                                                                                                                                                                                                                               
                                                                                                                                                                                                                                             
  用户输入：                                                                                                                                                                                                                               
                                                                                                                                                                                                                                             
  请讲解 DeepSeek 的 tool_calls 是怎么工作的                                                                                                                                                                                               
                                                                                                                                                                                                                                             
  selector 选中：                                                                                                                                                                                                                            
                                                                                                                                                                                                                                             
  agent-teacher                                                                                                                                                                                                                              
                                                                                                                                                                                                                                           
  请求给 DeepSeek 的 system prompt 变成：                                                                                                                                                                                                    
                                         
  你是这个 Go 学习项目里的 DeepSeek provider。                                                                                                                                                                                               
  你可以使用工具完成任务...                                                                                                                                                                                                                  
                      
  下面是当前任务相关的 skills：                                                                                                                                                                                                              
                                                                                                                                                                                                                                           
  <skill name="agent-teacher">                                                                                                                                                                                                               
  你是一个 AI Agent 教学助手。                                                                                                                                                                                                             
  回答时遵循：                                                                                                                                                                                                                               
  1. 先解释原理                                                                                                                                                                                                                              
  2. 再对应到本项目代码                                                                                                                                                                                                                      
  3. 最后给一个运行例子                                                                                                                                                                                                                      
  </skill>                                                                                                                                                                                                                                   
                                                                                                                                                                                                                                           
  模型就更倾向于给出教学型回答，而不是普通聊天式回答。                                                                                                                                                                                       
                                                                                                                                                                                                                                           
  ---                                                                                                                                                                                                                                        
  如果你要我继续实现，我建议下一步直接做这个最小版本：                                                                                                                                                                                     
                                                                                                                                                                                                                                             
  本地 Markdown skills + 关键词 selector + provider system prompt 注入
                                                                                                                                                                                                                                             
  不需要引入复杂依赖。    




























# harness 是什么？请详细讲解一下它的原理和用法。如果我想在该学习项目中引入harness，请详细讲解一下如何做

在 AI Agent 里，harness 可以理解为“Agent 的宿主运行时 / 执行外壳 / 控制框架”。

模型本身只会生成文本或结构化意图，例如：

{
  "tool_calls": [
    {
      "function": {
        "name": "read_file",
        "arguments": "{\"path\":\"README.md\"}"
      }
    }
  ]
}

但模型不会真的读取文件。
真正负责：

接收模型意图
校验参数
执行工具
记录日志
处理权限
管理上下文
把 observation 发回模型
控制循环终止

的那层程序，就可以叫 harness。

你现在这个 Go 项目，其实已经有一个最小 harness 了。

---
1. Harness 在 Agent 架构里的位置

可以把 Agent 系统拆成四层：

User
  ↓
Harness / Agent Runtime
  ↓
LLM Provider
  ↓
Model

更具体一点：

用户输入
  ↓
Harness 维护 messages / skills / memory / tools
  ↓
Provider 把请求发给 DeepSeek / Claude
  ↓
模型返回 final answer 或 tool_calls
  ↓
Harness 校验 tool_calls
  ↓
Harness 执行本地工具
  ↓
Harness 把 observation 写回 messages
  ↓
下一轮模型调用

所以 harness 不是模型，也不是工具本身，而是：

▎ 让模型、工具、上下文、权限、日志、记忆、终止条件协同工作的运行时。

---
2. 为什么需要 Harness

如果没有 harness，模型只能回答文本：

你应该读取 README.md。

但有了 harness，模型可以提出结构化动作：

{
  "name": "read_file",
  "arguments": "{\"path\":\"README.md\"}"
}

然后 harness 做这些事：

1. 检查 read_file 是否存在
2. 检查路径是否安全
3. 执行 Go 函数
4. 得到文件内容
5. 把结果作为 observation 发回模型
6. 让模型继续推理或总结

也就是：

模型负责“想做什么”
Harness 负责“能不能做、怎么做、做完怎么继续”
工具负责“实际动作”

---
3. Harness 和 Tool 的区别

┌──────────┬───────────────────────┐
│   概念   │         作用          │
├──────────┼───────────────────────┤
│ Model    │ 生成回答或 tool_calls │
├──────────┼───────────────────────┤
│ Tool     │ 执行某个具体动作      │
├──────────┼───────────────────────┤
│ Provider │ 连接具体模型 API      │
├──────────┼───────────────────────┤
│ Harness  │ 控制整个运行流程      │
├──────────┼───────────────────────┤
│ Skill    │ 给模型看的专项说明    │
├──────────┼───────────────────────┤
│ Memory   │ 持久化跨会话信息      │
└──────────┴───────────────────────┘

举例：

用户：读取 README 并总结

模型返回：

tool_call read_file {"path":"README.md"}

此时：

- read_file 工具负责真正读文件
- harness 负责决定是否允许读、调用哪个工具、把结果放回上下文、继续下一轮
- provider 负责 DeepSeek / Claude API 请求格式转换
- model 负责下一步是否继续搜索或最终回答

---
4. 你当前项目里的 Harness 是什么

你现在的核心 harness 主要在：

internal/agent/agent.go

尤其是：

func (a *Agent) Run(ctx context.Context, goal string) (Result, error)

它已经做了 harness 的核心工作：

messages := []model.Message{{Role: model.RoleUser, Content: goal}}

for step := 1; step <= a.config.MaxSteps; step++ {
    decision, err := a.provider.Next(ctx, model.Request{
        Messages:      messages,
        Tools:         a.toolSpecs(ctx),
        Skills:        skillSpecs(selectedSkills),
        MemoryContext: a.memoryContext(),
    })

    switch decision.Type {
    case model.DecisionFinal:
        return result, nil

    case model.DecisionToolCall:
        calls := decision.Calls()
        for _, call := range calls {
            tool := a.tools[call.ToolName]
            observation := tool.Execute(ctx, call.Arguments)
            messages append assistant tool_call
            messages append tool observation
        }
    }
}

这就是一个最小 harness：

维护 messages
控制 max steps
调用 provider
执行工具
处理多个 tool_calls
记录 trace
返回 final answer

---
5. 当前 harness 已经具备哪些能力

你的项目已经有很多 harness 能力：

5.1 Tool registry

tools.Registry(ctx, registeredTools...)

作用：

把工具名 -> 工具实例 存起来

例如：

calculator -> tools.Calculator{}
weather    -> tools.Weather{}
read_file  -> tools.ReadFile{}

当模型返回：

"name": "read_file"

harness 就能找到对应工具。

---
5.2 Agent Loop

for step := 1; step <= MaxSteps; step++

作用：

防止模型无限调用工具

没有这个限制，模型可能一直调用工具，不给最终答案。

---
5.3 Provider abstraction

type Provider interface {
    Next(ctx context.Context, req Request) (Decision, error)
}

作用：

Agent 不关心底层是 mock、DeepSeek 还是 Claude

Harness 只接收统一的：

DecisionFinal
DecisionToolCall

---
5.4 多工具调用

你已经补齐了：

calls := decision.Calls()
for _, call := range calls {
    ...
}

所以当 DeepSeek 一次返回多个 tool_calls，harness 能一次执行多个工具。

---
5.5 Trace log

trace.Log(ctx, ...)

这是 harness 的可观测性能力。

它能告诉你：

模型请求发了什么
模型返回了什么
Agent 执行了哪个工具
工具参数是什么
observation 是什么
代码位置在哪里

---
5.6 Skill 注入

selectedSkills := skills.Select(ctx, goal, a.skills)

Harness 根据用户目标选择 skill，然后把它放进模型上下文。

---
5.7 Memory 注入 / memory 工具

MemoryContext
tools.Memory

Harness 可以：

- 让模型按需调用 memory 工具
- 或者通过 -memory-in-context 自动把记忆注入 system prompt

---
5.8 工具安全边界

例如 read_file 和 file_search：

只允许项目目录内路径
拒绝绝对路径
拒绝 ..
跳过 .git / .claude / .cjadk

这就是 harness + tool 共同实现的安全控制。

---
6. 一个完整 Harness 执行流程

以：

go run ./cmd/agent -provider deepseek "请搜索 README 并总结项目能力"

为例：

1. CLI 解析参数
   provider=deepseek
   goal=请搜索 README 并总结项目能力

2. Harness 初始化
   加载 tools
   加载 skills
   加载 memory 配置
   建立 trace logger

3. Harness 构造 messages
   user: 请搜索 README 并总结项目能力

4. Harness 调用 provider.Next()
   provider 把 messages/tools/skills 转成 DeepSeek 请求

5. DeepSeek 返回 tool_calls
   file_search {"query":"README"}

6. Harness 校验工具
   工具是否存在？
   参数是否合法？

7. Harness 执行 file_search
   observation: README.md

8. Harness 把 observation 加回 messages

9. 下一轮 provider.Next()
   DeepSeek 看到 README.md，可能返回：
   read_file {"path":"README.md"}

10. Harness 执行 read_file

11. Harness 把 README 内容作为 observation 加回 messages

12. 下一轮 provider.Next()
   DeepSeek 输出最终总结

13. Harness 返回 Result

这就是一个完整 Agent harness 的职责。

---
7. 如果要在该学习项目中“正式引入 harness”，应该怎么做

你现在虽然已经有 harness，但它还散在 agent.go、CLI、tools、trace 里。
如果想把概念更清楚地引入项目，可以做一次结构化重构。

目标是让项目结构变成：

internal/harness/
├── harness.go
├── config.go
├── result.go
├── registry.go
├── policy.go
└── events.go

然后让：

internal/agent

变得更薄，甚至可以改名为：

internal/harness

---
8. 推荐的 Harness 目录设计

8.1 internal/harness/config.go

定义运行时配置：

type Config struct {
    MaxSteps        int
    SkillDir        string
    MemoryPath      string
    MemoryInContext bool
    TraceLog        bool
}

现在这些配置分散在：

cmd/agent/main.go
internal/agent/types.go

可以统一放到 harness config。

---
8.2 internal/harness/harness.go

核心运行时：

type Harness struct {
    provider Provider
    tools    map[string]tools.Tool
    skills   []skills.Skill
    memory   *memory.Store
    policy   Policy
    config   Config
}

func (h *Harness) Run(ctx context.Context, goal string) (Result, error)

这就是你现在 Agent.Run() 的升级版。

---
8.3 internal/harness/events.go

现在 trace 是直接打日志。
更进一步，可以定义结构化事件：

type Event struct {
    Step      int
    Type      string
    ToolName  string
    Input     json.RawMessage
    Output    string
    Error     string
}

例如：

model_request
model_response
tool_call
tool_result
final_answer

这样以后可以：

- 打印 CLI trace
- 输出 JSON trace
- 存日志文件
- 做 Web UI
- 做测试断言

---
8.4 internal/harness/policy.go

这个很重要。

当前工具只要模型返回，harness 就执行。

未来如果加入：

write_file
delete_file
shell
http_request
send_email

就必须有 policy。

例如：

type Policy interface {
    AllowTool(ctx context.Context, call model.ToolCall) (Decision, error)
}

可以先做：

type AllowAllPolicy struct{}

后面扩展：

type ConfirmPolicy struct{}
type DenyDangerousPolicy struct{}

例如：

read_file    allow
file_search  allow
calculator   allow
weather      allow
write_file   ask
delete_file  deny
shell         ask

这就是真正 harness 和普通脚本的区别。

---
8.5 internal/harness/registry.go

可以把工具注册做得更明确：

type Registry struct {
    tools map[string]tools.Tool
}

func (r *Registry) Get(name string) (tools.Tool, bool)
func (r *Registry) Specs() []model.ToolSpec

现在这部分在：

tools.Registry()
agent.toolSpecs()

可以集中管理。

---
9. 具体重构路线

我建议分 5 步做，不要一次大改。

---
第一步：保留现有 internal/agent，只引入 harness 概念文档

先在 README 加一节：

## Harness 是什么

说明当前 internal/agent/agent.go 就是最小 harness。

这一步不改代码。

---
第二步：把 agent.Config 扩展成 harness.Config

新增：

internal/harness/config.go

先只移动配置类型，不动运行逻辑。

type Config struct {
    MaxSteps        int
    SkillDir        string
    MemoryInContext bool
    MemoryContext   string
}

然后 agent.Config 可以暂时 alias：

type Config = harness.Config

这样改动小。

---
第三步：抽出 Tool Registry

新增：

internal/harness/registry.go

把：

a.toolSpecs(ctx)

和：

tools.Registry()

逐步迁移进去。

目标：

registry := harness.NewToolRegistry(tools...)
registry.Specs()
registry.Get(name)

---
第四步：抽出 Policy

新增：

internal/harness/policy.go

第一版：

type Policy interface {
    AllowTool(ctx context.Context, call model.ToolCall) error
}

type AllowAllPolicy struct{}

func (AllowAllPolicy) AllowTool(ctx context.Context, call model.ToolCall) error {
    return nil
}

在执行工具前加：

if err := h.policy.AllowTool(ctx, call); err != nil {
    return Result{}, err
}

这一步很有学习价值，因为它体现了：

模型提出动作，不代表程序必须执行

---
第五步：把 Agent 改名或包裹成 Harness

两种方案：

方案 A：保守

保留：

internal/agent

但 README 里说明：

本项目里的 Agent 类型同时承担 harness/runtime 职责。

优点：改动小。

方案 B：更清晰

新增：

internal/harness

把 Agent 迁移为：

type Harness struct {}

CLI 从：

agent.New(...)

改成：

harness.New(...)

优点：概念清楚。
缺点：改动较多。

我建议你先选 方案 A，因为这个项目还是学习项目，不必过早做大重构。

---
10. 如果现在就想“引入 harness”，最小实现是什么

最小有价值版本是加一个 policy 层。

因为这能马上体现 harness 的核心职责之一：

工具调用前的安全控制

可以新增：

internal/harness/policy.go

内容：

package harness

import (
    "context"
    "fmt"

    "github.com/zieckey/ai-study/internal/model"
)

type Policy interface {
    AllowTool(ctx context.Context, call model.ToolCall) error
}

type AllowAllPolicy struct{}

func (AllowAllPolicy) AllowTool(ctx context.Context, call model.ToolCall) error {
    return nil
}

type StaticPolicy struct {
    Denied map[string]bool
}

func (p StaticPolicy) AllowTool(ctx context.Context, call model.ToolCall) error {
    if p.Denied[call.ToolName] {
        return fmt.Errorf("tool %q is denied by harness policy", call.ToolName)
    }
    return nil
}

然后在 Agent 里加：

policy harness.Policy

执行工具前：

if err := a.policy.AllowTool(ctx, call); err != nil {
    return Result{}, err
}

这样你的 Agent 就多了一个非常重要的 harness 能力：

不是模型想调什么就调什么
而是 harness 先检查是否允许

---
11. Harness 的最终学习路线

如果你想继续沿着 harness 学，推荐顺序是：

1. Policy：工具调用前的 allow/deny
2. Event：统一记录 model_request/tool_call/tool_result
3. Approval：危险工具人工确认
4. Sandbox：限制文件路径、网络、命令执行范围
5. Replay：用日志重放一次 Agent 执行过程
6. Eval：用固定任务测试 Agent 是否按预期调用工具

其中最适合你当前项目的下一步是：

增加 harness policy 层

因为你已经有：

read_file
file_search
memory

接下来再加 write_file 或 shell 时，就必须有 policy 了。