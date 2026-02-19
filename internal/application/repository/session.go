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

// sessionRepository 会话数据访问实现 - GORM 版本
type sessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository 创建会话数据访问实例
func NewSessionRepository(db *gorm.DB) interfaces.SessionRepository {
	return &sessionRepository{db: db}
}

// Create 创建会话
func (r *sessionRepository) Create(ctx context.Context, userID int64, req *types.CreateSessionRequest) (*types.SessionEntity, error) {
	// 生成 UUID
	sessionID := common_repository.GenerateUUID()

	// 获取租户ID
	tenantID := getTenantIDFromContext(ctx)

	// 创建会话实体
	session := &types.SessionEntity{
		ID:          sessionID,
		TenantID:    tenantID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      1, // 默认正常状态
	}

	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, fmt.Errorf("创建会话失败: %w", err)
	}

	return session, nil
}

// FindByID 根据ID查找会话
func (r *sessionRepository) FindByID(ctx context.Context, id string) (*types.SessionEntity, error) {
	tenantID := getTenantIDFromContext(ctx)

	var session types.SessionEntity
	query := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	err := query.First(&session).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("会话不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询会话失败: %w", err)
	}

	return &session, nil
}

// FindByUserID 根据用户ID查找会话列表
func (r *sessionRepository) FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*types.SessionEntity, int64, error) {
	tenantID := getTenantIDFromContext(ctx)

	// 查询总数
	var total int64
	query := r.db.WithContext(ctx).Model(&types.SessionEntity{}).Where("user_id = ? AND deleted_at IS NULL", userID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询会话总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	var sessions []*types.SessionEntity
	query = r.db.WithContext(ctx).Where("user_id = ? AND deleted_at IS NULL", userID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.Order("updated_at DESC").Limit(pageSize).Offset(offset).Find(&sessions).Error; err != nil {
		return nil, 0, fmt.Errorf("查询会话列表失败: %w", err)
	}

	return sessions, total, nil
}

// FindByUserIDAndStatus 根据用户ID和状态查找会话
func (r *sessionRepository) FindByUserIDAndStatus(ctx context.Context, userID int64, status int8, page, pageSize int) ([]*types.SessionEntity, int64, error) {
	tenantID := getTenantIDFromContext(ctx)

	// 查询总数
	var total int64
	query := r.db.WithContext(ctx).Model(&types.SessionEntity{}).
		Where("user_id = ? AND status = ? AND deleted_at IS NULL", userID, status)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询会话总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	var sessions []*types.SessionEntity
	query = r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND deleted_at IS NULL", userID, status)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.Order("updated_at DESC").Limit(pageSize).Offset(offset).Find(&sessions).Error; err != nil {
		return nil, 0, fmt.Errorf("查询会话列表失败: %w", err)
	}

	return sessions, total, nil
}

// Update 更新会话
func (r *sessionRepository) Update(ctx context.Context, id string, req *types.UpdateSessionRequest) error {
	tenantID := getTenantIDFromContext(ctx)

	// 先验证会话是否存在且有权限
	var session types.SessionEntity
	query := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&session).Error; err != nil {
		return fmt.Errorf("会话不存在或无权访问: %w", err)
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := r.db.WithContext(ctx).Model(&session).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新会话失败: %w", err)
	}

	return nil
}

// UpdateMessageCount 更新会话消息数量（no-op，消息数量动态计算）
func (r *sessionRepository) UpdateMessageCount(ctx context.Context, sessionID string) error {
	return nil
}

// Delete 删除会话（软删除）
func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	tenantID := getTenantIDFromContext(ctx)

	// 使用 GORM 的软删除功能
	query := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	result := query.Delete(&types.SessionEntity{})

	if result.Error != nil {
		return fmt.Errorf("删除会话失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("会话不存在或无权访问")
	}

	return nil
}

// HardDelete 硬删除会话
func (r *sessionRepository) HardDelete(ctx context.Context, id string) error {
	tenantID := getTenantIDFromContext(ctx)

	// 使用 Unscoped 进行硬删除
	query := r.db.WithContext(ctx).Unscoped().Where("id = ?", id)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	result := query.Delete(&types.SessionEntity{})

	if result.Error != nil {
		return fmt.Errorf("硬删除会话失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("会话不存在或无权访问")
	}

	return nil
}

// CountByUserID 统计用户的会话数量
func (r *sessionRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	tenantID := getTenantIDFromContext(ctx)

	var count int64
	query := r.db.WithContext(ctx).Model(&types.SessionEntity{}).
		Where("user_id = ? AND deleted_at IS NULL", userID)
	if tenantID > 0 {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("统计会话数量失败: %w", err)
	}

	return count, nil
}

// IncrementMessageCount 增加消息计数（no-op）
func (r *sessionRepository) IncrementMessageCount(ctx context.Context, sessionID string) error {
	return nil
}

// ========================================
// 辅助方法
// ========================================

// getTenantIDFromContext 从上下文获取租户ID
func getTenantIDFromContext(ctx context.Context) int64 {
	if tenantID, ok := ctx.Value("tenant_id").(int64); ok {
		log.Printf("📋 [getTenantIDFromContext] 获取 tenant_id = %d", tenantID)
		return tenantID
	}
	log.Printf("⚠️  [getTenantIDFromContext] context 中没有 tenant_id，返回 0")
	return 0
}
