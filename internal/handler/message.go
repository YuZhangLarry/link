package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// MessageHandler 消息处理器
type MessageHandler struct {
	messageService interfaces.MessageService
}

// NewMessageHandler 创建消息处理器
func NewMessageHandler(messageService interfaces.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// CreateMessage 创建消息
// @Summary 创建消息
// @Description 创建新的聊天消息
// @Tags 消息
// @Accept json
// @Produce json
// @Param request body types.CreateMessageRequest true "创建消息请求"
// @Success 200 {object} Response{data=types.MessageResponse}
// @Router /api/v1/messages [post]
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	var req types.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.messageService.CreateMessage(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "创建消息失败",
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

// GetMessageByID 根据ID获取消息
// @Summary 获取消息详情
// @Description 根据ID获取消息详情
// @Tags 消息
// @Accept json
// @Produce json
// @Param id path int true "消息ID"
// @Success 200 {object} Response{data=types.MessageResponse}
// @Router /api/v1/messages/{id} [get]
func (h *MessageHandler) GetMessageByID(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.messageService.GetMessageByID(c.Request.Context(), uri.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    -1,
			"message": "消息不存在",
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

// ListMessages 查询消息列表
// @Summary 获取消息列表
// @Description 查询对话的消息列表
// @Tags 消息
// @Accept json
// @Produce json
// @Param chat_id query int true "对话ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Param role query string false "角色筛选" Enums(system, user, assistant, tool)
// @Success 200 {object} Response{data=types.MessageListResponse}
// @Router /api/v1/messages [get]
func (h *MessageHandler) ListMessages(c *gin.Context) {
	var req types.ListMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.messageService.ListMessages(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "查询消息列表失败",
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

// UpdateMessage 更新消息
// @Summary 更新消息
// @Description 更新消息内容
// @Tags 消息
// @Accept json
// @Produce json
// @Param id path int true "消息ID"
// @Param request body types.UpdateMessageRequest true "更新消息请求"
// @Success 200 {object} Response{data=types.MessageResponse}
// @Router /api/v1/messages/{id} [put]
func (h *MessageHandler) UpdateMessage(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	var req types.UpdateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	resp, err := h.messageService.UpdateMessage(c.Request.Context(), uri.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "更新消息失败",
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

// DeleteMessage 删除消息
// @Summary 删除消息
// @Description 删除指定消息
// @Tags 消息
// @Accept json
// @Produce json
// @Param id path int true "消息ID"
// @Success 200 {object} Response
// @Router /api/v1/messages/{id} [delete]
func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	err := h.messageService.DeleteMessage(c.Request.Context(), uri.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "删除消息失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "删除成功",
	})
}
