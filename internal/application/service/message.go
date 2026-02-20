package service

import (
	"context"
	"errors"
	"fmt"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// MessageService 消息服务实现 - 多租户版本
type MessageService struct {
	messageRepo interfaces.MessageRepository
}

// NewMessageService 创建消息服务实例
func NewMessageService(messageRepo interfaces.MessageRepository) interfaces.MessageService {
	return &MessageService{
		messageRepo: messageRepo,
	}
}

// CreateMessage 创建消息
func (s *MessageService) CreateMessage(ctx context.Context, req *types.CreateMessageRequest) (*types.MessageResponse, error) {
	// 验证角色
	if req.Role != "system" && req.Role != "user" && req.Role != "assistant" && req.Role != "tool" {
		return nil, errors.New("无效的消息角色")
	}

	// 验证内容
	if req.Content == "" {
		return nil, errors.New("消息内容不能为空")
	}

	// 调用仓储创建消息
	message, err := s.messageRepo.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("创建消息失败: %w", err)
	}

	// 返回响应
	return s.toMessageResponse(message), nil
}

// GetMessageByID 根据ID获取消息
func (s *MessageService) GetMessageByID(ctx context.Context, id string) (*types.MessageResponse, error) {
	message, err := s.messageRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toMessageResponse(message), nil
}

// ListMessages 查询消息列表
func (s *MessageService) ListMessages(ctx context.Context, req *types.ListMessagesRequest) (*types.MessageListResponse, error) {
	// 设置默认分页参数
	page := req.Page
	if page == 0 {
		page = 1
	}
	size := req.Size
	if size == 0 {
		size = 20
	}

	var messages []*types.MessageEntity
	var total int64
	var err error

	// 根据是否按角色筛选选择不同的查询方法
	if req.Role != "" {
		messages, total, err = s.messageRepo.FindBySessionIDAndRole(ctx, req.SessionID, req.Role, page, size)
	} else {
		messages, total, err = s.messageRepo.FindBySessionID(ctx, req.SessionID, page, size)
	}

	if err != nil {
		return nil, fmt.Errorf("查询消息列表失败: %w", err)
	}

	// 转换为响应格式
	messageResponses := make([]*types.MessageResponse, 0, len(messages))
	for _, msg := range messages {
		messageResponses = append(messageResponses, s.toMessageResponse(msg))
	}

	return &types.MessageListResponse{
		Messages: messageResponses,
		Total:    total,
		Page:     page,
		Size:     size,
	}, nil
}

// UpdateMessage 更新消息
func (s *MessageService) UpdateMessage(ctx context.Context, id string, req *types.UpdateMessageRequest) (*types.MessageResponse, error) {
	// 查找消息
	message, err := s.messageRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新
	err = s.messageRepo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("更新消息失败: %w", err)
	}

	// 重新获取更新后的消息
	message, err = s.messageRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toMessageResponse(message), nil
}

// DeleteMessage 删除消息
func (s *MessageService) DeleteMessage(ctx context.Context, id string) error {
	err := s.messageRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("删除消息失败: %w", err)
	}
	return nil
}

// DeleteMessagesBySessionID 删除会话的所有消息
func (s *MessageService) DeleteMessagesBySessionID(ctx context.Context, sessionID string) error {
	err := s.messageRepo.DeleteBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("删除会话消息失败: %w", err)
	}
	return nil
}

// toMessageResponse 转换为消息响应格式
func (s *MessageService) toMessageResponse(message *types.MessageEntity) *types.MessageResponse {
	return &types.MessageResponse{
		ID:                  message.ID,
		RequestID:           message.RequestID,
		SessionID:           message.SessionID,
		Role:                message.Role,
		Content:             message.Content,
		KnowledgeReferences: message.KnowledgeReferences,
		AgentSteps:          message.AgentSteps,
		ToolCalls:           message.ToolCalls,
		IsCompleted:         message.IsCompleted,
		TokenCount:          message.TokenCount,
		CreatedAt:           message.CreatedAt,
	}
}
