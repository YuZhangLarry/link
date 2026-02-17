package rag

import (
	"context"
	"fmt"
	"sort"
)

// ========================================
// 重排策略接口
// ========================================

// RerankStrategy 重排策略接口
type RerankStrategy interface {
	// Rerank 执行重排，返回重排后的结果
	Rerank(ctx context.Context, query string, resultLists [][]*RetrieveResult) ([]*RetrieveResult, error)

	// Name 返回策略名称
	Name() string
}

// ========================================
// 重排配置
// ========================================

// RerankOptions 重排选项
type RerankOptions struct {
	TopK        int     // 返回结果数量
	Strategy    string  // 重排策略：rrf, weighted, model, weighted_rrf
	Alpha       float32 // 向量检索权重（用于加权融合）
	Beta        float32 // 关键词检索权重（用于加权融合）
	Gamma       float32 // 图谱检索权重（用于加权融合）
	RerankModel string  // 重排模型名称（如果使用模型重排）
	RerankTopK  int     // 重排前保留的候选数量
}

// DefaultRerankOptions 默认重排选项
func DefaultRerankOptions() *RerankOptions {
	return &RerankOptions{
		TopK:       5,
		Strategy:   "rrf",
		Alpha:      0.5,
		Beta:       0.3,
		Gamma:      0.2,
		RerankTopK: 20,
	}
}

// ========================================
// Reranker 重排器
// ========================================

// Reranker 重排器，支持多种重排策略
type Reranker struct {
	strategies map[string]RerankStrategy
	embedder   RerankEmbedder // 可选：用于模型重排
}

// RerankEmbedder 重排模型接口
type RerankEmbedder interface {
	// Rerank 计算查询与文档的相关性分数
	// 返回文档索引对应的分数列表
	Rerank(ctx context.Context, query string, documents []string) ([]float32, error)
}

// NewReranker 创建重排器
func NewReranker() *Reranker {
	r := &Reranker{
		strategies: make(map[string]RerankStrategy),
	}

	// 注册默认策略
	r.RegisterStrategy(&RRFStrategy{})
	r.RegisterStrategy(&WeightedFusionStrategy{})
	r.RegisterStrategy(&WeightedRRFStrategy{})
	r.RegisterStrategy(&ModelRerankStrategy{})

	return r
}

// NewRerankerWithEmbedder 创建带重排模型的重排器
func NewRerankerWithEmbedder(embedder RerankEmbedder) *Reranker {
	r := NewReranker()
	r.embedder = embedder

	// 更新模型策略的 embedder
	if strategy, ok := r.strategies["model"]; ok {
		if mr, ok := strategy.(*ModelRerankStrategy); ok {
			mr.SetEmbedder(embedder)
		}
	}

	return r
}

// RegisterStrategy 注册重排策略
func (r *Reranker) RegisterStrategy(strategy RerankStrategy) {
	r.strategies[strategy.Name()] = strategy
}

// Rerank 执行重排（主入口）
func (r *Reranker) Rerank(ctx context.Context, query string, resultLists [][]*RetrieveResult, opts *RerankOptions) ([]*RetrieveResult, error) {
	if opts == nil {
		opts = DefaultRerankOptions()
	}

	// 过滤空列表
	validLists := make([][]*RetrieveResult, 0, len(resultLists))
	for _, list := range resultLists {
		if len(list) > 0 {
			validLists = append(validLists, list)
		}
	}

	if len(validLists) == 0 {
		return []*RetrieveResult{}, nil
	}

	// 只有一个结果列表，不需要重排
	if len(validLists) == 1 {
		return sortAndTrimResults(validLists[0], opts.TopK), nil
	}

	// 获取策略
	strategy, ok := r.strategies[opts.Strategy]
	if !ok {
		return nil, fmt.Errorf("未知的重排策略: %s", opts.Strategy)
	}

	// 执行重排
	results, err := strategy.Rerank(ctx, query, validLists)
	if err != nil {
		return nil, fmt.Errorf("重排失败: %w", err)
	}

	// 截取 TopK
	return sortAndTrimResults(results, opts.TopK), nil
}

// ========================================
// RRF 策略（倒数排名融合）
// ========================================

// RRFStrategy Reciprocal Rank Fusion 策略
type RRFStrategy struct {
	K float64 // RRF 常数，默认 60
}

// Name 返回策略名称
func (s *RRFStrategy) Name() string {
	return "rrf"
}

// Rerank 执行 RRF 重排
func (s *RRFStrategy) Rerank(ctx context.Context, query string, resultLists [][]*RetrieveResult) ([]*RetrieveResult, error) {
	k := s.K
	if k == 0 {
		k = 60.0
	}

	// 构建分块 ID 到分数的映射
	scoreMap := make(map[string]float32)
	dataMap := make(map[string]*RetrieveResult)

	// 对每个结果列表计算 RRF 分数
	for _, results := range resultLists {
		// 为每个列表分配权重（这里简单均分，可以根据需要调整）
		weight := float32(1.0) / float32(len(resultLists))

		for rank, result := range results {
			if _, exists := dataMap[result.ChunkID]; !exists {
				dataMap[result.ChunkID] = result
			}
			// RRF 公式: weight / (k + rank)
			scoreMap[result.ChunkID] += weight / float32(k+float64(rank+1))
		}
	}

	// 构建重排结果
	reranked := make([]*RetrieveResult, 0, len(dataMap))
	for chunkID, score := range scoreMap {
		result := dataMap[chunkID]
		newResult := &RetrieveResult{
			ChunkID:       result.ChunkID,
			KnowledgeID:   result.KnowledgeID,
			KBID:          result.KBID,
			Content:       result.Content,
			ChunkIndex:    result.ChunkIndex,
			Score:         score,
			MatchType:     inferMatchType(resultLists, chunkID),
			StartPosition: result.StartPosition,
			EndPosition:   result.EndPosition,
		}
		reranked = append(reranked, newResult)
	}

	return reranked, nil
}

// ========================================
// 加权分数融合策略
// ========================================

// WeightedFusionStrategy 加权分数融合策略
// 直接对各个检索器的原始分数进行加权融合
type WeightedFusionStrategy struct{}

// Name 返回策略名称
func (s *WeightedFusionStrategy) Name() string {
	return "weighted"
}

// Rerank 执行加权分数融合
func (s *WeightedFusionStrategy) Rerank(ctx context.Context, query string, resultLists [][]*RetrieveResult) ([]*RetrieveResult, error) {
	// 定义默认权重：两路或三路平均分配
	weights := make([]float32, len(resultLists))
	for i := range weights {
		weights[i] = 1.0 / float32(len(resultLists))
	}

	// 归一化分数（因为不同检索器的分数范围可能不同）
	normalizedLists := make([][]*RetrieveResult, len(resultLists))
	for i, results := range resultLists {
		normalizedLists[i] = normalizeScores(results)
	}

	// 构建分块 ID 到分数的映射
	scoreMap := make(map[string]float32)
	dataMap := make(map[string]*RetrieveResult)

	for listIdx, results := range normalizedLists {
		weight := weights[listIdx]

		for _, result := range results {
			if _, exists := dataMap[result.ChunkID]; !exists {
				dataMap[result.ChunkID] = result
			}
			scoreMap[result.ChunkID] += result.Score * weight
		}
	}

	// 构建重排结果
	reranked := make([]*RetrieveResult, 0, len(dataMap))
	for chunkID, score := range scoreMap {
		result := dataMap[chunkID]
		newResult := &RetrieveResult{
			ChunkID:       result.ChunkID,
			KnowledgeID:   result.KnowledgeID,
			KBID:          result.KBID,
			Content:       result.Content,
			ChunkIndex:    result.ChunkIndex,
			Score:         score,
			MatchType:     inferMatchType(resultLists, chunkID),
			StartPosition: result.StartPosition,
			EndPosition:   result.EndPosition,
		}
		reranked = append(reranked, newResult)
	}

	return reranked, nil
}

// ========================================
// 加权 RRF 策略
// ========================================

// WeightedRRFStrategy 加权 RRF 策略
// 结合 RRF 和加权融合的优点
type WeightedRRFStrategy struct {
	K       float64   // RRF 常数
	Weights []float32 // 各列表的权重
}

// Name 返回策略名称
func (s *WeightedRRFStrategy) Name() string {
	return "weighted_rrf"
}

// Rerank 执行加权 RRF
func (s *WeightedRRFStrategy) Rerank(ctx context.Context, query string, resultLists [][]*RetrieveResult) ([]*RetrieveResult, error) {
	k := s.K
	if k == 0 {
		k = 60.0
	}

	// 如果没有指定权重，使用默认权重
	weights := s.Weights
	if len(weights) != len(resultLists) {
		weights = make([]float32, len(resultLists))
		for i := range weights {
			// 根据列表数量分配默认权重
			// 两路：[0.6, 0.4]，三路：[0.5, 0.3, 0.2]
			switch len(resultLists) {
			case 2:
				if i == 0 {
					weights[i] = 0.6
				} else {
					weights[i] = 0.4
				}
			case 3:
				if i == 0 {
					weights[i] = 0.5
				} else if i == 1 {
					weights[i] = 0.3
				} else {
					weights[i] = 0.2
				}
			default:
				weights[i] = 1.0 / float32(len(resultLists))
			}
		}
	}

	// 构建分块 ID 到分数的映射
	scoreMap := make(map[string]float32)
	dataMap := make(map[string]*RetrieveResult)

	for listIdx, results := range resultLists {
		weight := weights[listIdx]

		for rank, result := range results {
			if _, exists := dataMap[result.ChunkID]; !exists {
				dataMap[result.ChunkID] = result
			}
			// 加权 RRF 公式: weight / (k + rank)
			scoreMap[result.ChunkID] += weight / float32(k+float64(rank+1))
		}
	}

	// 构建重排结果
	reranked := make([]*RetrieveResult, 0, len(dataMap))
	for chunkID, score := range scoreMap {
		result := dataMap[chunkID]
		newResult := &RetrieveResult{
			ChunkID:       result.ChunkID,
			KnowledgeID:   result.KnowledgeID,
			KBID:          result.KBID,
			Content:       result.Content,
			ChunkIndex:    result.ChunkIndex,
			Score:         score,
			MatchType:     inferMatchType(resultLists, chunkID),
			StartPosition: result.StartPosition,
			EndPosition:   result.EndPosition,
		}
		reranked = append(reranked, newResult)
	}

	return reranked, nil
}

// SetWeights 设置权重
func (s *WeightedRRFStrategy) SetWeights(weights []float32) {
	s.Weights = weights
}

// ========================================
// 模型重排策略
// ========================================

// ModelRerankStrategy 模型重排策略
// 使用专门的重排模型（如 Cohere Rerank, BGE Reranker）
type ModelRerankStrategy struct {
	embedder RerankEmbedder
	topK     int // 重排前保留的候选数量
}

// Name 返回策略名称
func (s *ModelRerankStrategy) Name() string {
	return "model"
}

// SetEmbedder 设置重排模型
func (s *ModelRerankStrategy) SetEmbedder(embedder RerankEmbedder) {
	s.embedder = embedder
}

// Rerank 执行模型重排
func (s *ModelRerankStrategy) Rerank(ctx context.Context, query string, resultLists [][]*RetrieveResult) ([]*RetrieveResult, error) {
	if s.embedder == nil {
		// 如果没有配置重排模型，回退到 RRF
		fallback := &RRFStrategy{}
		return fallback.Rerank(ctx, query, resultLists)
	}

	// 合并所有结果，去重
	candidates := mergeAndDeduplicate(resultLists)

	// 限制候选数量
	topK := s.topK
	if topK <= 0 {
		topK = 20
	}
	if len(candidates) > topK {
		candidates = candidates[:topK]
	}

	// 准备文档内容
	documents := make([]string, len(candidates))
	for i, result := range candidates {
		documents[i] = result.Content
	}

	// 调用重排模型
	scores, err := s.embedder.Rerank(ctx, query, documents)
	if err != nil {
		// 重排失败，回退到 RRF
		fallback := &RRFStrategy{}
		return fallback.Rerank(ctx, query, resultLists)
	}

	// 使用模型分数重新排序
	for i, result := range candidates {
		if i < len(scores) {
			result.Score = scores[i]
		}
	}

	// 按新的分数排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	// 更新 MatchType
	for _, result := range candidates {
		result.MatchType = "model_rerank"
	}

	return candidates, nil
}

// SetTopK 设置重排前保留的候选数量
func (s *ModelRerankStrategy) SetTopK(topK int) {
	s.topK = topK
}

// ========================================
// 混合重排策略（模型 + RRF）
// ========================================

// HybridRerankStrategy 混合重排策略
// 先用 RRF 融合，再用模型重排 TopK
type HybridRerankStrategy struct {
	embedder  RerankEmbedder
	rrfTopK   int // RRF 后保留的数量
	modelTopK int // 模型重排前保留的候选数量
}

// Name 返回策略名称
func (s *HybridRerankStrategy) Name() string {
	return "hybrid"
}

// SetEmbedder 设置重排模型
func (s *HybridRerankStrategy) SetEmbedder(embedder RerankEmbedder) {
	s.embedder = embedder
}

// Rerank 执行混合重排
func (s *HybridRerankStrategy) Rerank(ctx context.Context, query string, resultLists [][]*RetrieveResult) ([]*RetrieveResult, error) {
	// 第一步：使用 RRF 融合
	rrf := &RRFStrategy{K: 60}
	rrfResults, err := rrf.Rerank(ctx, query, resultLists)
	if err != nil {
		return nil, err
	}

	// 保留 RRF TopK
	rrfTopK := s.rrfTopK
	if rrfTopK <= 0 {
		rrfTopK = 10
	}
	if len(rrfResults) > rrfTopK {
		rrfResults = rrfResults[:rrfTopK]
	}

	// 如果没有配置模型，直接返回 RRF 结果
	if s.embedder == nil {
		return rrfResults, nil
	}

	// 第二步：使用模型重排
	modelStrategy := &ModelRerankStrategy{
		embedder: s.embedder,
		topK:     s.modelTopK,
	}

	// 将 RRF 结果包装成列表形式
	modelLists := [][]*RetrieveResult{rrfResults}
	return modelStrategy.Rerank(ctx, query, modelLists)
}

// SetRRFTopK 设置 RRF 后保留的数量
func (s *HybridRerankStrategy) SetRRFTopK(topK int) {
	s.rrfTopK = topK
}

// SetModelTopK 设置模型重排前保留的候选数量
func (s *HybridRerankStrategy) SetModelTopK(topK int) {
	s.modelTopK = topK
}

// ========================================
// 辅助函数
// ========================================

// inferMatchType 推断匹配类型
func inferMatchType(resultLists [][]*RetrieveResult, chunkID string) string {
	hasVector := false
	hasKeyword := false
	hasGraph := false

	for _, results := range resultLists {
		for _, r := range results {
			if r.ChunkID == chunkID {
				switch r.MatchType {
				case "vector":
					hasVector = true
				case "keyword", "bm25":
					hasKeyword = true
				case "graph":
					hasGraph = true
				case "hybrid":
					hasVector = true
					hasKeyword = true
				}
				break
			}
		}
	}

	// 根据组合推断类型
	count := 0
	if hasVector {
		count++
	}
	if hasKeyword {
		count++
	}
	if hasGraph {
		count++
	}

	if count > 1 {
		return "hybrid"
	}

	if hasGraph {
		return "graph"
	}
	if hasKeyword {
		return "keyword"
	}
	return "vector"
}

// normalizeScores 归一化分数到 [0, 1]
func normalizeScores(results []*RetrieveResult) []*RetrieveResult {
	if len(results) == 0 {
		return results
	}

	// 找到最大和最小分数
	maxScore := results[0].Score
	minScore := results[0].Score

	for _, r := range results {
		if r.Score > maxScore {
			maxScore = r.Score
		}
		if r.Score < minScore {
			minScore = r.Score
		}
	}

	// 如果所有分数相同，直接返回
	if maxScore == minScore {
		return results
	}

	// 归一化
	normalized := make([]*RetrieveResult, len(results))
	for i, r := range results {
		normalized[i] = &RetrieveResult{
			ChunkID:       r.ChunkID,
			KnowledgeID:   r.KnowledgeID,
			KBID:          r.KBID,
			Content:       r.Content,
			ChunkIndex:    r.ChunkIndex,
			Score:         (r.Score - minScore) / (maxScore - minScore),
			MatchType:     r.MatchType,
			StartPosition: r.StartPosition,
			EndPosition:   r.EndPosition,
		}
	}

	return normalized
}

// mergeAndDeduplicate 合并多个结果列表并去重
// 保留每个 chunkID 的最高分数版本
func mergeAndDeduplicate(resultLists [][]*RetrieveResult) []*RetrieveResult {
	seen := make(map[string]*RetrieveResult)

	for _, results := range resultLists {
		for _, result := range results {
			if existing, exists := seen[result.ChunkID]; exists {
				// 保留分数更高的
				if result.Score > existing.Score {
					seen[result.ChunkID] = result
				}
			} else {
				seen[result.ChunkID] = result
			}
		}
	}

	// 转换为列表
	merged := make([]*RetrieveResult, 0, len(seen))
	for _, result := range seen {
		merged = append(merged, result)
	}

	// 按分数排序
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].Score > merged[j].Score
	})

	return merged
}

// sortAndTrimResults 按分数排序并返回 TopK
func sortAndTrimResults(results []*RetrieveResult, topK int) []*RetrieveResult {
	if len(results) == 0 {
		return results
	}

	// 按分数降序排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// 返回 TopK
	if topK > 0 && len(results) > topK {
		return results[:topK]
	}
	return results
}

// ========================================
// 工厂函数
// ========================================

// NewRerankerForStrategy 根据策略名创建重排器
func NewRerankerForStrategy(strategy string) (*Reranker, error) {
	reranker := NewReranker()

	// 根据策略配置
	switch strategy {
	case "rrf", "weighted", "weighted_rrf":
		// 使用默认策略即可
	case "model", "hybrid":
		// 需要外部设置 embedder
	default:
		return nil, fmt.Errorf("未知的重排策略: %s", strategy)
	}

	return reranker, nil
}

// ========================================
// 重排结果包装
// ========================================

// RerankResult 重排结果
type RerankResult struct {
	Results      []*RetrieveResult
	Query        string
	Strategy     string
	SourceTypes  []string // 来源类型列表，如 ["vector", "bm25", "graph"]
	InputCount   int      // 输入结果总数
	OutputCount  int      // 输出结果数
	StrategyUsed string   // 实际使用的策略
}

// ToRetrieveResponse 转换为 RetrieveResponse
func (rr *RerankResult) ToRetrieveResponse() *RetrieveResponse {
	return &RetrieveResponse{
		Results: rr.Results,
		Query:   rr.Query,
		Mode:    rr.Strategy,
	}
}
