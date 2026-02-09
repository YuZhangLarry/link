package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"link/internal/application/service"
	"link/internal/middleware"
	"link/internal/types"
)

// TenantHandler 租户处理器
type TenantHandler struct {
	tenantService *service.TenantService
}

// NewTenantHandler 创建租户处理器
func NewTenantHandler(tenantService *service.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// TenantRequired 租户必填中间件
// @Summary 租户ID验证中间件
// @Description 验证请求头中是否包含租户ID
// @Tags Middleware
// @Accept json
// @Produce json
// @Router /api/v1/ [get]
func (h *TenantHandler) TenantRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-ID")
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    40000,
				"message": "缺少租户ID，请在请求头中添加 X-Tenant-ID",
			})
			c.Abort()
			return
		}

		// 转换为 int64 并设置到上下文
		tenantIDInt, err := strconv.ParseInt(tenantID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    40001,
				"message": "租户ID格式错误",
			})
			c.Abort()
			return
		}

		// 设置租户ID到上下文
		middleware.SetTenantContext(c, &middleware.TenantContext{
			ID: tenantIDInt,
		})

		c.Next()
	}
}

// CreateTenant 创建租户
// @Summary 创建租户
// @Description 创建新租户
// @Tags Tenant
// @Accept json
// @Produce json
// @Param request body types.CreateTenantRequest true "创建租户请求"
// @Success 200 {object} types.TenantResponse
// @Router /api/v1/tenants [post]
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req types.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenant, err := h.tenantService.CreateTenant(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// GetTenant 获取租户信息
// @Summary 获取租户信息
// @Description 根据ID获取租户信息
// @Tags Tenant
// @Accept json
// @Produce json
// @Param id path int true "租户ID"
// @Success 200 {object} types.TenantResponse
// @Router /api/v1/tenants/{id} [get]
func (h *TenantHandler) GetTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的租户ID"})
		return
	}

	tenant, err := h.tenantService.GetTenantByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tenant.ToResponse())
}

// ListTenants 获取租户列表
// @Summary 获取租户列表
// @Description 分页获取租户列表
// @Tags Tenant
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tenants [get]
func (h *TenantHandler) ListTenants(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	tenants, total, err := h.tenantService.ListTenants(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenants": tenants,
		"total": total,
		"page": page,
		"page_size": pageSize,
	})
}

// UpdateTenant 更新租户
// @Summary 更新租户
// @Description 更新租户信息
// @Tags Tenant
// @Accept json
// @Produce json
// @Param id path int true "租户ID"
// @Param request body types.UpdateTenantRequest true "更新租户请求"
// @Success 200 {object} types.TenantResponse
// @Router /api/v1/tenants/{id} [put]
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的租户ID"})
		return
	}

	var req types.UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenant, err := h.tenantService.UpdateTenant(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// DeleteTenant 删除租户
// @Summary 删除租户
// @Description 删除租户（软删除）
// @Tags Tenant
// @Accept json
// @Produce json
// @Param id path int true "租户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tenants/{id} [delete]
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的租户ID"})
		return
	}

	err = h.tenantService.DeleteTenant(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "租户删除成功"})
}

// RegenerateAPIKey 重新生成API Key
// @Summary 重新生成API Key
// @Description 为租户重新生成API Key
// @Tags Tenant
// @Accept json
// @Produce json
// @Param id path int true "租户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tenants/{id}/api-key [post]
func (h *TenantHandler) RegenerateAPIKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的租户ID"})
		return
	}

	apiKey, err := h.tenantService.RegenerateAPIKey(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API Key 重新生成成功",
		"api_key": apiKey,
	})
}

// GetStorageUsage 获取存储使用情况
// @Summary 获取存储使用情况
// @Description 获取租户的存储使用情况
// @Tags Tenant
// @Accept json
// @Produce json
// @Param id path int true "租户ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/tenants/{id}/storage [get]
func (h *TenantHandler) GetStorageUsage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的租户ID"})
		return
	}

	quota, used, percentage, err := h.tenantService.GetStorageUsage(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"storage_quota":     quota,
		"storage_used":      used,
		"usage_percentage":  percentage,
	})
}

// RegisterTenantRoutes 注册租户相关路由
func RegisterTenantRoutes(router *gin.RouterGroup, tenantService *service.TenantService) {
	h := NewTenantHandler(tenantService)

	// 需要认证的路由
	authGroup := router.Group("")
	authGroup.Use(middleware.Auth(nil)) // 需要传入 userService
	{
		authGroup.POST("/tenants", h.CreateTenant)
		authGroup.GET("/tenants", h.ListTenants)
		authGroup.GET("/tenants/:id", h.GetTenant)
		authGroup.PUT("/tenants/:id", h.UpdateTenant)
		authGroup.DELETE("/tenants/:id", h.DeleteTenant)
		authGroup.POST("/tenants/:id/api-key", h.RegenerateAPIKey)
		authGroup.GET("/tenants/:id/storage", h.GetStorageUsage)
	}
}
