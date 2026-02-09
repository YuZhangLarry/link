package interfaces

import (
	"context"
	"link/internal/types"
)

// MessageRepository 消息数据访问接口 - 多租户版本
type MessageRepository interface {
	// Create 创建消息 - 返回创建的实体
	Create(ctx context.Context, req *types.CreateMessageRequest) (*types.MessageEntity, error)

	// FindByID 根据ID查找消息
	FindByID(ctx context.Context, id string) (*types.MessageEntity, error)

	// FindBySessionID 根据会话ID查找消息列表（原 FindByChatID）
	FindBySessionID(ctx context.Context, sessionID string, page, pageSize int) ([]*types.MessageEntity, int64, error)

	// FindBySessionIDAndRole 根据会话ID和角色查找消息（原 FindByChatIDAndRole）
	FindBySessionIDAndRole(ctx context.Context, sessionID string, role string, page, pageSize int) ([]*types.MessageEntity, int64, error)

	// Update 更新消息
	Update(ctx context.Context, id string, req *types.UpdateMessageRequest) error

	// Delete 删除消息（软删除）
	Delete(ctx context.Context, id string) error

	// DeleteBySessionID 删除会话的所有消息
	DeleteBySessionID(ctx context.Context, sessionID string) error

	// CountBySessionID 统计会话的消息数量
	CountBySessionID(ctx context.Context, sessionID string) (int64, error)
}

// MessageService 消息服务接口
type MessageService interface {
	// CreateMessage 创建消息
	CreateMessage(ctx context.Context, req *types.CreateMessageRequest) (*types.MessageResponse, error)

	// GetMessageByID 根据ID获取消息
	GetMessageByID(ctx context.Context, id string) (*types.MessageResponse, error)

	// ListMessages 查询消息列表
	ListMessages(ctx context.Context, req *types.ListMessagesRequest) (*types.MessageListResponse, error)

	// UpdateMessage 更新消息
	UpdateMessage(ctx context.Context, id string, req *types.UpdateMessageRequest) (*types.MessageResponse, error)

	// DeleteMessage 删除消息
	DeleteMessage(ctx context.Context, id string) error

	// DeleteMessagesBySessionID 删除会话的所有消息（原 DeleteMessagesByChatID）
	DeleteMessagesBySessionID(ctx context.Context, sessionID string) error
}
