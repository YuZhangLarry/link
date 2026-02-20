// Package tool 提供 RAG 查询工具
package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// ========================================
// RAG 查询工具
// ========================================

// RAGQueryService RAG 服务接口（避免循环依赖）
type RAGQueryService interface {
	Query(ctx context.Context, req *RAGQueryRequest) (*RAGQueryResult, error)
}

// 全局 RAG 服务实例
var ragService RAGQueryService

// InitRAGQueryTool 初始化 RAG 查询工具
func InitRAGQueryTool(service RAGQueryService) {
	ragService = service
}

// SetRAGService 设置 RAG 服务（用于测试）
func SetRAGService(service RAGQueryService) {
	ragService = service
}

// RAGQueryRequest RAG 查询请求
type RAGQueryRequest struct {
	// Query 查询内容
	Query string `json:"query" jsonschema:"required,description=用户的问题或查询内容"`

	// KBID 知识库ID，0表示查询所有启用的知识库
	KBID int64 `json:"kb_id" jsonschema:"description=知识库ID，0或不传表示查询所有启用的知识库"`

	// TopK 返回结果数量，默认5
	TopK int `json:"top_k" jsonschema:"description=返回结果数量，默认5，范围1-20"`

	// RetrievalMode 检索模式：vector(向量)、bm25(关键词)、hybrid(混合)、graph(图谱)
	RetrievalMode string `json:"retrieval_mode" jsonschema:"description=检索模式：vector/bm25/hybrid/graph，默认hybrid"`

	// MinScore 最小相似度阈值，默认0.7
	MinScore float64 `json:"min_score" jsonschema:"description=最小相似度阈值，范围0-1，默认0.7"`

	// EnableRerank 是否启用重排序，默认false
	EnableRerank bool `json:"enable_rerank" jsonschema:"description=是否启用重排序，默认false"`
}

// RAGQueryResult RAG 查询结果
type RAGQueryResult struct {
	// Answer 基于检索内容生成的答案
	Answer string `json:"answer"`

	// Chunks 检索到的文档片段
	Chunks []DocumentChunk `json:"chunks"`

	// Count 检索到的片段数量
	Count int `json:"count"`

	// Query 原始查询
	Query string `json:"query"`

	// KBID 使用的知识库ID
	KBID int64 `json:"kb_id"`

	// Latency 查询耗时（毫秒）
	Latency int64 `json:"latency_ms"`

	// RetrievalMode 实际使用的检索模式
	RetrievalMode string `json:"retrieval_mode"`

	// HasAnswer 是否有答案
	HasAnswer bool `json:"has_answer"`
}

// DocumentChunk 文档片段
type DocumentChunk struct {
	// Content 片段内容
	Content string `json:"content"`

	// Score 相似度分数
	Score float64 `json:"score"`

	// Source 来源文档
	Source string `json:"source"`

	// DocumentID 文档ID
	DocumentID int64 `json:"document_id"`

	// ChunkIndex 片段索引
	ChunkIndex int `json:"chunk_index"`

	// Highlight 高亮内容
	Highlight string `json:"highlight,omitempty"`

	// Metadata 额外元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewRAGQueryTool 创建 RAG 查询工具
func NewRAGQueryTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"rag_query",
		`使用 RAG（检索增强生成）技术从知识库中查询信息并生成答案。

这是最强大的查询工具，能够：
1. 从用户上传的文档中检索相关内容
2. 使用向量相似度、关键词匹配、知识图谱等多种检索方式
3. 基于检索结果生成准确的答案
4. 标注信息来源，便于验证

支持多种检索模式：
- vector: 向量检索，基于语义相似度，适合概念性查询
- bm25: 关键词检索，基于精确匹配，适合事实性查询
- hybrid: 混合检索，结合向量和关键词，推荐使用
- graph: 图谱检索，基于实体关系，适合关联性查询

适用场景：
- 查询文档内容："API文档中关于认证的部分"
- 概念解释："什么是微服务架构？"
- 关联查询："张三负责哪些项目？"
- 综合分析："对比分析两个方案的优劣"

参数说明：
- query: 查询内容（必需）
- kb_id: 知识库ID（可选，0表示查询所有）
- top_k: 返回结果数量（可选，默认5）
- retrieval_mode: 检索模式（可选，默认hybrid）
- enable_rerank: 是否启用重排序（可选，默认false）`,
		ragQuery,
	)
}

// ragQuery 执行 RAG 查询
func ragQuery(ctx context.Context, req *RAGQueryRequest) (*RAGQueryResult, error) {
	startTime := time.Now()

	// 1. 参数验证
	if req.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// 设置默认值
	if req.TopK <= 0 {
		req.TopK = 5
	}
	if req.TopK > 20 {
		req.TopK = 20
	}
	if req.MinScore <= 0 {
		req.MinScore = 0.7
	}
	if req.RetrievalMode == "" {
		req.RetrievalMode = "hybrid"
	}

	// 验证检索模式
	validModes := map[string]bool{
		"vector": true,
		"bm25":   true,
		"hybrid": true,
		"graph":  true,
	}
	if !validModes[req.RetrievalMode] {
		return nil, fmt.Errorf("invalid retrieval_mode: %s, must be one of: vector, bm25, hybrid, graph", req.RetrievalMode)
	}

	// 2. 检查 RAG 服务是否已初始化
	if ragService == nil {
		// 服务未初始化，返回模拟数据
		return mockRAGQuery(ctx, req, startTime)
	}

	// 3. 调用真实的 RAG 服务
	result, err := ragService.Query(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("RAG query failed: %w", err)
	}

	// 4. 更新耗时
	result.Latency = time.Since(startTime).Milliseconds()
	result.Query = req.Query
	result.KBID = req.KBID
	result.RetrievalMode = req.RetrievalMode

	return result, nil
}

// mockRAGQuery 模拟 RAG 查询（服务未初始化时的降级方案）
func mockRAGQuery(ctx context.Context, req *RAGQueryRequest, startTime time.Time) (*RAGQueryResult, error) {
	// 模拟检索延迟
	latency := 50 + time.Since(startTime).Milliseconds()

	// 根据查询内容生成模拟结果
	chunks := generateMockChunks(req.Query, req.TopK)

	// 生成模拟答案
	answer := generateMockAnswer(req.Query, chunks)

	return &RAGQueryResult{
		Answer:        answer,
		Chunks:        chunks,
		Count:         len(chunks),
		Query:         req.Query,
		KBID:          req.KBID,
		Latency:       latency,
		RetrievalMode: req.RetrievalMode,
		HasAnswer:     true,
	}, nil
}

// generateMockChunks 生成模拟片段
func generateMockChunks(query string, count int) []DocumentChunk {
	chunks := make([]DocumentChunk, 0, count)

	for i := 0; i < count && i < 5; i++ {
		score := 0.95 - float64(i)*0.05
		chunks = append(chunks, DocumentChunk{
			Content:    fmt.Sprintf("这是与查询「%s」相关的文档内容片段 #%d。实际项目中，这里会包含从知识库检索到的真实文档内容。", query, i+1),
			Score:      score,
			Source:     fmt.Sprintf("文档_%d.pdf", i+1),
			DocumentID: int64(i + 1),
			ChunkIndex: i,
		})
	}

	return chunks
}

// generateMockAnswer 生成模拟答案
func generateMockAnswer(query string, chunks []DocumentChunk) string {
	var answer string

	if len(chunks) > 0 {
		answer = fmt.Sprintf("关于「%s」的回答：\n\n", query)
		answer += "根据知识库检索结果，我找到以下相关信息：\n\n"

		for i, chunk := range chunks {
			if i < 3 { // 最多引用3个片段
				answer += fmt.Sprintf("%d. %s (相似度: %.2f)\n", i+1, chunk.Content, chunk.Score)
			}
		}

		answer += "\n注意：这是模拟数据，实际项目中需要配置 RAG 服务以获取真实答案。"
	} else {
		answer = fmt.Sprintf("未找到与「%s」相关的内容。", query)
	}

	return answer
}

// ========================================
// 工具工厂
// ========================================

// RAGToolFactory RAG 工具工厂
type RAGToolFactory struct {
	service RAGQueryService
}

// NewRAGToolFactory 创建 RAG 工具工厂
func NewRAGToolFactory(service RAGQueryService) *RAGToolFactory {
	return &RAGToolFactory{
		service: service,
	}
}

// CreateTool 创建工具
func (f *RAGToolFactory) CreateTool() (tool.InvokableTool, error) {
	InitRAGQueryTool(f.service)
	return NewRAGQueryTool()
}

// ========================================
// JSON 转换辅助函数
// ========================================

// RAGQueryRequestFromJSON 从 JSON 解析请求
func RAGQueryRequestFromJSON(jsonStr string) (*RAGQueryRequest, error) {
	var req RAGQueryRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}
	return &req, nil
}

// RAGQueryResultToJSON 将结果转为 JSON
func RAGQueryResultToJSON(result *RAGQueryResult) (string, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(data), nil
}
