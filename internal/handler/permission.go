package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"link/internal/middleware"
	"link/internal/types"
	"link/internal/types/interfaces"
)

// PermissionHandler 权限处理器
type PermissionHandler struct {
	permissionService interfaces.PermissionService
}

// NewPermissionHandler 创建权限处理器
func NewPermissionHandler(permissionService interfaces.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
	}
}

// ========================================
// 角色管理接口
// ========================================

// CreateRole 创建角色
// @Summary 创建角色
// @Description 为租户创建新角色
// @Tags Permission
// @Accept json
// @Produce json
// @Param request body types.CreateRoleRequest true "创建角色请求"
// @Success 200 {object} types.Role
// @Router /api/v1/roles [post]
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少租户ID"})
		return
	}

	var req types.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role, err := h.permissionService.CreateRole(c.Request.Context(), tenantID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, role)
}

// GetRoles 获取角色列表
// @Summary 获取角色列表
// @Description 获取租户的角色列表
// @Tags Permission
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} types.RoleListResponse
// @Router /api/v1/roles [get]
func (h *PermissionHandler) GetRoles(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少租户ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	roles, total, err := h.permissionService.GetRoles(c.Request.Context(), tenantID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"roles": roles,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// UpdateRole 更新角色
// @Summary 更新角色
// @Description 更新角色信息
// @Tags Permission
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Param request body types.UpdateRoleRequest true "更新角色请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id} [put]
func (h *PermissionHandler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	var req types.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.permissionService.UpdateRole(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色更新成功"})
}

// DeleteRole 删除角色
// @Summary 删除角色
// @Description 删除角色
// @Tags Permission
// @Accept json
// @Produce json
// @Param id path int true "角色ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/{id} [delete]
func (h *PermissionHandler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的角色ID"})
		return
	}

	err = h.permissionService.DeleteRole(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色删除成功"})
}

// ========================================
// 用户角色管理接口
// ========================================

// AssignRole 给用户分配角色
// @Summary 分配角色
// @Description 给用户分配角色
// @Tags Permission
// @Accept json
// @Produce json
// @Param request body types.AssignRoleRequest true "分配角色请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/roles/assign [post]
func (h *PermissionHandler) AssignRole(c *gin.Context) {
	var req types.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取操作人ID
	operatorID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	err := h.permissionService.AssignRole(c.Request.Context(), &req, operatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色分配成功"})
}

// RevokeRole 撤销用户角色
// @Summary 撤销角色
// @Description 撤销用户的角色
// @Tags Permission
// @Accept json
// @Produce json
// @Param user_id path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/users/{user_id}/role [delete]
func (h *PermissionHandler) RevokeRole(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少租户ID"})
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取操作人ID
	operatorID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	err = h.permissionService.RevokeRole(c.Request.Context(), tenantID, userID, operatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "角色撤销成功"})
}

// ========================================
// 权限查询接口
// ========================================

// GetUserPermissions 获取用户权限列表
// @Summary 获取用户权限
// @Description 获取用户的所有权限
// @Tags Permission
// @Accept json
// @Produce json
// @Param user_id query int false "用户ID（不传则查询当前用户）"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/permissions [get]
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少租户ID"})
		return
	}

	// 获取用户ID
	userIDStr := c.Query("user_id")
	var userID int64
	var err error

	if userIDStr != "" {
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
			return
		}
	} else {
		// 使用当前登录用户ID
		var ok bool
		userID, ok = middleware.GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			return
		}
	}

	permissions, err := h.permissionService.GetUserPermissions(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"permissions": permissions,
	})
}

// CheckPermission 检查用户权限
// @Summary 检查权限
// @Description 检查用户是否有指定权限
// @Tags Permission
// @Accept json
// @Produce json
// @Param request body types.UserPermissionCheckRequest true "权限检查请求"
// @Success 200 {object} types.CheckPermissionResponse
// @Router /api/v1/permissions/check [post]
func (h *PermissionHandler) CheckPermission(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少租户ID"})
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	var req types.UserPermissionCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hasPermission, err := h.permissionService.CheckPermission(c.Request.Context(), tenantID, userID, req.ResourceType, req.Action)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取所有权限列表
	permissions, _ := h.permissionService.GetUserPermissions(c.Request.Context(), tenantID, userID)
	permList := make([]string, len(permissions))
	for i, p := range permissions {
		permList[i] = p.ResourceType + ":" + p.Action
	}

	c.JSON(http.StatusOK, types.CheckPermissionResponse{
		HasPermission: hasPermission,
		Permissions:   permList,
	})
}

// GetUserRole 获取用户角色
// @Summary 获取用户角色
// @Description 获取用户的角色信息
// @Tags Permission
// @Accept json
// @Produce json
// @Param user_id query int false "用户ID（不传则查询当前用户）"
// @Success 200 {object} types.UserRolesView
// @Router /api/v1/role [get]
func (h *PermissionHandler) GetUserRole(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少租户ID"})
		return
	}

	// 获取用户ID
	userIDStr := c.Query("user_id")
	var userID int64

	if userIDStr != "" {
		userID, _ = strconv.ParseInt(userIDStr, 10, 64)
	} else {
		// 使用当前登录用户ID
		var ok bool
		userID, ok = middleware.GetUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			return
		}
	}

	userRole, err := h.permissionService.GetUserRole(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, userRole)
}

// ========================================
// 注册路由
// ========================================

// RegisterPermissionRoutes 注册权限相关路由
func RegisterPermissionRoutes(router *gin.RouterGroup, permissionService interfaces.PermissionService) {
	h := NewPermissionHandler(permissionService)

	// 角色管理路由
	roles := router.Group("/roles")
	{
		roles.GET("", h.GetRoles)           // 获取角色列表
		roles.POST("", h.CreateRole)         // 创建角色
		roles.PUT("/:id", h.UpdateRole)      // 更新角色
		roles.DELETE("/:id", h.DeleteRole)   // 删除角色

		// 分配角色
		roles.POST("/assign", h.AssignRole)  // 分配角色
	}

	// 用户角色路由
	users := router.Group("/users")
	{
		users.DELETE("/:user_id/role", h.RevokeRole) // 撤销角色
		users.GET("/role", h.GetUserRole)             // 获取用户角色
	}

	// 权限查询路由
	permissions := router.Group("/permissions")
	{
		permissions.GET("", h.GetUserPermissions)     // 获取用户权限
		permissions.POST("/check", h.CheckPermission) // 检查权限
	}
}
