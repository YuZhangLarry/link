package service

import (
	"context"
	"fmt"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// SessionService 会话服务实现 - 多租户版本
type SessionService struct {
	sessionRepo interfaces.SessionRepository
}

// NewSessionService 创建会话服务实例
func NewSessionService(sessionRepo interfaces.SessionRepository) interfaces.SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
	}
}

// CreateSession 创建会话
func (s *SessionService) CreateSession(ctx context.Context, userID int64, req *types.CreateSessionRequest) (*types.SessionResponse, error) {
	// 调用仓储创建会话（repository会处理默认值）
	session, err := s.sessionRepo.Create(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	return s.toSessionResponse(session), nil
}

// GetSessionByID 根据ID获取会话
func (s *SessionService) GetSessionByID(ctx context.Context, id string) (*types.SessionResponse, error) {
	// 从上下文获取用户ID，进行权限验证
	userID, ok := ctx.Value("user_id").(int64)
	if ok && userID > 0 {
		// 验证会话是否属于该用户
		session, err := s.sessionRepo.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		// 检查会话是否属于当前用户
		if session.UserID != userID {
			return nil, fmt.Errorf("无权访问该会话")
		}
		return s.toSessionResponse(session), nil
	}

	// 如果没有 user_id，则直接查询（兼容旧逻辑）
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toSessionResponse(session), nil
}

// GetSessionDetail 获取会话详情（包含消息）
func (s *SessionService) GetSessionDetail(ctx context.Context, id string) (*types.SessionDetailResponse, error) {
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &types.SessionDetailResponse{
		SessionResponse: *s.toSessionResponse(session),
		// 消息列表需要通过 MessageService 获取
		Messages: []*types.MessageResponse{},
	}, nil
}

// ListSessions 查询会话列表
func (s *SessionService) ListSessions(ctx context.Context, req *types.ListSessionsRequest) (*types.SessionListResponse, error) {
	// 设置默认分页参数
	page := req.Page
	if page == 0 {
		page = 1
	}
	size := req.Size
	if size == 0 {
		size = 20
	}

	// 从上下文获取用户ID
	userID, ok := ctx.Value("user_id").(int64)
	if !ok || userID == 0 {
		return nil, fmt.Errorf("未找到用户ID")
	}

	var sessions []*types.SessionEntity
	var total int64
	var err error

	// 根据是否筛选状态选择不同的查询方法
	if req.Status != nil {
		sessions, total, err = s.sessionRepo.FindByUserIDAndStatus(ctx, userID, *req.Status, page, size)
	} else {
		sessions, total, err = s.sessionRepo.FindByUserID(ctx, userID, page, size)
	}

	if err != nil {
		return nil, fmt.Errorf("查询会话列表失败: %w", err)
	}

	// 转换为响应格式
	sessionResponses := make([]*types.SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		sessionResponses = append(sessionResponses, s.toSessionResponse(session))
	}

	return &types.SessionListResponse{
		Sessions: sessionResponses,
		Total:    total,
		Page:     page,
		Size:     size,
	}, nil
}

// UpdateSession 更新会话
func (s *SessionService) UpdateSession(ctx context.Context, id string, req *types.UpdateSessionRequest) (*types.SessionResponse, error) {
	// 更新
	err := s.sessionRepo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("更新会话失败: %w", err)
	}

	// 重新获取更新后的会话
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toSessionResponse(session), nil
}

// DeleteSession 删除会话
func (s *SessionService) DeleteSession(ctx context.Context, id string) error {
	err := s.sessionRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("删除会话失败: %w", err)
	}
	return nil
}

// ArchiveSession 归档会话
func (s *SessionService) ArchiveSession(ctx context.Context, id string) error {
	// 归档状态为 0
	status := int8(0)
	req := &types.UpdateSessionRequest{
		Status: &status,
	}

	err := s.sessionRepo.Update(ctx, id, req)
	if err != nil {
		return fmt.Errorf("归档会话失败: %w", err)
	}

	return nil
}

// ActivateSession 激活会话
func (s *SessionService) ActivateSession(ctx context.Context, id string) error {
	// 正常状态为 1
	status := int8(1)
	req := &types.UpdateSessionRequest{
		Status: &status,
	}

	err := s.sessionRepo.Update(ctx, id, req)
	if err != nil {
		return fmt.Errorf("激活会话失败: %w", err)
	}

	return nil
}

// toSessionResponse 转换为会话响应格式
func (s *SessionService) toSessionResponse(session *types.SessionEntity) *types.SessionResponse {
	return &types.SessionResponse{
		ID:              session.ID,
		TenantID:        session.TenantID,
		UserID:          session.UserID,
		Title:           session.Title,
		Description:     session.Description,
		KBID:            session.KBID,
		MaxRounds:       session.MaxRounds,
		EnableRewrite:   session.EnableRewrite,
		MessageCount:    session.MessageCount,
		Status:          session.Status,
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
	}
}
