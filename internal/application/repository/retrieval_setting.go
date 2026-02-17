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
// 检索设置仓储实现
// ========================================

// retrievalSettingRepository 检索设置仓储实现
type retrievalSettingRepository struct {
	base *common_repository.BaseRepository
}

// NewRetrievalSettingRepository 创建检索设置仓储
func NewRetrievalSettingRepository(db *gorm.DB, tenantEnabled bool) interfaces.RetrievalSettingRepository {
	return &retrievalSettingRepository{
		base: common_repository.NewBaseRepository(db, tenantEnabled),
	}
}

// Create 创建检索设置
func (r *retrievalSettingRepository) Create(ctx context.Context, setting *types.RetrievalSetting) error {
	return r.base.Create(ctx, setting)
}

// FindByKBID 根据知识库ID查找检索设置
// 注意：retrieval_settings 表使用 session_id，这里按 tenant_id 查找并返回第一个
func (r *retrievalSettingRepository) FindByKBID(ctx context.Context, kbID string) (*types.RetrievalSetting, error) {
	var setting types.RetrievalSetting
	err := r.base.WithContext(ctx).
		First(&setting).Error

	if err == gorm.ErrRecordNotFound {
		// 返回默认设置
		return r.defaultSetting(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询检索设置失败: %w", err)
	}

	return &setting, nil
}

// FindByID 根据ID查找检索设置
func (r *retrievalSettingRepository) FindByID(ctx context.Context, id int64) (*types.RetrievalSetting, error) {
	var setting types.RetrievalSetting
	err := r.base.WithContext(ctx).
		Where("id = ?", id).
		First(&setting).Error

	if err != nil {
		return nil, fmt.Errorf("查询检索设置失败: %w", err)
	}

	return &setting, nil
}

// Update 更新检索设置
func (r *retrievalSettingRepository) Update(ctx context.Context, setting *types.RetrievalSetting) error {
	return r.base.Update(ctx, setting)
}

// Delete 删除检索设置
func (r *retrievalSettingRepository) Delete(ctx context.Context, kbID string) error {
	db := r.base.WithContext(ctx)
	return db.Delete(&types.RetrievalSetting{}).Error
}

// UpdateVectorConfig 更新向量检索配置
func (r *retrievalSettingRepository) UpdateVectorConfig(ctx context.Context, kbID string, topK int, threshold float64, modelID string) error {
	db := r.base.WithContext(ctx)

	updates := map[string]interface{}{
		"vector_top_k":     topK,
		"vector_threshold": threshold,
	}
	if modelID != "" {
		updates["vector_model_id"] = modelID
	}

	result := db.Model(&types.RetrievalSetting{}).Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("更新向量检索配置失败: %w", result.Error)
	}

	return nil
}

// UpdateBM25Config 更新BM25检索配置
func (r *retrievalSettingRepository) UpdateBM25Config(ctx context.Context, kbID string, topK int) error {
	db := r.base.WithContext(ctx)

	result := db.Model(&types.RetrievalSetting{}).
		Update("bm25_top_k", topK)

	if result.Error != nil {
		return fmt.Errorf("更新BM25检索配置失败: %w", result.Error)
	}

	return nil
}

// UpdateGraphConfig 更新图谱检索配置
func (r *retrievalSettingRepository) UpdateGraphConfig(ctx context.Context, kbID string, enabled bool, topK int, minStrength float64) error {
	db := r.base.WithContext(ctx)

	result := db.Model(&types.RetrievalSetting{}).
		Updates(map[string]interface{}{
			"graph_enabled":      enabled,
			"graph_top_k":        topK,
			"graph_min_strength": minStrength,
		})

	if result.Error != nil {
		return fmt.Errorf("更新图谱检索配置失败: %w", result.Error)
	}

	return nil
}

// UpdateHybridConfig 更新混合检索配置
func (r *retrievalSettingRepository) UpdateHybridConfig(ctx context.Context, kbID string, alpha float64, rerankEnabled bool) error {
	db := r.base.WithContext(ctx)

	result := db.Model(&types.RetrievalSetting{}).
		Updates(map[string]interface{}{
			"hybrid_alpha":          alpha,
			"hybrid_rerank_enabled": rerankEnabled,
		})

	if result.Error != nil {
		return fmt.Errorf("更新混合检索配置失败: %w", result.Error)
	}

	return nil
}

// UpdateWebConfig 更新网络搜索配置
func (r *retrievalSettingRepository) UpdateWebConfig(ctx context.Context, kbID string, enabled bool, topK int, engine, apiKey string, searchDepth int) error {
	db := r.base.WithContext(ctx)

	updates := map[string]interface{}{
		"web_enabled":      enabled,
		"web_top_k":        topK,
		"web_search_depth": searchDepth,
	}
	if engine != "" {
		updates["web_engine"] = engine
	}
	if apiKey != "" {
		updates["web_api_key"] = apiKey
	}

	result := db.Model(&types.RetrievalSetting{}).Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("更新网络搜索配置失败: %w", result.Error)
	}

	return nil
}

// UpdateRerankConfig 更新重排序配置
func (r *retrievalSettingRepository) UpdateRerankConfig(ctx context.Context, kbID string, enabled bool, modelID string) error {
	db := r.base.WithContext(ctx)

	updates := map[string]interface{}{
		"rerank_enabled": enabled,
	}
	if modelID != "" {
		updates["rerank_model_id"] = modelID
	}

	result := db.Model(&types.RetrievalSetting{}).Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("更新重排序配置失败: %w", result.Error)
	}

	return nil
}

// UpdateDefaultMode 更新默认检索模式
func (r *retrievalSettingRepository) UpdateDefaultMode(ctx context.Context, kbID string, mode string, availableModes []string) error {
	// 注意：新的 RetrievalSetting 类型没有 default_mode 和 available_modes 字段
	// 这些配置应该存储在 advanced_config 中
	db := r.base.WithContext(ctx)

	advancedConfig := make(map[string]interface{})
	advancedConfig["default_mode"] = mode
	advancedConfig["available_modes"] = availableModes

	// 将配置序列化为 JSON 字符串
	// 这里简化处理，实际可能需要更复杂的逻辑

	result := db.Model(&types.RetrievalSetting{}).
		Update("advanced_config", advancedConfig)

	if result.Error != nil {
		return fmt.Errorf("更新默认检索模式失败: %w", result.Error)
	}

	return nil
}

// Exists 检查设置是否存在
func (r *retrievalSettingRepository) Exists(ctx context.Context, kbID string) (bool, error) {
	var count int64
	err := r.base.WithContext(ctx).
		Model(&types.RetrievalSetting{}).
		Count(&count).Error

	return count > 0, err
}

// defaultSetting 返回默认检索设置
func (r *retrievalSettingRepository) defaultSetting() *types.RetrievalSetting {
	defaultTopK := 5
	defaultThreshold := 0.7
	bm25TopK := 5
	graphTopK := 5
	graphMinStrength := 1.0
	webSearchDepth := 1

	return &types.RetrievalSetting{
		VectorTopK:          &defaultTopK,
		VectorThreshold:     &defaultThreshold,
		BM25Enable:          nil,
		BM25TopK:            &bm25TopK,
		GraphEnabled:        nil,
		GraphTopK:           &graphTopK,
		GraphMinStrength:    &graphMinStrength,
		HybridAlpha:         nil,
		HybridRerankEnabled: nil,
		WebEnabled:          nil,
		WebSearchDepth:      &webSearchDepth,
		RerankEnabled:       nil,
	}
}

// FindBySessionID 根据会话ID查找检索设置
func (r *retrievalSettingRepository) FindBySessionID(ctx context.Context, sessionID string) (*types.RetrievalSetting, error) {
	var setting types.RetrievalSetting
	err := r.base.WithContext(ctx).
		Where("session_id = ?", sessionID).
		First(&setting).Error

	if err == gorm.ErrRecordNotFound {
		// 返回默认设置
		return r.defaultSetting(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询检索设置失败: %w", err)
	}

	return &setting, nil
}

// UpsertBySessionID 根据会话ID创建或更新检索设置
func (r *retrievalSettingRepository) UpsertBySessionID(ctx context.Context, sessionID string, tenantID int64, ragConfig *types.RAGConfig) error {
	// 先查找是否存在
	var existing types.RetrievalSetting
	err := r.base.WithContext(ctx).
		Where("session_id = ? AND tenant_id = ?", sessionID, tenantID).
		First(&existing).Error

	// 从 retrieval_modes 判断是否启用图谱
	graphEnabled := false
	for _, mode := range ragConfig.RetrievalModes {
		if mode == "graph" {
			graphEnabled = true
			break
		}
	}

	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新记录
		vectorTopK := ragConfig.VectorTopK
		vectorThreshold := ragConfig.SimilarityThreshold
		bm25TopK := ragConfig.KeywordTopK
		graphTopK := ragConfig.GraphTopK
		hybridAlpha := types.Number(float64(ragConfig.Alpha))

		setting := &types.RetrievalSetting{
			TenantID:        tenantID,
			SessionID:       &sessionID,
			VectorTopK:      &vectorTopK,
			VectorThreshold: &vectorThreshold,
			BM25TopK:        &bm25TopK,
			GraphEnabled:    &graphEnabled,
			GraphTopK:       &graphTopK,
			HybridAlpha:     &hybridAlpha,
		}
		return r.base.Create(ctx, setting)
	}

	if err != nil {
		return fmt.Errorf("查询检索设置失败: %w", err)
	}

	// 存在，更新记录
	hybridAlpha := types.Number(float64(ragConfig.Alpha))
	updates := map[string]interface{}{
		"vector_top_k":     ragConfig.VectorTopK,
		"vector_threshold": ragConfig.SimilarityThreshold,
		"bm25_top_k":       ragConfig.KeywordTopK,
		"graph_enabled":    graphEnabled,
		"graph_top_k":      ragConfig.GraphTopK,
		"hybrid_alpha":     hybridAlpha,
	}

	return r.base.WithContext(ctx).Model(&existing).Updates(updates).Error
}
