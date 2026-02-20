package middleware

import (
	"github.com/gin-gonic/gin"
)

// DefaultMiddleware 默认中间件链
func DefaultMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		CORS(),
		RequestID(),
		Recovery(),
		Logger(),
		// TracingMiddleware(), // 可选：如果需要追踪功能
	}
}

// SetupMiddleware 设置中间件到路由
func SetupMiddleware(router *gin.Engine) {
	middlewares := DefaultMiddleware()
	router.Use(middlewares...)
}

// DefaultMiddlewareWithAuth 默认中间件链 + 认证
func DefaultMiddlewareWithAuth(authFunc gin.HandlerFunc) []gin.HandlerFunc {
	return append(
		[]gin.HandlerFunc{
			CORS(),
			RequestID(),
			Recovery(),
			Logger(),
		},
		authFunc,
		ContextToRequest(), // 在认证后，将 Gin 上下文传递到 request.Context
	)
}
