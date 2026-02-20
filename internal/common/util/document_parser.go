package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DocumentParser 文档解析器接口
type DocumentParser interface {
	// Parse 解析文档并返回内容和元数据
	Parse(reader io.Reader) (*Document, error)

	// Supports 检查是否支持该文件格式
	Supports(ext string) bool

	// GetSupportedExtensions 获取支持的文件扩展名
	GetSupportedExtensions() []string
}

// Document 解析后的文档
type Document struct {
	// Content 文档内容
	Content string `json:"content"`

	// Metadata 文档元数据
	Metadata *DocumentMetadata `json:"metadata,omitempty"`

	// Error 解析错误（如果有）
	Error error `json:"error,omitempty"`
}

// DocumentMetadata 文档元数据
type DocumentMetadata struct {
	// Title 文档标题
	Title string `json:"title,omitempty"`

	// Author 文档作者
	Author string `json:"author,omitempty"`

	// Subject 文档主题
	Subject string `json:"subject,omitempty"`

	// Keywords 关键词
	Keywords []string `json:"keywords,omitempty"`

	// CreationDate 创建时间
	CreationDate string `json:"creation_date,omitempty"`

	// ModificationDate 修改时间
	ModificationDate string `json:"modification_date,omitempty"`

	// PageCount 页数（对于支持页数的格式）
	PageCount int `json:"page_count,omitempty"`

	// WordCount 字数
	WordCount int `json:"word_count"`

	// CharCount 字符数
	CharCount int `json:"char_count"`

	// Language 语言（自动检测）
	Language string `json:"language,omitempty"`
}

// ParserRegistry 解析器注册表
type ParserRegistry struct {
	parsers map[string]DocumentParser
}

// NewParserRegistry 创建解析器注册表
func NewParserRegistry() *ParserRegistry {
	registry := &ParserRegistry{
		parsers: make(map[string]DocumentParser),
	}

	// 注册内置解析器
	registry.Register(&TextParser{})
	registry.Register(&MarkdownParser{})
	registry.Register(&HTMLParser{})

	return registry
}

// Register 注册解析器
func (r *ParserRegistry) Register(parser DocumentParser) {
	exts := parser.GetSupportedExtensions()
	for _, ext := range exts {
		r.parsers[ext] = parser
	}
}

// GetParser 根据文件扩展名获取解析器
func (r *ParserRegistry) GetParser(ext string) (DocumentParser, error) {
	ext = strings.ToLower(ext)
	if parser, exists := r.parsers[ext]; exists {
		return parser, nil
	}
	return nil, fmt.Errorf("unsupported file format: %s", ext)
}

// ParseFile 解析文件
func (r *ParserRegistry) ParseFile(filePath string) (*Document, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	ext := strings.TrimPrefix(filepath.Ext(filePath), ".")
	parser, err := r.GetParser(ext)
	if err != nil {
		return nil, err
	}

	doc, err := parser.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	// 如果没有标题，使用文件名作为标题
	if doc.Metadata != nil && doc.Metadata.Title == "" {
		doc.Metadata.Title = filepath.Base(filePath)
	}

	return doc, nil
}

// GetSupportedExtensions 获取支持的文件扩展名
func (r *ParserRegistry) GetSupportedExtensions() []string {
	exts := make([]string, 0, len(r.parsers))
	for ext := range r.parsers {
		exts = append(exts, ext)
	}
	return exts
}

// ========================================
// 文本解析器
// ========================================

// TextParser 纯文本解析器
type TextParser struct{}

// Parse 解析纯文本
func (p *TextParser) Parse(reader io.Reader) (*Document, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	text := string(content)

	return &Document{
		Content: text,
		Metadata: &DocumentMetadata{
			CharCount: len(text),
			WordCount: countWords(text),
		},
	}, nil
}

// Supports 检查是否支持该格式
func (p *TextParser) Supports(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".txt" || ext == ".text"
}

// GetSupportedExtensions 获取支持的扩展名
func (p *TextParser) GetSupportedExtensions() []string {
	return []string{".txt", ".text"}
}

// ========================================
// Markdown 解析器
// ========================================

// MarkdownParser Markdown 解析器
type MarkdownParser struct{}

// Parse 解析 Markdown
func (p *MarkdownParser) Parse(reader io.Reader) (*Document, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	text := string(content)

	// 提取标题作为文档标题
	title := extractMarkdownTitle(text)

	return &Document{
		Content: text,
		Metadata: &DocumentMetadata{
			Title:     title,
			CharCount: len(text),
			WordCount: countWords(text),
		},
	}, nil
}

// Supports 检查是否支持该格式
func (p *MarkdownParser) Supports(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".md" || ext == ".markdown"
}

// GetSupportedExtensions 获取支持的扩展名
func (p *MarkdownParser) GetSupportedExtensions() []string {
	return []string{".md", ".markdown"}
}

// ========================================
// HTML 解析器
// ========================================

// HTMLParser HTML 解析器
type HTMLParser struct{}

// Parse 解析 HTML（去除标签）
func (p *HTMLParser) Parse(reader io.Reader) (*Document, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	html := string(content)
	text := stripHTMLTags(html)

	return &Document{
		Content: text,
		Metadata: &DocumentMetadata{
			CharCount: len(text),
			WordCount: countWords(text),
		},
	}, nil
}

// Supports 检查是否支持该格式
func (p *HTMLParser) Supports(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".html" || ext == ".htm"
}

// GetSupportedExtensions 获取支持的扩展名
func (p *HTMLParser) GetSupportedExtensions() []string {
	return []string{".html", ".htm"}
}

// ========================================
// 辅助函数
// ========================================

// countWords 统计字数
func countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}

// extractMarkdownTitle 提取 Markdown 标题
func extractMarkdownTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// 提取第一个 # 标题，去除所有 # 前缀
			title := strings.TrimLeft(line, "#")
			title = strings.TrimSpace(title)
			if title != "" {
				return title
			}
		}
	}
	return ""
}

// stripHTMLTags 去除 HTML 标签
func stripHTMLTags(html string) string {
	// 简单的 HTML 标签去除实现
	// 对于更复杂的需求，可以使用 golang.org/x/net/html 包
	var result strings.Builder
	inTag := false

	for _, r := range html {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				result.WriteRune(r)
			}
		}
	}

	text := result.String()
	// 替换 HTML 实体
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	// 清理多余的空白（将连续空格压缩为单个）
	text = strings.Join(strings.Fields(text), " ")

	lines := strings.Split(text, "\n")
	cleanedLines := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	return strings.Join(cleanedLines, "\n")
}
