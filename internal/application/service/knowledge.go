package service

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// KnowledgeService 知识库服务
type KnowledgeService struct {
	knowledgeRepo interfaces.KnowledgeRepository
	chunkRepo     interfaces.ChunkRepository
	kbSettingRepo interfaces.KBSettingRepository
	db            *gorm.DB
}

// NewKnowledgeService 创建知识库服务
func NewKnowledgeService(
	knowledgeRepo interfaces.KnowledgeRepository,
	chunkRepo interfaces.ChunkRepository,
	kbSettingRepo interfaces.KBSettingRepository,
	db *gorm.DB,
) *KnowledgeService {
	return &KnowledgeService{
		knowledgeRepo: knowledgeRepo,
		chunkRepo:     chunkRepo,
		kbSettingRepo: kbSettingRepo,
		db:            db,
	}
}

// ========================================
// Knowledge CRUD 操作
// ========================================

// Create 创建知识条目
func (s *KnowledgeService) Create(ctx context.Context, knowledge *types.Knowledge) error {
	return s.knowledgeRepo.Create(ctx, knowledge)
}

// CreateBatch 批量创建知识条目
func (s *KnowledgeService) CreateBatch(ctx context.Context, knowledgeList []*types.Knowledge) error {
	return s.knowledgeRepo.CreateBatch(ctx, knowledgeList)
}

// FindByID 根据ID查找知识条目
func (s *KnowledgeService) FindByID(ctx context.Context, id string) (*types.Knowledge, error) {
	return s.knowledgeRepo.FindByID(ctx, id)
}

// FindByKBID 根据知识库ID查找知识条目列表
func (s *KnowledgeService) FindByKBID(
	ctx context.Context,
	kbID string,
	query *types.KnowledgeListQuery,
) ([]*types.Knowledge, int64, error) {
	return s.knowledgeRepo.FindByKBID(ctx, kbID, query)
}

// FindByTenantID 根据租户ID查找知识条目列表
func (s *KnowledgeService) FindByTenantID(
	ctx context.Context,
	tenantID int64,
	page, pageSize int,
) ([]*types.Knowledge, int64, error) {
	return s.knowledgeRepo.FindByTenantID(ctx, tenantID, page, pageSize)
}

// Update 更新知识条目
func (s *KnowledgeService) Update(ctx context.Context, knowledge *types.Knowledge) error {
	return s.knowledgeRepo.Update(ctx, knowledge)
}

// UpdateParseStatus 更新解析状态
func (s *KnowledgeService) UpdateParseStatus(
	ctx context.Context,
	id string,
	parseStatus string,
	errorMessage string,
) error {
	return s.knowledgeRepo.UpdateParseStatus(ctx, id, parseStatus, errorMessage)
}

// UpdateChunkCount 更新分块数量
func (s *KnowledgeService) UpdateChunkCount(ctx context.Context, id string, chunkCount int) error {
	return s.knowledgeRepo.UpdateChunkCount(ctx, id, chunkCount)
}

// Delete 删除知识条目（软删除）
func (s *KnowledgeService) Delete(ctx context.Context, id string) error {
	return s.knowledgeRepo.Delete(ctx, id)
}

// DeleteBatch 批量删除知识条目（软删除）
func (s *KnowledgeService) DeleteBatch(ctx context.Context, ids []string) error {
	return s.knowledgeRepo.DeleteBatch(ctx, ids)
}

// HardDelete 硬删除知识条目
func (s *KnowledgeService) HardDelete(ctx context.Context, id string) error {
	return s.knowledgeRepo.HardDelete(ctx, id)
}

// FindByStatus 根据解析状态查找待处理的知识条目
func (s *KnowledgeService) FindByStatus(
	ctx context.Context,
	tenantID int64,
	parseStatus string,
	limit int,
) ([]*types.Knowledge, error) {
	return s.knowledgeRepo.FindByStatus(ctx, tenantID, parseStatus, limit)
}

// FindByFileHash 根据文件哈希查找（去重用）
func (s *KnowledgeService) FindByFileHash(
	ctx context.Context,
	tenantID int64,
	fileHash string,
) (*types.Knowledge, error) {
	return s.knowledgeRepo.FindByFileHash(ctx, tenantID, fileHash)
}

// ========================================
// Chunk 操作
// ========================================

// FindChunkByID 查找分块
func (s *KnowledgeService) FindChunkByID(ctx context.Context, id string) (*types.Chunk, error) {
	return s.chunkRepo.FindByID(ctx, id)
}

// FindChunkByKBID 查找知识库的所有分块
func (s *KnowledgeService) FindChunkByKBID(
	ctx context.Context,
	kbID string,
	page, pageSize int,
) ([]*types.Chunk, int64, error) {
	query := &types.ChunkListQuery{
		Page:     page,
		PageSize: pageSize,
	}
	return s.chunkRepo.FindByKBID(ctx, kbID, query)
}

// FindChunkByKnowledgeID 根据知识条目ID查找分块列表
func (s *KnowledgeService) FindChunkByKnowledgeID(
	ctx context.Context,
	knowledgeID string,
	enabledOnly bool,
) ([]*types.Chunk, error) {
	return s.chunkRepo.FindByKnowledgeID(ctx, knowledgeID, enabledOnly)
}

// CreateChunk 创建分块
func (s *KnowledgeService) CreateChunk(ctx context.Context, chunk *types.Chunk) error {
	return s.chunkRepo.Create(ctx, chunk)
}

// CreateChunkBatch 批量创建分块
func (s *KnowledgeService) CreateChunkBatch(ctx context.Context, chunks []*types.Chunk) error {
	return s.chunkRepo.CreateBatch(ctx, chunks)
}

// UpdateChunk 更新分块
func (s *KnowledgeService) UpdateChunk(ctx context.Context, chunk *types.Chunk) error {
	return s.chunkRepo.Update(ctx, chunk)
}

// DeleteChunk 删除分块（软删除）
func (s *KnowledgeService) DeleteChunk(ctx context.Context, id string) error {
	return s.chunkRepo.Delete(ctx, id)
}

// DeleteChunkByKnowledgeID 删除知识条目的所有分块
func (s *KnowledgeService) DeleteChunkByKnowledgeID(ctx context.Context, knowledgeID string) error {
	return s.chunkRepo.DeleteByKnowledgeID(ctx, knowledgeID)
}

// ========================================
// KBSetting 操作
// ========================================

// FindKBSettingByKBID 根据知识库ID查找设置
func (s *KnowledgeService) FindKBSettingByKBID(
	ctx context.Context,
	kbID string,
) (*types.KBSetting, error) {
	return s.kbSettingRepo.FindByKBID(ctx, kbID)
}

// FindKBSettingByKBIDAndKey 根据知识库ID和键查找设置
func (s *KnowledgeService) FindKBSettingByKBIDAndKey(
	ctx context.Context,
	kbID string,
	key string,
) (*types.KBSetting, error) {
	setting, err := s.kbSettingRepo.FindByKBID(ctx, kbID)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, nil
	}
	// 简化实现：返回整个设置
	return setting, nil
}

// CreateKBSetting 创建知识库设置
func (s *KnowledgeService) CreateKBSetting(ctx context.Context, setting *types.KBSetting) error {
	return s.kbSettingRepo.Create(ctx, setting)
}

// UpdateKBSetting 更新知识库设置
func (s *KnowledgeService) UpdateKBSetting(ctx context.Context, setting *types.KBSetting) error {
	return s.kbSettingRepo.Update(ctx, setting)
}

// UpdateRetrievalConfig 更新检索配置
func (s *KnowledgeService) UpdateRetrievalConfig(
	ctx context.Context,
	kbID string,
	mode string,
	threshold float64,
	topK int,
) error {
	return s.kbSettingRepo.UpdateRetrievalConfig(ctx, kbID, mode, threshold, topK)
}

// ========================================
// 辅助方法
// ========================================

// GetDB 获取数据库连接（供 handler 使用事务）
func (s *KnowledgeService) GetDB() *gorm.DB {
	return s.db
}

// GetKnowledgeByID 根据ID获取知识条目（别名）
func (s *KnowledgeService) GetKnowledgeByID(ctx context.Context, id string) (*types.Knowledge, error) {
	return s.FindByID(ctx, id)
}

// GetChunksByKnowledgeID 获取知识条目的所有分块（别名）
func (s *KnowledgeService) GetChunksByKnowledgeID(
	ctx context.Context,
	knowledgeID string,
) ([]*types.Chunk, error) {
	return s.FindChunkByKnowledgeID(ctx, knowledgeID, false)
}

// CountChunksByKnowledgeID 统计知识条目的分块数量
func (s *KnowledgeService) CountChunksByKnowledgeID(
	ctx context.Context,
	knowledgeID string,
) (int, error) {
	chunks, err := s.FindChunkByKnowledgeID(ctx, knowledgeID, false)
	if err != nil {
		return 0, err
	}
	return len(chunks), nil
}

// ========================================
// 事务操作
// ========================================

// WithTransaction 执行事务操作
func (s *KnowledgeService) WithTransaction(fn func(tx *gorm.DB) error) error {
	return s.db.Transaction(fn)
}

// Transaction 开始事务
func (s *KnowledgeService) Transaction() *gorm.DB {
	return s.db.Begin()
}

// ========================================
// 统计操作
// ========================================

// CountByKBID 统计知识库的知识条目数量
func (s *KnowledgeService) CountByKBID(ctx context.Context, kbID string) (int64, error) {
	knowledges, _, err := s.FindByKBID(ctx, kbID, &types.KnowledgeListQuery{})
	if err != nil {
		return 0, err
	}
	return int64(len(knowledges)), nil
}

// CountByTenantID 统计租户的知识条目数量
func (s *KnowledgeService) CountByTenantID(ctx context.Context, tenantID int64) (int64, error) {
	knowledges, _, err := s.FindByTenantID(ctx, tenantID, 1, 1000)
	if err != nil {
		return 0, err
	}
	return int64(len(knowledges)), nil
}

// GetPendingKnowledge 获取待处理的知识条目
func (s *KnowledgeService) GetPendingKnowledge(
	ctx context.Context,
	tenantID int64,
	limit int,
) ([]*types.Knowledge, error) {
	return s.FindByStatus(ctx, tenantID, "pending", limit)
}

// GetProcessingKnowledge 获取处理中的知识条目
func (s *KnowledgeService) GetProcessingKnowledge(
	ctx context.Context,
	tenantID int64,
	limit int,
) ([]*types.Knowledge, error) {
	return s.FindByStatus(ctx, tenantID, "processing", limit)
}

// MarkAsProcessing 标记为处理中
func (s *KnowledgeService) MarkAsProcessing(ctx context.Context, id string) error {
	return s.UpdateParseStatus(ctx, id, "processing", "")
}

// MarkAsCompleted 标记为已完成
func (s *KnowledgeService) MarkAsCompleted(ctx context.Context, id string, chunkCount int) error {
	if err := s.UpdateChunkCount(ctx, id, chunkCount); err != nil {
		return fmt.Errorf("failed to update chunk count: %w", err)
	}
	return s.UpdateParseStatus(ctx, id, "completed", "")
}

// MarkAsFailed 标记为失败
func (s *KnowledgeService) MarkAsFailed(ctx context.Context, id string, errMsg string) error {
	return s.UpdateParseStatus(ctx, id, "failed", errMsg)
}
