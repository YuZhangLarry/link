package types

import "time"

// SessionEntity 会话实体（对应数据库表结构）
type SessionEntity struct {
	ID           string     `json:"id" gorm:"primaryKey;size:36"`
	TenantID     int64      `json:"tenant_id" gorm:"not null;index:idx_tenant_id"`
	UserID       int64      `json:"user_id" gorm:"not null;index:idx_user_id"`
	Title        string     `json:"title" gorm:"size:255"`
	Description  string     `json:"description" gorm:"type:text"`
	Status       int8       `json:"status" gorm:"type:tinyint;default:1;index:idx_status"` // 0=归档, 1=正常
	MessageCount int        `json:"message_count" gorm:"-"`                                // 不映射到数据库，仅用于响应
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName 指定表名
func (SessionEntity) TableName() string {
	return "sessions"
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	Title       string     `json:"title" binding:"max=255"`
	Description string     `json:"description"`
	RAGConfig   *RAGConfig `json:"rag_config,omitempty"` // RAG 配置（可选）
}

// UpdateSessionRequest 更新会话请求
type UpdateSessionRequest struct {
	Title       *string    `json:"title" binding:"omitempty,max=255"`
	Description *string    `json:"description"`
	Status      *int8      `json:"status" binding:"omitempty,oneof=0 1"`
	RAGConfig   *RAGConfig `json:"rag_config,omitempty"` // RAG 配置更新（可选）
}

// SessionResponse 会话响应（聚合 RAG 配置）
type SessionResponse struct {
	ID           string     `json:"id"`
	TenantID     int64      `json:"tenant_id"`
	UserID       int64      `json:"user_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Status       int8       `json:"status"`
	MessageCount int        `json:"message_count"`
	RAGConfig    *RAGConfig `json:"rag_config,omitempty"` // RAG 配置（从 retrieval_settings 表聚合）
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
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
