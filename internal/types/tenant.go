package types

import "time"

// Tenant 租户实体
type Tenant struct {
	ID               int64              `json:"id" gorm:"primaryKey;autoIncrement"`
	Name             string             `json:"name" gorm:"type:varchar(255);not null"`
	Description      string             `json:"description" gorm:"type:text"`
	APIKey           string             `json:"api_key,omitempty" gorm:"type:varchar(64);not null;uniqueIndex"`
	RetrieverEngines *string            `json:"retriever_engines,omitempty" gorm:"type:json"` // {"vector": "milvus", "graph": "neo4j", "bm25": "redis"} - 租户后续配置
	Status           string             `json:"status" gorm:"type:varchar(50);default:'active'"` // active/suspended/deleted
	Business         string             `json:"business" gorm:"type:varchar(255);not null"`
	StorageQuota     int64              `json:"storage_quota" gorm:"not null;default:10737418240"` // 10GB in bytes
	StorageUsed      int64              `json:"storage_used" gorm:"not null;default:0"`
	AgentConfig      *string            `json:"agent_config,omitempty" gorm:"type:json"` // 租户级Agent配置
	Settings         *string            `json:"settings,omitempty" gorm:"type:json"` // {"embedding_model", "rerank_model", "summary_model"}
	CreatedAt        time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time          `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        *time.Time         `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName 指定表名
func (Tenant) TableName() string {
	return "tenants"
}

// TenantUser 租户用户关联
type TenantUser struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID  int64     `json:"tenant_id" gorm:"not null;index:idx_tenant_user;uniqueIndex:uk_tenant_user"`
	UserID    int64     `json:"user_id" gorm:"not null;index:idx_user_id;uniqueIndex:uk_tenant_user"`
	Role      string    `json:"role" gorm:"type:varchar(50);not null;default:'member';index:idx_role"` // owner/admin/member
	Status    string    `json:"status" gorm:"type:varchar(50);default:'active'"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName 指定表名
func (TenantUser) TableName() string {
	return "tenant_users"
}

// CreateTenantRequest 创建租户请求
type CreateTenantRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=255"`
	Description  string `json:"description" binding:"max=500"`
	Business     string `json:"business" binding:"required,max=255"`
	StorageQuota int64  `json:"storage_quota" binding:"min=0"`
}

// UpdateTenantRequest 更新租户请求
type UpdateTenantRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=255"`
	Description string `json:"description" binding:"omitempty,max=500"`
	Business    string `json:"business" binding:"omitempty,max=255"`
	Status      string `json:"status" binding:"omitempty,oneof=active suspended"`
}

// TenantResponse 租户响应
type TenantResponse struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Business         string    `json:"business"`
	Status           string    `json:"status"`
	StorageQuota     int64     `json:"storage_quota"`
	StorageUsed      int64     `json:"storage_used"`
	RetrieverEngines *string   `json:"retriever_engines,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ToResponse 转换为响应格式
func (t *Tenant) ToResponse() *TenantResponse {
	return &TenantResponse{
		ID:               t.ID,
		Name:             t.Name,
		Description:      t.Description,
		Business:         t.Business,
		Status:           t.Status,
		StorageQuota:     t.StorageQuota,
		StorageUsed:      t.StorageUsed,
		RetrieverEngines: t.RetrieverEngines,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}
