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
func NewNeo4jRepository(driver neo4j.DriverWithContext) interfaces.Neo4jGraphRepository {
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
	log.Printf("[Repo] addGraph START: namespace.KBID=%s, nodes=%d, relations=%d",
		namespace.KBID, len(graph.Node), len(graph.Relation))
	log.Printf("[Repo] addGraph Label: %s", n.Label(namespace))

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
			SET n.kb_id = $kb_id
			SET n.chunks = COALESCE(n.chunks, [])
			SET n.attributes = COALESCE(row.attributes, [])
			RETURN n
		`

		nodeData := make([]map[string]interface{}, 0)
		for _, node := range graph.Node {
			nodeData = append(nodeData, map[string]interface{}{
				"id":          node.ID,
				"name":        node.Name,
				"entity_type": node.EntityType,
				"attributes":  node.Attributes,
				"chunks":      node.Chunks,
			})
			log.Printf("[Repo] addGraph Node: ID=%s, Name=%q", node.ID, node.Name)
		}

		if len(nodeData) > 0 {
			log.Printf("[Repo] addGraph Executing node query with %d nodes", len(nodeData))
			result, err := tx.Run(ctx, nodeQuery, map[string]interface{}{
				"data":         nodeData,
				"knowledge_id": namespace.Knowledge,
				"kb_id":        namespace.KBID,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create nodes: %w", err)
			}
			// 消费结果并打印返回值
			count := 0
			for result.Next(ctx) {
				count++
			}
			if _, err := result.Consume(ctx); err != nil {
				return nil, fmt.Errorf("failed to consume node result: %w", err)
			}
			log.Printf("[Repo] addGraph Nodes created successfully: count=%d", count)
		}

		// 2. 创建关系（使用新的关系结构）
		// 先查找节点的度数，用于计算 CombinedDegree
		// 注意：优先使用 source_id/target_id（节点ID），如果为空则使用 source/target（节点名称）
		// 如果节点不存在，使用 MERGE 自动创建
		relQuery := `
			UNWIND $data AS row
			// 使用 MERGE 匹配或创建源节点：优先用 source_id，否则用 source 名称
			MERGE (source:` + n.Label(namespace) + ` {id: COALESCE(row.source_id, row.source)})
			ON CREATE SET
				source.id = COALESCE(row.source_id, row.source),
				source.name = row.source,
				source.kb_id = $kb_id,
				source.knowledge_id = $knowledge_id
			// 使用 MERGE 匹配或创建目标节点：使用 row.target 匹配
			MERGE (target:` + n.Label(namespace) + ` {name: row.target, kb_id: $kb_id})
			ON CREATE SET
				target.id = COALESCE(row.target_id, row.target),
				target.name = row.target,
				target.kb_id = $kb_id,
				target.knowledge_id = $knowledge_id

			// 获取源节点和目标节点的度数
			WITH source, target, row
			OPTIONAL MATCH (source)-[outRel]-()
			WITH source, target, row, count(outRel) AS sourceDegree
			OPTIONAL MATCH (target)-[inRel]-()
			WITH source, target, row, sourceDegree, count(inRel) AS targetDegree

			// 计算 CombinedDegree
			WITH source, target, row, sourceDegree + targetDegree AS combinedDegree

			// 使用关系ID创建唯一关系
			MERGE (source)-[r:RELATES_TO {id: row.id, kb_id: $kb_id}]->(target)
			SET r.source = row.source
			SET r.target = row.target
			SET r.type = COALESCE(row.type, '关联')
			SET r.description = row.description
			SET r.strength = COALESCE(row.strength, 5.0)
			SET r.weight = COALESCE(row.weight, 5.0)
			SET r.combined_degree = combinedDegree
			SET r.kb_id = $kb_id

			// 更新 ChunkIDs - 如果有 chunk_id 且不在列表中，则添加
			WITH r, row
			WHERE row.chunk_id IS NOT NULL
			SET r.chunk_ids = CASE
				WHEN row.chunk_id IN r.chunk_ids THEN r.chunk_ids
				ELSE COALESCE(r.chunk_ids, []) + row.chunk_id
			END

			RETURN r.id, r.source, r.target, r.type, r.description, r.strength
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
						"type":        rel.Type,
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
					"type":        rel.Type,
					"description": rel.Description,
					"strength":    rel.Strength,
					"weight":      rel.Weight,
					"chunk_id":    nil,
				})
			}
			log.Printf("[Repo] addGraph Relation: ID=%s, Source=%q, Target=%q, Type=%q, source_id=%q, target_id=%q",
				rel.ID, rel.Source, rel.Target, rel.Type,
				n.findNodeIDByName(graph.Node, rel.Source),
				n.findNodeIDByName(graph.Node, rel.Target))
		}

		if len(relData) > 0 {
			log.Printf("[Repo] addGraph Executing relation query with %d relations", len(relData))
			log.Printf("[Repo] addGraph relData[0]: id=%s, source=%q, target=%q, source_id=%q, target_id=%q",
				relData[0]["id"], relData[0]["source"], relData[0]["target"],
				relData[0]["source_id"], relData[0]["target_id"])

			result, err := tx.Run(ctx, relQuery, map[string]interface{}{
				"data":         relData,
				"knowledge_id": namespace.Knowledge,
				"kb_id":        namespace.KBID,
			})
			if err != nil {
				log.Printf("[Repo] addGraph Run ERROR: %v", err)
				return nil, fmt.Errorf("failed to create relationships: %w", err)
			}

			// 消费结果并打印返回值
			count := 0
			for result.Next(ctx) {
				count++
				record := result.Record()
				id, _ := record.Get("r.id")
				source, _ := record.Get("r.source")
				target, _ := record.Get("r.target")
				relType, _ := record.Get("r.type")
				log.Printf("[Repo] addGraph RETURN: r.id=%v, r.source=%v, r.target=%v, r.type=%v",
					id, source, target, relType)
			}
			if _, err := result.Consume(ctx); err != nil {
				return nil, fmt.Errorf("failed to consume relation result: %w", err)
			}
			log.Printf("[Repo] addGraph Relations created: count=%d", count)
			if count == 0 {
				log.Printf("[Repo] WARNING: Neo4j returned 0 relations! Nodes may not exist.")
			}
		}

		return nil, nil
	})

	if err != nil {
		log.Printf("[Repo] addGraph ERROR: failed to add graph: %v", err)
		return err
	}

	log.Printf("[Repo] addGraph SUCCESS: knowledge_id=%s, nodes=%d, relations=%d",
		namespace.Knowledge, len(graph.Node), len(graph.Relation))
	return nil
}

// findNodeIDByName 根据实体名称查找节点ID
func (n *Neo4jRepository) findNodeIDByName(nodes []*types.GraphNode, name string) string {
	for _, node := range nodes {
		if node.Name == name {
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
			// 不使用 label 过滤，只使用 kb_id 和 knowledge_id 属性（兼容不同命名风格的节点）
			var deleteRelsQuery, deleteNodesQuery string
			var params map[string]interface{}

			if namespace.Knowledge != "" {
				// 按 knowledge_id 删除
				deleteRelsQuery = `
					MATCH (n {knowledge_id: $knowledge_id})-[r]-(m {knowledge_id: $knowledge_id})
					DELETE r
					RETURN count(r) AS deleted_count
				`
				deleteNodesQuery = `
					MATCH (n {knowledge_id: $knowledge_id})
					DELETE n
					RETURN count(n) AS deleted_count
				`
				params = map[string]interface{}{
					"knowledge_id": namespace.Knowledge,
				}
				log.Printf("[Neo4j] DeleteGraph: deleting by knowledge_id=%s", namespace.Knowledge)
			} else {
				// 按 kb_id 删除
				deleteRelsQuery = `
					MATCH (n {kb_id: $kb_id})-[r]-(m {kb_id: $kb_id})
					DELETE r
					RETURN count(r) AS deleted_count
				`
				deleteNodesQuery = `
					MATCH (n {kb_id: $kb_id})
					DELETE n
					RETURN count(n) AS deleted_count
				`
				params = map[string]interface{}{
					"kb_id": namespace.KBID,
				}
				log.Printf("[Neo4j] DeleteGraph: deleting by kb_id=%s", namespace.KBID)
			}

			result, err := tx.Run(ctx, deleteRelsQuery, params)
			if err != nil {
				return nil, fmt.Errorf("failed to delete relationships: %w", err)
			}
			if _, err := result.Consume(ctx); err != nil {
				return nil, fmt.Errorf("failed to consume delete rels result: %w", err)
			}

			result, err = tx.Run(ctx, deleteNodesQuery, params)
			if err != nil {
				return nil, fmt.Errorf("failed to delete nodes: %w", err)
			}
			if _, err := result.Consume(ctx); err != nil {
				return nil, fmt.Errorf("failed to consume delete nodes result: %w", err)
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

// GetGraph 获取知识库的完整图谱数据
func (n *Neo4jRepository) GetGraph(
	ctx context.Context,
	namespace types.NameSpace,
) (*types.GraphData, error) {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT RETRIEVE GRAPH - driver is nil")
		return &types.GraphData{
			Node:     []*types.GraphNode{},
			Relation: []*types.GraphRelation{},
		}, nil
	}

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		log.Printf("[Neo4j] GetGraph: namespace.Knowledge=%s, namespace.KBID=%s, namespace.Type=%s",
			namespace.Knowledge, namespace.KBID, namespace.Type)

		// 构建节点查询
		var nodeQuery string

		if namespace.Knowledge != "" {
			// 查询单个知识条目的图谱（使用 knowledge_id）
			nodeQuery = `
				MATCH (n {knowledge_id: $knowledge_id, kb_id: $kb_id})
				RETURN n.id, n.name, n.entity_type
				ORDER BY n.name
				LIMIT 1000
			`
			log.Printf("[Neo4j] Querying single knowledge graph: knowledge_id=%s", namespace.Knowledge)
		} else {
			// 查询整个知识库的图谱（使用 kb_id 属性）
			nodeQuery = `
				MATCH (n {kb_id: $kb_id})
				RETURN n.id, n.name, n.entity_type
				ORDER BY n.name
				LIMIT 1000
			`
			log.Printf("[Neo4j] Querying entire KB graph with kb_id: %s", namespace.KBID)
		}

		result, err := tx.Run(ctx, nodeQuery, map[string]interface{}{
			"knowledge_id": namespace.Knowledge,
			"kb_id":        namespace.KBID,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to query nodes: %w", err)
		}

		// 收集所有节点ID
		nodeIDs := []string{}
		nodesMap := make(map[string]*types.GraphNode)

		for result.Next(ctx) {
			record := result.Record()
			idVal, ok := record.Get("n.id")
			if !ok || idVal == nil {
				continue
			}

			node := &types.GraphNode{
				ID:         fmt.Sprintf("%v", idVal),
				Attributes: []string{},
				Chunks:     []string{},
			}

			if name, ok := record.Get("n.name"); ok && name != nil {
				node.Name = fmt.Sprintf("%v", name)
			}

			if entityType, ok := record.Get("n.entity_type"); ok && entityType != nil {
				node.EntityType = fmt.Sprintf("%v", entityType)
			}

			nodesMap[node.ID] = node
			nodeIDs = append(nodeIDs, node.ID)
		}

		// 如果没有节点，返回空
		if len(nodeIDs) == 0 {
			return &types.GraphData{
				Node:     nodesMapToList(nodesMap),
				Relation: []*types.GraphRelation{},
			}, nil
		}

		// 获取节点之间的关系 - 使用 COALESCE 处理可能的空值
		var relQuery string
		if namespace.Knowledge != "" {
			// 查询单个知识条目的关系
			relQuery = `
				MATCH (n {kb_id: $kb_id})-[r:RELATES_TO]->(m {kb_id: $kb_id})
				WHERE n.id IN $node_ids
				RETURN n.id as source_id, n.name as source_name,
				       r.id as rel_id, COALESCE(r.type, '关联') as type, r.description as description,
				       COALESCE(r.strength, 5.0) as strength, COALESCE(r.weight, 5.0) as weight,
				       m.id as target_id, m.name as target_name
				LIMIT 2000
			`
		} else {
			// 查询整个知识库的关系
			relQuery = `
				MATCH (n {kb_id: $kb_id})-[r:RELATES_TO]->(m {kb_id: $kb_id})
				WHERE n.id IN $node_ids
				RETURN n.id as source_id, n.name as source_name,
				       r.id as rel_id, COALESCE(r.type, '关联') as type, r.description as description,
				       COALESCE(r.strength, 5.0) as strength, COALESCE(r.weight, 5.0) as weight,
				       m.id as target_id, m.name as target_name
				LIMIT 2000
			`
		}

		result, err = tx.Run(ctx, relQuery, map[string]interface{}{
			"kb_id":    namespace.KBID,
			"node_ids": nodeIDs,
		})

		if err != nil {
			return nil, fmt.Errorf("failed to query relations: %w", err)
		}

		graphData := &types.GraphData{
			Node:     nodesMapToList(nodesMap),
			Relation: make([]*types.GraphRelation, 0),
		}
		relSeen := make(map[string]bool)

		for result.Next(ctx) {
			record := result.Record()

			_, hasSource := record.Get("source_id")
			_, hasTarget := record.Get("target_id")
			if !hasSource || !hasTarget {
				continue
			}

			relIDVal, ok := record.Get("rel_id")
			if !ok || relIDVal == nil {
				continue
			}

			relID := fmt.Sprintf("%v", relIDVal)

			if _, seen := relSeen[relID]; !seen {
				relSeen[relID] = true

				graphRelation := &types.GraphRelation{
					ID:       relID,
					ChunkIDs: []string{},
					Type:     "关联", // 默认关系类型
					Strength: 5.0,  // 默认强度
					Weight:   5.0,  // 默认权重
				}

				// 获取属性（需要 nil 检查）
				if sourceName, ok := record.Get("source_name"); ok && sourceName != nil {
					graphRelation.Source = fmt.Sprintf("%v", sourceName)
				}
				if targetName, ok := record.Get("target_name"); ok && targetName != nil {
					graphRelation.Target = fmt.Sprintf("%v", targetName)
				}
				if relType, ok := record.Get("type"); ok && relType != nil {
					graphRelation.Type = fmt.Sprintf("%v", relType)
				}
				if description, ok := record.Get("description"); ok && description != nil {
					graphRelation.Description = fmt.Sprintf("%v", description)
				}
				if strength, ok := record.Get("strength"); ok && strength != nil {
					graphRelation.Strength = convertToFloat64(strength)
				}
				if weight, ok := record.Get("weight"); ok && weight != nil {
					graphRelation.Weight = convertToFloat64(weight)
				}

				log.Printf("[Neo4j] Retrieved relation: ID=%s, Source=%s, Target=%s, Type=%s, Strength=%f, Weight=%f",
					graphRelation.ID, graphRelation.Source, graphRelation.Target, graphRelation.Type, graphRelation.Strength, graphRelation.Weight)

				graphData.Relation = append(graphData.Relation, graphRelation)
			}
		}

		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("error iterating results: %w", err)
		}

		return graphData, nil
	})

	if err != nil {
		log.Printf("[Neo4j] ERROR: get graph failed: %v", err)
		return nil, err
	}

	graphData, ok := result.(*types.GraphData)
	if !ok {
		log.Printf("[Neo4j] ERROR: unexpected result type from ExecuteRead")
		return &types.GraphData{
			Node:     []*types.GraphNode{},
			Relation: []*types.GraphRelation{},
		}, nil
	}

	log.Printf("[Neo4j] Get graph completed: found %d nodes, %d relations",
		len(graphData.Node), len(graphData.Relation))
	return graphData, nil
}

// 辅助函数：将 map 转换为列表
func nodesMapToList(m map[string]*types.GraphNode) []*types.GraphNode {
	list := make([]*types.GraphNode, 0, len(m))
	for _, node := range m {
		list = append(list, node)
	}
	return list
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
		// 不使用 label 过滤，只使用 kb_id 属性（兼容不同命名风格的节点）
		// 根据是否有 knowledge_id 决定使用哪个字段查询
		var query string
		var knowledgeIDCondition string
		if namespace.Knowledge != "" {
			// 查询单个知识条目的图谱
			knowledgeIDCondition = " AND n.knowledge_id = $knowledge_id AND m.knowledge_id = $knowledge_id"
			log.Printf("[Neo4j] SearchNode: querying single knowledge graph: knowledge_id=%s", namespace.Knowledge)
		} else {
			// 查询整个知识库的图谱（使用 kb_id 属性）
			knowledgeIDCondition = ""
			log.Printf("[Neo4j] SearchNode: querying entire KB graph with kb_id: %s", namespace.KBID)
		}

		query = `
			MATCH (n {kb_id: $kb_id})-[r:RELATES_TO]->(m {kb_id: $kb_id})
			WHERE ANY(nodeText IN $nodes WHERE toLower(n.name) CONTAINS toLower(nodeText))` + knowledgeIDCondition + `
			RETURN n, r, m
			LIMIT 1000
		`

		params := map[string]interface{}{
			"nodes": nodes,
			"kb_id": namespace.KBID,
		}
		if namespace.Knowledge != "" {
			params["knowledge_id"] = namespace.Knowledge
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
					// 处理 type 字段，提供默认值
					if relType, ok := rel.Props["type"]; ok && relType != nil {
						graphRelation.Type = fmt.Sprintf("%v", relType)
					} else {
						graphRelation.Type = "关联"
					}
					if description, ok := rel.Props["description"]; ok {
						graphRelation.Description = fmt.Sprintf("%v", description)
					}
					if strength, ok := rel.Props["strength"]; ok {
						graphRelation.Strength = convertToFloat64(strength)
					} else {
						graphRelation.Strength = 5.0
					}
					if weight, ok := rel.Props["weight"]; ok {
						graphRelation.Weight = convertToFloat64(weight)
					} else {
						graphRelation.Weight = 5.0
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

// SearchNodeV2 搜索节点（改进版 - 直接返回节点名称）
func (n *Neo4jRepository) SearchNodeV2(
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
		var knowledgeIDCondition string
		if namespace.Knowledge != "" {
			knowledgeIDCondition = " AND n.knowledge_id = $knowledge_id AND m.knowledge_id = $knowledge_id"
			log.Printf("[Neo4j] SearchNodeV2: querying single knowledge graph: knowledge_id=%s", namespace.Knowledge)
		} else {
			knowledgeIDCondition = ""
			log.Printf("[Neo4j] SearchNodeV2: querying entire KB graph with kb_id: %s", namespace.KBID)
		}

		// 改进版查询：直接返回节点名称
		query := `
			MATCH (n {kb_id: $kb_id})-[r:RELATES_TO]->(m {kb_id: $kb_id})
			WHERE ANY(nodeText IN $nodes WHERE toLower(n.name) CONTAINS toLower(nodeText))` + knowledgeIDCondition + `
			RETURN n.id as n_id, n.name as n_name, n.entity_type as n_type,
			       n.chunks as n_chunks, n.attributes as n_attrs,
			       r.id as r_id, r.type as r_type, r.description as r_desc,
			       r.strength as r_strength, r.weight as r_weight, r.chunk_ids as r_chunk_ids,
			       m.id as m_id, m.name as m_name, m.entity_type as m_type
			LIMIT 1000
		`

		params := map[string]interface{}{
			"nodes": nodes,
			"kb_id": namespace.KBID,
		}
		if namespace.Knowledge != "" {
			params["knowledge_id"] = namespace.Knowledge
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

			// 获取节点数据 - 直接从返回的字段读取
			nIDVal, _ := record.Get("n_id")
			nNameVal, _ := record.Get("n_name")
			nTypeVal, _ := record.Get("n_type")
			mIDVal, _ := record.Get("m_id")
			mNameVal, _ := record.Get("m_name")
			mTypeVal, _ := record.Get("m_type")

			if nIDVal == nil || mIDVal == nil {
				continue
			}

			// 处理源节点 n
			nID := fmt.Sprintf("%v", nIDVal)
			if _, seen := nodeSeen[nID]; !seen {
				nodeSeen[nID] = true
				graphNode := &types.GraphNode{
					ID:         nID,
					Attributes: []string{},
					Chunks:     []string{},
				}
				if nNameVal != nil {
					graphNode.Name = fmt.Sprintf("%v", nNameVal)
				}
				if nTypeVal != nil {
					graphNode.EntityType = fmt.Sprintf("%v", nTypeVal)
				}
				if nChunks, ok := record.Get("n_chunks"); ok && nChunks != nil {
					graphNode.Chunks = interfaceListToStringList(nChunks.([]interface{}))
				}
				graphData.Node = append(graphData.Node, graphNode)
			}

			// 处理目标节点 m
			mID := fmt.Sprintf("%v", mIDVal)
			if _, seen := nodeSeen[mID]; !seen {
				nodeSeen[mID] = true
				graphNode := &types.GraphNode{
					ID:         mID,
					Attributes: []string{},
					Chunks:     []string{},
				}
				if mNameVal != nil {
					graphNode.Name = fmt.Sprintf("%v", mNameVal)
				}
				if mTypeVal != nil {
					graphNode.EntityType = fmt.Sprintf("%v", mTypeVal)
				}
				graphData.Node = append(graphData.Node, graphNode)
			}

			// 处理关系 - 使用节点名称
			rIDVal, hasR := record.Get("r_id")
			if !hasR {
				continue
			}
			relID := fmt.Sprintf("%v", rIDVal)

			if _, seen := relSeen[relID]; !seen {
				relSeen[relID] = true

				graphRelation := &types.GraphRelation{
					ID:       relID,
					ChunkIDs: []string{},
					Type:     "关联",
					Strength: 5.0,
					Weight:   5.0,
				}

				// 直接使用节点名称
				if nNameVal != nil {
					graphRelation.Source = fmt.Sprintf("%v", nNameVal)
				}
				if mNameVal != nil {
					graphRelation.Target = fmt.Sprintf("%v", mNameVal)
				}
				if rType, ok := record.Get("r_type"); ok && rType != nil {
					graphRelation.Type = fmt.Sprintf("%v", rType)
				}
				if rDesc, ok := record.Get("r_desc"); ok && rDesc != nil {
					graphRelation.Description = fmt.Sprintf("%v", rDesc)
				}
				if rStrength, ok := record.Get("r_strength"); ok && rStrength != nil {
					graphRelation.Strength = convertToFloat64(rStrength)
				}
				if rWeight, ok := record.Get("r_weight"); ok && rWeight != nil {
					graphRelation.Weight = convertToFloat64(rWeight)
				}
				if rChunkIDs, ok := record.Get("r_chunk_ids"); ok && rChunkIDs != nil {
					graphRelation.ChunkIDs = interfaceListToStringList(rChunkIDs.([]interface{}))
				}

				graphData.Relation = append(graphData.Relation, graphRelation)
			}
		}

		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("error iterating results: %w", err)
		}

		return graphData, nil
	})

	if err != nil {
		log.Printf("[Neo4j] ERROR: search node V2 failed: %v", err)
		return nil, err
	}

	log.Printf("[Neo4j] SearchV2 completed: found %d nodes, %d relations",
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
		// 不使用 label 过滤，只使用 kb_id 属性（兼容不同命名风格的节点）
		// 根据是否有 knowledge_id 决定查询条件
		var query string
		if namespace.Knowledge != "" {
			// 查询单个知识条目的图谱
			query = `
				MATCH path = shortestPath(
					(start {kb_id: $kb_id, knowledge_id: $knowledge_id})-[*1..` + fmt.Sprint(maxDepth) + `]-(end {kb_id: $kb_id, knowledge_id: $knowledge_id})
				)
				WHERE start.name = $start_node
				AND end.name = $end_node
				RETURN path
				LIMIT 10
			`
			log.Printf("[Neo4j] SearchPath: querying single knowledge graph: knowledge_id=%s", namespace.Knowledge)
		} else {
			// 查询整个知识库的图谱（使用 kb_id 属性）
			query = `
				MATCH path = shortestPath(
					(start {kb_id: $kb_id})-[*1..` + fmt.Sprint(maxDepth) + `]-(end {kb_id: $kb_id})
				)
				WHERE start.name = $start_node
				AND end.name = $end_node
				RETURN path
				LIMIT 10
			`
			log.Printf("[Neo4j] SearchPath: querying entire KB graph with kb_id: %s", namespace.KBID)
		}

		params := map[string]interface{}{
			"start_node": startNode,
			"end_node":   endNode,
			"kb_id":      namespace.KBID,
		}
		if namespace.Knowledge != "" {
			params["knowledge_id"] = namespace.Knowledge
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

					// 从节点获取名称
					if sourceName, ok := sourceNode.Props["name"]; ok {
						graphRelation.Source = fmt.Sprintf("%v", sourceName)
					}

					if targetName, ok := targetNode.Props["name"]; ok {
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

// UpdateNode 更新节点属性
func (n *Neo4jRepository) UpdateNode(ctx context.Context, namespace types.NameSpace, node *types.GraphNode) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT RETRIEVE GRAPH - driver is nil")
		return nil
	}

	log.Printf("[Neo4j] UpdateNode: id=%s, name=%s, entity_type=%s", node.ID, node.Name, node.EntityType)

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 不使用 label 过滤，直接用 id 匹配（id 是唯一的）
		nodeQuery := `
			MATCH (n {id: $id})
			SET n.name = $name
			SET n.entity_type = $entity_type
			SET n.attributes = $attributes
			RETURN n
		`

		result, err := tx.Run(ctx, nodeQuery, map[string]interface{}{
			"id":          node.ID,
			"name":        node.Name,
			"entity_type": node.EntityType,
			"attributes":  node.Attributes,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update node: %w", err)
		}

		// 检查是否匹配到节点并打印返回数据
		if result.Next(ctx) {
			record := result.Record()
			if name, ok := record.Get("n.name"); ok && name != nil {
				log.Printf("[Neo4j] UpdateNode SUCCESS: id=%s, returned name=%q", node.ID, name)
			} else {
				log.Printf("[Neo4j] UpdateNode SUCCESS: id=%s (no name returned)", node.ID)
			}
		} else {
			log.Printf("[Neo4j] UpdateNode WARNING: no node found with id=%s", node.ID)
		}

		// 消费结果
		if _, err := result.Consume(ctx); err != nil {
			return nil, fmt.Errorf("failed to consume result: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		log.Printf("[Neo4j] ERROR: failed to update node: %v", err)
		return err
	}

	return nil
}

// AddRelation 添加单个关系
// 返回添加后的关系
func (n *Neo4jRepository) AddRelation(ctx context.Context, namespace types.NameSpace, relation *types.GraphRelation) (*types.GraphRelation, error) {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT RETRIEVE GRAPH - driver is nil")
		return nil, nil
	}

	log.Printf("[Neo4j] AddRelation START: id=%s, kb_id=%s, Source=%q, Target=%q, Type=%q, Description=%q, Strength=%f",
		relation.ID, namespace.KBID, relation.Source, relation.Target, relation.Type, relation.Description, relation.Strength)

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("[Neo4j] AddRelation session.Close error: %v", err)
		}
	}()

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 使用 MERGE 匹配或创建源节点：使用 source 名称作为 ID
		mergeQuery := `
			// 使用 MERGE 匹配或创建源节点
			MERGE (source:` + n.Label(namespace) + ` {name: $source})
			ON CREATE SET
				source.id = COALESCE($source_id, source.name),
				source.name = $source,
				source.kb_id = $kb_id,
				source.knowledge_id = $knowledge_id
			// 使用 MERGE 匹配或创建目标节点
			WITH source
			MERGE (target:` + n.Label(namespace) + ` {name: $target})
			ON CREATE SET
				target.id = COALESCE($target_id, target.name),
				target.name = $target,
				target.kb_id = $kb_id,
				target.knowledge_id = $knowledge_id

			// 获取源节点和目标节点的度数
			WITH source, target
			OPTIONAL MATCH (source)-[outRel]-()
			WITH source, target, count(outRel) AS sourceDegree
			OPTIONAL MATCH (target)-[inRel]-()
			WITH source, target, sourceDegree, count(inRel) AS targetDegree

			// 计算 CombinedDegree
			WITH source, target, sourceDegree + targetDegree AS combinedDegree

			// 使用关系ID创建唯一关系（只通过 id 匹配，兼容旧数据）
			MERGE (source)-[r:RELATES_TO {id: $id}]->(target)
			ON CREATE SET
				r.id = $id,
				r.source = $source,
				r.target = $target,
				r.type = COALESCE($type, '关联'),
				r.description = $description,
				r.strength = COALESCE($strength, 5.0),
				r.weight = COALESCE($weight, 5.0),
				r.combined_degree = combinedDegree,
				r.kb_id = $kb_id
			ON MATCH SET
				r.type = COALESCE($type, '关联'),
				r.description = $description,
				r.strength = COALESCE($strength, 5.0),
				r.weight = COALESCE($weight, 5.0),
				r.kb_id = $kb_id
			RETURN r.id as rel_id, r.source as source, r.target as target, r.type as type, r.description as description, r.strength as strength, r.weight as weight, r.combined_degree as combined_degree
		`

		params := map[string]interface{}{
			"id":           relation.ID,
			"source_id":    relation.ID + "_source", // 临时 ID
			"source":       relation.Source,
			"target_id":    relation.ID + "_target", // 临时 ID
			"target":       relation.Target,
			"type":         relation.Type,
			"description":  relation.Description,
			"strength":     relation.Strength,
			"weight":       relation.Weight,
			"kb_id":        namespace.KBID,
			"knowledge_id": namespace.Knowledge,
		}
		log.Printf("[Neo4j] AddRelation params: id=%s, source=%q, target=%q, type=%q, strength=%f, weight=%f",
			params["id"], params["source"], params["target"], params["type"], params["strength"], params["weight"])

		result, err := tx.Run(ctx, mergeQuery, params)
		if err != nil {
			log.Printf("[Neo4j] AddRelation Run ERROR: %v", err)
			return nil, fmt.Errorf("failed to create relation: %w", err)
		}

		// 收集创建的关系数据
		var createdRel *types.GraphRelation
		if result.Next(ctx) {
			record := result.Record()
			createdRel = &types.GraphRelation{
				ID:       relation.ID,
				ChunkIDs: []string{},
			}

			if source, ok := record.Get("source"); ok && source != nil {
				createdRel.Source = fmt.Sprintf("%v", source)
			}
			if target, ok := record.Get("target"); ok && target != nil {
				createdRel.Target = fmt.Sprintf("%v", target)
			}
			if relType, ok := record.Get("type"); ok && relType != nil {
				createdRel.Type = fmt.Sprintf("%v", relType)
			}
			if relDesc, ok := record.Get("description"); ok && relDesc != nil {
				createdRel.Description = fmt.Sprintf("%v", relDesc)
			}
			if relStrength, ok := record.Get("strength"); ok && relStrength != nil {
				createdRel.Strength = convertToFloat64(relStrength)
			}
			if relWeight, ok := record.Get("weight"); ok && relWeight != nil {
				createdRel.Weight = convertToFloat64(relWeight)
			}
			if combinedDegree, ok := record.Get("combined_degree"); ok && combinedDegree != nil {
				createdRel.CombinedDegree = int(convertToFloat64(combinedDegree))
			}

			log.Printf("[Neo4j] AddRelation SUCCESS: id=%s, Source=%q, Target=%q, Type=%q, Strength=%f, Weight=%f",
				createdRel.ID, createdRel.Source, createdRel.Target, createdRel.Type, createdRel.Strength, createdRel.Weight)
		} else {
			log.Printf("[Neo4j] AddRelation WARNING: no relation returned after create")
		}

		// 消费剩余结果并处理错误
		if _, err := result.Consume(ctx); err != nil {
			return nil, fmt.Errorf("failed to consume result: %w", err)
		}

		return createdRel, nil
	})

	if err != nil {
		log.Printf("[Neo4j] AddRelation ERROR: failed to add relation: %v", err)
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return result.(*types.GraphRelation), nil
}

// AddNode 添加单个节点
func (n *Neo4jRepository) AddNode(ctx context.Context, namespace types.NameSpace, node *types.GraphNode) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT RETRIEVE GRAPH - driver is nil")
		return nil
	}

	log.Printf("[Neo4j] AddNode START: id=%s, kb_id=%s, Name=%q, EntityType=%q",
		node.ID, namespace.KBID, node.Name, node.EntityType)

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("[Neo4j] AddNode session.Close error: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		nodeQuery := `
			MERGE (n:` + n.Label(namespace) + ` {id: $id})
			SET n.name = $name
			SET n.title = COALESCE($title, $name)
			SET n.entity_type = $entity_type
			SET n.knowledge_id = $knowledge_id
			SET n.kb_id = $kb_id
			SET n.chunks = COALESCE($chunks, [])
			SET n.attributes = COALESCE($attributes, [])
			RETURN n.id, n.name, n.entity_type
		`

		result, err := tx.Run(ctx, nodeQuery, map[string]interface{}{
			"id":           node.ID,
			"name":         node.Name,
			"title":        node.Name,
			"entity_type":  node.EntityType,
			"attributes":   node.Attributes,
			"chunks":       node.Chunks,
			"knowledge_id": namespace.Knowledge,
			"kb_id":        namespace.KBID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create node: %w", err)
		}

		// 检查是否匹配到节点并打印返回数据
		if result.Next(ctx) {
			record := result.Record()
			if id, ok := record.Get("n.id"); ok && id != nil {
				log.Printf("[Neo4j] AddNode SUCCESS: id=%v, name=%q", id, node.Name)
			}
		} else {
			log.Printf("[Neo4j] AddNode WARNING: no node returned")
		}

		// 消费结果
		if _, err := result.Consume(ctx); err != nil {
			return nil, fmt.Errorf("failed to consume result: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		log.Printf("[Neo4j] AddNode ERROR: failed to add node: %v", err)
		return err
	}

	return nil
}

// UpdateRelation 更新关系属性
// 返回更新后的关系，如果找不到则返回 nil
func (n *Neo4jRepository) UpdateRelation(ctx context.Context, namespace types.NameSpace, relation *types.GraphRelation) (*types.GraphRelation, error) {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT RETRIEVE GRAPH - driver is nil")
		return nil, nil
	}

	log.Printf("[Neo4j] UpdateRelation START: id=%s, kb_id=%s, type=%q, description=%q, strength=%f",
		relation.ID, namespace.KBID, relation.Type, relation.Description, relation.Strength)

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("[Neo4j] UpdateRelation session.Close error: %v", err)
		}
	}()

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 先查询是否存在该关系（只通过 id 查询，兼容旧数据）
		checkQuery := `
			MATCH ()-[r:RELATES_TO {id: $id}]->()
			RETURN r.id as rel_id, r.kb_id as kb_id, r.type as type, r.description as description, r.strength as strength
			LIMIT 1
		`

		checkResult, err := tx.Run(ctx, checkQuery, map[string]interface{}{
			"id": relation.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to check relation: %w", err)
		}

		var existingKBID string
		var existingType string
		var existingStrength float64
		if checkResult.Next(ctx) {
			record := checkResult.Record()
			if kbid, ok := record.Get("kb_id"); ok && kbid != nil {
				existingKBID = fmt.Sprintf("%v", kbid)
			}
			if typ, ok := record.Get("type"); ok && typ != nil {
				existingType = fmt.Sprintf("%v", typ)
			}
			if strength, ok := record.Get("strength"); ok && strength != nil {
				existingStrength = convertToFloat64(strength)
			}
			log.Printf("[Neo4j] Found existing relation: id=%s, kb_id=%s, type=%s, strength=%f", relation.ID, existingKBID, existingType, existingStrength)
		} else {
			log.Printf("[Neo4j] ERROR: Relation NOT FOUND in database! id=%s", relation.ID)
		}
		if _, err := checkResult.Consume(ctx); err != nil {
			return nil, fmt.Errorf("failed to consume check result: %w", err)
		}

		// 执行更新 - 只使用 id 匹配（不依赖 kb_id 以兼容旧数据）
		// 但要确保 kb_id 也被设置
		updateQuery := `
			MATCH ()-[r:RELATES_TO {id: $id}]->()
			SET r.type = $type
			SET r.description = $description
			SET r.strength = $strength
			SET r.weight = $weight
			SET r.kb_id = $kb_id
			RETURN r.id as rel_id, r.source as source, r.target as target, r.type as type, r.description as description, r.strength as strength, r.weight as weight
		`

		params := map[string]interface{}{
			"id":          relation.ID,
			"kb_id":       namespace.KBID,
			"type":        relation.Type,
			"description": relation.Description,
			"strength":    relation.Strength,
			"weight":      relation.Weight,
		}
		log.Printf("[Neo4j] UpdateRelation params: id=%s, kb_id=%s, type=%q, strength=%f, weight=%f",
			params["id"], params["kb_id"], params["type"], params["strength"], params["weight"])

		result, err := tx.Run(ctx, updateQuery, params)
		if err != nil {
			return nil, fmt.Errorf("failed to update relation: %w", err)
		}

		// 收集更新后的关系数据
		var updatedRel *types.GraphRelation
		if result.Next(ctx) {
			record := result.Record()
			updatedRel = &types.GraphRelation{
				ID:       relation.ID,
				ChunkIDs: []string{},
			}

			if source, ok := record.Get("source"); ok && source != nil {
				updatedRel.Source = fmt.Sprintf("%v", source)
			}
			if target, ok := record.Get("target"); ok && target != nil {
				updatedRel.Target = fmt.Sprintf("%v", target)
			}
			if relType, ok := record.Get("type"); ok && relType != nil {
				updatedRel.Type = fmt.Sprintf("%v", relType)
			}
			if relDesc, ok := record.Get("description"); ok && relDesc != nil {
				updatedRel.Description = fmt.Sprintf("%v", relDesc)
			}
			if relStrength, ok := record.Get("strength"); ok && relStrength != nil {
				updatedRel.Strength = convertToFloat64(relStrength)
			}
			if relWeight, ok := record.Get("weight"); ok && relWeight != nil {
				updatedRel.Weight = convertToFloat64(relWeight)
			}

			log.Printf("[Neo4j] UpdateRelation SUCCESS: id=%s, Source=%q, Target=%q, Type=%q, Desc=%q, Strength=%f, Weight=%f",
				updatedRel.ID, updatedRel.Source, updatedRel.Target, updatedRel.Type, updatedRel.Description, updatedRel.Strength, updatedRel.Weight)
		} else {
			log.Printf("[Neo4j] UpdateRelation WARNING: no relation returned after update")
		}

		// 消费剩余结果并处理错误
		if _, err := result.Consume(ctx); err != nil {
			return nil, fmt.Errorf("failed to consume result: %w", err)
		}

		return updatedRel, nil
	})

	if err != nil {
		log.Printf("[Neo4j] ERROR: failed to update relation: %v", err)
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	return result.(*types.GraphRelation), nil
}

// DeleteNode 删除单个节点
func (n *Neo4jRepository) DeleteNode(ctx context.Context, namespace types.NameSpace, nodeID string) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT DELETE NODE - driver is nil")
		return nil
	}

	log.Printf("[Neo4j] DeleteNode START: kb_id=%s, node_id=%s", namespace.KBID, nodeID)

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("[Neo4j] DeleteNode session.Close error: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 先删除与该节点相关的所有关系
		deleteRelQuery := `
			MATCH ()-[r:RELATES_TO {kb_id: $kb_id}]->()
			WHERE r.source = $node_id OR r.target = $node_id
			DELETE r
			RETURN count(r) as deleted_count
		`

		_, err := tx.Run(ctx, deleteRelQuery, map[string]interface{}{
			"kb_id":   namespace.KBID,
			"node_id": nodeID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete relations: %w", err)
		}

		// 删除节点本身（通过 name 或 id 匹配）
		// 不使用 label 过滤，只使用 kb_id 属性（兼容不同命名风格的节点）
		deleteNodeQuery := `
			MATCH (n)
			WHERE n.kb_id = $kb_id AND (n.name = $node_id OR n.id = $node_id)
			DETACH DELETE n
			RETURN count(n) as deleted_count
		`

		_, err = tx.Run(ctx, deleteNodeQuery, map[string]interface{}{
			"kb_id":   namespace.KBID,
			"node_id": nodeID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete node: %w", err)
		}

		log.Printf("[Neo4j] DeleteNode SUCCESS: node_id=%s", nodeID)

		return nil, nil
	})

	if err != nil {
		log.Printf("[Neo4j] DeleteNode ERROR: failed to delete node: %v", err)
		return err
	}

	log.Printf("[Neo4j] DeleteNode SUCCESS: node_id=%s", nodeID)

	return nil
}

// DeleteRelation 删除单个关系
func (n *Neo4jRepository) DeleteRelation(ctx context.Context, namespace types.NameSpace, relationID string) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT DELETE RELATION - driver is nil")
		return nil
	}

	log.Printf("[Neo4j] DeleteRelation START: kb_id=%s, relation_id=%s", namespace.KBID, relationID)

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("[Neo4j] DeleteRelation session.Close error: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 删除关系（只通过 id 匹配，兼容没有 kb_id 的旧数据）
		deleteRelQuery := `
			MATCH ()-[r:RELATES_TO {id: $id}]->()
			DELETE r
			RETURN r.id as deleted_id
		`

		result, err := tx.Run(ctx, deleteRelQuery, map[string]interface{}{
			"id": relationID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete relation: %w", err)
		}

		// 检查是否删除成功
		var deletedID string
		if result.Next(ctx) {
			record := result.Record()
			if id, ok := record.Get("deleted_id"); ok && id != nil {
				deletedID = fmt.Sprintf("%v", id)
			}
		}
		if _, err := result.Consume(ctx); err != nil {
			return nil, fmt.Errorf("failed to consume delete result: %w", err)
		}

		if deletedID != "" {
			log.Printf("[Neo4j] DeleteRelation SUCCESS: relation_id=%s", relationID)
		} else {
			log.Printf("[Neo4j] DeleteRelation WARNING: relation not found: %s", relationID)
		}

		return nil, nil
	})

	if err != nil {
		log.Printf("[Neo4j] DeleteRelation ERROR: failed to delete relation: %v", err)
		return err
	}

	log.Printf("[Neo4j] DeleteRelation SUCCESS: relation_id=%s", relationID)

	return nil
}

// ========================================
// 按知识库/分块删除 (用于文档删除时的清理)
// ========================================

// DeleteByChunkID 删除与指定 chunk_id 相关的节点（关系会自动删除）
func (n *Neo4jRepository) DeleteByChunkID(ctx context.Context, namespace types.NameSpace, chunkID string) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT DELETE BY CHUNK_ID - driver is nil")
		return nil
	}

	if chunkID == "" {
		return fmt.Errorf("chunk_id cannot be empty")
	}

	log.Printf("[Neo4j] DeleteByChunkID START: kb_id=%s, chunk_id=%s", namespace.KBID, chunkID)

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("[Neo4j] DeleteByChunkID session.Close error: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 删除该 chunk_id 对应的节点（关系会自动删除）
		// 节点通过 name 或 id 匹配
		deleteNodeQuery := `
			MATCH (n)
			WHERE (n.name = $chunk_id OR n.id = $chunk_id)
			DETACH DELETE n
			RETURN count(n) as deleted_count
		`

		result, err := tx.Run(ctx, deleteNodeQuery, map[string]interface{}{
			"chunk_id": chunkID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete node by chunk_id: %w", err)
		}

		if result.Next(ctx) {
			record := result.Record()
			if count, ok := record.Get("deleted_count"); ok {
				log.Printf("[Neo4j] Deleted %d nodes for chunk_id=%s", count, chunkID)
			}
		}

		log.Printf("[Neo4j] DeleteByChunkID SUCCESS: chunk_id=%s", chunkID)

		return nil, nil
	})

	if err != nil {
		log.Printf("[Neo4j] DeleteByChunkID ERROR: %v", err)
		return err
	}

	log.Printf("[Neo4j] DeleteByChunkID SUCCESS: chunk_id=%s", chunkID)

	return nil
}

// DeleteByKnowledgeID 删除与指定 knowledge_id 相关的所有节点（关系会自动删除）
func (n *Neo4jRepository) DeleteByKnowledgeID(ctx context.Context, namespace types.NameSpace, knowledgeID string) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT DELETE BY KNOWLEDGE_ID - driver is nil")
		return nil
	}

	if knowledgeID == "" {
		return fmt.Errorf("knowledge_id cannot be empty")
	}

	log.Printf("[Neo4j] DeleteByKnowledgeID START: kb_id=%s, knowledge_id=%s", namespace.KBID, knowledgeID)

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("[Neo4j] DeleteByKnowledgeID session.Close error: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 删除该 knowledge_id 对应的节点（关系会自动删除）
		deleteNodeQuery := `
			MATCH (n {knowledge_id: $knowledge_id})
			DETACH DELETE n
			RETURN count(n) as deleted_count
		`

		result, err := tx.Run(ctx, deleteNodeQuery, map[string]interface{}{
			"knowledge_id": knowledgeID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete node by knowledge_id: %w", err)
		}

		if result.Next(ctx) {
			record := result.Record()
			if count, ok := record.Get("deleted_count"); ok {
				log.Printf("[Neo4j] Deleted %d nodes for knowledge_id=%s", count, knowledgeID)
			}
		}

		log.Printf("[Neo4j] DeleteByKnowledgeID SUCCESS: knowledge_id=%s", knowledgeID)

		return nil, nil
	})

	if err != nil {
		log.Printf("[Neo4j] DeleteByKnowledgeID ERROR: %v", err)
		return err
	}

	log.Printf("[Neo4j] DeleteByKnowledgeID SUCCESS: knowledge_id=%s", knowledgeID)

	return nil
}

// ========================================
// 批量删除操作
// ========================================

// DeleteNodes 批量删除节点
func (n *Neo4jRepository) DeleteNodes(ctx context.Context, namespace types.NameSpace, nodeIDs []string) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT DELETE NODES - driver is nil")
		return nil
	}

	if len(nodeIDs) == 0 {
		return fmt.Errorf("node_ids cannot be empty")
	}

	log.Printf("[Neo4j] DeleteNodes START: kb_id=%s, count=%d", namespace.KBID, len(nodeIDs))

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("[Neo4j] DeleteNodes session.Close error: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 先删除与这些节点相关的所有关系
		deleteRelsQuery := `
			MATCH ()-[r:RELATES_TO {kb_id: $kb_id}]->()
			WHERE r.source in $node_ids OR r.target in $node_ids
			DELETE r
			RETURN count(r) as deleted_count
		`

		_, err := tx.Run(ctx, deleteRelsQuery, map[string]interface{}{
			"kb_id":    namespace.KBID,
			"node_ids": nodeIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete relations: %w", err)
		}

		// 删除节点本身
		deleteNodesQuery := `
			MATCH (n)
			WHERE n.kb_id = $kb_id AND (n.name in $node_ids OR n.id in $node_ids)
			DETACH DELETE n
			RETURN count(n) as deleted_count
		`

		result, err := tx.Run(ctx, deleteNodesQuery, map[string]interface{}{
			"kb_id":    namespace.KBID,
			"node_ids": nodeIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete nodes: %w", err)
		}

		// 获取删除数量
		if result.Next(ctx) {
			record := result.Record()
			if count, ok := record.Get("deleted_count"); ok {
				log.Printf("[Neo4j] DeleteNodes deleted %d nodes", count)
			}
		}

		log.Printf("[Neo4j] DeleteNodes SUCCESS: deleted %d nodes", len(nodeIDs))

		return nil, nil
	})

	if err != nil {
		log.Printf("[Neo4j] DeleteNodes ERROR: failed to delete nodes: %v", err)
		return err
	}

	return nil
}

// DeleteRelations 批量删除关系
func (n *Neo4jRepository) DeleteRelations(ctx context.Context, namespace types.NameSpace, relationIDs []string) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT DELETE RELATIONS - driver is nil")
		return nil
	}

	if len(relationIDs) == 0 {
		return fmt.Errorf("relation_ids cannot be empty")
	}

	log.Printf("[Neo4j] DeleteRelations START: kb_id=%s, count=%d", namespace.KBID, len(relationIDs))

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("[Neo4j] DeleteRelations session.Close error: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 构建删除查询（批量）
		deleteRelsQuery := `
			MATCH ()-[r:RELATES_TO]->()
			WHERE r.id in $relation_ids
			DELETE r
			RETURN count(r) as deleted_count
		`

		result, err := tx.Run(ctx, deleteRelsQuery, map[string]interface{}{
			"relation_ids": relationIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete relations: %w", err)
		}

		// 获取删除数量
		if result.Next(ctx) {
			record := result.Record()
			if count, ok := record.Get("deleted_count"); ok {
				log.Printf("[Neo4j] DeleteRelations deleted %d relations", count)
			}
		}

		log.Printf("[Neo4j] DeleteRelations SUCCESS: deleted %d relations", len(relationIDs))

		return nil, nil
	})

	if err != nil {
		log.Printf("[Neo4j] DeleteRelations ERROR: failed to delete relations: %v", err)
		return err
	}

	return nil
}

// DeleteByChunkIDs 批量按 chunk_id 删除相关节点
func (n *Neo4jRepository) DeleteByChunkIDs(ctx context.Context, namespace types.NameSpace, chunkIDs []string) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT DELETE BY CHUNK_IDS - driver is nil")
		return nil
	}

	if len(chunkIDs) == 0 {
		return fmt.Errorf("chunk_ids cannot be empty")
	}

	log.Printf("[Neo4j] DeleteByChunkIDs START: kb_id=%s, count=%d", namespace.KBID, len(chunkIDs))

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("[Neo4j] DeleteByChunkIDs session.Close error: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 删除这些 chunk_ids 对应的节点（关系会自动删除）
		deleteNodesQuery := `
			MATCH (n)
			WHERE n.name in $chunk_ids OR n.id in $chunk_ids
			DETACH DELETE n
			RETURN count(n) as deleted_count
		`

		result, err := tx.Run(ctx, deleteNodesQuery, map[string]interface{}{
			"chunk_ids": chunkIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete nodes by chunk_ids: %w", err)
		}

		// 获取删除数量
		if result.Next(ctx) {
			record := result.Record()
			if count, ok := record.Get("deleted_count"); ok {
				log.Printf("[Neo4j] DeleteByChunkIDs deleted %d nodes", count)
			}
		}

		log.Printf("[Neo4j] DeleteByChunkIDs SUCCESS: deleted %d chunk_ids", len(chunkIDs))

		return nil, nil
	})

	if err != nil {
		log.Printf("[Neo4j] DeleteByChunkIDs ERROR: %v", err)
		return err
	}

	return nil
}

// DeleteByKnowledgeIDs 批量按 knowledge_id 删除相关节点
func (n *Neo4jRepository) DeleteByKnowledgeIDs(ctx context.Context, namespace types.NameSpace, knowledgeIDs []string) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT DELETE BY KNOWLEDGE_IDS - driver is nil")
		return nil
	}

	if len(knowledgeIDs) == 0 {
		return fmt.Errorf("knowledge_ids cannot be empty")
	}

	log.Printf("[Neo4j] DeleteByKnowledgeIDs START: kb_id=%s, count=%d", namespace.KBID, len(knowledgeIDs))

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		if err := session.Close(ctx); err != nil {
			log.Printf("[Neo4j] DeleteByKnowledgeIDs session.Close error: %v", err)
		}
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 删除这些 knowledge_ids 对应的节点（关系会自动删除）
		deleteNodesQuery := `
			MATCH (n)
			WHERE n.knowledge_id in $knowledge_ids
			DETACH DELETE n
			RETURN count(n) as deleted_count
		`

		result, err := tx.Run(ctx, deleteNodesQuery, map[string]interface{}{
			"knowledge_ids": knowledgeIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete nodes by knowledge_ids: %w", err)
		}

		// 获取删除数量
		if result.Next(ctx) {
			record := result.Record()
			if count, ok := record.Get("deleted_count"); ok {
				log.Printf("[Neo4j] DeleteByKnowledgeIDs deleted %d nodes", count)
			}
		}

		log.Printf("[Neo4j] DeleteByKnowledgeIDs SUCCESS: deleted %d knowledge_ids", len(knowledgeIDs))

		return nil, nil
	})

	if err != nil {
		log.Printf("[Neo4j] DeleteByKnowledgeIDs ERROR: %v", err)
		return err
	}

	return nil
}

// ========================================
// 统计信息
// ========================================

// GetGraphStats 获取图谱统计信息
func (n *Neo4jRepository) GetGraphStats(ctx context.Context, namespace types.NameSpace) (*interfaces.GraphStats, error) {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: NOT SUPPORT GET GRAPH STATS - driver is nil")
		return &interfaces.GraphStats{}, nil
	}

	session := n.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		var nodeQuery string

		if namespace.Knowledge != "" {
			nodeQuery = `
				MATCH (n {knowledge_id: $knowledge_id, kb_id: $kb_id})
				RETURN count(n) as node_count
			`
		} else {
			nodeQuery = `
				MATCH (n {kb_id: $kb_id})
				RETURN count(n) as node_count
			`
		}

		nodeResult, err := tx.Run(ctx, nodeQuery, map[string]interface{}{
			"knowledge_id": namespace.Knowledge,
			"kb_id":        namespace.KBID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to query node count: %w", err)
		}

		var nodeCount int64
		if nodeResult.Next(ctx) {
			record := nodeResult.Record()
			if count, ok := record.Get("node_count"); ok {
				nodeCount = int64(convertToFloat64(count))
			}
		}

		// 查询关系数量
		var relQuery string
		if namespace.Knowledge != "" {
			relQuery = `
				MATCH (n {kb_id: $kb_id})-[r:RELATES_TO]->(m {kb_id: $kb_id})
				WHERE n.knowledge_id = $knowledge_id AND m.knowledge_id = $knowledge_id
				RETURN count(DISTINCT r.id) as rel_count
			`
		} else {
			relQuery = `
				MATCH (n {kb_id: $kb_id})-[r:RELATES_TO]->(m {kb_id: $kb_id})
				RETURN count(DISTINCT r.id) as rel_count
			`
		}

		relResult, err := tx.Run(ctx, relQuery, map[string]interface{}{
			"knowledge_id": namespace.Knowledge,
			"kb_id":        namespace.KBID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to query relation count: %w", err)
		}

		var relCount int64
		if relResult.Next(ctx) {
			record := relResult.Record()
			if count, ok := record.Get("rel_count"); ok {
				relCount = int64(convertToFloat64(count))
			}
		}

		// 收集关联的 chunk_ids
		chunkQuery := `
			MATCH (n {kb_id: $kb_id})
			WHERE n.chunks IS NOT NULL AND size(n.chunks) > 0
			UNWIND n.chunks as chunk_id
			RETURN DISTINCT chunk_id
			LIMIT 10000
		`

		chunkResult, err := tx.Run(ctx, chunkQuery, map[string]interface{}{
			"kb_id": namespace.KBID,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to query chunk ids: %w", err)
		}

		chunkIDs := make([]string, 0)
		for chunkResult.Next(ctx) {
			record := chunkResult.Record()
			if chunkID, ok := record.Get("chunk_id"); ok && chunkID != nil {
				chunkIDs = append(chunkIDs, fmt.Sprintf("%v", chunkID))
			}
		}

		return &interfaces.GraphStats{
			NodeCount:     nodeCount,
			RelationCount: relCount,
			ChunkCount:    int64(len(chunkIDs)),
			ChunkIDs:      chunkIDs,
		}, nil
	})

	if err != nil {
		log.Printf("[Neo4j] ERROR: get graph stats failed: %v", err)
		return nil, err
	}

	stats, ok := result.(*interfaces.GraphStats)
	if !ok {
		return &interfaces.GraphStats{}, nil
	}

	log.Printf("[Neo4j] GetGraphStats: nodes=%d, relations=%d, chunks=%d",
		stats.NodeCount, stats.RelationCount, stats.ChunkCount)

	return stats, nil
}

// Close 关闭 Neo4j 驱动连接
func (n *Neo4jRepository) Close(ctx context.Context) error {
	if n.driver == nil {
		log.Printf("[Neo4j] WARN: driver is nil, nothing to close")
		return nil
	}

	err := n.driver.Close(ctx)
	if err != nil {
		log.Printf("[Neo4j] ERROR: failed to close driver: %v", err)
		return fmt.Errorf("failed to close neo4j driver: %w", err)
	}

	log.Printf("[Neo4j] Driver closed successfully")
	return nil
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
