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