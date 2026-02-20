package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"link/internal/config"
	"link/internal/models/chat"
)

// QueryStrengthener 查询增强器
// 提供查询重写、查询拆分等前置优化功能
type QueryStrengthener struct {
	chatModel  chat.Chat
	chatConfig *chat.ChatConfig
}

// NewQueryStrengthener 创建查询增强器
func NewQueryStrengthener(chatConfig *chat.ChatConfig) (*QueryStrengthener, error) {
	model, err := chat.NewChat(&chat.ChatConfig{
		Source:    chatConfig.Source,
		BaseURL:   chatConfig.BaseURL,
		ModelName: chatConfig.ModelName,
		APIKey:    chatConfig.APIKey,
		Provider:  chatConfig.Provider,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	return &QueryStrengthener{
		chatModel:  model,
		chatConfig: chatConfig,
	}, nil
}

// ========================================
// 查询增强选项
// ========================================

// StrengthOptions 增强选项
type StrengthOptions struct {
	EnableRewrite bool    // 是否启用查询重写
	EnableSplit   bool    // 是否启用查询拆分
	Temperature   float64 // LLM 温度参数
	MaxTokens     int     // 最大 token 数
}

// DefaultStrengthOptions 默认增强选项
func DefaultStrengthOptions() *StrengthOptions {
	return &StrengthOptions{
		EnableRewrite: true,
		EnableSplit:   true,
		Temperature:   0.1, // 低温度保证稳定性
		MaxTokens:     2000,
	}
}

// ========================================
// 查询增强结果
// ========================================

// StrengthenedQuery 增强后的查询
type StrengthenedQuery struct {
	OriginalQuery  string   // 原始查询
	RewrittenQuery string   // 重写后的查询
	SubQueries     []string // 拆分的子查询
	RewriteApplied bool     // 是否应用了重写
	SplitApplied   bool     // 是否应用了拆分
	ProcessingTime int64    // 处理耗时（毫秒）
}

// ========================================
// 核心方法：增强查询
// ========================================

// StrengthenQuery 增强查询（主入口）
// 自动判断是否需要重写和拆分，并返回增强后的查询
func (s *QueryStrengthener) StrengthenQuery(
	ctx context.Context,
	query string,
	conversationHistory string,
	opts *StrengthOptions,
) (*StrengthenedQuery, error) {
	startTime := time.Now()

	if opts == nil {
		opts = DefaultStrengthOptions()
	}

	result := &StrengthenedQuery{
		OriginalQuery: query,
	}

	// 步骤1: 查询重写
	if opts.EnableRewrite && s.shouldRewrite(query) {
		rewritten, err := s.RewriteQuery(ctx, query, conversationHistory, opts)
		if err == nil && rewritten != "" {
			result.RewrittenQuery = rewritten
			result.RewriteApplied = true
		}
		// 重写失败时使用原查询，不中断流程
	}

	// 步骤2: 查询拆分（使用重写后的查询或原查询）
	queryToSplit := query
	if result.RewrittenQuery != "" {
		queryToSplit = result.RewrittenQuery
	}

	if opts.EnableSplit && s.shouldSplit(queryToSplit) {
		subQueries, err := s.SplitQuery(ctx, queryToSplit, conversationHistory, opts)
		if err == nil && len(subQueries) > 0 {
			result.SubQueries = subQueries
			result.SplitApplied = true
		}
		// 拆分失败时使用原查询，不中断流程
	}

	result.ProcessingTime = time.Since(startTime).Milliseconds()

	return result, nil
}

// ========================================
// 查询重写
// ========================================

// RewriteQuery 重写查询
// 使查询更加清晰、完整，便于检索
func (s *QueryStrengthener) RewriteQuery(
	ctx context.Context,
	query string,
	conversationHistory string,
	opts *StrengthOptions,
) (string, error) {
	// 加载查询重写模板
	template, err := config.LoadPromptTemplate("query/rewrite_user")
	if err != nil {
		return "", fmt.Errorf("failed to load rewrite template: %w", err)
	}

	// 创建提示词
	prompt := s.buildRewritePrompt(template, query, conversationHistory)

	// 调用 LLM
	response, err := s.callLLM(ctx, prompt, opts)
	if err != nil {
		return "", err
	}

	// 提取重写后的查询
	rewrittenQuery := s.extractRewrittenQuery(response)
	if rewrittenQuery == "" {
		return query, nil // 重写失败，返回原查询
	}

	return rewrittenQuery, nil
}

// buildRewritePrompt 创建重写提示词
func (s *QueryStrengthener) buildRewritePrompt(template, query, conversationHistory string) string {
	prompt := template

	// 替换历史对话
	if conversationHistory != "" {
		prompt = strings.Replace(prompt, "{{conversation}}", conversationHistory, 1)
	} else {
		// 如果没有历史，移除历史部分
		prompt = s.removeSection(prompt, "历史对话背景")
	}

	// 替换当前查询
	prompt = strings.Replace(prompt, "{{query}}", query, 1)

	return prompt
}

// extractRewrittenQuery 从 LLM 响应中提取重写后的查询
func (s *QueryStrengthener) extractRewrittenQuery(response string) string {
	response = strings.TrimSpace(response)

	// 尝试提取 JSON 格式
	if strings.Contains(response, "{") {
		re := regexp.MustCompile(`\{[^}]*"rewritten[^}]*":\s*"([^"]+)"`)
		matches := re.FindStringSubmatch(response)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
		re = regexp.MustCompile(`\{[^}]*"query[^}]*":\s*"([^"]+)"`)
		matches = re.FindStringSubmatch(response)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	// 尝试提取 "改写后的问题:" 后的内容
	re := regexp.MustCompile(`改写后的问题[:：]\s*\n*(.+?)(?:\n\n|$)`)
	matches := re.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// 尝试提取 markdown 代码块
	re = regexp.MustCompile("```(?:text|plain)?\n*(.+?)\n*```")
	matches = re.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// 如果响应简短且不是指令，直接使用
	if len(response) > 3 && len(response) < 200 &&
		!strings.Contains(response, "改写") &&
		!strings.Contains(response, "步骤") &&
		!strings.Contains(response, "注意") {
		return response
	}

	return ""
}

// ========================================
// 查询拆分
// ========================================

// SplitQuery 拆分查询
// 将复杂查询拆分为多个简单的子查询
func (s *QueryStrengthener) SplitQuery(
	ctx context.Context,
	query string,
	conversationHistory string,
	opts *StrengthOptions,
) ([]string, error) {
	// 加载查询拆分模板
	template, err := config.LoadPromptTemplate("query/query_split")
	if err != nil {
		return nil, fmt.Errorf("failed to load split template: %w", err)
	}

	// 创建提示词
	prompt := s.buildSplitPrompt(template, query, conversationHistory)

	// 调用 LLM
	response, err := s.callLLM(ctx, prompt, opts)
	if err != nil {
		return nil, err
	}

	// 提取拆分后的子查询
	subQueries := s.extractSubQueries(response)
	if len(subQueries) == 0 {
		return []string{query}, nil // 拆分失败，返回原查询
	}

	return subQueries, nil
}

// buildSplitPrompt 创建拆分提示词
func (s *QueryStrengthener) buildSplitPrompt(template, query, conversationHistory string) string {
	prompt := template

	// 替换历史对话
	if conversationHistory != "" {
		prompt = strings.Replace(prompt, "{{conversation}}", conversationHistory, 1)
	} else {
		prompt = s.removeSection(prompt, "历史对话背景")
	}

	// 替换当前查询
	prompt = strings.Replace(prompt, "{{query}}", query, 1)

	// 移除重写查询部分（如果存在）
	prompt = strings.Replace(prompt, "{{rewritten_query}}", "", 1)

	return prompt
}

// extractSubQueries 从 LLM 响应中提取子查询
func (s *QueryStrengthener) extractSubQueries(response string) []string {
	response = strings.TrimSpace(response)

	// 尝试解析 JSON 格式
	var result struct {
		SubQueries []struct {
			ID           string   `json:"id"`
			Query        string   `json:"query"`
			Description  string   `json:"description"`
			Order        int      `json:"order"`
			Dependencies []string `json:"dependencies"`
		} `json:"sub_queries"`
	}

	// 提取 JSON 部分
	re := regexp.MustCompile(`\{[\s\S]*sub_queries[\s\S]*\}`)
	jsonMatch := re.FindString(response)

	if err := json.Unmarshal([]byte(jsonMatch), &result); err == nil {
		if len(result.SubQueries) > 0 {
			// 按 order 排序
			subQueries := make([]string, len(result.SubQueries))
			for _, sq := range result.SubQueries {
				if sq.Order > 0 && sq.Order <= len(subQueries) {
					subQueries[sq.Order-1] = sq.Query
				}
			}
			// 过滤空查询
			var filtered []string
			for _, q := range subQueries {
				if q != "" {
					filtered = append(filtered, q)
				}
			}
			if len(filtered) > 0 {
				return filtered
			}
		}
	}

	// 尝试提取列表格式
	re = regexp.MustCompile(`(?:\d+\.|[-*])\s*["']?(.+?)["']?(?:\n|$)`)
	matches := re.FindAllStringSubmatch(response, -1)

	var subQueries []string
	for _, match := range matches {
		if len(match) > 1 {
			query := strings.TrimSpace(match[1])
			// 过滤掉非查询内容
			if query != "" &&
				!strings.HasPrefix(query, "步骤") &&
				!strings.HasPrefix(query, "注意") &&
				!strings.HasPrefix(query, "示例") &&
				len(query) > 2 {
				subQueries = append(subQueries, query)
			}
		}
	}

	if len(subQueries) > 0 {
		return subQueries
	}

	// 尝试按行分割，过滤掉说明文字
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过标题和说明
		if strings.HasPrefix(line, "#") ||
			strings.HasPrefix(line, "##") ||
			strings.HasPrefix(line, "-") ||
			strings.Contains(line, "拆分结果") ||
			strings.Contains(line, "子查询") {
			continue
		}
		if line != "" && len(line) > 2 && len(line) < 100 {
			subQueries = append(subQueries, line)
		}
	}

	return subQueries
}

// ========================================
// 判断逻辑
// ========================================

// shouldRewrite 判断是否需要重写查询
func (s *QueryStrengthener) shouldRewrite(query string) bool {
	// 过于简短的查询不需要重写
	if len(query) < 5 {
		return false
	}

	// 包含代词的查询需要重写
	pronouns := []string{"他", "她", "它", "他们", "它们", "这", "那", "这个", "那个", "这里", "那里"}
	for _, p := range pronouns {
		if strings.Contains(query, p) {
			return true
		}
	}

	// 包含省略号的查询需要重写
	if strings.Contains(query, "...") || strings.Contains(query, "…") {
		return true
	}

	// 问号开头的疑问句
	if strings.HasPrefix(strings.TrimSpace(query), "？") ||
		strings.HasPrefix(strings.TrimSpace(query), "?") {
		return true
	}

	// 过于简短的查询可能需要补充
	if len(query) < 15 && strings.Contains(query, "怎么") ||
		len(query) < 15 && strings.Contains(query, "如何") {
		return true
	}

	return false
}

// shouldSplit 判断是否需要拆分查询
func (s *QueryStrengthener) shouldSplit(query string) bool {
	query = strings.ToLower(query)

	// 包含多个问题的查询
	questionCount := 0
	questionWords := []string{"什么是", "如何", "怎么", "为什么", "哪个", "哪些",
		"what is", "how to", "why", "which", "list", "列出"}

	for _, word := range questionWords {
		if strings.Contains(query, word) {
			questionCount++
		}
	}

	if questionCount >= 2 {
		return true
	}

	// 包含"和"、"与"、"以及"等连接词，且长度较长
	if (strings.Contains(query, "和") || strings.Contains(query, "与") ||
		strings.Contains(query, "以及") || strings.Contains(query, "和")) &&
		len(query) > 20 {
		return true
	}

	// 包含"区别"、"对比"、"比较"等词
	comparisonWords := []string{"区别", "对比", "比较", "差异", "不同",
		"difference", "compare", "vs", "versus"}
	for _, word := range comparisonWords {
		if strings.Contains(query, word) {
			return true
		}
	}

	// 包含"列举"、"所有"等词
	if strings.Contains(query, "列举") || strings.Contains(query, "所有") {
		return true
	}

	// 包含"步骤"、"流程"等词，且查询较长
	if (strings.Contains(query, "步骤") || strings.Contains(query, "流程")) &&
		len(query) > 20 {
		return true
	}

	return false
}

// ========================================
// 辅助方法
// ========================================

// callLLM 调用 LLM
func (s *QueryStrengthener) callLLM(
	ctx context.Context,
	prompt string,
	opts *StrengthOptions,
) (string, error) {
	messages := []chat.Message{
		{Role: "user", Content: prompt},
	}

	chatOpts := &chat.ChatOptions{
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
	}

	response, err := s.chatModel.Chat(ctx, messages, chatOpts)
	if err != nil {
		return "", fmt.Errorf("LLM call failed: %w", err)
	}

	return response.Content, nil
}

// removeSection 移除提示词中的某个章节
func (s *QueryStrengthener) removeSection(prompt, sectionTitle string) string {
	re := regexp.MustCompile(`## ` + regexp.QuoteMeta(sectionTitle) + `[\s\S]*?(?=##|\Z)`)
	return re.ReplaceAllString(prompt, "")
}

// ========================================
// 批量增强查询
// ========================================

// BatchStrengthenQueries 批量增强查询
func (s *QueryStrengthener) BatchStrengthenQueries(
	ctx context.Context,
	queries []string,
	conversationHistory string,
	opts *StrengthOptions,
) ([]*StrengthenedQuery, error) {
	results := make([]*StrengthenedQuery, len(queries))

	for i, query := range queries {
		result, err := s.StrengthenQuery(ctx, query, conversationHistory, opts)
		if err != nil {
			// 单个失败不影响其他查询
			results[i] = &StrengthenedQuery{
				OriginalQuery: query,
			}
		} else {
			results[i] = result
		}
	}

	return results, nil
}

// ========================================
// 获取增强后的查询列表
// ========================================

// GetQueriesForRetrieve 获取用于检索的查询列表
// 优先使用拆分的子查询，其次使用重写后的查询，最后使用原查询
func (sq *StrengthenedQuery) GetQueriesForRetrieve() []string {
	var queries []string

	// 如果有拆分的子查询，使用子查询
	if len(sq.SubQueries) > 0 {
		queries = append(queries, sq.SubQueries...)
	}
	// 如果有重写后的查询，添加重写后的查询
	if sq.RewrittenQuery != "" {
		queries = append(queries, sq.RewrittenQuery)
	}
	// 始终包含原查询
	queries = append(queries, sq.OriginalQuery)

	return deduplicateQueries(queries)
}

// deduplicateQueries 去重查询列表
func deduplicateQueries(queries []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q != "" && !seen[q] {
			seen[q] = true
			result = append(result, q)
		}
	}

	return result
}

// ========================================
// 获取增强摘要
// ========================================

// GetSummary 获取增强操作的摘要信息
func (sq *StrengthenedQuery) GetSummary() map[string]interface{} {
	return map[string]interface{}{
		"original_query":     sq.OriginalQuery,
		"rewritten_query":    sq.RewrittenQuery,
		"sub_queries":        sq.SubQueries,
		"rewrite_applied":    sq.RewriteApplied,
		"split_applied":      sq.SplitApplied,
		"processing_time_ms": sq.ProcessingTime,
		"query_count":        len(sq.GetQueriesForRetrieve()),
	}
}
