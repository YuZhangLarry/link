package tool

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"link/internal/config"
	"link/internal/types"
	"link/internal/types/interfaces"
)

// ========================================
// 包级依赖管理
// ========================================

var (
	kbRepoInstance interfaces.KnowledgeBaseRepository
	kbRepoOnce     sync.Once
)

// InitKnowledgeBaseTool 初始化知识库工具的依赖
// 应该在应用启动时调用，传入实际的 repository 实现
func InitKnowledgeBaseTool(repo interfaces.KnowledgeBaseRepository) {
	kbRepoOnce.Do(func() {
		kbRepoInstance = repo
	})
}

// ========================================
// 请求/响应类型定义
// ========================================

// KbListRequestV2 知识库列表请求
type KbListRequestV2 struct {
	Status string `json:"status" jsonschema:"description=状态筛选：all(全部)/enabled(启用)/disabled(禁用)，默认all"`
}

// KbListResultV2 知识库列表结果
type KbListResultV2 struct {
	KnowledgeBases []KbInfoV2 `json:"knowledge_bases"`
	Count          int        `json:"count"`
	LatencyMs      int64      `json:"latency_ms"`
}

// KbInfoV2 知识库详细信息
type KbInfoV2 struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	DocumentCount int64  `json:"document_count"`
	ChunkCount    int64  `json:"chunk_count"`
	StorageSize   int64  `json:"storage_size"`
	IsPublic      bool   `json:"is_public"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// ========================================
// 工具创建函数
// ========================================

// NewKnowledgeBaseListTool 创建知识库列表工具
// 该工具可以查询当前租户下的所有知识库信息
func NewKnowledgeBaseListTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"kb_list",
		`获取当前租户下的所有知识库信息。

功能：
- 列出所有可用的知识库及其统计信息
- 获取知识库的配置详情
- 查看知识库中的文档和分块数量
- 支持按状态筛选

适用场景：
- 用户询问"有哪些知识库"
- 用户询问知识库统计信息
- 需要确定查询目标知识库时使用

参数：
- status: 状态筛选，可选值: all/enabled/disabled，默认 all`,
		listKnowledgeBases,
	)
}

// ========================================
// 工具执行逻辑
// ========================================

// listKnowledgeBases 查询知识库列表
func listKnowledgeBases(ctx context.Context, req *KbListRequestV2) (*KbListResultV2, error) {
	startTime := time.Now()

	// 1. 参数处理
	status := req.Status
	if status == "" {
		status = "all"
	}

	// 验证参数
	if status != "all" && status != "enabled" && status != "disabled" {
		return nil, fmt.Errorf("invalid status: %s, must be one of: all/enabled/disabled", status)
	}

	// 2. 检查依赖是否已初始化
	if kbRepoInstance == nil {
		return nil, fmt.Errorf("knowledge base repository not initialized, call InitKnowledgeBaseTool first")
	}

	// 3. 从上下文获取租户ID
	tenantID, err := getTenantIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取租户信息失败: %w", err)
	}

	// 4. 查询知识库列表
	knowledgeBases, err := queryKnowledgeBases(ctx, kbRepoInstance, tenantID, status)
	if err != nil {
		return nil, fmt.Errorf("查询知识库列表失败: %w", err)
	}

	// 5. 构建返回结果
	result := &KbListResultV2{
		KnowledgeBases: knowledgeBases,
		Count:          len(knowledgeBases),
		LatencyMs:      time.Since(startTime).Milliseconds(),
	}

	return result, nil
}

// queryKnowledgeBases 查询知识库列表（实际实现）
func queryKnowledgeBases(ctx context.Context, kbRepo interfaces.KnowledgeBaseRepository, tenantID int64, status string) ([]KbInfoV2, error) {
	// 设置分页参数（获取所有数据）
	page := 1
	pageSize := 1000

	// 查询知识库列表
	kbs, total, err := kbRepo.FindByTenantID(ctx, tenantID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询知识库失败: %w", err)
	}

	// 如果没有数据，返回空列表
	if total == 0 {
		return []KbInfoV2{}, nil
	}

	// 根据 status 筛选并转换为返回格式
	var results []KbInfoV2

	for _, kb := range kbs {
		// 跳过已删除的知识库
		if kb.DeletedAt != nil {
			continue
		}

		// 状态筛选
		kbStatus := "enabled"
		if kb.Status == 0 {
			kbStatus = "disabled"
		}

		if status != "all" && status != kbStatus {
			continue
		}

		// 转换为返回格式
		kbInfo := KbInfoV2{
			ID:            kb.ID,
			Name:          kb.Name,
			Description:   kb.Description,
			DocumentCount: int64(kb.DocumentCount),
			ChunkCount:    int64(kb.ChunkCount),
			StorageSize:   kb.StorageSize,
			IsPublic:      kb.IsPublic,
			Status:        kbStatus,
			CreatedAt:     kb.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     kb.UpdatedAt.Format(time.RFC3339),
		}

		results = append(results, kbInfo)
	}

	return results, nil
}

// ========================================
// 上下文辅助函数
// ========================================

var (
	// defaultTenantID 默认租户ID，用于测试环境或非多租户场景
	// 可通过 SetDefaultTenantID 修改
	defaultTenantID int64 = 7
)

// SetDefaultTenantID 设置默认租户ID
func SetDefaultTenantID(tid int64) {
	defaultTenantID = tid
}

// GetDefaultTenantID 获取默认租户ID
func GetDefaultTenantID() int64 {
	return defaultTenantID
}

// getTenantIDFromContext 从上下文中获取租户ID
// 如果上下文中没有租户ID，则返回默认租户ID（7）
func getTenantIDFromContext(ctx context.Context) (int64, error) {
	// 尝试从 context.Value 中获取 tenant_id
	// 这是通过 middleware.ContextToRequest() 中间件设置的
	if tid, ok := ctx.Value("tenant_id").(int64); ok && tid > 0 {
		return tid, nil
	}

	// 如果没有找到租户ID，返回默认租户ID
	// 这样可以支持测试环境和单租户场景
	if defaultTenantID > 0 {
		return defaultTenantID, nil
	}

	return 0, fmt.Errorf("tenant_id not found in context and no default tenant configured")
}

// ========================================
// 类型转换辅助函数
// ========================================

// typesToKbInfo 将 types.KnowledgeBase 转换为 KbInfoV2
func typesToKbInfo(kb *types.KnowledgeBase) KbInfoV2 {
	status := "enabled"
	if kb.Status == 0 {
		status = "disabled"
	}

	return KbInfoV2{
		ID:            kb.ID,
		Name:          kb.Name,
		Description:   kb.Description,
		DocumentCount: int64(kb.DocumentCount),
		ChunkCount:    int64(kb.ChunkCount),
		StorageSize:   kb.StorageSize,
		IsPublic:      kb.IsPublic,
		Status:        status,
		CreatedAt:     kb.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     kb.UpdatedAt.Format(time.RFC3339),
	}
}

// ========================================
// 工具工厂（支持依赖注入）
// ========================================

// KnowledgeBaseToolFactory 工具工厂
type KnowledgeBaseToolFactory struct {
	kbRepo interfaces.KnowledgeBaseRepository
}

// NewKnowledgeBaseToolFactory 创建工具工厂
func NewKnowledgeBaseToolFactory(kbRepo interfaces.KnowledgeBaseRepository) *KnowledgeBaseToolFactory {
	return &KnowledgeBaseToolFactory{
		kbRepo: kbRepo,
	}
}

// CreateToolUsingFactory 使用工厂创建工具
// 此方法会设置包级变量，然后返回使用 utils.InferTool 创建的工具
func (f *KnowledgeBaseToolFactory) CreateToolUsingFactory() (tool.InvokableTool, error) {
	InitKnowledgeBaseTool(f.kbRepo)
	return NewKnowledgeBaseListTool()
}

// ========================================
// 智能检索工具 - Smart Retrieval Tool
// ========================================
// 这个工具会：
// 1. 查询所有启用的知识库
// 2. 匹配与 query 最相关的知识库
// 3. 对匹配的知识库进行 RAG 检索
// 4. 判断是否需要网络搜索
// 5. 合并所有检索结果返回

// SmartRetrievalRequest 智能检索请求
type SmartRetrievalRequest struct {
	// Query 用户查询
	Query string `json:"query" jsonschema:"required,description=用户的问题或查询内容"`

	// TopK 每个知识库返回的片段数量，默认5
	TopK int `json:"top_k" jsonschema:"description=每个知识库返回的片段数量，默认5，范围1-10"`

	// EnableWebSearch 是否启用网络搜索，默认true
	EnableWebSearch bool `json:"enable_web_search" jsonschema:"description=是否启用网络搜索补充信息，默认true"`

	// WebSearchLimit 网络搜索结果数量，默认3
	WebSearchLimit int `json:"web_search_limit" jsonschema:"description=网络搜索结果数量，默认3，范围1-5"`

	// RetrievalMode 检索模式，默认hybrid
	RetrievalMode string `json:"retrieval_mode" jsonschema:"description=检索模式：vector/bm25/hybrid/graph，默认hybrid"`
}

// SmartRetrievalResult 智能检索结果
type SmartRetrievalResult struct {
	// Query 原始查询
	Query string `json:"query"`

	// MatchedKnowledgeBases 匹配的知识库列表
	MatchedKnowledgeBases []MatchedKBInfo `json:"matched_knowledge_bases"`

	// TotalChunks 检索到的总片段数
	TotalChunks int `json:"total_chunks"`

	// TopChunks 最重要的片段（合并后的top结果）
	TopChunks []DocumentChunk `json:"top_chunks"`

	// WebSearchResults 网络搜索结果（如果有）
	WebSearchResults *WebSearchResult `json:"web_search_results,omitempty"`

	// Summary 检索结果摘要
	Summary string `json:"summary"`

	// NeedsWebSearch 是否需要网络搜索
	NeedsWebSearch bool `json:"needs_web_search"`

	// LatencyMs 总耗时
	LatencyMs int64 `json:"latency_ms"`
}

// MatchedKBInfo 匹配的知识库信息
type MatchedKBInfo struct {
	KBID       string          `json:"kb_id"`
	Name       string          `json:"name"`
	MatchScore float64         `json:"match_score"`
	ChunkCount int             `json:"chunk_count"`
	Chunks     []DocumentChunk `json:"chunks"`
}

// NewSmartRetrievalTool 创建智能检索工具
func NewSmartRetrievalTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"smart_retrieval",
		`智能检索工具 - 自动匹配相关知识库并进行综合检索。

这是一个高级检索工具，能够：
1. 自动分析查询内容，匹配最相关的知识库
2. 对多个知识库进行并行检索
3. 智能判断是否需要网络搜索补充
4. 合并去重，返回最相关的结果

工作流程：
1. 分析查询，提取关键词
2. 查询所有启用的知识库
3. 计算每个知识库与查询的相关度
4. 对高相关度的知识库进行RAG检索
5. 检索结果不足时，自动进行网络搜索
6. 合并所有结果，按相关度排序

适用场景：
- 不确定哪个知识库包含相关信息
- 需要从多个知识库综合查询
- 需要获取最新信息补充知识库内容
- 复杂问题需要多源信息支撑

参数说明：
- query: 查询内容（必需）
- top_k: 每个知识库返回的片段数（可选，默认5）
- enable_web_search: 是否启用网络搜索（可选，默认true）
- web_search_limit: 网络搜索结果数（可选，默认3）
- retrieval_mode: 检索模式（可选，默认hybrid）`,
		smartRetrieval,
	)
}

// smartRetrieval 执行智能检索
func smartRetrieval(ctx context.Context, req *SmartRetrievalRequest) (*SmartRetrievalResult, error) {
	startTime := time.Now()

	// 1. 参数验证
	if req.Query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	// 设置默认值
	if req.TopK <= 0 {
		req.TopK = 5
	}
	if req.TopK > 10 {
		req.TopK = 10
	}
	if req.WebSearchLimit <= 0 {
		req.WebSearchLimit = 3
	}
	if req.WebSearchLimit > 5 {
		req.WebSearchLimit = 5
	}
	if req.RetrievalMode == "" {
		req.RetrievalMode = "hybrid"
	}

	// 2. 获取租户ID
	tenantID, err := getTenantIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取租户信息失败: %w", err)
	}

	// 3. 检查知识库 repository 是否已初始化
	if kbRepoInstance == nil {
		return nil, fmt.Errorf("knowledge base repository not initialized")
	}

	// 4. 查询所有启用的知识库
	knowledgeBases, _, err := kbRepoInstance.FindByTenantID(ctx, tenantID, 1, 100)
	if err != nil {
		return nil, fmt.Errorf("查询知识库列表失败: %w", err)
	}

	// 5. 匹配相关知识库
	matchedKBs := matchKnowledgeBases(req.Query, knowledgeBases)

	if len(matchedKBs) == 0 {
		// 没有匹配的知识库，直接使用网络搜索
		return handleNoMatchFound(ctx, req, startTime)
	}

	// 6. 对匹配的知识库进行并行检索
	allChunks := make([]DocumentChunk, 0)
	matchedKBInfo := make([]MatchedKBInfo, 0, len(matchedKBs))

	for _, kb := range matchedKBs {
		// 将知识库ID转换为int64（RAG服务需要）
		kbIDInt := parseKBIDToInt(kb.ID)
		if kbIDInt == 0 {
			continue
		}

		// 调用RAG检索
		chunks, err := performRAGQuery(ctx, kbIDInt, req.Query, req.TopK, req.RetrievalMode)
		if err != nil {
			// 记录错误但继续处理其他知识库
			continue
		}

		if len(chunks) > 0 {
			matchedKBInfo = append(matchedKBInfo, MatchedKBInfo{
				KBID:       kb.ID,
				Name:       kb.Name,
				MatchScore: kb.matchScore,
				ChunkCount: len(chunks),
				Chunks:     chunks,
			})
			allChunks = append(allChunks, chunks...)
		}
	}

	// 7. 判断是否需要网络搜索
	needsWebSearch := req.EnableWebSearch && shouldPerformWebSearch(req.Query, allChunks)

	// 8. 网络搜索（如果需要）
	var webSearchResults *WebSearchResult
	if needsWebSearch {
		webSearchResults, err = performWebSearch(ctx, req.Query, req.WebSearchLimit)
		if err != nil {
			// 网络搜索失败，不影响主流程
			webSearchResults = nil
		}
	}

	// 9. 合并和排序结果
	topChunks := mergeAndRankChunks(allChunks, req.TopK*2)

	// 10. 生成摘要
	summary := generateRetrievalSummary(req.Query, matchedKBInfo, topChunks, webSearchResults)

	// 11. 构建返回结果
	result := &SmartRetrievalResult{
		Query:                 req.Query,
		MatchedKnowledgeBases: matchedKBInfo,
		TotalChunks:           len(allChunks),
		TopChunks:             topChunks,
		WebSearchResults:      webSearchResults,
		Summary:               summary,
		NeedsWebSearch:        needsWebSearch,
		LatencyMs:             time.Since(startTime).Milliseconds(),
	}

	return result, nil
}

// ========================================
// 知识库匹配逻辑
// ========================================

// matchedKB 内部使用的匹配知识库结构
type matchedKB struct {
	*types.KnowledgeBase
	matchScore float64
}

// matchKnowledgeBases 匹配与查询最相关的知识库
func matchKnowledgeBases(query string, kbs []*types.KnowledgeBase) []matchedKB {
	// 提取查询关键词
	keywords := extractKeywords(query)

	var matched []matchedKB

	for _, kb := range kbs {
		// 跳过已删除和禁用的知识库
		if kb.DeletedAt != nil || kb.Status == 0 {
			continue
		}

		// 计算匹配分数
		score := calculateKBMatchScore(query, keywords, kb)

		// 设置阈值，只保留相关度较高的知识库
		if score > 0.1 {
			matched = append(matched, matchedKB{
				KnowledgeBase: kb,
				matchScore:    score,
			})
		}
	}

	// 按匹配分数排序
	sortMatchedKBs(matched)

	// 最多返回前5个最相关的知识库
	if len(matched) > 5 {
		matched = matched[:5]
	}

	return matched
}

// extractKeywords 从查询中提取关键词
func extractKeywords(query string) []string {
	// 简单的关键词提取：按空格和标点分割
	query = strings.ToLower(query)

	// 移除常见停用词
	stopWords := map[string]bool{
		"的": true, "了": true, "是": true, "在": true, "有": true,
		"和": true, "与": true, "或": true, "但": true, "而": true,
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"what": true, "how": true, "when": true, "where": true, "who": true,
	}

	// 分割并过滤
	words := strings.Fields(query)
	var keywords []string
	for _, word := range words {
		word = strings.Trim(word, ".,!?;:，。！？；：")
		if len(word) > 1 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// calculateKBMatchScore 计算知识库与查询的匹配分数
func calculateKBMatchScore(query string, keywords []string, kb *types.KnowledgeBase) float64 {
	score := 0.0

	// 1. 检查名称匹配
	kbName := strings.ToLower(kb.Name)
	queryLower := strings.ToLower(query)

	// 完全匹配
	if kbName == queryLower {
		score += 1.0
	} else if strings.Contains(kbName, queryLower) || strings.Contains(queryLower, kbName) {
		score += 0.5
	}

	// 2. 检查描述匹配
	if kb.Description != "" {
		desc := strings.ToLower(kb.Description)
		for _, keyword := range keywords {
			if strings.Contains(desc, keyword) {
				score += 0.2
			}
		}
	}

	// 3. 检查关键词匹配
	for _, keyword := range keywords {
		if strings.Contains(kbName, keyword) {
			score += 0.3
		}
	}

	return score
}

// sortMatchedKBs 按匹配分数排序
func sortMatchedKBs(kbs []matchedKB) {
	// 简单冒泡排序
	n := len(kbs)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if kbs[j].matchScore < kbs[j+1].matchScore {
				kbs[j], kbs[j+1] = kbs[j+1], kbs[j]
			}
		}
	}
}

// ========================================
// RAG 检索逻辑
// ========================================

// performRAGQuery 执行RAG检索
func performRAGQuery(ctx context.Context, kbID int64, query string, topK int, mode string) ([]DocumentChunk, error) {
	// 如果RAG服务未初始化，返回空结果
	if ragService == nil {
		return nil, fmt.Errorf("RAG service not initialized")
	}

	// 构建请求
	req := &RAGQueryRequest{
		Query:         query,
		KBID:          kbID,
		TopK:          topK,
		RetrievalMode: mode,
		MinScore:      0.5,
	}

	// 调用服务
	result, err := ragService.Query(ctx, req)
	if err != nil {
		return nil, err
	}

	return result.Chunks, nil
}

// parseKBIDToInt 将知识库ID字符串转换为int64
func parseKBIDToInt(kbID string) int64 {
	// 知识库ID可能是字符串格式，尝试转换
	var id int64
	if _, err := fmt.Sscanf(kbID, "%d", &id); err == nil {
		return id
	}
	return 0
}

// ========================================
// 网络搜索逻辑
// ========================================

// shouldPerformWebSearch 判断是否需要网络搜索
func shouldPerformWebSearch(query string, chunks []DocumentChunk) bool {
	// 1. 如果没有检索到任何结果，需要网络搜索
	if len(chunks) == 0 {
		return true
	}

	// 2. 检查平均相似度
	totalScore := 0.0
	for _, chunk := range chunks {
		totalScore += chunk.Score
	}
	avgScore := totalScore / float64(len(chunks))

	// 平均相似度低于0.6，需要网络搜索
	if avgScore < 0.6 {
		return true
	}

	// 3. 检查查询是否包含时间敏感词
	timeSensitiveKeywords := []string{
		"最新", "今天", "最近", "现在", "当前", "2024", "2025", "2026",
		"latest", "today", "recent", "now", "current", "news", "价格", "price",
	}

	queryLower := strings.ToLower(query)
	for _, keyword := range timeSensitiveKeywords {
		if strings.Contains(queryLower, keyword) {
			return true
		}
	}

	return false
}

// performWebSearch 执行网络搜索
func performWebSearch(ctx context.Context, query string, limit int) (*WebSearchResult, error) {
	if metasoClient == nil {
		// 客户端未初始化，返回模拟结果
		return mockSearchWeb(ctx, query, limit)
	}

	result, err := metasoClient.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	// 转换格式
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
		Query: query,
	}, nil
}

// ========================================
// 结果合并和摘要
// ========================================

// mergeAndRankChunks 合并和排序片段
func mergeAndRankChunks(chunks []DocumentChunk, topN int) []DocumentChunk {
	if len(chunks) == 0 {
		return chunks
	}

	// 按相似度排序（简单冒泡排序）
	n := len(chunks)
	sorted := make([]DocumentChunk, n)
	copy(sorted, chunks)

	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if sorted[j].Score < sorted[j+1].Score {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	// 去重（基于内容相似度）
	uniqueChunks := make([]DocumentChunk, 0, n)
	seen := make(map[string]bool)

	for _, chunk := range sorted {
		// 使用内容的前50个字符作为去重标识
		key := chunk.Content
		if len(key) > 50 {
			key = key[:50]
		}
		key = strings.TrimSpace(key)

		if !seen[key] {
			seen[key] = true
			uniqueChunks = append(uniqueChunks, chunk)
		}
	}

	// 返回topN
	if len(uniqueChunks) > topN {
		uniqueChunks = uniqueChunks[:topN]
	}

	return uniqueChunks
}

// generateRetrievalSummary 生成检索摘要
func generateRetrievalSummary(query string, matchedKBs []MatchedKBInfo, chunks []DocumentChunk, webResults *WebSearchResult) string {
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("## 检索摘要\n\n**查询**: %s\n\n", query))

	// 匹配的知识库
	if len(matchedKBs) > 0 {
		summary.WriteString("**匹配的知识库**:\n")
		for i, kb := range matchedKBs {
			summary.WriteString(fmt.Sprintf("%d. %s (相关度: %.2f, %d个片段)\n",
				i+1, kb.Name, kb.MatchScore, kb.ChunkCount))
		}
		summary.WriteString("\n")
	}

	// 检索统计
	summary.WriteString(fmt.Sprintf("**检索统计**: 共检索到 %d 个相关片段\n\n", len(chunks)))

	// 网络搜索
	if webResults != nil && webResults.Count > 0 {
		summary.WriteString(fmt.Sprintf("**网络搜索**: 补充了 %d 条网络结果\n\n", webResults.Count))
	}

	// 关键内容预览
	if len(chunks) > 0 {
		summary.WriteString("**关键内容**:\n\n")
		for i, chunk := range chunks {
			if i >= 3 { // 最多显示3个片段
				break
			}
			// 截取内容预览
			preview := chunk.Content
			if len(preview) > 150 {
				preview = preview[:150] + "..."
			}
			summary.WriteString(fmt.Sprintf("%d. %s\n", i+1, preview))
		}
	}

	return summary.String()
}

// ========================================
// 无匹配结果处理
// ========================================

// handleNoMatchFound 处理没有找到匹配知识库的情况
func handleNoMatchFound(ctx context.Context, req *SmartRetrievalRequest, startTime time.Time) (*SmartRetrievalResult, error) {
	var webResults *WebSearchResult
	var err error

	// 直接进行网络搜索
	if req.EnableWebSearch {
		webResults, err = performWebSearch(ctx, req.Query, req.WebSearchLimit)
		if err != nil {
			webResults = nil
		}
	}

	// 生成摘要
	summary := fmt.Sprintf("未找到匹配的知识库，使用网络搜索获取信息。", req.Query)
	if webResults != nil && webResults.Count > 0 {
		summary += fmt.Sprintf("\n\n网络搜索找到 %d 条相关结果。", webResults.Count)
	}

	result := &SmartRetrievalResult{
		Query:                 req.Query,
		MatchedKnowledgeBases: []MatchedKBInfo{},
		TotalChunks:           0,
		TopChunks:             []DocumentChunk{},
		WebSearchResults:      webResults,
		Summary:               summary,
		NeedsWebSearch:        true,
		LatencyMs:             time.Since(startTime).Milliseconds(),
	}

	return result, nil
}

// ========================================
// 智能检索工具工厂
// ========================================

// SmartRetrievalToolFactory 智能检索工具工厂
type SmartRetrievalToolFactory struct {
	kbRepo    interfaces.KnowledgeBaseRepository
	ragSvc    RAGQueryService
	searchCfg *config.SearchConfig
}

// NewSmartRetrievalToolFactory 创建智能检索工具工厂
func NewSmartRetrievalToolFactory(
	kbRepo interfaces.KnowledgeBaseRepository,
	ragSvc RAGQueryService,
	searchCfg *config.SearchConfig,
) *SmartRetrievalToolFactory {
	return &SmartRetrievalToolFactory{
		kbRepo:    kbRepo,
		ragSvc:    ragSvc,
		searchCfg: searchCfg,
	}
}

// CreateTool 创建工具
func (f *SmartRetrievalToolFactory) CreateTool() (tool.InvokableTool, error) {
	// 初始化依赖
	InitKnowledgeBaseTool(f.kbRepo)
	if f.ragSvc != nil {
		InitRAGQueryTool(f.ragSvc)
	}
	if f.searchCfg != nil {
		InitGlobalSearchClient(f.searchCfg)
		SetMetasoClient(GetGlobalSearchClient())
	}

	return NewSmartRetrievalTool()
}
