package interfaces

import (
	"context"

	"link/internal/types"
)

// TenantRepository 租户仓储接口
type TenantRepository interface {
	// Create 创建租户
	Create(ctx context.Context, tenant *types.Tenant) error

	// FindByID 根据ID查找租户
	FindByID(ctx context.Context, id int64) (*types.Tenant, error)

	// FindByName 根据名称查找租户
	FindByName(ctx context.Context, name string) (*types.Tenant, error)

	// FindByAPIKey 根据API Key查找租户
	FindByAPIKey(ctx context.Context, apiKey string) (*types.Tenant, error)

	// Find 查找租户列表
	Find(ctx context.Context, page, pageSize int) ([]*types.Tenant, int64, error)

	// Update 更新租户
	Update(ctx context.Context, tenant *types.Tenant) error

	// UpdateStorageUsed 更新存储使用量
	UpdateStorageUsed(ctx context.Context, tenantID int64, delta int64) error

	// Delete 删除租户（软删除）
	Delete(ctx context.Context, id int64) error

	// GenerateAPIKey 生成新的API Key
	GenerateAPIKey(ctx context.Context, tenantID int64) (string, error)

	// RotateAPIKey 轮换API Key
	RotateAPIKey(ctx context.Context, tenantID int64) (string, error)
}
