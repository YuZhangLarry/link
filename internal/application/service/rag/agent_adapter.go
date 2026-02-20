// Package rag 提供 RAG 服务适配器，用于 Agent 工具
package rag

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/cloudwego/eino/components/embedding"

	agentTool "link/internal/agent/tool"
	"link/internal/types/interfaces"
)

// ========================================
// Agent 工具适配器 - 直接使用 Retriever
// ========================================

// AgentRAGAdapter 直接使用 Retriever 进行检索
type AgentRAGAdapter struct {
	retriever *Retriever
	tenantID  int64
}

// NewAgentRAGAdapter 创建 Agent RAG 适配器
func NewAgentRAGAdapter(
	kbSettingRepo interfaces.KBSettingRepository,
	chunkRepo interfaces.ChunkRepository,
	embedder embedding.Embedder,
	neo4jRepo interfaces.Neo4jGraphRepository,
	graphQueryRepo interfaces.GraphQueryRepository,
	tenantID int64,
) *AgentRAGAdapter {
	retriever := NewRetriever(
		kbSettingRepo,
		chunkRepo,
		embedder,
		nil, // milvusRetriever
		neo4jRepo,
		graphQueryRepo,
	)
	return &AgentRAGAdapter{
		retriever: retriever,
		tenantID:  tenantID,
	}
}

// Query 实现 RAGQueryService 接口（只返回检索结果）
func (a *AgentRAGAdapter) Query(ctx context.Context, req *agentTool.RAGQueryRequest) (*agentTool.RAGQueryResult, error) {
	startTime := time.Now()

	// 参数验证
	if req.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// 设置默认值
	topK := req.TopK
	if topK <= 0 {
		topK = 5
	}
	if topK > 20 {
		topK = 20
	}

	// 转换 KBID（int64 -> string）
	kbID := strconv.FormatInt(req.KBID, 10)
	if req.KBID == 0 {
		// 使用默认知识库 ID
		kbID = "4b856e03-953a-4221-8d7e-b2ee7b0b30b3"
	}

	// 确定检索模式
	retrievalMode := req.RetrievalMode
	if retrievalMode == "" {
		retrievalMode = "hybrid"
	}

	// 构建检索选项
	opts := &RetrieveOptions{
		TopK:                topK,
		SimilarityThreshold: req.MinScore,
		RerankEnabled:       req.EnableRerank,
		GraphEnabled:        retrievalMode == "graph",
		Alpha:               0.5,
	}

	// 执行检索
	var response *RetrieveResponse
	var err error

	switch retrievalMode {
	case "vector":
		response, err = a.retriever.vectorRetrieveWithEmbedding(ctx, a.tenantID, kbID, req.Query, opts)
	case "bm25", "keyword":
		response, err = a.retriever.bm25Retrieve(ctx, a.tenantID, kbID, req.Query, opts)
	case "graph":
		response, err = a.retriever.GraphRetrieve(ctx, a.tenantID, kbID, req.Query, opts)
	default: // hybrid
		response, err = a.retriever.hybridRetrieve(ctx, a.tenantID, kbID, req.Query, opts)
	}

	if err != nil {
		return &agentTool.RAGQueryResult{
			Answer:        fmt.Sprintf("检索失败: %v", err),
			Chunks:        []agentTool.DocumentChunk{},
			Count:         0,
			Query:         req.Query,
			KBID:          req.KBID,
			Latency:       time.Since(startTime).Milliseconds(),
			RetrievalMode: retrievalMode,
			HasAnswer:     false,
		}, nil
	}

	// 转换结果
	chunks := make([]agentTool.DocumentChunk, 0, len(response.Results))
	for _, doc := range response.Results {
		var docID int64
		if doc.KnowledgeID != "" {
			if id, err := strconv.ParseInt(doc.KnowledgeID, 10, 64); err == nil {
				docID = id
			}
		}

		chunks = append(chunks, agentTool.DocumentChunk{
			Content:    doc.Content,
			Score:      float64(doc.Score),
			Source:     doc.MatchType,
			DocumentID: docID,
			ChunkIndex: doc.ChunkIndex,
			Metadata: map[string]interface{}{
				"chunk_id":     doc.ChunkID,
				"knowledge_id": doc.KnowledgeID,
				"kb_id":        doc.KBID,
			},
		})
	}

	// 构建简要状态信息
	var statusMsg string
	if len(chunks) == 0 {
		statusMsg = "未找到相关内容"
	} else {
		// 返回检索结果的摘要，让 Agent 看到具体内容
		statusMsg = fmt.Sprintf("检索到 %d 个相关片段：\n", len(chunks))
		for i, c := range chunks {
			if i >= 3 { // 最多显示前3个
				statusMsg += fmt.Sprintf("... 还有 %d 个片段\n", len(chunks)-3)
				break
			}
			// 截取内容前100字符
			content := c.Content
			if len(content) > 100 {
				content = content[:100] + "..."
			}
			statusMsg += fmt.Sprintf("%d. [相似度: %.2f] %s\n", i+1, c.Score, content)
		}
	}

	return &agentTool.RAGQueryResult{
		Answer:        statusMsg,
		Chunks:        chunks,
		Count:         len(chunks),
		Query:         req.Query,
		KBID:          req.KBID,
		Latency:       time.Since(startTime).Milliseconds(),
		RetrievalMode: retrievalMode,
		HasAnswer:     len(chunks) > 0,
	}, nil
}

// ========================================
// 初始化函数
// ========================================

// InitAgentRAGTool 初始化 Agent 的 RAG 工具（直接使用 Retriever）
func InitAgentRAGTool(
	kbSettingRepo interfaces.KBSettingRepository,
	chunkRepo interfaces.ChunkRepository,
	embedder embedding.Embedder,
	neo4jRepo interfaces.Neo4jGraphRepository,
	graphQueryRepo interfaces.GraphQueryRepository,
	tenantID int64,
) {
	adapter := NewAgentRAGAdapter(kbSettingRepo, chunkRepo, embedder, neo4jRepo, graphQueryRepo, tenantID)
	agentTool.InitRAGQueryTool(adapter)
}
