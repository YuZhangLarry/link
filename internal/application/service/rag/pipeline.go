package rag

import (
	"context"
	"log"

	"link/internal/models/chat"
	"link/internal/types/interfaces"

	"github.com/cloudwego/eino/components/embedding"
)

// ========================================
// RAG Pipeline 配置
// ========================================

// PipelineConfig RAG 管道配置（简化版，仅包含检索相关设置）
type PipelineConfig struct {
	// 检索模式（可多选，向量检索必选）
	RetrievalModes []string // 检索模式：vector(必选), bm25, graph

	// 检索参数
	VectorTopK          int     // 向量检索返回数量
	KeywordTopK         int     // 关键词检索返回数量
	GraphTopK           int     // 图谱检索返回数量
	SimilarityThreshold float64 // 相似度阈值
	Alpha               float32 // 向量检索权重（混合检索用）
}

// DefaultPipelineConfig 默认管道配置
func DefaultPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		RetrievalModes:      []string{"vector"}, // 默认仅向量检索
		VectorTopK:          15,
		KeywordTopK:         15,
		GraphTopK:           10,
		SimilarityThreshold: 0.0,
		Alpha:               0.6,
	}
}

// ========================================
// RAG Pipeline 管道
// ========================================

// Pipeline RAG 管道（简化版，仅包含检索）
type Pipeline struct {
	retriever       *Retriever
	embedder        embedding.Embedder
	milvusRetriever interface{} // *milvus.VectorRetriever
}

// NewPipeline 创建 RAG 管道
func NewPipeline(
	chatConfig *chat.ChatConfig,
	kbSettingRepo interfaces.KBSettingRepository,
	chunkRepo interfaces.ChunkRepository,
	embedder embedding.Embedder,
	milvusRetriever interface{}, // *milvus.VectorRetriever
	neo4jRepo interfaces.Neo4jGraphRepository,
	graphQueryRepo interfaces.GraphQueryRepository,
) (*Pipeline, error) {
	// 创建检索器
	retriever := NewRetriever(
		kbSettingRepo,
		chunkRepo,
		embedder,
		nil, // milvusRetriever, // 类型转换问题，需要外部处理
		neo4jRepo,
		graphQueryRepo,
	)

	return &Pipeline{
		retriever:       retriever,
		embedder:        embedder,
		milvusRetriever: milvusRetriever,
	}, nil
}

// NewPipelineWithReranker 创建 RAG 管道（简化版，不再使用重排模型）
func NewPipelineWithReranker(
	chatConfig *chat.ChatConfig,
	kbSettingRepo interfaces.KBSettingRepository,
	chunkRepo interfaces.ChunkRepository,
	embedder embedding.Embedder,
	milvusRetriever interface{},
	neo4jRepo interfaces.Neo4jGraphRepository,
	graphQueryRepo interfaces.GraphQueryRepository,
	rerankEmbedder RerankEmbedder,
) (*Pipeline, error) {
	// 重排模型参数已弃用，直接调用 NewPipeline
	return NewPipeline(
		chatConfig,
		kbSettingRepo,
		chunkRepo,
		embedder,
		milvusRetriever,
		neo4jRepo,
		graphQueryRepo,
	)
}

// ========================================
// 核心方法：执行完整 RAG 流程
// ========================================

// Execute 执行完整的 RAG 流程（简化版：仅包含检索）
func (p *Pipeline) Execute(
	ctx context.Context,
	tenantID int64,
	kbID string,
	query string,
	conversationHistory string,
	config *PipelineConfig,
) (*PipelineResult, error) {
	if config == nil {
		config = DefaultPipelineConfig()
	}

	// 确保 vector 在检索模式中（必选）
	hasVector := false
	for _, mode := range config.RetrievalModes {
		if mode == "vector" {
			hasVector = true
			break
		}
	}
	if !hasVector {
		config.RetrievalModes = append([]string{"vector"}, config.RetrievalModes...)
	}

	log.Printf("[RAG-Pipeline] START: kbID=%s, query=%s, modes=%v", kbID, query, config.RetrievalModes)

	result := &PipelineResult{
		Query:         query,
		FinalQuery:    query,
		Stages:        make(map[string]*StageResult),
		RetrievedDocs: make([]*RetrieveResult, 0),
	}

	// ========================================
	// 检索阶段：根据配置的检索模式执行
	// ========================================
	var allResults [][]*RetrieveResult
	var sourceTypes []string

	for _, mode := range config.RetrievalModes {
		switch mode {
		case "vector":
			// 向量检索
			vectorOpts := &RetrieveOptions{
				TopK:                config.VectorTopK,
				SimilarityThreshold: config.SimilarityThreshold,
				RerankEnabled:       false,
				GraphEnabled:        false,
				Alpha:               config.Alpha,
			}
			resp, err := p.retriever.vectorRetrieveWithEmbedding(ctx, tenantID, kbID, query, vectorOpts)
			if err == nil && len(resp.Results) > 0 {
				allResults = append(allResults, resp.Results)
			}
			sourceTypes = append(sourceTypes, "vector")

		case "bm25", "keyword":
			// BM25 检索
			keywordOpts := &RetrieveOptions{
				TopK:                config.KeywordTopK,
				SimilarityThreshold: config.SimilarityThreshold,
				RerankEnabled:       false,
				GraphEnabled:        false,
				Alpha:               config.Alpha,
			}
			resp, err := p.retriever.bm25Retrieve(ctx, tenantID, kbID, query, keywordOpts)
			if err == nil && len(resp.Results) > 0 {
				allResults = append(allResults, resp.Results)
			}
			sourceTypes = append(sourceTypes, "keyword")

		case "graph":
			// 图谱检索
			graphOpts := &RetrieveOptions{
				TopK:                config.GraphTopK,
				SimilarityThreshold: config.SimilarityThreshold,
				RerankEnabled:       false,
				GraphEnabled:        true,
				Alpha:               config.Alpha,
			}
			resp, err := p.retriever.GraphRetrieve(ctx, tenantID, kbID, query, graphOpts)
			if err == nil && len(resp.Results) > 0 {
				allResults = append(allResults, resp.Results)
				result.GraphRelations = resp.Relations
			}
			sourceTypes = append(sourceTypes, "graph")
		}
	}

	result.Stages["retriever"] = &StageResult{
		Name:       "retriever",
		Input:      query,
		Output:     map[string]interface{}{"result_count": len(allResults), "sources": sourceTypes},
		Success:    len(allResults) > 0,
		InputCount: 1,
		OutputCount: func() (total int) {
			for _, list := range allResults {
				total += len(list)
			}
			return
		}(),
	}

	if len(allResults) == 0 {
		log.Printf("[RAG-Pipeline] No results retrieved")
		result.Success = true
		return result, nil
	}

	// 合并所有检索结果
	finalResults := mergeAllResults(allResults)

	result.RetrievedDocs = finalResults
	result.SourceTypes = sourceTypes
	result.Success = true

	log.Printf("[RAG-Pipeline] COMPLETE: returned %d documents, sources=%v", len(finalResults), sourceTypes)

	return result, nil
}

// ========================================
// 快捷方法
// ========================================

// SimpleExecute 简单执行 RAG（使用默认配置）
func (p *Pipeline) SimpleExecute(
	ctx context.Context,
	tenantID int64,
	kbID string,
	query string,
) (*PipelineResult, error) {
	return p.Execute(ctx, tenantID, kbID, query, "", DefaultPipelineConfig())
}

// ExecuteWithMode 指定检索模式执行
func (p *Pipeline) ExecuteWithMode(
	ctx context.Context,
	tenantID int64,
	kbID string,
	query string,
	retrievalMode string,
	topK int,
) (*PipelineResult, error) {
	config := DefaultPipelineConfig()
	// 将 retrievalMode 转换为 RetrievalModes 数组
	switch retrievalMode {
	case "vector":
		config.RetrievalModes = []string{"vector"}
	case "bm25", "keyword":
		config.RetrievalModes = []string{"vector", "bm25"}
	case "graph":
		config.RetrievalModes = []string{"vector", "graph"}
	case "hybrid":
		config.RetrievalModes = []string{"vector", "bm25"}
	default:
		config.RetrievalModes = []string{"vector"}
	}
	config.VectorTopK = topK * 3
	config.KeywordTopK = topK * 3
	return p.Execute(ctx, tenantID, kbID, query, "", config)
}

// ========================================
// 结果类型
// ========================================

// PipelineResult RAG 管道执行结果
type PipelineResult struct {
	Query          string                  // 原始查询
	FinalQuery     string                  // 最终使用的查询（可能被重写）
	RetrievedDocs  []*RetrieveResult       // 检索到的文档
	GraphRelations []*GraphRelationRes     // 图谱关系
	SourceTypes    []string                // 来源类型列表
	Success        bool                    // 执行是否成功
	Stages         map[string]*StageResult // 各阶段执行详情
}

// GetContexts 获取检索到的文档内容（用于生成）
func (r *PipelineResult) GetContexts() []string {
	contexts := make([]string, len(r.RetrievedDocs))
	for i, doc := range r.RetrievedDocs {
		contexts[i] = doc.Content
	}
	return contexts
}

// GetContextsWithScore 获取带分数的文档内容
func (r *PipelineResult) GetContextsWithScore() []map[string]interface{} {
	result := make([]map[string]interface{}, len(r.RetrievedDocs))
	for i, doc := range r.RetrievedDocs {
		result[i] = map[string]interface{}{
			"content":  doc.Content,
			"score":    doc.Score,
			"chunk_id": doc.ChunkID,
			"source":   doc.MatchType,
		}
	}
	return result
}

// StageResult 阶段执行结果
type StageResult struct {
	Name        string      // 阶段名称
	Input       interface{} // 输入
	Output      interface{} // 输出
	Success     bool        // 是否成功
	Error       string      // 错误信息
	InputCount  int         // 输入数量
	OutputCount int         // 输出数量
}

// ========================================
// 辅助函数
// ========================================

// formatError 格式化错误信息
func formatError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// mergeAllResults 合并多个结果列表并去重
func mergeAllResults(resultLists [][]*RetrieveResult) []*RetrieveResult {
	seen := make(map[string]*RetrieveResult)

	for _, results := range resultLists {
		for _, result := range results {
			if existing, exists := seen[result.ChunkID]; exists {
				// 保留分数更高的
				if result.Score > existing.Score {
					seen[result.ChunkID] = result
				}
			} else {
				seen[result.ChunkID] = result
			}
		}
	}

	// 转换为列表
	merged := make([]*RetrieveResult, 0, len(seen))
	for _, result := range seen {
		merged = append(merged, result)
	}

	// 按分数排序
	sortResults(merged)

	return merged
}

// sortResults 按分数排序结果
func sortResults(results []*RetrieveResult) {
	// 这里简单实现，实际可以使用 sort.Slice
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Score < results[j].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// ========================================
// 工厂函数
// ========================================

// NewPipelineFromComponents 从已有组件创建管道
func NewPipelineFromComponents(
	retriever *Retriever,
) *Pipeline {
	return &Pipeline{
		retriever: retriever,
	}
}
