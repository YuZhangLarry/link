package chunker

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
)

// Chunker 文档分片器接口
type Chunker interface {
	// Split 将文档分割成多个片段
	Split(ctx context.Context, document string) ([]string, error)

	// SplitDocuments 将多个文档分割成片段
	SplitDocuments(ctx context.Context, docs []*schema.Document) ([]*schema.Document, error)

	// SplitWithIDs 将文档分割成片段，并返回带ID的文档
	SplitWithIDs(ctx context.Context, documentID, content string) ([]*schema.Document, error)
}

// SemanticChunker 语义分片器
type SemanticChunker struct {
	splitter document.Transformer
	config   *Config
}

// Config 语义分片器配置
type Config struct {
	// Embedding 嵌入器
	Embedding embedding.Embedder

	// BufferSize 上下文缓冲区大小，用于计算语义相似度时包含更多上下文
	BufferSize int

	// MinChunkSize 最小片段大小（字符数）
	MinChunkSize int

	// MaxChunkSize 最大片段大小（字符数），用于递归分割
	MaxChunkSize int

	// Separators 分隔符列表，按优先级顺序使用
	Separators []string

	// Percentile 分割阈值的百分位数，范围 0-1，越大分割越少
	Percentile float64

	// Overlap 片段之间的重叠大小（字符数）
	Overlap int
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		BufferSize:   2,
		MinChunkSize: 100,
		MaxChunkSize: 2000,
		Separators:   []string{"\n\n", "\n", "。", "！", "？", ".", "!"},
		Percentile:   0.9,
		Overlap:      50,
	}
}

// NewSemanticChunker 创建语义分片器
func NewSemanticChunker(ctx context.Context, config *Config) (*SemanticChunker, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if config.Embedding == nil {
		return nil, fmt.Errorf("embedding is required")
	}

	// 创建 eino 语义分割器
	splitter, err := semantic.NewSplitter(ctx, &semantic.Config{
		Embedding:    config.Embedding,
		BufferSize:   config.BufferSize,
		MinChunkSize: config.MinChunkSize,
		Separators:   config.Separators,
		Percentile:   config.Percentile,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create semantic splitter: %w", err)
	}

	return &SemanticChunker{
		splitter: splitter,
		config:   config,
	}, nil
}

// Split 将文档分割成多个片段
func (c *SemanticChunker) Split(ctx context.Context, document string) ([]string, error) {
	if document == "" {
		return []string{}, nil
	}

	docs := []*schema.Document{
		{
			ID:      "doc_0",
			Content: document,
		},
	}

	results, err := c.splitter.Transform(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("failed to split document: %w", err)
	}

	chunks := make([]string, 0, len(results))
	for _, doc := range results {
		chunks = append(chunks, doc.Content)
	}

	log.Printf("[Chunker] Split document into %d chunks", len(chunks))
	return chunks, nil
}

// SplitDocuments 将多个文档分割成片段
func (c *SemanticChunker) SplitDocuments(ctx context.Context, documents []*schema.Document) ([]*schema.Document, error) {
	if len(documents) == 0 {
		return []*schema.Document{}, nil
	}

	results, err := c.splitter.Transform(ctx, documents)
	if err != nil {
		return nil, fmt.Errorf("failed to split documents: %w", err)
	}

	log.Printf("[Chunker] Split %d documents into %d chunks", len(documents), len(results))
	return results, nil
}

// SplitWithIDs 将文档分割成片段，并返回带ID的文档
func (c *SemanticChunker) SplitWithIDs(ctx context.Context, documentID, content string) ([]*schema.Document, error) {
	chunks, err := c.Split(ctx, content)
	if err != nil {
		return nil, err
	}

	docs := make([]*schema.Document, 0, len(chunks))
	for i, chunk := range chunks {
		docs = append(docs, &schema.Document{
			ID:      fmt.Sprintf("%s_chunk_%d", documentID, i),
			Content: chunk,
		})
	}

	return docs, nil
}

// ========================================
// 简单分片器（基于分隔符，无需 Embedding）
// ========================================

// SimpleChunker 简单分片器，基于固定分隔符和大小
type SimpleChunker struct {
	config *SimpleConfig
}

// SimpleConfig 简单分片器配置
type SimpleConfig struct {
	// ChunkSize 每个片段的字符数
	ChunkSize int

	// Overlap 片段之间的重叠字符数
	Overlap int

	// Separators 分隔符列表，优先在这些分隔符处分割
	Separators []string

	// KeepSeparator 是否保留分隔符
	KeepSeparator bool
}

// DefaultSimpleConfig 返回默认简单配置
func DefaultSimpleConfig() *SimpleConfig {
	return &SimpleConfig{
		ChunkSize:     1000,
		Overlap:       100,
		Separators:    []string{"\n\n", "\n", "。", "！", "？"},
		KeepSeparator: true,
	}
}

// NewSimpleChunker 创建简单分片器
func NewSimpleChunker(config *SimpleConfig) *SimpleChunker {
	if config == nil {
		config = DefaultSimpleConfig()
	}

	return &SimpleChunker{
		config: config,
	}
}

// Split 将文档分割成多个片段
func (c *SimpleChunker) Split(ctx context.Context, document string) ([]string, error) {
	if document == "" {
		return []string{}, nil
	}

	chunks := c.splitBySize(document)

	log.Printf("[Chunker] Split document into %d chunks (simple)", len(chunks))
	return chunks, nil
}

// SplitDocuments 将多个文档分割成片段
func (c *SimpleChunker) SplitDocuments(ctx context.Context, documents []*schema.Document) ([]*schema.Document, error) {
	results := make([]*schema.Document, 0)

	for _, doc := range documents {
		chunks, err := c.Split(ctx, doc.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to split document %s: %w", doc.ID, err)
		}

		for i, chunk := range chunks {
			results = append(results, &schema.Document{
				ID:      fmt.Sprintf("%s_chunk_%d", doc.ID, i),
				Content: chunk,
			})
		}
	}

	return results, nil
}

// SplitWithIDs 将文档分割成片段，并返回带ID的文档
func (c *SimpleChunker) SplitWithIDs(ctx context.Context, documentID, content string) ([]*schema.Document, error) {
	chunks, err := c.Split(ctx, content)
	if err != nil {
		return nil, err
	}

	docs := make([]*schema.Document, 0, len(chunks))
	for i, chunk := range chunks {
		docs = append(docs, &schema.Document{
			ID:      fmt.Sprintf("%s_chunk_%d", documentID, i),
			Content: chunk,
		})
	}

	return docs, nil
}

// splitBySize 按大小分割文档（使用 rune 正确处理 UTF-8）
func (c *SimpleChunker) splitBySize(document string) []string {
	// 转换为 rune 切片以正确处理 UTF-8 字符（尤其是中文）
	runes := []rune(document)
	docLen := len(runes)

	if docLen <= c.config.ChunkSize {
		return []string{document}
	}

	chunks := []string{}
	start := 0
	overlap := c.config.Overlap

	if overlap >= c.config.ChunkSize {
		overlap = c.config.ChunkSize / 2
	}

	// 防止无限循环：记录上一次的 start 位置
	lastStart := -1
	iterations := 0
	// 设置最大迭代次数（安全限制：假设每块至少1个字符）
	maxIterations := docLen*2 + 1000

	for start < docLen && iterations < maxIterations {
		iterations++

		// 检测是否陷入无限循环
		if start <= lastStart {
			log.Printf("[Chunker] Warning: No progress detected at iteration %d, start=%d, adding remaining text", iterations, start)
			if start < docLen {
				chunks = append(chunks, string(runes[start:]))
			}
			break
		}
		lastStart = start

		end := start + c.config.ChunkSize
		if end > docLen {
			end = docLen
		}

		// 尝试在分隔符处找到更好的分割点
		if end < docLen {
			textSegment := string(runes[start:end])
			bestBreak := c.findBestBreakPoint(textSegment, end-start)
			if bestBreak > 0 {
				end = start + bestBreak
			}
		}

		chunks = append(chunks, string(runes[start:end]))

		// 计算新的 start 位置，确保至少前进1个字符
		newStart := end - overlap
		minProgress := 1 // 最小前进距离
		if newStart < start+minProgress {
			newStart = start + minProgress
		}
		if newStart > docLen {
			newStart = docLen
		}
		start = newStart
	}

	if iterations >= maxIterations {
		log.Printf("[Chunker] Warning: Reached max iterations (%d), possible infinite loop prevented", maxIterations)
	}

	return chunks
}

// findBestBreakPoint 在文本中查找最佳分割点
func (c *SimpleChunker) findBestBreakPoint(text string, maxLen int) int {
	// 从后往前搜索分隔符
	for _, sep := range c.config.Separators {
		// 在文本中查找最后一个出现的分隔符
		searchText := text
		if len(searchText) > maxLen {
			searchText = searchText[:maxLen]
		}

		lastIndex := strings.LastIndex(searchText, sep)
		if lastIndex > 0 && lastIndex < maxLen {
			sepLen := len(sep)
			if c.config.KeepSeparator {
				return lastIndex + sepLen
			}
			return lastIndex
		}
	}

	return -1
}
