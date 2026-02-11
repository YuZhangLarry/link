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
// 文档分块仓储实现
// ========================================

// chunkRepository 文档分块仓储实现
type chunkRepository struct {
	base *common_repository.BaseRepository
}

// NewChunkRepository 创建分块仓储
func NewChunkRepository(db *gorm.DB, tenantEnabled bool) interfaces.ChunkRepository {
	return &chunkRepository{
		base: common_repository.NewBaseRepository(db, tenantEnabled),
	}
}

// Create 创建分块
func (r *chunkRepository) Create(ctx context.Context, chunk *types.Chunk) error {
	return r.base.Create(ctx, chunk)
}

// CreateBatch 批量创建分块
func (r *chunkRepository) CreateBatch(ctx context.Context, chunks []*types.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}
	db := r.base.WithContext(ctx)

	// 使用事务批量插入
	return db.Transaction(func(tx *gorm.DB) error {
		// 分批插入，每批 1000 条
		batchSize := 1000
		for i := 0; i < len(chunks); i += batchSize {
			end := i + batchSize
			if end > len(chunks) {
				end = len(chunks)
			}

			batch := chunks[i:end]
			if err := tx.CreateInBatches(batch, 100).Error; err != nil {
				return fmt.Errorf("批量插入分块失败: %w", err)
			}
		}
		return nil
	})
}

// FindByID 根据ID查找分块
func (r *chunkRepository) FindByID(ctx context.Context, id string) (*types.Chunk, error) {
	var chunk types.Chunk
	err := r.base.WithContext(ctx).
		Where("id = ?", id).
		First(&chunk).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("分块不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询分块失败: %w", err)
	}

	return &chunk, nil
}

// FindByKBID 根据知识库ID查找分块列表
func (r *chunkRepository) FindByKBID(ctx context.Context, kbID string, query *types.ChunkListQuery) ([]*types.Chunk, int64, error) {
	var chunks []*types.Chunk
	var total int64

	db := r.base.WithContext(ctx).Model(&types.Chunk{}).Where("kb_id = ?", kbID)

	// 添加过滤条件
	if query != nil {
		if query.IsEnabled != nil {
			db = db.Where("is_enabled = ?", *query.IsEnabled)
		}
		if query.KnowledgeID != "" {
			db = db.Where("knowledge_id = ?", query.KnowledgeID)
		}
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计分块数量失败: %w", err)
	}

	// 分页查询
	page := 1
	pageSize := 20
	if query != nil {
		page = query.Page
		pageSize = query.PageSize
	}

	offset := (page - 1) * pageSize
	err := db.Order("chunk_index ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&chunks).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询分块列表失败: %w", err)
	}

	return chunks, total, nil
}

// FindByKnowledgeID 根据知识条目ID查找分块列表
func (r *chunkRepository) FindByKnowledgeID(ctx context.Context, knowledgeID string, enabledOnly bool) ([]*types.Chunk, error) {
	var chunks []*types.Chunk

	db := r.base.WithContext(ctx).
		Where("knowledge_id = ?", knowledgeID)

	if enabledOnly {
		db = db.Where("is_enabled = ?", true)
	}

	err := db.Order("chunk_index ASC").
		Find(&chunks).Error

	if err != nil {
		return nil, fmt.Errorf("查询分块列表失败: %w", err)
	}

	return chunks, nil
}

// Update 更新分块
func (r *chunkRepository) Update(ctx context.Context, chunk *types.Chunk) error {
	return r.base.Update(ctx, chunk)
}

// UpdateEmbeddingID 更新向量ID
func (r *chunkRepository) UpdateEmbeddingID(ctx context.Context, id string, embeddingID string) error {
	db := r.base.WithContext(ctx)
	return db.Model(&types.Chunk{}).
		Where("id = ?", id).
		Update("embedding_id", embeddingID).Error
}

// UpdateBatchStatus 批量更新启用状态
func (r *chunkRepository) UpdateBatchStatus(ctx context.Context, ids []string, isEnabled bool) error {
	if len(ids) == 0 {
		return nil
	}
	db := r.base.WithContext(ctx)
	return db.Model(&types.Chunk{}).
		Where("id IN ?", ids).
		Update("is_enabled", isEnabled).Error
}

// Delete 删除分块（软删除）
func (r *chunkRepository) Delete(ctx context.Context, id string) error {
	db := r.base.WithContext(ctx)
	return db.Delete(&types.Chunk{}, id).Error
}

// DeleteByKnowledgeID 删除知识条目的所有分块
func (r *chunkRepository) DeleteByKnowledgeID(ctx context.Context, knowledgeID string) error {
	db := r.base.WithContext(ctx)
	return db.Where("knowledge_id = ?", knowledgeID).Delete(&types.Chunk{}).Error
}

// DeleteByKBID 删除知识库的所有分块
func (r *chunkRepository) DeleteByKBID(ctx context.Context, kbID string) error {
	db := r.base.WithContext(ctx)
	return db.Where("kb_id = ?", kbID).Delete(&types.Chunk{}).Error
}

// HardDelete 硬删除分块
func (r *chunkRepository) HardDelete(ctx context.Context, id string) error {
	db := r.base.WithContext(ctx)
	return db.Unscoped().Delete(&types.Chunk{}, id).Error
}

// HardDeleteBatch 批量硬删除
func (r *chunkRepository) HardDeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	db := r.base.WithContext(ctx)
	return db.Unscoped().Delete(&types.Chunk{}, ids).Error
}

// FindEnabledChunks 查找启用的分块（用于检索）
func (r *chunkRepository) FindEnabledChunks(ctx context.Context, kbID string, limit int) ([]*types.Chunk, error) {
	var chunks []*types.Chunk

	db := r.base.WithContext(ctx)
	err := db.Where("kb_id = ? AND is_enabled = ?", kbID, true).
		Order("chunk_index ASC").
		Limit(limit).
		Find(&chunks).Error

	if err != nil {
		return nil, fmt.Errorf("查询启用分块失败: %w", err)
	}

	return chunks, nil
}

// CountByKBID 统计知识库的分块数量
func (r *chunkRepository) CountByKBID(ctx context.Context, kbID string) (int64, error) {
	var count int64

	db := r.base.WithContext(ctx)
	err := db.Model(&types.Chunk{}).
		Where("kb_id = ?", kbID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("统计分块数量失败: %w", err)
	}

	return count, nil
}

// UpdateBatchEmbeddingIDs 批量更新向量ID（用于向量检索后回写）
func (r *chunkRepository) UpdateBatchEmbeddingIDs(ctx context.Context, chunkIDs []string, embeddingIDs []string) error {
	if len(chunkIDs) == 0 || len(chunkIDs) != len(embeddingIDs) {
		return fmt.Errorf("参数错误：chunkIDs 和 embeddingIDs 长度必须相等且不为空")
	}

	db := r.base.WithContext(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		for i := 0; i < len(chunkIDs); i++ {
			if err := tx.Model(&types.Chunk{}).
				Where("id = ?", chunkIDs[i]).
				Update("embedding_id", embeddingIDs[i]).Error; err != nil {
				return fmt.Errorf("更新向量ID失败: %w", err)
			}
		}
		return nil
	})
}

// FindChunksWithoutEmbedding 查找没有向量的分块（用于批处理向量化）
func (r *chunkRepository) FindChunksWithoutEmbedding(ctx context.Context, kbID string, limit int) ([]*types.Chunk, error) {
	var chunks []*types.Chunk

	db := r.base.WithContext(ctx)
	query := db.Model(&types.Chunk{}).Where("embedding_id = ?", "")

	if kbID != "" {
		query = query.Where("kb_id = ?", kbID)
	}

	err := query.Where("is_enabled = ?", true).
		Order("created_at ASC").
		Limit(limit).
		Find(&chunks).Error

	if err != nil {
		return nil, fmt.Errorf("查询无向量分块失败: %w", err)
	}

	return chunks, nil
}

// ========================================
// TagID 相关操作
// ========================================

// UpdateTagID 更新分块的标签ID
func (r *chunkRepository) UpdateTagID(ctx context.Context, id string, tagID int64) error {
	db := r.base.WithContext(ctx)
	return db.Model(&types.Chunk{}).
		Where("id = ?", id).
		Update("tag_id", tagID).Error
}

// RemoveTagID 移除分块的标签ID（设置为0）
func (r *chunkRepository) RemoveTagID(ctx context.Context, id string) error {
	db := r.base.WithContext(ctx)
	return db.Model(&types.Chunk{}).
		Where("id = ?", id).
		Update("tag_id", 0).Error
}

// UpdateTagIDBatch 批量更新分块的标签ID
func (r *chunkRepository) UpdateTagIDBatch(ctx context.Context, ids []string, tagID int64) error {
	if len(ids) == 0 {
		return nil
	}
	db := r.base.WithContext(ctx)
	return db.Model(&types.Chunk{}).
		Where("id IN ?", ids).
		Update("tag_id", tagID).Error
}

// RemoveTagIDBatch 批量移除分块的标签ID
func (r *chunkRepository) RemoveTagIDBatch(ctx context.Context, ids []string, tagID int64) error {
	if len(ids) == 0 {
		return nil
	}
	db := r.base.WithContext(ctx)
	return db.Model(&types.Chunk{}).
		Where("id IN ?", ids).
		Update("tag_id", tagID).Error
}

// FindByTagID 根据标签ID查找分块列表
func (r *chunkRepository) FindByTagID(ctx context.Context, tenantID int64, tagID int64, page, pageSize int) ([]*types.Chunk, int64, error) {
	var chunks []*types.Chunk
	var total int64

	db := r.base.WithContext(ctx)

	// 统计总数
	countQuery := db.Model(&types.Chunk{}).Where("tenant_id = ? AND tag_id = ?", tenantID, tagID)
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计分块数量失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Where("tenant_id = ? AND tag_id = ?", tenantID, tagID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&chunks).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询分块列表失败: %w", err)
	}

	return chunks, total, nil
}

// AddTagIDBatch 批量为分块添加标签ID
func (r *chunkRepository) AddTagIDBatch(ctx context.Context, ids []string, tagID int64) error {
	if len(ids) == 0 {
		return nil
	}
	db := r.base.WithContext(ctx)
	return db.Model(&types.Chunk{}).
		Where("id IN ?", ids).
		Update("tag_id", tagID).Error
}
