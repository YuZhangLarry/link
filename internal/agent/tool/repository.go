package tool

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"link/internal/models"
)

// Repository 工具执行记录仓库
type Repository struct {
	db *sql.DB
}

// NewRepository 创建工具仓库
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// SaveExecution 保存工具执行记录
func (r *Repository) SaveExecution(ctx context.Context, messageID int64, execResult *ToolExecResult) error {
	// 查找或创建工具ID
	toolID, err := r.getOrCreateToolID(ctx, execResult.ToolName)
	if err != nil {
		return fmt.Errorf("failed to get tool id: %w", err)
	}

	var outputData sql.NullString
	if execResult.Success {
		outputData = sql.NullString{String: execResult.Data, Valid: true}
	}

	status := "success"
	var errorMessage sql.NullString
	if !execResult.Success {
		status = "failed"
		if execResult.Error != nil {
			errorMessage = sql.NullString{String: execResult.Error.Error(), Valid: true}
		}
	}

	query := `
		INSERT INTO tool_executions (message_id, tool_id, input_params, output_data, status, duration_ms, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		messageID,
		toolID,
		"{}", // input_params 需要从外部传入
		outputData,
		status,
		execResult.DurationMs,
		errorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to save execution: %w", err)
	}

	return nil
}

// SaveExecutionWithParams 保存工具执行记录（带参数）
func (r *Repository) SaveExecutionWithParams(
	ctx context.Context,
	messageID int64,
	toolName string,
	inputParams map[string]interface{},
	execResult *ToolExecResult,
) error {
	// 查找或创建工具ID
	toolID, err := r.getOrCreateToolID(ctx, toolName)
	if err != nil {
		return fmt.Errorf("failed to get tool id: %w", err)
	}

	// 序列化输入参数
	inputJSON, err := json.Marshal(inputParams)
	if err != nil {
		return fmt.Errorf("failed to marshal input params: %w", err)
	}

	var outputData sql.NullString
	if execResult.Success {
		outputData = sql.NullString{String: execResult.Data, Valid: true}
	}

	status := "success"
	var errorMessage sql.NullString
	if !execResult.Success {
		status = "failed"
		if execResult.Error != nil {
			errorMessage = sql.NullString{String: execResult.Error.Error(), Valid: true}
		}
	}

	query := `
		INSERT INTO tool_executions (message_id, tool_id, input_params, output_data, status, duration_ms, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(ctx, query,
		messageID,
		toolID,
		string(inputJSON),
		outputData,
		status,
		execResult.DurationMs,
		errorMessage,
	)

	if err != nil {
		return fmt.Errorf("failed to save execution: %w", err)
	}

	return nil
}

// GetExecutionsByMessage 获取消息的所有工具执行记录
func (r *Repository) GetExecutionsByMessage(ctx context.Context, messageID int64) ([]*models.ToolExecution, error) {
	query := `
		SELECT id, message_id, tool_id, input_params, output_data,
		       status, duration_ms, error_message, created_at
		FROM tool_executions
		WHERE message_id = ?
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}
	defer rows.Close()

	executions := make([]*models.ToolExecution, 0)
	for rows.Next() {
		var exec models.ToolExecution
		var inputParams, outputData, errorMessage sql.NullString

		err := rows.Scan(
			&exec.ID,
			&exec.MessageID,
			&exec.ToolID,
			&inputParams,
			&outputData,
			&exec.Status,
			&exec.DurationMs,
			&errorMessage,
			&exec.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution: %w", err)
		}

		if inputParams.Valid {
			exec.InputParams = inputParams.String
		}
		if outputData.Valid {
			exec.OutputData = outputData.String
		}
		if errorMessage.Valid {
			exec.ErrorMessage = errorMessage.String
		}

		executions = append(executions, &exec)
	}

	return executions, nil
}

// getOrCreateToolID 获取或创建工具ID
func (r *Repository) getOrCreateToolID(ctx context.Context, toolName string) (int64, error) {
	// 先尝试查找
	var toolID int64
	query := `SELECT id FROM tools WHERE name = ? LIMIT 1`
	err := r.db.QueryRowContext(ctx, query, toolName).Scan(&toolID)

	if err == sql.ErrNoRows {
		// 不存在，创建新记录
		insertQuery := `
			INSERT INTO tools (name, type, description, config, enabled)
			VALUES (?, ?, ?, ?, ?)
		`
		result, err := r.db.ExecContext(ctx, insertQuery,
			toolName,
			"custom",
			"Auto-created tool",
			"{}",
			true,
		)
		if err != nil {
			return 0, fmt.Errorf("failed to create tool: %w", err)
		}

		return result.LastInsertId()
	}

	if err != nil {
		return 0, fmt.Errorf("failed to query tool: %w", err)
	}

	return toolID, nil
}

// GetToolStats 获取工具使用统计
func (r *Repository) GetToolStats(ctx context.Context, toolName string, days int) (*ToolStats, error) {
	query := `
		SELECT
			COUNT(*) as total_calls,
			SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END) as success_calls,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_calls,
			AVG(duration_ms) as avg_duration_ms,
			MAX(duration_ms) as max_duration_ms
		FROM tool_executions te
		JOIN tools t ON te.tool_id = t.id
		WHERE t.name = ?
		  AND te.created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)
	`

	var stats ToolStats
	var totalCalls, successCalls, failedCalls int
	var avgDuration, maxDuration sql.NullFloat64

	err := r.db.QueryRowContext(ctx, query, toolName, days).Scan(
		&totalCalls,
		&successCalls,
		&failedCalls,
		&avgDuration,
		&maxDuration,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	stats.TotalCalls = totalCalls
	stats.SuccessCalls = successCalls
	stats.FailedCalls = failedCalls
	if avgDuration.Valid {
		stats.AvgDurationMs = int(avgDuration.Float64)
	}
	if maxDuration.Valid {
		stats.MaxDurationMs = int(maxDuration.Float64)
	}

	return &stats, nil
}

// ToolStats 工具统计信息
type ToolStats struct {
	TotalCalls     int `json:"total_calls"`
	SuccessCalls   int `json:"success_calls"`
	FailedCalls    int `json:"failed_calls"`
	AvgDurationMs  int `json:"avg_duration_ms"`
	MaxDurationMs  int `json:"max_duration_ms"`
}

// ListRecentExecutions 列出最近的工具执行记录
func (r *Repository) ListRecentExecutions(ctx context.Context, limit int) ([]*models.ToolExecution, error) {
	query := `
		SELECT te.id, te.message_id, te.tool_id, t.name as tool_name,
		       te.input_params, te.output_data, te.status, te.duration_ms,
		       te.error_message, te.created_at
		FROM tool_executions te
		JOIN tools t ON te.tool_id = t.id
		ORDER BY te.created_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query executions: %w", err)
	}
	defer rows.Close()

	executions := make([]*models.ToolExecution, 0)
	for rows.Next() {
		var exec models.ToolExecution
		var inputParams, outputData, errorMessage sql.NullString
		var toolName sql.NullString

		err := rows.Scan(
			&exec.ID,
			&exec.MessageID,
			&exec.ToolID,
			&toolName,
			&inputParams,
			&outputData,
			&exec.Status,
			&exec.DurationMs,
			&errorMessage,
			&exec.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan execution: %w", err)
		}

		if inputParams.Valid {
			exec.InputParams = inputParams.String
		}
		if outputData.Valid {
			exec.OutputData = outputData.String
		}
		if errorMessage.Valid {
			exec.ErrorMessage = errorMessage.String
		}

		executions = append(executions, &exec)
	}

	return executions, nil
}
