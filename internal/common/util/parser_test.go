package util

import (
	"strings"
	"testing"
)

// TestTextParser 测试纯文本解析器
func TestTextParser(t *testing.T) {
	parser := &TextParser{}

	content := "Hello World!\nThis is a test document."
	doc, err := parser.Parse(strings.NewReader(content))

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if doc.Content != content {
		t.Errorf("Content mismatch: got %q, want %q", doc.Content, content)
	}

	if doc.Metadata == nil {
		t.Fatal("Metadata is nil")
	}

	if doc.Metadata.CharCount != len(content) {
		t.Errorf("CharCount mismatch: got %d, want %d",
			doc.Metadata.CharCount, len(content))
	}

	expectedWords := 7
	if doc.Metadata.WordCount != expectedWords {
		t.Errorf("WordCount mismatch: got %d, want %d",
			doc.Metadata.WordCount, expectedWords)
	}
}

// TestMarkdownParser 测试 Markdown 解析器
func TestMarkdownParser(t *testing.T) {
	parser := &MarkdownParser{}

	content := `# Test Document

This is a test markdown document.

## Subsection

Some content here.
`

	doc, err := parser.Parse(strings.NewReader(content))

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// 检查标题提取
	if doc.Metadata.Title != "Test Document" {
		t.Errorf("Title mismatch: got %q, want %q",
			doc.Metadata.Title, "Test Document")
	}

	// strings.Fields 会将 ## 也当作一个词，所以实际是 14 个词
	// 如果要精确计数，需要更复杂的分词逻辑
	if doc.Metadata.WordCount != 14 {
		t.Errorf("WordCount mismatch: got %d, want 14",
			doc.Metadata.WordCount)
	}
}

// TestHTMLParser 测试 HTML 解析器
func TestHTMLParser(t *testing.T) {
	parser := &HTMLParser{}

	content := `<html>
<head><title>Test Page</title></head>
<body>
<h1>Hello World</h1>
<p>This is a <strong>test</strong> paragraph.</p>
</body>
</html>`

	doc, err := parser.Parse(strings.NewReader(content))

	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// 检查标签是否被去除
	if strings.Contains(doc.Content, "<") && strings.Contains(doc.Content, ">") {
		t.Error("HTML tags not properly stripped")
	}

	// 检查文本内容
	if !strings.Contains(doc.Content, "Hello World") {
		t.Error("Expected content 'Hello World' not found")
	}

	if !strings.Contains(doc.Content, "test paragraph") {
		t.Error("Expected content 'test paragraph' not found")
	}
}

// TestParserRegistry 测试解析器注册表
func TestParserRegistry(t *testing.T) {
	registry := NewParserRegistry()

	// 检查内置解析器
	supportedExts := registry.GetSupportedExtensions()
	expectedExts := []string{".txt", ".text", ".md", ".markdown", ".html", ".htm"}

	for _, expected := range expectedExts {
		found := false
		for _, ext := range supportedExts {
			if ext == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected extension %s not found in supported extensions", expected)
		}
	}
}

// TestDocumentParserFactory 测试解析器工厂
func TestDocumentParserFactory(t *testing.T) {
	factory := NewDocumentParserFactory()

	// 测试从字符串解析
	content := "Test content for factory"
	doc, err := factory.ParseFromString(content, ".txt")

	if err != nil {
		t.Fatalf("ParseFromString failed: %v", err)
	}

	if doc.Content != content {
		t.Errorf("Content mismatch: got %q, want %q", doc.Content, content)
	}

	// 测试支持的格式检查 (IsSupported 期望带点的扩展名)
	if !factory.IsSupported(".txt") {
		t.Error(".txt should be supported")
	}

	if factory.IsSupported(".xyz") {
		t.Error(".xyz should not be supported")
	}
}

// TestChunkDocument 测试文档分块
func TestChunkDocument(t *testing.T) {
	content := strings.Repeat("word ", 200) // 1000 字符
	doc := &Document{
		Content: content,
	}

	opts := &ChunkOptions{
		ChunkSize: 200,
		Overlap:   20,
	}

	chunks := ChunkDocument(doc, opts)

	// 验证分块数量
	if len(chunks) == 0 {
		t.Fatal("No chunks generated")
	}

	// 验证每个块的大小
	for i, chunk := range chunks {
		if len(chunk) > opts.ChunkSize+50 { // 允许一些误差
			t.Errorf("Chunk %d too large: %d chars", i, len(chunk))
		}
	}

	// 验证内容完整性（分块可能有重叠，所以总长度应该大于等于原长度）
	totalLength := 0
	for _, chunk := range chunks {
		totalLength += len(chunk)
	}

	if totalLength < len(content) {
		t.Errorf("Content loss: got %d chars, want %d chars",
			totalLength, len(content))
	}
}

// TestChunkDocumentSmall 测试小文档分块
func TestChunkDocumentSmall(t *testing.T) {
	content := "Small document"
	doc := &Document{
		Content: content,
	}

	chunks := ChunkDocument(doc, DefaultChunkOptions())

	// 小文档不应该被分块
	if len(chunks) != 1 {
		t.Errorf("Expected 1 chunk for small document, got %d", len(chunks))
	}

	if chunks[0] != content {
		t.Errorf("Content mismatch: got %q, want %q", chunks[0], content)
	}
}

// TestExtractMarkdownTitle 测试 Markdown 标题提取
func TestExtractMarkdownTitle(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		expectedTitle string
	}{
		{
			name:          "H1 title",
			content:       "# Main Title\n\nContent here",
			expectedTitle: "Main Title",
		},
		{
			name:          "H2 title",
			content:       "## Sub Title\n\nContent here",
			expectedTitle: "Sub Title",
		},
		{
			name:          "No title",
			content:       "Just some content\nwithout title",
			expectedTitle: "",
		},
		{
			name:          "Title with spaces",
			content:       "#    Title with spaces   \n\nContent",
			expectedTitle: "Title with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title := extractMarkdownTitle(tt.content)
			if title != tt.expectedTitle {
				t.Errorf("extractMarkdownTitle() = %q, want %q", title, tt.expectedTitle)
			}
		})
	}
}

// TestStripHTMLTags 测试 HTML 标签去除
func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Simple tags",
			html:     "<p>Hello World</p>",
			expected: "Hello World",
		},
		{
			name:     "Nested tags",
			html:     "<div><p>Nested <strong>content</strong></p></div>",
			expected: "Nested content",
		},
		{
			name:     "HTML entities",
			html:     "Hello &nbsp; World &lt;3&gt;",
			expected: "Hello World <3>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripHTMLTags(tt.html)
			if result != tt.expected {
				t.Errorf("stripHTMLTags() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestCountWords 测试字数统计
func TestCountWords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "Simple text",
			text:     "Hello World",
			expected: 2,
		},
		{
			name:     "Multiple spaces",
			text:     "Hello    World   Test",
			expected: 3,
		},
		{
			name:     "Newlines",
			text:     "Line1\nLine2\nLine3",
			expected: 3,
		},
		{
			name:     "Empty",
			text:     "",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countWords(tt.text)
			if result != tt.expected {
				t.Errorf("countWords() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// BenchmarkTextParser 基准测试纯文本解析器
func BenchmarkTextParser(b *testing.B) {
	parser := &TextParser{}
	content := strings.Repeat("test content ", 1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parser.Parse(strings.NewReader(content))
	}
}

// BenchmarkChunkDocument 基准测试文档分块
func BenchmarkChunkDocument(b *testing.B) {
	content := strings.Repeat("word ", 1000) // 减小避免内存溢出
	doc := &Document{Content: content}
	opts := DefaultChunkOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ChunkDocument(doc, opts)
	}
}
