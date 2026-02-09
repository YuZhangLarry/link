package interfaces

import (
	"context"

	"link/internal/types"
)

// ========================================
// 知识库仓储接口
// ========================================

// KnowledgeBaseRepository 知识库仓储接口
type KnowledgeBaseRepository interface {
	// Create 创建知识库
	Create(ctx context.Context, kb *types.KnowledgeBase) error

	// FindByID 根据ID查找知识库
	FindByID(ctx context.Context, id string) (*types.KnowledgeBase, error)

	// FindByTenantID 根据租户ID查找知识库列表
	FindByTenantID(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.KnowledgeBase, int64, error)

	// FindByUser 根据用户ID查找知识库列表
	FindByUser(ctx context.Context, userID int64, page, pageSize int) ([]*types.KnowledgeBase, int64, error)

	// Update 更新知识库
	Update(ctx context.Context, kb *types.KnowledgeBase) error

	// UpdateStats 更新知识库统计信息
	UpdateStats(ctx context.Context, kbID string, documentCount, chunkCount int, storageSize int64) error

	// Delete 删除知识库（软删除）
	Delete(ctx context.Context, id string) error

	// HardDelete 硬删除知识库及其所有关联数据
	HardDelete(ctx context.Context, id string) error

	// Exists 检查知识库是否存在
	Exists(ctx context.Context, id string) (bool, error)

	// GetStorageStats 获取租户的存储统计
	GetStorageStats(ctx context.Context, tenantID int64) (totalSize, kbCount int64, err error)
}

// KnowledgeRepository 知识条目仓储接口
type KnowledgeRepository interface {
	// Create 创建知识条目
	Create(ctx context.Context, knowledge *types.Knowledge) error

	// CreateBatch 批量创建知识条目
	CreateBatch(ctx context.Context, knowledgeList []*types.Knowledge) error

	// FindByID 根据ID查找知识条目
	FindByID(ctx context.Context, id string) (*types.Knowledge, error)

	// FindByKBID 根据知识库ID查找知识条目列表
	FindByKBID(ctx context.Context, kbID string, query *types.KnowledgeListQuery) ([]*types.Knowledge, int64, error)

	// FindByTenantID 根据租户ID查找知识条目列表
	FindByTenantID(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.Knowledge, int64, error)

	// Update 更新知识条目
	Update(ctx context.Context, knowledge *types.Knowledge) error

	// UpdateParseStatus 更新解析状态
	UpdateParseStatus(ctx context.Context, id string, parseStatus string, errorMessage string) error

	// UpdateChunkCount 更新分块数量
	UpdateChunkCount(ctx context.Context, id string, chunkCount int) error

	// Delete 删除知识条目（软删除）
	Delete(ctx context.Context, id string) error

	// DeleteBatch 批量删除知识条目（软删除）
	DeleteBatch(ctx context.Context, ids []string) error

	// HardDelete 硬删除知识条目
	HardDelete(ctx context.Context, id string) error

	// HardDeleteBatch 批量硬删除
	HardDeleteBatch(ctx context.Context, ids []string) error

	// FindByStatus 根据解析状态查找待处理的知识条目
	FindByStatus(ctx context.Context, tenantID int64, parseStatus string, limit int) ([]*types.Knowledge, error)

	// FindByFileHash 根据文件哈希查找（去重用）
	FindByFileHash(ctx context.Context, tenantID int64, fileHash string) (*types.Knowledge, error)
}

// ChunkRepository 文档分块仓储接口
type ChunkRepository interface {
	// Create 创建分块
	Create(ctx context.Context, chunk *types.Chunk) error

	// CreateBatch 批量创建分块
	CreateBatch(ctx context.Context, chunks []*types.Chunk) error

	// FindByID 根据ID查找分块
	FindByID(ctx context.Context, id string) (*types.Chunk, error)

	// FindByKBID 根据知识库ID查找分块列表
	FindByKBID(ctx context.Context, kbID string, query *types.ChunkListQuery) ([]*types.Chunk, int64, error)

	// FindByKnowledgeID 根据知识条目ID查找分块列表
	FindByKnowledgeID(ctx context.Context, knowledgeID string, enabledOnly bool) ([]*types.Chunk, error)

	// Update 更新分块
	Update(ctx context.Context, chunk *types.Chunk) error

	// UpdateEmbeddingID 更新向量ID
	UpdateEmbeddingID(ctx context.Context, id string, embeddingID string) error

	// UpdateBatchStatus 批量更新启用状态
	UpdateBatchStatus(ctx context.Context, ids []string, isEnabled bool) error

	// Delete 删除分块（软删除）
	Delete(ctx context.Context, id string) error

	// DeleteByKnowledgeID 删除知识条目的所有分块
	DeleteByKnowledgeID(ctx context.Context, knowledgeID string) error

	// DeleteByKBID 删除知识库的所有分块
	DeleteByKBID(ctx context.Context, kbID string) error

	// HardDelete 硬删除分块
	HardDelete(ctx context.Context, id string) error

	// HardDeleteBatch 批量硬删除
	HardDeleteBatch(ctx context.Context, ids []string) error

	// FindEnabledChunks 查找启用的分块（用于检索）
	FindEnabledChunks(ctx context.Context, kbID string, limit int) ([]*types.Chunk, error)

	// CountByKBID 统计知识库的分块数量
	CountByKBID(ctx context.Context, kbID string) (int64, error)
}

// KBSettingRepository 知识库设置仓储接口
type KBSettingRepository interface {
	// Create 创建设置
	Create(ctx context.Context, setting *types.KBSetting) error

	// FindByKBID 根据知识库ID查找设置
	FindByKBID(ctx context.Context, kbID string) (*types.KBSetting, error)

	// Update 更新设置
	Update(ctx context.Context, setting *types.KBSetting) error

	// UpdateRetrievalConfig 更新检索配置
	UpdateRetrievalConfig(ctx context.Context, kbID string, mode string, threshold float64, topK int) error

	// Delete 删除设置
	Delete(ctx context.Context, kbID string) error

	// Exists 检查设置是否存在
	Exists(ctx context.Context, kbID string) (bool, error)
}
