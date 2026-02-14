# Document Parser - 文档解析工具

## 概述

文档解析工具提供了统一的接口来解析多种文档格式，包括纯文本、Markdown、HTML、PDF 等。

## 功能特性

### 1. 支持的文档格式

| 格式 | 扩展名 | 状态 | 说明 |
|------|---------|------|------|
| 纯文本 | `.txt`, `.text` | ✅ 完全支持 | 基础文本格式 |
| Markdown | `.md`, `.markdown` | ✅ 完全支持 | 支持 GitHub Flavored Markdown |
| HTML | `.html`, `.htm` | ✅ 基础支持 | 自动去除标签，提取文本 |
| PDF | `.pdf` | 🚧 框架完成 | 需集成 PDF 库 |
| Word | `.docx` | 📋 计划中 | 需集成 DOCX 库 |
| Excel | `.xlsx` | 📋 计划中 | 需集成 Excel 库 |

### 2. 核心功能

- **统一接口**: 所有解析器实现相同的 `DocumentParser` 接口
- **元数据提取**: 自动提取标题、作者、字数等信息
- **批量处理**: 支持批量解析多个文件或整个目录
- **文档分块**: 将长文档分割成适合处理的块

## 使用示例

### 基础使用

```go
package main

import (
    "fmt"
    "link/internal/common/util"
)

func main() {
    // 创建解析器工厂
    factory := util.NewDocumentParserFactory()

    // 解析单个文件
    doc, err := factory.ParseFromFile("document.txt", nil)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Title: %s\n", doc.Metadata.Title)
    fmt.Printf("Content length: %d chars\n", doc.Metadata.CharCount)
    fmt.Printf("Word count: %d\n", doc.Metadata.WordCount)
}
```

### 批量解析

```go
// 批量解析多个文件
filePaths := []string{
    "doc1.txt",
    "doc2.md",
    "doc3.html",
}

result := factory.BatchParseFiles(filePaths)
fmt.Printf("Parsed %d documents, %d failed\n",
    result.Success, result.Failed)

for _, doc := range result.Documents {
    fmt.Printf("- %s (%d words)\n",
        doc.Metadata.Title, doc.Metadata.WordCount)
}
```

### 批量解析目录

```go
// 解析整个目录（递归）
result := factory.BatchParseDirectory("/path/to/docs", true)
fmt.Printf("Found and parsed %d documents\n", result.Success)

// 或仅解析顶层文件
result = factory.BatchParseDirectory("/path/to/docs", false)
```

### 文档分块

```go
doc, _ := factory.ParseFromFile("large_doc.txt", nil)

// 使用默认选项分块
chunks := util.ChunkDocument(doc, nil)
fmt.Printf("Split into %d chunks\n", len(chunks))

// 自定义分块选项
opts := &util.ChunkOptions{
    ChunkSize: 500,  // 每块 500 字符
    Overlap:   50,    // 重叠 50 字符
    Separator: "\n---\n",
}
chunks = util.ChunkDocument(doc, opts)

for i, chunk := range chunks {
    fmt.Printf("Chunk %d: %d chars\n", i+1, len(chunk))
}
```

### 与图谱提取结合使用

```go
import (
    "context"
    "link/internal/application/service"
    "link/internal/common/util"
)

func main() {
    ctx := context.Background()

    // 1. 创建解析器工厂
    factory := util.NewDocumentParserFactory()

    // 2. 解析文档
    doc, err := factory.ParseFromFile("knowledge.txt", nil)
    if err != nil {
        panic(err)
    }

    // 3. 分块
    chunks := util.ChunkDocument(doc, nil)

    // 4. 为每个分片创建图谱提取输入
    inputs := make([]*service.ChunkExtractionInput, 0, len(chunks))
    for i, chunk := range chunks {
        inputs = append(inputs, &service.ChunkExtractionInput{
            ChunkID:  fmt.Sprintf("chunk_%03d", i),
            Document: chunk,
            Query:    "提取实体和关系",
        })
    }

    // 5. 提取图谱（需要 GraphService 实例）
    // graphService := service.NewGraphService(graphRepo)
    // graphData, _ := graphService.ExtractGraphFromChunks(ctx, inputs)
}
```

## API 参考

### DocumentParserFactory

| 方法 | 说明 |
|------|------|
| `NewDocumentParserFactory()` | 创建工厂实例 |
| `ParseFromFile(path, opts)` | 从文件解析 |
| `ParseFromString(content, type)` | 从字符串解析 |
| `BatchParseFiles(paths)` | 批量解析文件 |
| `BatchParseDirectory(path, recursive)` | 批量解析目录 |
| `GetSupportedExtensions()` | 获取支持的扩展名 |
| `IsSupported(ext)` | 检查是否支持某格式 |

### Document

| 字段 | 类型 | 说明 |
|------|------|------|
| `Content` | `string` | 文档内容 |
| `Metadata` | `*DocumentMetadata` | 文档元数据 |
| `Error` | `error` | 解析错误 |

### DocumentMetadata

| 字段 | 类型 | 说明 |
|------|------|------|
| `Title` | `string` | 文档标题 |
| `Author` | `string` | 作者 |
| `Subject` | `string` | 主题 |
| `Keywords` | `[]string` | 关键词 |
| `CreationDate` | `string` | 创建时间 |
| `PageCount` | `int` | 页数 |
| `WordCount` | `int` | 字数 |
| `CharCount` | `int` | 字符数 |
| `Language` | `string` | 语言 |

## 扩展指南

### 添加新的文档格式支持

1. **实现 DocumentParser 接口**:

```go
type MyCustomParser struct{}

func (p *MyCustomParser) Parse(reader io.Reader) (*util.Document, error) {
    // 实现解析逻辑
    return &util.Document{
        Content: "extracted content",
        Metadata: &util.DocumentMetadata{
            Title: "Document Title",
            CharCount: 100,
            WordCount: 20,
        },
    }, nil
}

func (p *MyCustomParser) Supports(ext string) bool {
    return ext == ".custom"
}

func (p *MyCustomParser) GetSupportedExtensions() []string {
    return []string{".custom"}
}
```

2. **注册解析器**:

```go
factory := util.NewDocumentParserFactory()
factory.registry.Register(&MyCustomParser{})
```

### 集成 PDF 解析库

推荐使用以下库之一：

1. **unidoc/unipdf** (功能强大，有商业限制)
   ```bash
   go get github.com/unidoc/unidoc/pdf
   ```

2. **ledongthuc/pdf** (轻量级，纯 Go)
   ```bash
   go get github.com/ledongthuc/pdf
   ```

3. **rsc/pdf** (简单，Google 出品)
   ```bash
   go get rsc.io/pdf
   ```

实现步骤：

1. 安装选择的库
2. 实现 `pdf_parser.go` 中的 `PDFExtractor` 接口
3. 更新 `NewParserRegistry()` 以注册 PDF 解析器

## 性能建议

1. **批量处理**: 使用 `BatchParseFiles` 或 `BatchParseDirectory` 而不是循环调用
2. **并发处理**: 批量解析时可以在 goroutine 中并发处理
3. **文档分块**: 对于大文档，使用分块功能避免超出 LLM 上下文限制
4. **内存管理**: 处理大文件时注意内存使用，考虑流式处理

## 依赖

- Go 标准库
- PDF 库（可选，根据选择的实现）

## 许可

MIT License
