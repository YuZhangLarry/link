package models

import (
	"mime/multipart"
	"time"
)

// ========================================
// 租户模块
// ========================================

// Tenant 租户
type Tenant struct {
	ID              int64              `json:"id" db:"id"`
	Name            string             `json:"name" db:"name"`
	Description     string             `json:"description" db:"description"`
	APIKey          string             `json:"api_key" db:"api_key"`
	RetrieverEngines *string            `json:"retriever_engines,omitempty" db:"retriever_engines"` // JSON
	Status          string             `json:"status" db:"status"`                 // active/suspended/deleted
	Business        string             `json:"business" db:"business"`
	StorageQuota    int64              `json:"storage_quota" db:"storage_quota"`
	StorageUsed     int64              `json:"storage_used" db:"storage_used"`
	AgentConfig     *string            `json:"agent_config,omitempty" db:"agent_config"`     // JSON
	Settings        *string            `json:"settings,omitempty" db:"settings"`            // JSON
	CreatedAt       time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time         `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ========================================
// 用户模块
// ========================================

// User 用户模型
type User struct {
	ID           int64      `json:"id" db:"id"`
	TenantID     int64      `json:"tenant_id" db:"tenant_id"` // 租户ID
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Avatar       string     `json:"avatar" db:"avatar"`
	Status       int8       `json:"status" db:"status"` // 0=禁用, 1=正常
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
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
// 模型管理模块
// ========================================

// Model AI模型
type Model struct {
	ID          string    `json:"id" db:"id"`          // VARCHAR(64) UUID
	TenantID    int64     `json:"tenant_id" db:"tenant_id"`
	Name        string    `json:"name" db:"name"`
	Type        string    `json:"type" db:"type"`        // embedding/chat/rerank/vlm/summary
	Source      string    `json:"source" db:"source"`    // openai/azure/dashscope/custom
	Description string    `json:"description" db:"description"`
	Parameters  string    `json:"parameters" db:"parameters"` // JSON
	IsDefault   bool      `json:"is_default" db:"is_default"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ========================================
// 知识库模块
// ========================================

// KnowledgeBase 知识库
type KnowledgeBase struct {
	ID                 string     `json:"id" db:"id"`                   // VARCHAR(36) UUID
	TenantID           int64      `json:"tenant_id" db:"tenant_id"`
	UserID             int64      `json:"user_id" db:"user_id"`
	Name               string     `json:"name" db:"name"`
	Description        string     `json:"description" db:"description"`
	Avatar             string     `json:"avatar" db:"avatar"`
	EmbeddingModelID   string     `json:"embedding_model_id" db:"embedding_model_id"` // VARCHAR(64) UUID
	ChunkingConfig     string     `json:"chunking_config" db:"chunking_config"`     // JSON
	ImageProcessingConfig string    `json:"image_processing_config" db:"image_processing_config"` // JSON
	SummaryModelID     string     `json:"summary_model_id" db:"summary_model_id"`     // VARCHAR(64) UUID
	RerankModelID      string     `json:"rerank_model_id" db:"rerank_model_id"`       // VARCHAR(64) UUID
	CosConfig          string     `json:"cos_config" db:"cos_config"`                // JSON
	VLMConfig          string     `json:"vlm_config" db:"vlm_config"`                // JSON
	ExtractConfig      string     `json:"extract_config" db:"extract_config"`      // JSON
	Status             int8       `json:"status" db:"status"`
	IsPublic           bool       `json:"is_public" db:"is_public"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// KBSetting 知识库设置
type KBSetting struct {
	ID                  int64     `json:"id" db:"id"`
	KBID                string    `json:"kb_id" db:"kb_id"`                     // VARCHAR(36) UUID
	RetrievalMode       string    `json:"retrieval_mode" db:"retrieval_mode"`   // vector/bm25/hybrid/graph
	SimilarityThreshold float64   `json:"similarity_threshold" db:"similarity_threshold"`
	TopK                int       `json:"top_k" db:"top_k"`
	RerankEnabled       bool      `json:"rerank_enabled" db:"rerank_enabled"`
	GraphEnabled        bool      `json:"graph_enabled" db:"graph_enabled"`
	SettingsJSON        string    `json:"settings_json" db:"settings_json"`    // JSON 字符串
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// ========================================
// 知识内容模块
// ========================================

// Knowledge 知识条目
type Knowledge struct {
	ID                 string     `json:"id" db:"id"`                      // VARCHAR(36) UUID
	TenantID           int64      `json:"tenant_id" db:"tenant_id"`
	KBID               string     `json:"kb_id" db:"kb_id"`                  // VARCHAR(36) UUID
	UserID             int64      `json:"user_id" db:"user_id"`
	Type               string     `json:"type" db:"type"`                    // document/file/url
	Title              string     `json:"title" db:"title"`
	Description        string     `json:"description" db:"description"`
	Source             string     `json:"source" db:"source"`                // upload/crawler/api
	ParseStatus        string     `json:"parse_status" db:"parse_status"`    // unprocessed/processing/completed/failed
	EnableStatus       string     `json:"enable_status" db:"enable_status"`    // enabled/disabled
	EmbeddingModelID   string     `json:"embedding_model_id" db:"embedding_model_id"` // VARCHAR(64) UUID
	FileName           string     `json:"file_name" db:"file_name"`
	FileType           string     `json:"file_type" db:"file_type"`
	FileSize           int64      `json:"file_size" db:"file_size"`
	FilePath           string     `json:"file_path" db:"file_path"`
	FileHash           string     `json:"file_hash" db:"file_hash"`
	StorageSize        int64      `json:"storage_size" db:"storage_size"`
	Metadata           string     `json:"metadata" db:"metadata"`              // JSON
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	ProcessedAt        *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	ErrorMessage       string     `json:"error_message" db:"error_message"`
}

// Chunk 文档分块
type Chunk struct {
	ID                  string     `json:"id" db:"id"`                         // VARCHAR(36) UUID
	TenantID            int64      `json:"tenant_id" db:"tenant_id"`
	KBID                string     `json:"kb_id" db:"kb_id"`                   // VARCHAR(36) UUID
	KnowledgeID         string     `json:"knowledge_id" db:"knowledge_id"`     // VARCHAR(36) UUID
	Content             string     `json:"content" db:"content"`
	ChunkIndex          int        `json:"chunk_index" db:"chunk_index"`
	IsEnabled           bool       `json:"is_enabled" db:"is_enabled"`
	StartAt             int        `json:"start_at" db:"start_at"`
	EndAt               int        `json:"end_at" db:"end_at"`
	PreChunkID          string     `json:"pre_chunk_id" db:"pre_chunk_id"`   // VARCHAR(36) UUID
	NextChunkID         string     `json:"next_chunk_id" db:"next_chunk_id"` // VARCHAR(36) UUID
	ChunkType           string     `json:"chunk_type" db:"chunk_type"`       // text/image/table
	ParentChunkID       string     `json:"parent_chunk_id" db:"parent_chunk_id"` // VARCHAR(36) UUID
	ImageInfo           string     `json:"image_info" db:"image_info"`
	RelationChunks      string     `json:"relation_chunks" db:"relation_chunks"`  // JSON
	IndirectRelationChunks string  `json:"indirect_relation_chunks" db:"indirect_relation_chunks"` // JSON
	EmbeddingID         string     `json:"embedding_id" db:"embedding_id"`     // Milvus ID
	TokenCount          int        `json:"token_count" db:"token_count"`
	Metadata            string     `json:"metadata" db:"metadata"`                // JSON
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ========================================
// 会话模块
// ========================================

// Session 会话
type Session struct {
	ID                  string     `json:"id" db:"id"`                       // VARCHAR(36) UUID
	TenantID            int64      `json:"tenant_id" db:"tenant_id"`
	UserID              int64      `json:"user_id" db:"user_id"`
	Title               string     `json:"title" db:"title"`
	Description         string     `json:"description" db:"description"`
	KBID                string     `json:"kb_id" db:"kb_id"`                     // VARCHAR(36) UUID
	MaxRounds           int        `json:"max_rounds" db:"max_rounds"`
	EnableRewrite       bool       `json:"enable_rewrite" db:"enable_rewrite"`
	FallbackStrategy    string     `json:"fallback_strategy" db:"fallback_strategy"`
	FallbackResponse    string     `json:"fallback_response" db:"fallback_response"`
	KeywordThreshold    float32    `json:"keyword_threshold" db:"keyword_threshold"`
	VectorThreshold     float32    `json:"vector_threshold" db:"vector_threshold"`
	RerankModelID       string     `json:"rerank_model_id" db:"rerank_model_id"` // VARCHAR(64) UUID
	EmbeddingTopK       int        `json:"embedding_top_k" db:"embedding_top_k"`
	RerankTopK          int        `json:"rerank_top_k" db:"rerank_top_k"`
	RerankThreshold     float32    `json:"rerank_threshold" db:"rerank_threshold"`
	SummaryModelID      string     `json:"summary_model_id" db:"summary_model_id"` // VARCHAR(64) UUID
	SummaryParameters   string     `json:"summary_parameters" db:"summary_parameters"` // JSON
	AgentConfig         string     `json:"agent_config" db:"agent_config"`         // JSON
	ContextConfig       string     `json:"context_config" db:"context_config"`     // JSON
	Status              int8       `json:"status" db:"status"`                 // 0=归档, 1=正常
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Message 消息
type Message struct {
	ID                  string     `json:"id" db:"id"`                       // VARCHAR(36) UUID
	RequestID           string     `json:"request_id" db:"request_id"`         // VARCHAR(36) UUID
	SessionID           string     `json:"session_id" db:"session_id"`       // VARCHAR(36) UUID
	Role                string     `json:"role" db:"role"`                   // system/user/assistant/tool
	Content             string     `json:"content" db:"content"`
	KnowledgeReferences string     `json:"knowledge_references" db:"knowledge_references"` // JSON
	AgentSteps          string     `json:"agent_steps" db:"agent_steps"`         // JSON
	ToolCalls           string     `json:"tool_calls" db:"tool_calls"`           // JSON
	IsCompleted         bool       `json:"is_completed" db:"is_completed"`
	TokenCount          int        `json:"token_count" db:"token_count"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// MessageFeedback 消息反馈
type MessageFeedback struct {
	ID        int64     `json:"id" db:"id"`
	MessageID string    `json:"message_id" db:"message_id"` // VARCHAR(36) UUID
	UserID    int64     `json:"user_id" db:"user_id"`
	Rating    int       `json:"rating" db:"rating"` // 1-5星
	Comment   string    `json:"comment" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ========================================
// 工具模块
// ========================================

// Tool 工具
type Tool struct {
	ID        int64     `json:"id" db:"id"`
	TenantID  int64     `json:"tenant_id" db:"tenant_id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"` // search/database/http/custom
	Description string   `json:"description" db:"description"`
	Config    string    `json:"config" db:"config"` // JSON
	Enabled   bool      `json:"enabled" db:"enabled"`
	CreatedBy *int64    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ToolExecution 工具执行记录
type ToolExecution struct {
	ID           int64     `json:"id" db:"id"`
	MessageID    string    `json:"message_id" db:"message_id"` // VARCHAR(36) UUID
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
	TenantID      int64     `json:"tenant_id" db:"tenant_id"`
	UserID        int64     `json:"user_id" db:"user_id"`
	KBID          *string   `json:"kb_id" db:"kb_id"`             // VARCHAR(36) UUID
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
	KeyHash    string     `json:"-" db:"key_hash"`
	KeyPrefix  string     `json:"key_prefix" db:"key_prefix"`
	Scopes     string     `json:"scopes" db:"scopes"` // JSON
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Status     int8       `json:"status" db:"status"` // 0=禁用, 1=启用
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// AuditLog 审计日志
type AuditLog struct {
	ID           int64      `json:"id" db:"id"`
	TenantID     *int64     `json:"tenant_id" db:"tenant_id"`
	UserID       *int64     `json:"user_id" db:"user_id"`
	Action       string     `json:"action" db:"action"`               // create/update/delete/login
	ResourceType string     `json:"resource_type" db:"resource_type"` // tenant/user/kb/document/chat
	ResourceID   string     `json:"resource_id" db:"resource_id"`     // VARCHAR(100) to support UUID
	Details      string     `json:"details" db:"details"` // JSON
	IPAddress    string     `json:"ip_address" db:"ip_address"`
	UserAgent    string     `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	ID          int64     `json:"id" db:"id"`
	ConfigKey   string    `json:"config_key" db:"config_key"`
	ConfigValue string    `json:"config_value" db:"config_value"`
	Description string    `json:"description" db:"description"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ========================================
// 请求/响应 DTO (保留向后兼容)
// ========================================

// CreateKnowledgeBaseRequest 创建知识库请求
type CreateKnowledgeBaseRequest struct {
	Name               string `json:"name" binding:"required,min=1,max=100"`
	Description        string `json:"description" binding:"max=500"`
	EmbeddingModelID   string `json:"embedding_model_id"`    // 改为 Model ID
	ChunkSize          int    `json:"chunk_size"`
	ChunkOverlap       int    `json:"chunk_overlap"`
}

// CreateKnowledgeBaseResponse 创建知识库响应
type CreateKnowledgeBaseResponse struct {
	KnowledgeBaseID string `json:"kb_id"`    // 改为 UUID
	Message          string `json:"message"`
}

// UploadDocumentRequest 上传文档请求
type UploadDocumentRequest struct {
	KBID int64                 `form:"kb_id" binding:"required"`
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// UploadDocumentResponse 上传文档响应
type UploadDocumentResponse struct {
	DocumentID string `json:"doc_id"`   // 改为 UUID
	Status     string `json:"status"`
	Message    string `json:"message"`
	Progress   int    `json:"progress"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	SessionID string `json:"session_id"` // 改为 Session UUID
	Content   string `json:"content" binding:"required"`
	Stream    bool   `json:"stream"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	MessageID  string     `json:"message_id"`     // 改为 UUID
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
