package service

import (
	"context"
	"link/internal/models/chat"
	"log"
	"math"
	"sync"

	"link/internal/types"
	"link/internal/types/interfaces"

	"github.com/google/uuid"
)

// GraphService 图谱服务
type GraphService struct {
	graphRepo interfaces.GraphRepository
	chatModel chat.Chat
	mutex     sync.RWMutex
}

// NewGraphService 创建图谱服务实例
func NewGraphService(graphRepo interfaces.GraphRepository) *GraphService {
	return &GraphService{
		graphRepo: graphRepo,
	}
}

// AddGraph 添加图谱数据
func (s *GraphService) AddGraph(ctx context.Context, namespace types.NameSpace, graphs []*types.GraphData) error {
	// 为节点生成UUID（如果还没有）
	for _, graph := range graphs {
		for _, node := range graph.Node {
			if node.ID == "" {
				node.ID = uuid.New().String()
			}
			// 如果没有标题，使用名称作为标题
			if node.Title == "" {
				node.Title = node.Name
			}
		}

		// 为关系生成UUID并计算属性
		for _, rel := range graph.Relation {
			if rel.ID == "" {
				rel.ID = uuid.New().String()
			}
			// 计算关系属性（占位符，实际计算在后续实现）
			s.calculateRelationProperties(graph, rel)
		}
	}

	return s.graphRepo.AddGraph(ctx, namespace, graphs)
}

// calculateRelationProperties 计算关系属性
func (s *GraphService) calculateRelationProperties(graph *types.GraphData, rel *types.GraphRelation) {
	// 设置默认值
	if rel.Strength == 0 {
		rel.Strength = 5.0 // 默认强度为5
	}
	if rel.Weight == 0 {
		rel.Weight = 5.0 // 默认权重为5
	}

	// TODO: 在后续实现中添加以下计算逻辑
	// 1. Weight 计算 (graph.go:575处设置)
	//    - 由 PMI (Pointwise Mutual Information) 和 Strength 组合计算
	//    - 公式: Weight = normalize(PMI + Strength) 到 1-10 范围
	//
	// 2. CombinedDegree 计算 (graph.go:615处计算)
	//    - 源实体和目标实体的度数之和
	//    - 需要查找两个节点的出度+入度
	//
	// 3. ChunkIDs 更新 (graph.go:271处生成)
	//    - 记录该关系在哪些文档块中被识别到
	//    - 通过 findRelationChunkIDs 计算共同出现的文档块
	//
	// 4. 加权平均更新 (graph.go:292处使用)
	//    - 当关系已存在时，使用加权平均更新现有属性
}

// calculatePMI 计算 PMI (Pointwise Mutual Information)
// TODO: 实现具体的 PMI 计算逻辑
func (s *GraphService) calculatePMI(source, target string, chunkIDs []string) float64 {
	// PMI(x,y) = log( P(x,y) / (P(x) * P(y)) )
	// 需要统计:
	// - P(x,y): 源和目标同时出现的概率
	// - P(x): 源出现的概率
	// - P(y): 目标出现的概率
	log.Printf("[GraphService] PMI calculation for %s -> %s (not yet implemented)", source, target)
	return 0.0
}

// findRelationChunkIDs 查找关系中共同出现的文档块
// TODO: 实现具体的查找逻辑
func (s *GraphService) findRelationChunkIDs(source, target string) []string {
	// 需要查询:
	// 1. 源实体出现在哪些文档块
	// 2. 目标实体出现在哪些文档块
	// 3. 返回交集（共同出现的文档块）
	log.Printf("[GraphService] Finding chunk IDs for %s -> %s (not yet implemented)", source, target)
	return []string{}
}

// calculateNodeDegree 计算节点的度数（出度+入度）
// TODO: 实现具体的度数计算逻辑
func (s *GraphService) calculateNodeDegree(nodeID string) (int, error) {
	// 需要查询:
	// 1. 该节点所有出度关系数量
	// 2. 该节点所有入度关系数量
	// 3. 返回总和
	log.Printf("[GraphService] Calculating degree for node %s (not yet implemented)", nodeID)
	return 0, nil
}

// calculateCombinedDegree 计算两个节点的组合度数
func (s *GraphService) calculateCombinedDegree(sourceID, targetID string) (int, error) {
	sourceDegree, err := s.calculateNodeDegree(sourceID)
	if err != nil {
		return 0, err
	}
	targetDegree, err := s.calculateNodeDegree(targetID)
	if err != nil {
		return 0, err
	}
	return sourceDegree + targetDegree, nil
}

// normalizeTo1To10 将数值标准化到 1-10 范围
func normalizeTo1To10(value, min, max float64) float64 {
	if max == min {
		return 5.0 // 默认中间值
	}
	normalized := (value - min) / (max - min) * 9  // 0-9 范围
	return math.Max(1, math.Min(10, normalized+1)) // 转换到 1-10 范围
}

// updateRelationWithWeightedAverage 使用加权平均更新关系属性
func (s *GraphService) updateRelationWithWeightedAverage(
	existing *types.GraphRelation,
	new *types.GraphRelation,
	weight float64,
) {
	// weight 是新数据的权重（0-1），现有数据的权重为 (1-weight)
	// 使用加权平均更新各个属性

	// Strength 加权平均更新 (graph.go:292处使用)
	if existing.Strength > 0 && new.Strength > 0 {
		existing.Strength = existing.Strength*(1-weight) + new.Strength*weight
	} else if new.Strength > 0 {
		existing.Strength = new.Strength
	}

	// Weight 加权平均更新
	if existing.Weight > 0 && new.Weight > 0 {
		existing.Weight = existing.Weight*(1-weight) + new.Weight*weight
	} else if new.Weight > 0 {
		existing.Weight = new.Weight
	}

	// 合并 ChunkIDs（不重复）
	for _, chunkID := range new.ChunkIDs {
		existing.AddChunkID(chunkID)
	}

	// Description 可能需要更复杂的合并逻辑
	if new.Description != "" && existing.Description == "" {
		existing.Description = new.Description
	}
}

// DeleteGraph 删除图谱数据
func (s *GraphService) DeleteGraph(ctx context.Context, namespaces []types.NameSpace) error {
	return s.graphRepo.DeleteGraph(ctx, namespaces)
}

// SearchNode 搜索节点
func (s *GraphService) SearchNode(
	ctx context.Context,
	namespace types.NameSpace,
	nodes []string,
) (*types.GraphData, error) {
	return s.graphRepo.SearchNode(ctx, namespace, nodes)
}

// SearchPath 搜索路径
func (s *GraphService) SearchPath(
	ctx context.Context,
	namespace types.NameSpace,
	startNode, endNode string,
	maxDepth int,
) ([]*types.GraphData, error) {
	return s.graphRepo.SearchPath(ctx, namespace, startNode, endNode, maxDepth)
}

// CheckHealth 检查图谱服务健康状态
func (s *GraphService) CheckHealth(ctx context.Context) error {
	return s.graphRepo.CheckHealth(ctx)
}

// ========================================
// 关系分析相关方法（占位符实现）
// ========================================

// FindRelationsByChunk 查找涉及特定文档块的所有关系
func (s *GraphService) FindRelationsByChunk(
	ctx context.Context,
	namespace types.NameSpace,
	chunkID string,
) ([]*types.GraphRelation, error) {
	// TODO: 实现 Neo4j 查询
	// MATCH (a)-[r:RELATES_TO]->(b)
	// WHERE $chunkID IN r.chunk_ids
	// RETURN r
	log.Printf("[GraphService] Finding relations for chunk %s (not yet implemented)", chunkID)
	return []*types.GraphRelation{}, nil
}

// FindStrongRelations 查找强度高于阈值的关系
func (s *GraphService) FindStrongRelations(
	ctx context.Context,
	namespace types.NameSpace,
	minStrength float64,
) ([]*types.GraphRelation, error) {
	// TODO: 实现 Neo4j 查询
	// MATCH (a)-[r:RELATES_TO]->(b)
	// WHERE r.strength >= $minStrength
	// RETURN a, r, b
	// ORDER BY r.strength DESC
	log.Printf("[GraphService] Finding strong relations with strength >= %.2f (not yet implemented)", minStrength)
	return []*types.GraphRelation{}, nil
}

// GetRelationStatistics 获取关系统计信息
func (s *GraphService) GetRelationStatistics(
	ctx context.Context,
	namespace types.NameSpace,
) (*RelationStatistics, error) {
	// TODO: 实现统计逻辑
	// - 总关系数
	// - 平均强度
	// - 平均权重
	// - 强度分布
	return &RelationStatistics{}, nil
}

// RelationStatistics 关系统计信息
type RelationStatistics struct {
	TotalRelations       int         // 总关系数
	AverageStrength      float64     // 平均强度
	AverageWeight        float64     // 平均权重
	AverageChunkCount    float64     // 平均每关系关联的文档块数
	StrengthDistribution map[int]int // 强度分布 (1-10 各有多少关系)
}

// UpdateRelationWeight 更新关系权重（由 PMI 和 Strength 计算）
func (s *GraphService) UpdateRelationWeight(
	ctx context.Context,
	namespace types.NameSpace,
	relationID string,
) error {
	// TODO: 实现权重更新逻辑
	// 1. 计算 PMI
	// 2. 获取 Strength
	// 3. 组合计算 Weight
	// 4. 更新 Neo4j
	log.Printf("[GraphService] Updating weight for relation %s (not yet implemented)", relationID)
	return nil
}

// BatchUpdateRelations 批量更新关系属性
func (s *GraphService) BatchUpdateRelations(
	ctx context.Context,
	namespace types.NameSpace,
	relations []*types.GraphRelation,
) error {
	for _, rel := range relations {
		if err := s.UpdateRelationWeight(ctx, namespace, rel.ID); err != nil {
			log.Printf("[GraphService] Failed to update relation %s: %v", rel.ID, err)
			// 继续处理其他关系
		}
	}
	return nil
}

// GetRelatedChunks 获取与特定实体相关的所有文档块
func (s *GraphService) GetRelatedChunks(
	ctx context.Context,
	namespace types.NameSpace,
	entityTitle string,
) ([]string, error) {
	// TODO: 实现 Neo4j 查询
	// MATCH (n {title: $entity_title})
	// RETURN n.chunks AS chunks
	log.Printf("[GraphService] Getting related chunks for entity %s (not yet implemented)", entityTitle)
	return []string{}, nil
}

// FindCommonChunks 查找两个实体共同出现的文档块
func (s *GraphService) FindCommonChunks(
	ctx context.Context,
	namespace types.NameSpace,
	entity1, entity2 string,
) ([]string, error) {
	// TODO: 实现 Neo4j 查询
	// MATCH (a {title: $entity1})-[:RELATES_TO]->(b {title: $entity2})
	// RETURN a.chunks + b.chunks AS common_chunks
	// 或者更精确地，查找同时包含两个实体的文档块
	log.Printf("[GraphService] Finding common chunks between %s and %s (not yet implemented)", entity1, entity2)
	return []string{}, nil
}

// CalculateRelationMetrics 计算关系的所有指标
func (s *GraphService) CalculateRelationMetrics(
	ctx context.Context,
	namespace types.NameSpace,
	rel *types.GraphRelation,
	sourceNode, targetNode *types.GraphNode,
) error {
	// 1. 计算 CombinedDegree
	combinedDegree, err := s.calculateCombinedDegree(sourceNode.ID, targetNode.ID)
	if err != nil {
		log.Printf("[GraphService] Failed to calculate combined degree: %v", err)
	}
	rel.CombinedDegree = combinedDegree

	// 2. 计算 PMI
	// TODO: 需要从图谱数据中统计概率
	// pmi := s.calculatePMI(rel.Source, rel.Target, rel.ChunkIDs)

	// 3. 计算 Weight (由 PMI 和 Strength 组合)
	// TODO: 实际公式待定
	// rel.Weight = normalizeTo1To10(pmi + rel.Strength, 0, 20)

	log.Printf("[GraphService] Calculated metrics for relation %s: CombinedDegree=%d",
		rel.ID, rel.CombinedDegree)

	return nil
}
