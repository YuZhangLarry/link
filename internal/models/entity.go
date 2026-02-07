package models

import (
	"mime/multipart"
	"time"
)

// ========================================
// 用户模块
// ========================================

// User 用户模型
type User struct {
	ID           int64      `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Avatar       string     `json:"avatar" db:"avatar"`
	Status       int8       `json:"status" db:"status"` // 0=禁用, 1=正常
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}

// UserPreference 用户偏好设置
type UserPreference struct {
	ID                  int64     `json:"id" db:"id"`
	UserID              int64     `json:"user_id" db:"user_id"`
	Language            string    `json:"language" db:"language"`
	Theme               string    `json:"theme" db:"theme"`
	NotificationEnabled bool      `json:"notification_enabled" db:"notification_enabled"`
	PreferenceJSON      string    `json:"preference_json" db:"preference_json"` // JSON 字符串
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// ========================================
// 知识库模块
// ========================================

// KnowledgeBase 知识库
type KnowledgeBase struct {
	ID             int64     `json:"id" db:"id"`
	UserID         int64     `json:"user_id" db:"user_id"`
	Name           string    `json:"name" db:"name"`
	Description    string    `json:"description" db:"description"`
	Avatar         string    `json:"avatar" db:"avatar"`
	EmbeddingModel string    `json:"embedding_model" db:"embedding_model"`
	ChunkSize      int       `json:"chunk_size" db:"chunk_size"`
	ChunkOverlap   int       `json:"chunk_overlap" db:"chunk_overlap"`
	Status         int8      `json:"status" db:"status"`
	IsPublic       bool      `json:"is_public" db:"is_public"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// KBSetting 知识库设置
type KBSetting struct {
	ID                  int64     `json:"id" db:"id"`
	KBID                int64     `json:"kb_id" db:"kb_id"`
	RetrievalMode       string    `json:"retrieval_mode" db:"retrieval_mode"` // vector/bm25/hybrid/graph
	SimilarityThreshold float64   `json:"similarity_threshold" db:"similarity_threshold"`
	TopK                int       `json:"top_k" db:"top_k"`
	RerankEnabled       bool      `json:"rerank_enabled" db:"rerank_enabled"`
	GraphEnabled        bool      `json:"graph_enabled" db:"graph_enabled"`
	SettingsJSON        string    `json:"settings_json" db:"settings_json"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// ========================================
// 文档模块
// ========================================

// Document 文档
type Document struct {
	ID           int64      `json:"id" db:"id"`
	KBID         int64      `json:"kb_id" db:"kb_id"`
	UserID       int64      `json:"user_id" db:"user_id"`
	FileName     string     `json:"file_name" db:"file_name"`
	FileType     string     `json:"file_type" db:"file_type"` // pdf/docx/txt/md
	FileSize     int64      `json:"file_size" db:"file_size"`
	FilePath     string     `json:"file_path" db:"file_path"`
	FileHash     string     `json:"file_hash" db:"file_hash"`
	Status       string     `json:"status" db:"status"` // pending/processing/completed/failed
	ErrorMessage string     `json:"error_message" db:"error_message"`
	ChunkCount   int        `json:"chunk_count" db:"chunk_count"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// Chunk 文档分块
type Chunk struct {
	ID          int64     `json:"id" db:"id"`
	DocumentID  int64     `json:"document_id" db:"document_id"`
	ChunkIndex  int       `json:"chunk_index" db:"chunk_index"`
	Content     string    `json:"content" db:"content"`
	TokenCount  int       `json:"token_count" db:"token_count"`
	EmbeddingID string    `json:"embedding_id" db:"embedding_id"` // Milvus ID
	Metadata    string    `json:"metadata" db:"metadata"`         // JSON
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ========================================
// 对话模块
// ========================================

// Chat 对话
type Chat struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	KBID         *int64    `json:"kb_id" db:"kb_id"`
	Title        string    `json:"title" db:"title"`
	MessageCount int       `json:"message_count" db:"message_count"`
	Model        string    `json:"model" db:"model"`
	Status       int8      `json:"status" db:"status"` // 0=归档, 1=正常
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Message 消息
type Message struct {
	ID         int64     `json:"id" db:"id"`
	ChatID     int64     `json:"chat_id" db:"chat_id"`
	Role       string    `json:"role" db:"role"` // system/user/assistant/tool
	Content    string    `json:"content" db:"content"`
	ToolCalls  string    `json:"tool_calls" db:"tool_calls"` // JSON
	TokenCount int       `json:"token_count" db:"token_count"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// MessageFeedback 消息反馈
type MessageFeedback struct {
	ID        int64     `json:"id" db:"id"`
	MessageID int64     `json:"message_id" db:"message_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Rating    int       `json:"rating" db:"rating"` // 1-5星
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ========================================
// Agent & 工具模块
// ========================================

// Tool 工具
type Tool struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Type        string    `json:"type" db:"type"` // search/database/http/custom
	Description string    `json:"description" db:"description"`
	Config      string    `json:"config" db:"config"` // JSON
	Enabled     bool      `json:"enabled" db:"enabled"`
	CreatedBy   int64     `json:"created_by" db:"created_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ToolExecution 工具执行记录
type ToolExecution struct {
	ID           int64     `json:"id" db:"id"`
	MessageID    int64     `json:"message_id" db:"message_id"`
	ToolID       int64     `json:"tool_id" db:"tool_id"`
	InputParams  string    `json:"input_params" db:"input_params"` // JSON
	OutputData   string    `json:"output_data" db:"output_data"`   // JSON
	Status       string    `json:"status" db:"status"`             // success/failed/timeout
	DurationMs   int       `json:"duration_ms" db:"duration_ms"`
	ErrorMessage string    `json:"error_message" db:"error_message"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ========================================
// 检索模块
// ========================================

// SearchHistory 搜索历史
type SearchHistory struct {
	ID            int64     `json:"id" db:"id"`
	UserID        int64     `json:"user_id" db:"user_id"`
	KBID          *int64    `json:"kb_id" db:"kb_id"`
	Query         string    `json:"query" db:"query"`
	RetrievalType string    `json:"retrieval_type" db:"retrieval_type"`
	ResultCount   int       `json:"result_count" db:"result_count"`
	LatencyMs     int       `json:"latency_ms" db:"latency_ms"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// ========================================
// 系统模块
// ========================================

// APIKey API密钥
type APIKey struct {
	ID         int64      `json:"id" db:"id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	Name       string     `json:"name" db:"name"`
	KeyHash    string     `json:"-" db:"key_hash`
	KeyPrefix  string     `json:"key_prefix" db:"key_prefix"`
	Scopes     string     `json:"scopes" db:"scopes"` // JSON
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Status     int8       `json:"status" db:"status"` // 0=禁用, 1=启用
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID           int64     `json:"id" db:"id"`
	UserID       *int64    `json:"user_id" db:"user_id"`
	Action       string    `json:"action" db:"action"`               // create/update/delete/login
	ResourceType string    `json:"resource_type" db:"resource_type"` // user/kb/document/chat
	ResourceID   *int64    `json:"resource_id" db:"resource_id"`
	Details      string    `json:"details" db:"details"` // JSON
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ========================================
// 请求/响应 DTO
// ========================================

// CreateKnowledgeBaseRequest 创建知识库请求
type CreateKnowledgeBaseRequest struct {
	Name           string `json:"name" binding:"required,min=1,max=100"`
	Description    string `json:"description" binding:"max=500"`
	EmbeddingModel string `json:"embedding_model"`
	ChunkSize      int    `json:"chunk_size"`
	ChunkOverlap   int    `json:"chunk_overlap"`
}

// CreateKnowledgeBaseResponse 创建知识库响应
type CreateKnowledgeBaseResponse struct {
	KnowledgeBaseID int64  `json:"kb_id"`
	Message         string `json:"message"`
}

// UploadDocumentRequest 上传文档请求
type UploadDocumentRequest struct {
	KBID int64                 `form:"kb_id" binding:"required"`
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// UploadDocumentResponse 上传文档响应
type UploadDocumentResponse struct {
	DocumentID int64  `json:"doc_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	Progress   int    `json:"progress"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	KBID    int64  `json:"kb_id"`
	Content string `json:"content" binding:"required"`
	Stream  bool   `json:"stream"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	ID        string `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	MessageID  string     `json:"message_id"`
	Content    string     `json:"content"`
	Role       string     `json:"role"`
	CreatedAt  int64      `json:"created_at"`
	TokenCount int        `json:"token_count"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
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
	Event     string `json:"event"` // start/content/end/error
	Content   string `json:"content"`
	MessageID string `json:"message_id"`
	Error     string `json:"error,omitempty"`
}
