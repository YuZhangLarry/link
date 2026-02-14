package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"link/internal/application/service"
	"link/internal/middleware"
	"link/internal/models/chat"
	"link/internal/types"
	"link/internal/types/interfaces"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	chatService    *service.ChatService
	sessionService interfaces.SessionService
	messageService interfaces.MessageService
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(
	chatService *service.ChatService,
	sessionService interfaces.SessionService,
	messageService interfaces.MessageService,
) *ChatHandler {
	return &ChatHandler{
		chatService:    chatService,
		sessionService: sessionService,
		messageService: messageService,
	}
}

// Chat 聊天接口（非流式）
// @Summary 聊天对话
// @Description 发送聊天消息并获取回复，自动保存到会话
// @Tags 聊天
// @Accept json
// @Produce json
// @Param request body types.ChatRequest true "聊天请求"
// @Success 200 {object} Response{data=types.ChatResponse}
// @Router /api/v1/chat [post]
func (h *ChatHandler) Chat(c *gin.Context) {
	var req types.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 如果设置了流式，返回错误
	if req.Stream {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "流式聊天请使用 /api/v1/chat/stream 接口",
		})
		return
	}

	// 获取用户ID
	userID := h.getUserID(c)

	// 获取或创建会话ID
	sessionID := h.getSessionID(c, req, userID)

	// 保存用户消息
	h.saveUserMessage(c.Request.Context(), sessionID, &req)

	resp, err := h.chatService.Chat(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "聊天失败",
			"error":   err.Error(),
		})
		return
	}

	// 保存 AI 回复
	h.saveAssistantMessage(c.Request.Context(), sessionID, resp)

	c.JSON(http.StatusOK, gin.H{
		"code":       0,
		"message":    "成功",
		"data":       resp,
		"session_id": sessionID,
	})
}

// ChatStream 流式聊天接口
// @Summary 流式聊天
// @Description 发送聊天消息并以流式方式获取回复，自动保存到会话
// @Tags 聊天
// @Accept json
// @Produce text/event-stream
// @Param request body types.ChatRequest true "聊天请求"
// @Router /api/v1/chat/stream [post]
func (h *ChatHandler) ChatStream(c *gin.Context) {
	log.Printf("🤖 [ChatStream] 收到流式聊天请求")

	var req types.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ [ChatStream] 参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    -1,
			"message": "请求参数错误",
			"error":   err.Error(),
		})
		return
	}

	// 强制设置流式模式
	req.Stream = true

	// 获取用户ID
	userID := h.getUserID(c)
	log.Printf("✅ [ChatStream] 用户ID: %d, 内容: %s", userID, req.Content)

	// 获取或创建会话ID
	sessionID := h.getSessionID(c, req, userID)
	log.Printf("✅ [ChatStream] 会话ID: %s", sessionID)

	// 保存用户消息
	h.saveUserMessage(c.Request.Context(), sessionID, &req)

	log.Printf("📡 [ChatStream] 调用聊天服务...")
	eventChan, err := h.chatService.ChatStream(c.Request.Context(), &req)
	if err != nil {
		log.Printf("❌ [ChatStream] 调用失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1,
			"message": "流式聊天失败",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("✅ [ChatStream] 开始流式响应...")
	// 使用SSE写入器发送流式响应，并保存完整的 AI 回复
	h.handleStreamWithSave(c.Request.Context(), c, sessionID, eventChan)
	log.Printf("✅ [ChatStream] 流式响应完成")
}

// convertToModelEvents 转换为模型事件
func (h *ChatHandler) convertToModelEvents(eventChan <-chan types.StreamChatEvent) <-chan chat.StreamResponse {
	respChan := make(chan chat.StreamResponse, 10)

	go func() {
		defer close(respChan)
		for event := range eventChan {
			respChan <- chat.StreamResponse{
				Event:      event.Event,
				Content:    event.Content,
				MessageID:  event.MessageID,
				TokenCount: event.TokenCount,
				ToolCalls:  h.convertToModelToolCalls(event.ToolCalls),
				Error:      event.Error,
			}
		}
	}()

	return respChan
}

// convertToModelToolCalls 转换为模型工具调用
func (h *ChatHandler) convertToModelToolCalls(calls []types.ToolCall) []chat.ToolCall {
	if calls == nil {
		return nil
	}

	result := make([]chat.ToolCall, len(calls))
	for i, call := range calls {
		result[i] = chat.ToolCall{
			ID:   call.ID,
			Type: call.Type,
			Function: chat.FunctionCall{
				Name:      call.Function.Name,
				Arguments: call.Function.Arguments,
			},
		}
	}
	return result
}

// ChatWithAuth 带认证的聊天接口（可选）
// @Summary 带认证的聊天对话
// @Description 发送聊天消息并获取回复（需要认证）
// @Tags 聊天
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body types.ChatRequest true "聊天请求"
// @Success 200 {object} Response{data=types.ChatResponse}
// @Router /api/v1/chat/auth [post]
func (h *ChatHandler) ChatWithAuth(c *gin.Context) {
	// 获取用户ID（可选）
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "未认证",
		})
		return
	}

	// 可以在这里添加用户相关的逻辑，例如记录用户聊天历史
	_ = userID

	// 调用普通的聊天处理
	h.Chat(c)
}

// ChatStreamWithAuth 带认证的流式聊天接口（可选）
// @Summary 带认证的流式聊天
// @Description 发送聊天消息并以流式方式获取回复（需要认证）
// @Tags 聊天
// @Accept json
// @Produce text/event-stream
// @Security BearerAuth
// @Param request body types.ChatRequest true "聊天请求"
// @Router /api/v1/chat/stream/auth [post]
func (h *ChatHandler) ChatStreamWithAuth(c *gin.Context) {
	// 获取用户ID（可选）
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    -1,
			"message": "未认证",
		})
		return
	}

	// 可以在这里添加用户相关的逻辑，例如记录用户聊天历史
	_ = userID

	// 调用普通的流式聊天处理
	h.ChatStream(c)
}

// ========================================
// 辅助方法
// ========================================

// getUserID 获取用户ID
func (h *ChatHandler) getUserID(c *gin.Context) int64 {
	if userID, exists := middleware.GetUserID(c); exists {
		log.Printf("🔑 [getUserID] 从中间件获取用户ID: %d", userID)
		return userID
	}
	log.Printf("🔑 [getUserID] 使用默认用户ID: 1")
	return 1 // 默认用户ID
}

// getSessionID 获取或创建会话ID
func (h *ChatHandler) getSessionID(c *gin.Context, req types.ChatRequest, userID int64) string {
	// 📊 诊断：打印 context 中的 tenant_id 和 user_id
	if tenantID, exists := c.Get("tenant_id"); exists {
		log.Printf("🔍 [getSessionID] Gin Context tenant_id = %v (type: %T)", tenantID, tenantID)
	} else {
		log.Printf("⚠️ [getSessionID] Gin Context 中没有 tenant_id")
	}
	if uid, exists := c.Get("user_id"); exists {
		log.Printf("🔍 [getSessionID] Gin Context user_id = %v (type: %T)", uid, uid)
	} else {
		log.Printf("⚠️ [getSessionID] Gin Context 中没有 user_id")
	}
	// 检查 request.Context() 中的值
	if ctxTenantID := c.Request.Context().Value("tenant_id"); ctxTenantID != nil {
		log.Printf("🔍 [getSessionID] Request.Context tenant_id = %v (type: %T)", ctxTenantID, ctxTenantID)
	} else {
		log.Printf("⚠️ [getSessionID] Request.Context 中没有 tenant_id")
	}

	// 优先从请求体获取会话ID
	if req.SessionID != "" {
		log.Printf("📌 [getSessionID] 从请求体获取会话ID: %s", req.SessionID)
		return req.SessionID
	}

	// 其次尝试从请求头获取会话ID（兼容旧版）
	if sessionID := c.GetHeader("X-Session-ID"); sessionID != "" {
		log.Printf("📌 [getSessionID] 从请求头获取会话ID: %s", sessionID)
		return sessionID
	}

	// 如果没有会话ID，创建新会话
	log.Printf("➕ [getSessionID] 创建新会话...")
	session, err := h.sessionService.CreateSession(c.Request.Context(), userID, &types.CreateSessionRequest{
		Title:       generateSessionTitle(req.Content),
		Description: "自动创建的会话",
		MaxRounds:   50,
	})
	if err != nil {
		log.Printf("❌ [getSessionID] 创建会话失败: %v", err)
		return "" // 创建失败返回空字符串
	}

	log.Printf("✅ [getSessionID] 新会话创建成功: ID=%s, TenantID=%d, UserID=%d", session.ID, session.TenantID, session.UserID)
	return session.ID
}

// saveUserMessage 保存用户消息
func (h *ChatHandler) saveUserMessage(ctx context.Context, sessionID string, req *types.ChatRequest) {
	if sessionID == "" {
		log.Printf("⚠️ [saveUserMessage] sessionID 为空，跳过保存")
		return
	}

	log.Printf("💾 [saveUserMessage] 保存用户消息: sessionID=%s, content=%s", sessionID, req.Content[:min(20, len(req.Content))]+"...")

	// 用户消息没有 tool_calls，不需要传
	_, err := h.messageService.CreateMessage(ctx, &types.CreateMessageRequest{
		SessionID:  sessionID,
		Role:       "user",
		Content:    req.Content,
		TokenCount: len(req.Content) / 3, // 简单估算
		ToolCalls:  "",                   // 空字符串，数据库层会处理
	})
	if err != nil {
		log.Printf("❌ [saveUserMessage] 保存失败: %v", err)
	} else {
		log.Printf("✅ [saveUserMessage] 保存成功")
	}
}

// saveAssistantMessage 保存 AI 回复
func (h *ChatHandler) saveAssistantMessage(ctx context.Context, sessionID string, resp *types.ChatResponse) {
	if sessionID == "" {
		return
	}

	// 序列化 tool_calls
	var toolCallsJSON string
	if len(resp.ToolCalls) > 0 {
		data, _ := json.Marshal(resp.ToolCalls)
		toolCallsJSON = string(data)
	}

	h.messageService.CreateMessage(ctx, &types.CreateMessageRequest{
		SessionID:  sessionID,
		Role:       resp.Role,
		Content:    resp.Content,
		ToolCalls:  toolCallsJSON,
		TokenCount: resp.TokenCount,
	})
}

// handleStreamWithSave 处理流式响应并保存完整内容
func (h *ChatHandler) handleStreamWithSave(ctx context.Context, c *gin.Context, sessionID string, eventChan <-chan types.StreamChatEvent) {
	sseWriter := chat.NewSSEResponseWriter(c)
	defer sseWriter.Close()

	// 首先发送 session_id 给前端
	if sessionID != "" {
		sessionData := gin.H{"session_id": sessionID}
		if err := sseWriter.WriteEvent("session", sessionData); err != nil {
			log.Printf("❌ [handleStreamWithSave] 发送session_id失败: %v", err)
		}
	}

	var fullContent string
	var totalTokenCount int
	var toolCalls []types.ToolCall

	for event := range eventChan {
		// 转换并发送事件
		modelEvent := chat.StreamResponse{
			Event:      event.Event,
			Content:    event.Content,
			MessageID:  event.MessageID,
			TokenCount: event.TokenCount,
			ToolCalls:  h.convertToModelToolCalls(event.ToolCalls),
			Error:      event.Error,
		}

		if err := sseWriter.WriteEvent(event.Event, modelEvent); err != nil {
			return
		}

		// 累积内容和 TokenCount
		if event.Event == "content" {
			fullContent += event.Content
			// 累加 TokenCount（如果提供了）
			if event.TokenCount > 0 {
				totalTokenCount += event.TokenCount
			}
		} else if event.Event == "end" && len(event.ToolCalls) > 0 {
			toolCalls = event.ToolCalls
		}
	}

	// 保存完整的 AI 回复
	if sessionID != "" && fullContent != "" {
		var toolCallsJSON string
		if len(toolCalls) > 0 {
			data, _ := json.Marshal(toolCalls)
			toolCallsJSON = string(data)
		}

		h.messageService.CreateMessage(ctx, &types.CreateMessageRequest{
			SessionID:  sessionID,
			Role:       "assistant",
			Content:    fullContent,
			ToolCalls:  toolCallsJSON,
			TokenCount: totalTokenCount,
		})
	}
}

// generateSessionTitle 生成会话标题
func generateSessionTitle(content string) string {
	if len(content) > 30 {
		return content[:30] + "..."
	}
	return content
}

// parseInt 字符串转整数
func parseInt(s string) int64 {
	var result int64
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			result = result*10 + int64(ch-'0')
		} else {
			break
		}
	}
	return result
}
