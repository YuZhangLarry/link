package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/gin-gonic/gin"

	"link/internal/application/chunker"
	"link/internal/application/repository"
	"link/internal/application/repository/retriever/neo4j"
	repoService "link/internal/application/service"
	"link/internal/config"
	"link/internal/container"
	"link/internal/handler"
	"link/internal/middleware"
	embeddingModel "link/internal/models/embedding"
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

	// 初始化 Milvus
	if err := container.InitMilvus(cfg.Milvus); err != nil {
		log.Printf("⚠️  Milvus 初始化失败: %v", err)
		log.Println("继续运行（向量检索功能将不可用）...")
	} else {
		log.Println("✅ Milvus 初始化成功")
		defer container.CloseMilvus()
	}

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

	// 初始化知识库Repository
	kbBaseRepo := repository.NewKnowledgeBaseRepository(gormDB.DB, true)
	knowledgeRepo := repository.NewKnowledgeRepository(gormDB.DB, true)
	chunkRepo := repository.NewChunkRepository(gormDB.DB, true)
	kbSettingRepo := repository.NewKBSettingRepository(gormDB.DB, true)

	// 初始化 Service
	tenantService := repoService.NewTenantService(tenantRepo)
	userService := repoService.NewUserService(userRepo, refreshTokenRepo, tenantRepo, cfg.JWT)
	chatService := repoService.NewChatService(cfg.Chat)
	messageService := repoService.NewMessageService(messageRepo)
	sessionService := repoService.NewSessionService(sessionRepo)

	// 初始化知识库Service
	kbBaseService := repoService.NewKnowledgeBaseService(kbBaseRepo, knowledgeRepo, chunkRepo, gormDB.DB)
	knowledgeService := repoService.NewKnowledgeService(knowledgeRepo, chunkRepo, kbSettingRepo, gormDB.DB)

	// 初始化 Graph Service（使用 neo4j retriever 的仓储，确保写入和查询一致）
	var graphService *repoService.GraphService
	if cfg.Neo4j != nil && cfg.Neo4j.URI != "" {
		ctx := context.Background()
		neo4jCfg := container.Config{
			URI:      cfg.Neo4j.URI,
			Username: cfg.Neo4j.Username,
			Password: cfg.Neo4j.Password,
		}
		driver, err := container.CreateDriver(ctx, neo4jCfg)
		if err != nil {
			log.Printf("⚠️  Neo4j driver 创建失败: %v", err)
		} else {
			defer func() {
				log.Println("🔌 关闭 Neo4j driver...")
				driver.Close(ctx)
			}()
			// 使用 neo4j retriever 的仓储实现（Neo4j 操作）
			graphRepo := neo4j.NewNeo4jRepository(driver)
			// 创建图谱查询仓储（与知识库的关联查询）
			graphQueryRepo := repository.NewGraphQueryRepository(gormDB.DB, true)
			// 使用包含查询仓储的构造函数
			graphService = repoService.NewGraphServiceWithQuery(graphRepo, graphQueryRepo)
			log.Println("✅ Graph Service 初始化成功 (使用 neo4j retriever 仓储 + 图谱查询仓储)")
		}
	}

	// 初始化 Embedder
	var embedder embedding.Embedder
	if cfg.Embedding != nil && cfg.Embedding.APIKey != "" {
		embedder, err = embeddingModel.NewEmbedder(cfg.Embedding)
		if err != nil {
			log.Printf("⚠️  Embedder 初始化失败: %v", err)
		} else {
			log.Println("✅ Embedder 初始化成功")
		}
	}

	// 初始化 Milvus Schema（创建collection和索引）
	if embedder != nil && container.MilvusClient != nil {
		if err := container.InitMilvusSchema(embedder); err != nil {
			log.Printf("⚠️  Milvus Schema 初始化失败: %v", err)
			log.Println("文件上传功能可能无法正常工作")
		} else {
			log.Println("✅ Milvus Schema 初始化成功")
		}
	} else if container.MilvusClient != nil {
		log.Println("⚠️  Embedder 未初始化，跳过 Milvus Schema 初始化")
	}

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(userService)
	tenantHandler := handler.NewTenantHandler(tenantService)
	chatHandler := handler.NewChatHandler(chatService, sessionService, messageService)
	messageHandler := handler.NewMessageHandler(messageService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	kbBaseHandler := handler.NewKnowledgeBaseHandler(kbBaseService)

	// 初始化图谱Handler
	var graphHandler *handler.GraphHandler
	if graphService != nil {
		graphHandler = handler.NewGraphHandler(graphService)
	}

	// 初始化完整知识库处理器
	var knowledgeHandler *handler.KnowledgeHandlerFull
	if graphService != nil && embedder != nil && container.MilvusClient != nil {
		chunkConfig := &chunker.SimpleConfig{
			ChunkSize:     512,
			Overlap:       100,
			Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!"},
			KeepSeparator: true,
		}
		knowledgeHandler = handler.NewKnowledgeHandlerFull(
			knowledgeService,
			graphService,
			embedder,
			container.MilvusClient,
			chunkConfig,
		)
		log.Println("✅ Knowledge Handler 初始化成功")
	} else {
		log.Println("⚠️  Knowledge Handler 未完全初始化（缺少 GraphService/Embedder/Milvus）")
	}

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
	// 租户拦截器：需要租户ID的接口
	tenantMiddleware := tenantHandler.TenantRequired()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
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
			chat.POST("", chatHandler.Chat)              // 非流式聊天
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
		// 分块管理路由（用于测试）- 必须在knowledge-bases之前定义
		// ========================================
		chunks := api.Group("/chunks")
		chunks.Use(authMiddleware, tenantMiddleware, contextToRequest)
		{
			chunks.POST("", kbBaseHandler.CreateChunk)
		}

		// ========================================
		// 知识库管理路由（需要认证 + 租户ID）
		// ========================================
		knowledgeBases := api.Group("/knowledge-bases")
		knowledgeBases.Use(authMiddleware, tenantMiddleware, contextToRequest)
		{
			knowledgeBases.POST("", kbBaseHandler.Create)
			knowledgeBases.GET("", kbBaseHandler.GetList)
			knowledgeBases.GET("/:id", kbBaseHandler.GetDetail)
			knowledgeBases.PUT("/:id", kbBaseHandler.Update)
			knowledgeBases.DELETE("/:id", kbBaseHandler.Delete)
			knowledgeBases.GET("/:id/stats", kbBaseHandler.GetStats)
			knowledgeBases.GET("/:id/knowledge", kbBaseHandler.GetKnowledgeList)
			knowledgeBases.DELETE("/:id/knowledge/:knowledge_id", kbBaseHandler.DeleteKnowledge)
			knowledgeBases.GET("/:id/chunks", kbBaseHandler.GetChunks)

			// 知识图谱路由（需要图谱服务）
			if graphHandler != nil {
				knowledgeBases.GET("/:id/graph", graphHandler.GetGraph)
				knowledgeBases.POST("/:id/graph/search", graphHandler.SearchNode)
				knowledgeBases.GET("/:id/graph/nodes/:nodeId", graphHandler.GetNodeDetail)
				knowledgeBases.POST("/:id/graph/nodes", graphHandler.AddNode)
				knowledgeBases.PUT("/:id/graph/nodes/:nodeId", graphHandler.UpdateNode)
				knowledgeBases.DELETE("/:id/graph/nodes/:nodeId", graphHandler.DeleteNode)
				knowledgeBases.POST("/:id/graph/relations", graphHandler.AddRelation)
				knowledgeBases.PUT("/:id/graph/relations/:relationId", graphHandler.UpdateRelation)
				knowledgeBases.DELETE("/:id/graph/relations/:relationId", graphHandler.DeleteRelation)
				knowledgeBases.GET("/:id/graph/relation-types", graphHandler.GetRelationTypes)
				knowledgeBases.DELETE("/:id/graph", graphHandler.DeleteGraph)
			} else {
				// 如果未初始化图谱处理器，返回提示信息
				knowledgeBases.GET("/:id/graph", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Graph handler not initialized"})
				})
				knowledgeBases.POST("/:id/graph/search", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Graph handler not initialized"})
				})
				knowledgeBases.GET("/:id/graph/nodes/:nodeId", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Graph handler not initialized"})
				})
				knowledgeBases.POST("/:id/graph/nodes", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Graph handler not initialized"})
				})
				knowledgeBases.POST("/:id/graph/relations", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Graph handler not initialized"})
				})
				knowledgeBases.DELETE("/:id/graph", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Graph handler not initialized"})
				})
			}

			// 知识库文件操作路由（需要完整知识库处理器）
			if knowledgeHandler != nil {
				// 上传文件
				knowledgeBases.POST("/:id/knowledge/file", knowledgeHandler.UploadKnowledgeFile)
				// 获取处理状态
				knowledgeBases.GET("/:id/knowledge/:knowledge_id/status", knowledgeHandler.GetKnowledgeStatus)
			} else {
				// 如果未初始化完整处理器，返回提示信息
				knowledgeBases.POST("/:id/knowledge/file", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Knowledge handler not fully initialized"})
				})
				knowledgeBases.GET("/:id/knowledge/:knowledge_id/status", func(c *gin.Context) {
					c.JSON(503, gin.H{"error": "Knowledge handler not fully initialized"})
				})
			}
		}

		// 知识搜索路由（需要认证）
		api.POST("/knowledge/search", authMiddleware, contextToRequest, handler.SearchKnowledge)

		// ========================================
		// 权限管理路由（需要认证 + 租户ID）
		// ========================================
		permissionService := container.GetPermissionService()
		if permissionService != nil {
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
