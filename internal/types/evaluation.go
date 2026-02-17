package types

import "time"

// ========================================
// Evaluation Status Constants
// ========================================

const (
	// EvaluationStatuePending 等待开始
	EvaluationStatuePending = 0
	// EvaluationStatueRunning 执行中
	EvaluationStatueRunning = 1
	// EvaluationStatueSuccess 成功完成
	EvaluationStatueSuccess = 2
	// EvaluationStatueFailed 执行失败
	EvaluationStatueFailed = 3
)

// ========================================
// Evaluation Task
// ========================================

// EvaluationTask 测评任务
type EvaluationTask struct {
	ID          string     `json:"id" gorm:"primaryKey;size:36"`
	TenantID    int64      `json:"tenant_id" gorm:"not null;index:idx_tenant_id"`
	DatasetID   string     `json:"dataset_id" gorm:"size:100;index:idx_dataset_id"`
	KBID        string     `json:"kb_id" gorm:"size:36;index:idx_kb_id"`
	ChatModelID string     `json:"chat_model_id" gorm:"size:64;index:idx_chat_model"`
	Status      int        `json:"status" gorm:"default:0;index:idx_status"`
	Total       int        `json:"total" gorm:"default:0"`
	Finished    int        `json:"finished" gorm:"default:0"`
	ErrMsg      string     `json:"err_msg" gorm:"type:text"`
	StartTime   time.Time  `json:"start_time" gorm:"autoCreateTime"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	CreatedAt   time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TableName 指定表名
func (EvaluationTask) TableName() string {
	return "evaluation_tasks"
}

// ========================================
// Evaluation Metrics
// ========================================

// RetrievalMetrics 检索指标
type RetrievalMetrics struct {
	Precision float64 `json:"precision"`
	Recall    float64 `json:"recall"`
	NDCG3     float64 `json:"ndcg3"`
	NDCG10    float64 `json:"ndcg10"`
	MRR       float64 `json:"mrr"`
	MAP       float64 `json:"map"`
}

// GenerationMetrics 生成指标
type GenerationMetrics struct {
	BLEU1  float64 `json:"bleu1"`
	BLEU2  float64 `json:"bleu2"`
	BLEU4  float64 `json:"bleu4"`
	ROUGE1 float64 `json:"rouge1"`
	ROUGE2 float64 `json:"rouge2"`
	ROUGEL float64 `json:"rougel"`
}

// MetricResult 综合指标结果
type MetricResult struct {
	RetrievalMetrics  *RetrievalMetrics  `json:"retrieval_metrics,omitempty"`
	GenerationMetrics *GenerationMetrics `json:"generation_metrics,omitempty"`
}

// ========================================
// Evaluation Detail
// ========================================

// EvaluationDetail 测评详情
type EvaluationDetail struct {
	Task   *EvaluationTask   `json:"task"`
	Metric *MetricResult     `json:"metric,omitempty"`
	Params *EvaluationParams `json:"params,omitempty"`
}

// EvaluationParams 测评参数
type EvaluationParams struct {
	VectorThreshold  float64 `json:"vector_threshold"`
	KeywordThreshold float64 `json:"keyword_threshold"`
	EmbeddingTopK    int     `json:"embedding_top_k"`
	ChatModelID      string  `json:"chat_model_id"`
}

// ========================================
// QAPair 数据集QA对
// ========================================

// QAPair QA对
type QAPair struct {
	Question string   `json:"question"`
	Answer   string   `json:"answer"`
	PIDs     []int    `json:"pids"`
	Passages []string `json:"passages"`
}

// DatasetRecord 数据集记录
type DatasetRecord struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	TenantID  int64     `json:"tenant_id" gorm:"not null;index:idx_tenant_id;column:tenant_id"`
	DatasetID string    `json:"dataset_id" gorm:"size:100;not null;index:idx_dataset;column:dataset_id"`
	Question  string    `json:"question" gorm:"type:text;not null;column:question"`
	Answer    string    `json:"answer" gorm:"type:text;column:answer"`
	PIDs      string    `json:"pids" gorm:"type:text;column:pids"`
	Passages  string    `json:"passages" gorm:"type:text;column:passages"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;column:created_at"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime;column:updated_at"`
}

// TableName 指定表名
func (DatasetRecord) TableName() string {
	return "dataset_records"
}

// ========================================
// Request/Response Types
// ========================================

// CreateEvaluationRequest 创建测评任务请求
type CreateEvaluationRequest struct {
	DatasetID       string `json:"dataset_id" binding:"required"`
	KnowledgeBaseID string `json:"knowledge_base_id"`
	ChatID          string `json:"chat_id"`
}

// GetEvaluationRequest 获取测评结果请求
type GetEvaluationRequest struct {
	TaskID string `form:"task_id" binding:"required"`
}

// EvaluationResultResponse 测评结果响应
type EvaluationResultResponse struct {
	Task   *EvaluationTask   `json:"task"`
	Metric *MetricResult     `json:"metric,omitempty"`
	Params *EvaluationParams `json:"params,omitempty"`
}
