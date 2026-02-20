package repository

import (
	"context"
	"database/sql"
	"fmt"
	"link/internal/types"
	"link/internal/types/interfaces"
	"time"

	"github.com/jmoiron/sqlx"
)

// ========================================
// 仓储类型别名
// ========================================
type RoleRepository = interfaces.RoleRepository

// ========================================
// 角色仓储实现
// ========================================

// roleRepository 角色仓储实现
type roleRepository struct {
	db *sqlx.DB
}

// NewRoleRepository 创建角色仓储
func NewRoleRepository(db *sqlx.DB) RoleRepository {
	return &roleRepository{db: db}
}

// Create 创建角色
func (r *roleRepository) Create(ctx context.Context, role *types.Role) error {
	query := `
		INSERT INTO roles (tenant_id, name, code, description, is_system, is_default, level, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		role.TenantID, role.Name, role.Code, role.Description,
		role.IsSystem, role.IsDefault, role.Level, role.Status,
	)
	if err != nil {
		return fmt.Errorf("创建角色失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	role.ID = id
	return nil
}

// FindByID 根据ID查找角色
func (r *roleRepository) FindByID(ctx context.Context, id int64) (*types.Role, error) {
	query := `
		SELECT id, tenant_id, name, code, description, is_system, is_default, level, status, created_at, updated_at, deleted_at
		FROM roles
		WHERE id = ? AND deleted_at IS NULL
	`
	var role types.Role
	err := r.db.GetContext(ctx, &role, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("角色不存在")
		}
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}
	return &role, nil
}

// FindByTenantID 根据租户ID查找角色列表
func (r *roleRepository) FindByTenantID(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.Role, int64, error) {
	// 查询总数
	countQuery := `SELECT COUNT(*) FROM roles WHERE tenant_id = ? AND deleted_at IS NULL`
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("查询角色总数失败: %w", err)
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := `
		SELECT id, tenant_id, name, code, description, is_system, is_default, level, status, created_at, updated_at, deleted_at
		FROM roles
		WHERE tenant_id = ? AND deleted_at IS NULL
		ORDER BY level DESC, created_at ASC
		LIMIT ? OFFSET ?
	`
	var roles []*types.Role
	err = r.db.SelectContext(ctx, &roles, query, tenantID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询角色列表失败: %w", err)
	}

	return roles, total, nil
}

// FindByCode 根据租户ID和角色编码查找角色
func (r *roleRepository) FindByCode(ctx context.Context, tenantID int64, code string) (*types.Role, error) {
	query := `
		SELECT id, tenant_id, name, code, description, is_system, is_default, level, status, created_at, updated_at, deleted_at
		FROM roles
		WHERE tenant_id = ? AND code = ? AND deleted_at IS NULL
		LIMIT 1
	`
	var role types.Role
	err := r.db.GetContext(ctx, &role, query, tenantID, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("角色不存在")
		}
		return nil, fmt.Errorf("查询角色失败: %w", err)
	}
	return &role, nil
}

// Update 更新角色
func (r *roleRepository) Update(ctx context.Context, role *types.Role) error {
	query := `
		UPDATE roles
		SET name = ?, description = ?, level = ?, status = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	role.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		role.Name, role.Description, role.Level, role.Status, role.UpdatedAt, role.ID,
	)
	if err != nil {
		return fmt.Errorf("更新角色失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("角色不存在或已删除")
	}

	return nil
}

// Delete 删除角色（软删除）
func (r *roleRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE roles SET deleted_at = ? WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("删除角色失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("角色不存在")
	}

	return nil
}

// List 分页查询角色列表
func (r *roleRepository) List(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.Role, int64, error) {
	return r.FindByTenantID(ctx, tenantID, page, pageSize)
}
