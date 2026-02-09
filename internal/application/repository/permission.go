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
// 权限仓储实现
// ========================================

// ========================================
// 仓储类型别名
// ========================================
// 为了避免命名冲突，使用类型别名
type (
	// PermissionRepository 权限仓储接口
	PermissionRepository = interfaces.PermissionRepository
	// RolePermissionRepository 角色权限仓储接口
	RolePermissionRepository = interfaces.RolePermissionRepository
	// ResourcePermissionRepository 资源权限仓储接口
	ResourcePermissionRepository = interfaces.ResourcePermissionRepository
	// PermissionAuditLogRepository 权限审计日志仓储接口
	PermissionAuditLogRepository = interfaces.PermissionAuditLogRepository
)

// ========================================
// 权限仓储实现
// ========================================

// permissionRepository 权限仓储实现
type permissionRepository struct {
	db *sqlx.DB
}

// NewPermissionRepository 创建权限仓储
func NewPermissionRepository(db *sqlx.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

// Create 创建权限
func (r *permissionRepository) Create(ctx context.Context, permission *types.Permission) error {
	query := `
		INSERT INTO permissions (resource_type, action, description, is_system)
		VALUES (?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		permission.ResourceType, permission.Action,
		permission.Description, permission.IsSystem,
	)
	if err != nil {
		return fmt.Errorf("创建权限失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	permission.ID = id
	return nil
}

// FindByID 根据ID查找权限
func (r *permissionRepository) FindByID(ctx context.Context, id int64) (*types.Permission, error) {
	query := `
		SELECT id, resource_type, action, description, is_system, created_at, updated_at
		FROM permissions
		WHERE id = ?
	`
	var permission types.Permission
	err := r.db.GetContext(ctx, &permission, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("权限不存在")
		}
		return nil, fmt.Errorf("查询权限失败: %w", err)
	}
	return &permission, nil
}

// FindAll 查找所有权限
func (r *permissionRepository) FindAll(ctx context.Context) ([]*types.Permission, error) {
	query := `
		SELECT id, resource_type, action, description, is_system, created_at, updated_at
		FROM permissions
		ORDER BY resource_type, action
	`
	var permissions []*types.Permission
	err := r.db.SelectContext(ctx, &permissions, query)
	if err != nil {
		return nil, fmt.Errorf("查询所有权限失败: %w", err)
	}
	return permissions, nil
}

// FindByResourceType 根据资源类型查找权限
func (r *permissionRepository) FindByResourceType(ctx context.Context, resourceType string) ([]*types.Permission, error) {
	query := `
		SELECT id, resource_type, action, description, is_system, created_at, updated_at
		FROM permissions
		WHERE resource_type = ?
		ORDER BY action
	`
	var permissions []*types.Permission
	err := r.db.SelectContext(ctx, &permissions, query, resourceType)
	if err != nil {
		return nil, fmt.Errorf("查询资源权限失败: %w", err)
	}
	return permissions, nil
}

// FindByRoleID 根据角色ID查找其拥有的所有权限
func (r *permissionRepository) FindByRoleID(ctx context.Context, roleID int64) ([]*types.Permission, error) {
	query := `
		SELECT p.id, p.resource_type, p.action, p.description, p.is_system, p.created_at, p.updated_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ?
		ORDER BY p.resource_type, p.action
	`
	var permissions []*types.Permission
	err := r.db.SelectContext(ctx, &permissions, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("查询角色权限失败: %w", err)
	}
	return permissions, nil
}

// ========================================
// 角色权限仓储实现
// ========================================

// rolePermissionRepository 角色权限仓储实现
type rolePermissionRepository struct {
	db *sqlx.DB
}

// NewRolePermissionRepository 创建角色权限仓储
func NewRolePermissionRepository(db *sqlx.DB) RolePermissionRepository {
	return &rolePermissionRepository{db: db}
}

// Create 创建角色权限关联
func (r *rolePermissionRepository) Create(ctx context.Context, rolePermission *types.RolePermission) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES (?, ?)
	`
	result, err := r.db.ExecContext(ctx, query, rolePermission.RoleID, rolePermission.PermissionID)
	if err != nil {
		return fmt.Errorf("创建角色权限关联失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	rolePermission.ID = id
	return nil
}

// FindByRoleID 根据角色ID查找权限ID列表
func (r *rolePermissionRepository) FindByRoleID(ctx context.Context, roleID int64) ([]int64, error) {
	query := `SELECT permission_id FROM role_permissions WHERE role_id = ?`
	var permissionIDs []int64
	err := r.db.SelectContext(ctx, &permissionIDs, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("查询角色权限ID列表失败: %w", err)
	}
	return permissionIDs, nil
}

// Delete 删除角色的某个权限
func (r *rolePermissionRepository) Delete(ctx context.Context, roleID, permissionID int64) error {
	query := `DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?`
	_, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("删除角色权限失败: %w", err)
	}
	return nil
}

// DeleteByRoleID 删除角色的所有权限
func (r *rolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID int64) error {
	query := `DELETE FROM role_permissions WHERE role_id = ?`
	_, err := r.db.ExecContext(ctx, query, roleID)
	if err != nil {
		return fmt.Errorf("删除角色所有权限失败: %w", err)
	}
	return nil
}

// BatchCreate 批量创建角色权限关联
func (r *rolePermissionRepository) BatchCreate(ctx context.Context, rolePermissions []*types.RolePermission) error {
	if len(rolePermissions) == 0 {
		return nil
	}

	query := `
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES (?, ?)
	`
	stmt, err := r.db.Preparex(query)
	if err != nil {
		return fmt.Errorf("准备批量插入语句失败: %w", err)
	}
	defer stmt.Close()

	for _, rp := range rolePermissions {
		_, err := stmt.ExecContext(ctx, rp.RoleID, rp.PermissionID)
		if err != nil {
			return fmt.Errorf("批量插入角色权限失败: %w", err)
		}
	}

	return nil
}

// ========================================
// 资源级权限仓储实现
// ========================================

// resourcePermissionRepository 资源级权限仓储实现
type resourcePermissionRepository struct {
	db *sqlx.DB
}

// NewResourcePermissionRepository 创建资源级权限仓储
func NewResourcePermissionRepository(db *sqlx.DB) ResourcePermissionRepository {
	return &resourcePermissionRepository{db: db}
}

// Create 创建资源级权限
func (r *resourcePermissionRepository) Create(ctx context.Context, resourcePermission *types.ResourcePermission) error {
	query := `
		INSERT INTO resource_permissions (tenant_id, user_id, resource_type, resource_id, permission_type, granted_by, granted_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		resourcePermission.TenantID, resourcePermission.UserID,
		resourcePermission.ResourceType, resourcePermission.ResourceID,
		resourcePermission.PermissionType, resourcePermission.GrantedBy,
		resourcePermission.GrantedAt, resourcePermission.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("创建资源级权限失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	resourcePermission.ID = id
	return nil
}

// FindByUserID 根据用户ID查找其资源权限
func (r *resourcePermissionRepository) FindByUserID(ctx context.Context, tenantID, userID int64) ([]*types.ResourcePermission, error) {
	query := `
		SELECT id, tenant_id, user_id, resource_type, resource_id, permission_type, granted_by, granted_at, expires_at
		FROM resource_permissions
		WHERE tenant_id = ? AND user_id = ?
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY resource_type, resource_id
	`
	var permissions []*types.ResourcePermission
	err := r.db.SelectContext(ctx, &permissions, query, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("查询用户资源权限失败: %w", err)
	}
	return permissions, nil
}

// FindByResource 根据资源查找权限列表
func (r *resourcePermissionRepository) FindByResource(ctx context.Context, tenantID int64, resourceType, resourceID string) ([]*types.ResourcePermission, error) {
	query := `
		SELECT id, tenant_id, user_id, resource_type, resource_id, permission_type, granted_by, granted_at, expires_at
		FROM resource_permissions
		WHERE tenant_id = ? AND resource_type = ? AND resource_id = ?
		AND (expires_at IS NULL OR expires_at > NOW())
		ORDER BY permission_type
	`
	var permissions []*types.ResourcePermission
	err := r.db.SelectContext(ctx, &permissions, query, tenantID, resourceType, resourceID)
	if err != nil {
		return nil, fmt.Errorf("查询资源权限失败: %w", err)
	}
	return permissions, nil
}

// CheckPermission 检查用户对资源是否有指定权限
func (r *resourcePermissionRepository) CheckPermission(ctx context.Context, tenantID, userID int64, resourceType, resourceID, permissionType string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM resource_permissions
		WHERE tenant_id = ? AND user_id = ?
		AND resource_type = ? AND resource_id = ? AND permission_type = ?
		AND (expires_at IS NULL OR expires_at > NOW())
		LIMIT 1
	`
	var count int
	err := r.db.GetContext(ctx, &count, query, tenantID, userID, resourceType, resourceID, permissionType)
	if err != nil {
		return false, fmt.Errorf("检查资源权限失败: %w", err)
	}
	return count > 0, nil
}

// Delete 删除资源级权限
func (r *resourcePermissionRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM resource_permissions WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除资源级权限失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("资源级权限不存在")
	}

	return nil
}

// DeleteByResource 删除资源的所有权限
func (r *resourcePermissionRepository) DeleteByResource(ctx context.Context, tenantID int64, resourceType, resourceID string) error {
	query := `DELETE FROM resource_permissions WHERE tenant_id = ? AND resource_type = ? AND resource_id = ?`
	_, err := r.db.ExecContext(ctx, query, tenantID, resourceType, resourceID)
	if err != nil {
		return fmt.Errorf("删除资源所有权限失败: %w", err)
	}
	return nil
}

// ========================================
// 权限审计日志仓储实现
// ========================================

// permissionAuditLogRepository 权限审计日志仓储实现
type permissionAuditLogRepository struct {
	db *sqlx.DB
}

// NewPermissionAuditLogRepository 创建权限审计日志仓储
func NewPermissionAuditLogRepository(db *sqlx.DB) PermissionAuditLogRepository {
	return &permissionAuditLogRepository{db: db}
}

// Create 创建审计日志
func (r *permissionAuditLogRepository) Create(ctx context.Context, log *types.PermissionAuditLog) error {
	query := `
		INSERT INTO permission_audit_logs (tenant_id, user_id, operator_id, operation_type, target_type, target_id, before_value, after_value, reason, ip_address, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, query,
		log.TenantID, log.UserID, log.OperatorID,
		log.OperationType, log.TargetType, log.TargetID,
		log.BeforeValue, log.AfterValue, log.Reason, log.IPAddress,
		log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("创建权限审计日志失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	log.ID = id
	return nil
}

// FindByUserID 根据用户ID查找审计日志
func (r *permissionAuditLogRepository) FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*types.PermissionAuditLog, int64, error) {
	// 查询总数
	countQuery := `SELECT COUNT(*) FROM permission_audit_logs WHERE user_id = ?`
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("查询审计日志总数失败: %w", err)
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := `
		SELECT id, tenant_id, user_id, operator_id, operation_type, target_type, target_id, before_value, after_value, reason, ip_address, created_at
		FROM permission_audit_logs
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	var logs []*types.PermissionAuditLog
	err = r.db.SelectContext(ctx, &logs, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询审计日志列表失败: %w", err)
	}

	return logs, total, nil
}

// FindByOperatorID 根据操作人ID查找审计日志
func (r *permissionAuditLogRepository) FindByOperatorID(ctx context.Context, operatorID int64, page, pageSize int) ([]*types.PermissionAuditLog, int64, error) {
	// 查询总数
	countQuery := `SELECT COUNT(*) FROM permission_audit_logs WHERE operator_id = ?`
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, operatorID)
	if err != nil {
		return nil, 0, fmt.Errorf("查询审计日志总数失败: %w", err)
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := `
		SELECT id, tenant_id, user_id, operator_id, operation_type, target_type, target_id, before_value, after_value, reason, ip_address, created_at
		FROM permission_audit_logs
		WHERE operator_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	var logs []*types.PermissionAuditLog
	err = r.db.SelectContext(ctx, &logs, query, operatorID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询审计日志列表失败: %w", err)
	}

	return logs, total, nil
}

// FindByTenantID 根据租户ID查找审计日志
func (r *permissionAuditLogRepository) FindByTenantID(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.PermissionAuditLog, int64, error) {
	// 查询总数
	countQuery := `SELECT COUNT(*) FROM permission_audit_logs WHERE tenant_id = ?`
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("查询审计日志总数失败: %w", err)
	}

	// 查询列表
	offset := (page - 1) * pageSize
	query := `
		SELECT id, tenant_id, user_id, operator_id, operation_type, target_type, target_id, before_value, after_value, reason, ip_address, created_at
		FROM permission_audit_logs
		WHERE tenant_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	var logs []*types.PermissionAuditLog
	err = r.db.SelectContext(ctx, &logs, query, tenantID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询审计日志列表失败: %w", err)
	}

	return logs, total, nil
}
