package common_repository

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ========================================
// GORM 租户 Scope
// ========================================

// TenantScope 租户过滤 Scope
// 使用方式: db.Scopes(TenantScope(tenantID)).Find(&results)
func TenantScope(tenantID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if tenantID == 0 {
			return db
		}
		return db.Where("tenant_id = ?", tenantID)
	}
}

// SoftDeleteScope 软删除过滤 Scope
// 使用方式: db.Scopes(SoftDeleteScope()).Find(&results)
func SoftDeleteScope() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("deleted_at IS NULL")
	}
}

// TenantWithSoftDeleteScope 租户过滤 + 软删除过滤 Scope
// 使用方式: db.Scopes(TenantWithSoftDeleteScope(tenantID)).Find(&results)
func TenantWithSoftDeleteScope(tenantID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Where("deleted_at IS NULL")
		if tenantID > 0 {
			db = db.Where("tenant_id = ?", tenantID)
		}
		return db
	}
}

// ========================================
// 基础仓储 - 使用 GORM
// ========================================

// BaseRepository 基础仓储，提供通用的租户过滤和 GORM 操作
type BaseRepository struct {
	db            *gorm.DB
	tenantEnabled bool
}

// NewBaseRepository 创建基础仓储
func NewBaseRepository(db *gorm.DB, tenantEnabled bool) *BaseRepository {
	return &BaseRepository{
		db:            db,
		tenantEnabled: tenantEnabled,
	}
}

// GetDB 获取数据库连接
func (r *BaseRepository) GetDB() *gorm.DB {
	return r.db
}

// WithContext 返回带上下文的 DB
func (r *BaseRepository) WithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

// WithTenantScope 返回带租户过滤的 DB
func (r *BaseRepository) WithTenantScope(ctx context.Context, tenantID int64) *gorm.DB {
	db := r.db.WithContext(ctx)
	if r.tenantEnabled && tenantID > 0 {
		db = db.Scopes(TenantScope(tenantID))
	}
	return db
}

// WithTenantAndSoftDeleteScope 返回带租户过滤和软删除过滤的 DB
func (r *BaseRepository) WithTenantAndSoftDeleteScope(ctx context.Context, tenantID int64) *gorm.DB {
	db := r.db.WithContext(ctx)
	if r.tenantEnabled {
		db = db.Scopes(TenantWithSoftDeleteScope(tenantID))
	} else {
		db = db.Scopes(SoftDeleteScope())
	}
	return db
}

// GetTenantID 从上下文获取租户ID
func (r *BaseRepository) GetTenantID(ctx context.Context) int64 {
	if tenantID, ok := ctx.Value("tenant_id").(int64); ok {
		return tenantID
	}
	return 0
}

// MustGetTenantID 必须获取租户ID，否则panic
func (r *BaseRepository) MustGetTenantID(ctx context.Context) int64 {
	tenantID := r.GetTenantID(ctx)
	if tenantID == 0 && r.tenantEnabled {
		panic("tenant_id is required but not set in context")
	}
	return tenantID
}

// ========================================
// 通用 CRUD 操作
// ========================================

// Create 创建记录
func (r *BaseRepository) Create(ctx context.Context, entity interface{}) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// FindByID 根据ID查找记录
func (r *BaseRepository) FindByID(ctx context.Context, dest interface{}, id string, tenantID int64) error {
	db := r.WithTenantAndSoftDeleteScope(ctx, tenantID)
	return db.First(dest, "id = ?", id).Error
}

// Find 查找多条记录
func (r *BaseRepository) Find(ctx context.Context, dest interface{}, tenantID int64, conds ...interface{}) error {
	db := r.WithTenantAndSoftDeleteScope(ctx, tenantID)
	if len(conds) > 0 {
		db = db.Where(conds[0], conds[1:]...)
	}
	return db.Find(dest).Error
}

// Update 更新记录
func (r *BaseRepository) Update(ctx context.Context, entity interface{}) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// UpdateFields 更新指定字段
func (r *BaseRepository) UpdateFields(ctx context.Context, table string, tenantID int64, id string, fields map[string]interface{}) error {
	db := r.WithTenantScope(ctx, tenantID)
	return db.Table(table).Where("id = ?", id).Updates(fields).Error
}

// Delete 软删除记录
func (r *BaseRepository) Delete(ctx context.Context, table string, tenantID int64, id string) error {
	db := r.WithTenantScope(ctx, tenantID)
	return db.Table(table).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// HardDelete 硬删除记录
func (r *BaseRepository) HardDelete(ctx context.Context, table string, tenantID int64, id string) error {
	db := r.WithTenantScope(ctx, tenantID)
	return db.Table(table).Where("id = ?", id).Delete(nil).Error
}

// Count 统计记录数
func (r *BaseRepository) Count(ctx context.Context, table string, tenantID int64, conds ...interface{}) (int64, error) {
	db := r.WithTenantAndSoftDeleteScope(ctx, tenantID)
	var count int64
	query := db.Table(table)
	if len(conds) > 0 {
		query = query.Where(conds[0], conds[1:]...)
	}
	err := query.Count(&count).Error
	return count, err
}

// ========================================
// 分页查询
// ========================================

// Pagination 分页参数
type Pagination struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	OrderBy  string `json:"order_by"`
}

// DefaultPagination 默认分页参数
func DefaultPagination() *Pagination {
	return &Pagination{
		Page:     1,
		PageSize: 20,
		OrderBy:  "created_at DESC",
	}
}

// Paginate 分页查询
func (r *BaseRepository) Paginate(ctx context.Context, dest interface{}, tenantID int64, pagination *Pagination, conds ...interface{}) (int64, error) {
	if pagination == nil {
		pagination = DefaultPagination()
	}

	// 查询总数
	db := r.WithTenantAndSoftDeleteScope(ctx, tenantID)
	var total int64
	countQuery := db.Model(dest)
	if len(conds) > 0 {
		countQuery = countQuery.Where(conds[0], conds[1:]...)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return 0, err
	}

	// 分页查询
	offset := (pagination.Page - 1) * pagination.PageSize
	query := db.Model(dest)
	if len(conds) > 0 {
		query = query.Where(conds[0], conds[1:]...)
	}

	if pagination.OrderBy != "" {
		query = query.Order(pagination.OrderBy)
	}

	if err := query.Offset(offset).Limit(pagination.PageSize).Find(dest).Error; err != nil {
		return 0, err
	}

	return total, nil
}
