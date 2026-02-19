package tool

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/tool"
	"link/internal/config"
)

// 全局搜索客户端（单例）
var (
	globalMetasoClient *MetasoClient
	globalSearchConfig *config.SearchConfig
	searchClientOnce   sync.Once
)

// Registry 工具注册中心
type Registry struct {
	mu    sync.RWMutex
	tools map[string]tool.BaseTool
}

// NewRegistry 创建工具注册中心
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]tool.BaseTool),
	}
}

// InitGlobalSearchClient 初始化全局搜索客户端（单例）
func InitGlobalSearchClient(cfg *config.SearchConfig) {
	globalSearchConfig = cfg
	searchClientOnce.Do(func() {
		if cfg != nil && cfg.MetasoAPIKey != "" {
			globalMetasoClient = NewMetasoClient(cfg)
		}
	})
}

// GetGlobalSearchClient 获取全局搜索客户端
func GetGlobalSearchClient() *MetasoClient {
	return globalMetasoClient
}

// ========================================
// 工具注册器
// ========================================

// ToolRegistrar 工具注册器接口
type ToolRegistrar interface {
	// Register 注册单个工具
	Register(name string, t tool.BaseTool) error
	// MustRegister 注册单个工具，失败时 panic
	MustRegister(name string, t tool.BaseTool)
	// RegisterBatch 批量注册工具
	RegisterBatch(tools map[string]tool.BaseTool) error
	// RegisterRAGTool 注册 RAG 检索工具
	RegisterRAGTool() error
	// RegisterWebSearchTool 注册网络搜索工具
	RegisterWebSearchTool() error
	// RegisterSmartRetrievalTool 注册智能检索工具（高级）
	RegisterSmartRetrievalTool() error
	// RegisterUtilityTools 注册实用工具（时间、计算器、HTTP请求）
	RegisterUtilityTools() error
	// RegisterDefaultTools 注册所有默认工具
	RegisterDefaultTools() error
}

// 确保实现了接口
var _ ToolRegistrar = (*Registry)(nil)

// MustRegister 注册单个工具，失败时 panic
func (r *Registry) MustRegister(name string, t tool.BaseTool) {
	if err := r.Register(name, t); err != nil {
		panic(fmt.Errorf("failed to register tool %s: %w", name, err))
	}
}

// RegisterBatch 批量注册工具
func (r *Registry) RegisterBatch(tools map[string]tool.BaseTool) error {
	for name, t := range tools {
		if err := r.Register(name, t); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", name, err)
		}
	}
	return nil
}

// RegisterRAGTool 注册 RAG 检索工具
func (r *Registry) RegisterRAGTool() error {
	ragQueryTool, err := NewRAGQueryTool()
	if err != nil {
		return fmt.Errorf("failed to create rag_query tool: %w", err)
	}
	return r.Register("rag_query", ragQueryTool)
}

// RegisterWebSearchTool 注册网络搜索工具
func (r *Registry) RegisterWebSearchTool() error {
	// 确保全局搜索客户端已初始化
	if globalMetasoClient == nil && globalSearchConfig != nil {
		InitGlobalSearchClient(globalSearchConfig)
	}
	// 设置全局客户端供工具使用
	SetMetasoClient(globalMetasoClient)

	webSearchTool, err := NewWebSearchTool()
	if err != nil {
		return fmt.Errorf("failed to create web_search tool: %w", err)
	}
	return r.Register("web_search", webSearchTool)
}

// RegisterUtilityTools 注册实用工具（时间、计算器、HTTP请求）
func (r *Registry) RegisterUtilityTools() error {
	// 获取当前时间工具
	getTimeTool, err := NewGetCurrentTimeTool()
	if err != nil {
		return fmt.Errorf("failed to create get_current_time tool: %w", err)
	}
	if err := r.Register("get_current_time", getTimeTool); err != nil {
		return err
	}

	// 计算器工具
	calcTool, err := NewCalculatorTool()
	if err != nil {
		return fmt.Errorf("failed to create calculator tool: %w", err)
	}
	if err := r.Register("calculator", calcTool); err != nil {
		return err
	}

	// HTTP 请求工具
	httpTool, err := NewHttpRequestTool()
	if err != nil {
		return fmt.Errorf("failed to create http_request tool: %w", err)
	}
	if err := r.Register("http_request", httpTool); err != nil {
		return err
	}

	return nil
}

// RegisterSmartRetrievalTool 注册智能检索工具
// 注意：此工具需要额外的依赖（知识库repository），默认不注册
// 需要使用 SmartRetrievalToolFactory 来创建和注册
func (r *Registry) RegisterSmartRetrievalTool() error {
	smartTool, err := NewSmartRetrievalTool()
	if err != nil {
		return fmt.Errorf("failed to create smart_retrieval tool: %w", err)
	}
	return r.Register("smart_retrieval", smartTool)
}

// RegisterDefaultTools 注册所有默认工具
func (r *Registry) RegisterDefaultTools() error {
	// 1. RAG 检索工具（最优先）
	if err := r.RegisterRAGTool(); err != nil {
		return err
	}

	// 2. 网络搜索工具
	if err := r.RegisterWebSearchTool(); err != nil {
		return err
	}

	// 3. 实用工具
	if err := r.RegisterUtilityTools(); err != nil {
		return err
	}

	// 注意：smart_retrieval 工具不在此处注册，需要使用工厂创建

	return nil
}

// ========================================
// 全局工具注册中心（单例）
// ========================================

var (
	defaultRegistry *Registry
	registryOnce    sync.Once
)

// GetDefaultRegistry 获取默认工具注册中心（单例）
func GetDefaultRegistry() *Registry {
	registryOnce.Do(func() {
		defaultRegistry = NewRegistry()
	})
	return defaultRegistry
}

// InitDefaultTools 初始化默认工具集（使用全局注册中心）
// 保留此函数以向后兼容
func InitDefaultTools() (*Registry, error) {
	return InitDefaultToolsWithConfig(nil)
}

// InitDefaultToolsWithConfig 使用指定配置初始化默认工具集
func InitDefaultToolsWithConfig(cfg *config.SearchConfig) (*Registry, error) {
	// 初始化搜索客户端
	if cfg != nil {
		InitGlobalSearchClient(cfg)
	}

	registry := GetDefaultRegistry()

	// 如果工具已经注册过，直接返回
	if registry.Count() > 0 {
		return registry, nil
	}

	// 注册所有默认工具
	if err := registry.RegisterDefaultTools(); err != nil {
		return nil, err
	}

	return registry, nil
}

// InitDefaultToolsWithSearch 初始化默认工具集（带搜索配置）
// 保留此函数以向后兼容
func InitDefaultToolsWithSearch(searchConfig interface{}) (*Registry, error) {
	var cfg *config.SearchConfig
	if sc, ok := searchConfig.(*config.SearchConfig); ok {
		cfg = sc
	}
	return InitDefaultToolsWithConfig(cfg)
}

// InitCustomTools 初始化自定义工具集
// 保留此函数以向后兼容
func InitCustomTools(toolNames []string) (*Registry, error) {
	registry, err := InitDefaultTools()
	if err != nil {
		return nil, err
	}

	// 如果指定了工具名称，只保留这些工具
	if len(toolNames) > 0 {
		filteredRegistry := NewRegistry()
		for _, name := range toolNames {
			if t, ok := registry.Get(name); ok {
				filteredRegistry.Register(name, t)
			}
		}
		return filteredRegistry, nil
	}

	return registry, nil
}

// GetToolsByName 根据名称列表获取工具
// 保留此函数以向后兼容
func GetToolsByName(toolNames []string) ([]tool.BaseTool, error) {
	registry, err := InitDefaultTools()
	if err != nil {
		return nil, err
	}
	tools, _ := registry.GetToolsByNames(toolNames)
	return tools, nil
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
