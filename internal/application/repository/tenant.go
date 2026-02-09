package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"gorm.io/gorm"

	common_repository "link/internal/common/repository"
	"link/internal/types"
	"link/internal/types/interfaces"
)

// tenantRepository 租户仓储实现
type tenantRepository struct {
	base *common_repository.BaseRepository
}

// NewTenantRepository 创建租户仓储
func NewTenantRepository(db *gorm.DB, tenantEnabled bool) interfaces.TenantRepository {
	return &tenantRepository{
		base: common_repository.NewBaseRepository(db, tenantEnabled),
	}
}

// Create 创建租户
func (r *tenantRepository) Create(ctx context.Context, tenant *types.Tenant) error {
	// 生成 API Key
	apiKey, err := r.generateAPIKey()
	if err != nil {
		return fmt.Errorf("生成 API Key 失败: %w", err)
	}
	tenant.APIKey = apiKey

	return r.base.Create(ctx, tenant)
}

// FindByID 根据ID查找租户
func (r *tenantRepository) FindByID(ctx context.Context, id int64) (*types.Tenant, error) {
	var tenant types.Tenant
	err := r.base.WithContext(ctx).
		Where("id = ?", id).
		First(&tenant).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("租户不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询租户失败: %w", err)
	}

	return &tenant, nil
}

// FindByName 根据名称查找租户
func (r *tenantRepository) FindByName(ctx context.Context, name string) (*types.Tenant, error) {
	var tenant types.Tenant
	err := r.base.WithContext(ctx).
		Where("name = ?", name).
		First(&tenant).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("租户不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询租户失败: %w", err)
	}

	return &tenant, nil
}

// FindByAPIKey 根据API Key查找租户
func (r *tenantRepository) FindByAPIKey(ctx context.Context, apiKey string) (*types.Tenant, error) {
	var tenant types.Tenant
	err := r.base.WithContext(ctx).
		Where("api_key = ?", apiKey).
		First(&tenant).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("无效的 API Key")
	}
	if err != nil {
		return nil, fmt.Errorf("查询租户失败: %w", err)
	}

	return &tenant, nil
}

// Find 查找租户列表
func (r *tenantRepository) Find(ctx context.Context, page, pageSize int) ([]*types.Tenant, int64, error) {
	var tenants []*types.Tenant
	var total int64

	db := r.base.WithContext(ctx)

	// 查询总数
	if err := db.Model(&types.Tenant{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询租户总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&tenants).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询租户列表失败: %w", err)
	}

	return tenants, total, nil
}

// Update 更新租户
func (r *tenantRepository) Update(ctx context.Context, tenant *types.Tenant) error {
	return r.base.Update(ctx, tenant)
}

// UpdateStorageUsed 更新存储使用量
func (r *tenantRepository) UpdateStorageUsed(ctx context.Context, tenantID int64, delta int64) error {
	db := r.base.WithContext(ctx)
	return db.Model(&types.Tenant{}).
		Where("id = ?", tenantID).
		UpdateColumn("storage_used", gorm.Expr("storage_used + ?", delta)).
		Error
}

// Delete 删除租户（软删除）
func (r *tenantRepository) Delete(ctx context.Context, id int64) error {
	db := r.base.WithContext(ctx)
	return db.Delete(&types.Tenant{}, id).Error
}

// GenerateAPIKey 生成新的API Key
func (r *tenantRepository) GenerateAPIKey(ctx context.Context, tenantID int64) (string, error) {
	apiKey, err := r.generateAPIKey()
	if err != nil {
		return "", err
	}

	db := r.base.WithContext(ctx)
	err = db.Model(&types.Tenant{}).
		Where("id = ?", tenantID).
		Update("api_key", apiKey).
		Error

	if err != nil {
		return "", fmt.Errorf("更新 API Key 失败: %w", err)
	}

	return apiKey, nil
}

// RotateAPIKey 轮换API Key
func (r *tenantRepository) RotateAPIKey(ctx context.Context, tenantID int64) (string, error) {
	return r.GenerateAPIKey(ctx, tenantID)
}

// generateAPIKey 生成随机 API Key
func (r *tenantRepository) generateAPIKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "tenant_" + hex.EncodeToString(bytes), nil
}
