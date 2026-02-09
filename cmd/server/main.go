package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"link/internal/application/repository"
	"link/internal/application/service"
	"link/internal/config"
	"link/internal/container"
	"link/internal/handler"
	"link/internal/middleware"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化 GORM 数据库
	gormDB, err := container.InitGORMDatabase(cfg.Database, "info")
	if err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	defer container.CloseDatabase()

	// 获取底层的 sql.DB
	sqlDB, err := gormDB.DB.DB()
	if err != nil {
		log.Fatalf("❌ 获取数据库连接失败: %v", err)
	}

	log.Println("✅ 数据库初始化成功")
	log.Println("🔧 正在初始化 Repository...")

	// 初始化权限系统（使用 database/sql）
	if err := container.InitPermissionSystem(); err != nil {
		log.Printf("⚠️  权限系统初始化失败: %v", err)
	} else {
		log.Println("✅ 权限系统初始化成功")
	}

	// 初始化 Repository (使用 GORM，启用多租户)
	tenantRepo := repository.NewTenantRepository(gormDB.DB, true)
	// User 和 RefreshToken 仓库仍然使用 database/sql
	userRepo := repository.NewUserRepository(sqlDB)
	refreshTokenRepo := repository.NewRefreshTokenRepository(sqlDB)
	sessionRepo := repository.NewSessionRepository(gormDB.DB)
	messageRepo := repository.NewMessageRepository(gormDB.DB)

	// 初始化 Service
	tenantService := service.NewTenantService(tenantRepo)
	userService := service.NewUserService(userRepo, refreshTokenRepo, tenantRepo, cfg.JWT)
	chatService := service.NewChatService(cfg.Chat)
	messageService := service.NewMessageService(messageRepo)
	sessionService := service.NewSessionService(sessionRepo)

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(userService)
	tenantHandler := handler.NewTenantHandler(tenantService)
	chatHandler := handler.NewChatHandler(chatService, sessionService, messageService)
	messageHandler := handler.NewMessageHandler(messageService)
	sessionHandler := handler.NewSessionHandler(sessionService)

	// 设置 Gin 运行模式
	gin.SetMode(cfg.Server.Mode)

	// 创建 Gin 路由
	r := gin.Default()

	// 应用全局中间件
	middleware.SetupMiddleware(r)

	// 应用认证中间件（在路由中使用）
	authMiddleware := middleware.Auth(userService)
	// ContextToRequest 中间件：将 Gin 上下文传递到 request.Context
	contextToRequest := middleware.ContextToRequest()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"message": "服务器运行正常",
			"version": "2.0.0-multi-tenant",
		})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 认证路由（无需 Token）
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authMiddleware, authHandler.Logout)
		}

		// 租户管理路由（需要认证）
		tenants := api.Group("/tenants")
		tenants.Use(authMiddleware, contextToRequest)
		{
			tenants.POST("", tenantHandler.CreateTenant)
			tenants.GET("", tenantHandler.ListTenants)
			tenants.GET("/:id", tenantHandler.GetTenant)
			tenants.PUT("/:id", tenantHandler.UpdateTenant)
			tenants.DELETE("/:id", tenantHandler.DeleteTenant)
			tenants.POST("/:id/api-key", tenantHandler.RegenerateAPIKey)
			tenants.GET("/:id/storage", tenantHandler.GetStorageUsage)
		}

		// 用户路由（需要 Token）
		user := api.Group("/user")
		user.Use(authMiddleware, contextToRequest)
		{
			user.GET("/profile", authHandler.GetProfile)
		}

		// 聊天路由
		chat := api.Group("/chat")
		{
			chat.POST("", chatHandler.Chat)             // 非流式聊天
			chat.POST("/stream", chatHandler.ChatStream) // 流式聊天
		}

		// 聊天路由（需要认证）
		chatAuth := api.Group("/chat/auth")
		chatAuth.Use(authMiddleware, contextToRequest)
		{
			chatAuth.POST("", chatHandler.ChatWithAuth)
			chatAuth.POST("/stream", chatHandler.ChatStreamWithAuth)
		}

		// 消息路由（需要认证）
		messages := api.Group("/messages")
		messages.Use(authMiddleware, contextToRequest)
		{
			messages.POST("", messageHandler.CreateMessage)
			messages.GET("", messageHandler.ListMessages)
			messages.GET("/:id", messageHandler.GetMessageByID)
			messages.PUT("/:id", messageHandler.UpdateMessage)
			messages.DELETE("/:id", messageHandler.DeleteMessage)
		}

		// 会话路由（需要认证）
		sessions := api.Group("/sessions")
		sessions.Use(authMiddleware, contextToRequest)
		{
			sessions.POST("", sessionHandler.CreateSession)
			sessions.GET("", sessionHandler.ListSessions)
			sessions.GET("/:id", sessionHandler.GetSessionByID)
			sessions.GET("/:id/detail", sessionHandler.GetSessionDetail)
			sessions.PUT("/:id", sessionHandler.UpdateSession)
			sessions.DELETE("/:id", sessionHandler.DeleteSession)
			sessions.POST("/:id/archive", sessionHandler.ArchiveSession)
			sessions.POST("/:id/activate", sessionHandler.ActivateSession)
		}

		// ========================================
		// 权限管理路由（需要认证 + 租户ID）
		// ========================================
		permissionService := container.GetPermissionService()
		if permissionService != nil {
			// 创建租户拦截器
			tenantMiddleware := tenantHandler.TenantRequired()

			// 权限路由组
			permGroup := api.Group("")
			permGroup.Use(authMiddleware, tenantMiddleware, contextToRequest)
			{
				// 角色管理
				roles := permGroup.Group("/roles")
				{
					roles.GET("", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "获取角色列表 - 需要实现"})
					})
					roles.POST("", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "创建角色 - 需要实现"})
					})
					roles.PUT("/:id", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "更新角色 - 需要实现"})
					})
					roles.DELETE("/:id", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "删除角色 - 需要实现"})
					})
					roles.POST("/assign", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "分配角色 - 需要实现"})
					})
				}

				// 用户角色
				permGroup.DELETE("/users/:user_id/role", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "撤销角色 - 需要实现"})
				})
				permGroup.GET("/role", func(c *gin.Context) {
					c.JSON(200, gin.H{"message": "获取用户角色 - 需要实现"})
				})

				// 权限查询
				permissions := permGroup.Group("/permissions")
				{
					permissions.GET("", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "获取用户权限 - 需要实现"})
					})
					permissions.POST("/check", func(c *gin.Context) {
						c.JSON(200, gin.H{"message": "检查权限 - 需要实现"})
					})
				}
			}

			log.Println("✅ 权限路由已注册")
		}
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("🚀 服务器启动在 http://localhost:%s\n", cfg.Server.Port)
	log.Printf("📚 API文档: http://localhost:%s/api/v1\n", cfg.Server.Port)
	log.Printf("🔐 认证方式: JWT Bearer Token / X-API-Key\n")

	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ 服务器启动失败: %v", err)
	}
}
