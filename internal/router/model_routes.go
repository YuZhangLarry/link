package router

import (
	"github.com/gin-gonic/gin"
)

// SetupModelRoutes 设置模型路由
func SetupModelRoutes(
	api *gin.RouterGroup,
	authMiddleware gin.HandlerFunc,
	listHandler gin.HandlerFunc,
	detailHandler gin.HandlerFunc,
) {
	// 模型管理路由
	m := api.Group("/models")
	m.Use(authMiddleware) // JWT 已包含 tenant_id
	{
		// 获取模型列表（支持按类型筛选）
		m.GET("", listHandler)

		// 获取单个模型
		m.GET("/:id", detailHandler)
	}
}
