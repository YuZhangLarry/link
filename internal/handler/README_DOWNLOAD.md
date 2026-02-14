# 文件下载功能说明

## 已实现的功能

### 文件结构
```
internal/handler/download.go      # 下载处理器
docs/download_api.md            # API 文档
D:\link\download\                # 本地下载目录
```

### 功能列表

| 功能 | 端点 | 方法 | 说明 |
|------|--------|------|------|
| 列出下载文件 | `/api/v1/download/list` | GET | 列出 download 目录下的所有文件 |
| 下载指定文件 | `/api/v1/download/file` | GET | 下载指定路径的文件 |
| 创建本地文件 | `/api/v1/download/local` | POST | 生成示例文件到 download 目录 |
| 删除下载文件 | `/api/v1/download/:filename` | DELETE | 删除指定的已下载文件 |
| 批量下载 | `/api/v1/download/batch` | POST | 批量下载（待实现 zip 打包） |
| 导出知识库 | `/api/v1/download/export` | POST | 导出知识库为文本文件 |

### 安全特性

1. **路径遍历防护**: 检测并拒绝包含 `..` 的文件名
2. **认证要求**: 所有接口都需要有效的 JWT Token 和 X-Tenant-ID
3. **文件类型检查**: 基于扩展名设置正确的 Content-Type

### Content-Type 映射

支持以下文件类型的正确 Content-Type:
- `.txt`, `.text` → `text/plain; charset=utf-8`
- `.md`, `.markdown` → `text/markdown; charset=utf-8`
- `.html`, `.htm` → `text/html; charset=utf-8`
- `.json` → `application/json; charset=utf-8`
- `.pdf` → `application/pdf`
- `.doc` → `application/msword`
- `.docx` → `application/vnd.openxmlformats-officedocument.wordprocessingml.document`
- `.xlsx` → `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
- `.zip` → `application/zip`

## 使用示例

### 1. 列出下载目录
```bash
curl "http://localhost:8080/api/v1/download/list" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Tenant-ID: 1"
```

### 2. 创建示例 Markdown 文件
```bash
curl "http://localhost:8080/api/v1/download/local?filename=example.md" \
  -X POST \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Tenant-ID: 1"
```

### 3. 删除已下载的文件
```bash
curl "http://localhost:8080/api/v1/download/example.md" \
  -X DELETE \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Tenant-ID: 1"
```

### 4. 导出知识库
```bash
curl "http://localhost:8080/api/v1/download/export?format=markdown" \
  -X POST \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "X-Tenant-ID: 1" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "我的知识库",
    "entities": [{"name": "实体1", "type": "Person"}],
    "relations": []
  }'
```

## 待扩展功能

- [ ] 远程下载（HTTP/FTP）
- [ ] ZIP 打包批量下载
- [ ] 下载历史记录
- [ ] 大文件下载进度
- [ ] 断点续传支持
- [ ] 细粒度权限控制
- [ ] 从知识库直接导出文件

## 注意事项

1. **下载目录固定**: 当前实现中，下载目录固定为 `D:\link\download`
2. **文件来源**: DownloadToLocal 当前生成示例文件，实际应从知识库等数据源生成
3. **批量下载**: 当前只返回确认信息，未实际实现 ZIP 打包
4. **认证**: 所有接口都需要 JWT Token 认证

## 代码结构

```go
// DownloadHandler 文件下载处理器
type DownloadHandler struct{}

// 主要方法
func (h *DownloadHandler) DownloadToLocal(c *gin.Context)
func (h *DownloadHandler) DownloadFile(c *gin.Context)
func (h *DownloadHandler) ListDownloadedFiles(c *gin.Context)
func (h *DownloadHandler) DeleteDownloadedFile(c *gin.Context)
func (h *DownloadHandler) BatchDownloadFiles(c *gin.Context)
func (h *DownloadHandler) ExportKnowledgeToText(c *gin.Context)

// 辅助方法
func (h *DownloadHandler) generateSampleContent(filename string) string
func (h *DownloadHandler) getContentType(filename string) string
```
