// Package handler 提供 Agent 相关的 HTTP 处理器
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"link/internal/application/service/agent"
	"link/internal/middleware"
	"link/internal/types"
	"link/internal/types/interfaces"
)

// ========================================
// Agent 接口定义
// ========================================

// EinoAgentProvider 提供获取 Eino Agent 的接口
// DeepSearchAgent 和 MultiAgentOrchestrator 都实现这个接口
type EinoAgentProvider interface {
	GetEinoAgent() adk.Agent
}

// ========================================
// Agent Handler
// ========================================

// AgentHandler Agent 处理器
type AgentHandler struct {
	agent          EinoAgentProvider
	sessionService interfaces.SessionService
	messageService interfaces.MessageService
}

// NewAgentHandler 创建 Agent Handler（支持 DeepSearchAgent 和 MultiAgentOrchestrator）
func NewAgentHandler(
	agentProvider EinoAgentProvider,
	sessionService interfaces.SessionService,
	messageService interfaces.MessageService,
) *AgentHandler {
	return &AgentHandler{
		agent:          agentProvider,
		sessionService: sessionService,
		messageService: messageService,
	}
}

// NewAgentHandlerWithDeepSearch 使用 DeepSearchAgent 创建 Handler（向后兼容）
func NewAgentHandlerWithDeepSearch(
	deepSearchAgent *agent.DeepSearchAgent,
	sessionService interfaces.SessionService,
	messageService interfaces.MessageService,
) *AgentHandler {
	return &AgentHandler{
		agent:          deepSearchAgent,
		sessionService: sessionService,
		messageService: messageService,
	}
}

// ChatRequest Agent 聊天请求
type ChatRequest struct {
	Query string `json:"query" binding:"required"`
}

// ChatResponse Agent 聊天响应
type ChatResponse struct {
	Answer    string                 `json:"answer"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	ToolCalls []*ToolCallInfo        `json:"tool_calls,omitempty"`
	Sources   []string               `json:"sources,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCallInfo 工具调用信息
type ToolCallInfo struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Input    string                 `json:"input"`
	Output   string                 `json:"output,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChatStreamRequest 流式聊天请求
type ChatStreamRequest struct {
	Query     string `json:"query" binding:"required"`
	SessionID string `json:"session_id,omitempty"`
}

// Chat 处理 Agent 聊天请求
func (h *AgentHandler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ChatResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	// 使用 Eino Runner 运行 Agent
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           h.agent.GetEinoAgent(),
		EnableStreaming: false,
	})

	messages := []adk.Message{schema.UserMessage(req.Query)}
	iter := runner.Run(ctx, messages)

	var answer strings.Builder
	var success = true
	var errMsg string
	toolCalls := make([]*ToolCallInfo, 0)

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			success = false
			errMsg = event.Err.Error()
			break
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err == nil {
				// 记录工具调用
				for _, tc := range msg.ToolCalls {
					toolCalls = append(toolCalls, &ToolCallInfo{
						ID:    tc.ID,
						Name:  tc.Function.Name,
						Input: tc.Function.Arguments,
					})
				}

				// 收集最终答案
				if (msg.Role == schema.Assistant || msg.Role == "") && len(msg.ToolCalls) == 0 && msg.Content != "" {
					answer.WriteString(msg.Content)
				}
			}
		}

		if event.Action != nil && event.Action.Exit {
			break
		}
	}

	if errMsg != "" {
		c.JSON(http.StatusInternalServerError, ChatResponse{
			Success: false,
			Error:   errMsg,
		})
		return
	}

	c.JSON(http.StatusOK, ChatResponse{
		Answer:    answer.String(),
		Success:   success,
		ToolCalls: toolCalls,
	})
}

// ChatStream 流式聊天，实时输出思考过程
func (h *AgentHandler) ChatStream(c *gin.Context) {
	var req ChatStreamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	userID := h.getUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	// 获取会话ID（必须提供）
	sessionID := h.getOrCreateSessionID(ctx, req, userID)
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id 是必需的"})
		return
	}
	log.Printf("🤖 [AgentChatStream] 会话ID: %s, 查询: %s", sessionID, req.Query)

	// 保存用户消息
	h.saveUserMessage(ctx, sessionID, req.Query)

	// 收集 Agent 步骤（仅用于实时展示，不持久化）
	agentSteps := make([]map[string]interface{}, 0)

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// 首先发送 session_id 给前端
	sendSSEvent(c.Writer, "session", map[string]interface{}{"session_id": sessionID})

	// 使用 Eino ADK Runner 流式运行
	log.Printf("🤖 [AgentChatStream] 开始执行 Agent 查询: %s", req.Query)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           h.agent.GetEinoAgent(),
		EnableStreaming: true,
	})

	messages := []adk.Message{schema.UserMessage(req.Query)}
	log.Printf("📋 [AgentChatStream] 开始迭代处理...")
	iter := runner.Run(ctx, messages)

	stepCount := 0
	var finalAnswer strings.Builder
	eventCount := 0

	// 跟踪当前活跃的工具调用
	activeToolCalls := make(map[string]*toolCallInfo)

	for {
		event, ok := iter.Next()
		if !ok {
			log.Printf("📋 [AgentChatStream] 迭代结束，共处理 %d 个事件", eventCount)
			break
		}
		eventCount++

		// 检查是否应该退出
		if event.Action != nil && event.Action.Exit {
			log.Printf("🔚 [AgentChatStream] 收到退出信号")
			break
		}

		log.Printf("📋 [AgentChatStream] 事件 #%d: Err=%v, Output=%v, Action=%v",
			eventCount, event.Err, event.Output != nil, event.Action != nil)

		// 发生错误
		if event.Err != nil {
			errorMsg := event.Err.Error()
			// 检查是否是内容过滤错误
			if strings.Contains(errorMsg, "content_filter") {
				errorMsg = "抱歉，您的问题触发了内容安全策略，请修改后重试。"
			}
			sendSSEvent(c.Writer, "error", map[string]interface{}{
				"step":    stepCount,
				"type":    "error",
				"content": errorMsg,
			})
			// 保存错误消息
			h.saveAssistantMessage(ctx, sessionID, "", fmt.Sprintf("Error: %v", event.Err.Error()), agentSteps)
			return
		}

		// 处理输出事件
		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				log.Printf("⚠️  [AgentChatStream] 获取消息失败: %v", err)
				continue
			}

			log.Printf("📝 [AgentChatStream] 收到消息: Role=%s, Content=%q, ToolCalls=%d",
				msg.Role, truncateString(msg.Content, 100), len(msg.ToolCalls))

			// 有工具调用 - 输出 action 步骤
			if len(msg.ToolCalls) > 0 {
				log.Printf("🔧 [AgentChatStream] 有 %d 个工具调用", len(msg.ToolCalls))

				// 先发送思考步骤（如果有内容）
				if msg.Content != "" {
					stepCount++
					thoughtData := map[string]interface{}{
						"step":    stepCount,
						"type":    "thinking",
						"content": msg.Content,
					}
					agentSteps = append(agentSteps, thoughtData)
					sendSSEvent(c.Writer, "step", thoughtData)
				}

				// 发送工具调用步骤
				for _, tc := range msg.ToolCalls {
					// 解析工具参数
					var params map[string]interface{}
					json.Unmarshal([]byte(tc.Function.Arguments), &params)

					// 获取工具信息
					toolType, toolDesc, stage := getToolInfo(tc.Function.Name)
					isAgentTool := isAgentTool(tc.Function.Name)

					stepCount++
					stepData := map[string]interface{}{
						"step":      stepCount,
						"type":      toolType,
						"stage":     stage,
						"tool_name": tc.Function.Name,
						"tool_desc": toolDesc,
						"tool_id":   tc.ID,
						"is_agent":  isAgentTool,
					}
					if params != nil {
						stepData["tool_params"] = params
					}

					// 如果是子 Agent 调用，添加额外信息
					if isAgentTool {
						agentName := getAgentDisplayName(tc.Function.Name)
						stepData["agent_name"] = agentName
						stepData["agent_stage"] = getAgentStage(tc.Function.Name)
					}

					agentSteps = append(agentSteps, stepData)

					log.Printf("🔧 [AgentChatStream] 发送工具调用: %s (类型: %s, 阶段: %s)",
						tc.Function.Name, toolType, stage)

					sendSSEvent(c.Writer, "step", stepData)

					// 记录活跃的工具调用，等待结果
					activeToolCalls[tc.ID] = &toolCallInfo{
						ID:      tc.ID,
						Name:    tc.Function.Name,
						Params:  params,
						Step:    stepCount,
						IsAgent: isAgentTool,
					}
				}
			} else if (msg.Role == schema.Assistant || msg.Role == "") && msg.Content != "" {
				log.Printf("💬 [AgentChatStream] 收到助手回复: Role=%q, Content=%q",
					msg.Role, truncateString(msg.Content, 100))

				// 确定步骤类型
				stepType := determineStepType(msg.Content)

				// 检查是否是工具调用的结果
				if isToolResult(msg.Content) {
					// 尝试匹配到之前的工具调用
					stepType = "tool_result"
				}

				// 收集最终答案或中间步骤
				if isFinalAnswer(msg.Content) {
					finalAnswer.WriteString(msg.Content)
				}

				stepCount++
				stepData := map[string]interface{}{
					"step":    stepCount,
					"type":    stepType,
					"content": msg.Content,
				}

				// 如果是工具结果，尝试关联到工具调用
				if stepType == "tool_result" {
					// 查找最近的活跃工具调用
					for _, tc := range activeToolCalls {
						stepData["related_tool"] = tc.Name
						stepData["related_step"] = tc.Step
						break
					}
				}

				agentSteps = append(agentSteps, stepData)

				sendSSEvent(c.Writer, "step", stepData)

				// 清理已完成的工具调用
				if stepType == "tool_result" || stepType == "agent_output" {
					activeToolCalls = make(map[string]*toolCallInfo)
				}
			}
		}

		// 处理 Action 事件
		if event.Action != nil && !event.Action.Exit {
			log.Printf("📋 [AgentChatStream] 收到非退出 Action 事件")
			// 可以在这里处理其他类型的 Action
		}
	}

	// 循环结束，发送完成事件并保存消息
	answer := finalAnswer.String()
	if answer == "" {
		answer = "执行完成"
	}

	log.Printf("✅ [AgentChatStream] 迭代完成，发送 done 事件，answer_len=%d", len(answer))

	// 添加完成步骤
	stepCount++
	completeStep := map[string]interface{}{
		"step":   stepCount,
		"type":   "complete",
		"reason": "Agent 完成执行",
	}
	agentSteps = append(agentSteps, completeStep)

	sendSSEvent(c.Writer, "done", map[string]interface{}{
		"step":       stepCount,
		"type":       "complete",
		"reason":     "Agent 完成执行",
		"answer":     answer,
		"step_count": stepCount,
	})

	// 保存助手消息（包含 agent_steps）
	h.saveAssistantMessage(ctx, sessionID, answer, "", agentSteps)

	// 刷新缓冲区，确保数据发送
	c.Writer.Flush()
}

// getUserID 获取用户ID
func (h *AgentHandler) getUserID(c *gin.Context) int64 {
	// 使用 middleware 的 GetUserID 函数获取用户 ID
	if uid, exists := middleware.GetUserID(c); exists {
		return uid
	}
	// 如果没有找到用户 ID，返回 0 表示未认证
	// 不再使用默认值 1，避免创建错误用户的会话
	log.Printf("⚠️  [AgentHandler] 未获取到用户 ID，请求可能未通过认证中间件")
	return 0
}

// getOrCreateSessionID 获取会话ID（不再自动创建，要求前端必须提供）
func (h *AgentHandler) getOrCreateSessionID(ctx context.Context, req ChatStreamRequest, userID int64) string {
	// 必须从请求获取会话ID
	if req.SessionID != "" {
		return req.SessionID
	}

	// 没有提供 session_id，记录错误
	log.Printf("❌ [AgentChatStream] 未提供 session_id，拒绝请求")
	return ""
}

// saveUserMessage 保存用户消息
func (h *AgentHandler) saveUserMessage(ctx context.Context, sessionID string, content string) {
	if sessionID == "" {
		return
	}

	log.Printf("💾 [AgentChatStream] 保存用户消息: sessionID=%s", sessionID)

	_, err := h.messageService.CreateMessage(ctx, &types.CreateMessageRequest{
		SessionID:  sessionID,
		Role:       "user",
		Content:    content,
		TokenCount: len(content) / 3,
	})
	if err != nil {
		log.Printf("❌ [AgentChatStream] 保存用户消息失败: %v", err)
	}
}

// saveAssistantMessage 保存助手消息（不持久化 agent_steps）
func (h *AgentHandler) saveAssistantMessage(ctx context.Context, sessionID string, content string, errorMsg string, agentSteps []map[string]interface{}) {
	if sessionID == "" {
		return
	}

	// 如果有错误，保存错误信息
	finalContent := content
	if errorMsg != "" {
		finalContent = errorMsg
	}

	log.Printf("💾 [AgentChatStream] 保存助手消息: sessionID=%s", sessionID)

	_, err := h.messageService.CreateMessage(ctx, &types.CreateMessageRequest{
		SessionID:  sessionID,
		Role:       "assistant",
		Content:    finalContent,
		TokenCount: len(finalContent) / 3,
		AgentSteps: "", // 不持久化 agent_steps
	})
	if err != nil {
		log.Printf("❌ [AgentChatStream] 保存助手消息失败: %v", err)
	}
}

// ListTools 列出可用工具
func (h *AgentHandler) ListTools(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"tools": []string{
			"rag_query",
			"web_search",
			"get_current_time",
			"calculator",
			"http_request",
		},
	})
}

// ========================================
// EinoAgentHandler 直接使用 Eino Agent 的 Handler
// ========================================

// EinoAgentHandler Eino Agent Handler
type EinoAgentHandler struct {
	agent adk.Agent
}

// NewEinoAgentHandler 创建 Eino Agent Handler
func NewEinoAgentHandler(einoAgent adk.Agent) *EinoAgentHandler {
	return &EinoAgentHandler{
		agent: einoAgent,
	}
}

// Chat 处理聊天请求
func (h *EinoAgentHandler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ChatResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	// 使用 Eino ADK Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           h.agent,
		EnableStreaming: false,
	})

	messages := []adk.Message{schema.UserMessage(req.Query)}
	iter := runner.Run(ctx, messages)

	var answer strings.Builder
	success := true
	var errMsg string

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			success = false
			errMsg = event.Err.Error()
			break
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err == nil && msg.Role == schema.Assistant && len(msg.ToolCalls) == 0 {
				answer.WriteString(msg.Content)
			}
		}

		if event.Action != nil && event.Action.Exit {
			break
		}
	}

	c.JSON(http.StatusOK, ChatResponse{
		Answer:  answer.String(),
		Success: success,
		Error:   errMsg,
	})
}

// sendSSEvent 发送 SSE 事件
func sendSSEvent(w io.Writer, eventType string, data map[string]interface{}) error {
	// 构建 SSE 事件
	var event strings.Builder
	event.WriteString("event: ")
	event.WriteString(eventType)
	event.WriteString("\n")

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	event.WriteString("data: ")
	event.WriteString(string(jsonData))
	event.WriteString("\n\n")

	_, err = w.Write([]byte(event.String()))
	if err != nil {
		return err
	}

	// 刷新缓冲区，确保实时发送
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// ========================================
// 辅助类型定义
// ========================================

// toolCallInfo 工具调用信息
type toolCallInfo struct {
	ID      string
	Name    string
	Params  map[string]interface{}
	Step    int
	IsAgent bool
}

// ========================================
// 辅助函数 - 工具类型和阶段判断
// ========================================

// getToolInfo 获取工具的详细信息
// 返回: (工具类型, 工具描述, 阶段)
func getToolInfo(toolName string) (toolType, toolDesc, stage string) {
	switch toolName {
	case "rag_query":
		return "search", "知识库检索", "信息检索"
	case "web_search":
		return "search", "网络搜索", "信息检索"
	case "calculator":
		return "utility", "计算器", "工具调用"
	case "get_current_time":
		return "utility", "获取时间", "工具调用"
	case "http_request":
		return "utility", "HTTP请求", "工具调用"
	default:
		// 检查是否是 Agent 工具
		if isAgentTool(toolName) {
			return "agent_call", "子代理调用", getAgentStage(toolName)
		}
		return "action", "工具调用", "其他"
	}
}

// isAgentTool 判断是否是子 Agent 工具
func isAgentTool(toolName string) bool {
	agentTools := []string{
		"planner", "planner_agent",
		"retriever", "retriever_agent",
		"analyzer", "analyzer_agent",
		"synthesizer", "synthesizer_agent",
		"critic", "critic_agent",
	}
	for _, at := range agentTools {
		if toolName == at {
			return true
		}
	}
	return false
}

// getAgentDisplayName 获取 Agent 的显示名称
func getAgentDisplayName(toolName string) string {
	switch toolName {
	case "planner", "planner_agent":
		return "规划代理 (Planner)"
	case "retriever", "retriever_agent":
		return "检索代理 (Retriever)"
	case "analyzer", "analyzer_agent":
		return "分析代理 (Analyzer)"
	case "synthesizer", "synthesizer_agent":
		return "合成代理 (Synthesizer)"
	case "critic", "critic_agent":
		return "评审代理 (Critic)"
	default:
		return toolName
	}
}

// getAgentStage 获取 Agent 所属阶段
func getAgentStage(toolName string) string {
	switch toolName {
	case "planner", "planner_agent":
		return "规划阶段"
	case "retriever", "retriever_agent":
		return "检索阶段"
	case "analyzer", "analyzer_agent":
		return "分析阶段"
	case "synthesizer", "synthesizer_agent":
		return "合成阶段"
	case "critic", "critic_agent":
		return "评审阶段"
	default:
		return "处理阶段"
	}
}

// determineStepType 根据内容确定步骤类型
func determineStepType(content string) string {
	// 按优先级检查

	// 1. 检查是否是规划内容
	if isPlanContent(content) {
		return "plan"
	}

	// 2. 检查是否是分析内容
	if isAnalysisContent(content) {
		return "analysis"
	}

	// 3. 检查是否是评审内容
	if isReviewContent(content) {
		return "review"
	}

	// 4. 检查是否是合成/报告内容
	if isSynthesisContent(content) {
		return "synthesis"
	}

	// 5. 检查是否是检索内容
	if isRetrievalContent(content) {
		return "retrieval"
	}

	// 6. 默认为思考
	return "thought"
}

// isPlanContent 判断是否是规划内容
func isPlanContent(content string) bool {
	keywords := []string{
		"研究目标", "子任务", "关键词", "数据来源", "待验证假设",
		"研究计划", "任务分解", "执行计划", "搜索策略",
		"研究目标", "subtask", "keyword", "data source",
	}
	return containsAny(content, keywords)
}

// isAnalysisContent 判断是否是分析内容
func isAnalysisContent(content string) bool {
	keywords := []string{
		"关键洞见", "事实提取", "矛盾", "一致性", "置信度评估",
		"综合分析", "交叉验证", "分析结果", "key insight",
		"analysis", "fact", "inconsistency",
	}
	return containsAny(content, keywords)
}

// isReviewContent 判断是否是评审内容
func isReviewContent(content string) bool {
	keywords := []string{
		"评审", "评分", "准确性", "完整性", "逻辑性",
		"改进建议", "修订建议", "质量评估", "review",
		"score", "improvement", "accuracy",
	}
	return containsAny(content, keywords)
}

// isSynthesisContent 判断是否是合成/报告内容
func isSynthesisContent(content string) bool {
	keywords := []string{
		"执行摘要", "研究背景", "核心发现", "详细分析",
		"结论", "报告", "整合", "synthesis",
		"executive summary", "key findings", "report",
	}
	// 同时检查是否有明显的结构标记
	hasStructure := strings.Contains(content, "###") ||
		strings.Contains(content, "####") ||
		strings.Contains(content, "##")
	return hasStructure || containsAny(content, keywords)
}

// isRetrievalContent 判断是否是检索内容
func isRetrievalContent(content string) bool {
	keywords := []string{
		"检索结果", "查询结果", "搜索结果", "初步结论",
		"信息质量", "数据来源", "retrieval", "search result",
	}
	return containsAny(content, keywords)
}

// isToolResult 判断是否是工具结果
func isToolResult(content string) bool {
	// 工具结果通常较短，且包含特定标记
	if len(content) < 500 {
		return containsAny(content, []string{"结果", "完成", "success", "done", "完成"})
	}
	return false
}

// isReflectionResult 判断是否是反思结果（保留兼容）
func isReflectionResult(content string) bool {
	keywords := []string{"信息质量评分", "数据来源", "反思", "校验", "信息充分", "缺失信息"}
	return containsAny(content, keywords)
}

// isSearchPlan 判断是否是搜索计划（保留兼容）
func isSearchPlan(content string) bool {
	return isPlanContent(content)
}

// isFinalAnswer 判断是否是最终答案
func isFinalAnswer(content string) bool {
	// 如果包含结构化的答案格式且内容较长，认为是最终答案
	hasStructure := strings.Contains(content, "###") || strings.Contains(content, "##") || strings.Contains(content, "# ")
	isLong := len([]rune(content)) > 200

	// 检查是否包含典型的报告结构
	hasReportStructure := containsAny(content, []string{
		"执行摘要", "核心发现", "研究背景", "结论",
		"executive summary", "key findings", "conclusion",
	})

	// 检查是否是简短的直接答案（不太可能是最终答案）
	hasThinkingMarkers := containsAny(content, []string{
		"让我", "我将", "需要", "首先", "然后", "正在",
		"思考", "分析一下", "检查",
	})

	return (hasStructure && isLong) || (hasReportStructure && isLong) ||
		(isLong && !hasThinkingMarkers && !strings.Contains(content, "调用"))
}

// containsAny 检查字符串是否包含任意一个关键词
func containsAny(content string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}
