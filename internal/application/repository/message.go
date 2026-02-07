package repository

import (
	"context"
	"database/sql"
	"fmt"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// messageRepository 消息数据访问实现
type messageRepository struct {
	db *sql.DB
}

// NewMessageRepository 创建消息数据访问实例
func NewMessageRepository(db *sql.DB) interfaces.MessageRepository {
	return &messageRepository{db: db}
}

// Create 创建消息
func (r *messageRepository) Create(ctx context.Context, message *types.MessageEntity) error {
	// 处理 tool_calls：如果是空字符串，使用 NULL
	var toolCalls interface{} = message.ToolCalls
	if toolCalls == "" {
		toolCalls = nil
	}

	query := `
		INSERT INTO messages (chat_id, role, content, tool_calls, token_count, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
	`
	result, err := r.db.ExecContext(ctx, query,
		message.ChatID, message.Role, message.Content, toolCalls, message.TokenCount,
	)
	if err != nil {
		return fmt.Errorf("创建消息失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取消息ID失败: %w", err)
	}

	message.ID = id
	return nil
}

// FindByID 根据ID查找消息
func (r *messageRepository) FindByID(ctx context.Context, id int64) (*types.MessageEntity, error) {
	query := `
		SELECT id, chat_id, role, content, tool_calls, token_count, created_at
		FROM messages
		WHERE id = ?
	`
	message := &types.MessageEntity{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&message.ID, &message.ChatID, &message.Role, &message.Content,
		&message.ToolCalls, &message.TokenCount, &message.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("消息不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("查询消息失败: %w", err)
	}
	return message, nil
}

// FindByChatID 根据对话ID查找消息列表
func (r *messageRepository) FindByChatID(ctx context.Context, chatID int64, page, pageSize int) ([]*types.MessageEntity, int64, error) {
	// 查询总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM messages WHERE chat_id = ?`
	err := r.db.QueryRowContext(ctx, countQuery, chatID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询消息总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `
		SELECT id, chat_id, role, content, tool_calls, token_count, created_at
		FROM messages
		WHERE chat_id = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, chatID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询消息列表失败: %w", err)
	}
	defer rows.Close()

	var messages []*types.MessageEntity
	for rows.Next() {
		message := &types.MessageEntity{}
		err := rows.Scan(
			&message.ID, &message.ChatID, &message.Role, &message.Content,
			&message.ToolCalls, &message.TokenCount, &message.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描消息数据失败: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, total, nil
}

// FindByChatIDAndRole 根据对话ID和角色查找消息
func (r *messageRepository) FindByChatIDAndRole(ctx context.Context, chatID int64, role string, page, pageSize int) ([]*types.MessageEntity, int64, error) {
	// 查询总数
	var total int64
	countQuery := `SELECT COUNT(*) FROM messages WHERE chat_id = ? AND role = ?`
	err := r.db.QueryRowContext(ctx, countQuery, chatID, role).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("查询消息总数失败: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	query := `
		SELECT id, chat_id, role, content, tool_calls, token_count, created_at
		FROM messages
		WHERE chat_id = ? AND role = ?
		ORDER BY created_at ASC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, chatID, role, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("查询消息列表失败: %w", err)
	}
	defer rows.Close()

	var messages []*types.MessageEntity
	for rows.Next() {
		message := &types.MessageEntity{}
		err := rows.Scan(
			&message.ID, &message.ChatID, &message.Role, &message.Content,
			&message.ToolCalls, &message.TokenCount, &message.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("扫描消息数据失败: %w", err)
		}
		messages = append(messages, message)
	}

	return messages, total, nil
}

// Update 更新消息
func (r *messageRepository) Update(ctx context.Context, message *types.MessageEntity) error {
	query := `
		UPDATE messages
		SET content = ?, tool_calls = ?, token_count = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query,
		message.Content, message.ToolCalls, message.TokenCount, message.ID,
	)
	if err != nil {
		return fmt.Errorf("更新消息失败: %w", err)
	}
	return nil
}

// Delete 删除消息
func (r *messageRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM messages WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除消息失败: %w", err)
	}
	return nil
}

// DeleteByChatID 删除对话的所有消息
func (r *messageRepository) DeleteByChatID(ctx context.Context, chatID int64) error {
	query := `DELETE FROM messages WHERE chat_id = ?`
	_, err := r.db.ExecContext(ctx, query, chatID)
	if err != nil {
		return fmt.Errorf("删除对话消息失败: %w", err)
	}
	return nil
}

// CountByChatID 统计对话的消息数量
func (r *messageRepository) CountByChatID(ctx context.Context, chatID int64) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM messages WHERE chat_id = ?`
	err := r.db.QueryRowContext(ctx, query, chatID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("统计消息数量失败: %w", err)
	}
	return count, nil
}
