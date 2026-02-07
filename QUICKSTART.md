# 🚀 Chat API - 快速开始

## ✅ 服务器已启动

服务器正在运行: `http://localhost:8080`

## 📡 可用接口

### 1. 健康检查
```
GET http://localhost:8080/health
```

### 2. 非流式聊天
```
POST http://localhost:8080/api/v1/chat
Content-Type: application/json

{
  "content": "你好",
  "stream": false
}
```

### 3. 流式聊天 (SSE) ⭐
```
POST http://localhost:8080/api/v1/chat/stream
Content-Type: application/json

{
  "content": "请写一首诗",
  "stream": true,
  "options": {
    "temperature": 0.7,
    "max_tokens": 500
  }
}
```

## 🎨 使用方式

### 方式1: Web界面 (最简单)

直接双击打开:
```
web/chat-demo.html
```

### 方式2: Apifox

**非流式接口配置:**
- 方法: `POST`
- URL: `http://localhost:8080/api/v1/chat`
- Body:
```json
{
  "content": "你好",
  "stream": false
}
```

**流式接口配置 (SSE):**
- 方法: `POST`
- URL: `http://localhost:8080/api/v1/chat/stream`
- Body:
```json
{
  "content": "写一首诗",
  "stream": true
}
```
- ⚠️ 在Apifox中启用"SSE支持"或"流式响应"

### 方式3: curl 测试

```bash
# Windows
curl -X POST http://localhost:8080/api/v1/chat/stream ^
  -H "Content-Type: application/json" ^
  -d "{\"content\": \"你好\", \"stream\": true}"

# Linux/Mac
curl -X POST http://localhost:8080/api/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"content": "你好", "stream": true}'
```

## 📝 SSE 响应示例

```
event: content
data: {"event":"content","content":"你","message_id":"msg_123","token_count":0}

event: content
data: {"event":"content","content":"好","message_id":"msg_123","token_count":0}

event: content
data: {"event":"content","content":"！","message_id":"msg_123","token_count":0}

event: end
data: {"event":"end","content":"","message_id":"msg_123","token_count":0}
```

## 📊 测试结果

刚才的测试显示:
- ✅ 服务器运行正常
- ✅ 健康检查通过
- ✅ SSE流式接口工作正常
- ✅ AI响应正确接收

## 🔧 当前配置

- Provider: OpenAI
- BaseURL: https://api.gpts.vin/v1
- Model: gpt-3.5-turbo
- API Key: 已配置

## 📚 相关文件

| 文件 | 说明 |
|------|------|
| `API.md` | 完整API文档 |
| `INTEGRATION.md` | 集成指南 |
| `web/chat-demo.html` | Web演示界面 |
| `test-api.bat` | Windows测试脚本 |

## 🎯 下一步

1. ✅ 在Apifox中配置接口
2. ✅ 测试非流式接口
3. ✅ 测试SSE流式接口
4. ✅ 集成到你的应用

## 💡 提示

- SSE接口会持续推送事件，直到收到`event: end`
- 每个事件都是JSON格式的`data:`行
- 建议在Apifox中启用SSE支持以获得最佳体验
- Web界面提供最直观的测试方式

## 🆘 遇到问题?

1. **服务器未运行**: 执行 `go run cmd/server/main.go`
2. **端口被占用**: 修改 `cmd/server/main.go` 中的端口号
3. **API错误**: 检查 `.env` 中的配置
4. **查看文档**: `API.md` 或 `INTEGRATION.md`

---

**服务器状态**: 🟢 运行中
**最后测试**: ✅ SSE流式接口正常
