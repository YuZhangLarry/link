package repository

import (
	"context"
	"fmt"
	common_repository "link/internal/common"
	"log"

	"gorm.io/gorm"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// messageRepository 消息数据访问实现 - GORM 版本
type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息数据访问实例
func NewMessageRepository(db *gorm.DB) interfaces.MessageRepository {
	return &messageRepository{db: db}
}

// Create 创建消息
func (r *messageRepository) Create(ctx context.Context, req *types.CreateMessageRequest) (*types.MessageEntity, error) {
	// 生成 UUID
	messageID := common_repository.GenerateUUID()

	// 从上下文获取租户ID（用于验证 session 属于该租户）
	tenantID := getTenantIDFromContext(ctx)
	log.Printf("📋 [MessageRepository.Create] 开始创建消息: messageID=%s, sessionID=%s, tenantID=%d", messageID, req.SessionID, tenantID)

	// 验证 session 是否存在（如果 tenantID > 0，还需要验证租户）
	var session types.SessionEntity
	query := r.db.WithContext(ctx).Where("id = ?", req.SessionID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&session).Error; err != nil {
		log.Printf("❌ [MessageRepository.Create] Session 查询失败: sessionID=%s, tenantID=%d, error=%v", req.SessionID, tenantID, err)
		return nil, fmt.Errorf("会话不存在或无权访问: %w", err)
	}
	log.Printf("✅ [MessageRepository.Create] Session 验证成功: session.tenantID=%d, session.userID=%d", session.TenantID, session.UserID)

	// 创建消息实体
	message := &types.MessageEntity{
		ID:                  messageID,
		RequestID:           common_repository.GenerateUUID(),
		SessionID:           req.SessionID,
		Role:                req.Role,
		Content:             req.Content,
		KnowledgeReferences: "[]", // 默认空数组
		AgentSteps:          "[]", // 默认空数组
		ToolCalls:           "[]", // 默认空数组
		IsCompleted:         false,
		TokenCount:          req.TokenCount,
	}

	// 如果请求中提供了值，覆盖默认值
	if req.KnowledgeReferences != "" {
		message.KnowledgeReferences = req.KnowledgeReferences
	}
	if req.AgentSteps != "" {
		message.AgentSteps = req.AgentSteps
	}
	if req.ToolCalls != "" {
		message.ToolCalls = req.ToolCalls
	}

	if err := r.db.WithContext(ctx).Create(message).Error; err != nil {
		return nil, fmt.Errorf("创建消息失败: %w", err)
	}

	return message, nil
}

// FindByID 根据ID查找消息
func (r *messageRepository) FindByID(ctx context.Context, id string) (*types.MessageEntity, error) {
	tenantID := getTenantIDFromContext(ctx)

	var message types.MessageEntity
	query := r.db.WithContext(ctx).Where("messages.id = ? AND messages.deleted_at IS NULL", id)
	// 通过 JOIN session 来验证租户权限（仅当 tenantID > 0 时）
	if tenantID > 0 {
		query = query.Joins("JOIN sessions ON sessions.id = messages.session_id").
			Where("sessions.tenant_id = ?", tenantID)
	}
	err := query.First(&message).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("消息不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}

	return &message, nil
}

// FindBySessionID 根据会话ID查找消息列表
func (r *messageRepository) FindBySessionID(ctx context.Context, sessionID string, page, pageSize int) ([]*types.MessageEntity, int64, error) {
	tenantID := getTenantIDFromContext(ctx)

	// 验证 session 是否存在（如果 tenantID > 0，还需要验证租户）
	var session types.SessionEntity
	query := r.db.WithContext(ctx).Where("id = ?", sessionID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&session).Error; err != nil {
		return nil, 0, fmt.Errorf("会话不存在或无权访问: %w", err)
	}

	// 查询总数
	var total int64
	if err := r.db.WithContext(ctx).Model(&types.MessageEntity{}).Where("session_id = ? AND deleted_at IS NULL", sessionID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询消息总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	var messages []*types.MessageEntity
	if err := r.db.WithContext(ctx).
		Where("session_id = ? AND deleted_at IS NULL", sessionID).
		Order("created_at ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("查询消息列表失败: %w", err)
	}

	return messages, total, nil
}

// FindBySessionIDAndRole 根据会话ID和角色查找消息
func (r *messageRepository) FindBySessionIDAndRole(ctx context.Context, sessionID string, role string, page, pageSize int) ([]*types.MessageEntity, int64, error) {
	tenantID := getTenantIDFromContext(ctx)

	// 验证 session 是否存在（如果 tenantID > 0，还需要验证租户）
	var session types.SessionEntity
	query := r.db.WithContext(ctx).Where("id = ?", sessionID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&session).Error; err != nil {
		return nil, 0, fmt.Errorf("会话不存在或无权访问: %w", err)
	}

	// 查询总数
	var total int64
	if err := r.db.WithContext(ctx).Model(&types.MessageEntity{}).
		Where("session_id = ? AND role = ? AND deleted_at IS NULL", sessionID, role).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询消息总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	var messages []*types.MessageEntity
	if err := r.db.WithContext(ctx).
		Where("session_id = ? AND role = ? AND deleted_at IS NULL", sessionID, role).
		Order("created_at ASC").
		Limit(pageSize).
		Offset(offset).
		Find(&messages).Error; err != nil {
		return nil, 0, fmt.Errorf("查询消息列表失败: %w", err)
	}

	return messages, total, nil
}

// Update 更新消息
func (r *messageRepository) Update(ctx context.Context, id string, req *types.UpdateMessageRequest) error {
	tenantID := getTenantIDFromContext(ctx)

	// 先验证消息是否存在且有权限
	var message types.MessageEntity
	query := r.db.WithContext(ctx).Where("messages.id = ? AND messages.deleted_at IS NULL", id)
	if tenantID > 0 {
		query = query.Joins("JOIN sessions ON sessions.id = messages.session_id").
			Where("sessions.tenant_id = ?", tenantID)
	}
	if err := query.First(&message).Error; err != nil {
		return fmt.Errorf("消息不存在或无权访问: %w", err)
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Content != nil {
		updates["content"] = *req.Content
	}
	if req.IsCompleted != nil {
		updates["is_completed"] = *req.IsCompleted
	}
	if req.TokenCount != nil {
		updates["token_count"] = *req.TokenCount
	}
	if req.KnowledgeReferences != nil {
		updates["knowledge_references"] = *req.KnowledgeReferences
	}
	if req.AgentSteps != nil {
		updates["agent_steps"] = *req.AgentSteps
	}

	if err := r.db.WithContext(ctx).Model(&message).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新消息失败: %w", err)
	}

	return nil
}

// Delete 删除消息（软删除）
func (r *messageRepository) Delete(ctx context.Context, id string) error {
	tenantID := getTenantIDFromContext(ctx)

	// 使用 GORM 的软删除功能
	query := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id)
	if tenantID > 0 {
		query = query.Where("id IN (SELECT messages.id FROM messages JOIN sessions ON sessions.id = messages.session_id WHERE sessions.tenant_id = ?)", tenantID)
	}
	result := query.Delete(&types.MessageEntity{})

	if result.Error != nil {
		return fmt.Errorf("删除消息失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("消息不存在或无权访问")
	}

	return nil
}

// DeleteBySessionID 删除会话的所有消息
func (r *messageRepository) DeleteBySessionID(ctx context.Context, sessionID string) error {
	tenantID := getTenantIDFromContext(ctx)

	// 验证 session 是否存在（如果 tenantID > 0，还需要验证租户）
	var session types.SessionEntity
	query := r.db.WithContext(ctx).Where("id = ?", sessionID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&session).Error; err != nil {
		return fmt.Errorf("会话不存在或无权访问: %w", err)
	}

	// 软删除该会话的所有消息
	if err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Delete(&types.MessageEntity{}).Error; err != nil {
		return fmt.Errorf("删除会话消息失败: %w", err)
	}

	return nil
}

// CountBySessionID 统计会话的消息数量
func (r *messageRepository) CountBySessionID(ctx context.Context, sessionID string) (int64, error) {
	tenantID := getTenantIDFromContext(ctx)

	// 验证 session 是否存在（如果 tenantID > 0，还需要验证租户）
	var session types.SessionEntity
	query := r.db.WithContext(ctx).Where("id = ?", sessionID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&session).Error; err != nil {
		return 0, fmt.Errorf("会话不存在或无权访问: %w", err)
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&types.MessageEntity{}).
		Where("session_id = ? AND deleted_at IS NULL", sessionID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计消息数量失败: %w", err)
	}

	return count, nil
}

// ========================================
// 辅助方法
// ========================================
// getTenantIDFromContext 从上下文获取租户ID（在 session.go 中定义）
