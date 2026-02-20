package types

// KnowledgeBaseStats 知识库统计信息
type KnowledgeBaseStats struct {
	KBID           string `json:"kb_id"`           // 知识库ID
	KnowledgeCount int64  `json:"knowledge_count"` // 文档数量
	ChunkCount     int64  `json:"chunk_count"`     // 分块数量
	TotalSize      int64  `json:"total_size"`      // 总存储大小（字节）
}
