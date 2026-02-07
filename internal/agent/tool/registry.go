package tool

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/tool"
)

// Registry 工具注册中心
type Registry struct {
	mu   sync.RWMutex
	tools map[string]tool.BaseTool
}

// NewRegistry 创建工具注册中心
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]tool.BaseTool),
	}
}

// Register 注册工具
func (r *Registry) Register(name string, t tool.BaseTool) error {
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if t == nil {
		return fmt.Errorf("tool cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools[strings.ToLower(name)] = t
	return nil
}

// Unregister 注销工具
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, strings.ToLower(name))
}

// Get 获取工具
func (r *Registry) Get(name string) (tool.BaseTool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[strings.ToLower(name)]
	return t, ok
}

// List 列出所有工具名称
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}

// GetTools 获取所有工具
func (r *Registry) GetTools() []tool.BaseTool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]tool.BaseTool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// GetToolsByNames 根据名称列表获取工具
func (r *Registry) GetToolsByNames(names []string) ([]tool.BaseTool, error) {
	if len(names) == 0 {
		return r.GetTools(), nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]tool.BaseTool, 0, len(names))
	for _, name := range names {
		t, ok := r.tools[strings.ToLower(name)]
		if !ok {
			return nil, fmt.Errorf("tool not found: %s", name)
		}
		tools = append(tools, t)
	}
	return tools, nil
}

// GetToolInfo 获取工具信息
func (r *Registry) GetToolInfo(ctx context.Context, name string) (map[string]interface{}, error) {
	t, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	info, err := t.Info(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tool info: %w", err)
	}

	return map[string]interface{}{
		"name":        info.Name,
		"description": info.Desc,
		"params":      info.ParamsOneOf,
	}, nil
}

// GetAllToolsInfo 获取所有工具信息
func (r *Registry) GetAllToolsInfo(ctx context.Context) ([]map[string]interface{}, error) {
	tools := r.GetTools()
	infos := make([]map[string]interface{}, 0, len(tools))

	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		infos = append(infos, map[string]interface{}{
			"name":        info.Name,
			"description": info.Desc,
			"params":      info.ParamsOneOf,
		})
	}

	return infos, nil
}

// Count 获取工具数量
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// Clear 清空所有工具
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools = make(map[string]tool.BaseTool)
}
