package repository

import (
	"context"
	"fmt"
	common_repository "link/internal/common"

	"gorm.io/gorm"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// ========================================
// 知识标签仓储实现
// ========================================

// tagRepository 知识标签仓储实现
type tagRepository struct {
	base *common_repository.BaseRepository
}

// NewTagRepository 创建知识标签仓储
func NewTagRepository(db *gorm.DB, tenantEnabled bool) interfaces.TagRepository {
	return &tagRepository{
		base: common_repository.NewBaseRepository(db, tenantEnabled),
	}
}

// Create 创建标签
func (r *tagRepository) Create(ctx context.Context, tag *types.Tag) error {
	return r.base.Create(ctx, tag)
}

// CreateBatch 批量创建标签
func (r *tagRepository) CreateBatch(ctx context.Context, tags []*types.Tag) error {
	if len(tags) == 0 {
		return nil
	}
	return r.base.WithContext(ctx).Create(&tags).Error
}

// FindByID 根据ID查找标签
func (r *tagRepository) FindByID(ctx context.Context, id int64) (*types.Tag, error) {
	var tag types.Tag
	err := r.base.WithContext(ctx).
		Where("id = ?", id).
		First(&tag).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("标签不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}

	return &tag, nil
}

// FindByKnowledgeBaseID 根据知识库ID查找标签列表
func (r *tagRepository) FindByKnowledgeBaseID(
	ctx context.Context,
	tenantID string,
	kbID int64,
	query *types.TagListQuery,
) ([]*types.Tag, int64, error) {
	var tags []*types.Tag
	var total int64

	db := r.base.WithContext(ctx)

	// 构建查询条件
	whereClause := db.Model(&types.Tag{}).
		Where("tenant_id = ? AND knowledge_base_id = ?", tenantID, kbID)

	// 名称模糊搜索
	if query.Name != "" {
		whereClause = whereClause.Where("name LIKE ?", "%"+query.Name+"%")
	}

	// 统计总数
	if err := whereClause.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计标签数量失败: %w", err)
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	err := whereClause.
		Order("sort_order ASC, created_at DESC").
		Limit(query.PageSize).
		Offset(offset).
		Find(&tags).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询标签列表失败: %w", err)
	}

	return tags, total, nil
}

// FindByTenantID 根据租户ID查找标签列表
func (r *tagRepository) FindByTenantID(
	ctx context.Context,
	tenantID string,
	page, pageSize int,
) ([]*types.Tag, int64, error) {
	var tags []*types.Tag
	var total int64

	db := r.base.WithContext(ctx)

	// 统计总数
	if err := db.Model(&types.Tag{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计标签数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Where("tenant_id = ?", tenantID).
		Order("sort_order ASC, created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&tags).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询标签列表失败: %w", err)
	}

	return tags, total, nil
}

// FindByName 根据名称查找标签
func (r *tagRepository) FindByName(
	ctx context.Context,
	tenantID string,
	kbID int64,
	name string,
) (*types.Tag, error) {
	var tag types.Tag
	err := r.base.WithContext(ctx).
		Where("tenant_id = ? AND knowledge_base_id = ? AND name = ?", tenantID, kbID, name).
		First(&tag).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("标签不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询标签失败: %w", err)
	}

	return &tag, nil
}

// Update 更新标签
func (r *tagRepository) Update(ctx context.Context, tag *types.Tag) error {
	return r.base.Update(ctx, tag)
}

// Delete 删除标签（软删除）
func (r *tagRepository) Delete(ctx context.Context, id int64) error {
	db := r.base.WithContext(ctx)
	return db.Delete(&types.Tag{}, id).Error
}

// DeleteBatch 批量删除标签（软删除）
func (r *tagRepository) DeleteBatch(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	db := r.base.WithContext(ctx)
	return db.Delete(&types.Tag{}, ids).Error
}

// DeleteByKnowledgeBaseID 删除知识库的所有标签
func (r *tagRepository) DeleteByKnowledgeBaseID(
	ctx context.Context,
	tenantID string,
	kbID int64,
) error {
	db := r.base.WithContext(ctx)
	return db.
		Where("tenant_id = ? AND knowledge_base_id = ?", tenantID, kbID).
		Delete(&types.Tag{}).Error
}

// Exists 检查标签是否存在
func (r *tagRepository) Exists(
	ctx context.Context,
	tenantID string,
	kbID int64,
	name string,
) (bool, error) {
	var count int64
	err := r.base.WithContext(ctx).
		Model(&types.Tag{}).
		Where("tenant_id = ? AND knowledge_base_id = ? AND name = ?", tenantID, kbID, name).
		Count(&count).Error

	return count > 0, err
}

// CountByKnowledgeBaseID 统计知识库的标签数量
func (r *tagRepository) CountByKnowledgeBaseID(
	ctx context.Context,
	tenantID string,
	kbID int64,
) (int64, error) {
	var count int64
	err := r.base.WithContext(ctx).
		Model(&types.Tag{}).
		Where("tenant_id = ? AND knowledge_base_id = ?", tenantID, kbID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("统计标签数量失败: %w", err)
	}

	return count, nil
}

// UpdateSortOrder 批量更新排序
func (r *tagRepository) UpdateSortOrder(
	ctx context.Context,
	tagOrders []interfaces.TagSortOrder,
) error {
	if len(tagOrders) == 0 {
		return nil
	}

	db := r.base.WithContext(ctx)

	// 使用事务处理
	return db.Transaction(func(tx *gorm.DB) error {
		for _, order := range tagOrders {
			if err := tx.Model(&types.Tag{}).
				Where("id = ?", order.ID).
				Update("sort_order", order.SortOrder).Error; err != nil {
				return fmt.Errorf("更新标签 %d 排序失败: %w", order.ID, err)
			}
		}
		return nil
	})
}
