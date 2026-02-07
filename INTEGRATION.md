# Chat API 集成指南

## 快速开始

### 1. 启动服务器

```bash
# 方式1: 直接运行
go run cmd/server/main.go

# 方式2: 编译后运行
go build -o server.exe ./cmd/server
./server.exe
```

服务器启动后，访问 http://localhost:8080

### 2. 验证服务

```bash
# 健康检查
curl http://localhost:8080/health
```

## API 接口

### 基础URL
```
http://localhost:8080/api/v1
```

### 可用接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/health` | GET | 健康检查 |
| `/chat` | POST | 非流式聊天 |
| `/chat/stream` | POST | 流式聊天(SSE) |
| `/chat/auth` | POST | 带认证的聊天 |
| `/chat/auth/stream` | POST | 带认证的流式聊天 |

## Apifox 配置

### 方式1: 手动配置

#### 非流式聊天接口

**请求配置:**
- 方法: `POST`
- URL: `http://localhost:8080/api/v1/chat`
- Content-Type: `application/json`

**请求体示例:**
```json
{
  "content": "你好",
  "stream": false,
  "options": {
    "temperature": 0.7,
    "max_tokens": 1000
  }
}
```

#### 流式聊天接口 (SSE)

**请求配置:**
- 方法: `POST`
- URL: `http://localhost:8080/api/v1/chat/stream`
- Content-Type: `application/json`

**请求体示例:**
```json
{
  "content": "请写一首关于春天的诗",
  "stream": true,
  "options": {
    "temperature": 0.8,
    "max_tokens": 500
  }
}
```

**在Apifox中配置SSE:**
1. 点击接口设置
2. 找到"响应"或"Response"设置
3. 启用"流式响应"或"SSE支持"
4. Apifox会自动解析并显示流式事件

### 方式2: 导入OpenAPI规范

将 `API.md` 文件中的OpenAPI JSON导入Apifox

## 测试方式

### 1. Web界面测试 (推荐)

打开浏览器访问:
```
web/chat-demo.html
```

功能:
- ✅ 实时流式对话
- ✅ 消息历史记录
- ✅ 可调整参数
- ✅ 美观的UI界面

### 2. 命令行测试

**Windows:**
```bash
test-api.bat
```

**Linux/Mac:**
```bash
chmod +x test-api.sh
./test-api.sh
```

### 3. curl 测试

**非流式:**
```bash
curl -X POST http://localhost:8080/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"content": "你好"}'
```

**流式:**
```bash
curl -X POST http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"content": "写一首诗", "stream": true}'
```

### 4. JavaScript 测试

```javascript
async function chat() {
  const response = await fetch('http://localhost:8080/api/v1/chat/stream', {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify({
      content: '你好',
      stream: true
    })
  });

  const reader = response.body.getReader();
  const decoder = new TextDecoder();

  while (true) {
    const {done, value} = await reader.read();
    if (done) break;

    const chunk = decoder.decode(value);
    const lines = chunk.split('\n');

    for (const line of lines) {
      if (line.startsWith('data: ')) {
        const data = JSON.parse(line.substring(6));
        if (data.event === 'content') {
          console.log(data.content);
        }
      }
    }
  }
}

chat();
```

## SSE 事件格式

流式响应返回 Server-Sent Events 格式:

```
data: {"event":"start","content":"","message_id":"msg_123","token_count":0}

data: {"event":"content","content":"你","message_id":"msg_123","token_count":1}

data: {"event":"content","content":"好","message_id":"msg_123","token_count":1}

data: {"event":"end","content":"","message_id":"msg_123","token_count":0}
```

### 事件类型

| event | 说明 |
|-------|------|
| `start` | 开始流式传输 |
| `content` | 内容片段 |
| `end` | 传输完成 |
| `error` | 错误信息 |

## 请求参数

### 基础参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| content | string | 是 | 消息内容 |
| stream | boolean | 是 | 是否流式响应 |
| history | array | 否 | 对话历史 |
| options | object | 否 | 生成选项 |

### Options 参数

| 参数 | 类型 | 范围 | 说明 |
|------|------|------|------|
| temperature | number | 0-2 | 温度参数，越高越随机 |
| max_tokens | integer | 1-32000 | 最大生成token数 |
| top_p | number | 0-1 | Top P采样 |
| presence_penalty | number | -2到2 | 存在惩罚 |
| frequency_penalty | number | -2到2 | 频率惩罚 |

### 对话历史格式

```json
{
  "content": "我叫什么名字？",
  "history": [
    {"role": "user", "content": "我的名字叫小明"},
    {"role": "assistant", "content": "你好小明！"}
  ]
}
```

## 响应格式

### 非流式响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "message_id": "msg_1234567890",
    "content": "你好！我是AI助手",
    "role": "assistant",
    "token_count": 25,
    "finish_reason": "stop"
  }
}
```

### 错误响应

```json
{
  "code": -1,
  "message": "错误描述",
  "error": "详细错误信息"
}
```

## 当前配置

根据 `.env` 配置:
- Provider: OpenAI
- BaseURL: https://api.gpts.vin/v1
- Model: gpt-3.5-turbo
- API Key: 已配置

## 故障排查

### 问题1: 无法连接服务器

**解决方案:**
```bash
# 检查服务器是否运行
curl http://localhost:8080/health

# 查看服务器日志
# 检查端口8080是否被占用
```

### 问题2: API返回错误

**解决方案:**
- 检查 `.env` 中的API密钥配置
- 确认API服务是否可用
- 查看服务器日志

### 问题3: SSE无响应

**解决方案:**
- 确保 `stream: true`
- 检查Content-Type是否为 `application/json`
- 使用Apifox或curl测试

## 项目文件

| 文件 | 说明 |
|------|------|
| `API.md` | 完整API文档 |
| `web/chat-demo.html` | Web演示界面 |
| `test-api.bat` | Windows测试脚本 |
| `test-api.sh` | Linux/Mac测试脚本 |
| `cmd/server/main.go` | 服务器主程序 |
| `internal/handler/chat.go` | Chat处理器 |
| `internal/models/chat/` | Chat核心实现 |

## 下一步

1. ✅ 启动服务器
2. ✅ 使用Apifox测试接口
3. ✅ 打开Web界面体验
4. ⏳ 集成到你的应用中

## 技术支持

- 查看完整文档: `API.md`
- 运行单元测试: `go test -v ./internal/models/chat`
- 运行测试程序: `go run cmd/test-sse/main.go stream`
