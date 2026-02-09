package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"link/internal/config"
	"link/internal/types"
	"link/internal/types/interfaces"
)

// UserService 用户服务实现
type UserService struct {
	userRepo        interfaces.UserRepository
	refreshTokenRepo interfaces.RefreshTokenRepository
	tenantRepo      interfaces.TenantRepository
	jwtConfig       *config.JWTConfig
}

// NewUserService 创建用户服务实例
func NewUserService(
	userRepo interfaces.UserRepository,
	refreshTokenRepo interfaces.RefreshTokenRepository,
	tenantRepo interfaces.TenantRepository,
	jwtConfig *config.JWTConfig,
) *UserService {
	return &UserService{
		userRepo:        userRepo,
		refreshTokenRepo: refreshTokenRepo,
		tenantRepo:      tenantRepo,
		jwtConfig:       jwtConfig,
	}
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *types.RegisterRequest) (*types.AuthResponse, error) {
	// 如果没有提供租户ID，自动创建一个新租户
	tenantID := req.TenantID
	if tenantID == 0 {
		// 自动创建租户（使用用户名作为租户名）
		tenant := &types.Tenant{
			Name:         req.Username + "_tenant",
			Description:  "自动创建的租户",
			Business:     "personal",
			Status:       "active",
			StorageQuota: 10 * 1024 * 1024 * 1024, // 10GB
			StorageUsed:  0,
		}
		err := s.tenantRepo.Create(ctx, tenant)
		if err != nil {
			return nil, fmt.Errorf("自动创建租户失败: %w", err)
		}
		tenantID = tenant.ID
	}

	// 检查邮箱是否已存在（在租户内）
	existingUser, err := s.userRepo.FindByEmail(ctx, tenantID, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("邮箱已被注册")
	}

	// 检查用户名是否已存在（在租户内）
	existingUser, err = s.userRepo.FindByUsername(ctx, tenantID, req.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("用户名已被使用")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 创建用户（包含租户ID）
	user := &types.User{
		TenantID:     tenantID,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Avatar:       "",
		Status:       1, // 默认启用
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败 (tenantID=%d, username=%s): %w", tenantID, req.Username, err)
	}

	// 生成Token（包含租户ID）
	accessToken, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	// 存储刷新Token
	err = s.saveRefreshToken(ctx, user.ID, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("保存刷新Token失败: %w", err)
	}

	return &types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TenantID:     user.TenantID,
		User: types.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			TenantID:  user.TenantID,
		},
	}, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, req *types.LoginRequest) (*types.AuthResponse, error) {
	var user *types.User
	var err error

	// 如果没有提供租户ID，则仅通过邮箱查找用户（自动获取租户ID）
	if req.TenantID == 0 {
		user, err = s.userRepo.FindByEmailOnly(ctx, req.Email)
	} else {
		user, err = s.userRepo.FindByEmail(ctx, req.TenantID, req.Email)
	}

	if err != nil {
		return nil, errors.New("邮箱或密码错误")
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("邮箱或密码错误")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	// 更新最后登录时间
	err = s.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("更新登录时间失败: %w", err)
	}

	// 生成Token（包含租户ID）
	accessToken, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	// 存储刷新Token
	err = s.saveRefreshToken(ctx, user.ID, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("保存刷新Token失败: %w", err)
	}

	return &types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TenantID:     user.TenantID,
		User: types.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			TenantID:  user.TenantID,
		},
	}, nil
}

// Logout 用户登出
func (s *UserService) Logout(ctx context.Context, userID int64) error {
	// 删除用户的所有刷新Token
	err := s.refreshTokenRepo.DeleteByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("登出失败: %w", err)
	}
	return nil
}

// RefreshToken 刷新Token
func (s *UserService) RefreshToken(ctx context.Context, req *types.RefreshTokenRequest) (*types.AuthResponse, error) {
	// 验证刷新Token
	claims, err := s.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("无效的刷新Token")
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("Token类型错误")
	}

	// 获取用户信息
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	// 验证刷新Token是否在数据库中
	tokenHash := s.hashToken(req.RefreshToken)
	_, err = s.refreshTokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.New("刷新Token不存在或已失效")
	}

	// 生成新的Token（包含租户ID）
	accessToken, refreshToken, expiresAt, err := s.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("生成Token失败: %w", err)
	}

	// 删除旧的刷新Token
	err = s.refreshTokenRepo.DeleteByUserID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("删除旧Token失败: %w", err)
	}

	// 存储新的刷新Token
	err = s.saveRefreshToken(ctx, user.ID, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("保存刷新Token失败: %w", err)
	}

	return &types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TenantID:     user.TenantID,
		User: types.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			TenantID:  user.TenantID,
		},
	}, nil
}

// GetUserByID 根据ID获取用户信息
func (s *UserService) GetUserByID(ctx context.Context, userID int64) (*types.UserInfo, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &types.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		TenantID:  user.TenantID,
	}, nil
}

// ValidateToken 验证Token
func (s *UserService) ValidateToken(tokenString string) (*types.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("无效的签名方法: %v", token.Header["alg"])
		}
		return []byte(s.jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenClaims := &types.TokenClaims{
			UserID:    int64(claims["user_id"].(float64)),
			Username:  claims["username"].(string),
			Email:     claims["email"].(string),
			TokenType: claims["token_type"].(string),
		}
		// 尝试获取 tenant_id（如果存在）
		if tenantID, ok := claims["tenant_id"].(float64); ok {
			tokenClaims.TenantID = int64(tenantID)
		}
		return tokenClaims, nil
	}

	return nil, errors.New("无效的Token")
}

// generateTokens 生成访问令牌和刷新令牌
func (s *UserService) generateTokens(user *types.User) (string, string, int64, error) {
	now := time.Now()
	accessExpiresAt := now.Add(time.Duration(s.jwtConfig.AccessTokenExpire) * time.Second)
	refreshExpiresAt := now.Add(time.Duration(s.jwtConfig.RefreshTokenExpire) * time.Second)

	// 生成访问Token（包含租户ID）
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"tenant_id":  user.TenantID,
		"username":   user.Username,
		"email":      user.Email,
		"token_type": "access",
		"exp":        accessExpiresAt.Unix(),
		"iat":        now.Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(s.jwtConfig.Secret))
	if err != nil {
		return "", "", 0, err
	}

	// 生成刷新Token（包含租户ID）
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID,
		"tenant_id":  user.TenantID,
		"username":   user.Username,
		"email":      user.Email,
		"token_type": "refresh",
		"exp":        refreshExpiresAt.Unix(),
		"iat":        now.Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtConfig.Secret))
	if err != nil {
		return "", "", 0, err
	}

	return accessTokenString, refreshTokenString, accessExpiresAt.Unix(), nil
}

// saveRefreshToken 保存刷新Token到数据库
func (s *UserService) saveRefreshToken(ctx context.Context, userID int64, token string) error {
	tokenHash := s.hashToken(token)
	expiresAt := time.Now().Add(time.Duration(s.jwtConfig.RefreshTokenExpire) * time.Second)

	refreshToken := &types.RefreshTokenEntity{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	return s.refreshTokenRepo.Create(ctx, refreshToken)
}

// hashToken 对Token进行哈希
func (s *UserService) hashToken(token string) string {
	h := sha256.New()
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}
