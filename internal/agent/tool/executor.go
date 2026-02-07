package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

// Executor 工具执行器
type Executor struct {
	registry *Registry
}

// NewExecutor 创建工具执行器
func NewExecutor(registry *Registry) *Executor {
	return &Executor{
		registry: registry,
	}
}

// Execute 执行工具调用
func (e *Executor) Execute(ctx context.Context, toolCall *schema.ToolCall, opts ...tool.Option) *ToolExecResult {
	startTime := time.Now()

	// 获取工具
	t, ok := e.registry.Get(toolCall.Function.Name)
	if !ok {
		return &ToolExecResult{
			Success:    false,
			Error:      fmt.Errorf("tool not found: %s", toolCall.Function.Name),
			DurationMs: 0,
			ToolName:   toolCall.Function.Name,
		}
	}

	// 根据工具类型执行
	var result string
	var err error

	switch tl := t.(type) {
	case tool.InvokableTool:
		result, err = tl.InvokableRun(ctx, toolCall.Function.Arguments, opts...)
	case tool.StreamableTool:
		stream, streamErr := tl.StreamableRun(ctx, toolCall.Function.Arguments, opts...)
		if streamErr != nil {
			err = streamErr
			break
		}
		// 收集流式结果
		var builder strings.Builder
		for {
			chunk, recvErr := stream.Recv()
			if recvErr != nil {
				if recvErr.Error() == "EOF" {
					break
				}
				err = recvErr
				break
			}
			builder.WriteString(chunk)
		}
		stream.Close()
		result = builder.String()
	default:
		err = fmt.Errorf("unsupported tool type: %T", t)
	}

	duration := time.Since(startTime)

	if err != nil {
		return &ToolExecResult{
			Success:    false,
			Error:      err,
			DurationMs: int(duration.Milliseconds()),
			ToolName:   toolCall.Function.Name,
		}
	}

	return &ToolExecResult{
		Success:    true,
		Data:       result,
		DurationMs: int(duration.Milliseconds()),
		ToolName:   toolCall.Function.Name,
	}
}

// ExecuteAll 执行多个工具调用
func (e *Executor) ExecuteAll(ctx context.Context, toolCalls []*schema.ToolCall, opts ...tool.Option) []*ToolExecResult {
	results := make([]*ToolExecResult, len(toolCalls))
	for i, tc := range toolCalls {
		results[i] = e.Execute(ctx, tc, opts...)
	}
	return results
}

// ExecuteByName 根据工具名和参数执行
func (e *Executor) ExecuteByName(ctx context.Context, toolName string, arguments map[string]interface{}) *ToolExecResult {
	argsJSON, err := json.Marshal(arguments)
	if err != nil {
		return &ToolExecResult{
			Success: false,
			Error:   fmt.Errorf("failed to marshal arguments: %w", err),
			ToolName: toolName,
		}
	}

	toolCall := &schema.ToolCall{
		Function: schema.FunctionCall{
			Name:      toolName,
			Arguments: string(argsJSON),
		},
	}

	return e.Execute(ctx, toolCall)
}

// ExecuteWithTimeout 带超时的工具执行
func (e *Executor) ExecuteWithTimeout(ctx context.Context, toolCall *schema.ToolCall, timeout time.Duration, opts ...tool.Option) *ToolExecResult {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return e.Execute(ctx, toolCall, opts...)
}

// ValidateToolCall 验证工具调用参数
func (e *Executor) ValidateToolCall(ctx context.Context, toolCall *schema.ToolCall) error {
	t, ok := e.registry.Get(toolCall.Function.Name)
	if !ok {
		return fmt.Errorf("tool not found: %s", toolCall.Function.Name)
	}

	// 验证参数格式
	var arguments map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &arguments); err != nil {
		return fmt.Errorf("invalid arguments JSON: %w", err)
	}

	// 获取工具信息进行参数验证
	info, err := t.Info(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tool info: %w", err)
	}

	// 如果有参数约束，进行验证
	if info.ParamsOneOf != nil {
		// 这里可以添加更详细的参数验证逻辑
		// 目前只验证 JSON 格式
	}

	return nil
}

// FormatToolResult 格式化工具执行结果为消息内容
func FormatToolResult(result *ToolExecResult) string {
	if result.Success {
		return result.Data
	}
	return fmt.Sprintf("Error: %v", result.Error)
}

// FormatToolResults 格式化多个工具执行结果
func FormatToolResults(results []*ToolExecResult) []string {
	formatted := make([]string, len(results))
	for i, r := range results {
		formatted[i] = FormatToolResult(r)
	}
	return formatted
}
