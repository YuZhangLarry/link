package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"link/internal/application/service"
)

// 无需认证的API列表
var noAuthAPI = map[string][]string{
	"/health":               {"GET"},
	"/api/v1/auth/register": {"POST"},
	"/api/v1/auth/login":    {"POST"},
	"/api/v1/auth/refresh":  {"POST"},
}

// 检查请求是否在无需认证的API列表中
func isNoAuthAPI(path string, method string) bool {
	for api, methods := range noAuthAPI {
		// 如果以*结尾，按照前缀匹配，否则按照全路径匹配
		if strings.HasSuffix(api, "*") {
			if strings.HasPrefix(path, strings.TrimSuffix(api, "*")) && slices.Contains(methods, method) {
				return true
			}
		} else if path == api && slices.Contains(methods, method) {
			return true
		}
	}
	return false
}

// Auth 认证中间件
func Auth(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 忽略 OPTIONS 请求
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// 检查请求是否在无需认证的API列表中
		if isNoAuthAPI(c.Request.URL.Path, c.Request.Method) {
			c.Next()
			return
		}

		// JWT Token 认证
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: missing or invalid authorization header",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := userService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: invalid token",
			})
			c.Abort()
			return
		}

		// 检查 Token 类型
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: invalid token type",
			})
			c.Abort()
			return
		}

		// 存储用户信息到上下文（包含租户ID）
		SetUserContext(c, &UserContext{
			ID:       claims.UserID,
			Username: claims.Username,
			Email:    claims.Email,
			Status:   1, // 默认正常状态
			Role:     "",
		})

		// 如果Token包含租户ID，也设置到上下文
		if claims.TenantID > 0 {
			c.Set(TenantIDKey, claims.TenantID)
		}

		c.Next()
	}
}

// ========================================
// 辅助函数
// ========================================

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (int64, bool) {
	if uid, exists := c.Get(UserIDKey); exists {
		if uidInt, ok := uid.(int64); ok {
			return uidInt, true
		}
	}
	return 0, false
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	if username, exists := c.Get(UsernameKey); exists {
		if usernameStr, ok := username.(string); ok {
			return usernameStr, true
		}
	}
	return "", false
}
