package handler

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"link/internal/application/service"
	"link/internal/middleware"
	"link/internal/types"
)

// KnowledgeBaseHandler 知识库处理器
type KnowledgeBaseHandler struct {
	kbBaseService *service.KnowledgeBaseService
}

// NewKnowledgeBaseHandler 创建知识库处理器
func NewKnowledgeBaseHandler(kbBaseService *service.KnowledgeBaseService) *KnowledgeBaseHandler {
	return &KnowledgeBaseHandler{
		kbBaseService: kbBaseService,
	}
}

// Create 创建知识库
// POST /api/v1/knowledge-bases
func (h *KnowledgeBaseHandler) Create(c *gin.Context) {
	var req types.CreateKnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 从上下文获取租户ID和用户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(401, gin.H{"error": "unauthorized: missing tenant_id"})
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"error": "unauthorized: missing user_id"})
		return
	}

	// 生成 UUID 和设置基础字段
	kbID := uuid.New().String()

	// 创建 KnowledgeBase 核心实体
	kb := &types.KnowledgeBase{
		ID:          kbID,
		TenantID:    tenantID,
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Avatar:      req.Avatar,
		IsPublic:    req.IsPublic,
		Status:      1,
	}

	// 创建 KBSetting 配置实体
	// 将检索相关配置存储到 settings_json 中
	settingsJSON := make(map[string]interface{})

	// 处理检索模式
	retrievalMode := "vector" // 默认为向量检索
	graphEnabled := false
	bm25Enabled := false
	if len(req.RetrievalModes) > 0 {
		hasBM25 := false
		hasGraph := false
		for _, mode := range req.RetrievalModes {
			switch mode {
			case "vector":
			case "bm25":
				hasBM25 = true
			case "graph":
				hasGraph = true
			}
		}
		if hasBM25 {
			retrievalMode = "hybrid"
			bm25Enabled = true
		}
		graphEnabled = hasGraph
	} else if req.RetrievalMode != nil {
		retrievalMode = *req.RetrievalMode
		graphEnabled = req.GraphEnabled != nil && *req.GraphEnabled
		bm25Enabled = req.BM25Enabled != nil && *req.BM25Enabled
	}

	// 将检索配置存入 settings_json
	if req.SimilarityThreshold != nil {
		settingsJSON["similarity_threshold"] = *req.SimilarityThreshold
	}
	if req.TopK != nil {
		settingsJSON["top_k"] = *req.TopK
	}
	if req.RerankEnabled != nil {
		settingsJSON["rerank_enabled"] = *req.RerankEnabled
	}
	if req.EmbeddingModelID != nil {
		settingsJSON["embedding_model_id"] = *req.EmbeddingModelID
	}
	if req.SummaryModelID != nil {
		settingsJSON["summary_model_id"] = *req.SummaryModelID
	}
	if req.RerankModelID != nil {
		settingsJSON["rerank_model_id"] = *req.RerankModelID
	}
	settingsJSON["retrieval_mode"] = retrievalMode

	// 处理分块配置
	chunkingConfig := make(map[string]interface{})
	if req.ChunkSize != nil {
		chunkingConfig["chunk_size"] = *req.ChunkSize
	}
	if req.ChunkOverlap != nil {
		chunkingConfig["chunk_overlap"] = *req.ChunkOverlap
	}
	if req.ChunkingConfig != "" {
		var customConfig map[string]interface{}
		if err := json.Unmarshal([]byte(req.ChunkingConfig), &customConfig); err == nil {
			for k, v := range customConfig {
				chunkingConfig[k] = v
			}
		}
	}

	var chunkingConfigJSON *string
	if len(chunkingConfig) > 0 {
		if configJSON, err := json.Marshal(chunkingConfig); err == nil {
			str := string(configJSON)
			chunkingConfigJSON = &str
		}
	}

	setting := &types.KBSetting{
		KBID:           kbID,
		GraphEnabled:   graphEnabled,
		BM25Enabled:    &bm25Enabled,
		ChunkingConfig: chunkingConfigJSON,
	}

	// 将其他配置也存入 settings_json
	if req.CosConfig != "" && req.CosConfig != "null" {
		settingsJSON["cos_config"] = req.CosConfig
	}
	if req.VLMConfig != "" && req.VLMConfig != "null" {
		settingsJSON["vlm_config"] = req.VLMConfig
	}

	// 序列化 settings_json
	if len(settingsJSON) > 0 {
		if settingsJSONBytes, err := json.Marshal(settingsJSON); err == nil {
			str := string(settingsJSONBytes)
			setting.SettingsJSON = &str
		}
	}

	log.Printf("Creating knowledge base: ID=%s, TenantID=%d, UserID=%d, Name=%s", kbID, tenantID, userID, req.Name)

	if err := h.kbBaseService.Create(c.Request.Context(), kb, setting); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// 将设置附加到响应
	kb.Setting = setting

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    kb,
	})
}

// GetList 获取知识库列表
// GET /api/v1/knowledge-bases?page=1&page_size=10
func (h *KnowledgeBaseHandler) GetList(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	kbs, total, err := h.kbBaseService.FindByTenantID(c.Request.Context(), tenantID, page, pageSize)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":     kbs,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// GetDetail 获取知识库详情
// GET /api/v1/knowledge-bases/:id
func (h *KnowledgeBaseHandler) GetDetail(c *gin.Context) {
	id := c.Param("id")

	kb, err := h.kbBaseService.FindByIDWithSettings(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "knowledge base not found"})
		return
	}

	// 构造响应数据，解析配置供前端使用
	retrievalModes := []string{"vector"}

	// 构建展开后的 setting 对象供前端使用
	settingData := gin.H{}

	// 基础布尔字段
	if kb.Setting != nil {
		settingData["graph_enabled"] = kb.Setting.GraphEnabled
		if kb.Setting.BM25Enabled != nil {
			settingData["bm25_enabled"] = *kb.Setting.BM25Enabled
		}
		if kb.Setting.BM25Enabled != nil && *kb.Setting.BM25Enabled {
			if !contains(retrievalModes, "bm25") {
				retrievalModes = append(retrievalModes, "bm25")
			}
		}
		if kb.Setting.GraphEnabled {
			retrievalModes = append(retrievalModes, "graph")
		}

		// 解析 chunking_config JSON
		if kb.Setting.ChunkingConfig != nil {
			var chunkingConfig map[string]interface{}
			if err := json.Unmarshal([]byte(*kb.Setting.ChunkingConfig), &chunkingConfig); err == nil {
				if chunkSize, ok := chunkingConfig["chunk_size"].(float64); ok {
					settingData["chunk_size"] = int(chunkSize)
				}
				if chunkOverlap, ok := chunkingConfig["chunk_overlap"].(float64); ok {
					settingData["chunk_overlap"] = int(chunkOverlap)
				}
			}
		}

		// 解析 settings_json 获取其他配置
		if kb.Setting.SettingsJSON != nil {
			var settingsMap map[string]interface{}
			if err := json.Unmarshal([]byte(*kb.Setting.SettingsJSON), &settingsMap); err == nil {
				// 检索模式
				if mode, ok := settingsMap["retrieval_mode"].(string); ok {
					if mode == "hybrid" && !contains(retrievalModes, "bm25") {
						retrievalModes = append(retrievalModes, "bm25")
					}
				}
				// 其他配置直接复制
				for key, val := range settingsMap {
					switch key {
					case "image_processing_mode":
						if str, ok := val.(string); ok {
							settingData["image_processing_mode"] = str
						}
					case "extract_mode":
						if str, ok := val.(string); ok {
							settingData["extract_mode"] = str
						}
					case "cos_config":
						if str, ok := val.(string); ok {
							settingData["cos_config"] = str
						}
					case "vlm_config":
						if str, ok := val.(string); ok {
							settingData["vlm_config"] = str
						}
					}
				}
			}
		}
	}

	// 设置默认值
	if _, exists := settingData["chunk_size"]; !exists {
		settingData["chunk_size"] = 512
	}
	if _, exists := settingData["chunk_overlap"]; !exists {
		settingData["chunk_overlap"] = 100
	}
	if _, exists := settingData["graph_enabled"]; !exists {
		settingData["graph_enabled"] = false
	}
	if _, exists := settingData["bm25_enabled"]; !exists {
		settingData["bm25_enabled"] = false
	}
	if _, exists := settingData["image_processing_mode"]; !exists {
		settingData["image_processing_mode"] = "none"
	}
	if _, exists := settingData["extract_mode"]; !exists {
		settingData["extract_mode"] = "none"
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":              kb.ID,
			"tenant_id":       kb.TenantID,
			"user_id":         kb.UserID,
			"name":            kb.Name,
			"description":     kb.Description,
			"avatar":          kb.Avatar,
			"status":          kb.Status,
			"is_public":       kb.IsPublic,
			"document_count":  kb.DocumentCount,
			"chunk_count":     kb.ChunkCount,
			"storage_size":    kb.StorageSize,
			"created_at":      kb.CreatedAt,
			"updated_at":      kb.UpdatedAt,
			"setting":         settingData,
			"retrieval_modes": retrievalModes,
		},
	})
}

// Update 更新知识库
// PUT /api/v1/knowledge-bases/:id
func (h *KnowledgeBaseHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req types.UpdateKnowledgeBaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 获取现有知识库及设置
	existingKb, err := h.kbBaseService.FindByIDWithSettings(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "knowledge base not found"})
		return
	}

	// 从上下文获取租户ID
	tenantID := middleware.GetTenantID(c)
	existingKb.TenantID = tenantID

	// 更新核心字段
	if req.Name != nil {
		existingKb.Name = *req.Name
	}
	if req.Description != nil {
		existingKb.Description = *req.Description
	}
	if req.Avatar != nil {
		existingKb.Avatar = *req.Avatar
	}
	if req.IsPublic != nil {
		existingKb.IsPublic = *req.IsPublic
	}
	if req.Status != nil {
		existingKb.Status = *req.Status
	}

	// 获取或创建设置对象
	setting := existingKb.Setting
	if setting == nil {
		setting = &types.KBSetting{
			KBID:         id,
			GraphEnabled: false,
		}
	}

	// 解析现有的 settings_json
	var settingsJSON map[string]interface{}
	if setting.SettingsJSON != nil {
		json.Unmarshal([]byte(*setting.SettingsJSON), &settingsJSON)
	}
	if settingsJSON == nil {
		settingsJSON = make(map[string]interface{})
	}

	// 更新配置字段到 settings_json
	if req.EmbeddingModelID != nil {
		settingsJSON["embedding_model_id"] = *req.EmbeddingModelID
	}
	if req.SummaryModelID != nil {
		settingsJSON["summary_model_id"] = *req.SummaryModelID
	}
	if req.RerankModelID != nil {
		settingsJSON["rerank_model_id"] = *req.RerankModelID
	}
	if req.SimilarityThreshold != nil {
		settingsJSON["similarity_threshold"] = *req.SimilarityThreshold
	}
	if req.TopK != nil {
		settingsJSON["top_k"] = *req.TopK
	}
	if req.RerankEnabled != nil {
		settingsJSON["rerank_enabled"] = *req.RerankEnabled
	}
	if req.CosConfig != nil {
		settingsJSON["cos_config"] = *req.CosConfig
	}
	if req.VLMConfig != nil {
		settingsJSON["vlm_config"] = *req.VLMConfig
	}

	// 处理分块配置
	if req.ChunkSize != nil || req.ChunkOverlap != nil || req.ChunkingConfig != nil {
		chunkingConfig := make(map[string]interface{})
		// 解析现有配置
		if setting.ChunkingConfig != nil {
			json.Unmarshal([]byte(*setting.ChunkingConfig), &chunkingConfig)
		}
		if req.ChunkSize != nil {
			chunkingConfig["chunk_size"] = *req.ChunkSize
		}
		if req.ChunkOverlap != nil {
			chunkingConfig["chunk_overlap"] = *req.ChunkOverlap
		}
		if req.ChunkingConfig != nil {
			if customConfig := parseJSON(*req.ChunkingConfig); customConfig != nil {
				for k, v := range customConfig {
					chunkingConfig[k] = v
				}
			}
		}
		if configJSON, err := json.Marshal(chunkingConfig); err == nil {
			str := string(configJSON)
			setting.ChunkingConfig = &str
		}
	}

	// 处理检索模式
	retrievalMode := "vector"
	if setting.SettingsJSON != nil {
		if existingMode, ok := settingsJSON["retrieval_mode"].(string); ok {
			retrievalMode = existingMode
		}
	}

	graphEnabled := setting.GraphEnabled
	bm25Enabled := false
	if setting.BM25Enabled != nil {
		bm25Enabled = *setting.BM25Enabled
	}

	if len(req.RetrievalModes) > 0 {
		hasBM25 := false
		hasGraph := false
		for _, mode := range req.RetrievalModes {
			switch mode {
			case "vector":
			case "bm25":
				hasBM25 = true
			case "graph":
				hasGraph = true
			}
		}
		if hasBM25 {
			retrievalMode = "hybrid"
			bm25Enabled = true
		} else {
			retrievalMode = "vector"
			bm25Enabled = false
		}
		graphEnabled = hasGraph
	} else {
		if req.RetrievalMode != nil {
			retrievalMode = *req.RetrievalMode
		}
		if req.GraphEnabled != nil {
			graphEnabled = *req.GraphEnabled
		}
		if req.BM25Enabled != nil {
			bm25Enabled = *req.BM25Enabled
		}
	}

	setting.GraphEnabled = graphEnabled
	setting.BM25Enabled = &bm25Enabled
	settingsJSON["retrieval_mode"] = retrievalMode

	// 序列化 settings_json
	if len(settingsJSON) > 0 {
		if settingsJSONBytes, err := json.Marshal(settingsJSON); err == nil {
			str := string(settingsJSONBytes)
			setting.SettingsJSON = &str
		}
	}

	// 使用 UpdateWithSettings 更新
	if err := h.kbBaseService.UpdateWithSettings(c.Request.Context(), existingKb, setting); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    existingKb,
	})
}

// Delete 删除知识库
// DELETE /api/v1/knowledge-bases/:id
func (h *KnowledgeBaseHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.kbBaseService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
	})
}

// GetStats 获取知识库统计信息
// GET /api/v1/knowledge-bases/:id/stats
func (h *KnowledgeBaseHandler) GetStats(c *gin.Context) {
	id := c.Param("id")

	stats, err := h.kbBaseService.GetStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// GetKnowledgeList 获取知识库的文档列表
// GET /api/v1/knowledge-bases/:id/knowledge?page=1&page_size=10
func (h *KnowledgeBaseHandler) GetKnowledgeList(c *gin.Context) {
	kbID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	knowledges, total, err := h.kbBaseService.GetKnowledgeList(c.Request.Context(), kbID, page, pageSize, status)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"kb_id":     kbID,
			"items":     knowledges,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// DeleteKnowledge 删除知识库文档
// DELETE /api/v1/knowledge-bases/:id/knowledge/:knowledge_id
func (h *KnowledgeBaseHandler) DeleteKnowledge(c *gin.Context) {
	kbID := c.Param("id")
	knowledgeID := c.Param("knowledge_id")

	log.Printf("Deleting knowledge: kb_id=%s, knowledge_id=%s", kbID, knowledgeID)

	if err := h.kbBaseService.DeleteKnowledge(c.Request.Context(), kbID, knowledgeID); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
	})
}

// CreateChunk 创建分块（用于测试）
// POST /api/v1/chunks
func (h *KnowledgeBaseHandler) CreateChunk(c *gin.Context) {
	var req types.Chunk
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": -1, "message": err.Error()})
		return
	}

	// 从上下文获取租户ID
	tenantID := middleware.GetTenantID(c)
	if tenantID == 0 {
		c.JSON(401, gin.H{"code": -1, "message": "unauthorized: missing tenant_id"})
		return
	}
	_, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(401, gin.H{"code": -1, "message": "unauthorized: missing user_id"})
		return
	}

	// 生成 UUID 和设置基础字段
	req.ID = uuid.New().String()
	req.TenantID = tenantID

	// 确保 knowledge_id 已设置（测试场景可能不需要）
	if req.KnowledgeID == "" {
		req.KnowledgeID = "test_knowledge"
	}

	log.Printf("Creating chunk: ID=%s, KB_ID=%s, TenantID=%d", req.ID, req.KBID, req.TenantID)

	// 通过 kbBaseService 获取 chunkRepo 并创建
	// 这里需要访问 chunkRepo，暂时通过 service 传递
	if err := h.kbBaseService.CreateChunk(c.Request.Context(), &req); err != nil {
		c.JSON(500, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    req,
	})
}

// GetChunks 获取知识库的分块列表
// GET /api/v1/knowledge-bases/:id/chunks?page=1&page_size=10&knowledge_id=xxx
func (h *KnowledgeBaseHandler) GetChunks(c *gin.Context) {
	kbID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	// 支持 page_size 和 size 两种参数名（前端使用 size）
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", c.DefaultQuery("size", "10")))
	knowledgeID := c.Query("knowledge_id")

	chunks, total, err := h.kbBaseService.GetChunks(c.Request.Context(), kbID, page, pageSize, knowledgeID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"kb_id":     kbID,
			"items":     chunks,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// 辅助函数

// contains 检查字符串切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// parseJSON 解析 JSON 字符串为 map
func parseJSON(jsonStr string) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil
	}
	return result
}
