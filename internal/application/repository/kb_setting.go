package repository

import (
	"context"
	"fmt"
	common_repository "link/internal/common"

	"gorm.io/gorm"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// ========================================
// 知识库设置仓储实现
// ========================================

// kbSettingRepository 知识库设置仓储实现
type kbSettingRepository struct {
	base *common_repository.BaseRepository
}

// NewKBSettingRepository 创建知识库设置仓储
func NewKBSettingRepository(db *gorm.DB, tenantEnabled bool) interfaces.KBSettingRepository {
	return &kbSettingRepository{
		base: common_repository.NewBaseRepository(db, tenantEnabled),
	}
}

// Create 创建设置
func (r *kbSettingRepository) Create(ctx context.Context, setting *types.KBSetting) error {
	return r.base.Create(ctx, setting)
}

// FindByKBID 根据知识库ID查找设置
func (r *kbSettingRepository) FindByKBID(ctx context.Context, kbID string) (*types.KBSetting, error) {
	var setting types.KBSetting
	err := r.base.WithContext(ctx).
		Where("kb_id = ?", kbID).
		First(&setting).Error

	if err == gorm.ErrRecordNotFound {
		// 如果没有找到，返回默认设置
		return &types.KBSetting{
			KBID:                kbID,
			RetrievalMode:       "vector",
			SimilarityThreshold: 0.7,
			TopK:                5,
			RerankEnabled:       false,
			GraphEnabled:        false,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询知识库设置失败: %w", err)
	}

	return &setting, nil
}

// Update 更新设置
func (r *kbSettingRepository) Update(ctx context.Context, setting *types.KBSetting) error {
	return r.base.Update(ctx, setting)
}

// UpdateRetrievalConfig 更新检索配置
func (r *kbSettingRepository) UpdateRetrievalConfig(ctx context.Context, kbID string, mode string, threshold float64, topK int) error {
	db := r.base.WithContext(ctx)

	// 先尝试更新
	result := db.Model(&types.KBSetting{}).
		Where("kb_id = ?", kbID).
		Updates(map[string]interface{}{
			"retrieval_mode":       mode,
			"similarity_threshold": threshold,
			"top_k":                topK,
		})

	if result.Error != nil {
		return fmt.Errorf("更新检索配置失败: %w", result.Error)
	}

	// 如果没有更新任何行，创建新记录
	if result.RowsAffected == 0 {
		setting := &types.KBSetting{
			KBID:                kbID,
			RetrievalMode:       mode,
			SimilarityThreshold: threshold,
			TopK:                topK,
			RerankEnabled:       false,
			GraphEnabled:        false,
		}
		return r.Create(ctx, setting)
	}

	return nil
}

// Delete 删除设置
func (r *kbSettingRepository) Delete(ctx context.Context, kbID string) error {
	db := r.base.WithContext(ctx)
	return db.Where("kb_id = ?", kbID).Delete(&types.KBSetting{}).Error
}

// Exists 检查设置是否存在
func (r *kbSettingRepository) Exists(ctx context.Context, kbID string) (bool, error) {
	var count int64
	err := r.base.WithContext(ctx).
		Model(&types.KBSetting{}).
		Where("kb_id = ?", kbID).
		Count(&count).Error

	return count > 0, err
}

// GetOrCreate 获取或创建设置（内部辅助方法）
func (r *kbSettingRepository) GetOrCreate(ctx context.Context, kbID string) (*types.KBSetting, error) {
	setting, err := r.FindByKBID(ctx, kbID)
	if err != nil {
		return nil, err
	}

	// 如果是新创建的（ID为0），保存到数据库
	if setting.ID == 0 {
		if err := r.Create(ctx, setting); err != nil {
			return nil, fmt.Errorf("创建默认设置失败: %w", err)
		}
	}

	return setting, nil
}

// UpdateRerankConfig 更新重排序配置
func (r *kbSettingRepository) UpdateRerankConfig(ctx context.Context, kbID string, enabled bool) error {
	db := r.base.WithContext(ctx)

	result := db.Model(&types.KBSetting{}).
		Where("kb_id = ?", kbID).
		Update("rerank_enabled", enabled)

	if result.Error != nil {
		return fmt.Errorf("更新重排序配置失败: %w", result.Error)
	}

	// 如果没有更新任何行，创建新记录
	if result.RowsAffected == 0 {
		setting, err := r.FindByKBID(ctx, kbID)
		if err != nil {
			return err
		}
		setting.RerankEnabled = enabled
		return r.Update(ctx, setting)
	}

	return nil
}

// UpdateGraphConfig 更新知识图谱配置
func (r *kbSettingRepository) UpdateGraphConfig(ctx context.Context, kbID string, enabled bool) error {
	db := r.base.WithContext(ctx)

	result := db.Model(&types.KBSetting{}).
		Where("kb_id = ?", kbID).
		Update("graph_enabled", enabled)

	if result.Error != nil {
		return fmt.Errorf("更新知识图谱配置失败: %w", result.Error)
	}

	// 如果没有更新任何行，创建新记录
	if result.RowsAffected == 0 {
		setting, err := r.FindByKBID(ctx, kbID)
		if err != nil {
			return err
		}
		setting.GraphEnabled = enabled
		return r.Update(ctx, setting)
	}

	return nil
}

// BatchGetSettings 批量获取多个知识库的设置
func (r *kbSettingRepository) BatchGetSettings(ctx context.Context, kbIDs []string) (map[string]*types.KBSetting, error) {
	var settings []*types.KBSetting

	db := r.base.WithContext(ctx)
	err := db.Where("kb_id IN ?", kbIDs).Find(&settings).Error

	if err != nil {
		return nil, fmt.Errorf("批量查询知识库设置失败: %w", err)
	}

	// 构建 map
	result := make(map[string]*types.KBSetting)
	for _, setting := range settings {
		result[setting.KBID] = setting
	}

	// 为没有设置的知识库添加默认设置
	defaultSetting := &types.KBSetting{
		RetrievalMode:       "vector",
		SimilarityThreshold: 0.7,
		TopK:                5,
		RerankEnabled:       false,
		GraphEnabled:        false,
	}

	for _, kbID := range kbIDs {
		if _, exists := result[kbID]; !exists {
			newSetting := *defaultSetting
			newSetting.KBID = kbID
			result[kbID] = &newSetting
		}
	}

	return result, nil
}
