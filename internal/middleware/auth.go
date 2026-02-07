package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"link/internal/application/service"
	"link/internal/types"
)

type AuthMiddleware struct {
	userService *service.UserService
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(userService *service.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		userService: userService,
	}
}

// AuthRequired JWT认证中间件
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Header获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "未提供认证Token",
			})
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "Token格式错误",
			})
			c.Abort()
			return
		}

		// 验证Token
		claims, err := m.userService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "Token无效或已过期",
			})
			c.Abort()
			return
		}

		// 检查Token类型
		if claims.TokenType != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    -1,
				"message": "Token类型错误",
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(int64), true
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	return username.(string), true
}

// GetUserClaims 从上下文获取完整的用户信息
func GetUserClaims(c *gin.Context) (*types.TokenClaims, bool) {
	userID, hasID := c.Get("user_id")
	username, hasUsername := c.Get("username")
	email, hasEmail := c.Get("email")

	if !hasID || !hasUsername || !hasEmail {
		return nil, false
	}

	return &types.TokenClaims{
		UserID:   userID.(int64),
		Username: username.(string),
		Email:    email.(string),
	}, true
}

// WhiteList 白名单中间件（跳过认证的路径）
func (m *AuthMiddleware) WhiteList(whiteListPaths []string) gin.HandlerFunc {
	// 创建路径映射
	whiteList := make(map[string]bool)
	for _, path := range whiteListPaths {
		whiteList[path] = true
	}

	return func(c *gin.Context) {
		// 检查当前路径是否在白名单中
		for path := range whiteList {
			if c.Request.URL.Path == path || strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// 不在白名单中，执行认证
		m.AuthRequired()(c)
	}
}
