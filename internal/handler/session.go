package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"link/internal/middleware"
	"link/internal/types"
	"link/internal/types/interfaces"
)

// SessionHandler 会话处理器
type SessionHandler struct {
	sessionService interfaces.SessionService
}

// NewSessionHandler 创建会话处理器
func NewSessionHandler(sessionService interfaces.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// CreateSession 创建会话
// @Summary 创建会话
// @Description 创建新的聊天会话
// @Tags 会话
// @Accept json
// @Produce json
// @Param request body types.CreateSessionRequest true "创建会话请求"
// @Success 200 {object} Response{data=types.SessionResponse}
// @Router /api/v1/sessions [post]
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req types.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ [CreateSession] 参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 从认证中间件获取用户ID
	userID, ok := middleware.GetUserID(c)
	if !ok || userID == 0 {
		log.Printf("❌ [CreateSession] 未找到用户ID")
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "未授权",
		})
		return
	}
	log.Printf("✅ [CreateSession] 用户ID: %d, 标题: %s", userID, req.Title)

	resp, err := h.sessionService.CreateSession(c.Request.Context(), userID, &req)
	if err != nil {
		log.Printf("❌ [CreateSession] 创建失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "创建会话失败",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("✅ [CreateSession] 创建成功: ID=%s", resp.ID)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    resp,
	})
}

// GetSessionByID 根据ID获取会话
// @Summary 获取会话详情
// @Description 根据ID获取会话详情
// @Tags 会话
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} Response{data=types.SessionResponse}
// @Router /api/v1/sessions/{id} [get]
func (h *SessionHandler) GetSessionByID(c *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		log.Printf("❌ [GetSessionByID] 参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("📖 [GetSessionByID] 查询会话: ID=%s", uri.ID)

	resp, err := h.sessionService.GetSessionByID(c.Request.Context(), uri.ID)
	if err != nil {
		log.Printf("❌ [GetSessionByID] 查询失败: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": "会话不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    resp,
	})
}

// GetSessionDetail 获取会话详情（包含消息）
// @Summary 获取会话完整详情
// @Description 获取会话详情及消息列表
// @Tags 会话
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} Response{data=types.SessionDetailResponse}
// @Router /api/v1/sessions/{id}/detail [get]
func (h *SessionHandler) GetSessionDetail(c *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.sessionService.GetSessionDetail(c.Request.Context(), uri.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": "会话不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data": gin.H{
			"session":  resp.Session,
			"messages": resp.Messages,
		},
	})
}

// ListSessions 查询会话列表
// @Summary 获取会话列表
// @Description 查询用户的会话列表
// @Tags 会话
// @Accept json
// @Produce json
// @Param user_id query int true "用户ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Param status query int false "状态筛选" Enums(0, 1)
// @Success 200 {object} Response{data=types.SessionListResponse}
// @Router /api/v1/sessions [get]
func (h *SessionHandler) ListSessions(c *gin.Context) {
	var req types.ListSessionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("❌ [ListSessions] 参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 用户ID从认证中间件获取
	log.Printf("📋 [ListSessions] 查询会话列表: Page=%d, Size=%d", req.Page, req.Size)

	resp, err := h.sessionService.ListSessions(c.Request.Context(), &req)
	if err != nil {
		log.Printf("❌ [ListSessions] 查询失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "查询会话列表失败",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("✅ [ListSessions] 查询成功: 总数=%d", resp.Total)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    resp,
	})
}

// UpdateSession 更新会话
// @Summary 更新会话
// @Description 更新会话信息
// @Tags 会话
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Param request body types.UpdateSessionRequest true "更新会话请求"
// @Success 200 {object} Response{data=types.SessionResponse}
// @Router /api/v1/sessions/{id} [put]
func (h *SessionHandler) UpdateSession(c *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	var req types.UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.sessionService.UpdateSession(c.Request.Context(), uri.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "更新会话失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "成功",
		"data":    resp,
	})
}

// DeleteSession 删除会话
// @Summary 删除会话
// @Description 删除指定会话（软删除）
// @Tags 会话
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} Response
// @Router /api/v1/sessions/{id} [delete]
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("🗑️ [DeleteSession] 删除会话: ID=%s", uri.ID)

	err := h.sessionService.DeleteSession(c.Request.Context(), uri.ID)
	if err != nil {
		log.Printf("❌ [DeleteSession] 删除失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "删除会话失败",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("✅ [DeleteSession] 删除成功")
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}

// ArchiveSession 归档会话
// @Summary 归档会话
// @Description 归档指定会话
// @Tags 会话
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} Response
// @Router /api/v1/sessions/{id}/archive [post]
func (h *SessionHandler) ArchiveSession(c *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	err := h.sessionService.ArchiveSession(c.Request.Context(), uri.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "归档会话失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "归档成功",
	})
}

// ActivateSession 激活会话
// @Summary 激活会话
// @Description 激活已归档的会话
// @Tags 会话
// @Accept json
// @Produce json
// @Param id path string true "会话ID"
// @Success 200 {object} Response
// @Router /api/v1/sessions/{id}/activate [post]
func (h *SessionHandler) ActivateSession(c *gin.Context) {
	var uri struct {
		ID string `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	err := h.sessionService.ActivateSession(c.Request.Context(), uri.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "激活会话失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "激活成功",
	})
}
