package types

import "time"

// Model AI模型
type Model struct {
	ID          string     `json:"id" gorm:"primaryKey;size:64"` // VARCHAR(64) UUID
	TenantID    int64      `json:"tenant_id" gorm:"not null;index:idx_tenant_id"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	Type        string     `json:"type" gorm:"type:varchar(50);not null;index:idx_type"` // embedding/chat/rerank/vlm/summary
	Source      string     `json:"source" gorm:"type:varchar(50);not null"`              // openai/azure/dashscope/custom
	Description string     `json:"description" gorm:"type:text"`
	Parameters  string     `json:"parameters" gorm:"type:json"` // JSON
	IsDefault   bool       `json:"is_default" gorm:"default:false"`
	Status      string     `json:"status" gorm:"type:varchar(50);default:'active'"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName 指定表名
func (Model) TableName() string {
	return "models"
}

// CreateModelRequest 创建模型请求
type CreateModelRequest struct {
	Name        string `json:"name" binding:"required,max=255"`
	Type        string `json:"type" binding:"required,oneof=embedding chat rerank vlm summary"`
	Source      string `json:"source" binding:"required,max=50"`
	Description string `json:"description" binding:"max=500"`
	Parameters  string `json:"parameters"` // JSON
	IsDefault   bool   `json:"is_default"`
}

// UpdateModelRequest 更新模型请求
type UpdateModelRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=255"`
	Description *string `json:"description" binding:"omitempty,max=500"`
	Parameters  *string `json:"parameters"`
	IsDefault   *bool   `json:"is_default"`
	Status      *string `json:"status" binding:"omitempty,oneof=active inactive"`
}

// ModelResponse 模型响应
type ModelResponse struct {
	ID          string    `json:"id"`
	TenantID    int64     `json:"tenant_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Source      string    `json:"source"`
	Description string    `json:"description"`
	Parameters  string    `json:"parameters"`
	IsDefault   bool      `json:"is_default"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
