package middleware

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ========================================
// 上下文键定义
// ========================================

type contextKey string

const (
	// 请求上下文键
	TenantIDKey   contextKey = "tenant_id"   // 租户ID
	UserIDKey     contextKey = "user_id"     // 用户ID
	UserRoleKey   contextKey = "user_role"   // 用户角色
	UsernameKey   contextKey = "username"   // 用户名
	RequestIDKey  contextKey = "request_id"   // 请求ID
	StartTimeKey  contextKey = "start_time"   // 请求开始时间

	// 租户上下文键
	TenantInfoKey contextKey = "tenant_info" // 租户信息
	UserInfoKey   contextKey = "user_info"    // 用户信息
)

// ========================================
// 上下文数据结构
// ========================================

// TenantContext 租户上下文
type TenantContext struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Business    string `json:"business"`
	Status      string `json:"status"`
	StorageQuota int64  `json:"storage_quota"`
	StorageUsed  int64  `json:"storage_used"`
	Settings     string `json:"settings"` // JSON
}

// UserContext 用户上下文
type UserContext struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   int8   `json:"status"`
	Role     string `json:"role"` // 在当前租户中的角色
}

// RequestContext 请求上下文
type RequestContext struct {
	TraceID    string // 请求追踪ID
	StartTime int64  // 请求开始时间
	Path       string // 请求路径
	Method     string // 请求方法
	IP         string // 客户端IP
	UserAgent   string // 用户代理
}

// ========================================
// 上下文管理器
// ========================================

var contextPool = sync.Pool{
	New: func() interface{} {
		return &RequestContext{}
	},
}

// GetRequestContext 获取请求上下文
func GetRequestContext(c *gin.Context) *RequestContext {
	if startTimeValue, exists := c.Get(string(StartTimeKey)); exists {
		if ctx, ok := startTimeValue.(*time.Time); ok {
			rc := contextPool.Get().(*RequestContext)
			rc.StartTime = ctx.Unix()
			rc.Path = c.Request.URL.Path
			rc.Method = c.Request.Method
			rc.IP = c.ClientIP()
			rc.UserAgent = c.Request.UserAgent()
			return rc
		}
	}
	return nil
}

// SetRequestContext 设置请求上下文
func SetRequestContext(c *gin.Context) *RequestContext {
	rc := GetRequestContext(c)
	if rc != nil {
		c.Set(string(StartTimeKey), rc.StartTime)
	}
	return rc
}

// ========================================
// 上下文获取辅助函数
// ========================================

// GetTenantID 获取租户ID
func GetTenantID(c *gin.Context) int64 {
	if tid, exists := c.Get(TenantIDKey); exists {
		if tidInt, ok := tid.(int64); ok {
			return tidInt
		}
	}
	return 0
}

// GetUserRole 获取用户角色
func GetUserRole(c *gin.Context) string {
	if role, exists := c.Get(UserRoleKey); exists {
		if roleStr, ok := role.(string); ok {
			return roleStr
		}
	}
	return ""
}

// GetTenantContext 获取租户上下文
func GetTenantContext(c *gin.Context) *TenantContext {
	if tc, exists := c.Get(TenantInfoKey); exists {
		if tenantCtx, ok := tc.(*TenantContext); ok {
			return tenantCtx
		}
	}
	return nil
}

// GetUserContext 获取用户上下文
func GetUserContext(c *gin.Context) *UserContext {
	if uc, exists := c.Get(UserInfoKey); exists {
		if userCtx, ok := uc.(*UserContext); ok {
			return userCtx
		}
	}
	return nil
}

// ========================================
// 上下文设置函数
// ========================================

// SetTenantContext 设置租户上下文
func SetTenantContext(c *gin.Context, tenant *TenantContext) {
	c.Set(TenantInfoKey, tenant)
	c.Set(TenantIDKey, tenant.ID)
}

// SetUserContext 设置用户上下文
func SetUserContext(c *gin.Context, user *UserContext) {
	c.Set(UserInfoKey, user)
	c.Set(UserIDKey, user.ID)
	c.Set(UsernameKey, user.Username)
	if user.Role != "" {
		c.Set(UserRoleKey, user.Role)
	}
}

// SetAuthContext 设置认证上下文
func SetAuthContext(c *gin.Context, tenantID, userID int64, username, role string) {
	c.Set(TenantIDKey, tenantID)
	c.Set(UserIDKey, userID)
	c.Set(UsernameKey, username)
	c.Set(UserRoleKey, role)
}

// SetRequestID 设置请求ID
func SetRequestID(c *gin.Context, requestID string) {
	c.Set(RequestIDKey, requestID)
}

// GetRequestID 获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if requestIDStr, ok := requestID.(string); ok {
			return requestIDStr
		}
	}
	return ""
}

// MustGetTenantID 必须获取租户ID，否则返回错误
func MustGetTenantID(c *gin.Context) (int64, error) {
	tid := GetTenantID(c)
	if tid == 0 {
		return 0, ErrTenantNotSet
	}
	return tid, nil
}

// MustGetUserID 必须获取用户ID，否则返回错误
func MustGetUserID(c *gin.Context) (int64, error) {
	// Use the GetUserID from auth.go which returns (int64, bool)
	if uid, exists := GetUserID(c); exists {
		return uid, nil
	}
	return 0, ErrUserNotSet
}

// ========================================
// 错误定义
// ========================================

var (
	ErrTenantNotSet = errors.New("tenant context not set")
	ErrUserNotSet   = errors.New("user context not set")
)

// TenantRequiredError 租户必填错误
type TenantRequiredError struct{}

func (e *TenantRequiredError) Error() string {
	return "tenant required"
}

// UserRequiredError 用户必填错误
type UserRequiredError struct{}

func (e *UserRequiredError) Error() string {
	return "user required"
}

// ========================================
// 上下文清理
// ========================================

func ReleaseRequestContext(c *gin.Context) {
	rc := GetRequestContext(c)
	if rc != nil {
		contextPool.Put(rc)
		c.Set(string(StartTimeKey), nil)
		c.Set(TenantInfoKey, nil)
		c.Set(UserInfoKey, nil)
	}
}

// ========================================
// Context 传递中间件
// ========================================

// ContextToRequest 将 Gin 上下文中的用户/租户信息传递到 request.Context
// 这样 service 层可以从 ctx.Value() 中获取这些信息
func ContextToRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Gin 上下文中的值，并传递到 request.Context
		// 使用字符串键，与 service 层保持一致
		if userID, exists := c.Get(UserIDKey); exists {
			if uid, ok := userID.(int64); ok {
				ctx := context.WithValue(c.Request.Context(), "user_id", uid)
				c.Request = c.Request.WithContext(ctx)
			}
		}
		if tenantID, exists := c.Get(TenantIDKey); exists {
			if tid, ok := tenantID.(int64); ok {
				ctx := context.WithValue(c.Request.Context(), "tenant_id", tid)
				c.Request = c.Request.WithContext(ctx)
			}
		}
		if username, exists := c.Get(UsernameKey); exists {
			if uname, ok := username.(string); ok {
				ctx := context.WithValue(c.Request.Context(), "username", uname)
				c.Request = c.Request.WithContext(ctx)
			}
		}
		c.Next()
	}
}
