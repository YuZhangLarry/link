package container

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"link/internal/config"
)

var MilvusClient client.Client

// InitMilvus 初始化Milvus连接
func InitMilvus(cfg *config.MilvusConfig) error {
	if cfg.Host == "" || cfg.Token == "" {
		return fmt.Errorf("Milvus配置不完整: host或token为空")
	}

	// 配置 Milvus 客户端
	milvusCfg := client.Config{
		Address:  cfg.Host,
		APIKey:   cfg.Token,
		Username: "",
		Password: "",
	}

	// 创建客户端连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := client.NewClient(ctx, milvusCfg)
	if err != nil {
		return fmt.Errorf("创建Milvus客户端失败: %w", err)
	}

	// 测试连接 - 通过列出collections来验证
	_, err = c.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("Milvus连接测试失败: %w", err)
	}

	MilvusClient = c
	log.Printf("✅ Milvus连接成功: %s\n", cfg.Host)

	return nil
}

// CloseMilvus 关闭Milvus连接
func CloseMilvus() error {
	if MilvusClient != nil {
		return MilvusClient.Close()
	}
	return nil
}

// GetMilvus 获取Milvus客户端
func GetMilvus() client.Client {
	return MilvusClient
}

// InitMilvusSchema 初始化 Milvus Schema（创建 link collection 并插入测试数据）
func InitMilvusSchema(embedder embedding.Embedder) error {
	if MilvusClient == nil {
		return fmt.Errorf("Milvus 客户端未初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collectionName := "link"

	// 1. 检查 collection 是否已存在
	has, err := MilvusClient.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("检查 collection 失败: %w", err)
	}

	if has {
		// 删除已存在的 collection
		log.Printf("⚠️  Collection '%s' 已存在，正在删除...\n", collectionName)
		err = MilvusClient.DropCollection(ctx, collectionName)
		if err != nil {
			return fmt.Errorf("删除已有 collection 失败: %w", err)
		}
		log.Printf("✅ 已删除旧的 collection\n")
	}

	// 2. 创建 Schema
	schema := createLinkSchema()

	// 3. 创建 collection
	err = MilvusClient.CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		return fmt.Errorf("创建 collection 失败: %w", err)
	}
	log.Printf("✅ Collection '%s' 创建成功\n", collectionName)

	// 4. 创建索引
	err = createLinkIndex(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("创建索引失败: %w", err)
	}
	log.Printf("✅ 索引创建成功\n")

	// 5. 加载 collection 到内存
	err = MilvusClient.LoadCollection(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("加载 collection 失败: %w", err)
	}
	log.Printf("✅ Collection 已加载到内存\n")

	// 6. 插入测试数据
	err = insertTestData(ctx, collectionName, embedder)
	if err != nil {
		return fmt.Errorf("插入测试数据失败: %w", err)
	}

	log.Printf("🎉 Milvus Schema 初始化完成\n")
	return nil
}

// createLinkSchema 创建 link collection 的 schema
func createLinkSchema() *entity.Schema {
	// 使用 entity.NewSchema() 创建 schema
	schema := entity.NewSchema().WithName("link").WithDescription("Link 知识库向量索引")

	// 1. 主键字段
	schema = schema.WithField(
		entity.NewField().WithName("id").WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true).WithIsAutoID(true),
	)

	// 2. 稠密向量字段（Dense Vector）- 用于语义检索
	schema = schema.WithField(
		entity.NewField().WithName("dense_vector").WithDataType(entity.FieldTypeFloatVector).WithDim(1024), // DashScope text-embedding-v4 实际维度
	)

	// 3. 稀疏向量字段（Sparse Vector）- 用于关键词检索（BM25）
	schema = schema.WithField(
		entity.NewField().WithName("sparse_vector").WithDataType(entity.FieldTypeSparseVector),
	)

	// 4. 元数据字段
	schema = schema.WithField(
		entity.NewField().WithName("chunk_id").WithDataType(entity.FieldTypeVarChar).WithMaxLength(36),
	)
	schema = schema.WithField(
		entity.NewField().WithName("knowledge_id").WithDataType(entity.FieldTypeVarChar).WithMaxLength(36),
	)
	schema = schema.WithField(
		entity.NewField().WithName("kb_id").WithDataType(entity.FieldTypeVarChar).WithMaxLength(36),
	)
	schema = schema.WithField(
		entity.NewField().WithName("tenant_id").WithDataType(entity.FieldTypeInt64),
	)
	schema = schema.WithField(
		entity.NewField().WithName("chunk_index").WithDataType(entity.FieldTypeInt64),
	)
	schema = schema.WithField(
		entity.NewField().WithName("content").WithDataType(entity.FieldTypeVarChar).WithMaxLength(65535),
	)
	schema = schema.WithField(
		entity.NewField().WithName("is_enabled").WithDataType(entity.FieldTypeBool),
	)
	schema = schema.WithField(
		entity.NewField().WithName("start_at").WithDataType(entity.FieldTypeInt64),
	)
	schema = schema.WithField(
		entity.NewField().WithName("end_at").WithDataType(entity.FieldTypeInt64),
	)
	schema = schema.WithField(
		entity.NewField().WithName("token_count").WithDataType(entity.FieldTypeInt64),
	)

	return schema
}

// createLinkIndex 创建索引
func createLinkIndex(ctx context.Context, collectionName string) error {
	// 1. 创建稠密向量索引（使用 HNSW）
	denseIdx, err := entity.NewIndexHNSW(entity.L2, 16, 256)
	if err != nil {
		return fmt.Errorf("创建稠密向量索引对象失败: %w", err)
	}

	err = MilvusClient.CreateIndex(ctx, collectionName, "dense_vector", denseIdx, false)
	if err != nil {
		return fmt.Errorf("创建 dense_vector 索引失败: %w", err)
	}

	// 2. 创建稀疏向量索引（使用 SPARSE_INVERTED_INDEX）
	sparseIdx, err := entity.NewIndexSparseInverted(entity.IP, 0.2)
	if err != nil {
		return fmt.Errorf("创建稀疏向量索引对象失败: %w", err)
	}

	err = MilvusClient.CreateIndex(ctx, collectionName, "sparse_vector", sparseIdx, false)
	if err != nil {
		return fmt.Errorf("创建 sparse_vector 索引失败: %w", err)
	}

	return nil
}

// insertTestData 插入测试数据
func insertTestData(ctx context.Context, collectionName string, embedder embedding.Embedder) error {
	// 准备测试数据
	testContents := []string{
		"Link 是一个基于 Go 语言开发的智能对话系统，支持多租户架构和知识库管理。",
		"向量检索是通过将文本转换为向量表示，然后在向量空间中查找最相似的文本。",
		"BM25 是一种常用的文本检索算法，通过计算词频和逆文档频率来评估文档的相关性。",
		"Milvus 是一个开源的向量数据库，专门用于存储和检索大规模的向量数据。",
		"知识图谱是一种用图结构表示知识的技术，可以描述实体之间的关系。",
		"混合检索结合了向量检索和关键词检索的优点，能够提高检索的准确率。",
		"租户隔离是指在一个系统中，不同租户的数据相互独立，互不干扰。",
		"Chunk（分块）是将长文档切分成较短的文本片段，便于向量化和检索。",
	}

	// 1. 生成稠密向量
	embeddings, err := embedder.EmbedStrings(ctx, testContents)
	if err != nil {
		return fmt.Errorf("生成测试数据稠密向量失败: %w", err)
	}

	// 转换为 float32
	denseVectors := make([][]float32, len(embeddings))
	for i, emb := range embeddings {
		denseVectors[i] = make([]float32, len(emb))
		for j, v := range emb {
			denseVectors[i][j] = float32(v)
		}
	}

	// 2. 生成稀疏向量（基于词频的简单模拟）
	sparseVectors, err := generateMockSparseVectors(testContents)
	if err != nil {
		return fmt.Errorf("生成稀疏向量失败: %w", err)
	}

	// 构建列数据
	chunkIDs := []string{
		"test-chunk-001",
		"test-chunk-002",
		"test-chunk-003",
		"test-chunk-004",
		"test-chunk-005",
		"test-chunk-006",
		"test-chunk-007",
		"test-chunk-008",
	}
	knowledgeIDs := make([]string, len(testContents))
	kbIDs := make([]string, len(testContents))
	for i := range testContents {
		knowledgeIDs[i] = "test-knowledge-001"
		kbIDs[i] = "test-kb-001"
	}

	tenantIDs := make([]int64, len(testContents))
	chunkIndexes := make([]int64, len(testContents))
	contents := testContents
	isEnableds := make([]bool, len(testContents))
	startAts := make([]int64, len(testContents))
	endAts := make([]int64, len(testContents))
	tokenCounts := make([]int64, len(testContents))

	for i := range testContents {
		tenantIDs[i] = 1
		chunkIndexes[i] = int64(i)
		isEnableds[i] = true
		startAts[i] = 0
		endAts[i] = int64(len(testContents[i]))
		tokenCounts[i] = int64(len(testContents[i])) / 2 // 估算
	}

	// 构建插入列
	dim := len(denseVectors[0])
	columns := []entity.Column{
		entity.NewColumnFloatVector("dense_vector", dim, denseVectors),
		entity.NewColumnSparseVectors("sparse_vector", sparseVectors),
		entity.NewColumnVarChar("chunk_id", chunkIDs),
		entity.NewColumnVarChar("knowledge_id", knowledgeIDs),
		entity.NewColumnVarChar("kb_id", kbIDs),
		entity.NewColumnInt64("tenant_id", tenantIDs),
		entity.NewColumnInt64("chunk_index", chunkIndexes),
		entity.NewColumnVarChar("content", contents),
		entity.NewColumnBool("is_enabled", isEnableds),
		entity.NewColumnInt64("start_at", startAts),
		entity.NewColumnInt64("end_at", endAts),
		entity.NewColumnInt64("token_count", tokenCounts),
	}

	// 插入数据
	_, err = MilvusClient.Insert(ctx, collectionName, "", columns...)
	if err != nil {
		return fmt.Errorf("插入数据失败: %w", err)
	}

	// 刷新数据以确保可搜索
	err = MilvusClient.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("刷新 collection 失败: %w", err)
	}

	log.Printf("✅ 已插入 %d 条测试数据\n", len(testContents))
	return nil
}

// generateMockSparseVectors 生成模拟稀疏向量（基于简单词频统计）
func generateMockSparseVectors(texts []string) ([]entity.SparseEmbedding, error) {
	sparseVectors := make([]entity.SparseEmbedding, len(texts))

	// 构建全局词表和文档频率
	wordDocFreq := make(map[string]int)                 // 词出现在多少文档中
	docWordCounts := make([]map[string]int, len(texts)) // 每个文档的词频
	wordPositions := make(map[string]uint32)            // 词汇位置映射

	// 第一遍扫描：收集所有词汇并分配位置
	position := uint32(0)
	for _, text := range texts {
		words := tokenize(text)
		seenWords := make(map[string]bool)
		for _, word := range words {
			if !seenWords[word] {
				if _, exists := wordPositions[word]; !exists {
					wordPositions[word] = position
					position++
				}
				seenWords[word] = true
			}
		}
	}

	// 第二遍扫描：计算文档频率
	for docIdx, text := range texts {
		words := tokenize(text)
		docWordCounts[docIdx] = make(map[string]int)
		seenWords := make(map[string]bool)

		for _, word := range words {
			docWordCounts[docIdx][word]++
			if !seenWords[word] {
				wordDocFreq[word]++
				seenWords[word] = true
			}
		}
	}

	// 计算 BM25 权重生成稀疏向量
	totalDocs := len(texts)
	avgDocLen := 0.0
	for _, counts := range docWordCounts {
		totalWords := 0
		for _, count := range counts {
			totalWords += count
		}
		avgDocLen += float64(totalWords)
	}
	avgDocLen /= float64(totalDocs)

	k1 := 1.5
	b := 0.75

	for docIdx := range texts {
		wordCounts := docWordCounts[docIdx]

		indices := make([]uint32, 0)
		values := make([]float32, 0)

		// 为文档中的每个唯一词计算 BM25 分数
		for word, count := range wordCounts {
			df := wordDocFreq[word]
			idf := math.Log(float64(totalDocs-df+1)/float64(float64(df)+0.5) + 1.0)

			tf := float64(count)
			numerator := tf * (k1 + 1)
			denominator := tf + k1*(1-b+b*float64(len(wordCounts))/avgDocLen)
			bm25 := idf * (numerator / denominator)

			if bm25 > 0.01 { // 只保留显著权重
				pos := wordPositions[word]
				indices = append(indices, pos)
				values = append(values, float32(bm25))
			}
		}

		sparseEmb, err := entity.NewSliceSparseEmbedding(indices, values)
		if err != nil {
			return nil, fmt.Errorf("生成稀疏向量失败 (文档 %d): %w", docIdx, err)
		}
		sparseVectors[docIdx] = sparseEmb
	}

	return sparseVectors, nil
}

// tokenize 简单分词函数（按字符和常见分隔符切分）
func tokenize(text string) []string {
	// 简单分词：按空格、标点符号分割
	words := make([]string, 0)
	currentWord := ""

	for _, r := range text {
		if r == ' ' || r == '\t' || r == '\n' || r == '，' || r == '。' || r == '、' || r == '（' || r == '）' {
			if len(currentWord) > 0 {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			// 英文字母和数字
			currentWord += string(r)
		} else {
			// 中文字符单独成词
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
