package handler

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"golang.org/x/sync/errgroup"

	"link/internal/application/chunker"
	"link/internal/application/service"
	"link/internal/middleware"
	"link/internal/types"
	"link/internal/types/interfaces"
)

// KnowledgeHandlerFull 完整知识库处理器
// 实现完整的文件处理流程：上传 -> 分片 -> 并行构建(稠密向量/稀疏向量/图谱)
type KnowledgeHandlerFull struct {
	kbService     *service.KnowledgeService
	graphService  *service.GraphService
	embedder      embedding.Embedder
	milvusClient  client.Client
	uploadDir     string
	chunker       *chunker.SimpleChunker
	taskProcessor *TaskProcessor
	kbSettingRepo interfaces.KBSettingRepository // 用于读取知识库配置
}

// NewKnowledgeHandlerFull 创建完整知识库处理器
func NewKnowledgeHandlerFull(
	kbService *service.KnowledgeService,
	graphService *service.GraphService,
	embedder embedding.Embedder,
	milvusClient client.Client,
	chunkConfig *chunker.SimpleConfig,
	kbSettingRepo interfaces.KBSettingRepository,
) *KnowledgeHandlerFull {
	uploadDir := "D:\\link\\uploads"
	os.MkdirAll(uploadDir, os.ModePerm)

	handler := &KnowledgeHandlerFull{
		kbService:     kbService,
		graphService:  graphService,
		embedder:      embedder,
		milvusClient:  milvusClient,
		uploadDir:     uploadDir,
		chunker:       chunker.NewSimpleChunker(chunkConfig),
		kbSettingRepo: kbSettingRepo,
	}

	// 创建并启动任务处理器
	handler.taskProcessor = NewTaskProcessor(handler, 2) // 2个worker并发处理
	handler.taskProcessor.Start()

	log.Println("[KnowledgeHandler] Knowledge handler initialized with task processor")
	return handler
}

// UploadKnowledgeFile 上传知识库文件
// POST /api/v1/knowledge-bases/{id}/knowledge/file
func (h *KnowledgeHandlerFull) UploadKnowledgeFile(c *gin.Context) {
	kbID := c.Param("id")

	// 使用 middleware 辅助函数获取租户ID和用户ID
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

	// 解析表单
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file upload failed"})
		return
	}

	// 获取参数
	title := c.PostForm("title")
	fileType := c.PostForm("file_type")
	if fileType == "" {
		fileType = "document"
	}

	chunkSize := h.postFormIntDefault(c, "chunk_size", 512)
	chunkOverlap := h.postFormIntDefault(c, "chunk_overlap", 100)

	// 保存文件
	ext := filepath.Ext(fileHeader.Filename)
	filename := uuid.New().String() + ext
	filePath := filepath.Join(h.uploadDir, filename)

	if err := c.SaveUploadedFile(fileHeader, filePath); err != nil {
		c.JSON(500, gin.H{"error": "failed to save file"})
		return
	}

	// 获取文件信息
	storageSize := fileHeader.Size

	// 创建 Knowledge 记录
	knowledge := &types.Knowledge{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		UserID:       userID,
		KBID:         kbID,
		Type:         fileType,
		Title:        title,
		Source:       "upload",
		FilePath:     filePath,
		StorageSize:  storageSize,
		ParseStatus:  "pending",
		EnableStatus: "enabled",
	}

	if err := h.kbService.Create(c.Request.Context(), knowledge); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Knowledge] File uploaded: %s (%s)", filename, formatFileSize(storageSize))

	// 提交异步处理任务
	task := &Task{
		ID:   uuid.New().String(),
		Type: TaskTypeKnowledgeProcess,
		Data: &KnowledgeTaskData{
			KnowledgeID:  knowledge.ID,
			TenantID:     knowledge.TenantID,
			UserID:       knowledge.UserID,
			KBID:         knowledge.KBID,
			ChunkSize:    chunkSize,
			ChunkOverlap: chunkOverlap,
		},
		CreatedAt: time.Now(),
	}

	if err := h.taskProcessor.Submit(task); err != nil {
		log.Printf("[Knowledge] Failed to submit task: %v", err)
		c.JSON(500, gin.H{"error": "failed to submit processing task"})
		return
	}

	c.JSON(200, gin.H{
		"message":      "file uploaded successfully, processing started",
		"knowledge_id": knowledge.ID,
		"status":       "pending",
		"storage_size": storageSize,
	})
}

// ProcessKnowledgeTask 处理知识库任务（完整流程）
// 实现：分片 -> 并行构建(稠密向量/稀疏向量/图谱数据)
func (h *KnowledgeHandlerFull) ProcessKnowledgeTask(ctx context.Context, knowledgeID string, chunkSize, chunkOverlap int) error {
	log.Printf("[Task] Processing knowledge: %s", knowledgeID)

	knowledge, err := h.kbService.FindByID(ctx, knowledgeID)
	if err != nil {
		return fmt.Errorf("knowledge not found: %w", err)
	}

	// 检查状态
	if knowledge.ParseStatus == "processing" || knowledge.ParseStatus == "completed" {
		log.Printf("[Task] Knowledge already processing/completed: %s", knowledgeID)
		return nil
	}

	// 更新状态为 processing
	knowledge.ParseStatus = "processing"
	h.kbService.Update(ctx, knowledge)

	// ========== 步骤1: 文档解析（分片） ==========
	content, err := os.ReadFile(knowledge.FilePath)
	if err != nil {
		h.markFailed(ctx, knowledge, "failed to read file: "+err.Error())
		return fmt.Errorf("failed to read file: %w", err)
	}

	// 使用 chunker 进行分片
	newChunker := chunker.NewSimpleChunker(&chunker.SimpleConfig{
		ChunkSize:     chunkSize,
		Overlap:       chunkOverlap,
		Separators:    []string{"\n\n", "\n", "。", "！", "？", ".", "!"},
		KeepSeparator: true,
	})

	chunks, err := newChunker.Split(ctx, string(content))
	if err != nil {
		h.markFailed(ctx, knowledge, "chunking failed: "+err.Error())
		return fmt.Errorf("chunking failed: %w", err)
	}

	log.Printf("[Task] Chunks created: %d", len(chunks))

	// ========== 步骤2: 写入 chunks 到数据库 ==========
	chunkIDs, err := h.saveChunksToDB(ctx, knowledge, chunks)
	if err != nil {
		h.markFailed(ctx, knowledge, "save chunks failed: "+err.Error())
		return fmt.Errorf("save chunks failed: %w", err)
	}

	log.Printf("[Task] Chunks saved to DB: %d", len(chunkIDs))

	// ========== 步骤3: 获取知识库配置，并行构建数据 ==========
	// 获取 KBSetting 配置
	setting, err := h.kbSettingRepo.FindByKBID(ctx, knowledge.KBID)
	if err != nil {
		log.Printf("[Task] Warning: failed to get kb_setting: %v, using defaults", err)
		setting = &types.KBSetting{
			GraphEnabled: false,
			BM25Enabled:  nil, // 默认不启用 BM25
		}
	}

	// 确定配置标志
	graphEnabled := setting.GraphEnabled
	bm25Enabled := false
	if setting.BM25Enabled != nil {
		bm25Enabled = *setting.BM25Enabled
	}

	log.Printf("[Task] Config: graph_enabled=%v, bm25_enabled=%v", graphEnabled, bm25Enabled)

	// 并行构建：稠密向量(必须) + 稀疏向量(可选) + 图谱(可选)
	if err := h.parallelBuildData(ctx, knowledge, chunkIDs, chunks, graphEnabled, bm25Enabled); err != nil {
		h.markFailed(ctx, knowledge, "data build failed: "+err.Error())
		return fmt.Errorf("data build failed: %w", err)
	}

	log.Printf("[Task] Data build completed: %d chunks", len(chunkIDs))

	// ========== 步骤4: 更新状态为 completed ==========
	knowledge.ParseStatus = "completed"
	knowledge.EnableStatus = "enabled"
	now := time.Now()
	knowledge.ProcessedAt = &now
	h.kbService.Update(ctx, knowledge)

	log.Printf("[Task] Completed processing knowledge %s: %d chunks", knowledgeID, len(chunkIDs))
	return nil
}

// parallelBuildData 并行构建稠密向量、稀疏向量、图谱数据
func (h *KnowledgeHandlerFull) parallelBuildData(
	ctx context.Context,
	knowledge *types.Knowledge,
	chunkIDs []string,
	chunks []string,
	graphEnabled bool,
	bm25Enabled bool,
) error {
	log.Printf("[ParallelBuild] Starting: chunks=%d, graph=%v, bm25=%v", len(chunkIDs), graphEnabled, bm25Enabled)

	// 使用 errgroup 并行执行
	g, ctx := errgroup.WithContext(ctx)
	// 限制并发数为 4
	g.SetLimit(4)

	// 结果通道
	denseVectors := make([][]float32, len(chunks))
	sparseVectors := make([]entity.SparseEmbedding, len(chunks))
	graphData := make([]*types.GraphData, 1)
	vectorMutex := sync.Mutex{}

	// ========== 并行生成稠密向量（必须） ==========
	g.Go(func() error {
		log.Printf("[ParallelBuild] Starting dense vector generation for %d chunks", len(chunks))
		for i, chunkContent := range chunks {
			embeddings, err := h.embedder.EmbedStrings(ctx, []string{chunkContent})
			if err != nil {
				return fmt.Errorf("dense embedding failed for chunk %d: %w", i, err)
			}

			// 转换为 float32
			denseVec := make([]float32, len(embeddings[0]))
			for j, v := range embeddings[0] {
				denseVec[j] = float32(v)
			}

			vectorMutex.Lock()
			denseVectors[i] = denseVec
			vectorMutex.Unlock()

			if (i+1)%10 == 0 {
				log.Printf("[ParallelBuild] Dense vectors: %d/%d", i+1, len(chunks))
			}
		}
		log.Printf("[ParallelBuild] Dense vector generation completed")
		return nil
	})

	// ========== 并行生成稀疏向量（可选） ==========
	if bm25Enabled {
		g.Go(func() error {
			log.Printf("[ParallelBuild] Starting sparse vector generation for %d chunks", len(chunks))
			for i, chunkContent := range chunks {
				sparseVec, err := h.generateSparseVector(chunkContent)
				if err != nil {
					return fmt.Errorf("sparse vector generation failed for chunk %d: %w", i, err)
				}

				vectorMutex.Lock()
				sparseVectors[i] = sparseVec
				vectorMutex.Unlock()

				if (i+1)%10 == 0 {
					log.Printf("[ParallelBuild] Sparse vectors: %d/%d", i+1, len(chunks))
				}
			}
			log.Printf("[ParallelBuild] Sparse vector generation completed")
			return nil
		})
	}

	// ========== 并行构建图谱数据（可选） ==========
	if graphEnabled {
		g.Go(func() error {
			log.Printf("[ParallelBuild] Starting graph build for %d chunks", len(chunks))

			// 准备提取输入
			inputs := make([]*service.ChunkExtractionInput, len(chunks))
			for i, chunk := range chunks {
				inputs[i] = &service.ChunkExtractionInput{
					KBID:     knowledge.KBID,
					ChunkID:  chunkIDs[i],
					Document: chunk,
					Query:    "请提取实体和关系",
				}
			}

			// 提取图谱
			graph, err := h.graphService.ExtractGraphFromChunks(ctx, inputs)
			if err != nil {
				log.Printf("[ParallelBuild] Graph extraction failed: %v", err)
				return fmt.Errorf("graph extraction failed: %w", err)
			}

			vectorMutex.Lock()
			graphData[0] = graph
			vectorMutex.Unlock()

			log.Printf("[ParallelBuild] Graph build completed: %d nodes, %d relations",
				len(graph.Node), len(graph.Relation))
			return nil
		})
	}

	// 等待所有任务完成
	if err := g.Wait(); err != nil {
		return fmt.Errorf("parallel build failed: %w", err)
	}

	// ========== 写入 Milvus ==========
	log.Printf("[ParallelBuild] Writing to Milvus: %d chunks", len(chunkIDs))
	if err := h.insertToMilvus(ctx, knowledge, chunkIDs, chunks, denseVectors, sparseVectors, bm25Enabled); err != nil {
		return fmt.Errorf("failed to insert to Milvus: %w", err)
	}

	// ========== 写入 Neo4j ==========
	if graphEnabled && graphData[0] != nil {
		log.Printf("[ParallelBuild] Writing to Neo4j")
		namespace := types.NameSpace{
			TenantID:  fmt.Sprintf("%d", knowledge.TenantID),
			KBID:      knowledge.KBID,
			Knowledge: knowledge.ID,
			Type:      knowledge.Type,
		}

		// 使用独立的 context 避免 HTTP 请求超时导致写入失败
		neo4jCtx, neo4jCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer neo4jCancel()

		if err := h.graphService.AddGraph(neo4jCtx, namespace, []*types.GraphData{graphData[0]}); err != nil {
			log.Printf("[ParallelBuild] Warning: failed to save graph to Neo4j: %v", err)
			// 图谱写入失败不影响主流程
		} else {
			log.Printf("[ParallelBuild] Graph saved to Neo4j")
		}
	}

	return nil
}

// insertToMilvus 将数据插入 Milvus
func (h *KnowledgeHandlerFull) insertToMilvus(
	ctx context.Context,
	knowledge *types.Knowledge,
	chunkIDs []string,
	chunks []string,
	denseVectors [][]float32,
	sparseVectors []entity.SparseEmbedding,
	bm25Enabled bool,
) error {
	collectionName := "link"

	// 使用独立的 context 避免被外部取消
	milvusCtx, milvusCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer milvusCancel()

	// 构建列
	chunkIDColumn := entity.NewColumnVarChar("chunk_id", chunkIDs)
	knowledgeIDColumn := entity.NewColumnVarChar("knowledge_id", make([]string, len(chunkIDs)))
	kbIDColumn := entity.NewColumnVarChar("kb_id", make([]string, len(chunkIDs)))
	tenantIDColumn := entity.NewColumnInt64("tenant_id", make([]int64, len(chunkIDs)))
	chunkIndexColumn := entity.NewColumnInt64("chunk_index", make([]int64, len(chunkIDs)))
	contentColumn := entity.NewColumnVarChar("content", chunks)
	isEnabledColumn := entity.NewColumnBool("is_enabled", make([]bool, len(chunkIDs)))
	startAtColumn := entity.NewColumnInt64("start_at", make([]int64, len(chunkIDs)))
	endAtColumn := entity.NewColumnInt64("end_at", make([]int64, len(chunkIDs)))
	tokenCountColumn := entity.NewColumnInt64("token_count", make([]int64, len(chunkIDs)))

	// 填充数据
	knowledgeIDs := make([]string, len(chunkIDs))
	kbIDs := make([]string, len(chunkIDs))
	tenantIDs := make([]int64, len(chunkIDs))
	chunkIndexes := make([]int64, len(chunkIDs))
	isEnableds := make([]bool, len(chunkIDs))
	startAts := make([]int64, len(chunkIDs))
	endAts := make([]int64, len(chunkIDs))
	tokenCounts := make([]int64, len(chunkIDs))

	for i := range chunkIDs {
		knowledgeIDs[i] = knowledge.ID
		kbIDs[i] = knowledge.KBID
		tenantIDs[i] = knowledge.TenantID
		chunkIndexes[i] = int64(i)
		isEnableds[i] = true
		startAts[i] = 0
		endAts[i] = int64(len(chunks[i]))
		tokenCounts[i] = int64(len(chunks[i])) / 2
	}

	knowledgeIDColumn = entity.NewColumnVarChar("knowledge_id", knowledgeIDs)
	kbIDColumn = entity.NewColumnVarChar("kb_id", kbIDs)
	tenantIDColumn = entity.NewColumnInt64("tenant_id", tenantIDs)
	chunkIndexColumn = entity.NewColumnInt64("chunk_index", chunkIndexes)
	isEnabledColumn = entity.NewColumnBool("is_enabled", isEnableds)
	startAtColumn = entity.NewColumnInt64("start_at", startAts)
	endAtColumn = entity.NewColumnInt64("end_at", endAts)
	tokenCountColumn = entity.NewColumnInt64("token_count", tokenCounts)

	// 向量列 - 稠密向量必须
	denseVectorColumn := entity.NewColumnFloatVector("dense_vector", len(denseVectors[0]), denseVectors)

	// 准备插入列
	columns := []entity.Column{
		chunkIDColumn,
		knowledgeIDColumn,
		kbIDColumn,
		tenantIDColumn,
		chunkIndexColumn,
		contentColumn,
		isEnabledColumn,
		startAtColumn,
		endAtColumn,
		tokenCountColumn,
		denseVectorColumn,
	}

	// 稀疏向量列（如果启用）
	if bm25Enabled {
		sparseVectorColumn := entity.NewColumnSparseVectors("sparse_vector", sparseVectors)
		columns = append(columns, sparseVectorColumn)
		log.Printf("[Milvus] Inserting with sparse vectors")
	}

	// 插入数据
	_, err := h.milvusClient.Insert(milvusCtx, collectionName, "", columns...)
	if err != nil {
		return fmt.Errorf("failed to insert into Milvus: %w", err)
	}

	// 刷新以确保可搜索
	if err := h.milvusClient.Flush(ctx, collectionName, false); err != nil {
		log.Printf("[Milvus] Warning: flush failed: %v", err)
	}

	log.Printf("[Milvus] All chunks inserted: dense=%d, sparse=%v",
		len(denseVectors), bm25Enabled)
	return nil
}

// saveChunksToDB 保存 chunks 到数据库（使用事务）
func (h *KnowledgeHandlerFull) saveChunksToDB(
	ctx context.Context,
	knowledge *types.Knowledge,
	chunks []string,
) ([]string, error) {
	chunkIDs := make([]string, len(chunks))
	db := h.kbService.GetDB()

	tx := db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	for i, chunkContent := range chunks {
		chunkID := uuid.New().String()
		log.Printf("[SaveChunk] Chunk %d: length=%d, content_preview=%q", i, len(chunkContent), truncateString(chunkContent, 50))

		// 计算前置和后置 chunk ID
		var preChunkID *string
		if i > 0 {
			preID := chunkIDs[i-1]
			preChunkID = &preID
		}

		tokenCount := len(chunkContent) / 2 // 粗略估算

		chunkRecord := &types.Chunk{
			ID:          chunkID,
			TenantID:    knowledge.TenantID,
			TagID:       nil, // 默认为 nil，后续可关联标签
			KBID:        knowledge.KBID,
			KnowledgeID: knowledge.ID,
			ChunkIndex:  i,
			Content:     chunkContent,
			IsEnabled:   true,
			StartAt:     0,
			EndAt:       len(chunkContent),
			PreChunkID:  preChunkID,
			ChunkType:   "text",
			TokenCount:  &tokenCount,
		}

		if err := tx.Create(chunkRecord).Error; err != nil {
			log.Printf("[SaveChunk] Error creating chunk %d: %v", i, err)
			tx.Rollback()
			return nil, fmt.Errorf("failed to create chunk %d: %w", i, err)
		}
		chunkIDs[i] = chunkID
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return chunkIDs, nil
}

// generateSparseVector 生成稀疏向量（基于简单词频的BM25）
func (h *KnowledgeHandlerFull) generateSparseVector(text string) (entity.SparseEmbedding, error) {
	// 简单分词
	words := tokenize(text)

	wordCount := make(map[string]uint32)
	// 统计词频
	position := uint32(0)
	wordPositions := make(map[string]uint32)

	for _, word := range words {
		if _, exists := wordPositions[word]; !exists {
			wordPositions[word] = position
			position++
		}
		wordCount[word]++
	}

	// 构建 sparse embedding
	indices := make([]uint32, 0, len(wordCount))
	values := make([]float32, 0, len(wordCount))

	for word, count := range wordCount {
		// 简单的 TF 权重
		tf := float32(count)
		pos := wordPositions[word]
		indices = append(indices, pos)
		values = append(values, tf)
	}

	return entity.NewSliceSparseEmbedding(indices, values)
}

// markFailed 标记任务失败
func (h *KnowledgeHandlerFull) markFailed(ctx context.Context, knowledge *types.Knowledge, errMsg string) {
	knowledge.ParseStatus = "failed"
	knowledge.EnableStatus = "error"
	now := time.Now()
	knowledge.ProcessedAt = &now
	h.kbService.Update(ctx, knowledge)
}

// GetKnowledgeStatus 获取知识库处理状态
// GET /api/v1/knowledge-bases/{id}/knowledge/{knowledge_id}/status
func (h *KnowledgeHandlerFull) GetKnowledgeStatus(c *gin.Context) {
	knowledgeID := c.Param("knowledge_id")

	knowledge, err := h.kbService.FindByID(c.Request.Context(), knowledgeID)
	if err != nil {
		c.JSON(404, gin.H{"error": "knowledge not found"})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data": gin.H{
			"knowledge_id":  knowledge.ID,
			"parse_status":  knowledge.ParseStatus,
			"enable_status": knowledge.EnableStatus,
			"chunk_count":   knowledge.ChunkCount,
			"created_at":    knowledge.CreatedAt,
			"processed_at":  knowledge.ProcessedAt,
		},
	})
}

// ========== Helper functions ==========

// postFormIntDefault 从表单获取整数值，带默认值
func (h *KnowledgeHandlerFull) postFormIntDefault(c *gin.Context, key string, defaultValue int) int {
	value := c.PostForm(key)
	if value == "" {
		return defaultValue
	}

	intVal, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Failed to parse %s: %v, using default %d", key, err, defaultValue)
		return defaultValue
	}
	return intVal
}

// formatFileSize 格式化文件大小显示
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div := int64(unit)
	exp := 0
	for size > div {
		size /= unit
		exp++
	}

	if exp > 3 {
		return fmt.Sprintf("%.1f MB", float64(size)/float64(unit*unit*unit*unit))
	}

	return fmt.Sprintf("%d.%d MB", size/unit/unit, size%unit)
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}

// getJSONField 从 JSON 字符串中获取字段值
func getJSONField(jsonStr, key, defaultValue string) string {
	// 简化实现：直接返回默认值
	// TODO: 实现 JSON 解析
	return defaultValue
}

// truncateString 截断字符串用于日志显示
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// tokenize 简单分词函数
func tokenize(text string) []string {
	words := make([]string, 0)
	currentWord := ""

	for _, r := range text {
		if r == ' ' || r == '\t' || r == '\n' || r == '，' || r == '。' || r == '、' {
			if len(currentWord) > 0 {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			currentWord += string(r)
		} else {
			if len(currentWord) > 0 {
				words = append(words, currentWord)
				currentWord = ""
			}
			words = append(words, string(r))
		}
	}

	if len(currentWord) > 0 {
		words = append(words, currentWord)
	}

	return words
}
