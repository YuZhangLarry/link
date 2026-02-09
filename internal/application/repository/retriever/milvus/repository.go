package milvus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"link/internal/container"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

// VectorRetriever 向量检索器
type VectorRetriever struct {
	client   client.Client
	embedder embedding.Embedder
}

// NewVectorRetriever 创建向量检索器
func NewVectorRetriever(embedder embedding.Embedder) (*VectorRetriever, error) {
	if embedder == nil {
		return nil, fmt.Errorf("embedder is required")
	}

	cli := container.GetMilvus()
	if cli == nil {
		return nil, fmt.Errorf("milvus client not initialized")
	}

	return &VectorRetriever{
		client:   cli,
		embedder: embedder,
	}, nil
}

// ========================================
// 知识库 (Collection) 管理
// ========================================

// CreateKnowledgeBaseOptions 创建知识库选项
type CreateKnowledgeBaseOptions struct {
	Dimension     int               // 向量维度
	IndexType     IndexType         // 索引类型
	MetricType    entity.MetricType // 距离度量类型
	AutoID        bool              // 是否自动生成ID
	EnableDynamic bool              // 是否启用动态字段
	Fields        []*entity.Field   // 字段定义
	Description   string            // 描述
}

// IndexType 索引类型
type IndexType string

const (
	IndexTypeFlat      IndexType = "FLAT"
	IndexTypeIvfFlat   IndexType = "IVF_FLAT"
	IndexTypeIvfSq8    IndexType = "IVF_SQ8"
	IndexTypeIvfPq     IndexType = "IVF_PQ"
	IndexTypeHnsw      IndexType = "HNSW"
	IndexTypeDiskAnn   IndexType = "DISKANN"
	IndexTypeAutoIndex IndexType = "AUTOINDEX"
	IndexTypeScalar    IndexType = "SCALAR"
)

// CreateKnowledgeBase 创建知识库 (Collection)
func (r *VectorRetriever) CreateKnowledgeBase(ctx context.Context, kbID int64, opts *CreateKnowledgeBaseOptions) error {
	schema := r.buildSchema(kbID, opts)

	// 创建 collection，使用默认分片数
	err := r.client.CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		return fmt.Errorf("create collection failed: %w", err)
	}

	log.Printf("[Milvus] Collection created: kb_%d", kbID)
	return nil
}

// buildSchema 构建集合 Schema
func (r *VectorRetriever) buildSchema(kbID int64, opts *CreateKnowledgeBaseOptions) *entity.Schema {
	collectionName := r.getCollectionName(kbID)

	// 使用 entity.NewSchema() 创建 schema
	schema := entity.NewSchema().WithName(collectionName).WithDescription(opts.Description)

	// 添加主键字段
	schema = schema.WithField(
		entity.NewField().WithName("id").WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true).WithIsAutoID(opts.AutoID),
	)

	// 添加向量字段
	schema = schema.WithField(
		entity.NewField().WithName("vector").WithDataType(entity.FieldTypeFloatVector).WithDim(int64(opts.Dimension)),
	)

	// 添加文档ID字段
	schema = schema.WithField(
		entity.NewField().WithName("document_id").WithDataType(entity.FieldTypeInt64),
	)

	// 添加分块索引字段
	schema = schema.WithField(
		entity.NewField().WithName("chunk_index").WithDataType(entity.FieldTypeInt64),
	)

	// 添加内容字段
	schema = schema.WithField(
		entity.NewField().WithName("content").WithDataType(entity.FieldTypeVarChar).WithMaxLength(65535),
	)

	// 添加元数据字段
	schema = schema.WithField(
		entity.NewField().WithName("metadata").WithDataType(entity.FieldTypeJSON),
	)

	// 启用动态字段
	if opts.EnableDynamic {
		schema = schema.WithDynamicFieldEnabled(true)
	}

	return schema
}

// DeleteKnowledgeBase 删除知识库 (Collection)
func (r *VectorRetriever) DeleteKnowledgeBase(ctx context.Context, kbID int64) error {
	collectionName := r.getCollectionName(kbID)

	err := r.client.DropCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("drop collection failed: %w", err)
	}

	log.Printf("[Milvus] Collection dropped: kb_%d", kbID)
	return nil
}

// HasKnowledgeBase 检查知识库是否存在
func (r *VectorRetriever) HasKnowledgeBase(ctx context.Context, kbID int64) (bool, error) {
	collectionName := r.getCollectionName(kbID)

	has, err := r.client.HasCollection(ctx, collectionName)
	if err != nil {
		return false, fmt.Errorf("check collection exists failed: %w", err)
	}

	return has, nil
}

// GetKnowledgeBaseInfo 获取知识库信息
func (r *VectorRetriever) GetKnowledgeBaseInfo(ctx context.Context, kbID int64) (*entity.Collection, error) {
	collectionName := r.getCollectionName(kbID)

	coll, err := r.client.DescribeCollection(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("describe collection failed: %w", err)
	}

	return coll, nil
}

// ListKnowledgeBase 列出所有知识库
func (r *VectorRetriever) ListKnowledgeBase(ctx context.Context) ([]*entity.Collection, error) {
	collections, err := r.client.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("list collections failed: %w", err)
	}

	return collections, nil
}

// ========================================
// 索引管理
// ========================================

// IndexOptions 索引选项
type IndexOptions struct {
	IndexType   IndexType         // 索引类型
	MetricType  entity.MetricType // 距离度量类型
	IndexParams map[string]string // 索引参数
	FieldName   string            // 字段名称
}

// CreateIndex 创建索引
func (r *VectorRetriever) CreateIndex(ctx context.Context, kbID int64, opts *IndexOptions) error {
	collectionName := r.getCollectionName(kbID)

	// 设置默认值
	if opts.FieldName == "" {
		opts.FieldName = "vector"
	}
	if opts.MetricType == "" {
		opts.MetricType = entity.L2
	}

	// 构建索引
	var idx entity.Index
	var err error

	switch opts.IndexType {
	case IndexTypeFlat, "":
		idx, err = entity.NewIndexFlat(opts.MetricType)
	case IndexTypeIvfFlat:
		nlist := 128
		if val, ok := opts.IndexParams["nlist"]; ok {
			fmt.Sscanf(val, "%d", &nlist)
		}
		idx, err = entity.NewIndexIvfFlat(opts.MetricType, nlist)
	case IndexTypeIvfSq8:
		nlist := 128
		if val, ok := opts.IndexParams["nlist"]; ok {
			fmt.Sscanf(val, "%d", &nlist)
		}
		idx, err = entity.NewIndexIvfSQ8(opts.MetricType, nlist)
	case IndexTypeHnsw:
		M := 16
		efConstruction := 256
		if val, ok := opts.IndexParams["M"]; ok {
			fmt.Sscanf(val, "%d", &M)
		}
		if val, ok := opts.IndexParams["efConstruction"]; ok {
			fmt.Sscanf(val, "%d", &efConstruction)
		}
		idx, err = entity.NewIndexHNSW(opts.MetricType, M, efConstruction)
	default:
		// 默认使用 IVF_FLAT
		nlist := 128
		if val, ok := opts.IndexParams["nlist"]; ok {
			fmt.Sscanf(val, "%d", &nlist)
		}
		idx, err = entity.NewIndexIvfFlat(opts.MetricType, nlist)
	}

	if err != nil {
		return fmt.Errorf("create index object failed: %w", err)
	}

	// 创建索引
	err = r.client.CreateIndex(ctx, collectionName, opts.FieldName, idx, false)
	if err != nil {
		return fmt.Errorf("create index failed: %w", err)
	}

	log.Printf("[Milvus] Index created on %s.%s", collectionName, opts.FieldName)
	return nil
}

// DropIndex 删除索引
func (r *VectorRetriever) DropIndex(ctx context.Context, kbID int64, fieldName string) error {
	collectionName := r.getCollectionName(kbID)

	if fieldName == "" {
		fieldName = "vector"
	}

	err := r.client.DropIndex(ctx, collectionName, fieldName)
	if err != nil {
		return fmt.Errorf("drop index failed: %w", err)
	}

	log.Printf("[Milvus] Index dropped on %s.%s", collectionName, fieldName)
	return nil
}

// DescribeIndex 描述索引
func (r *VectorRetriever) DescribeIndex(ctx context.Context, kbID int64, fieldName string) ([]entity.Index, error) {
	collectionName := r.getCollectionName(kbID)

	if fieldName == "" {
		fieldName = "vector"
	}

	indexes, err := r.client.DescribeIndex(ctx, collectionName, fieldName)
	if err != nil {
		return nil, fmt.Errorf("describe index failed: %w", err)
	}

	return indexes, nil
}

// ========================================
// 向量查询
// ========================================

// SearchResult 搜索结果
type SearchResult struct {
	ID         int64
	Score      float32
	Vector     []float32
	DocumentID int64
	ChunkIndex int
	Content    string
	Metadata   map[string]interface{}
}

// SearchOptions 搜索选项
type SearchOptions struct {
	TopK             int                     // 返回结果数量
	ScoreThreshold   float32                 // 相似度阈值
	MetricType       entity.MetricType       // 距离度量类型
	Expr             string                  // 过滤表达式
	OutputFields     []string                // 输出字段
	ConsistencyLevel entity.ConsistencyLevel // 一致性级别
	VectorFieldName  string                  // 向量字段名称
	SearchParams     map[string]interface{}  // 搜索参数
	IndexType        IndexType               // 索引类型
}

// Search 向量搜索
func (r *VectorRetriever) Search(ctx context.Context, kbID int64, queryText string, opts *SearchOptions) ([]*SearchResult, error) {
	// 1. 生成查询向量
	embeddings, err := r.embedder.EmbedStrings(ctx, []string{queryText})
	if err != nil {
		return nil, fmt.Errorf("embed query failed: %w", err)
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings generated")
	}

	// 2. 转换为 float32
	vectors := make([][]float32, len(embeddings))
	for i, emb := range embeddings {
		vectors[i] = make([]float32, len(emb))
		for j, val := range emb {
			vectors[i][j] = float32(val)
		}
	}

	// 3. 执行搜索
	return r.SearchVectors(ctx, kbID, vectors[0], opts)
}

// SearchVectors 直接使用向量搜索
func (r *VectorRetriever) SearchVectors(ctx context.Context, kbID int64, vector []float32, opts *SearchOptions) ([]*SearchResult, error) {
	collectionName := r.getCollectionName(kbID)

	// 设置默认值
	if opts == nil {
		opts = &SearchOptions{}
	}
	if opts.TopK <= 0 {
		opts.TopK = 10
	}
	if opts.MetricType == "" {
		opts.MetricType = entity.L2
	}
	if opts.VectorFieldName == "" {
		opts.VectorFieldName = "vector"
	}
	if len(opts.OutputFields) == 0 {
		opts.OutputFields = []string{"document_id", "chunk_index", "content", "metadata"}
	}

	// 构建搜索向量
	vectors := []entity.Vector{
		entity.FloatVector(vector),
	}

	// 构建搜索参数
	var sp entity.SearchParam
	switch opts.IndexType {
	case IndexTypeIvfFlat, IndexTypeIvfSq8:
		nprobe := 64
		if val, ok := opts.SearchParams["nprobe"]; ok {
			if v, ok := val.(float64); ok {
				nprobe = int(v)
			}
		}
		sp, _ = entity.NewIndexIvfFlatSearchParam(nprobe)
	case IndexTypeHnsw:
		ef := 64
		if val, ok := opts.SearchParams["ef"]; ok {
			if v, ok := val.(float64); ok {
				ef = int(v)
			}
		}
		sp, _ = entity.NewIndexHNSWSearchParam(ef)
	default:
		// 使用默认搜索参数
		sp, _ = entity.NewIndexFlatSearchParam()
	}

	// 执行搜索
	searchResult, err := r.client.Search(
		ctx,
		collectionName,
		[]string{}, // partitions
		opts.Expr,
		opts.OutputFields,
		vectors,
		opts.VectorFieldName,
		opts.MetricType,
		opts.TopK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// 解析结果
	results := make([]*SearchResult, 0)
	for _, res := range searchResult {
		for i := 0; i < res.ResultCount; i++ {
			// 应用阈值过滤
			if opts.ScoreThreshold > 0 && res.Scores[i] < opts.ScoreThreshold {
				continue
			}

			result := &SearchResult{
				Score: res.Scores[i],
			}

			// 提取字段值
			for _, field := range res.Fields {
				switch field.Name() {
				case "id":
					if ids, ok := field.(*entity.ColumnInt64); ok {
						val, _ := ids.ValueByIdx(i)
						result.ID = val
					}
				case "document_id":
					if ids, ok := field.(*entity.ColumnInt64); ok {
						val, _ := ids.ValueByIdx(i)
						result.DocumentID = val
					}
				case "chunk_index":
					if idxs, ok := field.(*entity.ColumnInt64); ok {
						val, _ := idxs.ValueByIdx(i)
						result.ChunkIndex = int(val)
					}
				case "content":
					if contents, ok := field.(*entity.ColumnVarChar); ok {
						val, _ := contents.ValueByIdx(i)
						result.Content = val
					}
				case "metadata":
					if metadatas, ok := field.(*entity.ColumnJSONBytes); ok {
						val, _ := metadatas.ValueByIdx(i)
						result.Metadata = parseJSONBytes(val)
					}
				}
			}

			results = append(results, result)
		}
	}

	return results, nil
}

// SearchBatchOptions 批量搜索选项
type SearchBatchOptions struct {
	TopK             int                     // 返回结果数量
	ScoreThreshold   float32                 // 相似度阈值
	MetricType       entity.MetricType       // 距离度量类型
	Expr             string                  // 过滤表达式
	OutputFields     []string                // 输出字段
	ConsistencyLevel entity.ConsistencyLevel // 一致性级别
	VectorFieldName  string                  // 向量字段名称
	SearchParams     map[string]interface{}  // 搜索参数
	IndexType        IndexType               // 索引类型
}

// BatchSearch 批量向量搜索
func (r *VectorRetriever) BatchSearch(ctx context.Context, kbID int64, queryTexts []string, opts *SearchBatchOptions) ([][]*SearchResult, error) {
	if len(queryTexts) == 0 {
		return nil, fmt.Errorf("query texts cannot be empty")
	}

	// 1. 生成查询向量
	embeddings, err := r.embedder.EmbedStrings(ctx, queryTexts)
	if err != nil {
		return nil, fmt.Errorf("embed queries failed: %w", err)
	}

	// 2. 转换为 float32
	vectors := make([][]float32, len(embeddings))
	for i, emb := range embeddings {
		vectors[i] = make([]float32, len(emb))
		for j, val := range emb {
			vectors[i][j] = float32(val)
		}
	}

	// 3. 执行批量搜索
	return r.BatchSearchVectors(ctx, kbID, vectors, opts)
}

// BatchSearchVectors 批量向量搜索（直接使用向量）
func (r *VectorRetriever) BatchSearchVectors(ctx context.Context, kbID int64, vectors [][]float32, opts *SearchBatchOptions) ([][]*SearchResult, error) {
	collectionName := r.getCollectionName(kbID)

	// 设置默认值
	if opts == nil {
		opts = &SearchBatchOptions{}
	}
	if opts.TopK <= 0 {
		opts.TopK = 10
	}
	if opts.MetricType == "" {
		opts.MetricType = entity.L2
	}
	if opts.VectorFieldName == "" {
		opts.VectorFieldName = "vector"
	}
	if len(opts.OutputFields) == 0 {
		opts.OutputFields = []string{"document_id", "chunk_index", "content", "metadata"}
	}

	// 构建搜索向量
	searchVectors := make([]entity.Vector, len(vectors))
	for i, vec := range vectors {
		searchVectors[i] = entity.FloatVector(vec)
	}

	// 构建搜索参数
	var sp entity.SearchParam
	switch opts.IndexType {
	case IndexTypeIvfFlat, IndexTypeIvfSq8:
		nprobe := 64
		if val, ok := opts.SearchParams["nprobe"]; ok {
			if v, ok := val.(float64); ok {
				nprobe = int(v)
			}
		}
		sp, _ = entity.NewIndexIvfFlatSearchParam(nprobe)
	case IndexTypeHnsw:
		ef := 64
		if val, ok := opts.SearchParams["ef"]; ok {
			if v, ok := val.(float64); ok {
				ef = int(v)
			}
		}
		sp, _ = entity.NewIndexHNSWSearchParam(ef)
	default:
		// 使用默认搜索参数
		sp, _ = entity.NewIndexFlatSearchParam()
	}

	// 执行批量搜索
	searchResults, err := r.client.Search(
		ctx,
		collectionName,
		[]string{}, // partitions
		opts.Expr,
		opts.OutputFields,
		searchVectors,
		opts.VectorFieldName,
		opts.MetricType,
		opts.TopK,
		sp,
	)
	if err != nil {
		return nil, fmt.Errorf("batch search failed: %w", err)
	}

	// 解析结果
	allResults := make([][]*SearchResult, len(vectors))
	for queryIdx, res := range searchResults {
		results := make([]*SearchResult, 0)
		for i := 0; i < res.ResultCount; i++ {
			// 应用阈值过滤
			if opts.ScoreThreshold > 0 && res.Scores[i] < opts.ScoreThreshold {
				continue
			}

			result := &SearchResult{
				Score: res.Scores[i],
			}

			// 提取字段值
			for _, field := range res.Fields {
				switch field.Name() {
				case "id":
					if ids, ok := field.(*entity.ColumnInt64); ok {
						val, _ := ids.ValueByIdx(i)
						result.ID = val
					}
				case "document_id":
					if ids, ok := field.(*entity.ColumnInt64); ok {
						val, _ := ids.ValueByIdx(i)
						result.DocumentID = val
					}
				case "chunk_index":
					if idxs, ok := field.(*entity.ColumnInt64); ok {
						val, _ := idxs.ValueByIdx(i)
						result.ChunkIndex = int(val)
					}
				case "content":
					if contents, ok := field.(*entity.ColumnVarChar); ok {
						val, _ := contents.ValueByIdx(i)
						result.Content = val
					}
				case "metadata":
					if metadatas, ok := field.(*entity.ColumnJSONBytes); ok {
						val, _ := metadatas.ValueByIdx(i)
						result.Metadata = parseJSONBytes(val)
					}
				}
			}

			results = append(results, result)
		}
		allResults[queryIdx] = results
	}

	return allResults, nil
}

// ========================================
// 数据管理
// ========================================

// InsertData 插入数据
func (r *VectorRetriever) InsertData(ctx context.Context, kbID int64, docs []*DocumentData) error {
	if len(docs) == 0 {
		return fmt.Errorf("documents cannot be empty")
	}

	collectionName := r.getCollectionName(kbID)

	// 构建列数据
	columns := r.buildColumns(docs)

	// 插入数据
	_, err := r.client.Insert(
		ctx,
		collectionName,
		"", // partition name
		columns...,
	)
	if err != nil {
		return fmt.Errorf("insert data failed: %w", err)
	}

	// 刷新数据以确保可搜索
	err = r.client.Flush(ctx, collectionName, false)
	if err != nil {
		return fmt.Errorf("flush collection failed: %w", err)
	}

	log.Printf("[Milvus] Inserted %d documents into %s", len(docs), collectionName)
	return nil
}

// DocumentData 文档数据
type DocumentData struct {
	ID            int64
	Vector        []float32
	DocumentID    int64
	ChunkIndex    int
	Content       string
	Metadata      map[string]interface{}
	DynamicFields map[string]interface{}
}

// buildColumns 构建插入数据的列
func (r *VectorRetriever) buildColumns(docs []*DocumentData) []entity.Column {
	ids := make([]int64, len(docs))
	vectors := make([][]float32, len(docs))
	documentIDs := make([]int64, len(docs))
	chunkIndexes := make([]int64, len(docs))
	contents := make([]string, len(docs))
	metadatas := make([][]byte, len(docs))

	for i, doc := range docs {
		ids[i] = doc.ID
		vectors[i] = doc.Vector
		documentIDs[i] = doc.DocumentID
		chunkIndexes[i] = int64(doc.ChunkIndex)
		contents[i] = doc.Content
		// 将 metadata 转为 JSON 字节数组
		if doc.Metadata != nil {
			metadatas[i], _ = json.Marshal(doc.Metadata)
		} else {
			metadatas[i] = []byte("{}")
		}
	}

	dim := len(vectors[0])
	return []entity.Column{
		entity.NewColumnInt64("id", ids),
		entity.NewColumnFloatVector("vector", dim, vectors),
		entity.NewColumnInt64("document_id", documentIDs),
		entity.NewColumnInt64("chunk_index", chunkIndexes),
		entity.NewColumnVarChar("content", contents),
		entity.NewColumnJSONBytes("metadata", metadatas),
	}
}

// DeleteData 删除数据
func (r *VectorRetriever) DeleteData(ctx context.Context, kbID int64, ids []int64) error {
	if len(ids) == 0 {
		return fmt.Errorf("ids cannot be empty")
	}

	collectionName := r.getCollectionName(kbID)

	// 构建删除表达式
	idStr := ""
	for i, id := range ids {
		if i > 0 {
			idStr += ", "
		}
		idStr += fmt.Sprintf("%d", id)
	}
	expr := fmt.Sprintf("id in [%s]", idStr)

	// 执行删除
	err := r.client.Delete(ctx, collectionName, "", expr)
	if err != nil {
		return fmt.Errorf("delete data failed: %w", err)
	}

	log.Printf("[Milvus] Deleted %d records from %s", len(ids), collectionName)
	return nil
}

// DeleteByExpr 按表达式删除数据
func (r *VectorRetriever) DeleteByExpr(ctx context.Context, kbID int64, expr string) error {
	if expr == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	collectionName := r.getCollectionName(kbID)

	err := r.client.Delete(ctx, collectionName, "", expr)
	if err != nil {
		return fmt.Errorf("delete by expression failed: %w", err)
	}

	log.Printf("[Milvus] Deleted by expr '%s' from %s", expr, collectionName)
	return nil
}

// QueryOptions 查询选项
type QueryOptions struct {
	Expr         []string // 输出字段
	OutputFields []string // 过滤表达式
	Limit        int64    // 限制数量
	Offset       int64    // 偏移量
}

// Query 查询数据
func (r *VectorRetriever) Query(ctx context.Context, kbID int64, opts *QueryOptions) ([]*DocumentData, error) {
	collectionName := r.getCollectionName(kbID)

	// 设置默认值
	if opts == nil {
		opts = &QueryOptions{}
	}
	if len(opts.Expr) == 0 {
		opts.Expr = []string{"id >= 0"} // 匹配所有
	}
	if len(opts.OutputFields) == 0 {
		opts.OutputFields = []string{"id", "document_id", "chunk_index", "content", "metadata"}
	}

	// 执行查询
	queryResult, err := r.client.Query(
		ctx,
		collectionName,
		[]string{}, // partitions
		opts.Expr[0],
		opts.OutputFields,
	)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	// 解析结果
	docs := make([]*DocumentData, 0)
	// 这里简化处理，实际需要根据字段类型解析
	_ = queryResult

	return docs, nil
}

// ========================================
// 字段管理
// ========================================

// AddField 添加字段
// 注意：Milvus 不支持直接向已有 collection 添加字段
func (r *VectorRetriever) AddField(ctx context.Context, kbID int64, fieldName string, fieldType entity.FieldType) error {
	return fmt.Errorf("Milvus does not support adding fields to existing collection. Use dynamic fields or recreate collection")
}

// UpdateField 更新字段
// 注意：Milvus 不支持修改字段
func (r *VectorRetriever) UpdateField(ctx context.Context, kbID int64, fieldName string) error {
	return fmt.Errorf("Milvus does not support updating field definitions")
}

// DropField 删除字段
// 注意：Milvus 不支持删除字段
func (r *VectorRetriever) DropField(ctx context.Context, kbID int64, fieldName string) error {
	return fmt.Errorf("Milvus does not support dropping fields. Use dynamic fields or recreate collection")
}

// ========================================
// 加载和释放
// ========================================

// LoadKnowledgeBase 加载知识库到内存
func (r *VectorRetriever) LoadKnowledgeBase(ctx context.Context, kbID int64, async bool) error {
	collectionName := r.getCollectionName(kbID)

	err := r.client.LoadCollection(ctx, collectionName, async)
	if err != nil {
		return fmt.Errorf("load collection failed: %w", err)
	}

	log.Printf("[Milvus] Collection loaded: %s", collectionName)
	return nil
}

// ReleaseKnowledgeBase 释放知识库内存
func (r *VectorRetriever) ReleaseKnowledgeBase(ctx context.Context, kbID int64) error {
	collectionName := r.getCollectionName(kbID)

	err := r.client.ReleaseCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("release collection failed: %w", err)
	}

	log.Printf("[Milvus] Collection released: %s", collectionName)
	return nil
}

// GetLoadingProgress 获取加载进度
func (r *VectorRetriever) GetLoadingProgress(ctx context.Context, kbID int64) (int64, error) {
	collectionName := r.getCollectionName(kbID)

	progress, err := r.client.GetLoadingProgress(ctx, collectionName, []string{})
	if err != nil {
		return 0, fmt.Errorf("get loading progress failed: %w", err)
	}

	return progress, nil
}

// GetLoadState 获取加载状态
func (r *VectorRetriever) GetLoadState(ctx context.Context, kbID int64) (entity.LoadState, error) {
	collectionName := r.getCollectionName(kbID)

	state, err := r.client.GetLoadState(ctx, collectionName, []string{})
	if err != nil {
		return entity.LoadStateNotExist, fmt.Errorf("get load state failed: %w", err)
	}

	return state, nil
}

// ========================================
// Partition 管理
// ========================================

// CreatePartition 创建分区
func (r *VectorRetriever) CreatePartition(ctx context.Context, kbID int64, partitionName string) error {
	collectionName := r.getCollectionName(kbID)

	err := r.client.CreatePartition(ctx, collectionName, partitionName)
	if err != nil {
		return fmt.Errorf("create partition failed: %w", err)
	}

	log.Printf("[Milvus] Partition created: %s.%s", collectionName, partitionName)
	return nil
}

// DropPartition 删除分区
func (r *VectorRetriever) DropPartition(ctx context.Context, kbID int64, partitionName string) error {
	collectionName := r.getCollectionName(kbID)

	err := r.client.DropPartition(ctx, collectionName, partitionName)
	if err != nil {
		return fmt.Errorf("drop partition failed: %w", err)
	}

	log.Printf("[Milvus] Partition dropped: %s.%s", collectionName, partitionName)
	return nil
}

// ShowPartitions 显示分区列表
func (r *VectorRetriever) ShowPartitions(ctx context.Context, kbID int64) ([]*entity.Partition, error) {
	collectionName := r.getCollectionName(kbID)

	partitions, err := r.client.ShowPartitions(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("show partitions failed: %w", err)
	}

	return partitions, nil
}

// HasPartition 检查分区是否存在
func (r *VectorRetriever) HasPartition(ctx context.Context, kbID int64, partitionName string) (bool, error) {
	collectionName := r.getCollectionName(kbID)

	has, err := r.client.HasPartition(ctx, collectionName, partitionName)
	if err != nil {
		return false, fmt.Errorf("check partition exists failed: %w", err)
	}

	return has, nil
}

// ========================================
// 辅助方法
// ========================================

// getCollectionName 获取集合名称
func (r *VectorRetriever) getCollectionName(kbID int64) string {
	return fmt.Sprintf("kb_%d", kbID)
}

// GetStats 获取集合统计信息
func (r *VectorRetriever) GetStats(ctx context.Context, kbID int64) (map[string]interface{}, error) {
	collectionName := r.getCollectionName(kbID)

	// 获取 collection 信息
	coll, err := r.client.DescribeCollection(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("describe collection failed: %w", err)
	}

	stats := map[string]interface{}{
		"name":        coll.Name,
		"field_count": len(coll.Schema.Fields),
	}

	// 获取实体数量
	entities, err := r.client.GetCollectionStatistics(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("get collection statistics failed: %w", err)
	}
	stats["statistics"] = entities

	return stats, nil
}

// CompactCollection 压缩集合（Milvus 会自动处理）
func (r *VectorRetriever) CompactCollection(ctx context.Context, kbID int64) error {
	// Milvus 会自动进行数据压缩
	// 这个方法保留为兼容性接口
	log.Printf("[Milvus] Collection auto-compaction enabled: %s", r.getCollectionName(kbID))
	return nil
}

// ========================================
// 工具函数
// ========================================

// parseJSONBytes 解析 JSON 字节数组为 map
func parseJSONBytes(data []byte) map[string]interface{} {
	if len(data) == 0 {
		return nil
	}
	var metadata map[string]interface{}
	_ = json.Unmarshal(data, &metadata)
	return metadata
}

// MetadataToJSON 将 metadata 转为 JSON 字符串
func MetadataToJSON(metadata map[string]interface{}) string {
	if metadata == nil {
		return "{}"
	}
	data, _ := json.Marshal(metadata)
	return string(data)
}

// JSONToMetadata 将 JSON 字符串转为 metadata
func JSONToMetadata(jsonStr string) map[string]interface{} {
	if jsonStr == "" {
		return nil
	}
	var metadata map[string]interface{}
	_ = json.Unmarshal([]byte(jsonStr), &metadata)
	return metadata
}

// ========================================
// 健康检查
// ========================================

// CheckHealth 检查 Milvus 连接健康状态
func (r *VectorRetriever) CheckHealth(ctx context.Context) error {
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 尝试列出集合
	_, err := r.client.ListCollections(checkCtx)
	if err != nil {
		return fmt.Errorf("milvus health check failed: %w", err)
	}

	return nil
}

// GetServerVersion 获取 Milvus 服务器版本
func (r *VectorRetriever) GetServerVersion(ctx context.Context) (string, error) {
	return r.client.GetVersion(ctx)
}
