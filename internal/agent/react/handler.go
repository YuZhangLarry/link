// Package react 提供 ReAct Agent 的 HTTP Handler
package react

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"

	"link/internal/agent/react"
)

// Handler ReAct Agent HTTP 处理器
type Handler struct {
	agent *react.Agent
}

// NewHandler 创建 ReAct Agent Handler
func NewHandler(chatModel model.ToolCallingChatModel, opts ...react.Option) (*Handler, error) {
	agent, err := react.NewAgent(chatModel, opts...)
	if err != nil {
		return nil, err
	}

	return &Handler{
		agent: agent,
	}, nil
}

// ========================================
// HTTP 处理方法
// ========================================

// ChatRequest 聊天请求
type ChatRequest struct {
	Query     string                 `json:"query" binding:"required"`
	SessionID string                 `json:"session_id,omitempty"`
	Streaming bool                   `json:"streaming"`
	Config    *react.Config          `json:"config,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Success    bool                   `json:"success"`
	Answer     string                 `json:"answer,omitempty"`
	Steps      []*react.RunStep       `json:"steps,omitempty"`
	ToolCalls  []react.ToolCallInfo   `json:"tool_calls,omitempty"`
	SessionID  string                 `json:"session_id,omitempty"`
	DurationMs int64                  `json:"duration_ms"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Chat 处理聊天请求
func (h *Handler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ChatResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	// 如果请求中带了新配置，创建新 Agent
	var agent *react.Agent
	var err error

	if req.Config != nil {
		// TODO: 从已有的 chatModel 创建新 Agent
		// agent, err = react.NewAgent(...)
		agent = h.agent
	} else {
		agent = h.agent
	}

	// 调用 Agent
	resp, err := agent.Chat(ctx, req.Query,
		react.WithSessionID(req.SessionID),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, ChatResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Success:    true,
		Answer:     resp.Answer,
		Steps:      resp.Steps,
		ToolCalls:  resp.ToolCalls,
		SessionID:  resp.SessionID,
		DurationMs: resp.DurationMs,
		Metadata:   resp.Metadata,
	})
}

// ChatStream 处理流式聊天请求
func (h *Handler) ChatStream(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	ctx := c.Request.Context()

	// 调用 Agent 流式接口
	eventChan, err := h.agent.ChatStream(ctx, req.Query,
		react.WithSessionID(req.SessionID),
	)

	if err != nil {
		sendSSEError(c.Writer, err)
		return
	}

	// 发送事件
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	for event := range eventChan {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}

		c.Writer.Write([]byte("data: "))
		c.Writer.Write(data)
		c.Writer.Write([]byte("\n\n"))
		flusher.Flush()
	}

	// 发送结束事件
	c.Writer.Write([]byte("data: [DONE]\n\n"))
	flusher.Flush()
}

// sendSSEError 发送 SSE 错误
func sendSSEError(w http.ResponseWriter, err error) {
	data, _ := json.Marshal(map[string]interface{}{
		"type":  "error",
		"error": err.Error(),
	})
	w.Write([]byte("data: "))
	w.Write(data)
	w.Write([]byte("\n\n"))
}

// ========================================
// 注册路由
// ========================================

// RegisterRoutes 注册 ReAct Agent 路由
func RegisterRoutes(r *gin.RouterGroup, chatModel model.ToolCallingChatModel) error {
	handler, err := NewHandler(chatModel)
	if err != nil {
		return err
	}

	agent := r.Group("/agent")
	{
		agent.POST("/chat", handler.Chat)
		agent.POST("/chat/stream", handler.ChatStream)
	}

	return nil
}

// ========================================
// 工具管理端点
// ========================================

// ListTools 列出可用工具
func (h *Handler) ListTools(c *gin.Context) {
	// TODO: 从 Agent 获取工具列表
	c.JSON(http.StatusOK, gin.H{
		"tools": []string{
			"kb_query",
			"kb_list",
			"document_list",
			"web_search",
			"get_current_time",
			"calculator",
			"http_request",
		},
	})
}

// GetToolInfo 获取工具详细信息
func (h *Handler) GetToolInfo(c *gin.Context) {
	toolName := c.Param("name")

	// TODO: 从 Agent 获取工具信息
	c.JSON(http.StatusOK, gin.H{
		"name":        toolName,
		"description": "Tool description",
	})
}
