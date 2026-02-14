package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ParserType 解析器类型
type ParserType string

const (
	ParserTypeAuto     ParserType = "auto"     // 自动检测
	ParserTypeText     ParserType = "text"     // 纯文本
	ParserTypePDF      ParserType = "pdf"      // PDF
	ParserTypeWord     ParserType = "word"     // Word (DOCX)
	ParserTypeExcel    ParserType = "excel"    // Excel (XLSX)
	ParserTypeHTML     ParserType = "html"     // HTML
	ParserTypeMarkdown ParserType = "markdown" // Markdown
)

// ParseOptions 解析选项
type ParseOptions struct {
	// ParserType 解析器类型（auto 时自动检测）
	ParserType ParserType

	// ExtractMetadata 是否提取元数据
	ExtractMetadata bool

	// PreserveFormatting 是否保留格式
	PreserveFormatting bool

	// Encoding 文件编码（用于文本文件）
	Encoding string
}

// DocumentParserFactory 文档解析器工厂
type DocumentParserFactory struct {
	registry *ParserRegistry
}

// NewDocumentParserFactory 创建解析器工厂
func NewDocumentParserFactory() *DocumentParserFactory {
	return &DocumentParserFactory{
		registry: NewParserRegistry(),
	}
}

// ParseFromString 从字符串解析文档
func (f *DocumentParserFactory) ParseFromString(content, fileType string) (*Document, error) {
	reader := strings.NewReader(content)
	parser, err := f.registry.GetParser(fileType)
	if err != nil {
		return nil, err
	}

	return parser.Parse(reader)
}

// ParseFromBytes 从字节数组解析文档
func (f *DocumentParserFactory) ParseFromBytes(content []byte, fileType string) (*Document, error) {
	reader := strings.NewReader(string(content))
	parser, err := f.registry.GetParser(fileType)
	if err != nil {
		return nil, err
	}

	return parser.Parse(reader)
}

// ParseFromFile 从文件解析文档
func (f *DocumentParserFactory) ParseFromFile(filePath string, opts *ParseOptions) (*Document, error) {
	if opts == nil {
		opts = &ParseOptions{
			ParserType:      ParserTypeAuto,
			ExtractMetadata: true,
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	// 确定文件类型
	ext := strings.TrimPrefix(filepath.Ext(filePath), ".")
	if opts.ParserType == ParserTypeAuto {
		// 自动检测
		parser, err := f.registry.GetParser("." + ext)
		if err != nil {
			return nil, fmt.Errorf("unsupported file type: %s", ext)
		}
		return f.parseWithParser(filePath, parser)
	}

	// 使用指定的解析器
	return f.registry.ParseFile(filePath)
}

// parseWithParser 使用指定解析器解析文件
func (f *DocumentParserFactory) parseWithParser(filePath string, parser DocumentParser) (*Document, error) {
	return f.registry.ParseFile(filePath)
}

// GetSupportedExtensions 获取支持的文件扩展名
func (f *DocumentParserFactory) GetSupportedExtensions() []string {
	return f.registry.GetSupportedExtensions()
}

// IsSupported 检查文件类型是否支持
func (f *DocumentParserFactory) IsSupported(ext string) bool {
	parser, err := f.registry.GetParser(ext)
	return err == nil && parser != nil
}

// ========================================
// 批量解析
// ========================================

// BatchParseResult 批量解析结果
type BatchParseResult struct {
	// Documents 成功解析的文档
	Documents []*Document `json:"documents"`

	// Errors 解析失败的文件路径和错误
	Errors map[string]error `json:"errors"`

	// Total 总数
	Total int `json:"total"`

	// Success 成功数
	Success int `json:"success"`

	// Failed 失败数
	Failed int `json:"failed"`
}

// BatchParseFiles 批量解析文件
func (f *DocumentParserFactory) BatchParseFiles(filePaths []string) *BatchParseResult {
	result := &BatchParseResult{
		Documents: make([]*Document, 0, len(filePaths)),
		Errors:    make(map[string]error),
		Total:     len(filePaths),
	}

	for _, filePath := range filePaths {
		doc, err := f.ParseFromFile(filePath, nil)
		if err != nil {
			result.Errors[filePath] = err
			result.Failed++
		} else {
			result.Documents = append(result.Documents, doc)
			result.Success++
		}
	}

	return result
}

// BatchParseDirectory 批量解析目录下的所有支持文件
func (f *DocumentParserFactory) BatchParseDirectory(dirPath string, recursive bool) *BatchParseResult {
	result := &BatchParseResult{
		Documents: make([]*Document, 0),
		Errors:    make(map[string]error),
	}

	// 遍历目录
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			if !recursive && path != dirPath {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件扩展名
		ext := strings.TrimPrefix(filepath.Ext(path), ".")
		if f.IsSupported(ext) {
			result.Total++
			doc, err := f.ParseFromFile(path, nil)
			if err != nil {
				result.Errors[path] = err
				result.Failed++
			} else {
				result.Documents = append(result.Documents, doc)
				result.Success++
			}
		}

		return nil
	}

	filepath.Walk(dirPath, walkFn)
	return result
}

// ========================================
// 文档分块工具
// ========================================

// ChunkOptions 分块选项
type ChunkOptions struct {
	// ChunkSize 每块的最大字符数
	ChunkSize int

	// Overlap 块之间的重叠字符数
	Overlap int

	// Separator 分隔符
	Separator string
}

// DefaultChunkOptions 默认分块选项
func DefaultChunkOptions() *ChunkOptions {
	return &ChunkOptions{
		ChunkSize: 1000,
		Overlap:   100,
		Separator: "\n\n",
	}
}

// ChunkDocument 将文档内容分块
func ChunkDocument(doc *Document, opts *ChunkOptions) []string {
	if opts == nil {
		opts = DefaultChunkOptions()
	}

	content := doc.Content
	if len(content) <= opts.ChunkSize {
		return []string{content}
	}

	chunks := []string{}
	start := 0
	overlap := opts.Overlap

	if overlap >= opts.ChunkSize {
		overlap = opts.ChunkSize / 2
	}

	for start < len(content) {
		end := start + opts.ChunkSize
		if end > len(content) {
			end = len(content)
		}

		chunks = append(chunks, content[start:end])

		// 如果已经到达文档末尾，退出循环
		if end >= len(content) {
			break
		}

		start = end - overlap
		// 确保 start 向前推进
		if start <= 0 {
			start = end
		}
	}

	return chunks
}

// ChunkDocuments 批量分块多个文档
func ChunkDocuments(docs []*Document, opts *ChunkOptions) map[string][]string {
	result := make(map[string][]string)

	for _, doc := range docs {
		if doc.Metadata != nil && doc.Metadata.Title != "" {
			result[doc.Metadata.Title] = ChunkDocument(doc, opts)
		} else {
			result[fmt.Sprintf("doc_%d", len(result))] = ChunkDocument(doc, opts)
		}
	}

	return result
}
