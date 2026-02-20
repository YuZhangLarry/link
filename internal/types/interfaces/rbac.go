package interfaces

import (
	"context"
	"link/internal/types"
)

// ========================================
// 角色仓储接口
// ========================================

// RoleRepository 角色数据访问接口
type RoleRepository interface {
	// Create 创建角色
	Create(ctx context.Context, role *types.Role) error

	// FindByID 根据ID查找角色
	FindByID(ctx context.Context, id int64) (*types.Role, error)

	// FindByTenantID 根据租户ID查找角色列表
	FindByTenantID(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.Role, int64, error)

	// FindByCode 根据租户ID和角色编码查找角色
	FindByCode(ctx context.Context, tenantID int64, code string) (*types.Role, error)

	// Update 更新角色
	Update(ctx context.Context, role *types.Role) error

	// Delete 删除角色
	Delete(ctx context.Context, id int64) error

	// List 分页查询角色列表
	List(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.Role, int64, error)
}

// ========================================
// 用户角色仓储接口
// ========================================

// UserRoleRepository 用户角色关联数据访问接口
type UserRoleRepository interface {
	// Create 创建用户角色关联
	Create(ctx context.Context, userRole *types.UserRole) error

	// FindByUserID 根据用户ID查找其角色
	FindByUserID(ctx context.Context, tenantID, userID int64) (*types.UserRole, error)

	// FindByRoleID 根据角色ID查找拥有该角色的用户列表
	FindByRoleID(ctx context.Context, roleID int64, page, pageSize int) ([]*types.UserRole, int64, error)

	// Update 更新用户角色（实际上应该是删除旧的，创建新的）
	Update(ctx context.Context, userRole *types.UserRole) error

	// Delete 删除用户角色关联
	Delete(ctx context.Context, tenantID, userID int64) error

	// DeleteByRoleID 删除角色的所有用户关联
	DeleteByRoleID(ctx context.Context, roleID int64) error
}

// ========================================
// 权限仓储接口
// ========================================

// PermissionRepository 权限数据访问接口
type PermissionRepository interface {
	// Create 创建权限
	Create(ctx context.Context, permission *types.Permission) error

	// FindByID 根据ID查找权限
	FindByID(ctx context.Context, id int64) (*types.Permission, error)

	// FindAll 查找所有权限
	FindAll(ctx context.Context) ([]*types.Permission, error)

	// FindByResourceType 根据资源类型查找权限
	FindByResourceType(ctx context.Context, resourceType string) ([]*types.Permission, error)

	// FindByRoleID 根据角色ID查找其拥有的所有权限
	FindByRoleID(ctx context.Context, roleID int64) ([]*types.Permission, error)
}

// ========================================
// 角色权限仓储接口
// ========================================

// RolePermissionRepository 角色权限关联数据访问接口
type RolePermissionRepository interface {
	// Create 创建角色权限关联
	Create(ctx context.Context, rolePermission *types.RolePermission) error

	// FindByRoleID 根据角色ID查找权限ID列表
	FindByRoleID(ctx context.Context, roleID int64) ([]int64, error)

	// Delete 删除角色的某个权限
	Delete(ctx context.Context, roleID, permissionID int64) error

	// DeleteByRoleID 删除角色的所有权限
	DeleteByRoleID(ctx context.Context, roleID int64) error

	// BatchCreate 批量创建角色权限关联
	BatchCreate(ctx context.Context, rolePermissions []*types.RolePermission) error
}

// ========================================
// 资源级权限仓储接口
// ========================================

// ResourcePermissionRepository 资源级权限数据访问接口
type ResourcePermissionRepository interface {
	// Create 创建资源级权限
	Create(ctx context.Context, resourcePermission *types.ResourcePermission) error

	// FindByUserID 根据用户ID查找其资源权限
	FindByUserID(ctx context.Context, tenantID, userID int64) ([]*types.ResourcePermission, error)

	// FindByResource 根据资源查找权限列表
	FindByResource(ctx context.Context, tenantID int64, resourceType, resourceID string) ([]*types.ResourcePermission, error)

	// CheckPermission 检查用户对资源是否有指定权限
	CheckPermission(ctx context.Context, tenantID, userID int64, resourceType, resourceID, permissionType string) (bool, error)

	// Delete 删除资源级权限
	Delete(ctx context.Context, id int64) error

	// DeleteByResource 删除资源的所有权限
	DeleteByResource(ctx context.Context, tenantID int64, resourceType, resourceID string) error
}

// ========================================
// 权限审计日志仓储接口
// ========================================

// PermissionAuditLogRepository 权限变更审计日志数据访问接口
type PermissionAuditLogRepository interface {
	// Create 创建审计日志
	Create(ctx context.Context, log *types.PermissionAuditLog) error

	// FindByUserID 根据用户ID查找审计日志
	FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*types.PermissionAuditLog, int64, error)

	// FindByOperatorID 根据操作人ID查找审计日志
	FindByOperatorID(ctx context.Context, operatorID int64, page, pageSize int) ([]*types.PermissionAuditLog, int64, error)

	// FindByTenantID 根据租户ID查找审计日志
	FindByTenantID(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.PermissionAuditLog, int64, error)
}

// ========================================
// 权限服务接口
// ========================================

// PermissionService 权限服务接口
type PermissionService interface {
	// CheckPermission 检查用户是否有指定权限
	CheckPermission(ctx context.Context, tenantID, userID int64, resourceType, action string) (bool, error)

	// GetUserPermissions 获取用户的所有权限列表
	GetUserPermissions(ctx context.Context, tenantID, userID int64) ([]*types.UserPermissionsView, error)

	// GetUserRole 获取用户的角色
	GetUserRole(ctx context.Context, tenantID, userID int64) (*types.UserRolesView, error)

	// AssignRole 给用户分配角色
	AssignRole(ctx context.Context, req *types.AssignRoleRequest, operatorID int64) error

	// RevokeRole 撤销用户角色
	RevokeRole(ctx context.Context, tenantID, userID int64, operatorID int64) error

	// CreateRole 创建角色
	CreateRole(ctx context.Context, tenantID int64, req *types.CreateRoleRequest) (*types.Role, error)

	// UpdateRole 更新角色
	UpdateRole(ctx context.Context, roleID int64, req *types.UpdateRoleRequest) error

	// DeleteRole 删除角色
	DeleteRole(ctx context.Context, roleID int64) error

	// GetRoles 获取角色列表
	GetRoles(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.Role, int64, error)

	// GrantPermissionToRole 给角色分配权限
	GrantPermissionToRole(ctx context.Context, roleID, permissionID int64) error

	// RevokePermissionFromRole 撤销角色的权限
	RevokePermissionFromRole(ctx context.Context, roleID, permissionID int64) error

	// GrantResourcePermission 授予用户对资源的权限
	GrantResourcePermission(ctx context.Context, req *types.ResourcePermission, operatorID int64) error

	// RevokeResourcePermission 撤销用户对资源的权限
	RevokeResourcePermission(ctx context.Context, id int64) error
}
