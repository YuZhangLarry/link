package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	common_repository "link/internal/common/repository"
	"link/internal/types"
)

// SessionRepositoryGORM 会话仓储 - 使用 GORM + Scope
type SessionRepositoryGORM struct {
	base *common_repository.BaseRepository
}

// NewSessionRepositoryGORM 创建会话仓储
func NewSessionRepositoryGORM(db *gorm.DB, tenantEnabled bool) *SessionRepositoryGORM {
	return &SessionRepositoryGORM{
		base: common_repository.NewBaseRepository(db, tenantEnabled),
	}
}

// Create 创建会话
func (r *SessionRepositoryGORM) Create(ctx context.Context, session *types.SessionEntity) error {
	return r.base.Create(ctx, session)
}

// FindByID 根据ID查找会话
func (r *SessionRepositoryGORM) FindByID(ctx context.Context, id string) (*types.SessionEntity, error) {
	tenantID := r.base.GetTenantID(ctx)

	session := &types.SessionEntity{}
	err := r.base.WithTenantAndSoftDeleteScope(ctx, tenantID).
		Where("id = ?", id).
		First(session).Error

	if err == gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("会话不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询会话失败: %w", err)
	}

	return session, nil
}

// FindByUserID 根据用户ID查找会话列表
func (r *SessionRepositoryGORM) FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*types.SessionEntity, int64, error) {
	tenantID := r.base.GetTenantID(ctx)

	var sessions []*types.SessionEntity
	var total int64

	db := r.base.WithTenantAndSoftDeleteScope(ctx, tenantID)

	// 查询总数
	if err := db.Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("查询会话总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&sessions).Error

	if err != nil {
		return nil, 0, fmt.Errorf("查询会话列表失败: %w", err)
	}

	return sessions, total, nil
}

// Update 更新会话
func (r *SessionRepositoryGORM) Update(ctx context.Context, session *types.SessionEntity) error {
	return r.base.Update(ctx, session)
}

// UpdateFields 更新指定字段
func (r *SessionRepositoryGORM) UpdateFields(ctx context.Context, id string, fields map[string]interface{}) error {
	tenantID := r.base.GetTenantID(ctx)
	return r.base.UpdateFields(ctx, "sessions", tenantID, id, fields)
}

// Delete 软删除会话
func (r *SessionRepositoryGORM) Delete(ctx context.Context, id string) error {
	tenantID := r.base.GetTenantID(ctx)
	return r.base.Delete(ctx, "sessions", tenantID, id)
}

// HardDelete 硬删除会话
func (r *SessionRepositoryGORM) HardDelete(ctx context.Context, id string) error {
	tenantID := r.base.GetTenantID(ctx)
	return r.base.HardDelete(ctx, "sessions", tenantID, id)
}

// CountByUserID 统计用户的会话数量
func (r *SessionRepositoryGORM) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	tenantID := r.base.GetTenantID(ctx)
	return r.base.Count(ctx, "sessions", tenantID, "user_id = ?", userID)
}

// ========================================
// 使用示例
// ========================================

// // 在 handler 中使用:
//
//	// 1. 从 context 获取 tenantID
//	tenantID := middleware.GetTenantID(c)
//
//	// 2. 将 tenantID 传入 context (供 repository 使用)
//	ctx := context.WithValue(c.Request.Context(), "tenant_id", tenantID)
//
//	// 3. 调用 repository
//	session, err := sessionRepo.FindByID(ctx, sessionID)
//
// // 或者直接使用 Scope:
//
//	db.Scopes(
//	    repository.TenantScope(tenantID),
//	    repository.SoftDeleteScope(),
//	).Where("user_id = ?", userID).Find(&sessions)
