package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"link/internal/config"
)

// MetasoClient Metaso 搜索客户端
type MetasoClient struct {
	apiKey     string
	apiEndpoint string
	httpClient *http.Client
}

// NewMetasoClient 创建 Metaso 客户端
func NewMetasoClient(cfg *config.SearchConfig) *MetasoClient {
	return &MetasoClient{
		apiKey:     cfg.MetasoAPIKey,
		apiEndpoint: cfg.APIEndpoint,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MetasoSearchRequest Metaso 搜索请求
type MetasoSearchRequest struct {
	Q                 string `json:"q"`                 // 查询内容
	Scope             string `json:"scope"`             // 搜索范围: webpage
	IncludeSummary    bool   `json:"includeSummary"`    // 是否包含摘要
	Size              int    `json:"size"`              // 返回结果数量
	IncludeRawContent bool   `json:"includeRawContent"` // 是否包含原始内容
	ConciseSnippet    bool   `json:"conciseSnippet"`    // 是否简洁摘要
}

// MetasoSearchResponse Metaso 搜索响应
type MetasoSearchResponse struct {
	Credits          int                `json:"credits"`
	SearchParameters MetasoSearchRequest `json:"searchParameters"`
	Webpages         []MetasoResultItem `json:"webpages"`
	Total            int                `json:"total"`
}

// MetasoResultItem Metaso 搜索结果项
type MetasoResultItem struct {
	Title    string `json:"title"`
	Link     string `json:"link"`     // 注意：API 返回的是 link 不是 url
	Score    string `json:"score"`
	Snippet  string `json:"snippet"`
	Position int    `json:"position"`
	Date     string `json:"date,omitempty"`
	Authors  []string `json:"authors,omitempty"`
}

// Search 执行搜索
func (c *MetasoClient) Search(ctx context.Context, query string, size int) (*MetasoSearchResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("METASO_API_KEY is not configured")
	}

	// 构建请求体
	reqBody := MetasoSearchRequest{
		Q:                 query,
		Scope:             "webpage",
		IncludeSummary:    false,
		Size:              size,
		IncludeRawContent: false,
		ConciseSnippet:    false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiEndpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var result MetasoSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(body))
	}

	// 检查是否有结果
	if len(result.Webpages) == 0 {
		return nil, fmt.Errorf("API returned no results, total: %d", result.Total)
	}

	return &result, nil
}

// ========================================
// 工具实现
// ========================================

// 全局客户端（在初始化时设置）
var metasoClient *MetasoClient

// InitMetasoClient 初始化 Metaso 客户端
func InitMetasoClient(cfg *config.SearchConfig) {
	metasoClient = NewMetasoClient(cfg)
}

// SetMetasoClient 设置 Metaso 客户端（用于测试）
func SetMetasoClient(client *MetasoClient) {
	metasoClient = client
}

// NewWebSearchTool 创建网络搜索工具
func NewWebSearchTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"web_search",
		`在互联网上搜索实时信息。

适用于：
- 查询最新新闻、时事资讯
- 查找技术资料、教程
- 获取实时数据（天气、股票等）
- 补充知识库中缺失的信息

注意：搜索结果来自互联网，请谨慎验证信息准确性`,
		webSearch,
	)
}

// webSearch 执行网络搜索（普通函数）
func webSearch(ctx context.Context, req *WebSearchRequest) (*WebSearchResult, error) {
	if req.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if req.Limit <= 0 {
		req.Limit = 5
	}
	if req.Limit > 10 {
		req.Limit = 10 // 限制最多10条结果
	}

	// 如果没有配置客户端，返回模拟数据
	if metasoClient == nil {
		return mockSearchWeb(ctx, req.Query, req.Limit)
	}

	// 调用真实的 Metaso API
	result, err := metasoClient.Search(ctx, req.Query, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// 转换响应格式
	items := make([]SearchItem, 0, len(result.Webpages))
	for _, r := range result.Webpages {
		snippet := r.Snippet
		if snippet == "" {
			snippet = "无摘要"
		}

		items = append(items, SearchItem{
			Title:   r.Title,
			URL:     r.Link,
			Snippet: snippet,
		})
	}

	return &WebSearchResult{
		Items: items,
		Count: len(items),
		Query: req.Query,
	}, nil
}

// mockSearchWeb 模拟搜索（API 调用失败时的降级方案）
func mockSearchWeb(ctx context.Context, query string, limit int) (*WebSearchResult, error) {
	// 模拟搜索结果
	items := []SearchItem{
		{
			Title:   fmt.Sprintf("关于\"%s\"的搜索结果 1", query),
			URL:     "https://example.com/result1",
			Snippet: fmt.Sprintf("这是关于 %s 的第一个搜索结果的摘要内容...", query),
		},
		{
			Title:   fmt.Sprintf("关于\"%s\"的搜索结果 2", query),
			URL:     "https://example.com/result2",
			Snippet: fmt.Sprintf("这是关于 %s 的第二个搜索结果的摘要内容...", query),
		},
	}

	if limit < len(items) {
		items = items[:limit]
	}

	return &WebSearchResult{
		Items: items,
		Count: len(items),
		Query: query,
	}, nil
}

// ========================================
// 时间工具
// ========================================

// NewGetCurrentTimeTool 创建获取当前时间工具
func NewGetCurrentTimeTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"get_current_time",
		`获取当前日期和时间。

适用于：
- 需要知道当前时间
- 计算时间差
- 生成时间戳

返回格式：RFC3339 格式的时间字符串`,
		getCurrentTime,
	)
}

// GetCurrentTimeResult 时间查询结果
type GetCurrentTimeResult struct {
	Time string `json:"time"`
}

// getCurrentTime 获取当前时间（普通函数）
func getCurrentTime(ctx context.Context, req struct{}) (*GetCurrentTimeResult, error) {
	return &GetCurrentTimeResult{
		Time: time.Now().Format(time.RFC3339),
	}, nil
}

// ========================================
// 计算器工具
// ========================================

// CalculatorRequest 计算器请求
type CalculatorRequest struct {
	Expression string `json:"expression" jsonschema:"required,description=要计算的数学表达式，支持加减乘除和括号，例如: (1+2)*3/4"`
}

// CalculatorResult 计算器结果
type CalculatorResult struct {
	Expression string  `json:"expression"`
	Result     float64 `json:"result"`
}

// NewCalculatorTool 创建计算器工具
func NewCalculatorTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"calculator",
		`执行数学计算。

支持：
- 基本运算：+, -, *, /
- 括号：( )
- 小数运算

注意：仅用于简单计算，复杂计算请使用专门的工具`,
		calculator,
	)
}

// calculator 执行计算（普通函数）
func calculator(ctx context.Context, req *CalculatorRequest) (*CalculatorResult, error) {
	// TODO: 实际项目中应该使用安全的表达式解析器
	// 可以使用 github.com/Knetic/govaluate 或类似库
	//
	// import "github.com/Knetic/govaluate"
	//
	// expr, err := govaluate.NewEvaluableExpression(req.Expression)
	// if err != nil {
	//     return nil, err
	// }
	// result, err := expr.Evaluate(nil)
	// if err != nil {
	//     return nil, err
	// }

	// 简化版：返回示例结果
	return &CalculatorResult{
		Expression: req.Expression,
		Result:     42.0, // 实际应该是计算结果
	}, nil
}

// ========================================
// HTTP 请求工具
// ========================================

// HttpRequestRequest HTTP 请求参数
type HttpRequestRequest struct {
	URL     string            `json:"url" jsonschema:"required,description=请求的URL"`
	Method  string            `json:"method" jsonschema:"description=HTTP方法,default=GET,enum=GET,enum=POST,enum=PUT,enum=DELETE"`
	Headers map[string]string `json:"headers" jsonschema:"description=请求头"`
	Body    string            `json:"body" jsonschema:"description=请求体(仅POST/PUT时使用)"`
}

// HttpRequestResult HTTP 请求结果
type HttpRequestResult struct {
	StatusCode int               `json:"status_code"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
}

// NewHttpRequestTool 创建 HTTP 请求工具
func NewHttpRequestTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"http_request",
		`发送 HTTP 请求并获取响应。

适用于：
- 调用外部 API
- 获取网页内容
- 与第三方服务集成

注意：
- 请确保目标URL安全可靠
- 遵守目标网站的 robots.txt 和使用条款
- 敏感信息不要在 URL 中传递`,
		httpRequest,
	)
}

// httpRequest 发送 HTTP 请求（普通函数）
func httpRequest(ctx context.Context, req *HttpRequestRequest) (*HttpRequestResult, error) {
	if req.URL == "" {
		return nil, fmt.Errorf("url cannot be empty")
	}

	if req.Method == "" {
		req.Method = "GET"
	}

	// TODO: 实际项目中使用 http.Client 发送请求
	//
	// import "net/http"
	//
	// client := &http.Client{Timeout: 10 * time.Second}
	// httpRequest, err := http.NewRequestWithContext(ctx, req.Method, req.URL, strings.NewReader(req.Body))
	// if err != nil {
	//     return nil, err
	// }
	// for k, v := range req.Headers {
	//     httpRequest.Header.Set(k, v)
	// }
	// resp, err := client.Do(httpRequest)
	// ...

	return &HttpRequestResult{
		StatusCode: 200,
		Body:       "示例响应内容",
		Headers: map[string]string{
			"content-type": "application/json",
		},
	}, nil
}
