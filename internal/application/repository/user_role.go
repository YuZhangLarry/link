package repository

import (
	"context"
	"database/sql"
	"fmt"
	"link/internal/types"
	"link/internal/types/interfaces"

	"github.com/jmoiron/sqlx"
)

// ========================================
// 仓储类型别名
// ========================================
type UserRoleRepository = interfaces.UserRoleRepository

// ========================================
// 用户角色仓储实现
// ========================================

// userRoleRepository 用户角色仓储实现
type userRoleRepository struct {
	db *sqlx.DB
}

// NewUserRoleRepository 创建用户角色仓储
func NewUserRoleRepository(db *sqlx.DB) UserRoleRepository {
	return &userRoleRepository{db: db}
}

// Create 创建用户角色关联
func (r *userRoleRepository) Create(ctx context.Context, userRole *types.UserRole) error {
	query := `
		INSERT INTO user_roles (tenant_id, user_id, role_id, assigned_by, assigned_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		userRole.TenantID, userRole.UserID, userRole.RoleID,
		userRole.AssignedBy, userRole.AssignedAt, userRole.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("创建用户角色关联失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	userRole.ID = id
	return nil
}

// FindByUserID 根据用户ID查找其角色
func (r *userRoleRepository) FindByUserID(ctx context.Context, tenantID, userID int64) (*types.UserRole, error) {
	query := `
		SELECT ur.id, ur.tenant_id, ur.user_id, ur.role_id, ur.assigned_by, ur.assigned_at, ur.expires_at,
		       r.name as role_name, r.code as role_code, r.level as role_level
		FROM user_roles ur
		LEFT JOIN roles r ON ur.role_id = r.id
		WHERE ur.tenant_id = ? AND ur.user_id = ?
		AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
		ORDER BY r.level DESC
		LIMIT 1
	`

	type UserRoleWithRole struct {
		types.UserRole
		RoleName  string `db:"role_name"`
		RoleCode  string `db:"role_code"`
		RoleLevel int    `db:"role_level"`
	}

	var result UserRoleWithRole
	err := r.db.GetContext(ctx, &result, query, tenantID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("用户角色不存在")
		}
		return nil, fmt.Errorf("查询用户角色失败: %w", err)
	}

	return &result.UserRole, nil
}

// FindByRoleID 根据角色ID查找拥有该角色的用户列表
func (r *userRoleRepository) FindByRoleID(ctx context.Context, roleID int64, page, pageSize int) ([]*types.UserRole, int64, error) {
	// 查询总数
	countQuery := `SELECT COUNT(*) FROM user_roles WHERE role_id = ? AND (expires_at IS NULL OR expires_at > NOW())`
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, roleID)
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户角色总数失败: %w", err)
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := `
		SELECT id, tenant_id, user_id, role_id, assigned_by, assigned_at, expires_at
		FROM user_roles
		WHERE role_id = ? AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY assigned_at DESC
		LIMIT ? OFFSET ?
	`
	var userRoles []*types.UserRole
	err = r.db.SelectContext(ctx, &userRoles, query, roleID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户角色列表失败: %w", err)
	}

	return userRoles, total, nil
}

// Update 更新用户角色（实际是删除旧的，创建新的）
func (r *userRoleRepository) Update(ctx context.Context, userRole *types.UserRole) error {
	// 先删除旧的角色关联
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM user_roles WHERE tenant_id = ? AND user_id = ?
	`, userRole.TenantID, userRole.UserID)
	if err != nil {
		return fmt.Errorf("删除旧的用户角色失败: %w", err)
	}

	// 创建新的角色关联
	return r.Create(ctx, userRole)
}

// Delete 删除用户角色关联
func (r *userRoleRepository) Delete(ctx context.Context, tenantID, userID int64) error {
	query := `DELETE FROM user_roles WHERE tenant_id = ? AND user_id = ?`
	result, err := r.db.ExecContext(ctx, query, tenantID, userID)
	if err != nil {
		return fmt.Errorf("删除用户角色失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("用户角色不存在")
	}

	return nil
}

// DeleteByRoleID 删除角色的所有用户关联
func (r *userRoleRepository) DeleteByRoleID(ctx context.Context, roleID int64) error {
	query := `DELETE FROM user_roles WHERE role_id = ?`
	_, err := r.db.ExecContext(ctx, query, roleID)
	if err != nil {
		return fmt.Errorf("删除角色用户关联失败: %w", err)
	}
	return nil
}
