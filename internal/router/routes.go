package router

import (
	"link/internal/application/service"
	"link/internal/container"
	"link/internal/handler"
	"link/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 设置完整路由
func SetupRoutes(
	r *gin.Engine,
	userService *service.UserService,
	chatHandler *handler.ChatHandler,
	sessionHandler *handler.SessionHandler,
	messageHandler *handler.MessageHandler,
	tenantService *service.TenantService,
) {
	// ========================================
	// 全局中间件
	// ========================================
	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())

	// 健康检查（不需要认证）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "link-api"})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		// ========================================
		// 公开接口（不需要认证）
		// ========================================
		auth := api.Group("/auth")
		{
			// 这些接口暂时直接实现，实际应该使用 handler
			auth.POST("/register", func(c *gin.Context) {
				// 注册时需要提供 tenant_id
				c.JSON(200, gin.H{"message": "注册接口 - 请提供 tenant_id"})
			})
			auth.POST("/login", func(c *gin.Context) {
				// 登录时需要提供 tenant_id
				c.JSON(200, gin.H{"message": "登录接口 - 请提供 tenant_id"})
			})
			auth.POST("/refresh", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "刷新Token接口"})
			})
		}

		// ========================================
		// 需要认证的接口
		// ========================================

		// 创建认证中间件
		authMiddleware := middleware.Auth(userService)

		// ========================================
		// 租户管理路由（需要认证）
		// ========================================
		tenantHandler := handler.NewTenantHandler(tenantService)
		tenants := api.Group("/tenants")
		tenants.Use(authMiddleware)
		{
			tenants.POST("", tenantHandler.CreateTenant)
			tenants.GET("", tenantHandler.ListTenants)
			tenants.GET("/:id", tenantHandler.GetTenant)
			tenants.PUT("/:id", tenantHandler.UpdateTenant)
			tenants.DELETE("/:id", tenantHandler.DeleteTenant)
			tenants.POST("/:id/api-key", tenantHandler.RegenerateAPIKey)
			tenants.GET("/:id/storage", tenantHandler.GetStorageUsage)
		}

		// ========================================
		// 需要租户ID的接口
		// ========================================

		// 创建租户拦截器
		h := handler.NewTenantHandler(tenantService)
		tenantMiddleware := h.TenantRequired()

		// 聊天相关接口（需要认证 + 租户ID）
		chat := api.Group("/chat")
		chat.Use(authMiddleware, tenantMiddleware)
		{
			chat.POST("", chatHandler.Chat)
			chat.POST("/stream", chatHandler.ChatStream)
		}

		// 会话相关接口（需要认证 + 租户ID）
		sessions := api.Group("/sessions")
		sessions.Use(authMiddleware, tenantMiddleware)
		{
			sessions.POST("", sessionHandler.CreateSession)
			sessions.GET("", sessionHandler.ListSessions)
			sessions.GET("/:id", sessionHandler.GetSessionByID)
			sessions.PUT("/:id", sessionHandler.UpdateSession)
			sessions.DELETE("/:id", sessionHandler.DeleteSession)
			sessions.POST("/:id/archive", sessionHandler.ArchiveSession)
			sessions.POST("/:id/activate", sessionHandler.ActivateSession)
			sessions.GET("/:id/detail", sessionHandler.GetSessionDetail)
		}

		// 消息相关接口（需要认证 + 租户ID）
		messages := api.Group("/messages")
		messages.Use(authMiddleware, tenantMiddleware)
		{
			messages.GET("", messageHandler.ListMessages)
			messages.GET("/:id", messageHandler.GetMessageByID)
			messages.PUT("/:id", messageHandler.UpdateMessage)
			messages.DELETE("/:id", messageHandler.DeleteMessage)
		}

		// ========================================
		// 权限管理路由（需要认证 + 租户ID）
		// ========================================

		// 初始化权限系统
		if err := container.InitPermissionSystem(); err == nil {
			permissionService := container.GetPermissionService()
			if permissionService != nil {
				permGroup := api.Group("")
				permGroup.Use(authMiddleware, tenantMiddleware)
				{
					handler.RegisterPermissionRoutes(permGroup, permissionService)
				}
			}
		}

		// ========================================
		// 用户信息路由（需要认证）
		// ========================================
		users := api.Group("/users")
		users.Use(authMiddleware)
		{
			users.GET("/profile", func(c *gin.Context) {
				// 从上下文获取用户信息
				userID, ok := middleware.GetUserID(c)
				if !ok {
					c.JSON(401, gin.H{"error": "Unauthorized"})
					return
				}

				userInfo, err := userService.GetUserByID(c.Request.Context(), userID)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}

				c.JSON(200, gin.H{
					"code": 0,
					"message": "成功",
					"data": userInfo,
				})
			})
		}
	}
}

// ========================================
// 路由使用说明
// ========================================

/*
1. 公开接口（不需要认证）：
   POST /api/v1/auth/register  - 用户注册（需要 tenant_id）
   POST /api/v1/auth/login     - 用户登录（需要 tenant_id）
   POST /api/v1/auth/refresh   - 刷新Token
   GET  /health                - 健康检查

2. 认证接口（需要JWT Token，不需要租户ID）：
   GET  /api/v1/tenants        - 获取租户列表
   POST /api/v1/tenants        - 创建租户
   GET  /api/v1/users/profile  - 获取用户信息

3. 租户接口（需要JWT Token + 租户ID）：
   所有在请求头中需要提供：
   - Authorization: Bearer <token>
   - X-Tenant-ID: <tenant_id>

   例如：
   POST /api/v1/sessions
   GET  /api/v1/sessions
   POST /api/v1/chat
   等...

请求示例：
```bash
# 1. 登录（注册时类似）
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": 1,
    "email": "admin@link.com",
    "password": "admin123"
  }'

# 2. 使用返回的 access_token 访问需要认证的接口
curl -X GET http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Tenant-ID: 1"

# 3. 创建会话
curl -X POST http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Tenant-ID: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "新会话",
    "kb_id": "kb-uuid-123",
    "max_rounds": 5
  }'

# 4. 查询用户权限
curl -X GET http://localhost:8080/api/v1/permissions \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Tenant-ID: 1"

# 5. 检查权限
curl -X POST http://localhost:8080/api/v1/permissions/check \
  -H "Authorization: Bearer <access_token>" \
  -H "X-Tenant-ID: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "resource_type": "kb",
    "action": "create"
  }'
```

响应格式：
成功：
{
    "code": 0,
    "message": "成功",
    "data": { ... }
}

错误：
{
    "error": "错误信息"
}

或者：
{
    "code": 40000,
    "message": "错误描述"
}
*/
