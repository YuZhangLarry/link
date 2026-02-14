package router

import (
	"github.com/gin-gonic/gin"
)

// SetupKnowledgeRoutes 设置知识库路由
func SetupKnowledgeRoutes(
	api *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
	tenantMiddleware gin.HandlerFunc,
	uploadHandler gin.HandlerFunc,
	statusHandler gin.HandlerFunc,
	getKnowledgeListHandler gin.HandlerFunc,
	deleteKnowledgeHandler gin.HandlerFunc,
	getChunksHandler gin.HandlerFunc,
) {
	// 知识库文件操作路由（JWT 已包含 tenant_id）
	knowledge := api.Group("/knowledge-bases/:id/knowledge")
	knowledge.Use(authMiddleware)
	{
		// 上传文件
		knowledge.POST("/file", uploadHandler)

		// 获取处理状态
		knowledge.GET("/:knowledge_id/status", statusHandler)
	}

	// 知识条目管理路由（JWT 已包含 tenant_id）
	knowledgeItems := api.Group("/knowledge-bases/:kb_id/knowledge")
	knowledgeItems.Use(authMiddleware)
	{
		// 列出知识条目
		knowledgeItems.GET("", getKnowledgeListHandler)

		// 获取单个知识条目（待实现）
		knowledgeItems.GET("/:knowledge_id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "get knowledge item - to be implemented"})
		})

		// 删除知识条目
		knowledgeItems.DELETE("/:knowledge_id", deleteKnowledgeHandler)

		// 更新知识条目（待实现）
		knowledgeItems.PUT("/:knowledge_id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "update knowledge item - to be implemented"})
		})
	}

	// 分块管理路由（JWT 已包含 tenant_id）
	chunks := api.Group("/knowledge-bases/:kb_id/chunks")
	chunks.Use(authMiddleware)
	{
		// 列出分块
		chunks.GET("", getChunksHandler)

		// 获取单个分块（待实现）
		chunks.GET("/:chunk_id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "get chunk - to be implemented"})
		})

		// 更新分块（待实现）
		chunks.PUT("/:chunk_id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "update chunk - to be implemented"})
		})

		// 删除分块（待实现）
		chunks.DELETE("/:chunk_id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "delete chunk - to be implemented"})
		})

		// 批量更新分块状态（待实现）
		chunks.POST("/batch/status", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "batch update chunk status - to be implemented"})
		})
	}
}

// SetupKnowledgeBaseRoutes 设置知识库管理路由
func SetupKnowledgeBaseRoutes(
	api *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
	tenantMiddleware gin.HandlerFunc,
	createHandler gin.HandlerFunc,
	listHandler gin.HandlerFunc,
	detailHandler gin.HandlerFunc,
	updateHandler gin.HandlerFunc,
	deleteHandler gin.HandlerFunc,
	statsHandler gin.HandlerFunc,
) {
	// 知识库管理路由
	kb := api.Group("/knowledge-bases")
	kb.Use(authMiddleware) // JWT 已包含 tenant_id，无需额外的 tenantMiddleware
	{
		// 创建知识库
		kb.POST("", createHandler)

		// 列出知识库
		kb.GET("", listHandler)

		// 获取单个知识库
		kb.GET("/:id", detailHandler)

		// 更新知识库
		kb.PUT("/:id", updateHandler)

		// 删除知识库
		kb.DELETE("/:id", deleteHandler)

		// 获取知识库统计
		kb.GET("/:id/stats", statsHandler)
	}

	// 知识搜索路由（JWT 已包含 tenant_id）
	api.POST("/knowledge/search", authMiddleware, func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "search knowledge - to be implemented"})
	})
}
