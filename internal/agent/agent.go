package agent

import (
	"context"
	"fmt"
	tool2 "link/internal/agent/tool"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// Agent AI Agent，负责协调对话和工具调用
type Agent struct {
	registry  *tool2.Registry
	executor  *tool2.Executor
	config    *tool2.AgentConfig
	chatModel model.BaseChatModel // 使用 BaseChatModel 而不是已弃用的 ChatModel
}

// NewAgent 创建 Agent
func NewAgent(chatModel model.BaseChatModel, config *tool2.AgentConfig) (*Agent, error) {
	if config == nil {
		config = tool2.DefaultAgentConfig()
	}

	// 初始化工具注册表
	registry, err := tool2.InitDefaultTools()
	if err != nil {
		return nil, fmt.Errorf("failed to init tools: %w", err)
	}

	// 如果指定了工具列表，只使用这些工具
	if len(config.ToolNames) > 0 {
		registry, err = tool2.InitCustomTools(config.ToolNames)
		if err != nil {
			return nil, fmt.Errorf("failed to init custom tools: %w", err)
		}
	}

	return &Agent{
		registry:  registry,
		executor:  tool2.NewExecutor(registry),
		config:    config,
		chatModel: chatModel,
	}, nil
}

// NewAgentWithRegistry 使用自定义注册表创建 Agent
func NewAgentWithRegistry(chatModel model.BaseChatModel, config *tool2.AgentConfig, registry *tool2.Registry) (*Agent, error) {
	if config == nil {
		config = tool2.DefaultAgentConfig()
	}

	return &Agent{
		registry:  registry,
		executor:  tool2.NewExecutor(registry),
		config:    config,
		chatModel: chatModel,
	}, nil
}

// Chat 进行对话（支持工具调用）
func (a *Agent) Chat(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	if !a.config.EnableTools {
		// 不启用工具，直接调用模型
		return a.chatModel.Generate(ctx, messages, opts...)
	}

	// 启用工具，获取工具信息
	toolInfos := a.getToolInfos(ctx)
	if len(toolInfos) > 0 {
		opts = append(opts, model.WithTools(toolInfos))
	}

	// 迭代处理工具调用
	currentMessages := messages
	var iteration int

	for iteration < a.config.MaxToolIterations {
		iteration++

		// 调用模型
		resp, err := a.chatModel.Generate(ctx, currentMessages, opts...)
		if err != nil {
			return nil, fmt.Errorf("model generate failed: %w", err)
		}

		// 检查是否有工具调用
		if len(resp.ToolCalls) == 0 {
			// 没有工具调用，返回最终结果
			return resp, nil
		}

		// 添加助手消息（包含工具调用）到消息历史
		currentMessages = append(currentMessages, resp)

		// 将 ToolCalls 转换为指针切片
		toolCallPtrs := make([]*schema.ToolCall, len(resp.ToolCalls))
		for i := range resp.ToolCalls {
			toolCallPtrs[i] = &resp.ToolCalls[i]
		}

		// 执行所有工具调用
		toolResults := a.executeToolCalls(ctx, toolCallPtrs)

		// 将工具结果添加为新的消息
		for _, tr := range toolResults {
			toolMsg := &schema.Message{
				Role:    schema.Assistant, // 使用 Assistant
				Content: tool2.FormatToolResult(tr),
				// 设置额外信息以关联到原始调用
			}
			currentMessages = append(currentMessages, toolMsg)
		}
	}

	// 达到最大迭代次数，返回最后一次模型响应
	return currentMessages[len(currentMessages)-1], nil
}

// ChatStream 流式对话（支持工具调用）
func (a *Agent) ChatStream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	if !a.config.EnableTools {
		return a.chatModel.Stream(ctx, messages, opts...)
	}

	// 启用工具，获取工具信息
	toolInfos := a.getToolInfos(ctx)
	if len(toolInfos) > 0 {
		opts = append(opts, model.WithTools(toolInfos))
	}

	// 流式输出时，工具调用的处理更复杂
	// 这里简化处理：先收集完整响应，再处理工具调用

	// TODO: 实现完整的流式工具调用处理
	// 目前先返回非流式结果
	resp, err := a.chatModel.Generate(ctx, messages, opts...)
	if err != nil {
		return nil, err
	}

	// 创建简单的流（使用 Pipe）
	reader, writer := schema.Pipe[*schema.Message](1)
	go func() {
		defer writer.Close()
		writer.Send(resp, nil)
	}()

	return reader, nil
}

// executeToolCalls 执行工具调用
func (a *Agent) executeToolCalls(ctx context.Context, toolCalls []*schema.ToolCall) []*tool2.ToolExecResult {
	// 使用超时上下文
	ctx, cancel := context.WithTimeout(ctx, a.config.ToolTimeout)
	defer cancel()

	return a.executor.ExecuteAll(ctx, toolCalls)
}

// getToolInfos 获取工具信息列表（用于绑定到模型）
func (a *Agent) getToolInfos(ctx context.Context) []*schema.ToolInfo {
	tools := a.registry.GetTools()
	infos := make([]*schema.ToolInfo, 0, len(tools))

	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}

	return infos
}

// GetToolRegistry 获取工具注册表
func (a *Agent) GetToolRegistry() *tool2.Registry {
	return a.registry
}

// GetExecutor 获取工具执行器
func (a *Agent) GetExecutor() *tool2.Executor {
	return a.executor
}

// UpdateConfig 更新配置
func (a *Agent) UpdateConfig(config *tool2.AgentConfig) {
	if config != nil {
		a.config = config
	}
}

// RegisterTool 注册新工具
func (a *Agent) RegisterTool(name string, t tool.BaseTool) error {
	return a.registry.Register(name, t)
}

// UnregisterTool 注销工具
func (a *Agent) UnregisterTool(name string) {
	a.registry.Unregister(name)
}

// ListTools 列出所有可用工具
func (a *Agent) ListTools() []string {
	return a.registry.List()
}

// GetToolsInfo 获取所有工具信息
func (a *Agent) GetToolsInfo(ctx context.Context) ([]map[string]interface{}, error) {
	return a.registry.GetAllToolsInfo(ctx)
}

// ========================================
// Agent 辅助方法
// ========================================

// BuildSystemPrompt 构建系统提示词
func BuildSystemPrompt(tools []tool.BaseTool) string {
	var sb strings.Builder

	sb.WriteString("你是一个智能助手，可以使用以下工具来帮助用户：\n\n")

	for _, t := range tools {
		info, err := t.Info(context.Background())
		if err != nil {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s\n", info.Name))
		sb.WriteString(fmt.Sprintf("%s\n\n", info.Desc))
	}

	sb.WriteString("\n使用工具时的注意事项：\n")
	sb.WriteString("1. 仔细理解用户需求，选择最合适的工具\n")
	sb.WriteString("2. 确保工具参数正确无误\n")
	sb.WriteString("3. 根据工具返回结果，给出有用的回答\n")
	sb.WriteString("4. 如果工具调用失败，向用户说明原因并提供替代方案\n")

	return sb.String()
}

// ShouldUseTool 判断是否需要使用工具
func ShouldUseTool(content string, tools []tool.BaseTool) bool {
	// 简单判断：如果内容包含某些关键词，可能需要使用工具
	keywords := []string{
		"搜索", "查询", "查找", "获取", "计算",
		"时间", "日期", "天气", "新闻",
		"知识库", "文档", "资料",
	}

	contentLower := strings.ToLower(content)
	for _, kw := range keywords {
		if strings.Contains(contentLower, kw) {
			return true
		}
	}

	return false
}
