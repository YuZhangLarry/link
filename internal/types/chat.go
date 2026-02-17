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
// RAG Types
// ========================================

// RAGConfig RAG 检索配置（简化版，仅包含检索相关设置）
type RAGConfig struct {
	// 是否启用 RAG
	Enabled bool `json:"enabled"`

	// 知识库 ID
	KBID string `json:"kb_id"`

	// 检索模式（可多选，向量检索必选）
	RetrievalModes []string `json:"retrieval_modes"` // 检索模式：vector(必选), bm25, graph

	// 检索参数
	VectorTopK          int     `json:"vector_top_k"`         // 向量检索返回数量
	KeywordTopK         int     `json:"keyword_top_k"`        // 关键词检索返回数量
	GraphTopK           int     `json:"graph_top_k"`          // 图谱检索返回数量
	SimilarityThreshold float64 `json:"similarity_threshold"` // 相似度阈值
	Alpha               float32 `json:"alpha"`                // 向量检索权重（混合检索用）
}

// DefaultRAGConfig 默认 RAG 配置
func DefaultRAGConfig() *RAGConfig {
	return &RAGConfig{
		Enabled:             false,
		RetrievalModes:      []string{"vector"}, // 默认仅向量检索
		VectorTopK:          15,
		KeywordTopK:         15,
		GraphTopK:           10,
		SimilarityThreshold: 0.0,
		Alpha:               0.6,
	}
}

// RAGContext RAG 检索结果上下文
type RAGContext struct {
	Query             string                   `json:"query"`               // 原始查询
	FinalQuery        string                   `json:"final_query"`         // 最终使用的查询
	Contexts          []string                 `json:"contexts"`            // 检索到的文档内容
	ContextsWithScore []map[string]interface{} `json:"contexts_with_score"` // 带分数的文档
	SourceTypes       []string                 `json:"source_types"`        // 来源类型列表
	RetrievedCount    int                      `json:"retrieved_count"`     // 检索到的文档数量
	Stages            map[string]interface{}   `json:"stages"`              // 各阶段执行详情
}

// ========================================
// Chat Types
// ========================================

// ChatRequest 聊天请求
type ChatRequest struct {
	SessionID string       `json:"session_id,omitempty"` // 会话ID（可选，用于继续对话）
	KBID      int64        `json:"kb_id,omitempty"`      // 知识库ID（已弃用，使用 RAGConfig.kb_id）
	Content   string       `json:"content"`              // 消息内容
	Stream    bool         `json:"stream"`               // 是否流式响应
	History   []Message    `json:"history,omitempty"`    // 历史消息
	Options   *ChatOptions `json:"options,omitempty"`    // 聊天选项
	RAGConfig *RAGConfig   `json:"rag_config,omitempty"` // RAG 检索配置
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
	MessageID    string      `json:"message_id"`            // 消息ID
	Content      string      `json:"content"`               // 响应内容
	Role         string      `json:"role"`                  // 角色
	TokenCount   int         `json:"token_count"`           // Token数量
	ToolCalls    []ToolCall  `json:"tool_calls"`            // 工具调用
	FinishReason string      `json:"finish_reason"`         // 结束原因
	RAGContext   *RAGContext `json:"rag_context,omitempty"` // RAG 检索上下文
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
	Event      string      `json:"event"`                 // start/content/end/error
	Content    string      `json:"content"`               // 内容片段
	MessageID  string      `json:"message_id"`            // 消息ID
	TokenCount int         `json:"token_count"`           // Token数量
	ToolCalls  []ToolCall  `json:"tool_calls"`            // 工具调用
	Error      string      `json:"error"`                 // 错误信息
	RAGContext *RAGContext `json:"rag_context,omitempty"` // RAG 检索上下文
}
