package tool

import (
	"time"
)

// ========================================
// Tool 请求/响应类型
// ========================================

// KbQueryRequest 知识库查询请求
type KbQueryRequest struct {
	Query         string  `json:"query" jsonschema:"required,description=用户查询的问题"`
	KBID          int64   `json:"kb_id" jsonschema:"required,description=知识库ID"`
	TopK          int     `json:"top_k" jsonschema:"description=返回结果数量,default=5"`
	Similarity    float64 `json:"similarity" jsonschema:"description=相似度阈值(0-1),default=0.7"`
	RetrievalMode string  `json:"retrieval_mode" jsonschema:"description=检索模式,default=hybrid,enum=vector,enum=bm25,enum=hybrid,enum=graph"`
}

// KbQueryResult 知识库查询结果
type KbQueryResult struct {
	Results []KbChunk `json:"results"`
	Count   int       `json:"count"`
	Query   string    `json:"query"`
	Latency int       `json:"latency_ms"`
}

// KbChunk 知识库分块
type KbChunk struct {
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	Source     string  `json:"source"`
	DocumentID int64   `json:"document_id"`
	ChunkIndex int     `json:"chunk_index"`
}

// WebSearchRequest 网络搜索请求
type WebSearchRequest struct {
	Query string `json:"query" jsonschema:"required,description=搜索关键词"`
	Limit int    `json:"limit" jsonschema:"description=返回结果数量,default=5"`
}

// WebSearchResult 网络搜索结果
type WebSearchResult struct {
	Items []SearchItem `json:"items"`
	Count int          `json:"count"`
	Query string       `json:"query"`
}

// SearchItem 搜索结果项
type SearchItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// KbListRequest 知识库列表请求
type KbListRequest struct {
	UserID int64  `json:"user_id" jsonschema:"description=用户ID(可选,为空则返回所有公开知识库)"`
	Status *int8  `json:"status" jsonschema:"description=状态筛选:0=禁用,1=启用,default=1"`
}

// KbListResult 知识库列表结果
type KbListResult struct {
	KnowledgeBases []KbInfo `json:"knowledge_bases"`
	Count          int      `json:"count"`
}

// KbInfo 知识库信息
type KbInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DocumentCount int  `json:"document_count"`
}

// DocumentListRequest 文档列表请求
type DocumentListRequest struct {
	KBID  int64  `json:"kb_id" jsonschema:"required,description=知识库ID"`
	Limit int    `json:"limit" jsonschema:"description=返回结果数量,default=10"`
}

// DocumentListResult 文档列表结果
type DocumentListResult struct {
	Documents []DocInfo `json:"documents"`
	Count     int       `json:"count"`
}

// DocInfo 文档信息
type DocInfo struct {
	ID         int64    `json:"id"`
	FileName   string   `json:"file_name"`
	FileType   string   `json:"file_type"`
	Status     string   `json:"status"`
	ChunkCount int      `json:"chunk_count"`
	CreatedAt  string   `json:"created_at"`
}

// ========================================
// Tool 执行记录
// ========================================

// ToolExecutionRecord 工具执行记录
type ToolExecutionRecord struct {
	ID           int64     `json:"id"`
	MessageID    int64     `json:"message_id"`
	ToolName     string    `json:"tool_name"`
	InputParams  string    `json:"input_params"`
	OutputData   string    `json:"output_data"`
	Status       string    `json:"status"`
	DurationMs   int       `json:"duration_ms"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// ToolExecOption 工具执行选项
type ToolExecOption struct {
	Timeout    time.Duration // 超时时间
	MaxRetries int           // 最大重试次数
}

// ToolExecResult 工具执行结果
type ToolExecResult struct {
	Success     bool          `json:"success"`
	Data        string        `json:"data"`
	Error       error         `json:"error,omitempty"`
	DurationMs  int           `json:"duration_ms"`
	ToolName    string        `json:"tool_name"`
}

// ========================================
// Agent 配置
// ========================================

// AgentConfig Agent 配置
type AgentConfig struct {
	MaxToolIterations int           // 最大工具迭代次数
	ToolTimeout       time.Duration // 工具调用超时时间
	EnableTools       bool          // 是否启用工具
	ToolNames         []string      // 启用的工具名称列表
}

// DefaultAgentConfig 默认 Agent 配置
func DefaultAgentConfig() *AgentConfig {
	return &AgentConfig{
		MaxToolIterations: 5,
		ToolTimeout:       30 * time.Second,
		EnableTools:       true,
		ToolNames:         []string{},
	}
}
