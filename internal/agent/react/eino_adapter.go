// Package react 提供与 Eino ADK 的适配器
package react

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// ========================================
// Eino ADK 适配器选项
// ========================================

// agentConfig 实现 adk.AgentConfig 接口的适配
type agentConfig struct {
	*Config
}

// toADKConfig 转换为 ADK 配置
func toADKConfig(cfg *Config) *adk.ChatModelAgentConfig {
	return &adk.ChatModelAgentConfig{
		MaxIterations: cfg.MaxIterations,
		SystemPrompt:  cfg.SystemPrompt,
	}
}

// ========================================
// Agent 构建选项
// ========================================

// AgentOption 构建选项
type AgentOption func(*agentBuilder)

// agentBuilder Agent 构建器
type agentBuilder struct {
	config      *Config
	model       model.ToolCallingChatModel
	tools       []tool.BaseTool
	toolOpts    []compose.ToolsNodeOption
	middlewares []adk.AgentMiddleware
}

// newAgentBuilder 创建构建器
func newAgentBuilder(model model.ToolCallingChatModel) *agentBuilder {
	return &agentBuilder{
		config: DefaultConfig(),
		model:  model,
	}
}

// WithConfig 设置配置
func WithConfig(cfg *Config) AgentOption {
	return func(b *agentBuilder) {
		b.config = cfg
	}
}

// WithAgentTools 设置工具
func WithAgentTools(tools ...tool.BaseTool) AgentOption {
	return func(b *agentBuilder) {
		b.tools = append(b.tools, tools...)
	}
}

// WithToolMiddlewares 设置工具中间件
func WithToolMiddlewares(middlewares ...compose.ToolMiddleware) AgentOption {
	return func(b *agentBuilder) {
		b.toolOpts = append(b.toolOpts, compose.WithToolCallMiddlewares(middlewares...))
	}
}

// ========================================
// NewChatModelAgent 创建 ChatModel Agent（使用 Eino ADK）
//
// 这是基于 Eino ADK 的简化封装，提供更友好的 API
func NewChatModelAgent(ctx context.Context, chatModel model.ToolCallingChatModel, opts ...AgentOption) (adk.Agent, error) {
	builder := newAgentBuilder(chatModel)

	for _, opt := range opts {
		opt(builder)
	}

	// 准备工具
	var tools []tool.BaseTool
	if len(builder.tools) > 0 {
		tools = builder.tools
	} else if builder.config.EnableTools {
		// 使用默认工具
		registry, err := InitToolRegistry(builder.config)
		if err != nil {
			return nil, fmt.Errorf("failed to init tool registry: %w", err)
		}
		tools = registry.GetTools()
	}

	// 构建 ADK 配置
	adkOpts := []adk.ChatModelAgentOption{
		adk.WithAgentName(builder.config.Name),
		adk.WithSystemPrompt(builder.config.SystemPrompt),
		adk.WithMaxIterations(builder.config.MaxIterations),
		adk.WithTools(tools...),
	}

	// 添加工具直接返回配置
	if len(builder.config.ToolReturnDirect) > 0 {
		adkOpts = append(adkOpts, adk.WithToolReturnDirect(builder.config.ToolReturnDirect))
	}

	// 使用 Eino ADK 创建 Agent
	agent, err := adk.NewChatModelAgent(ctx, chatModel, adkOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model agent: %w", err)
	}

	return agent, nil
}

// InitToolRegistry 初始化工具注册表
func InitToolRegistry(cfg *Config) (*tool.Registry, error) {
	registry, err := tool.InitDefaultTools()
	if err != nil {
		return nil, err
	}

	// 如果有白名单，过滤工具
	if len(cfg.AllowedTools) > 0 {
		filtered := tool.NewRegistry()
		for _, name := range cfg.AllowedTools {
			if t, ok := registry.Get(name); ok {
				filtered.Register(name, t)
			}
		}
		return filtered, nil
	}

	return registry, nil
}

// ========================================
// Runner 封装
// ========================================

// NewAgentRunner 创建 Agent Runner（简化版）
func NewAgentRunner(agent adk.Agent, streaming bool) *adk.Runner {
	return adk.NewRunner(context.Background(), adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: streaming,
	})
}

// RunAgent 运行 Agent（最简 API）
func RunAgent(ctx context.Context, chatModel model.ToolCallingChatModel, query string, opts ...AgentOption) (*RunResult, error) {
	agent, err := NewChatModelAgent(ctx, chatModel, opts...)
	if err != nil {
		return nil, err
	}

	runner := NewAgentRunner(agent, false)
	iter := runner.Query(ctx, query)

	result := &RunResult{
		Steps: make([]*RunStep, 0),
	}

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			result.Error = event.Err.Error()
			result.Success = false
			return result, nil
		}

		// 处理事件
		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				continue
			}

			result.Steps = append(result.Steps, &RunStep{
				StepNumber: len(result.Steps) + 1,
				Role:       string(msg.Role),
				Content:    msg.Content,
				Timestamp:  event.Output.MessageOutput.Message.Time,
			})

			// 提取工具调用
			for _, tc := range msg.ToolCalls {
				result.Steps = append(result.Steps, &RunStep{
					StepNumber: len(result.Steps) + 1,
					Role:       "tool_call",
					Content:    fmt.Sprintf("Call: %s", tc.Function.Name),
					ToolCalls: []ToolCallInfo{{
						ID:   tc.ID,
						Name: tc.Function.Name,
						Input: map[string]interface{}{
							"raw": tc.Function.Arguments,
						},
					}},
				})
			}
		}

		// 最终答案
		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, _ := event.Output.MessageOutput.GetMessage()
			if msg.Role == schema.Assistant && len(msg.ToolCalls) == 0 {
				result.Answer = msg.Content
			}
		}

		// 检查完成
		if event.Action != nil && event.Action.Exit {
			break
		}
	}

	result.TotalSteps = len(result.Steps)
	result.Success = result.Error == ""

	return result, nil
}
