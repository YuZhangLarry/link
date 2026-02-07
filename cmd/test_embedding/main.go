package main

import (
	"context"
	"fmt"

	"link/internal/config"
	"link/internal/models/embedding"
)

func main() {
	// 1. 加载配置
	embCfg := config.LoadEmbeddingConfig()
	fmt.Printf("Embedding 配置:\n")
	fmt.Printf("  提供商: %s\n", embCfg.Provider)
	fmt.Printf("  模型: %s\n", embCfg.Model)
	fmt.Printf("  API Key: %s...%s\n", embCfg.APIKey[:10], embCfg.APIKey[len(embCfg.APIKey)-5:])
	fmt.Printf("  Base URL: %s\n\n", embCfg.BaseURL)

	// 2. 创建 Embedder
	embedder, err := embedding.NewEmbedder(embCfg)
	if err != nil {
		fmt.Printf("创建 Embedder 失败: %v\n", err)
		return
	}

	fmt.Println("✅ Embedder 创建成功")

	// 3. 测试向量化
	texts := []string{
		"人工智能是计算机科学的一个分支",
		"机器学习是人工智能的核心技术",
		"深度学习使用神经网络进行学习",
	}

	fmt.Printf("\n开始向量化 %d 个文本...\n", len(texts))

	embeddings, err := embedder.EmbedStrings(context.Background(), texts)
	if err != nil {
		fmt.Printf("向量化失败: %v\n", err)
		return
	}

	fmt.Printf("\n✅ 向量化成功！\n")
	for i, text := range texts {
		vec := embeddings[i]
		fmt.Printf("\n[%d] 文本: %s\n", i+1, text)
		fmt.Printf("    向量维度: %d\n", len(vec))
		if len(vec) > 0 {
			fmt.Printf("    前5个值: [%.4f, %.4f, %.4f, %.4f, %.4f, ...]\n",
				vec[0], vec[1], vec[2], vec[3], vec[4])
		}
	}

	// 4. 测试单文本向量化
	fmt.Printf("\n\n测试单文本向量化...\n")
	singleTexts := []string{"Go 语言是一门开源编程语言"}
	singleEmbeddings, err := embedder.EmbedStrings(context.Background(), singleTexts)
	if err != nil {
		fmt.Printf("单文本向量化失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 单文本向量化成功！\n")
	fmt.Printf("文本: %s\n", singleTexts[0])
	fmt.Printf("向量维度: %d\n", len(singleEmbeddings[0]))

	// 5. 测试空文本
	fmt.Printf("\n\n测试空文本处理...\n")
	_, err = embedder.EmbedStrings(context.Background(), []string{})
	if err != nil {
		fmt.Printf("❌ 空文本错误（符合预期）: %v\n", err)
	}

	fmt.Printf("\n\n🎉 所有测试完成！\n")
}
