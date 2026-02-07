package types

import "time"

// ========================================
// Message Entity
// ========================================

// MessageEntity 消息实体（数据库模型）
type MessageEntity struct {
	ID         int64     `json:"id" db:"id"`
	ChatID     int64     `json:"chat_id" db:"chat_id"`
	Role       string    `json:"role" db:"role"` // system/user/assistant/tool
	Content    string    `json:"content" db:"content"`
	ToolCalls  string    `json:"tool_calls,omitempty" db:"tool_calls"` // JSON string
	TokenCount int       `json:"token_count" db:"token_count"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// ========================================
// Request DTO
// ========================================

// CreateMessageRequest 创建消息请求
type CreateMessageRequest struct {
	ChatID     int64              `json:"chat_id" binding:"required"`
	Role       string             `json:"role" binding:"required,oneof=system user assistant tool"`
	Content    string             `json:"content" binding:"required,max=10000"`
	ToolCalls  string             `json:"tool_calls,omitempty"`
	TokenCount int                `json:"token_count"`
}

// UpdateMessageRequest 更新消息请求
type UpdateMessageRequest struct {
	Content    string `json:"content" binding:"required,max=10000"`
	TokenCount int    `json:"token_count"`
}

// ListMessagesRequest 查询消息列表请求
type ListMessagesRequest struct {
	ChatID int64  `form:"chat_id" binding:"required"`
	Page   int    `form:"page" binding:"min=1"`
	Size   int    `form:"size" binding:"min=1,max=100"`
	Role   string `form:"role" binding:"omitempty,oneof=system user assistant tool"`
}

// ========================================
// Response DTO
// ========================================

// MessageResponse 消息响应
type MessageResponse struct {
	ID         int64     `json:"id"`
	ChatID     int64     `json:"chat_id"`
	Role       string    `json:"role"`
	Content    string    `json:"content"`
	ToolCalls  string    `json:"tool_calls,omitempty"`
	TokenCount int       `json:"token_count"`
	CreatedAt  time.Time `json:"created_at"`
}

// MessageListResponse 消息列表响应
type MessageListResponse struct {
	Messages []MessageResponse `json:"messages"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Size     int               `json:"size"`
}
