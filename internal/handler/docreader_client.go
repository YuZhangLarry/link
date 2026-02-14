package handler

import (
	"context"
	"link/internal/application/chunker"
	"link/internal/common/util"
	"log"
)

// DocreaderClient 文档解析客户端
// 用于调用文档解析和分片服务
type DocreaderClient struct {
	chunker *chunker.SimpleChunker
}

// NewDocreaderClient 创建文档解析客户端
func NewDocreaderClient() *DocreaderClient {
	return &DocreaderClient{
		chunker: chunker.NewSimpleChunker(chunker.DefaultSimpleConfig()),
	}
}

// ParseDocument 解析文档为分片
func (c *DocreaderClient) ParseDocument(ctx context.Context, content string) ([]string, error) {
	// 使用 chunker 进行分片
	chunks, err := c.chunker.Split(ctx, content)
	if err != nil {
		return nil, err
	}

	log.Printf("[Docreader] Parsed document into %d chunks", len(chunks))
	return chunks, nil
}

// ParseDocumentWithMeta 解析文档并返回元数据
func (c *DocreaderClient) ParseDocumentWithMeta(ctx context.Context, content string) (*util.Document, error) {
	chunks, err := c.ParseDocument(ctx, content)
	if err != nil {
		return nil, err
	}

	// 构建文档对象
	doc := &util.Document{
		Content: content,
		Metadata: &util.DocumentMetadata{
			Title:     "Document",
			CharCount: len(content),
			WordCount: 0, // TODO: 使用更好的分词
		},
	}

	log.Printf("[Docreader] Parsed document: %d chunks, %d chars", len(chunks), len(content))
	return doc, nil
}
