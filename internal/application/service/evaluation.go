package service

import (
	"context"
	"encoding/json"
	"fmt"
	metric2 "link/internal/metric"
	"log"
	"runtime"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"link/internal/config"
	"link/internal/types"
	"link/internal/types/interfaces"
)

// EvaluationService 测评服务
type EvaluationService struct {
	evalRepo         interfaces.EvaluationRepository
	datasetRepo      interfaces.DatasetRepository
	kbRepo           interfaces.KnowledgeBaseRepository
	chunkRepo        interfaces.ChunkRepository
	chatService      interface{}
	rerankService    interface{}
	embeddingService interface{}
	config           *config.Config
	metricsRepo      interface{} // EvaluationMetricsRepository
}

// NewEvaluationService 创建测评服务
func NewEvaluationService(
	evalRepo interfaces.EvaluationRepository,
	datasetRepo interfaces.DatasetRepository,
	kbRepo interfaces.KnowledgeBaseRepository,
	chunkRepo interfaces.ChunkRepository,
	chatService interface{},
	config *config.Config,
) *EvaluationService {
	return &EvaluationService{
		evalRepo:    evalRepo,
		datasetRepo: datasetRepo,
		kbRepo:      kbRepo,
		chunkRepo:   chunkRepo,
		chatService: chatService,
		config:      config,
	}
}

// SetMetricsRepo 设置指标仓储
func (s *EvaluationService) SetMetricsRepo(metricsRepo interface{}) {
	s.metricsRepo = metricsRepo
}

// Evaluation 执行测评
func (s *EvaluationService) Evaluation(
	ctx context.Context,
	tenantID int64,
	datasetID string,
	knowledgeBaseID string,
	chatModelID string,
) (*types.EvaluationDetail, error) {
	log.Printf("[Evaluation] ========== 开始创建测评任务 ==========")
	log.Printf("[Evaluation] tenantID=%d, datasetID=%s, kbID=%s", tenantID, datasetID, knowledgeBaseID)

	// 1. 创建测评任务
	taskID := uuid.New().String()
	now := time.Now()

	task := &types.EvaluationTask{
		ID:          taskID,
		TenantID:    tenantID,
		DatasetID:   datasetID,
		KBID:        knowledgeBaseID,
		ChatModelID: chatModelID,
		Status:      types.EvaluationStatuePending,
		StartTime:   now,
	}

	if err := s.evalRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("创建测评任务失败: %w", err)
	}
	log.Printf("[Evaluation] 任务已创建: %s", taskID)

	// 2. 获取数据集
	log.Printf("[Evaluation] 正在获取数据集...")
	qapairs, err := s.datasetRepo.FindByDatasetID(ctx, tenantID, datasetID)
	if err != nil {
		log.Printf("[Evaluation] 获取数据集失败: %v", err)
		s.evalRepo.UpdateStatus(ctx, taskID, types.EvaluationStatueFailed, fmt.Sprintf("获取数据集失败: %v", err))
		return &types.EvaluationDetail{Task: task}, nil
	}

	log.Printf("[Evaluation] === 数据集加载完成 ===")
	log.Printf("[Evaluation] QA数量: %d", len(qapairs))
	for i, qa := range qapairs {
		log.Printf("[Evaluation]   QA[%d]: question=%q, pids=%v, len(pids)=%d", i, qa.Question, qa.PIDs, len(qa.PIDs))
	}
	log.Printf("[Evaluation] ===============================")

	if len(qapairs) == 0 {
		log.Printf("[Evaluation] 数据集为空，返回失败")
		s.evalRepo.UpdateStatus(ctx, taskID, types.EvaluationStatueFailed, "数据集为空")
		return &types.EvaluationDetail{Task: task}, nil
	}

	// 3. 更新任务状态为运行中
	task.Status = types.EvaluationStatueRunning
	task.Total = len(qapairs)
	if err := s.evalRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("更新任务状态失败: %w", err)
	}
	log.Printf("[Evaluation] 任务状态已更新为运行中: total=%d", task.Total)

	// 4. 同步执行测评（调试模式）
	log.Printf("[Evaluation] 开始同步执行测评...")
	metricResult := s.executeEvaluationSync(ctx, task, qapairs)

	// 5. 保存指标
	if metricResult != nil {
		if s.metricsRepo != nil {
			if repo, ok := s.metricsRepo.(interface {
				Save(ctx context.Context, taskID string, metrics *types.MetricResult) error
			}); ok {
				if err := repo.Save(ctx, taskID, metricResult); err != nil {
					log.Printf("[Evaluation] 保存指标失败: %v", err)
				} else {
					log.Printf("[Evaluation] 指标已保存: precision=%.4f", metricResult.RetrievalMetrics.Precision)
				}
			}
		}
	}

	// 6. 更新任务状态为成功
	s.evalRepo.UpdateStatus(ctx, taskID, types.EvaluationStatueSuccess, "")
	log.Printf("[Evaluation] ========== 测评任务完成 ==========")

	// 7. 返回任务信息
	detail := &types.EvaluationDetail{
		Task:   task,
		Params: s.buildParams(chatModelID),
		Metric: metricResult,
	}

	return detail, nil
}

// executeEvaluationSync 同步执行测评（用于调试）
func (s *EvaluationService) executeEvaluationSync(ctx context.Context, task *types.EvaluationTask, qapairs []*types.QAPair) *types.MetricResult {
	taskID := task.ID

	log.Printf("[Evaluation] ========== 开始执行测评: %s, QA数量: %d ==========", taskID, len(qapairs))

	// 创建 HookMetric 用于记录指标（直接使用 s.metricsRepo）
	hook := NewHookMetric(s.metricsRepo, taskID)

	// 顺序处理每个 QA 对
	for i, qa := range qapairs {
		log.Printf("[Evaluation] --- 处理 QA %d/%d ---", i+1, len(qapairs))

		// 1. 记录 QA 对
		hook.RecordQAPair(qa)
		log.Printf("[Evaluation] QA: question=%q", qa.Question)
		log.Printf("[Evaluation] QA: pids=%v (长度=%d)", qa.PIDs, len(qa.PIDs))

		// 2. 执行检索（模拟）
		retrievedDocs := s.simulateRetrieval(ctx, task.KBID, qa.Question, qa.PIDs)
		log.Printf("[Evaluation] 检索结果: %v (长度=%d)", retrievedDocs, len(retrievedDocs))
		hook.RecordSearchResult(retrievedDocs)

		// 3. 执行重排（模拟）
		rerankedDocs := s.simulateRerank(ctx, retrievedDocs)
		hook.RecordRerankResult(rerankedDocs)

		// 4. 执行大模型生成（模拟）
		response := s.simulateChat(ctx, task.ChatModelID, qa.Question, rerankedDocs)
		log.Printf("[Evaluation] 生成响应: %q", response)
		log.Printf("[Evaluation] 标准答案: %q", qa.Answer)
		hook.RecordChatResponse(response, qa.Answer)

		// 5. 更新进度
		s.evalRepo.UpdateProgress(ctx, task.ID, i+1)
	}

	// 计算并返回指标
	metricResult := hook.Aggregate()
	log.Printf("[Evaluation] ========== 指标计算完成 ==========")
	log.Printf("[Evaluation] Precision: %.4f", metricResult.RetrievalMetrics.Precision)
	log.Printf("[Evaluation] Recall: %.4f", metricResult.RetrievalMetrics.Recall)
	log.Printf("[Evaluation] NDCG@3: %.4f", metricResult.RetrievalMetrics.NDCG3)
	log.Printf("[Evaluation] NDCG@10: %.4f", metricResult.RetrievalMetrics.NDCG10)
	log.Printf("[Evaluation] MRR: %.4f", metricResult.RetrievalMetrics.MRR)
	log.Printf("[Evaluation] MAP: %.4f", metricResult.RetrievalMetrics.MAP)
	log.Printf("[Evaluation] BLEU-1: %.4f", metricResult.GenerationMetrics.BLEU1)
	log.Printf("[Evaluation] ROUGE-1: %.4f", metricResult.GenerationMetrics.ROUGE1)

	return metricResult
}

// executeEvaluation 执行测评（异步）
func (s *EvaluationService) executeEvaluation(ctx context.Context, task *types.EvaluationTask, qapairs []*types.QAPair) {
	taskID := task.ID

	log.Printf("[Evaluation] 开始执行测评任务: %s, 总数: %d", taskID, len(qapairs))

	// 创建 HookMetric 用于记录指标（传入 metricsRepo 用于保存指标）
	hook := NewHookMetric(s.metricsRepo, taskID)

	// 计算工作数
	maxWorkers := runtime.GOMAXPROCS(0) - 1
	if maxWorkers < 1 {
		maxWorkers = 1
	}

	// 创建 errgroup
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxWorkers)

	// 并行处理每个 QA 对
	for _, qa := range qapairs {
		qa := qa // 避免闭包问题
		g.Go(func() error {
			return s.evaluateQA(gctx, task, qa, hook)
		})
	}

	// 等待所有任务完成
	if err := g.Wait(); err != nil {
		log.Printf("[Evaluation] 测评任务失败: %s, 错误: %v", taskID, err)
		s.evalRepo.UpdateStatus(ctx, taskID, types.EvaluationStatueFailed, err.Error())
		return
	}

	// 计算并保存最终指标
	metricResult := hook.Aggregate()
	if err := hook.SaveMetrics(ctx, metricResult); err != nil {
		log.Printf("[Evaluation] 保存指标失败: %v", err)
		s.evalRepo.UpdateStatus(ctx, taskID, types.EvaluationStatueFailed, err.Error())
		return
	}

	// 更新任务状态为成功
	s.evalRepo.UpdateStatus(ctx, taskID, types.EvaluationStatueSuccess, "")

	log.Printf("[Evaluation] 测评任务完成: %s", taskID)
}

// evaluateQA 评估单个 QA 对
func (s *EvaluationService) evaluateQA(ctx context.Context, task *types.EvaluationTask, qa *types.QAPair, hook *HookMetric) error {
	log.Printf("[Evaluation] 评估QA: question=%s, pids=%v", qa.Question, qa.PIDs)

	// 1. 记录 QA 对
	hook.RecordQAPair(qa)

	// 2. 执行检索（模拟）
	retrievedDocs := s.simulateRetrieval(ctx, task.KBID, qa.Question, qa.PIDs)
	log.Printf("[Evaluation] 检索结果: %v", retrievedDocs)
	hook.RecordSearchResult(retrievedDocs)

	// 3. 执行重排（模拟）- 不再需要 rerankModelID
	rerankedDocs := s.simulateRerank(ctx, retrievedDocs)
	hook.RecordRerankResult(rerankedDocs)

	// 4. 执行大模型生成（模拟）
	response := s.simulateChat(ctx, task.ChatModelID, qa.Question, rerankedDocs)
	log.Printf("[Evaluation] 生成响应: %s, 标准答案: %s", response, qa.Answer)
	hook.RecordChatResponse(response, qa.Answer)

	// 5. 更新进度
	s.evalRepo.UpdateProgress(ctx, task.ID, hook.GetFinishedCount())

	return nil
}

// simulateRetrieval 模拟检索（实际应调用 RAG Pipeline）
func (s *EvaluationService) simulateRetrieval(ctx context.Context, kbID string, query string, relevantPIDs []int) []int {
	// 模拟：返回相关文档和一些不相关文档
	results := make([]int, 0, 10)
	results = append(results, relevantPIDs...)
	// 添加一些随机文档模拟检索结果
	for i := 100; i < 105; i++ {
		results = append(results, i)
	}
	log.Printf("[Evaluation] simulateRetrieval: relevantPIDs=%v, results=%v", relevantPIDs, results)
	return results
}

// simulateRerank 模拟重排
func (s *EvaluationService) simulateRerank(ctx context.Context, docs []int) []int {
	// 简化：直接返回原结果
	return docs
}

// simulateChat 模拟大模型生成
func (s *EvaluationService) simulateChat(ctx context.Context, modelID string, question string, docs []int) string {
	// 模拟：根据问题生成包含关键词的回答（用于测试生成指标）
	// 实际使用时应调用真实的 LLM
	responses := map[string]string{
		"什么是知识图谱？":    "知识图谱是一种用图模型来描述知识和建模世界万物之间关联关系的技术方法，它使用节点和边来表示实体和关系。",
		"什么是RAG？":     "RAG（Retrieval-Augmented Generation）是一种结合检索和生成的AI技术，通过从知识库中检索相关信息来增强大模型的生成能力。",
		"向量检索的原理是什么？": "向量检索是通过计算查询向量和文档向量之间的相似度来找到最相关文档的方法，它使用余弦相似度等度量标准。",
		"什么是Neo4j？":   "Neo4j是一个高性能的图数据库管理系统，它使用节点和关系来存储和查询数据，特别适合处理复杂的关联查询。",
		"重排序的作用是什么？":  "重排序是对初次检索结果进行重新排序以提高检索准确性的重要组件，它使用专门的重排序模型对结果进行重新打分。",
	}

	// 如果问题有预设回答，返回预设回答；否则返回通用回答
	if resp, ok := responses[question]; ok {
		return resp
	}
	return "这是一个基于知识库检索内容生成的回答。"
}

// buildParams 构建测评参数
func (s *EvaluationService) buildParams(chatModelID string) *types.EvaluationParams {
	return &types.EvaluationParams{
		VectorThreshold:  0.5,
		KeywordThreshold: 0.3,
		EmbeddingTopK:    10,
		ChatModelID:      chatModelID,
	}
}

// EvaluationResult 获取测评结果
func (s *EvaluationService) EvaluationResult(ctx context.Context, taskID string) (*types.EvaluationDetail, error) {
	task, err := s.evalRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}

	log.Printf("[Evaluation] 获取结果: taskID=%s, status=%d, metricsRepo=%v", taskID, task.Status, s.metricsRepo != nil)

	detail := &types.EvaluationDetail{
		Task:   task,
		Params: s.buildParams(task.ChatModelID),
	}

	// 如果任务完成，从数据库获取指标
	if task.Status == types.EvaluationStatueSuccess && s.metricsRepo != nil {
		// 使用 metricsRepo 获取指标
		if repo, ok := s.metricsRepo.(interface {
			FindByTaskID(ctx context.Context, taskID string) (*types.MetricResult, error)
		}); ok {
			metric, err := repo.FindByTaskID(ctx, taskID)
			log.Printf("[Evaluation] 从数据库获取指标: metric=%v, err=%v", metric, err)
			if err == nil && metric != nil {
				detail.Metric = metric
			}
		} else {
			log.Printf("[Evaluation] metricsRepo 类型断言失败")
		}
	}

	// 如果没有获取到指标，返回空指标
	if detail.Metric == nil {
		log.Printf("[Evaluation] 未获取到指标，返回空指标")
		detail.Metric = &types.MetricResult{
			RetrievalMetrics:  &types.RetrievalMetrics{},
			GenerationMetrics: &types.GenerationMetrics{},
		}
	}

	return detail, nil
}

// ListEvaluations 列出测评任务
func (s *EvaluationService) ListEvaluations(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.EvaluationTask, int64, error) {
	return s.evalRepo.FindByTenantID(ctx, tenantID, page, pageSize)
}

// GetEvaluation 获取单个测评任务
func (s *EvaluationService) GetEvaluation(ctx context.Context, taskID string) (*types.EvaluationTask, error) {
	return s.evalRepo.FindByID(ctx, taskID)
}

// DeleteEvaluation 删除测评任务
func (s *EvaluationService) DeleteEvaluation(ctx context.Context, taskID string) error {
	return s.evalRepo.Delete(ctx, taskID)
}

// CreateDataset 创建数据集
func (s *EvaluationService) CreateDataset(ctx context.Context, tenantID int64, datasetID string, qapairs []*types.QAPair) error {
	records := make([]*types.DatasetRecord, len(qapairs))
	for i, qa := range qapairs {
		pidsJSON, _ := json.Marshal(qa.PIDs)
		passagesJSON, _ := json.Marshal(qa.Passages)

		records[i] = &types.DatasetRecord{
			TenantID:  tenantID,
			DatasetID: datasetID,
			Question:  qa.Question,
			Answer:    qa.Answer,
			PIDs:      string(pidsJSON),
			Passages:  string(passagesJSON),
		}
	}
	return s.datasetRepo.CreateBatch(ctx, records)
}

// ListDatasets 列出数据集
func (s *EvaluationService) ListDatasets(ctx context.Context, tenantID int64) ([]string, error) {
	return s.datasetRepo.FindByTenantID(ctx, tenantID)
}

// ========================================
// 指标计算辅助函数
// ========================================

// CalculateRetrievalMetrics 计算检索指标
func CalculateRetrievalMetrics(qapairs []*types.QAPair, retrievedDocsList [][]int) *types.RetrievalMetrics {
	if len(qapairs) == 0 {
		log.Printf("[Metrics] CalculateRetrievalMetrics: qapairs is empty")
		return &types.RetrievalMetrics{}
	}

	log.Printf("[Metrics] CalculateRetrievalMetrics: qapairs=%d, retrievedDocsList=%d", len(qapairs), len(retrievedDocsList))

	precisionCalc := metric2.NewPrecisionCalculator()
	recallCalc := metric2.NewRecallCalculator()
	ndcgCalc := metric2.NewNDCGCalculator()
	mrrCalc := metric2.NewMRRCalculator()
	mapCalc := metric2.NewMAPCalculator()

	var sumPrecision, sumRecall, sumNDCG3, sumNDCG10, sumMRR, sumMAP float64
	count := 0

	for i, qa := range qapairs {
		retrieved := retrievedDocsList[i]
		log.Printf("[Metrics] QA[%d]: PIDs=%v, retrieved=%v", i, qa.PIDs, retrieved)
		if len(retrieved) == 0 {
			continue
		}

		// Precision@10
		p := precisionCalc.Calculate(qa.PIDs, retrieved)
		sumPrecision += p

		// Recall
		r := recallCalc.Calculate(qa.PIDs, retrieved)
		sumRecall += r

		// NDCG@3, NDCG@10
		ndcg3 := ndcgCalc.NDCG3(qa.PIDs, retrieved)
		ndcg10 := ndcgCalc.NDCG10(qa.PIDs, retrieved)
		sumNDCG3 += ndcg3
		sumNDCG10 += ndcg10

		// MRR
		mrr := mrrCalc.Calculate(qa.PIDs, retrieved)
		sumMRR += mrr

		// MAP
		ap := mapCalc.Calculate(qa.PIDs, retrieved)
		sumMAP += ap

		count++
		log.Printf("[Metrics] QA[%d] metrics: precision=%.4f, recall=%.4f", i, p, r)
	}

	n := float64(len(qapairs))
	log.Printf("[Metrics] Final: n=%f, sumPrecision=%f, sumRecall=%f", n, sumPrecision, sumRecall)
	return &types.RetrievalMetrics{
		Precision: sumPrecision / n,
		Recall:    sumRecall / n,
		NDCG3:     sumNDCG3 / n,
		NDCG10:    sumNDCG10 / n,
		MRR:       sumMRR / n,
		MAP:       sumMAP / n,
	}
}

// CalculateGenerationMetrics 计算生成指标
func CalculateGenerationMetrics(references, candidates []string) *types.GenerationMetrics {
	if len(references) == 0 || len(references) != len(candidates) {
		return &types.GenerationMetrics{}
	}

	var sumBLEU1, sumBLEU2, sumBLEU4, sumROUGE1, sumROUGE2, sumROUGEL float64

	for i := range references {
		sumBLEU1 += metric2.BLEU1(references[i], candidates[i])
		sumBLEU2 += metric2.BLEU2(references[i], candidates[i])
		sumBLEU4 += metric2.BLEU4(references[i], candidates[i])
		sumROUGE1 += metric2.ROUGE1(references[i], candidates[i])
		sumROUGE2 += metric2.ROUGE2(references[i], candidates[i])
		sumROUGEL += metric2.ROUGEL(references[i], candidates[i])
	}

	n := float64(len(references))
	return &types.GenerationMetrics{
		BLEU1:  sumBLEU1 / n,
		BLEU2:  sumBLEU2 / n,
		BLEU4:  sumBLEU4 / n,
		ROUGE1: sumROUGE1 / n,
		ROUGE2: sumROUGE2 / n,
		ROUGEL: sumROUGEL / n,
	}
}
