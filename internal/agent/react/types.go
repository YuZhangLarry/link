// Package react 提供基于 Eino ADK 的 ReAct Agent 实现
//
// ReAct (Reasoning + Acting) 是一种经典的 Agent 推理模式：
// 1. Thought（思考）: 分析当前情况
// 2. Action（行动）: 执行工具调用
// 3. Observation（观察）: 获取工具结果
// 4. 重复直到得到最终答案
package react

import (
	"time"
)

// ========================================
// 配置类型
// ========================================

// Config ReAct Agent 配置
type Config struct {
	// 基础配置
	Name        string `json:"name"`        // Agent 名称
	Description string `json:"description"` // Agent 描述

	// 推理配置
	MaxIterations int           `json:"max_iterations"` // 最大迭代次数
	Timeout       time.Duration `json:"timeout"`        // 单次调用超时

	// 工具配置
	EnableTools      bool            `json:"enable_tools"`       // 是否启用工具
	AllowedTools     []string        `json:"allowed_tools"`      // 允许的工具列表（白名单）
	DeniedTools      []string        `json:"denied_tools"`       // 禁止的工具列表（黑名单）
	ToolReturnDirect map[string]bool `json:"tool_return_direct"` // 直接返回结果的工具

	// 输出配置
	Verbose          bool `json:"verbose"`           // 是否输出详细推理过程
	IncludeReasoning bool `json:"include_reasoning"` // 是否在最终响应中包含推理过程

	// 系统提示词
	SystemPrompt string `json:"system_prompt,omitempty"` // 自定义系统提示词
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Name:             "react_agent",
		Description:      "使用 ReAct 模式的智能助手",
		MaxIterations:    10,
		Timeout:          30 * time.Second,
		EnableTools:      true,
		Verbose:          true,
		IncludeReasoning: false,
		SystemPrompt:     defaultSystemPrompt(),
	}
}

// ========================================
// 运行时类型
// ========================================

// RunResult Agent 运行结果
type RunResult struct {
	Answer     string      `json:"answer"`                // 最终答案
	Steps      []*RunStep  `json:"steps"`                 // 所有步骤
	TotalSteps int         `json:"total_steps"`           // 总步骤数
	Success    bool        `json:"success"`               // 是否成功
	Error      string      `json:"error,omitempty"`       // 错误信息
	DurationMs int64       `json:"duration_ms"`           // 总耗时（毫秒）
	TokenUsage *TokenUsage `json:"token_usage,omitempty"` // Token 使用情况
}

// RunStep 单次运行步骤
type RunStep struct {
	StepNumber int            `json:"step_number"`          // 步骤编号
	Role       string         `json:"role"`                 // 角色（user/assistant/tool）
	Content    string         `json:"content"`              // 内容
	ToolCalls  []ToolCallInfo `json:"tool_calls,omitempty"` // 工具调用
	Timestamp  time.Time      `json:"timestamp"`            // 时间戳
	DurationMs int            `json:"duration_ms"`          // 耗时
}

// ToolCallInfo 工具调用信息
type ToolCallInfo struct {
	ID       string                 `json:"id"`              // 调用 ID
	Name     string                 `json:"name"`            // 工具名称
	Input    map[string]interface{} `json:"input"`           // 输入参数
	Output   string                 `json:"output"`          // 输出结果
	Duration int                    `json:"duration"`        // 耗时（毫秒）
	Success  bool                   `json:"success"`         // 是否成功
	Error    string                 `json:"error,omitempty"` // 错误信息
}

// TokenUsage Token 使用情况
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ========================================
// 请求/响应类型
// ========================================

// ChatRequest 聊天请求
type ChatRequest struct {
	Query               string                 `json:"query"`                          // 用户问题
	SessionID           string                 `json:"session_id,omitempty"`           // 会话 ID
	ConversationHistory []ConversationMessage  `json:"conversation_history,omitempty"` // 对话历史
	Metadata            map[string]interface{} `json:"metadata,omitempty"`             // 元数据
	Streaming           bool                   `json:"streaming"`                      // 是否流式输出
}

// ConversationMessage 对话消息
type ConversationMessage struct {
	Role      string    `json:"role"`                // user / assistant / system
	Content   string    `json:"content"`             // 消息内容
	Timestamp time.Time `json:"timestamp,omitempty"` // 时间戳
}

// ChatResponse 聊天响应
type ChatResponse struct {
	Answer     string                 `json:"answer"`                // 最终答案
	Steps      []*RunStep             `json:"steps"`                 // 推理步骤
	Reasoning  string                 `json:"reasoning,omitempty"`   // 推理过程
	ToolCalls  []ToolCallInfo         `json:"tool_calls,omitempty"`  // 工具调用记录
	SessionID  string                 `json:"session_id"`            // 会话 ID
	Metadata   map[string]interface{} `json:"metadata,omitempty"`    // 元数据
	DurationMs int64                  `json:"duration_ms"`           // 耗时
	TokenUsage *TokenUsage            `json:"token_usage,omitempty"` // Token 使用
	Finished   bool                   `json:"finished"`              // 是否完成
}

// ========================================
// 默认提示词
// ========================================

const defaultSystemPrompt = `你是一个智能助手，可以使用工具来帮助用户解决问题。

你的工作流程：
1. 仔细理解用户的问题
2. 分析需要什么信息来回答
3. 选择合适的工具获取信息
4. 基于获取的信息给出准确答案

注意事项：
- 在使用工具前，先思考为什么需要这个工具以及期望得到什么结果
- 工具调用失败时，分析原因并尝试其他方法
- 如果信息不足，可以多次调用工具
- 最终答案要基于工具返回的事实，不要编造
- 对于不确定的信息，明确告诉用户而不是猜测
`
