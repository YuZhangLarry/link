package types

import "time"

// ========================================
// 权限模块 (Permission)
// ========================================

// Permission 权限实体
type Permission struct {
	ID          int64     `json:"id" db:"id"`
	ResourceType string   `json:"resource_type" db:"resource_type"` // kb/session/document/user/role/tenant
	Action      string    `json:"action" db:"action"`               // create/read/update/delete/assign
	Description string    `json:"description" db:"description"`
	IsSystem    bool      `json:"is_system" db:"is_system"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// ========================================
// 角色模块 (Role)
// ========================================

// Role 角色实体
type Role struct {
	ID          int64      `json:"id" db:"id"`
	TenantID    int64      `json:"tenant_id" db:"tenant_id"`
	Name        string     `json:"name" db:"name"`
	Code        string     `json:"code" db:"code"`                 // owner/admin/user
	Description string     `json:"description" db:"description"`
	IsSystem    bool       `json:"is_system" db:"is_system"`
	IsDefault   bool       `json:"is_default" db:"is_default"`
	Level       int        `json:"level" db:"level"`               // 角色层级，数字越大权限越高
	Status      string     `json:"status" db:"status"`             // active/inactive
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// ========================================
// 用户角色关联 (UserRole)
// ========================================

// UserRole 用户角色关联
type UserRole struct {
	ID         int64      `json:"id" db:"id"`
	TenantID   int64      `json:"tenant_id" db:"tenant_id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	RoleID     int64      `json:"role_id" db:"role_id"`
	AssignedBy *int64     `json:"assigned_by,omitempty" db:"assigned_by"`
	AssignedAt time.Time  `json:"assigned_at" db:"assigned_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// TableName 指定表名
func (UserRole) TableName() string {
	return "user_roles"
}

// ========================================
// 角色权限关联 (RolePermission)
// ========================================

// RolePermission 角色权限关联
type RolePermission struct {
	ID          int64     `json:"id" db:"id"`
	RoleID      int64     `json:"role_id" db:"role_id"`
	PermissionID int64    `json:"permission_id" db:"permission_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// TableName 指定表名
func (RolePermission) TableName() string {
	return "role_permissions"
}

// ========================================
// 资源级权限 (ResourcePermission)
// ========================================

// ResourcePermission 资源级权限
type ResourcePermission struct {
	ID             int64      `json:"id" db:"id"`
	TenantID       int64      `json:"tenant_id" db:"tenant_id"`
	UserID         int64      `json:"user_id" db:"user_id"`
	ResourceType   string     `json:"resource_type" db:"resource_type"`   // kb/session/document
	ResourceID     string     `json:"resource_id" db:"resource_id"`       // 资源ID
	PermissionType string     `json:"permission_type" db:"permission_type"` // read/write/delete/admin
	GrantedBy      *int64     `json:"granted_by,omitempty" db:"granted_by"`
	GrantedAt      time.Time  `json:"granted_at" db:"granted_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// TableName 指定表名
func (ResourcePermission) TableName() string {
	return "resource_permissions"
}

// ========================================
// 权限变更审计日志 (PermissionAuditLog)
// ========================================

// PermissionAuditLog 权限变更审计日志
type PermissionAuditLog struct {
	ID            int64              `json:"id" db:"id"`
	TenantID      *int64             `json:"tenant_id,omitempty" db:"tenant_id"`
	UserID        *int64             `json:"user_id,omitempty" db:"user_id"`
	OperatorID    int64              `json:"operator_id" db:"operator_id"`
	OperationType string             `json:"operation_type" db:"operation_type"` // grant_role/revoke_role/modify_role
	TargetType    string             `json:"target_type" db:"target_type"`       // role/resource
	TargetID      string             `json:"target_id" db:"target_id"`
	BeforeValue   string             `json:"before_value,omitempty" db:"before_value"` // JSON
	AfterValue    string             `json:"after_value,omitempty" db:"after_value"`   // JSON
	Reason        string             `json:"reason,omitempty" db:"reason"`
	IPAddress     string             `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt     time.Time          `json:"created_at" db:"created_at"`
}

// TableName 指定表名
func (PermissionAuditLog) TableName() string {
	return "permission_audit_logs"
}

// ========================================
// 用户角色视图 (UserRolesView)
// ========================================

// UserRolesView 用户角色视图数据
type UserRolesView struct {
	UserID      int64      `json:"user_id" db:"user_id"`
	TenantID    int64      `json:"tenant_id" db:"tenant_id"`
	Username    string     `json:"username" db:"username"`
	Email       string     `json:"email" db:"email"`
	Status      int8       `json:"status" db:"status"`
	RoleID      *int64     `json:"role_id,omitempty" db:"role_id"`
	RoleName    *string    `json:"role_name,omitempty" db:"role_name"`
	RoleCode    *string    `json:"role_code,omitempty" db:"role_code"`
	RoleLevel   *int       `json:"role_level,omitempty" db:"role_level"`
	AssignedAt  *time.Time `json:"assigned_at,omitempty" db:"assigned_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// ========================================
// 用户权限视图 (UserPermissionsView)
// ========================================

// UserPermissionsView 用户权限视图数据
type UserPermissionsView struct {
	UserID       int64  `json:"user_id" db:"user_id"`
	TenantID     int64  `json:"tenant_id" db:"tenant_id"`
	Username     string `json:"username" db:"username"`
	ResourceType string `json:"resource_type" db:"resource_type"`
	Action       string `json:"action" db:"action"`
	RoleCode     string `json:"role_code" db:"role_code"`
	RoleLevel    int    `json:"role_level" db:"role_level"`
}

// ========================================
// 请求/响应 DTO
// ========================================

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Code        string `json:"code" binding:"required,min=1,max=50"`
	Description string `json:"description" binding:"max=500"`
	Level       int    `json:"level" binding:"min=0,max=100"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500"`
	Level       int    `json:"level" binding:"min=0,max=100"`
	Status      string `json:"status" binding:"oneof=active inactive"`
}

// AssignRoleRequest 分配角色请求
type AssignRoleRequest struct {
	TenantID  int64  `json:"tenant_id" binding:"required"` // 租户ID
	UserID    int64  `json:"user_id" binding:"required"`
	RoleID    int64  `json:"role_id" binding:"required"`
	Reason    string `json:"reason" binding:"max=500"`
	ExpiresAt *int64 `json:"expires_at,omitempty"` // Unix timestamp
}

// RoleListResponse 角色列表响应
type RoleListResponse struct {
	Roles []*RoleInfo `json:"roles"`
	Total int64       `json:"total"`
}

// RoleInfo 角色信息（不含敏感信息）
type RoleInfo struct {
	ID          int64  `json:"id"`
	TenantID    int64  `json:"tenant_id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
	IsDefault   bool   `json:"is_default"`
	Level       int    `json:"level"`
	Status      string `json:"status"`
	UserCount   int    `json:"user_count,omitempty"` // 拥有此角色的用户数量
}

// UserPermissionCheckRequest 用户权限检查请求
type UserPermissionCheckRequest struct {
	ResourceType string `json:"resource_type" binding:"required"`
	Action       string `json:"action" binding:"required"`
}

// CheckPermissionResponse 权限检查响应
type CheckPermissionResponse struct {
	HasPermission bool   `json:"has_permission"`
	Permissions   []string `json:"permissions,omitempty"` // 用户拥有的所有权限列表
}
