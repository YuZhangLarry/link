package types

import "time"

// MessageEntity 消息实体
type MessageEntity struct {
	ID                  string     `json:"id" gorm:"primaryKey;size:36"`
	RequestID           string     `json:"request_id" gorm:"not null;size:36;index:idx_request_id"`
	SessionID           string     `json:"session_id" gorm:"not null;size:36;index:idx_session_id"`
	Role                string     `json:"role" gorm:"type:varchar(50);not null;index:idx_role"` // system/user/assistant/tool
	Content             string     `json:"content" gorm:"type:text;not null"`
	KnowledgeReferences string     `json:"knowledge_references" gorm:"type:json"`
	AgentSteps          string     `json:"agent_steps" gorm:"type:json"`
	ToolCalls           string     `json:"tool_calls" gorm:"type:json"`
	IsCompleted         bool       `json:"is_completed" gorm:"default:false"`
	TokenCount          int        `json:"token_count"`
	CreatedAt           time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt           *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName 指定表名
func (MessageEntity) TableName() string {
	return "messages"
}

// CreateMessageRequest 创建消息请求
type CreateMessageRequest struct {
	SessionID           string `json:"session_id" binding:"required"`
	Role                string `json:"role" binding:"required,oneof=system user assistant tool"`
	Content             string `json:"content" binding:"required"`
	KnowledgeReferences string `json:"knowledge_references"`
	AgentSteps          string `json:"agent_steps"`
	ToolCalls           string `json:"tool_calls"`
	TokenCount          int    `json:"token_count"`
}

// UpdateMessageRequest 更新消息请求
type UpdateMessageRequest struct {
	Content             *string `json:"content"`
	IsCompleted         *bool   `json:"is_completed"`
	TokenCount          *int    `json:"token_count"`
	KnowledgeReferences *string `json:"knowledge_references"`
	AgentSteps          *string `json:"agent_steps"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID                  string    `json:"id"`
	RequestID           string    `json:"request_id"`
	SessionID           string    `json:"session_id"`
	Role                string    `json:"role"`
	Content             string    `json:"content"`
	KnowledgeReferences string    `json:"knowledge_references"`
	AgentSteps          string    `json:"agent_steps"`
	ToolCalls           string    `json:"tool_calls"`
	IsCompleted         bool      `json:"is_completed"`
	TokenCount          int       `json:"token_count"`
	CreatedAt           time.Time `json:"created_at"`
}

// ListMessagesRequest 查询消息列表请求
type ListMessagesRequest struct {
	SessionID string `form:"session_id" binding:"required"`
	Page      int    `form:"page" binding:"min=1"`
	Size      int    `form:"size" binding:"min=1,max=100"`
	Role      string `form:"role" binding:"omitempty,oneof=system user assistant tool"`
}

// MessageListResponse 消息列表响应
type MessageListResponse struct {
	Messages []*MessageResponse `json:"messages"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	Size     int                `json:"size"`
}

// MessageFeedback 消息反馈
type MessageFeedback struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	MessageID string    `json:"message_id" gorm:"not null;size:36;index:idx_message_id"`
	UserID    int64     `json:"user_id" gorm:"not null;index:idx_user_id"`
	Rating    int       `json:"rating" gorm:"not null"` // 1-5星
	Comment   string    `json:"comment" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (MessageFeedback) TableName() string {
	return "message_feedbacks"
}
