package types

import "time"

// ========================================
// Session Entity
// ========================================

// SessionEntity 会话实体（数据库模型）
type SessionEntity struct {
	ID              int64      `json:"id" db:"id"`
	UserID          int64      `json:"user_id" db:"user_id"`
	KBID            *int64     `json:"kb_id,omitempty" db:"kb_id"`
	Title           string     `json:"title" db:"title"`
	Description     string     `json:"description,omitempty" db:"description"`
	Model           string     `json:"model" db:"model"`
	MaxRounds       int        `json:"max_rounds" db:"max_rounds"`
	EnableRewrite   bool       `json:"enable_rewrite" db:"enable_rewrite"`
	KeywordThreshold float64   `json:"keyword_threshold" db:"keyword_threshold"`
	VectorThreshold float64    `json:"vector_threshold" db:"vector_threshold"`
	MessageCount    int        `json:"message_count" db:"message_count"`
	Status          int8       `json:"status" db:"status"` // 0=归档, 1=正常
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// ========================================
// Request DTO
// ========================================

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	KBID            *int64   `json:"kb_id,omitempty"`
	Title           string   `json:"title" binding:"required,min=1,max=255"`
	Description     string   `json:"description" binding:"max=1000"`
	Model           string   `json:"model" binding:"omitempty,max=50"`
	MaxRounds       int      `json:"max_rounds" binding:"omitempty,min=1,max=100"`
	EnableRewrite   bool     `json:"enable_rewrite"`
	KeywordThreshold float64 `json:"keyword_threshold" binding:"omitempty,min=0,max=1"`
	VectorThreshold float64  `json:"vector_threshold" binding:"omitempty,min=0,max=1"`
}

// UpdateSessionRequest 更新会话请求
type UpdateSessionRequest struct {
	Title           string   `json:"title" binding:"required,min=1,max=255"`
	Description     string   `json:"description" binding:"max=1000"`
	Model           string   `json:"model" binding:"omitempty,max=50"`
	MaxRounds       int      `json:"max_rounds" binding:"omitempty,min=1,max=100"`
	EnableRewrite   bool     `json:"enable_rewrite"`
	KeywordThreshold float64 `json:"keyword_threshold" binding:"omitempty,min=0,max=1"`
	VectorThreshold float64  `json:"vector_threshold" binding:"omitempty,min=0,max=1"`
	Status          *int8    `json:"status" binding:"omitempty,oneof=0 1"`
}

// ListSessionsRequest 查询会话列表请求
type ListSessionsRequest struct {
	UserID int64  `form:"user_id" binding:"required"`
	Page   int    `form:"page" binding:"min=1"`
	Size   int    `form:"size" binding:"min=1,max=100"`
	Status *int8  `form:"status" binding:"omitempty,oneof=0 1"`
}

// ========================================
// Response DTO
// ========================================

// SessionResponse 会话响应
type SessionResponse struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	KBID            *int64     `json:"kb_id,omitempty"`
	Title           string     `json:"title"`
	Description     string     `json:"description,omitempty"`
	Model           string     `json:"model"`
	MaxRounds       int        `json:"max_rounds"`
	EnableRewrite   bool       `json:"enable_rewrite"`
	KeywordThreshold float64   `json:"keyword_threshold"`
	VectorThreshold float64    `json:"vector_threshold"`
	MessageCount    int        `json:"message_count"`
	Status          int8       `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// SessionDetailResponse 会话详情响应（包含消息列表）
type SessionDetailResponse struct {
	SessionResponse
	Messages []MessageResponse `json:"messages,omitempty"`
}

// SessionListResponse 会话列表响应
type SessionListResponse struct {
	Sessions []SessionResponse `json:"sessions"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Size     int               `json:"size"`
}
