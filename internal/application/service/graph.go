package service

import (
	"context"
	"encoding/json"
	"fmt"
	"link/internal/config"
	"link/internal/models/chat"
	"log"
	"math"
	"slices"
	"strings"
	"sync"

	"link/internal/types"
	"link/internal/types/interfaces"

	"github.com/google/uuid"
)

// GraphService 图谱服务
type GraphService struct {
	graphRepo      interfaces.Neo4jGraphRepository // Neo4j 图谱操作仓储
	graphQueryRepo interfaces.GraphQueryRepository // 图谱与知识库关联查询仓储
	chunkRepo      interfaces.ChunkRepository      // Chunk 仓储
	graphCache     *graphCache
	mutex          sync.RWMutex
}

// NewGraphService 创建图谱服务实例
func NewGraphService(graphRepo interfaces.Neo4jGraphRepository) *GraphService {
	return &GraphService{
		graphRepo: graphRepo,
		graphCache: &graphCache{
			nodes:     make(map[string]*types.GraphNode),
			relations: make(map[string]*types.GraphRelation),
		},
	}
}

// NewGraphServiceWithQuery 创建图谱服务实例（包含查询仓储）
func NewGraphServiceWithQuery(graphRepo interfaces.Neo4jGraphRepository, graphQueryRepo interfaces.GraphQueryRepository) *GraphService {
	return &GraphService{
		graphRepo:      graphRepo,
		graphQueryRepo: graphQueryRepo,
		graphCache: &graphCache{
			nodes:     make(map[string]*types.GraphNode),
			relations: make(map[string]*types.GraphRelation),
		},
	}
}

// NewGraphServiceWithChunks 创建图谱服务实例（包含chunk仓储）
func NewGraphServiceWithChunks(graphRepo interfaces.Neo4jGraphRepository, graphQueryRepo interfaces.GraphQueryRepository, chunkRepo interfaces.ChunkRepository) *GraphService {
	return &GraphService{
		graphRepo:      graphRepo,
		graphQueryRepo: graphQueryRepo,
		chunkRepo:      chunkRepo,
		graphCache: &graphCache{
			nodes:     make(map[string]*types.GraphNode),
			relations: make(map[string]*types.GraphRelation),
		},
	}
}

// ========================================
// 图谱提取相关类型
// ========================================

// ExtractionMode 提取模式
type ExtractionMode string

const (
	ExtractionModeDocument ExtractionMode = "document" // 文档入库模式：完整提取
	ExtractionModeQuery    ExtractionMode = "query"    // 查询模式：仅提取相关
)

// ChunkExtractionInput 文档块提取输入
type ChunkExtractionInput struct {
	KBID     string         // 知识库ID
	ChunkID  string         // 文档块ID
	Document string         // 文档内容
	Query    string         // 提取查询
	Mode     ExtractionMode // 提取模式：document/query
}

// ExtractedGraphData 提取的图谱数据
type ExtractedGraphData struct {
	ChunkID   string                 // 文档块ID
	Nodes     []*types.GraphNode     // 提取的节点
	Relations []*types.GraphRelation // 提取的关系
}

// graphCache 图谱缓存，用于并发安全的数据合并
type graphCache struct {
	mutex     sync.RWMutex
	nodes     map[string]*types.GraphNode     // key: entity title
	relations map[string]*types.GraphRelation // key: "source#target"
}

// ExtractGraphFromChunks 从文档块中提取图谱（并发处理，最多4个线程）
// 默认使用文档入库模式（完整提取）
func (s *GraphService) ExtractGraphFromChunks(
	ctx context.Context,
	inputs []*ChunkExtractionInput,
) (*types.GraphData, error) {
	return s.ExtractGraphFromChunksWithMode(ctx, inputs, ExtractionModeDocument)
}

// ExtractGraphFromChunksWithQuery 从文档块中提取图谱（查询模式）
// 仅提取与查询相关的实体和关系
func (s *GraphService) ExtractGraphFromChunksWithQuery(
	ctx context.Context,
	inputs []*ChunkExtractionInput,
) (*types.GraphData, error) {
	return s.ExtractGraphFromChunksWithMode(ctx, inputs, ExtractionModeQuery)
}

// ExtractGraphFromChunksWithMode 从文档块中提取图谱（指定模式）
func (s *GraphService) ExtractGraphFromChunksWithMode(
	ctx context.Context,
	inputs []*ChunkExtractionInput,
	mode ExtractionMode,
) (*types.GraphData, error) {
	if len(inputs) == 0 {
		return &types.GraphData{}, nil
	}

	// 清空缓存
	s.graphCache.mutex.Lock()
	s.graphCache.nodes = make(map[string]*types.GraphNode)
	s.graphCache.relations = make(map[string]*types.GraphRelation)
	s.graphCache.mutex.Unlock()

	log.Printf("[GraphService] Starting concurrent extraction for %d chunks, mode=%s", len(inputs), mode)

	// 使用 semaphore 限制并发数为 4
	semaphore := make(chan struct{}, 4)
	var wg sync.WaitGroup

	results := make([]*ExtractedGraphData, len(inputs))
	errors := make([]error, len(inputs))

	// 并发提取每个文档块的图谱
	for i, input := range inputs {
		wg.Add(1)
		go func(idx int, inp *ChunkExtractionInput) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 确定输入的模式（如果未设置则使用传入的模式）
			extractionMode := inp.Mode
			if extractionMode == "" {
				extractionMode = mode
			}

			log.Printf("[GraphService] Processing chunk %s, mode=%s", inp.ChunkID, extractionMode)

			// 提取实体
			nodes, err := s.extractEntities(ctx, inp.ChunkID, inp.Document, inp.Query, extractionMode)
			if err != nil {
				errors[idx] = fmt.Errorf("failed to extract entities from chunk %s: %w", inp.ChunkID, err)
				log.Printf("[GraphService] Error extracting entities from chunk %s: %v", inp.ChunkID, err)
				return
			}

			// 提取关系（基于已提取的实体）
			relations, err := s.extractRelations(ctx, inp.KBID, inp.ChunkID, inp.Document, inp.Query, nodes, extractionMode)
			if err != nil {
				errors[idx] = fmt.Errorf("failed to extract relations from chunk %s: %w", inp.ChunkID, err)
				log.Printf("[GraphService] Error extracting relations from chunk %s: %v", inp.ChunkID, err)
				return
			}

			results[idx] = &ExtractedGraphData{
				ChunkID:   inp.ChunkID,
				Nodes:     nodes,
				Relations: relations,
			}

			log.Printf("[GraphService] Completed chunk %s: %d nodes, %d relations",
				inp.ChunkID, len(nodes), len(relations))
		}(i, input)
	}

	// 等待所有 goroutine 完成
	wg.Wait()

	// 检查错误
	for _, err := range errors {
		if err != nil {
			log.Printf("[GraphService] Warning: %v", err)
		}
	}

	// 过滤掉失败的结果
	validResults := make([]*ExtractedGraphData, 0, len(results))
	for _, result := range results {
		if result != nil {
			validResults = append(validResults, result)
		}
	}

	log.Printf("[GraphService] Extraction completed: %d/%d chunks succeeded, mode=%s", len(validResults), len(inputs), mode)

	// 合并所有提取的图谱数据
	mergedGraph, err := s.mergeExtractedGraphs(ctx, validResults)
	if err != nil {
		return nil, fmt.Errorf("failed to merge extracted graphs: %w", err)
	}

	log.Printf("[GraphService] Merged graph: %d nodes, %d relations",
		len(mergedGraph.Node), len(mergedGraph.Relation))

	return mergedGraph, nil
}

// extractEntities 从文档中提取实体
func (s *GraphService) extractEntities(
	ctx context.Context,
	chunkID, document, query string,
	mode ExtractionMode, // 提取模式：document/query
) ([]*types.GraphNode, error) {
	// 根据模式选择提示词模板
	var templateName string
	if mode == ExtractionModeQuery {
		templateName = "entity_extraction_query"
	} else {
		templateName = "entity_extraction"
	}

	promptTemplate, err := config.LoadPromptTemplate(templateName)
	if err != nil {
		return nil, fmt.Errorf("failed to load entity extraction template: %w", err)
	}

	// 替换占位符
	prompt := strings.Replace(promptTemplate, "{{document}}", document, 1)

	// 查询模式：添加查询信息
	if mode == ExtractionModeQuery && query != "" {
		prompt = strings.Replace(prompt, "{{query}}", query, 1)
	}

	log.Printf("[GraphService] Entity extraction mode=%s prompt length: %d", mode, len(prompt))

	// 创建 Chat
	chatConfig := config.LoadChatConfig()
	chatModel, err := chat.NewChat(&chat.ChatConfig{
		Source:    chatConfig.Source,
		BaseURL:   chatConfig.BaseURL,
		ModelName: chatConfig.ModelName,
		APIKey:    chatConfig.APIKey,
		Provider:  chatConfig.Provider,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	// 构建消息
	messages := []chat.Message{
		{Role: "user", Content: prompt},
	}

	// 调用 LLM
	response, err := chatModel.Chat(ctx, messages, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate entity extraction: %w", err)
	}

	log.Printf("[GraphService] Entity extraction raw response: %s", response.Content)

	// 清理响应内容
	cleanedContent := cleanJSONResponse(response.Content)
	log.Printf("[GraphService] Entity extraction cleaned response: %s", cleanedContent)

	// 解析 JSON 响应
	var result struct {
		Nodes []*types.GraphNode `json:"nodes"`
	}
	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		return nil, fmt.Errorf("failed to parse entity extraction response: %w, content: %s", err, cleanedContent)
	}

	// 为每个节点添加 chunk_id
	for _, node := range result.Nodes {
		if node.ID == "" {
			node.ID = uuid.New().String()
		}
		if !slices.Contains(node.Chunks, chunkID) {
			node.Chunks = append(node.Chunks, chunkID)
		}
	}

	return result.Nodes, nil
}

// extractRelations 从文档和实体中提取关系
func (s *GraphService) extractRelations(
	ctx context.Context,
	kbID, chunkID, document, query string,
	entities []*types.GraphNode,
	mode ExtractionMode, // 提取模式：document/query
) ([]*types.GraphRelation, error) {
	// 根据模式选择提示词模板
	var templateName string
	if mode == ExtractionModeQuery {
		templateName = "relation_extraction_query"
	} else {
		templateName = "relation_extraction"
	}

	promptTemplate, err := config.LoadPromptTemplate(templateName)
	if err != nil {
		return nil, fmt.Errorf("failed to load relation extraction template: %w", err)
	}

	// 构建实体列表 JSON（仅包含名称）
	entityNames := make([]string, 0, len(entities))
	for _, e := range entities {
		entityNames = append(entityNames, fmt.Sprintf(`"%s"`, e.Name))
	}
	entitiesList := fmt.Sprintf("[%s]", strings.Join(entityNames, ", "))

	// 替换占位符
	prompt := strings.Replace(promptTemplate, "{{entities}}", entitiesList, 1)
	prompt = strings.Replace(prompt, "{{document}}", document, 1)

	// 查询模式：添加查询信息
	if mode == ExtractionModeQuery && query != "" {
		prompt = strings.Replace(prompt, "{{query}}", query, 1)
	}

	log.Printf("[GraphService] Relation extraction mode=%s prompt length: %d", mode, len(prompt))

	// 创建 Chat
	chatConfig := config.LoadChatConfig()
	chatModel, err := chat.NewChat(&chat.ChatConfig{
		Source:    chatConfig.Source,
		BaseURL:   chatConfig.BaseURL,
		ModelName: chatConfig.ModelName,
		APIKey:    chatConfig.APIKey,
		Provider:  chatConfig.Provider,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	// 构建消息
	messages := []chat.Message{
		{Role: "user", Content: prompt},
	}

	// 调用 LLM
	response, err := chatModel.Chat(ctx, messages, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate relation extraction: %w", err)
	}

	log.Printf("[GraphService] Relation extraction raw response: %s", response.Content)

	// 清理响应内容
	cleanedContent := cleanJSONResponse(response.Content)
	log.Printf("[GraphService] Relation extraction cleaned response: %s", cleanedContent)

	// 解析 JSON 响应
	var result struct {
		Relations []*types.GraphRelation `json:"relations"`
	}
	if err := json.Unmarshal([]byte(cleanedContent), &result); err != nil {
		return nil, fmt.Errorf("failed to parse relation extraction response: %w, content: %s", err, cleanedContent)
	}

	// 验证 source 和 target 是否存在于实体列表中
	entitySet := make(map[string]bool)
	for _, entity := range entities {
		entitySet[entity.Name] = true
	}

	validRelations := make([]*types.GraphRelation, 0, len(result.Relations))
	for _, rel := range result.Relations {
		// 验证 source 和 target
		if !entitySet[rel.Source] {
			log.Printf("[GraphService] Warning: relation source '%s' not found in entities", rel.Source)
			continue
		}
		if !entitySet[rel.Target] {
			log.Printf("[GraphService] Warning: relation target '%s' not found in entities", rel.Target)
			continue
		}

		// 设置关系 ID
		if rel.ID == "" {
			rel.ID = uuid.New().String()
		}

		// 根据source和target实体名称从数据库中查找包含这些实体的chunk
		matchedChunks := s.findChunksByEntities(ctx, kbID, rel.Source, rel.Target)
		for _, dbChunkID := range matchedChunks {
			if !slices.Contains(rel.ChunkIDs, dbChunkID) {
				rel.ChunkIDs = append(rel.ChunkIDs, dbChunkID)
			}
		}

		validRelations = append(validRelations, rel)
	}

	return validRelations, nil
}

// mergeExtractedGraphs 合并多个提取的图谱数据
// 在构建实体时就加锁，保证并发安全
func (s *GraphService) mergeExtractedGraphs(
	ctx context.Context,
	dataList []*ExtractedGraphData,
) (*types.GraphData, error) {
	if len(dataList) == 0 {
		return &types.GraphData{}, nil
	}

	// 获取锁（在构建实体时加锁）
	s.graphCache.mutex.Lock()
	defer s.graphCache.mutex.Unlock()

	// 用于统计 PMI 计算
	totalChunks := len(dataList)
	entityChunkCount := make(map[string]int)  // 每个实体出现的文档块数
	coOccurrenceCount := make(map[string]int) // 每对实体共同出现的文档块数

	// 第一阶段：合并节点（已经在锁保护下）
	for _, data := range dataList {
		for _, node := range data.Nodes {
			if existingNode, exists := s.graphCache.nodes[node.Name]; exists {
				// 合并 chunks（去重）
				for _, chunk := range node.Chunks {
					if !slices.Contains(existingNode.Chunks, chunk) {
						existingNode.Chunks = append(existingNode.Chunks, chunk)
					}
				}
				// 合并 attributes（去重）
				for _, attr := range node.Attributes {
					if !slices.Contains(existingNode.Attributes, attr) {
						existingNode.Attributes = append(existingNode.Attributes, attr)
					}
				}
			} else {
				s.graphCache.nodes[node.Name] = node
			}
		}
	}

	// 第二阶段：合并关系
	for _, data := range dataList {
		for _, rel := range data.Relations {
			key := fmt.Sprintf("%s#%s", rel.Source, rel.Target)

			if existingRel, exists := s.graphCache.relations[key]; exists {
				// 合并 chunk_ids（去重）
				for _, chunkID := range rel.ChunkIDs {
					if !slices.Contains(existingRel.ChunkIDs, chunkID) {
						existingRel.ChunkIDs = append(existingRel.ChunkIDs, chunkID)
					}
				}
				// 加权平均更新 strength
				if existingRel.Strength > 0 && rel.Strength > 0 {
					existingRel.Strength = (existingRel.Strength + rel.Strength) / 2
				} else if rel.Strength > 0 {
					existingRel.Strength = rel.Strength
				}
			} else {
				s.graphCache.relations[key] = rel
			}
		}
	}

	// 统计实体出现的文档块数
	for _, node := range s.graphCache.nodes {
		entityChunkCount[node.Name] = len(node.Chunks)
	}

	// 统计实体对共同出现的文档块数
	for _, rel := range s.graphCache.relations {
		// 找到共同出现的文档块
		sourceChunks := s.graphCache.nodes[rel.Source].Chunks
		targetChunks := s.graphCache.nodes[rel.Target].Chunks
		commonChunks := intersection(sourceChunks, targetChunks)
		coOccurrenceCount[fmt.Sprintf("%s#%s", rel.Source, rel.Target)] = len(commonChunks)
	}

	// 第三阶段：计算 PMI 和 Weight（在关系去重后串行执行）
	for _, rel := range s.graphCache.relations {
		key := fmt.Sprintf("%s#%s", rel.Source, rel.Target)

		// 计算概率
		p_x_y := float64(coOccurrenceCount[key]) / float64(totalChunks)
		p_x := float64(entityChunkCount[rel.Source]) / float64(totalChunks)
		p_y := float64(entityChunkCount[rel.Target]) / float64(totalChunks)

		// PMI = log2(P(x,y) / (P(x) * P(y)))
		var pmi float64
		if p_x > 0 && p_y > 0 && p_x_y > 0 {
			pmi = math.Log2(p_x_y / (p_x * p_y))
		}

		// 归一化 PMI 到 [0, 1]（假设 PMI 范围为 [-5, 10]）
		normalizedPMI := (pmi + 5) / 15
		normalizedPMI = math.Max(0, math.Min(1, normalizedPMI))

		// 归一化 Strength 到 [0, 1]（假设范围 [1, 10]）
		normalizedStrength := (rel.Strength - 1) / 9
		normalizedStrength = math.Max(0, math.Min(1, normalizedStrength))

		// Weight = 1.0 + 9.0 * (normalizedPMI * 0.6 + normalizedStrength * 0.4)
		rel.Weight = 1.0 + 9.0*(normalizedPMI*0.6+normalizedStrength*0.4)

		// 计算 CombinedDegree
		sourceDegree := 0
		targetDegree := 0
		for _, r := range s.graphCache.relations {
			if r.Source == rel.Source || r.Target == rel.Source {
				sourceDegree++
			}
			if r.Source == rel.Target || r.Target == rel.Target {
				targetDegree++
			}
		}
		rel.CombinedDegree = sourceDegree + targetDegree

		log.Printf("[GraphService] Relation %s -> %s: PMI=%.2f, Weight=%.2f, CombinedDegree=%d",
			rel.Source, rel.Target, pmi, rel.Weight, rel.CombinedDegree)
	}

	// 构建最终结果
	result := &types.GraphData{
		Node:     make([]*types.GraphNode, 0, len(s.graphCache.nodes)),
		Relation: make([]*types.GraphRelation, 0, len(s.graphCache.relations)),
	}

	for _, node := range s.graphCache.nodes {
		result.Node = append(result.Node, node)
	}

	for _, rel := range s.graphCache.relations {
		result.Relation = append(result.Relation, rel)
	}

	return result, nil
}

// intersection 计算两个字符串数组的交集
func intersection(a, b []string) []string {
	set := make(map[string]bool)
	for _, item := range a {
		set[item] = true
	}

	result := make([]string, 0)
	for _, item := range b {
		if set[item] {
			result = append(result, item)
		}
	}
	return result
}

// AddGraph 添加图谱数据
func (s *GraphService) AddGraph(ctx context.Context, namespace types.NameSpace, graphs []*types.GraphData) error {
	log.Printf("[Service] AddGraph START: namespace.KBID=%s, len(graphs)=%d", namespace.KBID, len(graphs))

	// 为节点生成UUID（如果还没有）
	for _, graph := range graphs {
		log.Printf("[Service] AddGraph Processing graph: nodes=%d, relations=%d", len(graph.Node), len(graph.Relation))

		for _, node := range graph.Node {
			if node.ID == "" {
				node.ID = uuid.New().String()
				log.Printf("[Service] AddGraph Generated node ID: %s", node.ID)
			}
		}

		// 为关系生成UUID并计算属性
		for i, rel := range graph.Relation {
			if rel.ID == "" {
				rel.ID = uuid.New().String()
				log.Printf("[Service] AddGraph Generated relation ID[%d]: %s", i, rel.ID)
			}
			log.Printf("[Service] AddGraph Relation[%d]: ID=%s, Source=%q, Target=%q, Type=%q",
				i, rel.ID, rel.Source, rel.Target, rel.Type)
			// 计算关系属性（占位符，实际计算在后续实现）
			s.calculateRelationProperties(graph, rel)
		}
	}

	log.Printf("[Service] AddGraph Calling Repo.AddGraph")
	err := s.graphRepo.AddGraph(ctx, namespace, graphs)

	if err != nil {
		log.Printf("[Service] AddGraph ERROR from Repo: %v", err)
		return err
	}

	log.Printf("[Service] AddGraph SUCCESS")
	return nil
}

// calculateRelationProperties 计算关系属性
// 注意: PMI 和 Weight 计算现在在 mergeExtractedGraphs 中完成
func (s *GraphService) calculateRelationProperties(graph *types.GraphData, rel *types.GraphRelation) {
	// 设置默认值
	if rel.Strength == 0 {
		rel.Strength = 5.0
	}
	if rel.Weight == 0 {
		rel.Weight = 5.0
	}
}

// DeleteGraph 删除图谱数据
func (s *GraphService) DeleteGraph(ctx context.Context, namespaces []types.NameSpace) error {
	return s.graphRepo.DeleteGraph(ctx, namespaces)
}

// GetGraph 获取完整图谱数据
func (s *GraphService) GetGraph(ctx context.Context, namespace types.NameSpace) (*types.GraphData, error) {
	return s.graphRepo.GetGraph(ctx, namespace)
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

// UpdateNode 更新节点属性
func (s *GraphService) UpdateNode(ctx context.Context, namespace types.NameSpace, node *types.GraphNode) error {
	return s.graphRepo.UpdateNode(ctx, namespace, node)
}

// AddRelation 添加单个关系
func (s *GraphService) AddRelation(ctx context.Context, namespace types.NameSpace, relation *types.GraphRelation) (*types.GraphRelation, error) {
	log.Printf("[Service] AddRelation START: namespace.KBID=%s, ID=%s, Source=%q, Target=%q, Type=%q, Strength=%f",
		namespace.KBID, relation.ID, relation.Source, relation.Target, relation.Type, relation.Strength)

	result, err := s.graphRepo.AddRelation(ctx, namespace, relation)

	if err != nil {
		log.Printf("[Service] AddRelation ERROR from Repo: %v", err)
		return nil, err
	}

	if result == nil {
		log.Printf("[Service] AddRelation WARNING: Repo returned nil")
	} else {
		log.Printf("[Service] AddRelation SUCCESS: returning ID=%s, Source=%q, Target=%q, Type=%q, Strength=%f, Weight=%f",
			result.ID, result.Source, result.Target, result.Type, result.Strength, result.Weight)
	}

	return result, nil
}

// AddNode 添加单个节点
func (s *GraphService) AddNode(ctx context.Context, namespace types.NameSpace, node *types.GraphNode) error {
	log.Printf("[Service] AddNode START: namespace.KBID=%s, ID=%s, Name=%q, EntityType=%q",
		namespace.KBID, node.ID, node.Name, node.EntityType)

	err := s.graphRepo.AddNode(ctx, namespace, node)

	if err != nil {
		log.Printf("[Service] AddNode ERROR from Repo: %v", err)
		return err
	}

	log.Printf("[Service] AddNode SUCCESS: ID=%s, Name=%q", node.ID, node.Name)
	return nil
}

// UpdateRelation 更新关系属性
func (s *GraphService) UpdateRelation(ctx context.Context, namespace types.NameSpace, relation *types.GraphRelation) (*types.GraphRelation, error) {
	log.Printf("[Service] UpdateRelation START: namespace.KBID=%s, relation.ID=%s, relation.Type=%q",
		namespace.KBID, relation.ID, relation.Type)

	result, err := s.graphRepo.UpdateRelation(ctx, namespace, relation)

	if err != nil {
		log.Printf("[Service] UpdateRelation ERROR from Repo: %v", err)
		return nil, err
	}

	if result == nil {
		log.Printf("[Service] UpdateRelation WARNING: Repo returned nil")
	} else {
		log.Printf("[Service] UpdateRelation SUCCESS: returning ID=%s, Source=%q, Target=%q, Type=%q, Strength=%f, Weight=%f",
			result.ID, result.Source, result.Target, result.Type, result.Strength, result.Weight)
	}

	return result, nil
}

// DeleteNode 删除单个节点
func (s *GraphService) DeleteNode(ctx context.Context, namespace types.NameSpace, nodeID string) error {
	log.Printf("[Service] DeleteNode START: namespace.KBID=%s, node_id=%s", namespace.KBID, nodeID)

	err := s.graphRepo.DeleteNode(ctx, namespace, nodeID)

	if err != nil {
		log.Printf("[Service] DeleteNode ERROR: %v", err)
		return err
	}

	// 从缓存中移除该节点相关的关系
	s.graphCache.mutex.Lock()
	defer s.graphCache.mutex.Unlock()

	// 过滤掉要删除的关系
	for id, rel := range s.graphCache.relations {
		if rel.Source != nodeID && rel.Target != nodeID {
			delete(s.graphCache.relations, id)
		}
	}
	// 过滤掉要删除的节点
	for id, node := range s.graphCache.nodes {
		if node.Name != nodeID && node.ID != nodeID {
			delete(s.graphCache.nodes, id)
		}
	}

	log.Printf("[Service] DeleteNode SUCCESS: node_id=%s", nodeID)

	return nil
}

// DeleteRelation 删除单个关系
func (s *GraphService) DeleteRelation(ctx context.Context, namespace types.NameSpace, relationID string) error {
	log.Printf("[Service] DeleteRelation START: namespace.KBID=%s, relation_id=%s", namespace.KBID, relationID)

	err := s.graphRepo.DeleteRelation(ctx, namespace, relationID)

	if err != nil {
		log.Printf("[Service] DeleteRelation ERROR: %v", err)
		return err
	}

	log.Printf("[Service] DeleteRelation SUCCESS: relation_id=%s", relationID)

	// 从缓存中移除该关系
	s.graphCache.mutex.Lock()
	defer s.graphCache.mutex.Unlock()

	delete(s.graphCache.relations, relationID)

	return nil
}

// ========================================
// 按知识库/分块删除 (用于文档删除时的清理)
// ========================================

// DeleteByChunkID 删除与指定 chunk_id 相关的所有图谱数据
func (s *GraphService) DeleteByChunkID(ctx context.Context, namespace types.NameSpace, chunkID string) error {
	log.Printf("[Service] DeleteByChunkID START: namespace.KBID=%s, chunk_id=%s", namespace.KBID, chunkID)

	err := s.graphRepo.DeleteByChunkID(ctx, namespace, chunkID)

	if err != nil {
		log.Printf("[Service] DeleteByChunkID ERROR: %v", err)
		return err
	}

	// 从缓存中移除相关的节点和关系
	s.graphCache.mutex.Lock()
	defer s.graphCache.mutex.Unlock()

	// 移除包含该 chunk_id 的所有关系
	for id, rel := range s.graphCache.relations {
		if containsChunkID(rel.ChunkIDs, chunkID) {
			delete(s.graphCache.relations, id)
		}
	}
	// 移除该 chunk 对应的节点
	delete(s.graphCache.nodes, chunkID)

	log.Printf("[Service] DeleteByChunkID SUCCESS: chunk_id=%s", chunkID)

	return nil
}

// DeleteByKnowledgeID 删除与指定 knowledge_id 相关的所有图谱数据
func (s *GraphService) DeleteByKnowledgeID(ctx context.Context, namespace types.NameSpace, knowledgeID string) error {
	log.Printf("[Service] DeleteByKnowledgeID START: namespace.KBID=%s, knowledge_id=%s", namespace.KBID, knowledgeID)

	err := s.graphRepo.DeleteByKnowledgeID(ctx, namespace, knowledgeID)

	if err != nil {
		log.Printf("[Service] DeleteByKnowledgeID ERROR: %v", err)
		return err
	}

	// 从缓存中移除相关的节点和关系
	s.graphCache.mutex.Lock()
	defer s.graphCache.mutex.Unlock()

	// 移除该 knowledge 对应的节点
	delete(s.graphCache.nodes, knowledgeID)

	// 移除所有与该 knowledge_id 节点相连的关系
	for id, rel := range s.graphCache.relations {
		if rel.Source == knowledgeID || rel.Target == knowledgeID {
			delete(s.graphCache.relations, id)
		}
	}

	log.Printf("[Service] DeleteByKnowledgeID SUCCESS: knowledge_id=%s", knowledgeID)

	return nil
}

// containsChunkID 检查关系的 ChunkIDs 列表中是否包含指定的 chunk_id
func containsChunkID(chunkIDs []string, chunkID string) bool {
	for _, id := range chunkIDs {
		if id == chunkID {
			return true
		}
	}
	return false
}

// containsKnowledgeID 检查关系的 KnowledgeIDs 列表中是否包含指定的 knowledge_id
func containsKnowledgeID(knowledgeIDs []string, knowledgeID string) bool {
	for _, id := range knowledgeIDs {
		if id == knowledgeID {
			return true
		}
	}
	return false
}

// ========================================
// 图谱查询相关方法
// ========================================

// FindRelationsByChunk 查找涉及特定文档块的所有关系
func (s *GraphService) FindRelationsByChunk(
	ctx context.Context,
	namespace types.NameSpace,
	chunkID string,
) ([]*types.GraphRelation, error) {
	// 遍历缓存中的关系进行过滤
	s.graphCache.mutex.RLock()
	defer s.graphCache.mutex.RUnlock()

	var result []*types.GraphRelation
	for _, rel := range s.graphCache.relations {
		for _, cid := range rel.ChunkIDs {
			if cid == chunkID {
				result = append(result, rel)
				break
			}
		}
	}
	return result, nil
}

// FindStrongRelations 查找强度高于阈值的关系
func (s *GraphService) FindStrongRelations(
	ctx context.Context,
	namespace types.NameSpace,
	minStrength float64,
) ([]*types.GraphRelation, error) {
	// 遍历缓存中的关系进行过滤
	s.graphCache.mutex.RLock()
	defer s.graphCache.mutex.RUnlock()

	var result []*types.GraphRelation
	for _, rel := range s.graphCache.relations {
		if rel.Strength >= minStrength {
			result = append(result, rel)
		}
	}
	return result, nil
}

// GetRelationStatistics 获取关系统计信息
func (s *GraphService) GetRelationStatistics(
	ctx context.Context,
	namespace types.NameSpace,
) (*RelationStatistics, error) {
	s.graphCache.mutex.RLock()
	defer s.graphCache.mutex.RUnlock()

	if len(s.graphCache.relations) == 0 {
		return &RelationStatistics{}, nil
	}

	stats := &RelationStatistics{
		TotalRelations:       len(s.graphCache.relations),
		StrengthDistribution: make(map[int]int),
	}

	totalStrength := 0.0
	totalWeight := 0.0
	totalChunkCount := 0

	for _, rel := range s.graphCache.relations {
		totalStrength += rel.Strength
		totalWeight += rel.Weight
		totalChunkCount += len(rel.ChunkIDs)

		strengthBucket := int(rel.Strength)
		stats.StrengthDistribution[strengthBucket]++
	}

	stats.AverageStrength = totalStrength / float64(stats.TotalRelations)
	stats.AverageWeight = totalWeight / float64(stats.TotalRelations)
	stats.AverageChunkCount = float64(totalChunkCount) / float64(stats.TotalRelations)

	return stats, nil
}

// RelationStatistics 关系统计信息
type RelationStatistics struct {
	TotalRelations       int         // 总关系数
	AverageStrength      float64     // 平均强度
	AverageWeight        float64     // 平均权重
	AverageChunkCount    float64     // 平均每关系关联的文档块数
	StrengthDistribution map[int]int // 强度分布 (1-10 各有多少关系)
}

// GetRelatedChunks 获取与特定实体相关的所有文档块
func (s *GraphService) GetRelatedChunks(
	ctx context.Context,
	namespace types.NameSpace,
	entityTitle string,
) ([]string, error) {
	s.graphCache.mutex.RLock()
	defer s.graphCache.mutex.RUnlock()

	if node, exists := s.graphCache.nodes[entityTitle]; exists {
		return node.Chunks, nil
	}
	return []string{}, nil
}

// FindCommonChunks 查找两个实体共同出现的文档块
func (s *GraphService) FindCommonChunks(
	ctx context.Context,
	namespace types.NameSpace,
	entity1, entity2 string,
) ([]string, error) {
	s.graphCache.mutex.RLock()
	defer s.graphCache.mutex.RUnlock()

	node1, exists1 := s.graphCache.nodes[entity1]
	node2, exists2 := s.graphCache.nodes[entity2]

	if !exists1 || !exists2 {
		return []string{}, nil
	}

	return intersection(node1.Chunks, node2.Chunks), nil
}

// cleanJSONResponse 清理LLM响应中的JSON内容
// 处理markdown代码块、前后空白、多余文字等问题
func cleanJSONResponse(response string) string {
	// 去除首尾空白
	response = strings.TrimSpace(response)

	// 检查是否包含markdown代码块标记
	if strings.HasPrefix(response, "```json") || strings.HasPrefix(response, "```") {
		// 找到代码块结束标记
		start := strings.Index(response, "{")
		end := strings.LastIndex(response, "}")

		if start >= 0 && end > start {
			response = response[start : end+1]
		}
	}

	// 如果响应以 "结果:"、"输出:"、"返回:"等开头，截取JSON部分
	prefixes := []string{"结果：", "输出：", "返回：", "Result:", "Output:", "Return:"}
	for _, prefix := range prefixes {
		if idx := strings.Index(response, prefix); idx >= 0 {
			// 从prefix后找第一个{
			if jsonStart := strings.Index(response[idx:], "{"); jsonStart >= 0 {
				response = response[idx+jsonStart:]
			}
		}
	}

	// 如果最后有 ``` 或其他多余内容，截取到JSON结束
	if lastBrace := strings.LastIndex(response, "}"); lastBrace >= 0 {
		response = response[:lastBrace+1]
	}

	return strings.TrimSpace(response)
}

// ========================================
// 图谱与知识库关联查询方法
// ========================================

// GetChunksByGraphNodes 根据图谱节点名称获取关联的分片
func (s *GraphService) GetChunksByGraphNodes(ctx context.Context, kbID string, nodeNames []string) ([]*types.Chunk, error) {
	if s.graphQueryRepo == nil {
		return nil, fmt.Errorf("GraphQueryRepository 未初始化")
	}
	return s.graphQueryRepo.GetChunksByGraphNodes(ctx, kbID, nodeNames)
}

// GetKnowledgeByGraphNodes 根据图谱节点名称获取关联的知识条目
func (s *GraphService) GetKnowledgeByGraphNodes(ctx context.Context, kbID string, nodeNames []string) (*types.Knowledge, error) {
	if s.graphQueryRepo == nil {
		return nil, fmt.Errorf("GraphQueryRepository 未初始化")
	}
	return s.graphQueryRepo.GetKnowledgeByGraphNodes(ctx, kbID, nodeNames)
}

// GetGraphStats 获取图谱统计信息
func (s *GraphService) GetGraphStats(ctx context.Context, kbID string) (*interfaces.GraphStats, error) {
	if s.graphQueryRepo == nil {
		return nil, fmt.Errorf("GraphQueryRepository 未初始化")
	}
	return s.graphQueryRepo.GetGraphStats(ctx, kbID)
}

// findChunksByEntities 根据实体名称查找包含这些实体的 chunk
// 遍历所有启用 chunk 的内容，检查是否包含任一实体名称
func (s *GraphService) findChunksByEntities(
	ctx context.Context,
	kbID string,
	entity1, entity2 string,
) []string {
	// 获取该知识库的所有启用 chunk（限制数量以避免查询过大）
	chunks, err := s.chunkRepo.FindEnabledChunks(ctx, kbID, 1000)
	if err != nil {
		return []string{}
	}

	var matchedChunkIDs []string

	// 检查每个 chunk 内容是否包含实体名称
	for _, chunk := range chunks {
		content := chunk.Content
		// 检查是否包含 entity1
		if len(content) >= len(entity1) && containsEntity(content, entity1) {
			matchedChunkIDs = append(matchedChunkIDs, chunk.ID)
			break // 只需要匹配一个实体即可
		}
		// 如果没有匹配 entity1，检查 entity2
		if len(matchedChunkIDs) == 0 && len(content) >= len(entity2) && containsEntity(content, entity2) {
			matchedChunkIDs = append(matchedChunkIDs, chunk.ID)
			break
		}
	}

	return matchedChunkIDs
}

// containsEntity 检查 chunk 内容是否包含实体名称
func containsEntity(content, entityName string) bool {
	// 简单匹配：检查实体名称是否完整出现在内容中
	return strings.Contains(content, entityName)
}
