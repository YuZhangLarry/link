package tool

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// NewKbQueryTool 创建知识库查询工具
func NewKbQueryTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"kb_query",
		`从知识库中检索相关信息。

支持多种检索模式：
- vector: 向量检索，基于语义相似度
- bm25: 全文检索，基于关键词匹配
- hybrid: 混合检索，结合向量和关键词
- graph: 图谱检索，基于知识关系

适用于：需要从已上传的文档中查找答案的场景`,
		queryKb,
	)
}

// queryKb 执行知识库查询（普通函数，非方法）
func queryKb(ctx context.Context, req *KbQueryRequest) (*KbQueryResult, error) {
	startTime := time.Now()

	// 1. 参数验证
	if req.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}
	if req.KBID <= 0 {
		return nil, fmt.Errorf("invalid kb_id")
	}

	// 设置默认值
	if req.TopK <= 0 {
		req.TopK = 5
	}
	if req.Similarity <= 0 {
		req.Similarity = 0.7
	}
	if req.RetrievalMode == "" {
		req.RetrievalMode = "hybrid"
	}

	// 2. 实际检索逻辑
	// TODO: 实现真实的检索逻辑
	// 这里需要调用检索服务，从 Milvus/Neo4j 等存储中查询

	// 模拟检索结果
	chunks := mockRetrieveKb(req)

	// 3. 构建返回结果
	result := &KbQueryResult{
		Results: chunks,
		Count:   len(chunks),
		Query:   req.Query,
		Latency: int(time.Since(startTime).Milliseconds()),
	}

	// 4. 记录搜索历史（可选）
	// k.saveSearchHistory(ctx, req, result)

	return result, nil
}

// mockRetrieveKb 模拟检索逻辑（实际项目中需要替换为真实实现）
func mockRetrieveKb(req *KbQueryRequest) []KbChunk {
	// 这里是模拟数据，实际项目中应该：
	// 1. 将 query 转为 embedding
	// 2. 从 Milvus 搜索相似向量
	// 3. 从数据库获取对应的 chunk 内容
	// 4. 如果启用图谱，从 Neo4j 查询关联信息

	return []KbChunk{
		{
			Content:    "这是从知识库中检索到的相关内容片段。实际项目中，这里会包含真实的文档内容。",
			Score:      0.92,
			Source:     "example.pdf",
			DocumentID: 1,
			ChunkIndex: 0,
		},
		{
			Content:    "这是第二个相关的内容片段，与用户查询问题相关。",
			Score:      0.87,
			Source:     "example.pdf",
			DocumentID: 1,
			ChunkIndex: 1,
		},
	}
}

// ========================================
// 知识库列表工具
// ========================================

// NewKbListTool 创建知识库列表工具
func NewKbListTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"kb_list",
		`获取用户的可用知识库列表。

返回知识库的基本信息，包括名称、描述、文档数量等。
适用于：用户需要查看有哪些知识库可用，或者需要选择特定知识库进行查询的场景`,
		listKb,
	)
}

// listKb 获取知识库列表（普通函数）
func listKb(ctx context.Context, req *KbListRequest) (*KbListResult, error) {
	// TODO: 实际项目中从数据库查询
	// SELECT kb.*, COUNT(d.id) as document_count
	// FROM knowledge_bases kb
	// LEFT JOIN documents d ON kb.id = d.kb_id
	// WHERE kb.user_id = ? OR kb.is_public = true
	// GROUP BY kb.id

	// 模拟数据
	kbs := []KbInfo{
		{
			ID:          1,
			Name:        "技术文档",
			Description: "项目相关的技术文档和API说明",
			DocumentCount: 15,
		},
		{
			ID:          2,
			Name:        "产品手册",
			Description: "产品使用手册和用户指南",
			DocumentCount: 8,
		},
	}

	return &KbListResult{
		KnowledgeBases: kbs,
		Count:          len(kbs),
	}, nil
}

// ========================================
// 文档列表工具
// ========================================

// NewDocumentListTool 创建文档列表工具
func NewDocumentListTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"document_list",
		`获取指定知识库中的文档列表。

返回文档的基本信息，包括文件名、类型、处理状态等。
适用于：用户需要查看某个知识库中有哪些文档的场景`,
		listDocuments,
	)
}

// listDocuments 获取文档列表（普通函数）
func listDocuments(ctx context.Context, req *DocumentListRequest) (*DocumentListResult, error) {
	// TODO: 实际项目中从数据库查询
	// SELECT * FROM documents WHERE kb_id = ? LIMIT ?

	// 模拟数据
	docs := []DocInfo{
		{
			ID:         1,
			FileName:   "API设计文档.pdf",
			FileType:   "pdf",
			Status:     "completed",
			ChunkCount: 45,
			CreatedAt:  "2024-01-15T10:30:00Z",
		},
		{
			ID:         2,
			FileName:   "系统架构.docx",
			FileType:   "docx",
			Status:     "completed",
			ChunkCount: 32,
			CreatedAt:  "2024-01-14T15:20:00Z",
		},
	}

	return &DocumentListResult{
		Documents: docs,
		Count:     len(docs),
	}, nil
}
