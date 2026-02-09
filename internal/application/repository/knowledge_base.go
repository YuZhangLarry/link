package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	common_repository "link/internal/common/repository"
	"link/internal/types"
	"link/internal/types/interfaces"
)

// ========================================
// 知识库仓储实现
// ========================================

// knowledgeBaseRepository 知识库仓储实现
type knowledgeBaseRepository struct {
	base *common_repository.BaseRepository
}

// NewKnowledgeBaseRepository 创建知识库仓储
func NewKnowledgeBaseRepository(db *gorm.DB, tenantEnabled bool) interfaces.KnowledgeBaseRepository {
	return &knowledgeBaseRepository{
		base: common_repository.NewBaseRepository(db, tenantEnabled),
	}
}

// Create 创建知识库
func (r *knowledgeBaseRepository) Create(ctx context.Context, kb *types.KnowledgeBase) error {
	return r.base.Create(ctx, kb)
}

// FindByID 根据ID查找知识库
func (r *knowledgeBaseRepository) FindByID(ctx context.Context, id string) (*types.KnowledgeBase, error) {
	var kb types.KnowledgeBase
	err := r.base.WithContext(ctx).
		Where("id = ?", id).
		First(&kb).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("知识库不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询知识库失败: %w", err)
	}

	return &kb, nil
}

// FindByTenantID 根据租户ID查找知识库列表
func (r *knowledgeBaseRepository) FindByTenantID(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.KnowledgeBase, int64, error) {
	var kbs []*types.KnowledgeBase
	var total int64

	db := r.base.WithContext(ctx)

	// 统计总数
	if err := db.Model(&types.KnowledgeBase{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计知识库数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&kbs).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询知识库列表失败: %w", err)
	}

	return kbs, total, nil
}

// FindByUser 根据用户ID查找知识库列表
func (r *knowledgeBaseRepository) FindByUser(ctx context.Context, userID int64, page, pageSize int) ([]*types.KnowledgeBase, int64, error) {
	var kbs []*types.KnowledgeBase
	var total int64

	db := r.base.WithContext(ctx)

	// 统计总数
	if err := db.Model(&types.KnowledgeBase{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计知识库数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&kbs).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询知识库列表失败: %w", err)
	}

	return kbs, total, nil
}

// Update 更新知识库
func (r *knowledgeBaseRepository) Update(ctx context.Context, kb *types.KnowledgeBase) error {
	return r.base.Update(ctx, kb)
}

// UpdateStats 更新知识库统计信息
func (r *knowledgeBaseRepository) UpdateStats(ctx context.Context, kbID string, documentCount, chunkCount int, storageSize int64) error {
	db := r.base.WithContext(ctx)
	return db.Model(&types.KnowledgeBase{}).
		Where("id = ?", kbID).
		Updates(map[string]interface{}{
			"document_count": documentCount,
			"chunk_count":    chunkCount,
			"storage_size":   storageSize,
		}).Error
}

// Delete 删除知识库（软删除）
func (r *knowledgeBaseRepository) Delete(ctx context.Context, id string) error {
	db := r.base.WithContext(ctx)
	return db.Delete(&types.KnowledgeBase{}, id).Error
}

// HardDelete 硬删除知识库及其所有关联数据
func (r *knowledgeBaseRepository) HardDelete(ctx context.Context, id string) error {
	db := r.base.WithContext(ctx)
	// 使用事务处理
	return db.Transaction(func(tx *gorm.DB) error {
		// 删除关联的分块
		if err := tx.Where("kb_id = ?", id).Delete(&types.Chunk{}).Error; err != nil {
			return fmt.Errorf("删除分块失败: %w", err)
		}

		// 删除关联的知识条目
		if err := tx.Where("kb_id = ?", id).Delete(&types.Knowledge{}).Error; err != nil {
			return fmt.Errorf("删除知识条目失败: %w", err)
		}

		// 删除知识库设置
		if err := tx.Where("kb_id = ?", id).Delete(&types.KBSetting{}).Error; err != nil {
			return fmt.Errorf("删除知识库设置失败: %w", err)
		}

		// 删除知识库
		if err := tx.Unscoped().Delete(&types.KnowledgeBase{}, id).Error; err != nil {
			return fmt.Errorf("删除知识库失败: %w", err)
		}

		return nil
	})
}

// Exists 检查知识库是否存在
func (r *knowledgeBaseRepository) Exists(ctx context.Context, id string) (bool, error) {
	var count int64
	err := r.base.WithContext(ctx).
		Model(&types.KnowledgeBase{}).
		Where("id = ?", id).
		Count(&count).Error

	return count > 0, err
}

// GetStorageStats 获取租户的存储统计
func (r *knowledgeBaseRepository) GetStorageStats(ctx context.Context, tenantID int64) (totalSize, kbCount int64, err error) {
	db := r.base.WithContext(ctx)

	// 统计总存储大小
	if err := db.Model(&types.KnowledgeBase{}).
		Where("tenant_id = ?", tenantID).
		Select("COALESCE(SUM(storage_size), 0)").
		Scan(&totalSize).Error; err != nil {
		return 0, 0, fmt.Errorf("统计存储大小失败: %w", err)
	}

	// 统计知识库数量
	if err := db.Model(&types.KnowledgeBase{}).
		Where("tenant_id = ?", tenantID).
		Count(&kbCount).Error; err != nil {
		return 0, 0, fmt.Errorf("统计知识库数量失败: %w", err)
	}

	return totalSize, kbCount, nil
}
