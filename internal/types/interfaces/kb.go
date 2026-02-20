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

	// ========================================
	// TagID 相关操作
	// ========================================

	// UpdateTagID 更新知识条目的标签ID
	UpdateTagID(ctx context.Context, id string, tagID int64) error

	// RemoveTagID 移除知识条目的标签ID（设置为0）
	RemoveTagID(ctx context.Context, id string) error

	// RemoveTagIDBatch 批量移除知识条目的标签ID
	RemoveTagIDBatch(ctx context.Context, ids []string, tagID int64) error

	// FindByTagID 根据标签ID查找知识条目列表
	FindByTagID(ctx context.Context, tenantID int64, tagID int64, page, pageSize int) ([]*types.Knowledge, int64, error)

	// AddTagIDBatch 批量为知识条目添加标签ID
	AddTagIDBatch(ctx context.Context, ids []string, tagID int64) error
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

	// ========================================
	// TagID 相关操作
	// ========================================

	// UpdateTagID 更新分块的标签ID
	UpdateTagID(ctx context.Context, id string, tagID int64) error

	// RemoveTagID 移除分块的标签ID（设置为0）
	RemoveTagID(ctx context.Context, id string) error

	// UpdateTagIDBatch 批量更新分块的标签ID
	UpdateTagIDBatch(ctx context.Context, ids []string, tagID int64) error

	// RemoveTagIDBatch 批量移除分块的标签ID
	RemoveTagIDBatch(ctx context.Context, ids []string, tagID int64) error

	// FindByTagID 根据标签ID查找分块列表
	FindByTagID(ctx context.Context, tenantID int64, tagID int64, page, pageSize int) ([]*types.Chunk, int64, error)

	// AddTagIDBatch 批量为分块添加标签ID
	AddTagIDBatch(ctx context.Context, ids []string, tagID int64) error

	// ========================================
	// 图谱关联查询相关操作
	// ========================================

	// FindByGraphNodes 根据图谱节点名称查找关联的分片
	FindByGraphNodes(ctx context.Context, kbID string, nodeNames []string) ([]*types.Chunk, error)

	// GetGraphStats 获取图谱关联的统计信息
	GetGraphStats(ctx context.Context, kbID string) (*GraphStats, error)
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

	// UpdateModelConfig 更新模型配置
	UpdateModelConfig(ctx context.Context, kbID string, embeddingModelID, summaryModelID, rerankModelID string) error

	// UpdateProcessingConfig 更新处理配置
	UpdateProcessingConfig(ctx context.Context, kbID string, chunkingConfig, imageProcessingConfig, cosConfig, vlmConfig, extractConfig string) error

	// Delete 删除设置
	Delete(ctx context.Context, kbID string) error

	// Exists 检查设置是否存在
	Exists(ctx context.Context, kbID string) (bool, error)
}

// RetrievalSettingRepository 检索设置仓储接口
type RetrievalSettingRepository interface {
	// Create 创建检索设置
	Create(ctx context.Context, setting *types.RetrievalSetting) error

	// FindByKBID 根据知识库ID查找检索设置
	FindByKBID(ctx context.Context, kbID string) (*types.RetrievalSetting, error)

	// FindByID 根据ID查找检索设置
	FindByID(ctx context.Context, id int64) (*types.RetrievalSetting, error)

	// FindBySessionID 根据会话ID查找检索设置
	FindBySessionID(ctx context.Context, sessionID string) (*types.RetrievalSetting, error)

	// Update 更新检索设置
	Update(ctx context.Context, setting *types.RetrievalSetting) error

	// Delete 删除检索设置
	Delete(ctx context.Context, kbID string) error

	// UpsertBySessionID 根据会话ID创建或更新检索设置
	UpsertBySessionID(ctx context.Context, sessionID string, tenantID int64, ragConfig *types.RAGConfig) error

	// UpdateVectorConfig 更新向量检索配置
	UpdateVectorConfig(ctx context.Context, kbID string, topK int, threshold float64, modelID string) error

	// UpdateBM25Config 更新BM25检索配置
	UpdateBM25Config(ctx context.Context, kbID string, topK int) error

	// UpdateGraphConfig 更新图谱检索配置
	UpdateGraphConfig(ctx context.Context, kbID string, enabled bool, topK int, minStrength float64) error

	// UpdateHybridConfig 更新混合检索配置
	UpdateHybridConfig(ctx context.Context, kbID string, alpha float64, rerankEnabled bool) error

	// UpdateWebConfig 更新网络搜索配置
	UpdateWebConfig(ctx context.Context, kbID string, enabled bool, topK int, engine, apiKey string, searchDepth int) error

	// UpdateRerankConfig 更新重排序配置
	UpdateRerankConfig(ctx context.Context, kbID string, enabled bool, modelID string) error

	// UpdateDefaultMode 更新默认检索模式
	UpdateDefaultMode(ctx context.Context, kbID string, mode string, availableModes []string) error

	// Exists 检查设置是否存在
	Exists(ctx context.Context, kbID string) (bool, error)
}

// ========================================
// 知识图谱仓储接口
// ========================================

// GraphStats 图谱统计信息
type GraphStats struct {
	NodeCount     int64    // 节点数量
	RelationCount int64    // 关系数量
	ChunkCount    int64    // 关联的分块数量
	ChunkIDs      []string // 关联的分块ID列表
}

// Neo4jGraphRepository Neo4j 图谱仓储（专门负责 Neo4j 操作）
type Neo4jGraphRepository interface {
	// AddGraph 添加图谱数据
	AddGraph(ctx context.Context, namespace types.NameSpace, graphs []*types.GraphData) error

	// AddRelation 添加单个关系
	AddRelation(ctx context.Context, namespace types.NameSpace, relation *types.GraphRelation) (*types.GraphRelation, error)

	// AddNode 添加单个节点
	AddNode(ctx context.Context, namespace types.NameSpace, node *types.GraphNode) error
	// DeleteNode 删除单个节点
	DeleteNode(ctx context.Context, namespace types.NameSpace, nodeID string) error
	// DeleteRelation 删除单个关系
	DeleteRelation(ctx context.Context, namespace types.NameSpace, relationID string) error

	// DeleteByChunkID 按 chunk_id 删除相关图谱数据（用于文档删除）
	DeleteByChunkID(ctx context.Context, namespace types.NameSpace, chunkID string) error

	// DeleteByKnowledgeID 按 knowledge_id 删除相关图谱数据（用于文档删除）
	DeleteByKnowledgeID(ctx context.Context, namespace types.NameSpace, knowledgeID string) error

	// DeleteGraph 删除图谱数据
	DeleteGraph(ctx context.Context, namespaces []types.NameSpace) error

	// GetGraph 获取知识库的完整图谱数据
	GetGraph(ctx context.Context, namespace types.NameSpace) (*types.GraphData, error)

	// SearchNode 搜索节点
	SearchNode(ctx context.Context, namespace types.NameSpace, nodes []string) (*types.GraphData, error)

	// SearchNodeV2 搜索节点（改进版 - 直接返回节点名称）
	SearchNodeV2(ctx context.Context, namespace types.NameSpace, nodes []string) (*types.GraphData, error)

	// SearchPath 搜索路径
	SearchPath(ctx context.Context, namespace types.NameSpace, startNode, endNode string, maxDepth int) ([]*types.GraphData, error)

	// CheckHealth 检查 Neo4j 连接健康状态
	CheckHealth(ctx context.Context) error

	// UpdateNode 更新节点属性
	UpdateNode(ctx context.Context, namespace types.NameSpace, node *types.GraphNode) error

	// UpdateRelation 更新关系属性，返回更新后的关系
	UpdateRelation(ctx context.Context, namespace types.NameSpace, relation *types.GraphRelation) (*types.GraphRelation, error)

	// Close 关闭连接
	Close(ctx context.Context) error
}

// GraphQueryRepository 图谱查询仓储（负责与知识库的关联查询）
type GraphQueryRepository interface {
	// GetChunksByGraphNodes 根据图谱节点名称获取关联的分片
	GetChunksByGraphNodes(ctx context.Context, kbID string, nodeNames []string) ([]*types.Chunk, error)

	// GetChunksByIDs 根据 chunk ID 列表获取分片内容（用于图谱检索）
	GetChunksByIDs(ctx context.Context, kbID string, chunkIDs []string) ([]*types.Chunk, error)

	// GetKnowledgeByGraphNodes 根据图谱节点名称获取关联的知识条目
	GetKnowledgeByGraphNodes(ctx context.Context, kbID string, nodeNames []string) (*types.Knowledge, error)

	// GetGraphStats 获取图谱统计信息
	GetGraphStats(ctx context.Context, kbID string) (*GraphStats, error)
}

// GraphRepository 知识图谱仓储接口（向后兼容的别名）
// Deprecated: 使用 Neo4jGraphRepository 替代
type GraphRepository = Neo4jGraphRepository

// ========================================
// 知识标签仓储接口
// ========================================

// TagRepository 知识标签仓储接口
type TagRepository interface {
	// Create 创建标签
	Create(ctx context.Context, tag *types.Tag) error

	// CreateBatch 批量创建标签
	CreateBatch(ctx context.Context, tags []*types.Tag) error

	// FindByID 根据ID查找标签
	FindByID(ctx context.Context, id int64) (*types.Tag, error)

	// FindByKnowledgeBaseID 根据知识库ID查找标签列表
	FindByKnowledgeBaseID(ctx context.Context, tenantID string, kbID int64, query *types.TagListQuery) ([]*types.Tag, int64, error)

	// FindByTenantID 根据租户ID查找标签列表
	FindByTenantID(ctx context.Context, tenantID string, page, pageSize int) ([]*types.Tag, int64, error)

	// FindByName 根据名称查找标签
	FindByName(ctx context.Context, tenantID string, kbID int64, name string) (*types.Tag, error)

	// Update 更新标签
	Update(ctx context.Context, tag *types.Tag) error

	// Delete 删除标签（软删除）
	Delete(ctx context.Context, id int64) error

	// DeleteBatch 批量删除标签（软删除）
	DeleteBatch(ctx context.Context, ids []int64) error

	// DeleteByKnowledgeBaseID 删除知识库的所有标签
	DeleteByKnowledgeBaseID(ctx context.Context, tenantID string, kbID int64) error

	// Exists 检查标签是否存在
	Exists(ctx context.Context, tenantID string, kbID int64, name string) (bool, error)

	// CountByKnowledgeBaseID 统计知识库的标签数量
	CountByKnowledgeBaseID(ctx context.Context, tenantID string, kbID int64) (int64, error)

	// UpdateSortOrder 批量更新排序
	UpdateSortOrder(ctx context.Context, tagOrders []TagSortOrder) error
}

// TagSortOrder 标签排序信息
type TagSortOrder struct {
	ID        int64 `json:"id"`
	SortOrder int   `json:"sort_order"`
}

// ========================================
// 测评仓储接口
// ========================================

// EvaluationRepository 测评任务仓储接口
type EvaluationRepository interface {
	// Create 创建测评任务
	Create(ctx context.Context, task *types.EvaluationTask) error

	// FindByID 根据ID查找测评任务
	FindByID(ctx context.Context, id string) (*types.EvaluationTask, error)

	// FindByTenantID 根据租户ID查找测评任务列表
	FindByTenantID(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.EvaluationTask, int64, error)

	// FindByStatus 根据状态查找测评任务
	FindByStatus(ctx context.Context, status int) ([]*types.EvaluationTask, error)

	// Update 更新测评任务
	Update(ctx context.Context, task *types.EvaluationTask) error

	// UpdateStatus 更新任务状态
	UpdateStatus(ctx context.Context, id string, status int, errMsg string) error

	// UpdateProgress 更新任务进度
	UpdateProgress(ctx context.Context, id string, finished int) error

	// Delete 删除测评任务（软删除）
	Delete(ctx context.Context, id string) error
}

// DatasetRepository 数据集仓储接口
type DatasetRepository interface {
	// FindByDatasetID 根据数据集ID获取QA对
	FindByDatasetID(ctx context.Context, tenantID int64, datasetID string) ([]*types.QAPair, error)

	// Create 创建数据集记录
	Create(ctx context.Context, record *types.DatasetRecord) error

	// CreateBatch 批量创建数据集记录
	CreateBatch(ctx context.Context, records []*types.DatasetRecord) error

	// FindByTenantID 根据租户ID查找数据集
	FindByTenantID(ctx context.Context, tenantID int64) ([]string, error)
}
