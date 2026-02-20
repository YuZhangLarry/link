package handler

import (
	"fmt"
	"link/internal/application/service"
	"link/internal/middleware"
	"link/internal/types"

	"github.com/gin-gonic/gin"
)

// EvaluationHandler 测评处理器
type EvaluationHandler struct {
	evaluationService *service.EvaluationService
}

// NewEvaluationHandler 创建测评处理器
func NewEvaluationHandler(evaluationService *service.EvaluationService) *EvaluationHandler {
	return &EvaluationHandler{
		evaluationService: evaluationService,
	}
}

// CreateEvaluation 创建测评任务
// POST /api/v1/evaluation
func (h *EvaluationHandler) CreateEvaluation(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	var req types.CreateEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 设置默认数据集ID
	if req.DatasetID == "" {
		req.DatasetID = "default"
	}

	detail, err := h.evaluationService.Evaluation(
		c.Request.Context(),
		tenantID,
		req.DatasetID,
		req.KnowledgeBaseID,
		req.ChatID,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    detail,
	})
}

// GetEvaluation 获取测评结果
// GET /api/v1/evaluation?task_id=xxx
func (h *EvaluationHandler) GetEvaluation(c *gin.Context) {
	taskID := c.Query("task_id")
	if taskID == "" {
		c.JSON(400, gin.H{"error": "task_id is required"})
		return
	}

	detail, err := h.evaluationService.EvaluationResult(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    detail,
	})
}

// ListEvaluations 列出测评任务
// GET /api/v1/evaluations
func (h *EvaluationHandler) ListEvaluations(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	// 分页参数
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		if n, err := parsePage(p); err == nil {
			page = n
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if n, err := parsePage(ps); err == nil {
			pageSize = n
		}
	}

	tasks, total, err := h.evaluationService.ListEvaluations(c.Request.Context(), tenantID, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"tasks":     tasks,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetEvaluationByID 获取单个测评任务
// GET /api/v1/evaluations/:id
func (h *EvaluationHandler) GetEvaluationByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "id is required"})
		return
	}

	task, err := h.evaluationService.GetEvaluation(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "evaluation not found"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    task,
	})
}

// DeleteEvaluation 删除测评任务
// DELETE /api/v1/evaluations/:id
func (h *EvaluationHandler) DeleteEvaluation(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(400, gin.H{"error": "id is required"})
		return
	}

	if err := h.evaluationService.DeleteEvaluation(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "evaluation deleted",
	})
}

// ========================================
// 数据集管理
// ========================================

// CreateDataset 创建数据集
// POST /api/v1/datasets
func (h *EvaluationHandler) CreateDataset(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	var req struct {
		DatasetID string          `json:"dataset_id" binding:"required"`
		QAPairs   []*types.QAPair `json:"qapairs" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.evaluationService.CreateDataset(c.Request.Context(), tenantID, req.DatasetID, req.QAPairs); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "dataset created",
	})
}

// ListDatasets 列出数据集
// GET /api/v1/datasets
func (h *EvaluationHandler) ListDatasets(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}

	datasets, err := h.evaluationService.ListDatasets(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    datasets,
	})
}

// parsePage 解析分页参数
func parsePage(s string) (int, error) {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0, err
	}
	if n < 1 {
		return 1, nil
	}
	return n, nil
}
