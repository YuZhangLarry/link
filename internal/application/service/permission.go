package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// ========================================
// 权限服务实现
// ========================================

// permissionService 权限服务实现
type permissionService struct {
	roleRepo            interfaces.RoleRepository
	userRoleRepo        interfaces.UserRoleRepository
	permissionRepo      interfaces.PermissionRepository
	rolePermissionRepo  interfaces.RolePermissionRepository
	resourcePermRepo    interfaces.ResourcePermissionRepository
	auditLogRepo        interfaces.PermissionAuditLogRepository
}

// NewPermissionService 创建权限服务
func NewPermissionService(
	roleRepo interfaces.RoleRepository,
	userRoleRepo interfaces.UserRoleRepository,
	permissionRepo interfaces.PermissionRepository,
	rolePermissionRepo interfaces.RolePermissionRepository,
	resourcePermRepo interfaces.ResourcePermissionRepository,
	auditLogRepo interfaces.PermissionAuditLogRepository,
) interfaces.PermissionService {
	return &permissionService{
		roleRepo:           roleRepo,
		userRoleRepo:       userRoleRepo,
		permissionRepo:     permissionRepo,
		rolePermissionRepo: rolePermissionRepo,
		resourcePermRepo:   resourcePermRepo,
		auditLogRepo:       auditLogRepo,
	}
}

// CheckPermission 检查用户是否有指定权限
func (s *permissionService) CheckPermission(ctx context.Context, tenantID, userID int64, resourceType, action string) (bool, error) {
	// 首先获取用户的角色
	userRole, err := s.userRoleRepo.FindByUserID(ctx, tenantID, userID)
	if err != nil {
		// 如果没有角色，则没有权限
		return false, nil
	}

	// 获取角色的权限
	permissions, err := s.permissionRepo.FindByRoleID(ctx, userRole.RoleID)
	if err != nil {
		return false, fmt.Errorf("查询角色权限失败: %w", err)
	}

	// 检查是否有匹配的权限
	for _, perm := range permissions {
		if perm.ResourceType == resourceType && perm.Action == action {
			return true, nil
		}
	}

	return false, nil
}

// GetUserPermissions 获取用户的所有权限列表
func (s *permissionService) GetUserPermissions(ctx context.Context, tenantID, userID int64) ([]*types.UserPermissionsView, error) {
	// 获取用户角色
	userRole, err := s.userRoleRepo.FindByUserID(ctx, tenantID, userID)
	if err != nil {
		return []*types.UserPermissionsView{}, nil
	}

	// 获取角色的权限
	permissions, err := s.permissionRepo.FindByRoleID(ctx, userRole.RoleID)
	if err != nil {
		return []*types.UserPermissionsView{}, nil
	}

	// 获取角色信息
	role, err := s.roleRepo.FindByID(ctx, userRole.RoleID)
	if err != nil {
		return []*types.UserPermissionsView{}, nil
	}

	// 获取用户信息（用于视图）
	// 这里简化处理，实际应该从 userRepo 获取
	result := make([]*types.UserPermissionsView, 0, len(permissions))
	for _, perm := range permissions {
		result = append(result, &types.UserPermissionsView{
			UserID:       userID,
			TenantID:     tenantID,
			Username:     "", // 需要从 userRepo 获取
			ResourceType: perm.ResourceType,
			Action:       perm.Action,
			RoleCode:     role.Code,
			RoleLevel:    role.Level,
		})
	}

	return result, nil
}

// GetUserRole 获取用户的角色
func (s *permissionService) GetUserRole(ctx context.Context, tenantID, userID int64) (*types.UserRolesView, error) {
	userRole, err := s.userRoleRepo.FindByUserID(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	role, err := s.roleRepo.FindByID(ctx, userRole.RoleID)
	if err != nil {
		return nil, err
	}

	return &types.UserRolesView{
		UserID:     userID,
		TenantID:   tenantID,
		RoleID:     &role.ID,
		RoleName:   &role.Name,
		RoleCode:   &role.Code,
		RoleLevel:  &role.Level,
		AssignedAt: &userRole.AssignedAt,
		ExpiresAt:  userRole.ExpiresAt,
	}, nil
}

// AssignRole 给用户分配角色
func (s *permissionService) AssignRole(ctx context.Context, req *types.AssignRoleRequest, operatorID int64) error {
	// 获取上下文中的租户ID和用户ID
	// 这里需要从context中获取，简化处理
	tenantID := req.TenantID // 假设从请求中获取
	userID := req.UserID

	// 获取旧角色（用于审计）
	oldRole, err := s.userRoleRepo.FindByUserID(ctx, tenantID, userID)
	var oldRoleID *int64
	if err == nil && oldRole != nil {
		oldRoleID = &oldRole.RoleID
	}

	// 创建新的用户角色关联
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t := time.Unix(*req.ExpiresAt, 0)
		expiresAt = &t
	}

	userRole := &types.UserRole{
		TenantID:   tenantID,
		UserID:     userID,
		RoleID:     req.RoleID,
		AssignedBy: &operatorID,
		AssignedAt: time.Now(),
		ExpiresAt:  expiresAt,
	}

	// 更新用户角色（实际上是删除旧的，创建新的）
	err = s.userRoleRepo.Update(ctx, userRole)
	if err != nil {
		return fmt.Errorf("分配角色失败: %w", err)
	}

	// 记录审计日志
	beforeValue, _ := json.Marshal(map[string]interface{}{"role_id": oldRoleID})
	afterValue, _ := json.Marshal(map[string]interface{}{"role_id": req.RoleID})

	auditLog := &types.PermissionAuditLog{
		TenantID:      &tenantID,
		UserID:        &userID,
		OperatorID:    operatorID,
		OperationType: "modify_role",
		TargetType:    "role",
		TargetID:      fmt.Sprintf("%d", req.RoleID),
		BeforeValue:   string(beforeValue),
		AfterValue:    string(afterValue),
		Reason:        req.Reason,
		CreatedAt:     time.Now(),
	}

	_ = s.auditLogRepo.Create(ctx, auditLog)

	return nil
}

// RevokeRole 撤销用户角色
func (s *permissionService) RevokeRole(ctx context.Context, tenantID, userID int64, operatorID int64) error {
	err := s.userRoleRepo.Delete(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("撤销角色失败: %w", err)
	}

	// 记录审计日志
	auditLog := &types.PermissionAuditLog{
		TenantID:      &tenantID,
		UserID:        &userID,
		OperatorID:    operatorID,
		OperationType: "revoke_role",
		TargetType:    "role",
		TargetID:      fmt.Sprintf("%d", userID),
		CreatedAt:     time.Now(),
	}

	_ = s.auditLogRepo.Create(ctx, auditLog)

	return nil
}

// CreateRole 创建角色
func (s *permissionService) CreateRole(ctx context.Context, tenantID int64, req *types.CreateRoleRequest) (*types.Role, error) {
	role := &types.Role{
		TenantID:    tenantID,
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
		Level:       req.Level,
		IsSystem:    false,
		IsDefault:   false,
		Status:      "active",
	}

	err := s.roleRepo.Create(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("创建角色失败: %w", err)
	}

	return role, nil
}

// UpdateRole 更新角色
func (s *permissionService) UpdateRole(ctx context.Context, roleID int64, req *types.UpdateRoleRequest) error {
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("角色不存在: %w", err)
	}

	// 系统角色不允许修改
	if role.IsSystem {
		return fmt.Errorf("系统角色不允许修改")
	}

	role.Name = req.Name
	role.Description = req.Description
	role.Level = req.Level
	role.Status = req.Status

	err = s.roleRepo.Update(ctx, role)
	if err != nil {
		return fmt.Errorf("更新角色失败: %w", err)
	}

	return nil
}

// DeleteRole 删除角色
func (s *permissionService) DeleteRole(ctx context.Context, roleID int64) error {
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("角色不存在: %w", err)
	}

	// 系统角色不允许删除
	if role.IsSystem {
		return fmt.Errorf("系统角色不允许删除")
	}

	err = s.roleRepo.Delete(ctx, roleID)
	if err != nil {
		return fmt.Errorf("删除角色失败: %w", err)
	}

	return nil
}

// GetRoles 获取角色列表
func (s *permissionService) GetRoles(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.Role, int64, error) {
	return s.roleRepo.FindByTenantID(ctx, tenantID, page, pageSize)
}

// GrantPermissionToRole 给角色分配权限
func (s *permissionService) GrantPermissionToRole(ctx context.Context, roleID, permissionID int64) error {
	rolePermission := &types.RolePermission{
		RoleID:      roleID,
		PermissionID: permissionID,
	}

	err := s.rolePermissionRepo.Create(ctx, rolePermission)
	if err != nil {
		return fmt.Errorf("分配权限失败: %w", err)
	}

	return nil
}

// RevokePermissionFromRole 撤销角色的权限
func (s *permissionService) RevokePermissionFromRole(ctx context.Context, roleID, permissionID int64) error {
	err := s.rolePermissionRepo.Delete(ctx, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("撤销权限失败: %w", err)
	}

	return nil
}

// GrantResourcePermission 授予用户对资源的权限
func (s *permissionService) GrantResourcePermission(ctx context.Context, req *types.ResourcePermission, operatorID int64) error {
	err := s.resourcePermRepo.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("授予资源权限失败: %w", err)
	}

	// 记录审计日志
	auditLog := &types.PermissionAuditLog{
		TenantID:      &req.TenantID,
		UserID:        &req.UserID,
		OperatorID:    operatorID,
		OperationType: "grant_resource",
		TargetType:    "resource",
		TargetID:      fmt.Sprintf("%s:%s", req.ResourceType, req.ResourceID),
		CreatedAt:     time.Now(),
	}

	_ = s.auditLogRepo.Create(ctx, auditLog)

	return nil
}

// RevokeResourcePermission 撤销用户对资源的权限
func (s *permissionService) RevokeResourcePermission(ctx context.Context, id int64) error {
	err := s.resourcePermRepo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("撤销资源权限失败: %w", err)
	}

	return nil
}
