package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"link/internal/types/interfaces"
)

// ========================================
// 权限检查中间件
// ========================================

// RequirePermission 权限检查中间件工厂函数
// resourceType: 资源类型 (kb/session/document/user/role/tenant)
// action: 操作类型 (create/read/update/delete/assign)
func RequirePermission(permissionService interfaces.PermissionService, resourceType, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取租户ID和用户ID
		tenantID := GetTenantID(c)
		userID, ok := GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized: user not authenticated",
			})
			c.Abort()
			return
		}

		// 检查用户是否有指定权限
		hasPermission, err := permissionService.CheckPermission(c.Request.Context(), tenantID, userID, resourceType, action)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check permission",
			})
			c.Abort()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":             "Permission denied",
				"resource_type":     resourceType,
				"action":            action,
				"required_permission": resourceType + ":" + action,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole 要求用户具有指定角色
func RequireRole(roleCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		userRole := GetUserRole(c)
		if userRole == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Permission denied: no role assigned",
			})
			c.Abort()
			return
		}

		// 检查角色是否匹配
		if userRole != roleCode {
			c.JSON(http.StatusForbidden, gin.H{
				"error":            "Permission denied: insufficient role",
				"required_role":    roleCode,
				"current_role":     userRole,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOwnerOrAdmin 要求用户是所有者或管理员
func RequireOwnerOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if userRole == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Permission denied: no role assigned",
			})
			c.Abort()
			return
		}

		// owner 或 admin 角色可以访问
		if userRole != "owner" && userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Permission denied: owner or admin role required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTenantOwner 要求用户是租户所有者
func RequireTenantOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := GetUserRole(c)
		if userRole != "owner" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Permission denied: tenant owner role required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
