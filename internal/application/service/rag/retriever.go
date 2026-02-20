package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"link/internal/application/repository/retriever/milvus"
	"link/internal/types"
	"link/internal/types/interfaces"

	"github.com/cloudwego/eino/components/embedding"
)

// Retriever 检索服务
type Retriever struct {
	kbSettingRepo        interfaces.KBSettingRepository
	retrievalSettingRepo interfaces.RetrievalSettingRepository // 新增：检索设置仓储
	chunkRepo            interfaces.ChunkRepository
	embedder             embedding.Embedder
	milvusRetriever      *milvus.VectorRetriever // Milvus 向量检索器（可选）
	neo4jRepo            interfaces.Neo4jGraphRepository
	graphQueryRepo       interfaces.GraphQueryRepository
}

// RetrieveOptions 检索选项
type RetrieveOptions struct {
	TopK                int     // 返回结果数量
	SimilarityThreshold float64 // 相似度阈值
	RerankEnabled       bool    // 是否重排序（暂未实现）
	GraphEnabled        bool    // 是否使用知识图谱（暂未实现）
	Alpha               float32 // 混合检索中向量检索的权重（默认0.5）
}

// RetrieveResponse 检索响应
type RetrieveResponse struct {
	Results   []*RetrieveResult
	Query     string
	Mode      string
	Relations []*GraphRelationRes // 图谱关系（简化版）
}

// GraphRelationRes 简化的图谱关系
type GraphRelationRes struct {
	Source      string // 源实体
	Target      string // 目标实体
	Type        string // 关系类型
	Description string // 关系描述
}

// RetrieveResult 检索结果
type RetrieveResult struct {
	ChunkID       string
	KnowledgeID   string
	KBID          string
	Content       string
	ChunkIndex    int
	Score         float32
	MatchType     string // "vector", "keyword", "hybrid"
	StartPosition int
	EndPosition   int
}

// NewRetriever 创建检索服务
func NewRetriever(
	kbSettingRepo interfaces.KBSettingRepository,
	chunkRepo interfaces.ChunkRepository,
	embedder embedding.Embedder,
	milvusRetriever *milvus.VectorRetriever, // 可选，如果提供则使用 Milvus 向量检索
	neo4jRepo interfaces.Neo4jGraphRepository, // 可选，用于图谱检索
	graphQueryRepo interfaces.GraphQueryRepository, // 可选，用于图谱检索
) *Retriever {
	return &Retriever{
		kbSettingRepo:   kbSettingRepo,
		chunkRepo:       chunkRepo,
		embedder:        embedder,
		milvusRetriever: milvusRetriever,
		neo4jRepo:       neo4jRepo,
		graphQueryRepo:  graphQueryRepo,
	}
}

// Retrieve 统一检索接口
// 根据知识库设置自动选择检索模式并返回结果
func (r *Retriever) Retrieve(ctx context.Context, tenantID int64, kbID string, query string, opts *RetrieveOptions) (*RetrieveResponse, error) {
	// 1. 获取知识库设置
	setting, err := r.kbSettingRepo.FindByKBID(ctx, kbID)
	if err != nil {
		return nil, fmt.Errorf("获取知识库设置失败: %w", err)
	}

	// 从 settings_json 解析配置
	retrievalMode := "vector"  // 默认检索模式
	topK := 5                  // 默认 topK
	similarityThreshold := 0.7 // 默认阈值
	rerankEnabled := false
	graphEnabled := setting.GraphEnabled

	if setting.SettingsJSON != nil {
		var settingsMap map[string]interface{}
		if err := json.Unmarshal([]byte(*setting.SettingsJSON), &settingsMap); err == nil {
			if mode, ok := settingsMap["retrieval_mode"].(string); ok {
				retrievalMode = mode
			}
			if tk, ok := settingsMap["top_k"].(float64); ok {
				topK = int(tk)
			}
			if st, ok := settingsMap["similarity_threshold"].(float64); ok {
				similarityThreshold = st
			}
			if re, ok := settingsMap["rerank_enabled"].(bool); ok {
				rerankEnabled = re
			}
		}
	}

	// 2. 如果没有提供选项，使用知识库设置中的默认值
	if opts == nil {
		opts = &RetrieveOptions{
			TopK:                topK,
			SimilarityThreshold: similarityThreshold,
			RerankEnabled:       rerankEnabled,
			GraphEnabled:        graphEnabled,
			Alpha:               0.5,
		}
	}

	// 3. 确保 TopK 至少为 1
	if opts.TopK < 1 {
		opts.TopK = 5
	}

	// 4. 根据检索模式选择检索方法
	switch retrievalMode {
	case "vector", "vector_search":
		return r.vectorRetrieveWithEmbedding(ctx, tenantID, kbID, query, opts)
	case "bm25", "keyword", "keywords":
		return r.bm25Retrieve(ctx, tenantID, kbID, query, opts)
	case "hybrid":
		return r.hybridRetrieve(ctx, tenantID, kbID, query, opts)
	case "graph":
		// 图谱检索暂未实现
		return r.graphRetrieve(ctx, tenantID, kbID, query, opts)
	default:
		// 默认使用向量检索
		return r.vectorRetrieveWithEmbedding(ctx, tenantID, kbID, query, opts)
	}
}

// vectorRetrieveWithEmbedding 向量检索（优先使用 Milvus，回退到应用层）
func (r *Retriever) vectorRetrieveWithEmbedding(ctx context.Context, tenantID int64, kbID string, query string, opts *RetrieveOptions) (*RetrieveResponse, error) {
	// 方案 A: 如果 Milvus 可用，使用 Milvus 向量检索
	if r.milvusRetriever != nil {
		return r.vectorRetrieveWithMilvus(ctx, kbID, query, opts)
	}

	// 方案 B: 回退到应用层向量检索（当前实现）
	return r.vectorRetrieveInMemory(ctx, kbID, query, opts)
}

// vectorRetrieveWithMilvus 使用 Milvus 进行向量检索
func (r *Retriever) vectorRetrieveWithMilvus(ctx context.Context, kbID string, query string, opts *RetrieveOptions) (*RetrieveResponse, error) {
	// 1. 将 kbID (UUID string) 转换为 Milvus 需要的 int64
	kbIDInt, err := r.kbIDToInt64(kbID)
	if err != nil {
		return nil, fmt.Errorf("kbID 转换失败: %w", err)
	}

	// 2. 检查 Milvus collection 是否存在
	hasCollection, err := r.milvusRetriever.HasKnowledgeBase(ctx, kbIDInt)
	if err != nil {
		return nil, fmt.Errorf("检查 Milvus collection 失败: %w", err)
	}

	// 如果 collection 不存在，回退到应用层检索
	if !hasCollection {
		return r.vectorRetrieveInMemory(ctx, kbID, query, opts)
	}

	// 3. 使用 Milvus 进行向量搜索
	searchOpts := &milvus.SearchOptions{
		TopK:           opts.TopK * 3, // 获取更多候选用于融合
		ScoreThreshold: float32(opts.SimilarityThreshold),
		OutputFields:   []string{"document_id", "chunk_index", "content", "metadata"},
	}

	milvusResults, err := r.milvusRetriever.Search(ctx, kbIDInt, query, searchOpts)
	if err != nil {
		return nil, fmt.Errorf("Milvus 搜索失败: %w", err)
	}

	// 4. 将 Milvus 结果转换为 RetrieveResult 格式
	var results []*RetrieveResult
	for _, mr := range milvusResults {
		results = append(results, &RetrieveResult{
			ChunkID:       mr.ChunkID,
			KnowledgeID:   mr.KnowledgeID,
			KBID:          mr.KBID,
			Content:       mr.Content,
			ChunkIndex:    mr.ChunkIndex,
			Score:         mr.Score,
			MatchType:     "vector",
			StartPosition: int(mr.StartAt),
			EndPosition:   int(mr.EndAt),
		})
	}

	// 5. 按 TopK 截取
	results = r.sortAndTrimResults(results, opts.TopK)

	return &RetrieveResponse{
		Results: results,
		Query:   query,
		Mode:    "vector",
	}, nil
}

// vectorRetrieveInMemory 应用层向量检索（回退方案）
func (r *Retriever) vectorRetrieveInMemory(ctx context.Context, kbID string, query string, opts *RetrieveOptions) (*RetrieveResponse, error) {
	// 1. 获取启用的分块（获取更多候选结果以提高检索质量）
	candidateChunks, err := r.chunkRepo.FindEnabledChunks(ctx, kbID, opts.TopK*5)
	if err != nil {
		return nil, fmt.Errorf("获取分块失败: %w", err)
	}

	if len(candidateChunks) == 0 {
		return &RetrieveResponse{
			Results: []*RetrieveResult{},
			Query:   query,
			Mode:    "vector",
		}, nil
	}

	// 2. 将查询文本向量化
	queryVec, err := r.embedQueryText(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("查询向量化失败: %w", err)
	}

	// 3. 将所有分块内容向量化（批量）
	chunkTexts := make([]string, len(candidateChunks))
	for i, chunk := range candidateChunks {
		chunkTexts[i] = chunk.Content
	}

	chunkEmbeddings, err := r.embedTexts(ctx, chunkTexts)
	if err != nil {
		return nil, fmt.Errorf("分块向量化失败: %w", err)
	}

	// 4. 计算相似度并过滤结果
	var results []*RetrieveResult
	for i, chunk := range candidateChunks {
		score := r.calculateCosineSimilarity(queryVec, chunkEmbeddings[i])
		if float64(score) >= opts.SimilarityThreshold {
			results = append(results, &RetrieveResult{
				ChunkID:       chunk.ID,
				KnowledgeID:   chunk.KnowledgeID,
				KBID:          chunk.KBID,
				Content:       chunk.Content,
				ChunkIndex:    chunk.ChunkIndex,
				Score:         score,
				MatchType:     "vector",
				StartPosition: chunk.StartAt,
				EndPosition:   chunk.EndAt,
			})
		}
	}

	// 5. 按相似度排序并返回 TopK
	results = r.sortAndTrimResults(results, opts.TopK)

	return &RetrieveResponse{
		Results: results,
		Query:   query,
		Mode:    "vector",
	}, nil
}

// bm25Retrieve BM25 关键词检索
func (r *Retriever) bm25Retrieve(ctx context.Context, tenantID int64, kbID string, query string, opts *RetrieveOptions) (*RetrieveResponse, error) {
	// 1. 获取启用的分块
	chunks, err := r.chunkRepo.FindEnabledChunks(ctx, kbID, opts.TopK*10)
	if err != nil {
		return nil, fmt.Errorf("获取分块失败: %w", err)
	}

	if len(chunks) == 0 {
		return &RetrieveResponse{
			Results: []*RetrieveResult{},
			Query:   query,
			Mode:    "bm25",
		}, nil
	}

	// 2. 对查询进行分词
	queryWords := tokenize(query)
	if len(queryWords) == 0 {
		return &RetrieveResponse{
			Results: []*RetrieveResult{},
			Query:   query,
			Mode:    "bm25",
		}, nil
	}

	// 3. 计算文档统计信息（用于 BM25）
	totalDocs := len(chunks)
	avgDocLen := r.calculateAverageDocLength(chunks)

	// 4. 构建词频统计（简化版 IDF）
	docFreq := make(map[string]int) // 包含某个词的文档数量
	for _, chunk := range chunks {
		contentWords := tokenize(chunk.Content)
		uniqueWords := make(map[string]bool)
		for _, word := range contentWords {
			uniqueWords[word] = true
		}
		for word := range uniqueWords {
			docFreq[word]++
		}
	}

	// 5. 计算每个分块的 BM25 分数
	var results []*RetrieveResult
	for _, chunk := range chunks {
		score := r.calculateBM25Score(queryWords, chunk.Content, docFreq, totalDocs, avgDocLen)
		if score > 0 {
			results = append(results, &RetrieveResult{
				ChunkID:       chunk.ID,
				KnowledgeID:   chunk.KnowledgeID,
				KBID:          chunk.KBID,
				Content:       chunk.Content,
				ChunkIndex:    chunk.ChunkIndex,
				Score:         score,
				MatchType:     "keyword",
				StartPosition: chunk.StartAt,
				EndPosition:   chunk.EndAt,
			})
		}
	}

	// 6. 按 BM25 分数排序并返回 TopK
	results = r.sortAndTrimResults(results, opts.TopK)

	return &RetrieveResponse{
		Results: results,
		Query:   query,
		Mode:    "bm25",
	}, nil
}

// hybridRetrieve 混合检索（向量 + BM25）
func (r *Retriever) hybridRetrieve(ctx context.Context, tenantID int64, kbID string, query string, opts *RetrieveOptions) (*RetrieveResponse, error) {
	// 1. 执行向量检索
	vectorResp, err := r.vectorRetrieveWithEmbedding(ctx, tenantID, kbID, query, opts)
	if err != nil {
		return nil, fmt.Errorf("向量检索失败: %w", err)
	}

	// 2. 执行 BM25 检索
	bm25Resp, err := r.bm25Retrieve(ctx, tenantID, kbID, query, opts)
	if err != nil {
		return nil, fmt.Errorf("关键词检索失败: %w", err)
	}

	// 3. 使用 RRF 融合结果
	fusedResults := r.reciprocalRankFusion(vectorResp.Results, bm25Resp.Results, opts.Alpha)

	// 4. 返回 TopK
	results := r.sortAndTrimResults(fusedResults, opts.TopK)

	return &RetrieveResponse{
		Results: results,
		Query:   query,
		Mode:    "hybrid",
	}, nil
}

// GraphRetrieve 公开的知识图谱检索方法
// 实现流程：从查询中提取实体 -> 获取关联实体 -> 根据 strength 排序拿到 top5 -> 去数据库获取具体内容
func (r *Retriever) GraphRetrieve(ctx context.Context, tenantID int64, kbID string, query string, opts *RetrieveOptions) (*RetrieveResponse, error) {
	return r.graphRetrieve(ctx, tenantID, kbID, query, opts)
}

// graphRetrieve 知识图谱检索
// 实现流程：从查询中提取实体 -> 获取关联实体 -> 根据 strength 排序拿到 top5 -> 去数据库获取具体内容
func (r *Retriever) graphRetrieve(ctx context.Context, tenantID int64, kbID string, query string, opts *RetrieveOptions) (*RetrieveResponse, error) {
	log.Printf("[GraphRetriever] START: kbID=%s, query=%s", kbID, query)

	// 检查必要的仓储是否可用
	if r.neo4jRepo == nil || r.graphQueryRepo == nil {
		log.Printf("[GraphRetriever] ERROR: neo4jRepo or graphQueryRepo is nil")
		return &RetrieveResponse{
			Results: []*RetrieveResult{},
			Query:   query,
			Mode:    "graph",
		}, fmt.Errorf("知识图谱检索仓储未配置")
	}

	// 1. 从查询中提取实体
	entities := r.extractEntities(query)
	if len(entities) == 0 {
		log.Printf("[GraphRetriever] WARNING: no entities extracted from query")
		return &RetrieveResponse{
			Results: []*RetrieveResult{},
			Query:   query,
			Mode:    "graph",
		}, nil
	}
	log.Printf("[GraphRetriever] Extracted entities: %v", entities)

	// 2. 从 Neo4j 获取关联实体
	namespace := types.NameSpace{
		KBID: kbID,
		// Knowledge 留空表示查询整个知识库的图谱
	}

	graphData, err := r.neo4jRepo.SearchNodeV2(ctx, namespace, entities)
	if err != nil {
		log.Printf("[GraphRetriever] ERROR: SearchNodeV2 failed: %v", err)
		return nil, fmt.Errorf("图谱节点查询失败: %w", err)
	}

	if graphData == nil || len(graphData.Relation) == 0 {
		log.Printf("[GraphRetriever] WARNING: no relations found for entities")
		return &RetrieveResponse{
			Results: []*RetrieveResult{},
			Query:   query,
			Mode:    "graph",
		}, nil
	}
	log.Printf("[GraphRetriever] Found %d relations", len(graphData.Relation))

	// 3. 根据关系 strength 排序，收集 chunk_ids
	type ChunkWithStrength struct {
		ChunkID  string
		Strength float64
		Relation *types.GraphRelation
	}

	var chunksWithStrength []ChunkWithStrength
	seenChunkIDs := make(map[string]bool)

	for _, rel := range graphData.Relation {
		// 从关系的 chunk_ids 中提取所有关联的 chunk
		for _, chunkID := range rel.ChunkIDs {
			if chunkID == "" {
				continue
			}
			// 去重：如果同一个 chunk 出现在多个关系中，取最高的 strength
			if seenChunkIDs[chunkID] {
				// 更新已存在的 chunk 的 strength（如果当前更高）
				for i, cws := range chunksWithStrength {
					if cws.ChunkID == chunkID && rel.Strength > cws.Strength {
						chunksWithStrength[i].Strength = rel.Strength
						chunksWithStrength[i].Relation = rel
					}
				}
				continue
			}
			seenChunkIDs[chunkID] = true
			chunksWithStrength = append(chunksWithStrength, ChunkWithStrength{
				ChunkID:  chunkID,
				Strength: rel.Strength,
				Relation: rel,
			})
		}
	}

	// 按 strength 降序排序
	sort.Slice(chunksWithStrength, func(i, j int) bool {
		return chunksWithStrength[i].Strength > chunksWithStrength[j].Strength
	})

	// 取 topK 个 chunk ID（如果 TopK 小于等于 0，默认取 5）
	topK := opts.TopK
	if topK <= 0 {
		topK = 5
	}
	if len(chunksWithStrength) > topK {
		chunksWithStrength = chunksWithStrength[:topK]
	}

	// 提取 chunk ID 列表
	chunkIDs := make([]string, len(chunksWithStrength))
	for i, cws := range chunksWithStrength {
		chunkIDs[i] = cws.ChunkID
	}
	log.Printf("[GraphRetriever] Selected top %d chunks by strength: %v", len(chunkIDs), chunkIDs)

	// 4. 根据 chunk ID 列表从数据库获取具体内容
	chunks, err := r.graphQueryRepo.GetChunksByIDs(ctx, kbID, chunkIDs)
	if err != nil {
		log.Printf("[GraphRetriever] ERROR: GetChunksByIDs failed: %v", err)
		return nil, fmt.Errorf("获取分块内容失败: %w", err)
	}

	if len(chunks) == 0 {
		log.Printf("[GraphRetriever] WARNING: no chunks found in database")
		return &RetrieveResponse{
			Results: []*RetrieveResult{},
			Query:   query,
			Mode:    "graph",
		}, nil
	}

	// 5. 构建检索结果
	// 创建 chunkID -> strength 的映射
	chunkStrengthMap := make(map[string]float64)
	for _, cws := range chunksWithStrength {
		chunkStrengthMap[cws.ChunkID] = cws.Strength
	}

	var results []*RetrieveResult
	for _, chunk := range chunks {
		strength := chunkStrengthMap[chunk.ID]
		results = append(results, &RetrieveResult{
			ChunkID:       chunk.ID,
			KnowledgeID:   chunk.KnowledgeID,
			KBID:          chunk.KBID,
			Content:       chunk.Content,
			ChunkIndex:    chunk.ChunkIndex,
			Score:         float32(strength), // 使用 strength 作为分数
			MatchType:     "graph",
			StartPosition: chunk.StartAt,
			EndPosition:   chunk.EndAt,
		})
	}

	// 按分数降序排序
	results = r.sortAndTrimResults(results, opts.TopK)

	// 构建简化关系列表 - 从节点中获取名称
	var relationRes []*GraphRelationRes
	nodeNameMap := make(map[string]string) // ID -> Name
	for _, node := range graphData.Node {
		nodeNameMap[node.ID] = node.Name
	}

	for _, rel := range graphData.Relation {
		relationRes = append(relationRes, &GraphRelationRes{
			Source:      rel.Source,
			Target:      rel.Target,
			Type:        rel.Type,
			Description: rel.Description,
		})
	}

	log.Printf("[GraphRetriever] COMPLETE: returned %d results, %d relations", len(results), len(relationRes))
	return &RetrieveResponse{
		Results:   results,
		Query:     query,
		Mode:      "graph",
		Relations: relationRes,
	}, nil
}

// extractEntities 从查询文本中提取实体
// 使用简单的分词方法，提取中文词汇和英文单词
func (r *Retriever) extractEntities(query string) []string {
	// 使用现有的 tokenize 方法进行分词
	words := tokenize(query)

	// 过滤掉停用词和单字符
	stopWords := map[string]bool{
		"的": true, "了": true, "在": true, "是": true, "我": true,
		"有": true, "和": true, "就": true, "不": true, "人": true,
		"都": true, "一": true, "一个": true, "上": true, "也": true,
		"很": true, "到": true, "说": true, "要": true, "去": true,
		"你": true, "会": true, "着": true, "没有": true, "看": true,
		"好": true, "自己": true, "这": true, "the": true, "a": true,
		"an": true, "and": true, "or": true, "but": true, "is": true,
		"are": true, "was": true, "were": true, "of": true, "to": true,
		"in": true, "on": true, "at": true, "for": true, "with": true,
	}

	var entities []string
	seen := make(map[string]bool)

	for _, word := range words {
		// 过滤条件：非停用词、长度大于 1、不是纯数字
		if !stopWords[word] && len(word) > 1 && !isNumeric(word) {
			if !seen[word] {
				seen[word] = true
				entities = append(entities, word)
			}
		}
	}

	// 限制实体数量，避免查询过大
	if len(entities) > 10 {
		entities = entities[:10]
	}

	return entities
}

// isNumeric 检查字符串是否为纯数字
func isNumeric(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// ========================================
// 辅助方法
// ========================================

// embedQueryText 将查询文本向量化
func (r *Retriever) embedQueryText(ctx context.Context, query string) ([]float32, error) {
	embeddings, err := r.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 || len(embeddings[0]) == 0 {
		return nil, fmt.Errorf("向量化结果为空")
	}

	// 将 []float64 转换为 []float32
	vec := make([]float32, len(embeddings[0]))
	for i, v := range embeddings[0] {
		vec[i] = float32(v)
	}
	return vec, nil
}

// embedTexts 批量将文本向量化
func (r *Retriever) embedTexts(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	embeddings, err := r.embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, err
	}

	// 将 [][]float64 转换为 [][]float32
	result := make([][]float32, len(embeddings))
	for i, emb := range embeddings {
		result[i] = make([]float32, len(emb))
		for j, v := range emb {
			result[i][j] = float32(v)
		}
	}
	return result, nil
}

// calculateCosineSimilarity 计算余弦相似度
func (r *Retriever) calculateCosineSimilarity(vec1, vec2 []float32) float32 {
	if len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct float32
	var norm1 float32
	var norm2 float32

	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(norm1))) * float32(math.Sqrt(float64(norm2))))
}

// tokenize 文本分词（简化版，支持中文）
func tokenize(text string) []string {
	// 移除标点符号和空格
	var words []string
	var currentWord strings.Builder

	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			// 中文字符单独作为一个词
			if currentWord.Len() > 0 {
				words = append(words, strings.ToLower(currentWord.String()))
				currentWord.Reset()
			}
			words = append(words, string(r))
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			currentWord.WriteRune(r)
		} else if unicode.IsSpace(r) {
			if currentWord.Len() > 0 {
				words = append(words, strings.ToLower(currentWord.String()))
				currentWord.Reset()
			}
		}
	}

	if currentWord.Len() > 0 {
		words = append(words, strings.ToLower(currentWord.String()))
	}

	return words
}

// calculateBM25Score 计算 BM25 分数
func (r *Retriever) calculateBM25Score(queryWords []string, content string, docFreq map[string]int, totalDocs int, avgDocLen float64) float32 {
	k1 := 1.5 // 调节词频饱和度
	b := 0.75 // 调节文档长度归一化程度

	contentWords := tokenize(content)
	docLen := float64(len(contentWords))

	// 计算词频
	wordFreq := make(map[string]int)
	for _, word := range contentWords {
		wordFreq[word]++
	}

	var score float32
	for _, queryWord := range queryWords {
		freq := wordFreq[queryWord]
		if freq == 0 {
			continue
		}

		// IDF 计算（加1平滑）
		idf := float32(math.Log(float64(totalDocs-docFreq[queryWord]+1)/float64(docFreq[queryWord]+1) + 1))

		// TF 计算
		tf := float32((float64(freq) * (k1 + 1)) / (float64(freq) + k1*(1-b+b*docLen/avgDocLen)))

		score += idf * tf
	}

	return score
}

// calculateAverageDocLength 计算平均文档长度
func (r *Retriever) calculateAverageDocLength(chunks []*types.Chunk) float64 {
	if len(chunks) == 0 {
		return 500.0 // 默认值
	}

	totalLength := 0
	for _, chunk := range chunks {
		totalLength += len(chunk.Content)
	}
	return float64(totalLength) / float64(len(chunks))
}

// reciprocalRankFusion 倒数排名融合（RRF）
func (r *Retriever) reciprocalRankFusion(vectorResults, keywordResults []*RetrieveResult, alpha float32) []*RetrieveResult {
	k := 60.0 // RRF 常数

	// 构建分块 ID 到结果的映射
	vectorMap := make(map[string]*RetrieveResult)
	for _, result := range vectorResults {
		vectorMap[result.ChunkID] = result
	}

	keywordMap := make(map[string]*RetrieveResult)
	for _, result := range keywordResults {
		keywordMap[result.ChunkID] = result
	}

	// 合并所有分块 ID
	allChunkIDs := make(map[string]bool)
	for id := range vectorMap {
		allChunkIDs[id] = true
	}
	for id := range keywordMap {
		allChunkIDs[id] = true
	}

	// 按 RRF 算法计算新分数
	var fusedResults []*RetrieveResult
	for chunkID := range allChunkIDs {
		var score float32
		var matchType string

		// 计算向量检索排名分数
		if _, exists := vectorMap[chunkID]; exists {
			// 找到在向量结果中的排名
			rank := r.findRank(vectorResults, chunkID)
			score += alpha / float32(k+float64(rank))
			matchType = "vector"
		}

		// 计算关键词检索排名分数
		if _, exists := keywordMap[chunkID]; exists {
			rank := r.findRank(keywordResults, chunkID)
			score += (1 - alpha) / float32(k+float64(rank))
			if matchType == "" {
				matchType = "keyword"
			} else {
				matchType = "hybrid"
			}
		}

		// 使用向量结果作为基础数据
		var baseResult *RetrieveResult
		if vecResult, exists := vectorMap[chunkID]; exists {
			baseResult = vecResult
		} else {
			baseResult = keywordMap[chunkID]
		}

		// 创建融合后的结果
		fusedResult := &RetrieveResult{
			ChunkID:       baseResult.ChunkID,
			KnowledgeID:   baseResult.KnowledgeID,
			KBID:          baseResult.KBID,
			Content:       baseResult.Content,
			ChunkIndex:    baseResult.ChunkIndex,
			Score:         score,
			MatchType:     matchType,
			StartPosition: baseResult.StartPosition,
			EndPosition:   baseResult.EndPosition,
		}
		fusedResults = append(fusedResults, fusedResult)
	}

	return fusedResults
}

// findRank 查找结果在排序列表中的排名
func (r *Retriever) findRank(results []*RetrieveResult, chunkID string) int {
	for i, result := range results {
		if result.ChunkID == chunkID {
			return i + 1 // 排名从 1 开始
		}
	}
	return len(results) // 如果没找到，返回最后一名
}

// sortAndTrimResults 按分数排序并返回 TopK
func (r *Retriever) sortAndTrimResults(results []*RetrieveResult, topK int) []*RetrieveResult {
	if len(results) == 0 {
		return results
	}

	// 按分数降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 返回 TopK
	if len(results) > topK {
		return results[:topK]
	}
	return results
}

// ========================================
// 批量操作
// ========================================

// BatchRetrieve 批量检索
func (r *Retriever) BatchRetrieve(ctx context.Context, tenantID int64, kbID string, queries []string, opts *RetrieveOptions) ([]*RetrieveResponse, error) {
	responses := make([]*RetrieveResponse, len(queries))

	for i, query := range queries {
		resp, err := r.Retrieve(ctx, tenantID, kbID, query, opts)
		if err != nil {
			// 单个检索失败不影响其他检索
			resp = &RetrieveResponse{
				Results: []*RetrieveResult{},
				Query:   query,
				Mode:    "error",
			}
		}
		responses[i] = resp
	}

	return responses, nil
}

// ========================================
// Milvus 辅助方法
// ========================================

// kbIDToInt64 将 UUID 字符串转换为 int64（用于 Milvus collection 名称）
// 注意：这是一个简化实现，实际项目中需要维护一个 kbID 到 int64 的映射表
func (r *Retriever) kbIDToInt64(kbID string) (int64, error) {
	// 方案 1: 如果 kbID 本身就是数字字符串，直接转换
	if val, err := strconv.ParseInt(kbID, 10, 64); err == nil {
		return val, nil
	}

	// 方案 2: 使用哈希值（可能会冲突，不推荐生产环境）
	// hash := fnv.New64a()
	// hash.Write([]byte(kbID))
	// return int64(hash.Sum64()), nil

	// 方案 3: 从数据库或缓存中获取映射（推荐）
	// 这里应该从 knowledge_base 表的 milvus_collection_id 字段获取
	// 或者维护一个 kb_id -> milvus_kb_id 的映射表

	return 0, fmt.Errorf("kbID '%s' 不是有效的 int64，需要实现映射机制", kbID)
}
