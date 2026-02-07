package tool

import (
	"fmt"

	"github.com/cloudwego/eino/components/tool"
)

// InitDefaultTools 初始化默认工具集
func InitDefaultTools() (*Registry, error) {
	registry := NewRegistry()

	// 1. 知识库相关工具
	kbQueryTool, err := NewKbQueryTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create kb_query tool: %w", err)
	}
	if err := registry.Register("kb_query", kbQueryTool); err != nil {
		return nil, err
	}

	kbListTool, err := NewKbListTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create kb_list tool: %w", err)
	}
	if err := registry.Register("kb_list", kbListTool); err != nil {
		return nil, err
	}

	docListTool, err := NewDocumentListTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create document_list tool: %w", err)
	}
	if err := registry.Register("document_list", docListTool); err != nil {
		return nil, err
	}

	// 2. 搜索相关工具
	webSearchTool, err := NewWebSearchTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create web_search tool: %w", err)
	}
	if err := registry.Register("web_search", webSearchTool); err != nil {
		return nil, err
	}

	// 3. 实用工具
	getTimeTool, err := NewGetCurrentTimeTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create get_current_time tool: %w", err)
	}
	if err := registry.Register("get_current_time", getTimeTool); err != nil {
		return nil, err
	}

	calcTool, err := NewCalculatorTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create calculator tool: %w", err)
	}
	if err := registry.Register("calculator", calcTool); err != nil {
		return nil, err
	}

	httpTool, err := NewHttpRequestTool()
	if err != nil {
		return nil, fmt.Errorf("failed to create http_request tool: %w", err)
	}
	if err := registry.Register("http_request", httpTool); err != nil {
		return nil, err
	}

	return registry, nil
}

// InitDefaultToolsWithSearch 初始化默认工具集（带搜索配置）
func InitDefaultToolsWithSearch(searchConfig interface{}) (*Registry, error) {
	// TODO: 可以在这里初始化搜索客户端
	// if cfg, ok := searchConfig.(*config.SearchConfig); ok {
	//     InitMetasoClient(cfg)
	// }

	return InitDefaultTools()
}

// InitCustomTools 初始化自定义工具集
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
func GetToolsByName(toolNames []string) ([]tool.BaseTool, error) {
	registry, err := InitDefaultTools()
	if err != nil {
		return nil, err
	}

	return registry.GetToolsByNames(toolNames)
}
