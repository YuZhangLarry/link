// Package agent 提供 Agent 基础接口和类型定义
package agent

import (
	"context"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// Agent 接口定义 Agent 的基础能力
type Agent interface {
	// Name 返回 Agent 名称
	Name(ctx context.Context) string

	// Description 返回 Agent 描述
	Description(ctx context.Context) string

	// Chat 处理单轮对话
	Chat(ctx context.Context, query string, opts ...Option) (*Response, error)

	// StreamChat 处理流式对话
	StreamChat(ctx context.Context, query string, opts ...Option) (*schema.StreamReader[*ChatChunk], error)
}

// Option Agent 配置选项
type Option func(*Options)

// Options Agent 运行时配置
type Options struct {
	// Tools 限制使用的工具
	Tools []string

	// MaxIterations 最大迭代次数
	MaxIterations int

	// SessionID 会话ID
	SessionID string

	// UserID 用户ID
	UserID string

	// TenantID 租户ID
	TenantID int64

	// Extra 额外参数
	Extra map[string]interface{}
}

// Response Agent 响应
type Response struct {
	Answer    string                 `json:"answer"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	ToolCalls []*ToolCallRecord      `json:"tool_calls,omitempty"`
	Sources   []string               `json:"sources,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ChatChunk 流式聊天片段
type ChatChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

// ToolCallRecord 工具调用记录
type ToolCallRecord struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Input    string                 `json:"input"`
	Output   string                 `json:"output,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BaseAgent 基础 Agent 实现
type BaseAgent struct {
	name        string
	description string
	chatModel   model.ToolCallingChatModel
	tools       []tool.BaseTool
}

// NewBaseAgent 创建基础 Agent
func NewBaseAgent(name, description string, chatModel model.ToolCallingChatModel, tools []tool.BaseTool) *BaseAgent {
	return &BaseAgent{
		name:        name,
		description: description,
		chatModel:   chatModel,
		tools:       tools,
	}
}

// Name 实现 Agent 接口
func (a *BaseAgent) Name(ctx context.Context) string {
	return a.name
}

// Description 实现 Agent 接口
func (a *BaseAgent) Description(ctx context.Context) string {
	return a.description
}

// GetTools 获取工具列表
func (a *BaseAgent) GetTools() []tool.BaseTool {
	return a.tools
}

// GetChatModel 获取聊天模型
func (a *BaseAgent) GetChatModel() model.ToolCallingChatModel {
	return a.chatModel
}

// ========================================
// 辅助函数
// ========================================

// ShouldUseTool 判断是否需要使用工具（保留用于兼容）
func ShouldUseTool(content string, tools []tool.BaseTool) bool {
	// 这个函数现在只是示例，实际判断由大模型完成
	return len(tools) > 0
}
