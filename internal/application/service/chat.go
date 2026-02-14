package service

import (
	"context"
	"fmt"
	"link/internal/agent"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"link/internal/config"
	"link/internal/models/chat"
	"link/internal/types"
)

// ChatService 聊天服务
type ChatService struct {
	chatConfig *config.ChatConfig
	agent      *agent.Agent
	enableTool bool
}

// NewChatService 创建聊天服务
func NewChatService(chatConfig *config.ChatConfig) *ChatService {
	return &ChatService{
		chatConfig: chatConfig,
		enableTool: false, // 默认不启用工具
	}
}

// NewChatServiceWithAgent 创建带 Agent 的聊天服务
func NewChatServiceWithAgent(chatConfig *config.ChatConfig, agent *agent.Agent) *ChatService {
	return &ChatService{
		chatConfig: chatConfig,
		agent:      agent,
		enableTool: true,
	}
}

// SetAgent 设置 Agent
func (s *ChatService) SetAgent(agent *agent.Agent) {
	s.agent = agent
	s.enableTool = true
}

// EnableTool 启用工具
func (s *ChatService) EnableTool(enable bool) {
	s.enableTool = enable
}

// Chat 非流式聊天
func (s *ChatService) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	// 如果启用工具且 Agent 可用，使用 Agent
	if s.enableTool && s.agent != nil {
		return s.chatWithAgent(ctx, req)
	}

	// 否则使用普通聊天
	return s.chatNormal(ctx, req)
}

// chatWithAgent 使用 Agent 进行聊天（支持工具调用）
func (s *ChatService) chatWithAgent(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	// 转换消息格式为 Eino 格式
	messages := s.convertToEinoMessages(req.History, req.Content)

	// 转换选项
	opts := s.convertToEinoOptions(req.Options)

	// 使用 Agent 进行对话
	resp, err := s.agent.Chat(ctx, messages, opts...)
	if err != nil {
		return nil, fmt.Errorf("agent chat failed: %w", err)
	}

	// 转换响应
	toolCallPtrs := make([]*schema.ToolCall, len(resp.ToolCalls))
	for i := range resp.ToolCalls {
		toolCallPtrs[i] = &resp.ToolCalls[i]
	}

	return &types.ChatResponse{
		MessageID:    generateMessageID(),
		Content:      resp.Content,
		Role:         string(resp.Role),
		TokenCount:   0, // 需要从 resp 中获取
		ToolCalls:    convertEinoToolCalls(toolCallPtrs),
		FinishReason: "stop",
	}, nil
}

// chatNormal 普通聊天（不使用工具）
func (s *ChatService) chatNormal(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	// 创建聊天实例
	chatInstance, err := s.createChatInstance()
	if err != nil {
		return nil, fmt.Errorf("failed to create chat instance: %w", err)
	}

	// 转换消息格式
	messages := s.convertMessages(req.History, req.Content)

	// 转换选项
	opts := s.convertOptions(req.Options)

	// 调用聊天
	resp, err := chatInstance.Chat(ctx, messages, opts)
	if err != nil {
		return nil, fmt.Errorf("chat failed: %w", err)
	}

	return &types.ChatResponse{
		MessageID:    resp.MessageID,
		Content:      resp.Content,
		Role:         resp.Role,
		TokenCount:   resp.TokenCount,
		ToolCalls:    s.convertToolCalls(resp.ToolCalls),
		FinishReason: resp.FinishReason,
	}, nil
}

// ChatStream 流式聊天
func (s *ChatService) ChatStream(ctx context.Context, req *types.ChatRequest) (<-chan types.StreamChatEvent, error) {
	// 创建聊天实例
	chatInstance, err := s.createChatInstance()
	if err != nil {
		return nil, fmt.Errorf("failed to create chat instance: %w", err)
	}

	// 转换消息格式
	messages := s.convertMessages(req.History, req.Content)

	// 转换选项
	opts := s.convertOptions(req.Options)

	// 调用流式聊天
	respChan, err := chatInstance.ChatStream(ctx, messages, opts)
	if err != nil {
		return nil, fmt.Errorf("stream chat failed: %w", err)
	}

	// 转换响应格式
	eventChan := make(chan types.StreamChatEvent, 10)
	go func() {
		defer close(eventChan)
		for resp := range respChan {
			event := types.StreamChatEvent{
				Event:      resp.Event,
				Content:    resp.Content,
				MessageID:  resp.MessageID,
				TokenCount: resp.TokenCount,
				ToolCalls:  s.convertToolCalls(resp.ToolCalls),
				Error:      resp.Error,
			}
			eventChan <- event
		}
	}()

	return eventChan, nil
}

// ========================================
// Private Methods
// ========================================

// createChatInstance 创建聊天实例
func (s *ChatService) createChatInstance() (chat.Chat, error) {
	config := &chat.ChatConfig{
		Source:    s.chatConfig.Source,
		BaseURL:   s.chatConfig.BaseURL,
		ModelName: s.chatConfig.ModelName,
		APIKey:    s.chatConfig.APIKey,
		Provider:  s.chatConfig.Provider,
		ModelID:   fmt.Sprintf("chat_%d", time.Now().UnixNano()),
	}

	return chat.NewChat(config)
}

// convertMessages 转换消息格式
func (s *ChatService) convertMessages(history []types.Message, currentContent string) []chat.Message {
	messages := make([]chat.Message, 0, len(history)+1)

	// 添加历史消息
	for _, msg := range history {
		messages = append(messages, chat.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 添加当前用户消息
	messages = append(messages, chat.Message{
		Role:    "user",
		Content: currentContent,
	})

	return messages
}

// convertOptions 转换选项
func (s *ChatService) convertOptions(opts *types.ChatOptions) *chat.ChatOptions {
	if opts == nil {
		return nil
	}

	return &chat.ChatOptions{
		Temperature:      opts.Temperature,
		TopP:             opts.TopP,
		MaxTokens:        opts.MaxTokens,
		FrequencyPenalty: opts.FrequencyPenalty,
		PresencePenalty:  opts.PresencePenalty,
		Thinking:         &opts.Thinking,
	}
}

// convertToolCalls 转换工具调用
func (s *ChatService) convertToolCalls(calls []chat.ToolCall) []types.ToolCall {
	if calls == nil {
		return nil
	}

	result := make([]types.ToolCall, len(calls))
	for i, call := range calls {
		result[i] = types.ToolCall{
			ID:   call.ID,
			Type: call.Type,
			Function: types.FunctionCall{
				Name:      call.Function.Name,
				Arguments: call.Function.Arguments,
			},
		}
	}
	return result
}

// ========================================
// Eino 格式转换方法
// ========================================

// convertToEinoMessages 转换为 Eino 消息格式
func (s *ChatService) convertToEinoMessages(history []types.Message, currentContent string) []*schema.Message {
	messages := make([]*schema.Message, 0, len(history)+1)

	// 添加历史消息
	for _, msg := range history {
		einoMsg := &schema.Message{
			Role:    schema.RoleType(msg.Role),
			Content: msg.Content,
		}
		messages = append(messages, einoMsg)
	}

	// 添加当前用户消息
	messages = append(messages, &schema.Message{
		Role:    schema.User,
		Content: currentContent,
	})

	return messages
}

// convertToEinoOptions 转换为 Eino 选项
func (s *ChatService) convertToEinoOptions(opts *types.ChatOptions) []model.Option {
	if opts == nil {
		return nil
	}

	var modelOpts []model.Option

	if opts.Temperature > 0 {
		modelOpts = append(modelOpts, model.WithTemperature(float32(opts.Temperature)))
	}
	if opts.TopP > 0 {
		modelOpts = append(modelOpts, model.WithTopP(float32(opts.TopP)))
	}
	if opts.MaxTokens > 0 {
		modelOpts = append(modelOpts, model.WithMaxTokens(opts.MaxTokens))
	}

	return modelOpts
}

// convertEinoToolCalls 转换 Eino 工具调用
func convertEinoToolCalls(calls []*schema.ToolCall) []types.ToolCall {
	if calls == nil {
		return nil
	}

	result := make([]types.ToolCall, len(calls))
	for i, call := range calls {
		result[i] = types.ToolCall{
			ID:   call.ID,
			Type: string(call.Type),
			Function: types.FunctionCall{
				Name:      call.Function.Name,
				Arguments: call.Function.Arguments,
			},
		}
	}
	return result
}

// generateMessageID 生成消息ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
