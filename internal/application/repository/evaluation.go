package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"link/internal/types"
)

// EvaluationRepository 测评任务仓储接口
type EvaluationRepository interface {
	// Create 创建测评任务
	Create(ctx context.Context, task *types.EvaluationTask) error

	// FindByID 根据ID查找测评任务
	FindByID(ctx context.Context, id string) (*types.EvaluationTask, error)

	// FindByTenantID 根据租户ID查找测评任务列表
	FindByTenantID(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.EvaluationTask, int64, error)

	// FindByStatus 根据状态查找测评任务
	FindByStatus(ctx context.Context, status int) ([]*types.EvaluationTask, error)

	// Update 更新测评任务
	Update(ctx context.Context, task *types.EvaluationTask) error

	// UpdateStatus 更新任务状态
	UpdateStatus(ctx context.Context, id string, status int, errMsg string) error

	// UpdateProgress 更新任务进度
	UpdateProgress(ctx context.Context, id string, finished int) error

	// Delete 删除测评任务（软删除）
	Delete(ctx context.Context, id string) error
}

// DatasetRepository 数据集仓储接口
type DatasetRepository interface {
	// FindByDatasetID 根据数据集ID获取QA对
	FindByDatasetID(ctx context.Context, tenantID int64, datasetID string) ([]*types.QAPair, error)

	// Create 创建数据集记录
	Create(ctx context.Context, record *types.DatasetRecord) error

	// CreateBatch 批量创建数据集记录
	CreateBatch(ctx context.Context, records []*types.DatasetRecord) error

	// FindByTenantID 根据租户ID查找数据集
	FindByTenantID(ctx context.Context, tenantID int64) ([]string, error)
}

// ========================================
// Evaluation Repository 实现
// ========================================

type evaluationRepository struct {
	db *gorm.DB
}

// NewEvaluationRepository 创建测评仓储
func NewEvaluationRepository(db *gorm.DB) EvaluationRepository {
	return &evaluationRepository{db: db}
}

func (r *evaluationRepository) Create(ctx context.Context, task *types.EvaluationTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *evaluationRepository) FindByID(ctx context.Context, id string) (*types.EvaluationTask, error) {
	var task types.EvaluationTask
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *evaluationRepository) FindByTenantID(ctx context.Context, tenantID int64, page, pageSize int) ([]*types.EvaluationTask, int64, error) {
	var tasks []*types.EvaluationTask
	var total int64

	db := r.db.WithContext(ctx)

	// 统计总数
	if err := db.Model(&types.EvaluationTask{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&tasks).Error

	return tasks, total, err
}

func (r *evaluationRepository) FindByStatus(ctx context.Context, status int) ([]*types.EvaluationTask, error) {
	var tasks []*types.EvaluationTask
	err := r.db.WithContext(ctx).
		Where("status = ? AND deleted_at IS NULL", status).
		Find(&tasks).Error
	return tasks, err
}

func (r *evaluationRepository) Update(ctx context.Context, task *types.EvaluationTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *evaluationRepository) UpdateStatus(ctx context.Context, id string, status int, errMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errMsg != "" {
		updates["err_msg"] = errMsg
	}
	if status == types.EvaluationStatueSuccess || status == types.EvaluationStatueFailed {
		updates["end_time"] = gorm.Expr("NOW()")
	}
	return r.db.WithContext(ctx).
		Model(&types.EvaluationTask{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *evaluationRepository) UpdateProgress(ctx context.Context, id string, finished int) error {
	return r.db.WithContext(ctx).
		Model(&types.EvaluationTask{}).
		Where("id = ?", id).
		Update("finished", finished).Error
}

func (r *evaluationRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&types.EvaluationTask{}).Error
}

// ========================================
// Dataset Repository 实现
// ========================================

type datasetRepository struct {
	db *gorm.DB
}

// NewDatasetRepository 创建数据集仓储
func NewDatasetRepository(db *gorm.DB) DatasetRepository {
	return &datasetRepository{db: db}
}

func (r *datasetRepository) FindByDatasetID(ctx context.Context, tenantID int64, datasetID string) ([]*types.QAPair, error) {
	var records []*types.DatasetRecord
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND dataset_id = ?", tenantID, datasetID).
		Find(&records).Error
	if err != nil {
		log.Printf("[Dataset] 查询数据集失败: tenantID=%d, datasetID=%s, error=%v", tenantID, datasetID, err)
		return nil, err
	}

	log.Printf("[Dataset] 查询到 %d 条记录 (tenantID=%d, datasetID=%s)", len(records), tenantID, datasetID)

	// 转换为 QAPair
	qapairs := make([]*types.QAPair, 0, len(records))
	for i, record := range records {
		log.Printf("[Dataset] 记录%d: question=%s, pids原始值=%q", i+1, record.Question, record.PIDs)

		qa := &types.QAPair{
			Question: record.Question,
			Answer:   record.Answer,
			PIDs:     []int{},
			Passages: []string{},
		}

		// 解析 PIDs JSON
		if record.PIDs != "" && record.PIDs != "null" {
			if err := json.Unmarshal([]byte(record.PIDs), &qa.PIDs); err != nil {
				log.Printf("[Dataset] 解析PIDs失败: %s, error: %v", record.PIDs, err)
			} else {
				log.Printf("[Dataset] 解析PIDs成功: %v (长度=%d)", qa.PIDs, len(qa.PIDs))
			}
		}

		// 解析 Passages JSON
		if record.Passages != "" && record.Passages != "null" {
			if err := json.Unmarshal([]byte(record.Passages), &qa.Passages); err != nil {
				log.Printf("[Dataset] 解析Passages失败: %s, error: %v", record.Passages, err)
			}
		}

		qapairs = append(qapairs, qa)
	}

	log.Printf("[Dataset] 加载数据集 %s 完成，共 %d 条记录", datasetID, len(qapairs))
	return qapairs, nil
}

func (r *datasetRepository) Create(ctx context.Context, record *types.DatasetRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *datasetRepository) CreateBatch(ctx context.Context, records []*types.DatasetRecord) error {
	if len(records) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(records, 100).Error
}

func (r *datasetRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]string, error) {
	var datasetIDs []string
	err := r.db.WithContext(ctx).
		Model(&types.DatasetRecord{}).
		Where("tenant_id = ?", tenantID).
		Distinct("dataset_id").
		Pluck("dataset_id", &datasetIDs).Error
	return datasetIDs, err
}

// ========================================
// Metrics Repository
// ========================================

// EvaluationMetricsRepository 测评指标仓储接口
type EvaluationMetricsRepository interface {
	// Save 保存测评指标
	Save(ctx context.Context, taskID string, metrics *types.MetricResult) error

	// FindByTaskID 根据任务ID查找指标
	FindByTaskID(ctx context.Context, taskID string) (*types.MetricResult, error)
}

type evaluationMetricsRepository struct {
	db *gorm.DB
}

// NewEvaluationMetricsRepository 创建测评指标仓储
func NewEvaluationMetricsRepository(db *gorm.DB) EvaluationMetricsRepository {
	return &evaluationMetricsRepository{db: db}
}

func (r *evaluationMetricsRepository) Save(ctx context.Context, taskID string, metrics *types.MetricResult) error {
	// 检查是否已存在
	var count int64
	r.db.WithContext(ctx).
		Model(&EvaluationMetricRecord{}).
		Where("task_id = ?", taskID).
		Count(&count)

	if count > 0 {
		// 更新
		return r.db.WithContext(ctx).
			Model(&EvaluationMetricRecord{}).
			Where("task_id = ?", taskID).
			Updates(map[string]interface{}{
				"retrieval_metrics":  toJSON(metrics.RetrievalMetrics),
				"generation_metrics": toJSON(metrics.GenerationMetrics),
			}).Error
	}

	// 创建
	record := &EvaluationMetricRecord{
		TaskID:            taskID,
		RetrievalMetrics:  toJSON(metrics.RetrievalMetrics),
		GenerationMetrics: toJSON(metrics.GenerationMetrics),
	}
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *evaluationMetricsRepository) FindByTaskID(ctx context.Context, taskID string) (*types.MetricResult, error) {
	var record EvaluationMetricRecord
	err := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		First(&record).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	result := &types.MetricResult{}

	// 解析检索指标
	if record.RetrievalMetrics != "" && record.RetrievalMetrics != "{}" {
		var retrievalMetrics types.RetrievalMetrics
		if err := json.Unmarshal([]byte(record.RetrievalMetrics), &retrievalMetrics); err == nil {
			result.RetrievalMetrics = &retrievalMetrics
		}
	}

	// 解析生成指标
	if record.GenerationMetrics != "" && record.GenerationMetrics != "{}" {
		var generationMetrics types.GenerationMetrics
		if err := json.Unmarshal([]byte(record.GenerationMetrics), &generationMetrics); err == nil {
			result.GenerationMetrics = &generationMetrics
		}
	}

	return result, nil
}

// EvaluationMetricRecord 测评指标记录
type EvaluationMetricRecord struct {
	ID                int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TaskID            string    `json:"task_id" gorm:"size:36;not null;uniqueIndex:idx_task_id"`
	RetrievalMetrics  string    `json:"retrieval_metrics" gorm:"type:json"`
	GenerationMetrics string    `json:"generation_metrics" gorm:"type:json"`
	CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (EvaluationMetricRecord) TableName() string {
	return "evaluation_metrics"
}

// toJSON JSON 转换
func toJSON(v interface{}) string {
	if v == nil {
		return "null"
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

// UpdateStatusAndMetrics 更新状态和指标
func (r *evaluationRepository) UpdateStatusAndMetrics(ctx context.Context, id string, status int, metrics *types.MetricResult) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 更新任务状态
		updates := map[string]interface{}{
			"status": status,
		}
		if status == types.EvaluationStatueSuccess || status == types.EvaluationStatueFailed {
			updates["end_time"] = gorm.Expr("NOW()")
		}

		if err := tx.Model(&types.EvaluationTask{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("更新任务状态失败: %w", err)
		}

		// 保存指标
		metricsRepo := NewEvaluationMetricsRepository(tx)
		if err := metricsRepo.Save(ctx, id, metrics); err != nil {
			return fmt.Errorf("保存指标失败: %w", err)
		}

		return nil
	})
}
