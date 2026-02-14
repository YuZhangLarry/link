package types

import "time"

// SessionEntity 会话实体
type SessionEntity struct {
	ID                string     `json:"id" gorm:"primaryKey;size:36"`
	TenantID          int64      `json:"tenant_id" gorm:"not null;index:idx_tenant_id"`
	UserID            int64      `json:"user_id" gorm:"not null;index:idx_user_id"`
	Title             string     `json:"title" gorm:"size:255"`
	Description       string     `json:"description" gorm:"type:text"`
	KBID              string     `json:"kb_id" gorm:"size:36;index:idx_kb_id"`
	MaxRounds         int        `json:"max_rounds" gorm:"not null;default:5"`
	EnableRewrite     bool       `json:"enable_rewrite" gorm:"not null;default:true"`
	FallbackStrategy  string     `json:"fallback_strategy" gorm:"size:255;not null;default:'fixed'"`
	FallbackResponse  string     `json:"fallback_response" gorm:"size:255;not null;default:'很抱歉，我暂时无法回答这个问题。'"`
	KeywordThreshold  float32    `json:"keyword_threshold" gorm:"type:float;not null;default:0.5"`
	VectorThreshold   float32    `json:"vector_threshold" gorm:"type:float;not null;default:0.5"`
	RerankModelID     string     `json:"rerank_model_id" gorm:"size:64"`
	EmbeddingTopK     int        `json:"embedding_top_k" gorm:"not null;default:10"`
	RerankTopK        int        `json:"rerank_top_k" gorm:"not null;default:10"`
	RerankThreshold   float32    `json:"rerank_threshold" gorm:"type:float;not null;default:0.65"`
	SummaryModelID    string     `json:"summary_model_id" gorm:"size:64"`
	SummaryParameters string     `json:"summary_parameters" gorm:"type:json;not null"`
	AgentConfig       string     `json:"agent_config" gorm:"type:json"`
	ContextConfig     string     `json:"context_config" gorm:"type:json"`
	Status            int8       `json:"status" gorm:"type:tinyint;default:1;index:idx_status"` // 0=归档, 1=正常
	MessageCount      int        `json:"message_count" gorm:"-"`                                // 不映射到数据库，仅用于响应
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName 指定表名
func (SessionEntity) TableName() string {
	return "sessions"
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	Title            string   `json:"title" binding:"required,max=255"`
	Description      string   `json:"description"`
	KBID             *string  `json:"kb_id"`
	MaxRounds        int      `json:"max_rounds" binding:"min=1,max=50"`
	EnableRewrite    bool     `json:"enable_rewrite"`
	KeywordThreshold *float32 `json:"keyword_threshold" binding:"omitempty,min=0,max=1"`
	VectorThreshold  *float32 `json:"vector_threshold" binding:"omitempty,min=0,max=1"`
	RerankModelID    *string  `json:"rerank_model_id"`
	EmbeddingTopK    *int     `json:"embedding_top_k" binding:"omitempty,min=1,max=100"`
	RerankTopK       *int     `json:"rerank_top_k" binding:"omitempty,min=1,max=100"`
	RerankThreshold  *float32 `json:"rerank_threshold" binding:"omitempty,min=0,max=1"`
}

// UpdateSessionRequest 更新会话请求
type UpdateSessionRequest struct {
	Title            *string  `json:"title" binding:"omitempty,max=255"`
	Description      *string  `json:"description"`
	MaxRounds        *int     `json:"max_rounds" binding:"omitempty,min=1,max=50"`
	EnableRewrite    *bool    `json:"enable_rewrite"`
	KeywordThreshold *float32 `json:"keyword_threshold" binding:"omitempty,min=0,max=1"`
	VectorThreshold  *float32 `json:"vector_threshold" binding:"omitempty,min=0,max=1"`
	Status           *int8    `json:"status" binding:"omitempty,oneof=0 1"`
}

// SessionResponse 会话响应
type SessionResponse struct {
	ID            string    `json:"id"`
	TenantID      int64     `json:"tenant_id"`
	UserID        int64     `json:"user_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	KBID          string    `json:"kb_id"`
	MaxRounds     int       `json:"max_rounds"`
	EnableRewrite bool      `json:"enable_rewrite"`
	Status        int8      `json:"status"`
	MessageCount  int       `json:"message_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ListSessionsRequest 查询会话列表请求
type ListSessionsRequest struct {
	Page   int   `form:"page" binding:"min=1"`
	Size   int   `form:"size" binding:"min=1,max=100"`
	Status *int8 `form:"status" binding:"omitempty,oneof=0 1"`
}

// SessionListResponse 会话列表响应
type SessionListResponse struct {
	Sessions []*SessionResponse `json:"sessions"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	Size     int                `json:"size"`
}

// SessionDetailResponse 会话详情响应（包含消息列表）
type SessionDetailResponse struct {
	Session  *SessionResponse   `json:"session"`
	Messages []*MessageResponse `json:"messages"`
}
