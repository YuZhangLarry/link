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
	dbConfig := config.LoadDatabaseConfig()
	jwtConfig := config.LoadJWTConfig()
	chatConfig := config.LoadChatConfig()

	// 初始化数据库
	if err := container.InitDatabase(dbConfig); err != nil {
		log.Fatalf("❌ 数据库初始化失败: %v", err)
	}
	defer container.CloseDatabase()

	// 获取数据库连接
	db := container.GetDB()

	// 初始化Repository
	userRepo := repository.NewUserRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	sessionRepo := repository.NewSessionRepository(db)

	// 初始化Service
	userService := service.NewUserService(userRepo, refreshTokenRepo, jwtConfig)
	chatService := service.NewChatService(chatConfig)
	messageService := service.NewMessageService(messageRepo)
	sessionService := service.NewSessionService(sessionRepo)

	// 初始化Handler
	authHandler := handler.NewAuthHandler(userService)
	chatHandler := handler.NewChatHandler(chatService, sessionService, messageService)
	messageHandler := handler.NewMessageHandler(messageService)
	sessionHandler := handler.NewSessionHandler(sessionService)

	// 初始化Middleware
	authMiddleware := middleware.NewAuthMiddleware(userService)

	// 创建Gin路由
	r := gin.Default()

	// 应用 CORS 中间件
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "服务器运行正常",
		})
	})

	// API路由组
	api := r.Group("/api/v1")
	{
		// 认证路由（无需Token）
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/logout", authMiddleware.AuthRequired(), authHandler.Logout)
		}

		// 用户路由（需要Token）
		user := api.Group("/user")
		user.Use(authMiddleware.AuthRequired())
		{
			user.GET("/profile", authHandler.GetProfile)
		}

		// 聊天路由（无需认证）
		chat := api.Group("/chat")
		{
			chat.POST("", chatHandler.Chat)             // 非流式聊天
			chat.POST("/stream", chatHandler.ChatStream) // 流式聊天
		}

		// 聊天路由（需要认证，可选）
		chatAuth := api.Group("/chat/auth")
		// TODO: 临时禁用认证，后续需要时再开启
		// chatAuth.Use(authMiddleware.AuthRequired())
		{
			chatAuth.POST("", chatHandler.ChatWithAuth)
			chatAuth.POST("/stream", chatHandler.ChatStreamWithAuth)
		}

		// 消息路由
		messages := api.Group("/messages")
		{
			messages.POST("", messageHandler.CreateMessage)           // 创建消息
			messages.GET("", messageHandler.ListMessages)             // 查询消息列表
			messages.GET("/:id", messageHandler.GetMessageByID)       // 获取消息详情
			messages.PUT("/:id", messageHandler.UpdateMessage)        // 更新消息
			messages.DELETE("/:id", messageHandler.DeleteMessage)     // 删除消息
		}

		// 会话路由
		sessions := api.Group("/sessions")
		{
			sessions.POST("", sessionHandler.CreateSession)                // 创建会话
			sessions.GET("", sessionHandler.ListSessions)                  // 查询会话列表
			sessions.GET("/:id", sessionHandler.GetSessionByID)            // 获取会话详情
			sessions.GET("/:id/detail", sessionHandler.GetSessionDetail)    // 获取会话完整详情
			sessions.PUT("/:id", sessionHandler.UpdateSession)             // 更新会话
			sessions.DELETE("/:id", sessionHandler.DeleteSession)          // 删除会话
			sessions.POST("/:id/archive", sessionHandler.ArchiveSession)    // 归档会话
			sessions.POST("/:id/activate", sessionHandler.ActivateSession)  // 激活会话
		}
	}

	// 启动服务器
	addr := fmt.Sprintf(":%s", dbConfig.Port)
	log.Printf("🚀 服务器启动在 http://localhost%s\n", addr)
	log.Printf("📚 API文档: http://localhost%s/api/v1\n", addr)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("❌ 服务器启动失败: %v", err)
	}
}
