package repository

import (
	"context"
	"database/sql"
	"fmt"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// sessionRepository 会话数据访问实现
type sessionRepository struct {
	db *sql.DB
}

// NewSessionRepository 创建会话数据访问实例
func NewSessionRepository(db *sql.DB) interfaces.SessionRepository {
	return &sessionRepository{db: db}
}

// Create 创建会话
func (r *sessionRepository) Create(ctx context.Context, session *types.SessionEntity) error {
	query := `
		INSERT INTO sessions (user_id, kb_id, title, description, model, max_rounds,
			enable_rewrite, keyword_threshold, vector_threshold, message_count, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 1, NOW(), NOW())
	`
	result, err := r.db.ExecContext(ctx, query,
		session.UserID, session.KBID, session.Title, session.Description, session.Model,
		session.MaxRounds, session.EnableRewrite, session.KeywordThreshold, session.VectorThreshold,
	)
	if err != nil {
		return fmt.Errorf("创建会话失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取会话ID失败: %w", err)
	}

	session.ID = id
	return nil
}

// FindByID 根据ID查找会话
func (r *sessionRepository) FindByID(ctx context.Context, id int64) (*types.SessionEntity, error) {
	query := `
		SELECT id, user_id, kb_id, title, description, model, max_rounds, enable_rewrite,
			keyword_threshold, vector_threshold, message_count, status, created_at, updated_at, deleted_at
		FROM sessions
		WHERE id = ? AND deleted_at IS NULL
	`
	session := &types.SessionEntity{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID, &session.UserID, &session.KBID, &session.Title, &session.Description,
		&session.Model, &session.MaxRounds, &session.EnableRewrite, &session.KeywordThreshold,
		&session.VectorThreshold, &session.MessageCount, &session.Status,
		&session.CreatedAt, &session.UpdatedAt, &session.DeletedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("会话不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询会话失败: %w", err)
	}
	return session, nil
}

// FindByUserID 根据用户ID查找会话列表
func (r *sessionRepository) FindByUserID(ctx context.Context, userID int64, page, pageSize int) ([]*types.SessionEntity, int64, error) {
	// 查询总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM sessions WHERE user_id = ? AND deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `
		SELECT id, user_id, kb_id, title, description, model, max_rounds, enable_rewrite,
			keyword_threshold, vector_threshold, message_count, status, created_at, updated_at, deleted_at
		FROM sessions
		WHERE user_id = ? AND deleted_at IS NULL
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话列表失败: %w", err)
	}
	defer rows.Close()

	var sessions []*types.SessionEntity
	for rows.Next() {
		session := &types.SessionEntity{}
		err := rows.Scan(
			&session.ID, &session.UserID, &session.KBID, &session.Title, &session.Description,
			&session.Model, &session.MaxRounds, &session.EnableRewrite, &session.KeywordThreshold,
			&session.VectorThreshold, &session.MessageCount, &session.Status,
			&session.CreatedAt, &session.UpdatedAt, &session.DeletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描会话数据失败: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, total, nil
}

// FindByUserIDAndStatus 根据用户ID和状态查找会话
func (r *sessionRepository) FindByUserIDAndStatus(ctx context.Context, userID int64, status int8, page, pageSize int) ([]*types.SessionEntity, int64, error) {
	// 查询总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM sessions WHERE user_id = ? AND status = ? AND deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, countQuery, userID, status).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `
		SELECT id, user_id, kb_id, title, description, model, max_rounds, enable_rewrite,
			keyword_threshold, vector_threshold, message_count, status, created_at, updated_at, deleted_at
		FROM sessions
		WHERE user_id = ? AND status = ? AND deleted_at IS NULL
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, userID, status, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询会话列表失败: %w", err)
	}
	defer rows.Close()

	var sessions []*types.SessionEntity
	for rows.Next() {
		session := &types.SessionEntity{}
		err := rows.Scan(
			&session.ID, &session.UserID, &session.KBID, &session.Title, &session.Description,
			&session.Model, &session.MaxRounds, &session.EnableRewrite, &session.KeywordThreshold,
			&session.VectorThreshold, &session.MessageCount, &session.Status,
			&session.CreatedAt, &session.UpdatedAt, &session.DeletedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描会话数据失败: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, total, nil
}

// Update 更新会话
func (r *sessionRepository) Update(ctx context.Context, session *types.SessionEntity) error {
	query := `
		UPDATE sessions
		SET title = ?, description = ?, model = ?, max_rounds = ?, enable_rewrite = ?,
			keyword_threshold = ?, vector_threshold = ?, status = ?, updated_at = NOW()
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		session.Title, session.Description, session.Model, session.MaxRounds,
		session.EnableRewrite, session.KeywordThreshold, session.VectorThreshold,
		session.Status, session.ID,
	)
	if err != nil {
		return fmt.Errorf("更新会话失败: %w", err)
	}
	return nil
}

// UpdateMessageCount 更新会话消息数量
func (r *sessionRepository) UpdateMessageCount(ctx context.Context, sessionID int64) error {
	query := `
		UPDATE sessions s
		SET message_count = (
			SELECT COUNT(*)
			FROM messages
			WHERE chat_id = ?
		), updated_at = NOW()
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, sessionID, sessionID)
	if err != nil {
		return fmt.Errorf("更新消息数量失败: %w", err)
	}
	return nil
}

// Delete 删除会话（软删除）
func (r *sessionRepository) Delete(ctx context.Context, id int64) error {
	query := `UPDATE sessions SET deleted_at = NOW(), updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除会话失败: %w", err)
	}
	return nil
}

// HardDelete 硬删除会话
func (r *sessionRepository) HardDelete(ctx context.Context, id int64) error {
	query := `DELETE FROM sessions WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("硬删除会话失败: %w", err)
	}
	return nil
}

// CountByUserID 统计用户的会话数量
func (r *sessionRepository) CountByUserID(ctx context.Context, userID int64) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM sessions WHERE user_id = ? AND deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("统计会话数量失败: %w", err)
	}
	return count, nil
}

// IncrementMessageCount 增加消息计数
func (r *sessionRepository) IncrementMessageCount(ctx context.Context, sessionID int64) error {
	query := `UPDATE sessions SET message_count = message_count + 1, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("增加消息计数失败: %w", err)
	}
	return nil
}
