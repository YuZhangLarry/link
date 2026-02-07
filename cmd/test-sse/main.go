package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"link/internal/config"
	"link/internal/models/chat"
)

func main() {
	log.Println("🚀 AI对话SSE测试程序")
	log.Println("==========================================")

	// 加载配置
	chatConfig := config.LoadChatConfig()

	// 检查API Key
	if chatConfig.APIKey == "" || chatConfig.APIKey == "your-openai-api-key-here" {
		log.Println("❌ 错误: 请先在.env文件中配置CHAT_API_KEY")
		log.Println("示例:")
		log.Println("  CHAT_BASE_URL=https://api.openai.com/v1")
		log.Println("  CHAT_MODEL_NAME=gpt-3.5-turbo")
		log.Println("  CHAT_API_KEY=sk-your-api-key-here")
		log.Println("  CHAT_PROVIDER=openai")
		os.Exit(1)
	}

	log.Printf("📝 配置信息:")
	log.Printf("  Provider: %s", chatConfig.Provider)
	log.Printf("  BaseURL: %s", chatConfig.BaseURL)
	log.Printf("  Model: %s", chatConfig.ModelName)
	log.Printf("  Source: %s", chatConfig.Source)
	log.Println("")

	// 创建聊天配置
	cfg := &chat.ChatConfig{
		Source:    chatConfig.Source,
		BaseURL:   chatConfig.BaseURL,
		ModelName: chatConfig.ModelName,
		APIKey:    chatConfig.APIKey,
		Provider:  chatConfig.Provider,
		ModelID:   fmt.Sprintf("test_%d", time.Now().UnixNano()),
	}

	// 创建聊天实例
	chatInstance, err := chat.NewChat(cfg)
	if err != nil {
		log.Fatalf("❌ 创建聊天实例失败: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 选择测试模式
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "simple":
			runSimpleTest(ctx, chatInstance)
		case "stream":
			runStreamTest(ctx, chatInstance)
		case "multi":
			runMultiTurnTest(ctx, chatInstance)
		default:
			printUsage()
		}
	} else {
		// 默认运行流式测试
		runStreamTest(ctx, chatInstance)
	}
}

func printUsage() {
	fmt.Println("使用方法:")
	fmt.Println("  go run cmd/test-sse/main.go [模式]")
	fmt.Println("")
	fmt.Println("模式:")
	fmt.Println("  simple  - 简单非流式对话测试")
	fmt.Println("  stream  - 流式对话测试(SSE) [默认]")
	fmt.Println("  multi   - 多轮对话测试")
}

func runSimpleTest(ctx context.Context, chatInstance chat.Chat) {
	log.Println("==========================================")
	log.Println("📝 模式: 简单非流式对话")
	log.Println("==========================================")

	messages := []chat.Message{
		{
			Role:    "user",
			Content: "你好！请用一句话介绍一下你自己。",
		},
	}

	opts := &chat.ChatOptions{
		Temperature: 0.7,
		MaxTokens:   100,
	}

	log.Println("📤 发送消息...")
	resp, err := chatInstance.Chat(ctx, messages, opts)
	if err != nil {
		log.Fatalf("❌ 对话失败: %v", err)
	}

	log.Println("")
	log.Println("📥 响应:")
	log.Printf("  MessageID: %s", resp.MessageID)
	log.Printf("  Role: %s", resp.Role)
	log.Printf("  Content: %s", resp.Content)
	log.Printf("  TokenCount: %d", resp.TokenCount)
	log.Println("")
	log.Println("✅ 测试完成")
}

func runStreamTest(ctx context.Context, chatInstance chat.Chat) {
	log.Println("==========================================")
	log.Println("📝 模式: 流式对话(SSE)")
	log.Println("==========================================")

	messages := []chat.Message{
		{
			Role:    "user",
			Content: "请写一首关于春天的短诗，不超过100字。",
		},
	}

	opts := &chat.ChatOptions{
		Temperature: 0.8,
		MaxTokens:   200,
	}

	log.Println("📤 发送消息...")
	log.Println("")
	log.Println("📥 流式响应:")

	respChan, err := chatInstance.ChatStream(ctx, messages, opts)
	if err != nil {
		log.Fatalf("❌ 流式对话失败: %v", err)
	}

	log.Println("----------------------------------------")

	var fullContent string
	eventCount := 0
	startTime := time.Now()
	firstChunkTime := time.Time{}

	for resp := range respChan {
		eventCount++

		switch resp.Event {
		case chat.EventStart:
			if firstChunkTime.IsZero() {
				firstChunkTime = time.Now()
			}
			log.Printf("  [START] MessageID: %s", resp.MessageID)

		case chat.EventContent:
			if firstChunkTime.IsZero() {
				firstChunkTime = time.Now()
			}
			log.Printf("  [CONTENT] %s", resp.Content)
			fullContent += resp.Content

		case chat.EventEnd:
			log.Printf("  [END] 流式传输完成")
			log.Println("")

		case chat.EventError:
			log.Printf("  [ERROR] %s", resp.Error)
		}
	}

	log.Println("==========================================")
	log.Printf("✅ 测试完成")
	log.Printf("  总事件数: %d", eventCount)
	log.Printf("  首字延迟: %v", firstChunkTime.Sub(startTime))
	log.Printf("  总耗时: %v", time.Since(startTime))
	log.Printf("  内容长度: %d 字符", len(fullContent))
	log.Println("")
	log.Printf("  完整内容:\n  %s", fullContent)
	log.Println("==========================================")
}

func runMultiTurnTest(ctx context.Context, chatInstance chat.Chat) {
	log.Println("==========================================")
	log.Println("📝 模式: 多轮对话")
	log.Println("==========================================")

	opts := &chat.ChatOptions{
		Temperature: 0.7,
		MaxTokens:   100,
	}

	// 第一轮对话
	log.Println("📤 第一轮对话")
	messages1 := []chat.Message{
		{Role: "user", Content: "我的名字叫张三，是一名程序员"},
	}

	resp1, err := chatInstance.Chat(ctx, messages1, opts)
	if err != nil {
		log.Fatalf("❌ 第一轮对话失败: %v", err)
	}

	log.Printf("  用户: 我的名字叫张三，是一名程序员")
	log.Printf("  AI: %s", resp1.Content)
	log.Println("")

	// 第二轮对话（带上下文）
	log.Println("📤 第二轮对话（带上下文）")
	messages2 := []chat.Message{
		{Role: "user", Content: "我的名字叫张三，是一名程序员"},
		{Role: "assistant", Content: resp1.Content},
		{Role: "user", Content: "我叫什么名字？做什么工作？"},
	}

	resp2, err := chatInstance.Chat(ctx, messages2, opts)
	if err != nil {
		log.Fatalf("❌ 第二轮对话失败: %v", err)
	}

	log.Printf("  用户: 我叫什么名字？做什么工作？")
	log.Printf("  AI: %s", resp2.Content)
	log.Println("")

	// 验证上下文记忆
	if containsChinese(resp2.Content, "张三") {
		log.Println("✅ AI成功记住了名字")
	} else {
		log.Println("⚠️  AI没有记住名字（某些模型可能不支持长上下文）")
	}

	log.Println("==========================================")
	log.Println("✅ 多轮对话测试完成")
}

// containsChinese 检查字符串是否包含中文
func containsChinese(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
