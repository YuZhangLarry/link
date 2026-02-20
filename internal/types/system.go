package types

import "time"

// ========================================
// 工具模块
// ========================================

// Tool 工具
type Tool struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID    int64     `json:"tenant_id" gorm:"not null;index:idx_tenant_id"`
	Name        string    `json:"name" gorm:"type:varchar(255);not null"`
	Type        string    `json:"type" gorm:"type:varchar(50);not null"` // search/database/http/custom
	Description string    `json:"description" gorm:"type:text"`
	Config      string    `json:"config" gorm:"type:json"` // JSON
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedBy   *int64    `json:"created_by" gorm:"index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Tool) TableName() string {
	return "tools"
}

// ToolExecution 工具执行记录
type ToolExecution struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	MessageID    string    `json:"message_id" gorm:"not null;size:36;index:idx_message_id"`
	ToolID       int64     `json:"tool_id" gorm:"not null;index:idx_tool_id"`
	InputParams  string    `json:"input_params" gorm:"type:json"`           // JSON
	OutputData   string    `json:"output_data" gorm:"type:json"`            // JSON
	Status       string    `json:"status" gorm:"type:varchar(50);not null"` // success/failed/timeout
	DurationMs   int       `json:"duration_ms" gorm:"not null"`
	ErrorMessage string    `json:"error_message" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (ToolExecution) TableName() string {
	return "tool_executions"
}

// ========================================
// 检索模块
// ========================================

// SearchHistory 搜索历史
type SearchHistory struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID      int64     `json:"tenant_id" gorm:"not null;index:idx_tenant_id"`
	UserID        int64     `json:"user_id" gorm:"not null;index:idx_user_id"`
	KBID          *string   `json:"kb_id" gorm:"size:36;index:idx_kb_id"`
	Query         string    `json:"query" gorm:"type:text;not null"`
	RetrievalType string    `json:"retrieval_type" gorm:"type:varchar(50);not null"` // vector/bm25/hybrid/graph
	ResultCount   int       `json:"result_count" gorm:"not null"`
	LatencyMs     int       `json:"latency_ms" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (SearchHistory) TableName() string {
	return "search_history"
}

// ========================================
// 系统模块
// ========================================

// AuditLog 审计日志
type AuditLog struct {
	ID           int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID     *int64    `json:"tenant_id,omitempty" gorm:"index:idx_tenant_id"`
	UserID       *int64    `json:"user_id,omitempty" gorm:"index:idx_user_id"`
	Action       string    `json:"action" gorm:"type:varchar(50);not null"`        // create/update/delete/login
	ResourceType string    `json:"resource_type" gorm:"type:varchar(50);not null"` // tenant/user/kb/document/chat
	ResourceID   string    `json:"resource_id" gorm:"type:varchar(100);not null"`  // VARCHAR(100) to support UUID
	Details      string    `json:"details" gorm:"type:json"`                       // JSON
	IPAddress    string    `json:"ip_address" gorm:"type:varchar(50)"`
	UserAgent    string    `json:"user_agent" gorm:"type:varchar(500)"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}

// SystemConfig 系统配置
type SystemConfig struct {
	ID          int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ConfigKey   string    `json:"config_key" gorm:"type:varchar(100);not null;uniqueIndex"`
	ConfigValue string    `json:"config_value" gorm:"type:text;not null"`
	Description string    `json:"description" gorm:"type:text"`
	IsPublic    bool      `json:"is_public" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (SystemConfig) TableName() string {
	return "system_configs"
}
