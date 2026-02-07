package interfaces

import (
	"context"
	"link/internal/types"
)

// SessionRepository 会话数据访问接口
type SessionRepository interface {
	// Create 创建会话
	Create(ctx context.Context, session *types.SessionEntity) error

	// FindByID 根据ID查找会话
	FindByID(ctx context.Context, id int64) (*types.SessionEntity, error)

	// FindByUserID 根据用户ID查找会话列表
	FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*types.SessionEntity, int64, error)

	// FindByUserIDAndStatus 根据用户ID和状态查找会话
	FindByUserIDAndStatus(ctx context.Context, userID int64, status int8, page, pageSize int) ([]*types.SessionEntity, int64, error)

	// Update 更新会话
	Update(ctx context.Context, session *types.SessionEntity) error

	// UpdateMessageCount 更新会话消息数量
	UpdateMessageCount(ctx context.Context, sessionID int64) error

	// Delete 删除会话（软删除）
	Delete(ctx context.Context, id int64) error

	// HardDelete 硬删除会话
	HardDelete(ctx context.Context, id int64) error

	// CountByUserID 统计用户的会话数量
	CountByUserID(ctx context.Context, userID int64) (int64, error)

	// IncrementMessageCount 增加消息计数
	IncrementMessageCount(ctx context.Context, sessionID int64) error
}

// SessionService 会话服务接口
type SessionService interface {
	// CreateSession 创建会话
	CreateSession(ctx context.Context, userID int64, req *types.CreateSessionRequest) (*types.SessionResponse, error)

	// GetSessionByID 根据ID获取会话
	GetSessionByID(ctx context.Context, id int64) (*types.SessionResponse, error)

	// GetSessionDetail 获取会话详情（包含消息）
	GetSessionDetail(ctx context.Context, id int64) (*types.SessionDetailResponse, error)

	// ListSessions 查询会话列表
	ListSessions(ctx context.Context, req *types.ListSessionsRequest) (*types.SessionListResponse, error)

	// UpdateSession 更新会话
	UpdateSession(ctx context.Context, id int64, req *types.UpdateSessionRequest) (*types.SessionResponse, error)

	// DeleteSession 删除会话
	DeleteSession(ctx context.Context, id int64) error

	// ArchiveSession 归档会话
	ArchiveSession(ctx context.Context, id int64) error

	// ActivateSession 激活会话
	ActivateSession(ctx context.Context, id int64) error
}
