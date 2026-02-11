package neo4j

import (
	"context"
	"fmt"
	"log"
	"strings"

	"link/internal/types"
	"link/internal/types/interfaces"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Neo4jRepository Neo4j 知识图谱仓储
type Neo4jRepository struct {
	driver     neo4j.DriverWithContext
	nodePrefix string
}

// NewNeo4jRepository 创建 Neo4j 仓储实例
func NewNeo4jRepository(driver neo4j.DriverWithContext) interfaces.GraphRepository {
	return &Neo4jRepository{
		driver:     driver,
		nodePrefix: "ENTITY",
	}
}

// removeHyphen 移除字符串中的连字符
func removeHyphen(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

// Labels 返回命名空间的标签列表
func (n *Neo4jRepository) Labels(namespace types.NameSpace) []string {
	res := make([]string, 0)
	// 添加节点前缀
	res = append(res, n.nodePrefix)
	// 添加类型标签
	if namespace.Type != "" {
		res = append(res, removeHyphen(namespace.Type))
	}
	// 添加 KB 标签
	if namespace.KBID != "" {
		res = append(res, "KB_"+removeHyphen(namespace.KBID[:8]))
	}
	return res
}

// Label 返回命名空间的标签表达式
func (n *Neo4jRepository) Label(namespace types.NameSpace) string {
	labels := n.Labels(namespace)
	return strings.Join(labels, ":")
}

// AddGraph 添加图谱数据
func (n *Neo4jRepository) AddGraph(ctx context.Context, namespace types.NameSpace, graphs []*types.GraphData) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT RETRIEVE GRAPH - driver is nil")
		return nil
	}

	for _, graph := range graphs {
		if err := n.addGraph(ctx, namespace, graph); err != nil {
			return fmt.Errorf("failed to add graph: %w", err)
		}
	}
	return nil
}

// addGraph 添加单个图谱
func (n *Neo4jRepository) addGraph(ctx context.Context, namespace types.NameSpace, graph *types.GraphData) error {
	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 1. 创建节点（使用 UUID 作为节点ID）
		nodeQuery := `
			UNWIND $data AS row
			MERGE (n:` + n.Label(namespace) + ` {id: row.id})
			SET n.name = row.name
			SET n.title = COALESCE(row.title, row.name)
			SET n.entity_type = row.entity_type
			SET n.knowledge_id = $knowledge_id
			SET n.chunks = COALESCE(n.chunks, [])
			SET n.attributes = COALESCE(row.attributes, [])
			RETURN n
		`

		nodeData := make([]map[string]interface{}, 0)
		for _, node := range graph.Node {
			nodeData = append(nodeData, map[string]interface{}{
				"id":          node.ID,
				"name":        node.Name,
				"title":       node.Title,
				"entity_type": node.EntityType,
				"attributes":  node.Attributes,
				"chunks":      node.Chunks,
			})
		}

		if len(nodeData) > 0 {
			_, err := tx.Run(ctx, nodeQuery, map[string]interface{}{
				"data":         nodeData,
				"knowledge_id": namespace.Knowledge,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create nodes: %w", err)
			}
		}

		// 2. 创建关系（使用新的关系结构）
		// 先查找节点的度数，用于计算 CombinedDegree
		relQuery := `
			UNWIND $data AS row
			MATCH (source:` + n.Label(namespace) + ` {id: row.source_id})
			MATCH (target:` + n.Label(namespace) + ` {id: row.target_id})

			// 获取源节点和目标节点的度数
			WITH source, target, row
			OPTIONAL MATCH (source)-[outRel]-()
			WITH source, target, row, count(outRel) AS sourceDegree
			OPTIONAL MATCH (target)-[inRel]-()
			WITH source, target, row, sourceDegree, count(inRel) AS targetDegree

			// 计算 CombinedDegree
			WITH source, target, row, sourceDegree + targetDegree AS combinedDegree

			// 使用关系ID创建唯一关系
			MERGE (source)-[r:RELATES_TO {id: row.id}]->(target)
			SET r.source = row.source
			SET r.target = row.target
			SET r.description = row.description
			SET r.strength = COALESCE(row.strength, 5.0)
			SET r.weight = COALESCE(row.weight, 5.0)
			SET r.combined_degree = combinedDegree

			// 更新 ChunkIDs - 如果有 chunk_id 且不在列表中，则添加
			WITH r, row
			WHERE row.chunk_id IS NOT NULL
			SET r.chunk_ids = CASE
				WHEN row.chunk_id IN r.chunk_ids THEN r.chunk_ids
				ELSE COALESCE(r.chunk_ids, []) + row.chunk_id
			END

			RETURN r
		`

		relData := make([]map[string]interface{}, 0)
		for _, rel := range graph.Relation {
			// 为每个 ChunkID 创建一条关系记录
			if len(rel.ChunkIDs) > 0 {
				for _, chunkID := range rel.ChunkIDs {
					relData = append(relData, map[string]interface{}{
						"id":          rel.ID,
						"source_id":   n.findNodeIDByName(graph.Node, rel.Source),
						"target_id":   n.findNodeIDByName(graph.Node, rel.Target),
						"source":      rel.Source,
						"target":      rel.Target,
						"description": rel.Description,
						"strength":    rel.Strength,
						"weight":      rel.Weight,
						"chunk_id":    chunkID,
					})
				}
			} else {
				// 如果没有 ChunkIDs，仍然创建关系（chunk_id 为 null）
				relData = append(relData, map[string]interface{}{
					"id":          rel.ID,
					"source_id":   n.findNodeIDByName(graph.Node, rel.Source),
					"target_id":   n.findNodeIDByName(graph.Node, rel.Target),
					"source":      rel.Source,
					"target":      rel.Target,
					"description": rel.Description,
					"strength":    rel.Strength,
					"weight":      rel.Weight,
					"chunk_id":    nil,
				})
			}
		}

		if len(relData) > 0 {
			_, err := tx.Run(ctx, relQuery, map[string]interface{}{
				"data":         relData,
				"knowledge_id": namespace.Knowledge,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create relationships: %w", err)
			}
		}

		return nil, nil
	})

	if err != nil {
		log.Printf("[Neo4j] ERROR: failed to add graph: %v", err)
		return err
	}

	log.Printf("[Neo4j] Successfully added graph for knowledge_id=%s, nodes=%d, relations=%d",
		namespace.Knowledge, len(graph.Node), len(graph.Relation))
	return nil
}

// findNodeIDByName 根据实体标题查找节点ID
func (n *Neo4jRepository) findNodeIDByName(nodes []*types.GraphNode, title string) string {
	for _, node := range nodes {
		if node.Title == title || node.Name == title {
			return node.ID
		}
	}
	return ""
}

// DeleteGraph 删除图谱数据
func (n *Neo4jRepository) DeleteGraph(ctx context.Context, namespaces []types.NameSpace) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT RETRIEVE GRAPH - driver is nil")
		return nil
	}

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		for _, namespace := range namespaces {
			labelExpr := n.Label(namespace)

			// 先删除关系
			deleteRelsQuery := fmt.Sprintf(`
				MATCH (n:%s {knowledge_id: $knowledge_id})-[r]-(m:%s {knowledge_id: $knowledge_id})
				DELETE r
				RETURN count(r) AS deleted_count
			`, labelExpr, labelExpr)

			_, err := tx.Run(ctx, deleteRelsQuery, map[string]interface{}{
				"knowledge_id": namespace.Knowledge,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to delete relationships: %w", err)
			}

			// 再删除节点
			deleteNodesQuery := fmt.Sprintf(`
				MATCH (n:%s {knowledge_id: $knowledge_id})
				DELETE n
				RETURN count(n) AS deleted_count
			`, labelExpr)

			_, err = tx.Run(ctx, deleteNodesQuery, map[string]interface{}{
				"knowledge_id": namespace.Knowledge,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to delete nodes: %w", err)
			}
		}
		return nil, nil
	})

	if err != nil {
		log.Printf("[Neo4j] ERROR: failed to delete graph: %v", err)
		return err
	}

	log.Printf("[Neo4j] Successfully deleted graph for %d namespaces", len(namespaces))
	return nil
}

// SearchNode 搜索节点
func (n *Neo4jRepository) SearchNode(
	ctx context.Context,
	namespace types.NameSpace,
	nodes []string,
) (*types.GraphData, error) {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT RETRIEVE GRAPH - driver is nil")
		return nil, nil
	}

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		labelExpr := n.Label(namespace)
		query := `
			MATCH (n:` + labelExpr + ` {knowledge_id: $knowledge_id})-[r:RELATES_TO]-(m:` + labelExpr + ` {knowledge_id: $knowledge_id})
			WHERE ANY(nodeText IN $nodes WHERE toLower(n.title) CONTAINS toLower(nodeText) OR toLower(n.name) CONTAINS toLower(nodeText))
			RETURN n, r, m
			LIMIT 1000
		`

		params := map[string]interface{}{
			"nodes":        nodes,
			"knowledge_id": namespace.Knowledge,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, fmt.Errorf("failed to run query: %w", err)
		}

		graphData := &types.GraphData{
			Node:     make([]*types.GraphNode, 0),
			Relation: make([]*types.GraphRelation, 0),
		}
		nodeSeen := make(map[string]bool)
		relSeen := make(map[string]bool)

		for result.Next(ctx) {
			record := result.Record()

			// 获取节点和关系
			nodeVal, hasNode := record.Get("n")
			if !hasNode {
				continue
			}
			relVal, hasRel := record.Get("r")
			if !hasRel {
				continue
			}
			targetNodeVal, hasTarget := record.Get("m")
			if !hasTarget {
				continue
			}

			node := nodeVal.(neo4j.Node)
			targetNode := targetNodeVal.(neo4j.Node)

			// 处理节点
			for _, n := range []neo4j.Node{node, targetNode} {
				idVal, ok := n.Props["id"]
				if !ok {
					continue
				}
				idStr := fmt.Sprintf("%v", idVal)

				if _, seen := nodeSeen[idStr]; !seen {
					nodeSeen[idStr] = true
					graphNode := &types.GraphNode{
						ID:         idStr,
						Attributes: []string{},
						Chunks:     []string{},
					}

					// 获取名称
					if name, ok := n.Props["name"]; ok {
						graphNode.Name = fmt.Sprintf("%v", name)
					}

					// 获取标题
					if title, ok := n.Props["title"]; ok {
						graphNode.Title = fmt.Sprintf("%v", title)
					} else {
						graphNode.Title = graphNode.Name
					}

					// 获取属性
					if attrs, ok := n.Props["attributes"]; ok {
						graphNode.Attributes = interfaceListToStringList(attrs.([]interface{}))
					}

					// 获取关联的分块
					if chunks, ok := n.Props["chunks"]; ok {
						graphNode.Chunks = interfaceListToStringList(chunks.([]interface{}))
					}

					// 获取实体类型
					if entityType, ok := n.Props["entity_type"]; ok {
						graphNode.EntityType = fmt.Sprintf("%v", entityType)
					}

					graphData.Node = append(graphData.Node, graphNode)
				}
			}

			// 处理关系
			if rel, ok := relVal.(neo4j.Relationship); ok {
				relIDVal, ok := rel.Props["id"]
				if !ok {
					continue
				}
				relID := fmt.Sprintf("%v", relIDVal)
				relKey := relID

				if _, seen := relSeen[relKey]; !seen {
					relSeen[relKey] = true

					graphRelation := &types.GraphRelation{
						ID:       relID,
						ChunkIDs: []string{},
					}

					// 获取基本属性
					if source, ok := rel.Props["source"]; ok {
						graphRelation.Source = fmt.Sprintf("%v", source)
					}
					if target, ok := rel.Props["target"]; ok {
						graphRelation.Target = fmt.Sprintf("%v", target)
					}
					if description, ok := rel.Props["description"]; ok {
						graphRelation.Description = fmt.Sprintf("%v", description)
					}
					if strength, ok := rel.Props["strength"]; ok {
						graphRelation.Strength = convertToFloat64(strength)
					}
					if weight, ok := rel.Props["weight"]; ok {
						graphRelation.Weight = convertToFloat64(weight)
					}
					if combinedDegree, ok := rel.Props["combined_degree"]; ok {
						graphRelation.CombinedDegree = int(convertToFloat64(combinedDegree))
					}
					if chunkIDs, ok := rel.Props["chunk_ids"]; ok {
						graphRelation.ChunkIDs = interfaceListToStringList(chunkIDs.([]interface{}))
					}

					graphData.Relation = append(graphData.Relation, graphRelation)
				}
			}
		}

		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("error iterating results: %w", err)
		}

		return graphData, nil
	})

	if err != nil {
		log.Printf("[Neo4j] ERROR: search node failed: %v", err)
		return nil, err
	}

	log.Printf("[Neo4j] Search completed: found %d nodes, %d relations",
		len(result.(*types.GraphData).Node), len(result.(*types.GraphData).Relation))
	return result.(*types.GraphData), nil
}

// SearchPath 搜索路径
func (n *Neo4jRepository) SearchPath(
	ctx context.Context,
	namespace types.NameSpace,
	startNode, endNode string,
	maxDepth int,
) ([]*types.GraphData, error) {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT RETRIEVE GRAPH - driver is nil")
		return nil, nil
	}

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		labelExpr := n.Label(namespace)
		query := `
			MATCH path = shortestPath(
				(start:` + labelExpr + ` {knowledge_id: $knowledge_id})-[*1..` + fmt.Sprint(maxDepth) + `]-(end:` + labelExpr + ` {knowledge_id: $knowledge_id})
			)
			WHERE (start.title = $start_node OR start.name = $start_node)
			AND (end.title = $end_node OR end.name = $end_node)
			RETURN path
			LIMIT 10
		`

		params := map[string]interface{}{
			"start_node":   startNode,
			"end_node":     endNode,
			"knowledge_id": namespace.Knowledge,
		}

		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, fmt.Errorf("failed to run path query: %w", err)
		}

		paths := make([]*types.GraphData, 0)

		for result.Next(ctx) {
			record := result.Record()
			pathVal, hasPath := record.Get("path")
			if !hasPath {
				continue
			}

			path, ok := pathVal.(neo4j.Path)
			if !ok {
				continue
			}

			graphData := &types.GraphData{
				Node:     make([]*types.GraphNode, 0),
				Relation: make([]*types.GraphRelation, 0),
			}
			nodeSeen := make(map[string]bool)
			relSeen := make(map[string]bool)

			// 处理路径中的节点
			for _, node := range path.Nodes {
				idVal, ok := node.Props["id"]
				if !ok {
					continue
				}
				idStr := fmt.Sprintf("%v", idVal)

				if _, seen := nodeSeen[idStr]; !seen {
					nodeSeen[idStr] = true
					graphNode := &types.GraphNode{
						ID:         idStr,
						Attributes: []string{},
						Chunks:     []string{},
					}

					if name, ok := node.Props["name"]; ok {
						graphNode.Name = fmt.Sprintf("%v", name)
					}
					if title, ok := node.Props["title"]; ok {
						graphNode.Title = fmt.Sprintf("%v", title)
					} else {
						graphNode.Title = graphNode.Name
					}
					if attrs, ok := node.Props["attributes"]; ok {
						graphNode.Attributes = interfaceListToStringList(attrs.([]interface{}))
					}
					if chunks, ok := node.Props["chunks"]; ok {
						graphNode.Chunks = interfaceListToStringList(chunks.([]interface{}))
					}

					graphData.Node = append(graphData.Node, graphNode)
				}
			}

			// 处理路径中的关系
			for _, rel := range path.Relationships {
				relIDVal, ok := rel.Props["id"]
				if !ok {
					continue
				}
				relID := fmt.Sprintf("%v", relIDVal)

				if _, seen := relSeen[relID]; !seen {
					relSeen[relID] = true

					// 获取源节点和目标节点
					sourceIdx := -1
					targetIdx := -1
					for i, node := range path.Nodes {
						if node.ElementId == rel.StartElementId {
							sourceIdx = i
						}
						if node.ElementId == rel.EndElementId {
							targetIdx = i
						}
					}

					if sourceIdx == -1 || targetIdx == -1 {
						continue
					}

					sourceNode := path.Nodes[sourceIdx]
					targetNode := path.Nodes[targetIdx]

					graphRelation := &types.GraphRelation{
						ID:       relID,
						ChunkIDs: []string{},
					}

					// 从节点获取标题
					if sourceTitle, ok := sourceNode.Props["title"]; ok {
						graphRelation.Source = fmt.Sprintf("%v", sourceTitle)
					} else if sourceName, ok := sourceNode.Props["name"]; ok {
						graphRelation.Source = fmt.Sprintf("%v", sourceName)
					}

					if targetTitle, ok := targetNode.Props["title"]; ok {
						graphRelation.Target = fmt.Sprintf("%v", targetTitle)
					} else if targetName, ok := targetNode.Props["name"]; ok {
						graphRelation.Target = fmt.Sprintf("%v", targetName)
					}

					if description, ok := rel.Props["description"]; ok {
						graphRelation.Description = fmt.Sprintf("%v", description)
					}
					if strength, ok := rel.Props["strength"]; ok {
						graphRelation.Strength = convertToFloat64(strength)
					}
					if weight, ok := rel.Props["weight"]; ok {
						graphRelation.Weight = convertToFloat64(weight)
					}
					if combinedDegree, ok := rel.Props["combined_degree"]; ok {
						graphRelation.CombinedDegree = int(convertToFloat64(combinedDegree))
					}
					if chunkIDs, ok := rel.Props["chunk_ids"]; ok {
						graphRelation.ChunkIDs = interfaceListToStringList(chunkIDs.([]interface{}))
					}

					graphData.Relation = append(graphData.Relation, graphRelation)
				}
			}

			paths = append(paths, graphData)
		}

		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("error iterating path results: %w", err)
		}

		return paths, nil
	})

	if err != nil {
		log.Printf("[Neo4j] ERROR: search path failed: %v", err)
		return nil, err
	}

	log.Printf("[Neo4j] Path search completed: found %d paths", len(result.([]*types.GraphData)))
	return result.([]*types.GraphData), nil
}

// CheckHealth 检查 Neo4j 连接健康状态
func (n *Neo4jRepository) CheckHealth(ctx context.Context) error {
	if n.driver == nil {
		return fmt.Errorf("neo4j driver is nil")
	}

	// 验证连接
	err := n.driver.VerifyConnectivity(ctx)
	if err != nil {
		return fmt.Errorf("neo4j connectivity check failed: %w", err)
	}

	// 尝试执行简单查询
	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	_, err = session.Run(ctx, "RETURN 1 AS result", nil)
	if err != nil {
		return fmt.Errorf("neo4j query failed: %w", err)
	}

	log.Printf("[Neo4j] Health check passed")
	return nil
}

// ========================================
// 辅助函数
// ========================================

// interfaceListToStringList 将 []interface{} 转换为 []string
func interfaceListToStringList(list []interface{}) []string {
	result := make([]string, 0, len(list))
	for _, v := range list {
		result = append(result, fmt.Sprintf("%v", v))
	}
	return result
}

// convertToFloat64 将各种数值类型转换为 float64
func convertToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return 0.0
	}
}
