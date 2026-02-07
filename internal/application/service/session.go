package service

import (
	"context"
	"fmt"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// SessionService 会话服务实现
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
	// 设置默认值
	if req.Model == "" {
		req.Model = "gpt-3.5-turbo"
	}
	if req.MaxRounds == 0 {
		req.MaxRounds = 10
	}
	if req.KeywordThreshold == 0 {
		req.KeywordThreshold = 0.5
	}
	if req.VectorThreshold == 0 {
		req.VectorThreshold = 0.5
	}

	// 创建会话实体
	session := &types.SessionEntity{
		UserID:           userID,
		KBID:             req.KBID,
		Title:            req.Title,
		Description:      req.Description,
		Model:            req.Model,
		MaxRounds:        req.MaxRounds,
		EnableRewrite:    req.EnableRewrite,
		KeywordThreshold: req.KeywordThreshold,
		VectorThreshold:  req.VectorThreshold,
		Status:           1, // 默认正常状态
	}

	// 保存到数据库
	err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	return s.toSessionResponse(session), nil
}

// GetSessionByID 根据ID获取会话
func (s *SessionService) GetSessionByID(ctx context.Context, id int64) (*types.SessionResponse, error) {
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.toSessionResponse(session), nil
}

// GetSessionDetail 获取会话详情（包含消息）
func (s *SessionService) GetSessionDetail(ctx context.Context, id int64) (*types.SessionDetailResponse, error) {
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &types.SessionDetailResponse{
		SessionResponse: *s.toSessionResponse(session),
		// 消息列表需要通过 MessageService 获取
		Messages: []types.MessageResponse{},
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

	var sessions []*types.SessionEntity
	var total int64
	var err error

	// 根据是否筛选状态选择不同的查询方法
	if req.Status != nil {
		sessions, total, err = s.sessionRepo.FindByUserIDAndStatus(ctx, req.UserID, *req.Status, page, size)
	} else {
		sessions, total, err = s.sessionRepo.FindByUserID(ctx, req.UserID, page, size)
	}

	if err != nil {
		return nil, fmt.Errorf("查询会话列表失败: %w", err)
	}

	// 转换为响应格式
	sessionResponses := make([]types.SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		sessionResponses = append(sessionResponses, *s.toSessionResponse(session))
	}

	return &types.SessionListResponse{
		Sessions: sessionResponses,
		Total:    total,
		Page:     page,
		Size:     size,
	}, nil
}

// UpdateSession 更新会话
func (s *SessionService) UpdateSession(ctx context.Context, id int64, req *types.UpdateSessionRequest) (*types.SessionResponse, error) {
	// 查找会话
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	session.Title = req.Title
	session.Description = req.Description
	session.Model = req.Model
	session.MaxRounds = req.MaxRounds
	session.EnableRewrite = req.EnableRewrite
	session.KeywordThreshold = req.KeywordThreshold
	session.VectorThreshold = req.VectorThreshold
	if req.Status != nil {
		session.Status = *req.Status
	}

	// 保存更新
	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("更新会话失败: %w", err)
	}

	return s.toSessionResponse(session), nil
}

// DeleteSession 删除会话
func (s *SessionService) DeleteSession(ctx context.Context, id int64) error {
	err := s.sessionRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("删除会话失败: %w", err)
	}
	return nil
}

// ArchiveSession 归档会话
func (s *SessionService) ArchiveSession(ctx context.Context, id int64) error {
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	session.Status = 0 // 归档状态

	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return fmt.Errorf("归档会话失败: %w", err)
	}

	return nil
}

// ActivateSession 激活会话
func (s *SessionService) ActivateSession(ctx context.Context, id int64) error {
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	session.Status = 1 // 正常状态

	err = s.sessionRepo.Update(ctx, session)
	if err != nil {
		return fmt.Errorf("激活会话失败: %w", err)
	}

	return nil
}

// toSessionResponse 转换为会话响应格式
func (s *SessionService) toSessionResponse(session *types.SessionEntity) *types.SessionResponse {
	return &types.SessionResponse{
		ID:              session.ID,
		UserID:          session.UserID,
		KBID:            session.KBID,
		Title:           session.Title,
		Description:     session.Description,
		Model:           session.Model,
		MaxRounds:       session.MaxRounds,
		EnableRewrite:   session.EnableRewrite,
		KeywordThreshold: session.KeywordThreshold,
		VectorThreshold:  session.VectorThreshold,
		MessageCount:    session.MessageCount,
		Status:          session.Status,
		CreatedAt:       session.CreatedAt,
		UpdatedAt:       session.UpdatedAt,
	}
}
