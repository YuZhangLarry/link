# Tool 调用系统实现指南

## 目录结构

```
internal/agent/tool/
├── types.go       # 类型定义
├── registry.go    # 工具注册中心
├── executor.go    # 工具执行器
├── agent.go       # Agent 主控制器
├── repository.go  # 数据库记录
├── init.go        # 工具初始化
├── kb_query.go    # 知识库查询工具
├── web_search.go  # 网络搜索等工具
└── example.go     # 使用示例
```

---

## 快速开始

### 1. 初始化 Agent

```go
import (
    "link/internal/agent/tool"
    "github.com/cloudwego/eino/components/model"
)

// 创建 ChatModel
chatModel := createYourChatModel() // 需要实现 model.BaseChatModel

// 创建 Agent
agent, err := tool.NewAgent(chatModel, &tool.AgentConfig{
    EnableTools:       true,
    MaxToolIterations: 5,
    ToolTimeout:       30 * time.Second,
})
```

### 2. 集成到 ChatService

```go
// 方式1：创建时传入 Agent
chatService := service.NewChatServiceWithAgent(chatConfig, agent)

// 方式2：动态设置 Agent
chatService := service.NewChatService(chatConfig)
chatService.SetAgent(agent)

// 方式3：动态启用/禁用工具
chatService.EnableTool(true)  // 启用
chatService.EnableTool(false) // 禁用
```

### 3. 进行对话（自动工具调用）

```go
resp, err := chatService.Chat(ctx, &types.ChatRequest{
    Content: "请帮我搜索 Go 语言并发编程的资料",
    History: []types.Message{},
    Options: &types.ChatOptions{
        Temperature: 0.7,
    },
})

fmt.Println(resp.Content)      // AI 回复
fmt.Println(resp.ToolCalls)    // 调用的工具（如果有）
```

---

## 内置工具列表

| 工具名称 | 功能描述 | 参数 |
|---------|---------|------|
| `kb_query` | 知识库查询 | query, kb_id, top_k, similarity, retrieval_mode |
| `kb_list` | 获取知识库列表 | user_id, status |
| `document_list` | 获取文档列表 | kb_id, limit |
| `web_search` | 网络搜索 | query, limit |
| `get_current_time` | 获取当前时间 | 无 |
| `calculator` | 计算器 | expression |
| `http_request` | HTTP 请求 | url, method, headers, body |

---

## 自定义工具

### 方式1：使用 InferTool（推荐）

```go
package mytool

import (
    "context"
    "github.com/cloudwego/eino/components/tool/utils"
)

// 1. 定义请求结构体（使用 jsonschema tag 描述参数）
type WeatherRequest struct {
    City string `json:"city" jsonschema:"required,description=城市名称"`
    Unit string `json:"unit" jsonschema:"description=温度单位,celsius 或 fahrenheit,default=celsius,enum=celsius,enum=fahrenheit"`
}

// 2. 定义结果结构体
type WeatherResult struct {
    City        string  `json:"city"`
    Temperature float64 `json:"temperature"`
    Description string  `json:"description"`
    Unit        string  `json:"unit"`
}

// 3. 实现工具函数
func GetWeather(ctx context.Context, req *WeatherRequest) (*WeatherResult, error) {
    // 实际的天气查询逻辑
    return &WeatherResult{
        City:        req.City,
        Temperature: 25.5,
        Description: "晴朗",
        Unit:        req.Unit,
    }, nil
}

// 4. 创建工具
func CreateWeatherTool() (tool.InvokableTool, error) {
    return utils.InferTool(
        "get_weather",
        "获取指定城市的天气信息",
        GetWeather,
    )
}
```

### 方式2：直接实现接口

```go
type MyTool struct{}

func (t *MyTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
    return &schema.ToolInfo{
        Name: "my_tool",
        Desc: "我的自定义工具",
        ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
            "param1": {
                Type:     schema.String,
                Required: true,
                Desc:     "参数1的描述",
            },
        }),
    }, nil
}

func (t *MyTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
    // 解析参数
    var args map[string]interface{}
    json.Unmarshal([]byte(argumentsInJSON), &args)

    // 执行逻辑
    result := doSomething(args["param1"])

    // 返回结果
    return json.Marshal(result)
}
```

---

## 工具注册

### 注册自定义工具

```go
// 创建 Agent 时指定工具列表
agent, err := tool.NewAgent(chatModel, &tool.AgentConfig{
    EnableTools: true,
    ToolNames: []string{
        "kb_query",
        "web_search",
        "get_weather", // 自定义工具
    },
})

// 或者动态添加工具
weatherTool, _ := CreateWeatherTool()
agent.RegisterTool("get_weather", weatherTool)

// 或者移除工具
agent.UnregisterTool("calculator")
```

---

## 数据库集成

### 保存工具执行记录

```go
import "link/internal/agent/tool"

// 创建 Repository
repo := tool.NewRepository(db)

// 保存执行记录
err := repo.SaveExecutionWithParams(ctx, messageID, toolName, inputParams, execResult)

// 获取执行历史
executions, err := repo.GetExecutionsByMessage(ctx, messageID)

// 获取工具统计
stats, err := repo.GetToolStats(ctx, "kb_query", 7) // 最近7天
```

---

## 高级用法

### 选择性启用工具

```go
// 只启用知识库相关工具
config := &tool.AgentConfig{
    EnableTools: true,
    ToolNames: []string{
        "kb_query",
        "kb_list",
        "document_list",
    },
}
agent, _ := tool.NewAgent(chatModel, config)
```

### 自定义系统提示词

```go
tools := agent.GetToolRegistry().GetTools()
systemPrompt := tool.BuildSystemPrompt(tools)

messages := []*schema.Message{
    {Role: schema.RoleSystem, Content: systemPrompt},
    {Role: schema.RoleUser, Content: "你好"},
}
```

### 带超时的工具执行

```go
executor := agent.GetExecutor()

// 执行单个工具（带超时）
result := executor.ExecuteWithTimeout(ctx, toolCall, 10*time.Second)

// 执行多个工具
results := executor.ExecuteAll(ctx, toolCalls)
```

---

## 数据库表设计

### tools 表

```sql
CREATE TABLE tools (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    config JSON NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_by BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

### tool_executions 表

```sql
CREATE TABLE tool_executions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    message_id BIGINT NOT NULL,
    tool_id BIGINT NOT NULL,
    input_params JSON,
    output_data JSON,
    status ENUM('success', 'failed', 'timeout'),
    duration_ms INT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(id) ON DELETE CASCADE,
    FOREIGN KEY (tool_id) REFERENCES tools(id)
);
```

---

## 完整调用流程

```
用户消息
    ↓
ChatService.Chat()
    ↓
判断是否启用工具
    ↓
[启用] → Agent.Chat()
         ↓
         1. 调用 BaseChatModel 生成响应
         ↓
         2. 检查是否有 ToolCalls
         ↓
         [有工具] → Executor.ExecuteAll()
                    ↓
                    3. 执行所有工具
                    ↓
                    4. 将工具结果作为新消息
                    ↓
                    5. 再次调用 BaseChatModel
                    ↓
                    6. 返回最终结果
         ↓
[未启用] → 普通 BaseChatModel 调用
    ↓
返回 ChatResponse
```

---

## 注意事项

1. **模型兼容性**：确保使用的 LLM 支持 Function Calling（如 GPT-4, Claude 3, Qwen 等）

2. **超时设置**：工具调用可能耗时较长，建议设置合理的超时时间

3. **错误处理**：工具执行失败时，会将错误信息返回给模型，模型会决定如何处理

4. **安全性**：
   - HTTP 工具请求要验证 URL 安全性
   - 搜索工具要过滤敏感关键词
   - 数据库操作要使用参数化查询

5. **性能优化**：
   - 工具执行可以并行
   - 缓存常用的工具结果
   - 限制最大迭代次数

---

## 参考资源

- [Eino 官方文档 - 如何创建 Tool](https://www.cloudwego.io/zh/docs/eino/core_modules/components/tools_node_guide/how_to_create_a_tool/)
- [Eino ReAct Agent 手册](https://www.cloudwego.io/zh/docs/eino/core_modules/flow_integration_components/react_agent_manual/)
- [OpenAI Function Calling](https://platform.openai.com/docs/guides/function-calling)
