package types

import "time"

// User 用户实体
type User struct {
	ID                int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID          int64      `json:"tenant_id" gorm:"not null;index:idx_tenant_id"` // 租户ID
	Username          string     `json:"username" gorm:"type:varchar(50);not null;index:idx_tenant_username,priority:1"`
	Email             string     `json:"email" gorm:"type:varchar(100);not null;index:idx_tenant_email,priority:1"`
	PasswordHash      string     `json:"-" gorm:"type:varchar(255);not null"`
	Avatar            string     `json:"avatar" gorm:"type:varchar(500)"`
	Status            int8       `json:"status" gorm:"type:tinyint;default:1;index:idx_status"` // 0=禁用, 1=正常
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	LastLoginAt       *time.Time `json:"last_login_at,omitempty" gorm:"index"`
	DeletedAt         *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	TenantID int64  `json:"tenant_id"` // 租户ID，可选（为空时自动创建）
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=100"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	TenantID int64  `json:"tenant_id"` // 租户ID，可选（为空时自动查找）
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    int64     `json:"expires_at"`
	User         UserInfo  `json:"user"`
	TenantID     int64     `json:"tenant_id,omitempty"`
}

// UserInfo 用户信息（不含敏感信息）
type UserInfo struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	TenantID  int64     `json:"tenant_id,omitempty"`
}

// TokenClaims JWT Token声明
type TokenClaims struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	TenantID  int64  `json:"tenant_id,omitempty"`
	TokenType string `json:"token_type"` // access or refresh
}

// RefreshTokenEntity 刷新Token实体（用于存储在数据库）
type RefreshTokenEntity struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int64     `json:"user_id" gorm:"not null;index"`
	TokenHash string    `json:"-" gorm:"type:varchar(64);not null;uniqueIndex"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}
