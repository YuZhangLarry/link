package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"log/slog"
)

// Recovery 中间件 - 恢复 panic
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get request ID
				requestID := GetRequestID(c)
				tenantID := GetTenantID(c)

				// Build stack trace
				stack := debug.Stack()

				// Log error
				slog.Error("Request panic",
					"request_id", requestID,
					"tenant_id", tenantID,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"ip", c.ClientIP(),
					"error", err,
					"stack", string(stack),
				)

				// 返回500错误
				c.JSON(500, gin.H{
					"error":   "Internal Server Error",
					"message": fmt.Sprintf("%v", err),
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}
