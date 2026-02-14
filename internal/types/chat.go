package types

// ========================================
// Model Types
// ========================================

// ModelSource 模型源
type ModelSource string

const (
	// ModelSourceLocal 本地模型（如Ollama）
	ModelSourceLocal ModelSource = "local"
	// ModelSourceRemote 远程API
	ModelSourceRemote ModelSource = "remote"
)

// ========================================
// Chat Types
// ========================================

// ChatRequest 聊天请求
type ChatRequest struct {
	SessionID string       `json:"session_id,omitempty"` // 会话ID（可选，用于继续对话）
	KBID      int64        `json:"kb_id"`                // 知识库ID（可选）
	Content   string       `json:"content"`              // 消息内容
	Stream    bool         `json:"stream"`               // 是否流式响应
	History   []Message    `json:"history,omitempty"`    // 历史消息
	Options   *ChatOptions `json:"options,omitempty"`    // 聊天选项
}

// Message 聊天消息
type Message struct {
	Role    string `json:"role"`    // system/user/assistant/tool
	Content string `json:"content"` // 消息内容
}

// ChatOptions 聊天选项（简化版）
type ChatOptions struct {
	Temperature      float64 `json:"temperature,omitempty"`       // 温度
	MaxTokens        int     `json:"max_tokens,omitempty"`        // 最大token数
	TopP             float64 `json:"top_p,omitempty"`             // Top P
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"` // 频率惩罚
	PresencePenalty  float64 `json:"presence_penalty,omitempty"`  // 存在惩罚
	Thinking         bool    `json:"thinking,omitempty"`          // 是否启用思考
}

// ChatResponse 聊天响应
type ChatResponse struct {
	MessageID    string     `json:"message_id"`    // 消息ID
	Content      string     `json:"content"`       // 响应内容
	Role         string     `json:"role"`          // 角色
	TokenCount   int        `json:"token_count"`   // Token数量
	ToolCalls    []ToolCall `json:"tool_calls"`    // 工具调用
	FinishReason string     `json:"finish_reason"` // 结束原因
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // function
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// StreamChatEvent 流式聊天事件
type StreamChatEvent struct {
	Event      string     `json:"event"`       // start/content/end/error
	Content    string     `json:"content"`     // 内容片段
	MessageID  string     `json:"message_id"`  // 消息ID
	TokenCount int        `json:"token_count"` // Token数量
	ToolCalls  []ToolCall `json:"tool_calls"`  // 工具调用
	Error      string     `json:"error"`       // 错误信息
}
