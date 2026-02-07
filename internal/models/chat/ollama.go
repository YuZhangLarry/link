package chat

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/model"
)

// ========================================
// Ollama Chat Implementation
// ========================================

// ollamaChat Ollama聊天实现（本地模型）
type ollamaChat struct {
	base *baseChat
}

// NewOllamaChat 创建Ollama聊天实例
func NewOllamaChat(config *ChatConfig) (Chat, error) {
	if config.ModelName == "" {
		return nil, fmt.Errorf("model_name is required for ollama")
	}

	// 默认baseURL
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434"
	}

	// TODO: 使用eino的Ollama adapter创建模型
	// 目前先返回占位符实现
	return &ollamaChat{
		base: &baseChat{
			config:  config,
			modelID: config.ModelID,
		},
	}, nil
}

// Chat 进行非流式聊天
func (c *ollamaChat) Chat(ctx context.Context, messages []Message, opts *ChatOptions) (*ChatResponse, error) {
	// TODO: 实现Ollama聊天
	return nil, fmt.Errorf("ollama chat not implemented yet - please integrate eino ollama adapter")
}

// ChatStream 进行流式聊天
func (c *ollamaChat) ChatStream(ctx context.Context, messages []Message, opts *ChatOptions) (<-chan StreamResponse, error) {
	// TODO: 实现Ollama流式聊天
	return nil, fmt.Errorf("ollama stream chat not implemented yet - please integrate eino ollama adapter")
}

// GetModelName 获取模型名称
func (c *ollamaChat) GetModelName() string {
	return c.base.config.ModelName
}

// GetModelID 获取模型ID
func (c *ollamaChat) GetModelID() string {
	return c.base.modelID
}

// ========================================
// Ollama Model Creation
// ========================================

// createOllamaModel 创建Ollama模型
func createOllamaModel(config *ChatConfig) (model.BaseChatModel, error) {
	// TODO: 实现Ollama模型创建
	// 需要集成eino的Ollama adapter
	return nil, fmt.Errorf("ollama model creation not implemented yet")
}
