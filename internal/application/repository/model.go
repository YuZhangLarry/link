package repository

import (
	"context"
	"link/internal/models"

	"gorm.io/gorm"
)

// ModelRepository 模型仓储接口
type ModelRepository interface {
	FindByTenantID(ctx context.Context, tenantID int64) ([]*models.Model, error)
	FindByType(ctx context.Context, tenantID int64, modelType string) ([]*models.Model, error)
	FindByID(ctx context.Context, id string) (*models.Model, error)
}

type modelRepository struct {
	db *gorm.DB
}

// NewModelRepository 创建模型仓储
func NewModelRepository(db *gorm.DB) ModelRepository {
	return &modelRepository{db: db}
}

func (r *modelRepository) FindByTenantID(ctx context.Context, tenantID int64) ([]*models.Model, error) {
	var models []*models.Model
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Find(&models).Error
	return models, err
}

func (r *modelRepository) FindByType(ctx context.Context, tenantID int64, modelType string) ([]*models.Model, error) {
	var models []*models.Model
	query := r.db.WithContext(ctx).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	if modelType != "" {
		query = query.Where("type = ?", modelType)
	}

	err := query.Find(&models).Error
	return models, err
}

func (r *modelRepository) FindByID(ctx context.Context, id string) (*models.Model, error) {
	var model models.Model
	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&model).Error
	if err != nil {
		return nil, err
	}
	return &model, nil
}
