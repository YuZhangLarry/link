package container

import (
	"link/internal/application/repository"
	"link/internal/application/service"
	"link/internal/types/interfaces"

	"github.com/jmoiron/sqlx"
)

// ========================================
// 权限系统仓储和服务
// ========================================

var (
	// 权限仓储
	roleRepo            repository.RoleRepository
	userRoleRepo        repository.UserRoleRepository
	permissionRepo      repository.PermissionRepository
	rolePermissionRepo  repository.RolePermissionRepository
	resourcePermRepo    repository.ResourcePermissionRepository
	auditLogRepo        repository.PermissionAuditLogRepository

	// 权限服务
	permissionService interfaces.PermissionService
)

// ========================================
// 初始化权限系统
// ========================================

// InitPermissionSystem 初始化权限系统（使用 sqlx）
func InitPermissionSystem() error {
	db := GetSQLDB()
	if db == nil {
		return nil // 如果没有数据库连接，跳过初始化
	}

	// 将 sql.DB 转换为 sqlx.DB
	sqlxDB := sqlx.NewDb(db, "mysql")

	// 初始化仓储
	roleRepo = repository.NewRoleRepository(sqlxDB)
	userRoleRepo = repository.NewUserRoleRepository(sqlxDB)
	permissionRepo = repository.NewPermissionRepository(sqlxDB)
	rolePermissionRepo = repository.NewRolePermissionRepository(sqlxDB)
	resourcePermRepo = repository.NewResourcePermissionRepository(sqlxDB)
	auditLogRepo = repository.NewPermissionAuditLogRepository(sqlxDB)

	// 初始化服务
	permissionService = service.NewPermissionService(
		roleRepo,
		userRoleRepo,
		permissionRepo,
		rolePermissionRepo,
		resourcePermRepo,
		auditLogRepo,
	)

	return nil
}

// ========================================
// Getter 方法
// ========================================

// GetRoleRepository 获取角色仓储
func GetRoleRepository() repository.RoleRepository {
	return roleRepo
}

// GetUserRoleRepository 获取用户角色仓储
func GetUserRoleRepository() repository.UserRoleRepository {
	return userRoleRepo
}

// GetPermissionRepository 获取权限仓储
func GetPermissionRepository() repository.PermissionRepository {
	return permissionRepo
}

// GetRolePermissionRepository 获取角色权限仓储
func GetRolePermissionRepository() repository.RolePermissionRepository {
	return rolePermissionRepo
}

// GetResourcePermissionRepository 获取资源权限仓储
func GetResourcePermissionRepository() repository.ResourcePermissionRepository {
	return resourcePermRepo
}

// GetPermissionAuditLogRepository 获取权限审计日志仓储
func GetPermissionAuditLogRepository() repository.PermissionAuditLogRepository {
	return auditLogRepo
}

// GetPermissionService 获取权限服务
func GetPermissionService() interfaces.PermissionService {
	return permissionService
}
