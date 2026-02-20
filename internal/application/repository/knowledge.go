package repository

import (
	"context"
	"fmt"
	"link/internal/common"

	"gorm.io/gorm"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// ========================================
// 知识条目仓储实现
// ========================================

// knowledgeRepository 知识条目仓储实现
type knowledgeRepository struct {
	base *common.BaseRepository
}

// NewKnowledgeRepository 创建知识条目仓储
func NewKnowledgeRepository(db *gorm.DB, tenantEnabled bool) interfaces.KnowledgeRepository {
	return &knowledgeRepository{
		base: common.NewBaseRepository(db, tenantEnabled),
	}
}

// Create 创建知识条目
func (r *knowledgeRepository) Create(ctx context.Context, knowledge *types.Knowledge) error {
	return r.base.Create(ctx, knowledge)
}

// CreateBatch 批量创建知识条目
func (r *knowledgeRepository) CreateBatch(ctx context.Context, knowledgeList []*types.Knowledge) error {
	if len(knowledgeList) == 0 {
		return nil
	}
	db := r.base.WithContext(ctx)
	return db.CreateInBatches(knowledgeList, 100).Error
}

// FindByID 根据ID查找知识条目
func (r *knowledgeRepository) FindByID(ctx context.Context, id string) (*types.Knowledge, error) {
	var knowledge types.Knowledge
	err := r.base.WithContext(ctx).
		Where("id = ?", id).
		First(&knowledge).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("知识条目不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询知识条目失败: %w", err)
	}

	return &knowledge, nil
}

// FindByKBID 根据知识库ID查找知识条目列表
func (r *knowledgeRepository) FindByKBID(ctx context.Context, kbID string, query *types.KnowledgeListQuery) ([]*types.Knowledge, int64, error) {
	var knowledges []*types.Knowledge
	var total int64

	db := r.base.WithContext(ctx).Model(&types.Knowledge{}).Where("kb_id = ?", kbID)

	// 添加过滤条件
	if query != nil {
		if query.Type != "" {
			db = db.Where("type = ?", query.Type)
		}
		if query.ParseStatus != "" {
			db = db.Where("parse_status = ?", query.ParseStatus)
		}
		if query.EnableStatus != "" {
			db = db.Where("enable_status = ?", query.EnableStatus)
		}
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计知识条目数量失败: %w", err)
	}

	// 分页查询
	page := 1
	pageSize := 20
	if query != nil {
		page = query.Page
		pageSize = query.PageSize
	}

	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&knowledges).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询知识条目列表失败: %w", err)
	}

	return knowledges, total, nil
}

// FindByTenantID 根据租户ID查找知识条目列表
func (r *knowledgeRepository) FindByTenantID(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.Knowledge, int64, error) {
	var knowledges []*types.Knowledge
	var total int64

	db := r.base.WithContext(ctx)

	// 统计总数
	if err := db.Model(&types.Knowledge{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计知识条目数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&knowledges).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询知识条目列表失败: %w", err)
	}

	return knowledges, total, nil
}

// Update 更新知识条目
func (r *knowledgeRepository) Update(ctx context.Context, knowledge *types.Knowledge) error {
	return r.base.Update(ctx, knowledge)
}

// UpdateParseStatus 更新解析状态
func (r *knowledgeRepository) UpdateParseStatus(ctx context.Context, id string, parseStatus string, errorMessage string) error {
	db := r.base.WithContext(ctx)
	updates := map[string]interface{}{
		"parse_status": parseStatus,
	}
	if parseStatus == "completed" {
		now := common.NowPtr()
		updates["processed_at"] = now
	}
	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}

	return db.Model(&types.Knowledge{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateChunkCount 更新分块数量
func (r *knowledgeRepository) UpdateChunkCount(ctx context.Context, id string, chunkCount int) error {
	db := r.base.WithContext(ctx)
	return db.Model(&types.Knowledge{}).
		Where("id = ?", id).
		Update("chunk_count", chunkCount).Error
}

// Delete 删除知识条目（软删除）
func (r *knowledgeRepository) Delete(ctx context.Context, id string) error {
	db := r.base.WithContext(ctx)
	return db.Delete(&types.Knowledge{}, id).Error
}

// DeleteBatch 批量删除知识条目（软删除）
func (r *knowledgeRepository) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	db := r.base.WithContext(ctx)
	return db.Delete(&types.Knowledge{}, ids).Error
}

// HardDelete 硬删除知识条目
func (r *knowledgeRepository) HardDelete(ctx context.Context, id string) error {
	db := r.base.WithContext(ctx)
	return db.Unscoped().Delete(&types.Knowledge{}, id).Error
}

// HardDeleteBatch 批量硬删除
func (r *knowledgeRepository) HardDeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	db := r.base.WithContext(ctx)
	return db.Unscoped().Delete(&types.Knowledge{}, ids).Error
}

// FindByStatus 根据解析状态查找待处理的知识条目
func (r *knowledgeRepository) FindByStatus(ctx context.Context, tenantID int64, parseStatus string, limit int) ([]*types.Knowledge, error) {
	var knowledges []*types.Knowledge

	db := r.base.WithContext(ctx)
	query := db.Model(&types.Knowledge{}).Where("parse_status = ?", parseStatus)

	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}

	err := query.
		Order("created_at ASC").
		Limit(limit).
		Find(&knowledges).Error

	if err != nil {
		return nil, fmt.Errorf("查询待处理知识条目失败: %w", err)
	}

	return knowledges, nil
}

// FindByFileHash 根据文件哈希查找（去重用）
func (r *knowledgeRepository) FindByFileHash(ctx context.Context, tenantID int64, fileHash string) (*types.Knowledge, error) {
	var knowledge types.Knowledge

	db := r.base.WithContext(ctx)
	err := db.Where("tenant_id = ? AND file_hash = ?", tenantID, fileHash).
		First(&knowledge).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil // 不存在不算错误
	}
	if err != nil {
		return nil, fmt.Errorf("查询知识条目失败: %w", err)
	}

	return &knowledge, nil
}

// ========================================
// TagID 相关操作
// ========================================

// UpdateTagID 更新知识条目的标签ID
func (r *knowledgeRepository) UpdateTagID(ctx context.Context, id string, tagID int64) error {
	db := r.base.WithContext(ctx)
	return db.Model(&types.Knowledge{}).
		Where("id = ?", id).
		Update("tag_id", tagID).Error
}

// RemoveTagID 移除知识条目的标签ID（设置为0）
func (r *knowledgeRepository) RemoveTagID(ctx context.Context, id string) error {
	return r.UpdateTagID(ctx, id, 0)
}

// RemoveTagIDBatch 批量移除知识条目的标签ID
func (r *knowledgeRepository) RemoveTagIDBatch(ctx context.Context, ids []string, tagID int64) error {
	if len(ids) == 0 {
		return nil
	}
	db := r.base.WithContext(ctx)
	return db.Model(&types.Knowledge{}).
		Where("id IN ?", ids).
		Update("tag_id", 0).Error
}

// FindByTagID 根据标签ID查找知识条目列表
func (r *knowledgeRepository) FindByTagID(ctx context.Context, tenantID int64, tagID int64, page, pageSize int) ([]*types.Knowledge, int64, error) {
	var knowledges []*types.Knowledge
	var total int64

	db := r.base.WithContext(ctx).Model(&types.Knowledge{}).Where("tenant_id = ? AND tag_id = ?", tenantID, tagID)

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计知识条目数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&knowledges).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询知识条目列表失败: %w", err)
	}

	return knowledges, total, nil
}

// AddTagIDBatch 批量为知识条目添加标签ID
func (r *knowledgeRepository) AddTagIDBatch(ctx context.Context, ids []string, tagID int64) error {
	if len(ids) == 0 {
		return nil
	}
	db := r.base.WithContext(ctx)
	return db.Model(&types.Knowledge{}).
		Where("id IN ?", ids).
		Update("tag_id", tagID).Error
}
