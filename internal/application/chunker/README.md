# Chunker - 文档分片服务

## 概述

Chunker 是一个基于 CloudWeGo Eino 的文档分片服务，支持语义分片和简单分片两种方式。

## 功能特性

### 1. 语义分片器 (SemanticChunker)

基于语义相似度智能分割文档，利用向量嵌入模型分析内容关联性。

**特点：**
- 使用向量嵌入模型分析语义
- 计算相邻片段的余弦相似度
- 自动在语义边界处分隔
- 合并过小的片段

**配置参数：**
| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `Embedding` | `embedding.Embedder` | 嵌入器（必需） | - |
| `BufferSize` | `int` | 上下文缓冲区大小 | 2 |
| `MinChunkSize` | `int` | 最小片段大小（字符） | 100 |
| `MaxChunkSize` | `int` | 最大片段大小（字符） | 2000 |
| `Separators` | `[]string` | 分隔符列表 | `["\n\n", "\n", "。", "！", "？", ".", "!"]` |
| `Percentile` | `float64` | 分割阈值（0-1） | 0.9 |
| `Overlap` | `int` | 重叠大小（字符） | 50 |

### 2. 简单分片器 (SimpleChunker)

基于固定大小和分隔符分片，无需嵌入模型。

**特点：**
- 按固定字符数分割
- 优先在分隔符处分割
- 支持片段重叠
- 轻量级，无需额外模型

**配置参数：**
| 参数 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `ChunkSize` | `int` | 每个片段的字符数 | 1000 |
| `Overlap` | `int` | 片段间重叠字符数 | 100 |
| `Separators` | `[]string` | 分隔符列表 | `["\n\n", "\n", "。", "！", "？"]` |
| `KeepSeparator` | `bool` | 是否保留分隔符 | true |

## 使用示例

### 语义分片

```go
import (
    "context"
    "link/internal/application/chunker"
    "link/internal/config"
    "link/internal/models/embedding"
)

func main() {
    ctx := context.Background()

    // 创建 Embedder
    embedConfig := config.LoadEmbeddingConfig()
    embedder, _ := embedding.NewEmbedder(embedConfig)

    // 创建语义分片器
    config := &chunker.Config{
        Embedding:    embedder,
        BufferSize:   2,
        MinChunkSize: 100,
        Separators:   []string{"\n\n", "\n", "。"},
        Percentile:   0.9,
    }

    chunker, _ := chunker.NewSemanticChunker(ctx, config)

    // 分片
    document := "长文档内容..."
    chunks, _ := chunker.Split(ctx, document)

    fmt.Printf("Split into %d chunks\n", len(chunks))
}
```

### 简单分片

```go
import (
    "context"
    "link/internal/application/chunker"
)

func main() {
    ctx := context.Background()

    // 创建简单分片器
    config := &chunker.SimpleConfig{
        ChunkSize:  500,
        Overlap:    50,
        Separators: []string{"\n\n", "\n", "。"},
    }

    chunker := chunker.NewSimpleChunker(config)

    // 分片
    document := "长文档内容..."
    chunks, _ := chunker.Split(ctx, document)

    fmt.Printf("Split into %d chunks\n", len(chunks))
}
```

### 与图谱提取结合使用

```go
import (
    "context"
    "link/internal/application/chunker"
    "link/internal/application/service"
)

func main() {
    ctx := context.Background()

    // 1. 创建分片器
    simpleConfig := chunker.DefaultSimpleConfig()
    chunker := chunker.NewSimpleChunker(simpleConfig)

    // 2. 分片文档
    document := loadDocument()
    chunks, _ := chunker.Split(ctx, document)

    // 3. 为每个分片创建图谱提取输入
    inputs := make([]*service.ChunkExtractionInput, 0, len(chunks))
    for i, chunk := range chunks {
        inputs = append(inputs, &service.ChunkExtractionInput{
            ChunkID:  fmt.Sprintf("chunk_%03d", i),
            Document: chunk,
            Query:    "提取实体和关系",
        })
    }

    // 4. 并发提取图谱（最多4个线程）
    graphService := service.NewGraphService(graphRepo, chatConfig)
    graphData, _ := graphService.ExtractGraphFromChunks(ctx, inputs)

    fmt.Printf("Extracted %d nodes, %d relations\n",
        len(graphData.Node), len(graphData.Relation))
}
```

## API 参考

### Chunker 接口

```go
type Chunker interface {
    // Split 将文档分割成多个片段
    Split(ctx context.Context, document string) ([]string, error)

    // SplitDocuments 将多个文档分割成片段
    SplitDocuments(ctx context.Context, docs []*schema.Document) ([]*schema.Document, error)

    // SplitWithIDs 将文档分割成片段，并返回带ID的文档
    SplitWithIDs(ctx context.Context, documentID, content string) ([]*schema.Document, error)
}
```

### 工厂函数

| 函数 | 说明 |
|------|------|
| `NewSemanticChunker(ctx, config)` | 创建语义分片器 |
| `NewSimpleChunker(config)` | 创建简单分片器 |
| `DefaultConfig()` | 返回默认语义配置 |
| `DefaultSimpleConfig()` | 返回默认简单配置 |

## 最佳实践

1. **选择合适的分片器**
   - 需要语义理解 → 语义分片器
   - 快速处理大文档 → 简单分片器

2. **调整参数**
   - `ChunkSize` 根据模型上下文窗口设置
   - `Overlap` 建议设置为 10-20% 的 `ChunkSize`
   - `Percentile` 越大分片越少，越小分片越多

3. **与图谱提取配合**
   - 分片大小建议 500-2000 字符
   - 重叠 100-200 字符确保上下文连续性
   - 使用并发提取提高效率

## 依赖

- `github.com/cloudwego/eino-ext/components/document/transformer/splitter/semantic`
- `github.com/cloudwego/eino/components/embedding`
- `github.com/cloudwego/eino/schema`
