package repository

import (
	"context"
	"fmt"
	"strings"

	common_repository "link/internal/common"
	"link/internal/types"
	"link/internal/types/interfaces"

	"gorm.io/gorm"
)

// graphQueryRepository 图谱查询仓储实现（负责图谱数据与知识库分片表的关联查询）
type graphQueryRepository struct {
	base *common_repository.BaseRepository
}

// NewGraphQueryRepository 创建图谱查询仓储
func NewGraphQueryRepository(db *gorm.DB, tenantEnabled bool) interfaces.GraphQueryRepository {
	return &graphQueryRepository{
		base: common_repository.NewBaseRepository(db, tenantEnabled),
	}
}

// GetChunksByGraphNodes 根据图谱节点名称获取关联的分片
// 通过分片内容中包含节点名称来查找关联的分片
func (r *graphQueryRepository) GetChunksByGraphNodes(ctx context.Context, kbID string, nodeNames []string) ([]*types.Chunk, error) {
	if len(nodeNames) == 0 {
		return []*types.Chunk{}, nil
	}

	db := r.base.WithContext(ctx)
	var chunks []*types.Chunk

	// 构建查询：查找分片内容中包含任意节点名称的分片
	query := db.Model(&types.Chunk{}).
		Where("kb_id = ?", kbID).
		Where("is_enabled = ?", true)

	// 使用 LIKE 查询匹配节点名称
	for _, nodeName := range nodeNames {
		query = query.Or("content LIKE ?", "%"+nodeName+"%")
	}

	err := query.Order("chunk_index ASC").
		Limit(500).
		Find(&chunks).Error

	if err != nil {
		return nil, fmt.Errorf("查询图谱节点关联分片失败: %w", err)
	}

	return chunks, nil
}

// GetKnowledgeByGraphNodes 根据图谱节点名称获取关联的知识条目
// 返回包含这些节点的第一个知识条目
func (r *graphQueryRepository) GetKnowledgeByGraphNodes(ctx context.Context, kbID string, nodeNames []string) (*types.Knowledge, error) {
	if len(nodeNames) == 0 {
		return nil, fmt.Errorf("节点名称列表为空")
	}

	db := r.base.WithContext(ctx)
	var knowledge types.Knowledge

	// 通过关联的分片查找知识条目
	subQuery := db.Model(&types.Chunk{}).
		Select("DISTINCT knowledge_id").
		Where("kb_id = ?", kbID).
		Where("is_enabled = ?", true)

	// 添加节点名称匹配条件
	for _, nodeName := range nodeNames {
		subQuery = subQuery.Or("content LIKE ?", "%"+nodeName+"%")
	}

	err := db.Where("id IN (?)", subQuery).
		Where("kb_id = ?", kbID).
		First(&knowledge).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("未找到关联的知识条目")
	}
	if err != nil {
		return nil, fmt.Errorf("查询图谱节点关联知识条目失败: %w", err)
	}

	return &knowledge, nil
}

// GetGraphStats 获取图谱统计信息
// 统计知识库中与图谱相关的数据
func (r *graphQueryRepository) GetGraphStats(ctx context.Context, kbID string) (*interfaces.GraphStats, error) {
	db := r.base.WithContext(ctx)

	stats := &interfaces.GraphStats{
		ChunkIDs: []string{},
	}

	// 统计关联的分块数量
	err := db.Model(&types.Chunk{}).
		Where("kb_id = ? AND is_enabled = ?", kbID, true).
		Count(&stats.ChunkCount).Error
	if err != nil {
		return nil, fmt.Errorf("统计分块数量失败: %w", err)
	}

	// 获取关联的分块ID列表（用于后续图谱查询）
	var chunks []*types.Chunk
	err = db.Model(&types.Chunk{}).
		Where("kb_id = ? AND is_enabled = ?", kbID, true).
		Order("chunk_index ASC").
		Limit(1000).
		Find(&chunks).Error
	if err != nil {
		return nil, fmt.Errorf("查询分块ID列表失败: %w", err)
	}

	for _, chunk := range chunks {
		stats.ChunkIDs = append(stats.ChunkIDs, chunk.ID)
	}

	// 节点数和关系数从 Neo4j 获取，这里只返回0作为占位
	// 实际使用时需要结合 Neo4j 仓储获取完整统计
	stats.NodeCount = 0
	stats.RelationCount = 0

	return stats, nil
}

// ========================================
// 辅助方法
// ========================================

// BuildNodeNameCondition 构建节点名称匹配的 SQL 条件
func BuildNodeNameCondition(nodeNames []string) string {
	if len(nodeNames) == 0 {
		return "1=0"
	}
	conditions := make([]string, len(nodeNames))
	for i, name := range nodeNames {
		// 转义单引号
		escapedName := strings.ReplaceAll(name, "'", "''")
		conditions[i] = fmt.Sprintf("content LIKE '%%%s%%'", escapedName)
	}
	return strings.Join(conditions, " OR ")
}

// GetChunksByIDs 根据 chunk ID 列表获取分片内容（用于图谱检索）
func (r *graphQueryRepository) GetChunksByIDs(ctx context.Context, kbID string, chunkIDs []string) ([]*types.Chunk, error) {
	if len(chunkIDs) == 0 {
		return []*types.Chunk{}, nil
	}

	db := r.base.WithContext(ctx)
	var chunks []*types.Chunk

	err := db.Model(&types.Chunk{}).
		Where("kb_id = ?", kbID).
		Where("id IN ?", chunkIDs).
		Where("is_enabled = ?", true).
		Order("chunk_index ASC").
		Find(&chunks).Error

	if err != nil {
		return nil, fmt.Errorf("根据 chunk ID 列表查询分片失败: %w", err)
	}

	return chunks, nil
}
