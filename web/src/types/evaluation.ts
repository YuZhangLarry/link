/**
 * 测评状态
 */
export enum EvaluationStatus {
  Pending = 0,
  Running = 1,
  Success = 2,
  Failed = 3
}

/**
 * 测评状态文本映射
 */
export const EvaluationStatusText: Record<EvaluationStatus, string> = {
  [EvaluationStatus.Pending]: '等待中',
  [EvaluationStatus.Running]: '执行中',
  [EvaluationStatus.Success]: '已完成',
  [EvaluationStatus.Failed]: '失败'
}

/**
 * 测评状态类型映射
 */
export const EvaluationStatusType: Record<EvaluationStatus, 'info' | 'warning' | 'success' | 'danger'> = {
  [EvaluationStatus.Pending]: 'info',
  [EvaluationStatus.Running]: 'warning',
  [EvaluationStatus.Success]: 'success',
  [EvaluationStatus.Failed]: 'danger'
}

/**
 * 检索指标
 */
export interface RetrievalMetrics {
  precision: number    // 精确率
  recall: number       // 召回率
  ndcg3: number        // NDCG@3
  ndcg10: number       // NDCG@10
  mrr: number          // 平均倒数排名
  map: number          // 平均精确率
}

/**
 * 生成指标
 */
export interface GenerationMetrics {
  bleu1: number   // BLEU-1
  bleu2: number   // BLEU-2
  bleu4: number   // BLEU-4
  rouge1: number  // ROUGE-1
  rouge2: number  // ROUGE-2
  rougeL: number  // ROUGE-L
}

/**
 * 综合指标结果
 */
export interface MetricResult {
  retrieval_metrics?: RetrievalMetrics
  generation_metrics?: GenerationMetrics
}

/**
 * 测评参数
 */
export interface EvaluationParams {
  vector_threshold: number
  keyword_threshold: number
  embedding_top_k: number
  rerank_threshold: number
  rerank_top_k: number
  chat_model_id: string
}

/**
 * 测评任务
 */
export interface EvaluationTask {
  id: string
  tenant_id: number
  dataset_id: string
  kb_id: string
  chat_model_id: string
  rerank_model_id: string
  status: EvaluationStatus
  total: number
  finished: number
  err_msg: string
  start_time: string
  end_time?: string
  created_at: string
  updated_at: string
}

/**
 * 测评详情
 */
export interface EvaluationDetail {
  task: EvaluationTask
  metric?: MetricResult
  params?: EvaluationParams
}

/**
 * QA对
 */
export interface QAPair {
  question: string
  answer: string
  pids: number[]
  passages: string[]
}

/**
 * 创建测评请求
 */
export interface CreateEvaluationRequest {
  dataset_id: string
  knowledge_base_id?: string
  chat_id?: string
}

/**
 * 创建数据集请求
 */
export interface CreateDatasetRequest {
  dataset_id: string
  qapairs: QAPair[]
}
