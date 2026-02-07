package repository

import (
	"context"
	"database/sql"
	"fmt"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// userRepository 用户数据访问实现
type userRepository struct {
	db *sql.DB
}

// NewUserRepository 创建用户数据访问实例
func NewUserRepository(db *sql.DB) interfaces.UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *types.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, avatar, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := r.db.ExecContext(ctx, query,
		user.Username, user.Email, user.PasswordHash, user.Avatar, user.Status,
	)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取用户ID失败: %w", err)
	}

	user.ID = id
	return nil
}

// FindByID 根据ID查找用户
func (r *userRepository) FindByID(ctx context.Context, id int64) (*types.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar, status, created_at, updated_at, last_login_at
		FROM users
		WHERE id = ?
	`
	user := &types.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Avatar, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("用户不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*types.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar, status, created_at, updated_at, last_login_at
		FROM users
		WHERE email = ?
	`
	user := &types.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Avatar, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("用户不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return user, nil
}

// FindByUsername 根据用户名查找用户
func (r *userRepository) FindByUsername(ctx context.Context, username string) (*types.User, error) {
	query := `
		SELECT id, username, email, password_hash, avatar, status, created_at, updated_at, last_login_at
		FROM users
		WHERE username = ?
	`
	user := &types.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Avatar, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("用户不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return user, nil
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *types.User) error {
	query := `
		UPDATE users
		SET username = ?, email = ?, avatar = ?, status = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		user.Username, user.Email, user.Avatar, user.Status, user.ID,
	)
	if err != nil {
		return fmt.Errorf("更新用户失败: %w", err)
	}
	return nil
}

// UpdateLastLogin 更新最后登录时间
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("更新最后登录时间失败: %w", err)
	}
	return nil
}

// Delete 删除用户
func (r *userRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}
	return nil
}

// List 分页查询用户列表
func (r *userRepository) List(ctx context.Context, page, pageSize int) ([]*types.User, int64, error) {
	// 查询总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM users`
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `
		SELECT id, username, email, password_hash, avatar, status, created_at, updated_at, last_login_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询用户列表失败: %w", err)
	}
	defer rows.Close()

	var users []*types.User
	for rows.Next() {
		user := &types.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.PasswordHash,
			&user.Avatar, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描用户数据失败: %w", err)
		}
		users = append(users, user)
	}

	return users, total, nil
}

// refreshTokenRepository 刷新Token数据访问实现
type refreshTokenRepository struct {
	db *sql.DB
}

// NewRefreshTokenRepository 创建刷新Token数据访问实例
func NewRefreshTokenRepository(db *sql.DB) interfaces.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// Create 创建刷新Token
func (r *refreshTokenRepository) Create(ctx context.Context, token *types.RefreshTokenEntity) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, created_at)
		VALUES (?, ?, ?, NOW())
	`
	result, err := r.db.ExecContext(ctx, query,
		token.UserID, token.TokenHash, token.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("创建刷新Token失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取Token ID失败: %w", err)
	}

	token.ID = id
	return nil
}

// FindByTokenHash 根据Token哈希查找
func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*types.RefreshTokenEntity, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = ?
	`
	token := &types.RefreshTokenEntity{}
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Token不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询Token失败: %w", err)
	}
	return token, nil
}

// Delete 删除Token
func (r *refreshTokenRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM refresh_tokens WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除Token失败: %w", err)
	}
	return nil
}

// DeleteByUserID 删除用户的所有Token
func (r *refreshTokenRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = ?`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("删除用户Token失败: %w", err)
	}
	return nil
}

// DeleteExpired 删除过期的Token
func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("删除过期Token失败: %w", err)
	}
	return nil
}
