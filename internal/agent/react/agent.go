// Package react 提供基于 Eino ADK 的 ReAct Agent 实现
package react

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"

	"link/internal/agent/tool"
)

// Agent ReAct Agent 实现
type Agent struct {
	config   *Config
	model    model.ToolCallingChatModel
	tools    []tool.BaseTool
	toolsMap map[string]tool.BaseTool
}

// NewAgent 创建新的 ReAct Agent
func NewAgent(chatModel model.ToolCallingChatModel, opts ...Option) (*Agent, error) {
	cfg := DefaultConfig()

	for _, opt := range opts {
		opt(cfg)
	}

	a := &Agent{
		config:   cfg,
		model:    chatModel,
		toolsMap: make(map[string]tool.BaseTool),
	}

	// 初始化默认工具
	if cfg.EnableTools {
		if err := a.initTools(); err != nil {
			return nil, fmt.Errorf("failed to init tools: %w", err)
		}
	}

	return a, nil
}

// Option 配置选项
type Option func(*Config)

// WithName 设置 Agent 名称
func WithName(name string) Option {
	return func(c *Config) {
		c.Name = name
	}
}

// WithDescription 设置 Agent 描述
func WithDescription(desc string) Option {
	return func(c *Config) {
		c.Description = desc
	}
}

// WithMaxIterations 设置最大迭代次数
func WithMaxIterations(n int) Option {
	return func(c *Config) {
		c.MaxIterations = n
	}
}

// WithTools 设置工具列表
func WithTools(tools []tool.BaseTool) Option {
	return func(c *Config) {
		// 由 initTools 处理
	}
}

// WithToolRegistry 使用工具注册表
func WithToolRegistry(registry *tool.Registry) Option {
	return func(c *Config) {
		// 由 initTools 处理
	}
}

// WithAllowedTools 设置允许的工具白名单
func WithAllowedTools(tools []string) Option {
	return func(c *Config) {
		c.AllowedTools = tools
	}
}

// WithSystemPrompt 设置系统提示词
func WithSystemPrompt(prompt string) Option {
	return func(c *Config) {
		c.SystemPrompt = prompt
	}
}

// initTools 初始化工具
func (a *Agent) initTools() error {
	registry, err := tool.InitDefaultTools()
	if err != nil {
		return err
	}

	// 过滤工具
	allTools := registry.GetTools()

	for _, t := range allTools {
		info, err := t.Info(context.Background())
		if err != nil {
			continue
		}

		// 检查白名单
		if len(a.config.AllowedTools) > 0 {
			allowed := false
			for _, name := range a.config.AllowedTools {
				if strings.EqualFold(info.Name, name) {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}

		// 检查黑名单
		blocked := false
		for _, name := range a.config.DeniedTools {
			if strings.EqualFold(info.Name, name) {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}

		a.tools = append(a.tools, t)
		a.toolsMap[info.Name] = t
	}

	return nil
}

// AddTool 添加工具
func (a *Agent) AddTool(t tool.BaseTool) error {
	info, err := t.Info(context.Background())
	if err != nil {
		return err
	}

	a.tools = append(a.tools, t)
	a.toolsMap[info.Name] = t
	return nil
}

// ========================================
// adk.Agent 接口实现
// ========================================

// Name 返回 Agent 名称
func (a *Agent) Name(ctx context.Context) string {
	if a.config.Name != "" {
		return a.config.Name
	}
	return "ReActAgent"
}

// Description 返回 Agent 描述
func (a *Agent) Description(ctx context.Context) string {
	if a.config.Description != "" {
		return a.config.Description
	}
	return "使用 ReAct 模式的智能助手，可以调用工具来解决问题"
}

// Run 运行 Agent
func (a *Agent) Run(ctx context.Context, input *adk.AgentInput, opts ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	// 构建 Eino ADK Agent
	aiAgent, err := a.buildADKAgent(ctx)
	if err != nil {
		gen := adk.NewAsyncGeneratorPair[*adk.AgentEvent]()
		gen.Send(&adk.AgentEvent{
			Err: fmt.Errorf("failed to build agent: %w", err),
		})
		gen.Close()
		return gen.NewIterator()
	}

	return aiAgent.Run(ctx, input, opts...)
}

// buildADKAgent 构建 Eino ADK Agent
func (a *Agent) buildADKAgent(ctx context.Context) (adk.Agent, error) {
	// 使用 Eino ADK 的 NewChatModelAgent 创建 ReAct Agent
	aiAgent, err := adk.NewChatModelAgent(ctx, a.model,
		adk.WithAgentConfig(*a.config),
		adk.WithTools(a.tools...),
		adk.WithAgentName(a.Name(ctx)),
		adk.WithSystemPrompt(a.buildSystemPrompt()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model agent: %w", err)
	}

	return aiAgent, nil
}

// buildSystemPrompt 构建系统提示词
func (a *Agent) buildSystemPrompt() string {
	var sb strings.Builder

	sb.WriteString(a.config.SystemPrompt)
	sb.WriteString("\n\n")

	// 添加工具信息
	if len(a.tools) > 0 {
		sb.WriteString("# 可用工具\n\n")
		for _, t := range a.tools {
			info, err := t.Info(context.Background())
			if err != nil {
				continue
			}
			sb.WriteString(fmt.Sprintf("## %s\n%s\n\n", info.Name, info.Desc))
		}
	}

	return sb.String()
}

// ========================================
// 高级接口
// ========================================

// Chat 对话接口（简化版）
func (a *Agent) Chat(ctx context.Context, query string, opts ...ChatOption) (*ChatResponse, error) {
	startTime := time.Now()
	cfg := &chatConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           a,
		EnableStreaming: false,
	})

	// 构建消息
	messages := []adk.Message{schema.UserMessage(query)}
	if cfg.sessionID != "" {
		// TODO: 从会话历史加载消息
	}

	// 运行
	iter := runner.Run(ctx, messages)

	// 收集结果
	response := &ChatResponse{
		SessionID: cfg.sessionID,
		Steps:     make([]*RunStep, 0),
		ToolCalls: make([]ToolCallInfo, 0),
	}

	var lastContent strings.Builder
	var lastRole string
	var stepNum int

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			return nil, event.Err
		}

		// 处理事件
		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				continue
			}

			// 记录步骤
			if msg.Role != lastRole {
				if lastContent.Len() > 0 {
					response.Steps = append(response.Steps, &RunStep{
						StepNumber: stepNum,
						Role:       lastRole,
						Content:    lastContent.String(),
						Timestamp:  time.Now(),
					})
					stepNum++
					lastContent.Reset()
				}
				lastRole = string(msg.Role)
			}

			lastContent.WriteString(msg.Content)

			// 记录工具调用
			for _, tc := range msg.ToolCalls {
				toolCall := ToolCallInfo{
					ID:   tc.ID,
					Name: tc.Function.Name,
				}
				if tc.Function.Arguments != "" {
					// 解析参数（简化处理）
					toolCall.Input = map[string]interface{}{
						"raw": tc.Function.Arguments,
					}
				}
				response.ToolCalls = append(response.ToolCalls, toolCall)
			}
		}

		// 处理工具结果事件
		if event.Output != nil && event.Output.MessageOutput != nil &&
			event.Output.MessageOutput.Role == schema.Tool {
			msg, _ := event.Output.MessageOutput.GetMessage()
			// 更新对应的工具调用结果
			for i := range response.ToolCalls {
				if response.ToolCalls[i].Output == "" {
					response.ToolCalls[i].Output = msg.Content
					response.ToolCalls[i].Success = true
					break
				}
			}
		}

		// 检查是否完成
		if event.Action != nil {
			if event.Action.Exit {
				break
			}
		}
	}

	// 添加最后一个步骤
	if lastContent.Len() > 0 {
		response.Steps = append(response.Steps, &RunStep{
			StepNumber: stepNum,
			Role:       lastRole,
			Content:    lastContent.String(),
			Timestamp:  time.Now(),
		})
	}

	// 提取最终答案（通常来自最后一条 Assistant 消息）
	for i := len(response.Steps) - 1; i >= 0; i-- {
		if response.Steps[i].Role == "assistant" && response.Steps[i].Content != "" {
			response.Answer = response.Steps[i].Content
			break
		}
	}

	response.DurationMs = time.Since(startTime).Milliseconds()
	response.Finished = true

	return response, nil
}

// ChatStream 流式对话接口
func (a *Agent) ChatStream(ctx context.Context, query string, opts ...ChatOption) (<-chan StreamEvent, error) {
	startTime := time.Now()
	cfg := &chatConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           a,
		EnableStreaming: true,
	})

	// 构建消息
	messages := []adk.Message{schema.UserMessage(query)}

	// 运行
	iter := runner.Run(ctx, messages)

	// 创建输出通道
	eventChan := make(chan StreamEvent, 16)

	go func() {
		defer close(eventChan)

		for {
			event, ok := iter.Next()
			if !ok {
				// 发送完成事件
				eventChan <- StreamEvent{
					Type:     EventTypeDone,
					Duration: time.Since(startTime),
				}
				return
			}

			if event.Err != nil {
				eventChan <- StreamEvent{
					Type:  EventTypeError,
					Error: event.Err.Error(),
				}
				return
			}

			// 处理消息事件
			if event.Output != nil && event.Output.MessageOutput != nil {
				output := event.Output.MessageOutput

				if output.MessageStream != nil {
					// 流式消息
					for {
						chunk, err := output.MessageStream.Recv()
						if err != nil {
							if err.Error() != "EOF" {
								eventChan <- StreamEvent{
									Type:  EventTypeError,
									Error: err.Error(),
								}
							}
							break
						}

						if chunk.Content != "" {
							eventChan <- StreamEvent{
								Type:    EventTypeContent,
								Content: chunk.Content,
								Role:    string(chunk.Role),
							}
						}

						if len(chunk.ToolCalls) > 0 {
							for _, tc := range chunk.ToolCalls {
								eventChan <- StreamEvent{
									Type: EventTypeToolCall,
									ToolCall: &ToolCallInfo{
										ID:   tc.ID,
										Name: tc.Function.Name,
										Input: map[string]interface{}{
											"raw": tc.Function.Arguments,
										},
									},
								}
							}
						}
					}
				} else if output.Message != nil {
					// 非流式消息
					msg := output.Message
					eventChan <- StreamEvent{
						Type:    EventTypeContent,
						Content: msg.Content,
						Role:    string(msg.Role),
					}

					if len(msg.ToolCalls) > 0 {
						for _, tc := range msg.ToolCalls {
							eventChan <- StreamEvent{
								Type: EventTypeToolCall,
								ToolCall: &ToolCallInfo{
									ID:   tc.ID,
									Name: tc.Function.Name,
									Input: map[string]interface{}{
										"raw": tc.Function.Arguments,
									},
								},
							}
						}
					}
				}
			}

			// 检查是否完成
			if event.Action != nil && event.Action.Exit {
				return
			}
		}
	}()

	return eventChan, nil
}

// ========================================
// 辅助类型
// ========================================

type chatConfig struct {
	sessionID string
}

type ChatOption func(*chatConfig)

func WithSessionID(id string) ChatOption {
	return func(c *chatConfig) {
		c.sessionID = id
	}
}

// StreamEvent 流式事件
type StreamEvent struct {
	Type     EventType      `json:"type"`
	Content  string         `json:"content,omitempty"`
	Role     string         `json:"role,omitempty"`
	ToolCall *ToolCallInfo  `json:"tool_call,omitempty"`
	Error    string         `json:"error,omitempty"`
	Duration time.Duration  `json:"duration,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// EventType 事件类型
type EventType string

const (
	EventTypeContent  EventType = "content"
	EventTypeToolCall EventType = "tool_call"
	EventTypeError    EventType = "error"
	EventTypeDone     EventType = "done"
)
