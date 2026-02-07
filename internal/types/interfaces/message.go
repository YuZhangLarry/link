package interfaces

import (
	"context"
	"link/internal/types"
)

// MessageRepository 消息数据访问接口
type MessageRepository interface {
	// Create 创建消息
	Create(ctx context.Context, message *types.MessageEntity) error

	// FindByID 根据ID查找消息
	FindByID(ctx context.Context, id int64) (*types.MessageEntity, error)

	// FindByChatID 根据对话ID查找消息列表
	FindByChatID(ctx context.Context, chatID int64, page, pageSize int) ([]*types.MessageEntity, int64, error)

	// FindByChatIDAndRole 根据对话ID和角色查找消息
	FindByChatIDAndRole(ctx context.Context, chatID int64, role string, page, pageSize int) ([]*types.MessageEntity, int64, error)

	// Update 更新消息
	Update(ctx context.Context, message *types.MessageEntity) error

	// Delete 删除消息
	Delete(ctx context.Context, id int64) error

	// DeleteByChatID 删除对话的所有消息
	DeleteByChatID(ctx context.Context, chatID int64) error

	// CountByChatID 统计对话的消息数量
	CountByChatID(ctx context.Context, chatID int64) (int64, error)
}

// MessageService 消息服务接口
type MessageService interface {
	// CreateMessage 创建消息
	CreateMessage(ctx context.Context, req *types.CreateMessageRequest) (*types.MessageResponse, error)

	// GetMessageByID 根据ID获取消息
	GetMessageByID(ctx context.Context, id int64) (*types.MessageResponse, error)

	// ListMessages 查询消息列表
	ListMessages(ctx context.Context, req *types.ListMessagesRequest) (*types.MessageListResponse, error)

	// UpdateMessage 更新消息
	UpdateMessage(ctx context.Context, id int64, req *types.UpdateMessageRequest) (*types.MessageResponse, error)

	// DeleteMessage 删除消息
	DeleteMessage(ctx context.Context, id int64) error

	// DeleteMessagesByChatID 删除对话的所有消息
	DeleteMessagesByChatID(ctx context.Context, chatID int64) error
}
