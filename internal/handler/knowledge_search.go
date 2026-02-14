package handler

import (
	"log"

	"github.com/gin-gonic/gin"

	"link/internal/middleware"
)

// SearchKnowledge 知识搜索
// POST /api/v1/knowledge/search
func SearchKnowledge(c *gin.Context) {
	var req struct {
		Query          string   `json:"query" binding:"required"`
		KBIDs          []string `json:"kb_ids"`
		TopK           int      `json:"top_k"`
		ScoreThreshold float64  `json:"score_threshold"`
		IncludeGraph   bool     `json:"include_graph"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 从上下文获取租户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	log.Printf("[KnowledgeSearch] tenant_id=%d, query=%q, kb_ids=%v, top_k=%d, threshold=%f",
		tenantID, req.Query, req.KBIDs, req.TopK, req.ScoreThreshold)

	// TODO: 实现知识搜索逻辑
	// 1. 从 Milvus 进行向量检索
	// 2. 如果 include_graph=true，从 Neo4j 获取图谱信息
	// 3. 合并结果返回

	// 临时返回空结果
	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"query":   req.Query,
			"results": []interface{}{},
			"total":   0,
		},
	})
}
