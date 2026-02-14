package service

import (
	"context"

	"gorm.io/gorm"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// KnowledgeBaseService 知识库服务
type KnowledgeBaseService struct {
	kbRepo        interfaces.KnowledgeBaseRepository
	kbSettingRepo interfaces.KBSettingRepository
	knowledgeRepo interfaces.KnowledgeRepository
	chunkRepo     interfaces.ChunkRepository
	db            *gorm.DB
}

// NewKnowledgeBaseService 创建知识库服务
func NewKnowledgeBaseService(
	kbRepo interfaces.KnowledgeBaseRepository,
	kbSettingRepo interfaces.KBSettingRepository,
	knowledgeRepo interfaces.KnowledgeRepository,
	chunkRepo interfaces.ChunkRepository,
	db *gorm.DB,
) *KnowledgeBaseService {
	return &KnowledgeBaseService{
		kbRepo:        kbRepo,
		kbSettingRepo: kbSettingRepo,
		knowledgeRepo: knowledgeRepo,
		chunkRepo:     chunkRepo,
		db:            db,
	}
}

// ========================================
// 知识库CRUD操作
// ========================================

// Create 创建知识库
func (s *KnowledgeBaseService) Create(ctx context.Context, kb *types.KnowledgeBase, setting *types.KBSetting) error {
	// 先创建知识库
	if err := s.kbRepo.Create(ctx, kb); err != nil {
		return err
	}

	// 如果提供了设置，创建设置记录
	if setting != nil {
		setting.KBID = kb.ID
		if err := s.kbSettingRepo.Create(ctx, setting); err != nil {
			return err
		}
	}

	return nil
}

// FindByID 根据ID查找知识库
func (s *KnowledgeBaseService) FindByID(ctx context.Context, id string) (*types.KnowledgeBase, error) {
	return s.kbRepo.FindByID(ctx, id)
}

// FindByIDWithSettings 根据ID查找知识库并加载设置
func (s *KnowledgeBaseService) FindByIDWithSettings(ctx context.Context, id string) (*types.KnowledgeBase, error) {
	kb, err := s.kbRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 加载设置
	setting, err := s.kbSettingRepo.FindByKBID(ctx, id)
	if err == nil {
		kb.Setting = setting
	}

	return kb, nil
}

// FindByTenantID 根据租户ID查找知识库列表
func (s *KnowledgeBaseService) FindByTenantID(
	ctx context.Context,
	tenantID int64,
	page, pageSize int,
) ([]*types.KnowledgeBase, int64, error) {
	return s.kbRepo.FindByTenantID(ctx, tenantID, page, pageSize)
}

// Update 更新知识库
func (s *KnowledgeBaseService) Update(ctx context.Context, kb *types.KnowledgeBase) error {
	return s.kbRepo.Update(ctx, kb)
}

// UpdateWithSettings 更新知识库及其设置
func (s *KnowledgeBaseService) UpdateWithSettings(ctx context.Context, kb *types.KnowledgeBase, setting *types.KBSetting) error {
	// 更新知识库
	if err := s.kbRepo.Update(ctx, kb); err != nil {
		return err
	}

	// 如果提供了设置，更新设置记录
	if setting != nil {
		setting.KBID = kb.ID
		if setting.ID == 0 {
			// 检查设置是否存在
			exists, err := s.kbSettingRepo.Exists(ctx, kb.ID)
			if err != nil {
				return err
			}
			if !exists {
				setting.KBID = kb.ID
				return s.kbSettingRepo.Create(ctx, setting)
			}
		}
		if err := s.kbSettingRepo.Update(ctx, setting); err != nil {
			return err
		}
	}

	return nil
}

// Delete 删除知识库（软删除）
func (s *KnowledgeBaseService) Delete(ctx context.Context, id string) error {
	return s.kbRepo.Delete(ctx, id)
}

// Exists 检查知识库是否存在
func (s *KnowledgeBaseService) Exists(ctx context.Context, id string) (bool, error) {
	return s.kbRepo.Exists(ctx, id)
}

// ========================================
// 统计相关
// ========================================

// CreateChunk 创建分块
func (s *KnowledgeBaseService) CreateChunk(ctx context.Context, chunk *types.Chunk) error {
	return s.chunkRepo.Create(ctx, chunk)
}

// GetStats 获取知识库统计信息
func (s *KnowledgeBaseService) GetStats(ctx context.Context, kbID string) (*types.KnowledgeBaseStats, error) {
	stats := &types.KnowledgeBaseStats{
		KBID: kbID,
	}

	// 统计文档数量
	type KnowledgeCountResult struct {
		Total int64
	}
	var knowledgeResult KnowledgeCountResult
	err := s.db.Table("knowledges").
		Select("COALESCE(COUNT(*), 0) as total").
		Where("kb_id = ? AND deleted_at IS NULL", kbID).
		Scan(&knowledgeResult).Error
	if err != nil {
		return nil, err
	}

	// 统计分块数
	type ChunkCountResult struct {
		Total int64
	}
	var chunkResult ChunkCountResult
	err = s.db.Table("chunks").
		Select("COALESCE(COUNT(*), 0) as total").
		Where("kb_id = ? AND deleted_at IS NULL", kbID).
		Scan(&chunkResult).Error
	if err != nil {
		return nil, err
	}

	// 统计总大小
	type SizeResult struct {
		Total int64
	}
	var sizeResult SizeResult
	err = s.db.Table("knowledges").
		Select("COALESCE(SUM(storage_size), 0) as total").
		Where("kb_id = ? AND deleted_at IS NULL", kbID).
		Scan(&sizeResult).Error
	if err != nil {
		return nil, err
	}

	stats.KnowledgeCount = knowledgeResult.Total
	stats.ChunkCount = chunkResult.Total
	stats.TotalSize = sizeResult.Total

	return stats, nil
}

// ========================================
// 文档管理相关方法
// ========================================

// GetKnowledgeList 获取知识库的文档列表
func (s *KnowledgeBaseService) GetKnowledgeList(
	ctx context.Context,
	kbID string,
	page, pageSize int,
	status string,
) ([]*types.Knowledge, int64, error) {
	// 构建查询
	query := s.db.Table("knowledges").
		Select("id, kb_id, title, type, storage_size, parse_status, enable_status, created_at, processed_at, chunk_count").
		Where("kb_id = ? AND deleted_at IS NULL", kbID)

	if status != "" {
		query = query.Where("parse_status = ?", status)
	}

	// 获取总数
	var total int64
	countQuery := s.db.Table("knowledges").
		Where("kb_id = ? AND deleted_at IS NULL", kbID)
	if status != "" {
		countQuery = countQuery.Where("parse_status = ?", status)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	var knowledges []*types.Knowledge
	if err := query.
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&knowledges).Error; err != nil {
		return nil, 0, err
	}

	return knowledges, total, nil
}

// DeleteKnowledge 删除知识库文档（级联删除分块）
func (s *KnowledgeBaseService) DeleteKnowledge(ctx context.Context, kbID, knowledgeID string) error {
	// 使用事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除关联的分块
		if err := tx.Table("chunks").
			Where("knowledge_id = ? AND kb_id = ?", knowledgeID, kbID).
			Update("deleted_at", gorm.Expr("NOW()")).Error; err != nil {
			return err
		}

		// 软删除知识条目
		if err := tx.Table("knowledges").
			Where("id = ? AND kb_id = ?", knowledgeID, kbID).
			Update("deleted_at", gorm.Expr("NOW()")).Error; err != nil {
			return err
		}

		return nil
	})
}

// ========================================
// 分块管理相关方法
// ========================================

// GetChunks 获取知识库的分块列表
func (s *KnowledgeBaseService) GetChunks(
	ctx context.Context,
	kbID string,
	page, pageSize int,
	knowledgeID string,
) ([]*types.Chunk, int64, error) {
	// 构建查询
	query := s.db.Table("chunks").
		Select("id, kb_id, knowledge_id, chunk_index, content, token_count, is_enabled, created_at").
		Where("kb_id = ? AND deleted_at IS NULL", kbID)

	if knowledgeID != "" {
		query = query.Where("knowledge_id = ?", knowledgeID)
	}

	// 获取总数
	var total int64
	countQuery := s.db.Table("chunks").
		Where("kb_id = ? AND deleted_at IS NULL", kbID)
	if knowledgeID != "" {
		countQuery = countQuery.Where("knowledge_id = ?", knowledgeID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	var chunks []*types.Chunk
	if err := query.
		Order("knowledge_id, chunk_index").
		Limit(pageSize).
		Offset(offset).
		Find(&chunks).Error; err != nil {
		return nil, 0, err
	}

	return chunks, total, nil
}
