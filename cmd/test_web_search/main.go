package main

import (
	"context"
	"encoding/json"
	"fmt"

	"link/internal/agent/tool"
	"link/internal/config"
)

func main() {
	// 1. 加载搜索配置
	searchCfg := config.LoadSearchConfig()
	fmt.Printf("搜索配置:\n")
	fmt.Printf("  API Endpoint: %s\n", searchCfg.APIEndpoint)
	fmt.Printf("  API Key: %s...%s\n", searchCfg.MetasoAPIKey[:5], searchCfg.MetasoAPIKey[len(searchCfg.MetasoAPIKey)-5:])

	// 2. 初始化 Metaso 客户端
	tool.InitMetasoClient(searchCfg)

	// 3. 创建搜索工具
	webSearchTool, err := tool.NewWebSearchTool()
	if err != nil {
		fmt.Printf("创建搜索工具失败: %v\n", err)
		return
	}

	// 4. 获取工具信息
	info, err := webSearchTool.Info(context.Background())
	if err != nil {
		fmt.Printf("获取工具信息失败: %v\n", err)
		return
	}

	fmt.Printf("\n工具信息:\n")
	fmt.Printf("  名称: %s\n", info.Name)
	fmt.Printf("  描述: %s\n", info.Desc)

	// 5. 测试搜索（直接调用底层函数）
	fmt.Printf("\n开始测试搜索...\n")

	// 创建客户端进行测试
	client := tool.NewMetasoClient(searchCfg)
	result, err := client.Search(context.Background(), "日本女优top10", 3)
	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}

	fmt.Printf("\n搜索结果 (共 %d 条, 总计 %d 条):\n", len(result.Webpages), result.Total)
	for i, item := range result.Webpages {
		fmt.Printf("\n[%d] %s\n", i+1, item.Title)
		fmt.Printf("    URL: %s\n", item.Link)
		fmt.Printf("    摘要: %s\n", item.Snippet)
	}

	// 6. 通过工具调用测试
	fmt.Printf("\n\n通过工具调用测试...\n")

	// 获取工具注册表
	registry, err := tool.InitDefaultTools()
	if err != nil {
		fmt.Printf("初始化工具注册表失败: %v\n", err)
		return
	}

	// 获取执行器
	executor := tool.NewExecutor(registry)

	execResult := executor.ExecuteByName(context.Background(), "web_search", map[string]interface{}{
		"query": "日本女优top10",
		"limit": 2,
	})

	if execResult.Success {
		fmt.Printf("工具调用成功!\n")

		// 解析结果
		var searchResult tool.WebSearchResult
		if err := json.Unmarshal([]byte(execResult.Data), &searchResult); err == nil {
			fmt.Printf("返回 %d 条结果:\n", searchResult.Count)
			for i, item := range searchResult.Items {
				fmt.Printf("\n[%d] %s\n", i+1, item.Title)
				fmt.Printf("    URL: %s\n", item.URL)
			}
		}
	} else {
		fmt.Printf("工具调用失败: %v\n", execResult.Error)
	}

	fmt.Printf("\n\n可用工具列表:\n")
	for _, name := range registry.List() {
		fmt.Printf("  - %s\n", name)
	}
}
