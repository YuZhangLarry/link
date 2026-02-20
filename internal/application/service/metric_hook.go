package service

import (
	"context"
	metric2 "link/internal/metric"
	"log"
	"sync"

	"link/internal/types"
)

// HookMetric 测评指标记录与聚合
type HookMetric struct {
	evalRepo interface{} // EvaluationRepository
	taskID   string

	mu sync.RWMutex

	// QA 对记录
	qapairs []*types.QAPair

	// 检索结果记录
	searchResults [][]int

	// 重排结果记录
	rerankResults [][]int

	// 生成响应记录
	responses  []string
	references []string

	// 完成计数
	finished int
}

// NewHookMetric 创建 HookMetric
func NewHookMetric(evalRepo interface{}, taskID string) *HookMetric {
	return &HookMetric{
		evalRepo:      evalRepo,
		taskID:        taskID,
		qapairs:       make([]*types.QAPair, 0),
		searchResults: make([][]int, 0),
		rerankResults: make([][]int, 0),
		responses:     make([]string, 0),
		references:    make([]string, 0),
	}
}

// RecordQAPair 记录 QA 对
func (h *HookMetric) RecordQAPair(qa *types.QAPair) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.qapairs = append(h.qapairs, qa)
}

// RecordSearchResult 记录检索结果
func (h *HookMetric) RecordSearchResult(docs []int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.searchResults = append(h.searchResults, docs)
}

// RecordRerankResult 记录重排结果
func (h *HookMetric) RecordRerankResult(docs []int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rerankResults = append(h.rerankResults, docs)
}

// RecordChatResponse 记录生成响应
func (h *HookMetric) RecordChatResponse(response, reference string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.responses = append(h.responses, response)
	h.references = append(h.references, reference)
	h.finished++
}

// GetFinishedCount 获取完成数量
func (h *HookMetric) GetFinishedCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.finished
}

// Aggregate 聚合所有指标
func (h *HookMetric) Aggregate() *types.MetricResult {
	h.mu.RLock()
	defer h.mu.RUnlock()

	log.Printf("[HookMetric] 聚合指标: qapairs=%d, rerankResults=%d, responses=%d, references=%d",
		len(h.qapairs), len(h.rerankResults), len(h.responses), len(h.references))

	result := &types.MetricResult{}

	// 计算检索指标
	if len(h.qapairs) > 0 && len(h.rerankResults) > 0 {
		result.RetrievalMetrics = h.calculateRetrievalMetrics()
		log.Printf("[HookMetric] 检索指标: precision=%.4f, recall=%.4f",
			result.RetrievalMetrics.Precision, result.RetrievalMetrics.Recall)
	}

	// 计算生成指标
	if len(h.responses) > 0 && len(h.references) > 0 {
		result.GenerationMetrics = h.calculateGenerationMetrics()
		log.Printf("[HookMetric] 生成指标: bleu1=%.4f, rouge1=%.4f",
			result.GenerationMetrics.BLEU1, result.GenerationMetrics.ROUGE1)
	}

	return result
}

// calculateRetrievalMetrics 计算检索指标
func (h *HookMetric) calculateRetrievalMetrics() *types.RetrievalMetrics {
	// 使用重排后的结果计算指标
	return CalculateRetrievalMetrics(h.qapairs, h.rerankResults)
}

// calculateGenerationMetrics 计算生成指标
func (h *HookMetric) calculateGenerationMetrics() *types.GenerationMetrics {
	return CalculateGenerationMetrics(h.references, h.responses)
}

// SaveMetrics 保存指标到数据库
func (h *HookMetric) SaveMetrics(ctx context.Context, result *types.MetricResult) error {
	// 类型断言获取 EvaluationMetricsRepository
	if repo, ok := h.evalRepo.(interface {
		Save(ctx context.Context, taskID string, metrics *types.MetricResult) error
	}); ok {
		return repo.Save(ctx, h.taskID, result)
	}
	return nil
}

// ========================================
// 指标计算辅助函数
// ========================================

// calculateQAPairMetrics 计算单个 QA 对的指标
type QAPairMetrics struct {
	QAPair        *types.QAPair
	RetrievedDocs []int
	Response      string

	// 检索指标
	Precision float64
	Recall    float64
	NDCG3     float64
	NDCG10    float64
	MRR       float64

	// 生成指标
	BLEU1  float64
	BLEU2  float64
	BLEU4  float64
	ROUGE1 float64
	ROUGE2 float64
	ROUGEL float64
}

// CalculateQAPairMetrics 计算单个 QA 对的详细指标
func CalculateQAPairMetrics(qa *types.QAPair, retrievedDocs []int, response string) *QAPairMetrics {
	precisionCalc := metric2.NewPrecisionCalculator()
	recallCalc := metric2.NewRecallCalculator()
	ndcgCalc := metric2.NewNDCGCalculator()
	mrrCalc := metric2.NewMRRCalculator()

	return &QAPairMetrics{
		QAPair:        qa,
		RetrievedDocs: retrievedDocs,
		Response:      response,
		Precision:     precisionCalc.Calculate(qa.PIDs, retrievedDocs),
		Recall:        recallCalc.Calculate(qa.PIDs, retrievedDocs),
		NDCG3:         ndcgCalc.NDCG3(qa.PIDs, retrievedDocs),
		NDCG10:        ndcgCalc.NDCG10(qa.PIDs, retrievedDocs),
		MRR:           mrrCalc.Calculate(qa.PIDs, retrievedDocs),
		BLEU1:         metric2.BLEU1(qa.Answer, response),
		BLEU2:         metric2.BLEU2(qa.Answer, response),
		BLEU4:         metric2.BLEU4(qa.Answer, response),
		ROUGE1:        metric2.ROUGE1(qa.Answer, response),
		ROUGE2:        metric2.ROUGE2(qa.Answer, response),
		ROUGEL:        metric2.ROUGEL(qa.Answer, response),
	}
}

// AggregateMetrics 聚合多个 QA 对的指标
func AggregateMetrics(metricsList []*QAPairMetrics) *types.MetricResult {
	if len(metricsList) == 0 {
		return &types.MetricResult{}
	}

	var sumPrecision, sumRecall, sumNDCG3, sumNDCG10, sumMRR float64
	var sumBLEU1, sumBLEU2, sumBLEU4, sumROUGE1, sumROUGE2, sumROUGEL float64

	for _, m := range metricsList {
		sumPrecision += m.Precision
		sumRecall += m.Recall
		sumNDCG3 += m.NDCG3
		sumNDCG10 += m.NDCG10
		sumMRR += m.MRR
		sumBLEU1 += m.BLEU1
		sumBLEU2 += m.BLEU2
		sumBLEU4 += m.BLEU4
		sumROUGE1 += m.ROUGE1
		sumROUGE2 += m.ROUGE2
		sumROUGEL += m.ROUGEL
	}

	n := float64(len(metricsList))

	return &types.MetricResult{
		RetrievalMetrics: &types.RetrievalMetrics{
			Precision: sumPrecision / n,
			Recall:    sumRecall / n,
			NDCG3:     sumNDCG3 / n,
			NDCG10:    sumNDCG10 / n,
			MRR:       sumMRR / n,
			MAP:       sumNDCG3 / n, // 简化：用 NDCG3 近似
		},
		GenerationMetrics: &types.GenerationMetrics{
			BLEU1:  sumBLEU1 / n,
			BLEU2:  sumBLEU2 / n,
			BLEU4:  sumBLEU4 / n,
			ROUGE1: sumROUGE1 / n,
			ROUGE2: sumROUGE2 / n,
			ROUGEL: sumROUGEL / n,
		},
	}
}
