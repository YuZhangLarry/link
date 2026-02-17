package handler

import (
	"link/internal/application/repository"
	"link/internal/middleware"
	"link/internal/types"

	"github.com/gin-gonic/gin"
)

// ModelHandler 模型处理器
type ModelHandler struct {
	modelRepo repository.ModelRepository
}

// NewModelHandler 创建模型处理器
func NewModelHandler(modelRepo repository.ModelRepository) *ModelHandler {
	return &ModelHandler{
		modelRepo: modelRepo,
	}
}

// GetList 获取模型列表
// GET /api/v1/models?type=embedding
func (h *ModelHandler) GetList(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	modelType := c.Query("type")

	var models []*types.Model
	var err error

	if modelType != "" {
		models, err = h.modelRepo.FindByType(c.Request.Context(), tenantID, modelType)
	} else {
		models, err = h.modelRepo.FindByTenantID(c.Request.Context(), tenantID)
	}

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    models,
	})
}

// GetByID 获取单个模型
// GET /api/v1/models/:id
func (h *ModelHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	model, err := h.modelRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "model not found"})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    model,
	})
}
