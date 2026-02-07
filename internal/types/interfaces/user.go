package interfaces

import (
	"context"
	"link/internal/types"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *types.User) error

	// FindByID 根据ID查找用户
	FindByID(ctx context.Context, id int64) (*types.User, error)

	// FindByEmail 根据邮箱查找用户
	FindByEmail(ctx context.Context, email string) (*types.User, error)

	// FindByUsername 根据用户名查找用户
	FindByUsername(ctx context.Context, username string) (*types.User, error)

	// Update 更新用户
	Update(ctx context.Context, user *types.User) error

	// UpdateLastLogin 更新最后登录时间
	UpdateLastLogin(ctx context.Context, userID int64) error

	// Delete 删除用户
	Delete(ctx context.Context, id int64) error

	// List 分页查询用户列表
	List(ctx context.Context, page, pageSize int) ([]*types.User, int64, error)
}

// RefreshTokenRepository 刷新Token数据访问接口
type RefreshTokenRepository interface {
	// Create 创建刷新Token
	Create(ctx context.Context, token *types.RefreshTokenEntity) error

	// FindByTokenHash 根据Token哈希查找
	FindByTokenHash(ctx context.Context, tokenHash string) (*types.RefreshTokenEntity, error)

	// Delete 删除Token
	Delete(ctx context.Context, id int64) error

	// DeleteByUserID 删除用户的所有Token
	DeleteByUserID(ctx context.Context, userID int64) error

	// DeleteExpired 删除过期的Token
	DeleteExpired(ctx context.Context) error
}

// UserService 用户服务接口
type UserService interface {
	// Register 用户注册
	Register(ctx context.Context, req *types.RegisterRequest) (*types.AuthResponse, error)

	// Login 用户登录
	Login(ctx context.Context, req *types.LoginRequest) (*types.AuthResponse, error)

	// Logout 用户登出
	Logout(ctx context.Context, userID int64) error

	// RefreshToken 刷新Token
	RefreshToken(ctx context.Context, req *types.RefreshTokenRequest) (*types.AuthResponse, error)

	// GetUserByID 根据ID获取用户信息
	GetUserByID(ctx context.Context, userID int64) (*types.UserInfo, error)

	// ValidateToken 验证Token
	ValidateToken(tokenString string) (*types.TokenClaims, error)
}
