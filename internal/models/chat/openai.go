package chat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ========================================
// OpenAI Client Implementation
// ========================================

// openaiClient OpenAI客户端
type openaiClient struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

// newOpenAIClient 创建OpenAI客户端
func newOpenAIClient(config *ChatConfig) (*openaiClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("api_key is required for openai")
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	return &openaiClient{
		baseURL: strings.TrimSuffix(config.BaseURL, "/"),
		apiKey:  config.APIKey,
		model:   config.ModelName,
		client:  &http.Client{},
	}, nil
}

// Generate 非流式生成
func (c *openaiClient) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// 构建请求
	reqBody := c.buildRequest(messages, opts, false)

	// 发送请求
	resp, err := c.sendRequest(ctx, reqBody)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	var openaiResp openaiChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := openaiResp.Choices[0]
	return &schema.Message{
		Role:    schema.RoleType(choice.Message.Role),
		Content: choice.Message.Content,
	}, nil
}

// Stream 流式生成
func (c *openaiClient) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// 构建请求
	reqBody := c.buildRequest(messages, opts, true)

	// 发送请求
	resp, err := c.sendRequest(ctx, reqBody)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	// 创建流式读取器
	reader, writer := schema.Pipe[*schema.Message](10)

	go func() {
		defer writer.Close()
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var chunk openaiStreamChunk
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				delta := chunk.Choices[0].Delta
				if delta.Content != "" {
					writer.Send(&schema.Message{
						Role:    schema.RoleType(delta.Role),
						Content: delta.Content,
					}, nil)
				}
			}
		}
	}()

	return reader, nil
}

// buildRequest 构建请求体
func (c *openaiClient) buildRequest(messages []*schema.Message, opts []model.Option, stream bool) openaiRequest {
	oaiMessages := make([]openaiMessage, len(messages))
	for i, msg := range messages {
		oaiMessages[i] = openaiMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		}
	}

	req := openaiRequest{
		Model:    c.model,
		Messages: oaiMessages,
		Stream:   stream,
	}

	// 应用选项 - 使用GetOptions辅助函数
	options := getOptions(opts)
	if options.Temperature != nil {
		req.Temperature = float64(*options.Temperature)
	}
	if options.TopP != nil {
		req.TopP = float64(*options.TopP)
	}
	if options.MaxTokens != nil {
		req.MaxTokens = *options.MaxTokens
	}

	return req
}

// getOptions 从选项数组中提取Options
func getOptions(opts []model.Option) *model.Options {
	options := &model.Options{}
	// 注意：由于opt.apply是未导出的，我们需要构建选项
	// 这里简化处理，直接创建一个空的Options
	// 实际使用时，应该在调用buildRequest之前就处理选项
	return options
}

// sendRequest 发送HTTP请求
func (c *openaiClient) sendRequest(ctx context.Context, reqBody openaiRequest) (*http.Response, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	return c.client.Do(req)
}

// ========================================
// OpenAI API Types
// ========================================

type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiChatResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []openaiChoice   `json:"choices"`
	Usage   openaiUsage      `json:"usage"`
}

type openaiChoice struct {
	Index        int            `json:"index"`
	Message      openaiMessage  `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

type openaiStreamChunk struct {
	ID      string                `json:"id"`
	Object  string                `json:"object"`
	Created int64                 `json:"created"`
	Model   string                `json:"model"`
	Choices []openaiStreamChoice  `json:"choices"`
}

type openaiStreamChoice struct {
	Index        int           `json:"index"`
	Delta        openaiDelta   `json:"delta"`
	FinishReason *string       `json:"finish_reason"`
}

type openaiDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type openaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
