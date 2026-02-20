package service

import (
	"context"
	"fmt"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// SessionService 会话服务实现（聚合 RAG 配置）
type SessionService struct {
	sessionRepo   interfaces.SessionRepository
	messageRepo   interfaces.MessageRepository
	retrievalRepo interfaces.RetrievalSettingRepository
}

// NewSessionService 创建会话服务实例
func NewSessionService(
	sessionRepo interfaces.SessionRepository,
	messageRepo interfaces.MessageRepository,
	retrievalRepo interfaces.RetrievalSettingRepository,
) interfaces.SessionService {
	return &SessionService{
		sessionRepo:   sessionRepo,
		messageRepo:   messageRepo,
		retrievalRepo: retrievalRepo,
	}
}

// CreateSession 创建会话（同时保存 RAG 配置）
func (s *SessionService) CreateSession(ctx context.Context, userID int64, req *types.CreateSessionRequest) (*types.SessionResponse, error) {
	// 调用仓储创建会话
	session, err := s.sessionRepo.Create(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	// 如果请求包含 RAG 配置，保存到 retrieval_settings 表
	if req.RAGConfig != nil {
		tenantID := session.TenantID
		if tenantID == 0 {
			if tid, ok := ctx.Value("tenant_id").(int64); ok {
				tenantID = tid
			}
		}
		if tenantID > 0 {
			if err := s.retrievalRepo.UpsertBySessionID(ctx, session.ID, tenantID, req.RAGConfig); err != nil {
				// 记录错误但不影响会话创建
				fmt.Printf("⚠️ [CreateSession] 保存 RAG 配置失败: %v\n", err)
			}
		}
	}

	// 构建响应，包含 RAG 配置
	resp := s.toSessionResponse(session)
	if req.RAGConfig != nil {
		resp.RAGConfig = req.RAGConfig
	}
	return resp, nil
}

// GetSessionByID 根据ID获取会话（包含 RAG 配置）
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
		return s.buildSessionResponseWithRAG(ctx, session)
	}

	// 如果没有 user_id，则直接查询（兼容旧逻辑）
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.buildSessionResponseWithRAG(ctx, session)
}

// GetSessionDetail 获取会话详情（包含消息和 RAG 配置）
func (s *SessionService) GetSessionDetail(ctx context.Context, id string) (*types.SessionDetailResponse, error) {
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 查询消息列表（获取所有消息，不分页）
	messageEntities, _, err := s.messageRepo.FindBySessionID(ctx, id, 1, 10000)
	if err != nil {
		return nil, fmt.Errorf("查询消息列表失败: %w", err)
	}

	// 转换为 MessageResponse
	messages := make([]*types.MessageResponse, len(messageEntities))
	for i, msg := range messageEntities {
		messages[i] = &types.MessageResponse{
			ID:                  msg.ID,
			RequestID:           msg.RequestID,
			SessionID:           msg.SessionID,
			Role:                msg.Role,
			Content:             msg.Content,
			KnowledgeReferences: msg.KnowledgeReferences,
			AgentSteps:          msg.AgentSteps,
			ToolCalls:           msg.ToolCalls,
			IsCompleted:         msg.IsCompleted,
			TokenCount:          msg.TokenCount,
			CreatedAt:           msg.CreatedAt,
		}
	}

	// 构建会话响应（包含 RAG 配置）
	sessionResp, err := s.buildSessionResponseWithRAG(ctx, session)
	if err != nil {
		return nil, err
	}

	return &types.SessionDetailResponse{
		Session:  sessionResp,
		Messages: messages,
	}, nil
}

// ListSessions 查询会话列表（包含 RAG 配置）
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

	// 转换为响应格式（批量加载 RAG 配置）
	sessionResponses := make([]*types.SessionResponse, 0, len(sessions))
	for _, session := range sessions {
		resp, err := s.buildSessionResponseWithRAG(ctx, session)
		if err != nil {
			// 如果加载 RAG 配置失败，仍然返回会话信息
			resp = s.toSessionResponse(session)
		}
		sessionResponses = append(sessionResponses, resp)
	}

	return &types.SessionListResponse{
		Sessions: sessionResponses,
		Total:    total,
		Page:     page,
		Size:     size,
	}, nil
}

// UpdateSession 更新会话（同时更新 RAG 配置）
func (s *SessionService) UpdateSession(ctx context.Context, id string, req *types.UpdateSessionRequest) (*types.SessionResponse, error) {
	// 更新会话基本信息
	err := s.sessionRepo.Update(ctx, id, req)
	if err != nil {
		return nil, fmt.Errorf("更新会话失败: %w", err)
	}

	// 如果请求包含 RAG 配置更新，更新到 retrieval_settings 表
	if req.RAGConfig != nil {
		// 获取会话以确定 tenant_id
		session, err := s.sessionRepo.FindByID(ctx, id)
		if err == nil && session.TenantID > 0 {
			if err := s.retrievalRepo.UpsertBySessionID(ctx, id, session.TenantID, req.RAGConfig); err != nil {
				fmt.Printf("⚠️ [UpdateSession] 更新 RAG 配置失败: %v\n", err)
			}
		}
	}

	// 重新获取更新后的会话
	session, err := s.sessionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 构建响应，包含 RAG 配置
	return s.buildSessionResponseWithRAG(ctx, session)
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

// buildSessionResponseWithRAG 构建包含 RAG 配置的会话响应
func (s *SessionService) buildSessionResponseWithRAG(ctx context.Context, session *types.SessionEntity) (*types.SessionResponse, error) {
	resp := s.toSessionResponse(session)

	// 尝试从 retrieval_settings 表加载 RAG 配置
	retrievalSetting, err := s.retrievalRepo.FindBySessionID(ctx, session.ID)
	if err == nil && retrievalSetting != nil {
		// 将 RetrievalSetting 转换为 RAGConfig
		resp.RAGConfig = s.convertToRAGConfig(retrievalSetting)
	}
	// 如果没有找到 RAG 配置，保持为 nil（前端使用默认配置）

	return resp, nil
}

// convertToRAGConfig 将 RetrievalSetting 转换为 RAGConfig
func (s *SessionService) convertToRAGConfig(setting *types.RetrievalSetting) *types.RAGConfig {
	config := &types.RAGConfig{
		Enabled:             false,
		KBID:                "",
		VectorTopK:          15,
		KeywordTopK:         15,
		GraphTopK:           10,
		SimilarityThreshold: 0.0,
		Alpha:               0.6,
		RetrievalModes:      []string{"vector"},
	}

	if setting.VectorTopK != nil {
		config.VectorTopK = *setting.VectorTopK
	}
	if setting.VectorThreshold != nil {
		config.SimilarityThreshold = float64(*setting.VectorThreshold)
	}
	if setting.BM25TopK != nil {
		config.KeywordTopK = *setting.BM25TopK
	}
	if setting.GraphTopK != nil {
		config.GraphTopK = *setting.GraphTopK
	}
	if setting.GraphEnabled != nil && *setting.GraphEnabled {
		if len(config.RetrievalModes) == 1 && config.RetrievalModes[0] == "vector" {
			config.RetrievalModes = []string{"vector", "graph"}
		} else {
			config.RetrievalModes = append(config.RetrievalModes, "graph")
		}
	}
	if setting.HybridAlpha != nil {
		config.Alpha = float32(*setting.HybridAlpha)
	}

	return config
}

// toSessionResponse 转换为会话响应格式（不包含 RAG 配置）
func (s *SessionService) toSessionResponse(session *types.SessionEntity) *types.SessionResponse {
	return &types.SessionResponse{
		ID:           session.ID,
		TenantID:     session.TenantID,
		UserID:       session.UserID,
		Title:        session.Title,
		Description:  session.Description,
		Status:       session.Status,
		MessageCount: session.MessageCount,
		CreatedAt:    session.CreatedAt,
		UpdatedAt:    session.UpdatedAt,
	}
}
