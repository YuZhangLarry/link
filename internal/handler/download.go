package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"link/internal/types"

	"github.com/gin-gonic/gin"
)

// DownloadHandler 文件下载处理器
type DownloadHandler struct {
	// 可扩展：添加下载历史记录等功能
}

// NewDownloadHandler 创建下载处理器
func NewDownloadHandler() *DownloadHandler {
	return &DownloadHandler{}
}

// DownloadToLocal 下载文件到本地
// QueryParam:
//   - filename: 文件名
func (h *DownloadHandler) DownloadToLocal(c *gin.Context) {
	// 获取文件名参数
	filename := c.Query("filename")
	if filename == "" {
		c.JSON(400, gin.H{"error": "filename parameter is required"})
		return
	}

	// 安全检查：防止路径遍历攻击
	if strings.Contains(filename, "..") || strings.Contains(filename, "\\") {
		c.JSON(400, gin.H{"error": "invalid filename"})
		return
	}

	// 设置下载目录
	downloadDir := "D:\\link\\download"

	// 确保下载目录存在
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to create download directory: %v", err)})
		return
	}

	// 构建文件路径
	// 这里简化实现：假设文件在某个可访问的位置
	// 实际使用时可以根据需要调整源文件位置
	//
	// 假设文件来自上传目录或知识库目录
	// 暂时返回一个提示，说明需要指定源文件路径
	//
	// 示例场景：
	// 1. 从已上传的文件下载（需要在 upload 目录查找）
	// 2. 从知识库导出文件下载

	// 简化实现：生成一个示例文件用于演示
	destPath := filepath.Join(downloadDir, filename)

	// 检查文件是否已存在
	if _, err := os.Stat(destPath); err == nil {
		// 文件已存在，返回文件信息
		c.JSON(200, gin.H{
			"message":  "file already exists",
			"path":     destPath,
			"filename": filename,
		})
		return
	}

	// 生成示例文件内容
	content := h.generateSampleContent(filename)
	if err := os.WriteFile(destPath, []byte(content), 0644); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to create file: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"message":  "file created successfully",
		"path":     destPath,
		"filename": filename,
		"size":     len(content),
	})
}

// DownloadFile 下载指定的文件
// Param:
//   - filepath: 文件相对路径
func (h *DownloadHandler) DownloadFile(c *gin.Context) {
	// 获取文件路径参数
	filePath := c.Query("filepath")
	if filePath == "" {
		c.JSON(400, gin.H{"error": "filepath parameter is required"})
		return
	}

	// 安全检查
	if strings.Contains(filePath, "..") {
		c.JSON(400, gin.H{"error": "invalid filepath"})
		return
	}

	// 源文件基础路径（可根据实际情况调整）
	// 这里假设从 knowledge 目录读取
	baseDir := "D:\\link"
	sourcePath := filepath.Join(baseDir, filePath)

	// 检查源文件是否存在
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		c.JSON(404, gin.H{"error": fmt.Sprintf("file not found: %v", err)})
		return
	}

	// 确保是文件而不是目录
	if sourceInfo.IsDir() {
		c.JSON(400, gin.H{"error": "path is a directory, not a file"})
		return
	}

	// 打开源文件
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to open source file: %v", err)})
		return
	}
	defer sourceFile.Close()

	// 获取文件信息用于响应头
	fileHeader := make([]byte, 512)
	_, err = sourceFile.Read(fileHeader)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to read file: %v", err)})
		return
	}

	// 确定内容类型
	contentType := h.getContentType(filepath.Base(sourcePath))

	// 设置响应头
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(sourcePath)))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Expires", "0")
	c.Header("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
	c.Header("Pragma", "public")

	// 发送文件
	c.File(sourcePath)
}

// ListDownloadedFiles 列出已下载的文件
func (h *DownloadHandler) ListDownloadedFiles(c *gin.Context) {
	downloadDir := "D:\\link\\download"

	// 读取目录
	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to read download directory: %v", err)})
		return
	}

	// 构建文件列表
	files := make([]gin.H, 0, len(entries))
	for _, entry := range entries {
		info, _ := entry.Info()
		files = append(files, gin.H{
			"name":    entry.Name(),
			"size":    info.Size(),
			"modTime": info.ModTime(),
			"isDir":   entry.IsDir(),
		})
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    files,
		"count":   len(files),
	})
}

// DeleteDownloadedFile 删除已下载的文件
// Param:
//   - filename: 要删除的文件名
func (h *DownloadHandler) DeleteDownloadedFile(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(400, gin.H{"error": "filename parameter is required"})
		return
	}

	// 安全检查
	if strings.Contains(filename, "..") || strings.Contains(filename, "\\") {
		c.JSON(400, gin.H{"error": "invalid filename"})
		return
	}

	downloadDir := "D:\\link\\download"
	filePath := filepath.Join(downloadDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(404, gin.H{"error": "file not found"})
		return
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to delete file: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"message":  "file deleted successfully",
		"filename": filename,
	})
}

// BatchDownloadFiles 批量下载多个文件（打包为 zip）
// Body: JSON {"files": ["file1.txt", "file2.txt"]}
func (h *DownloadHandler) BatchDownloadFiles(c *gin.Context) {
	var req struct {
		Files []string `json:"files" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 简化实现：返回文件列表
	// 实际实现可以使用 archive/zip 包创建 zip 文件
	c.JSON(200, gin.H{
		"message": "batch download not yet implemented",
		"files":   req.Files,
		"count":   len(req.Files),
		"note":    "will be implemented later with zip packaging",
	})
}

// ========================================
// 辅助方法
// ========================================

// generateSampleContent 生成示例文件内容
func (h *DownloadHandler) generateSampleContent(filename string) string {
	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	switch ext {
	case ".txt", ".text":
		return fmt.Sprintf("示例文本文件\n\n文件名: %s\n创建时间: %s\n",
			filename, "2024-02-12")

	case ".md":
		return fmt.Sprintf("# %s\n\n这是一个示例 Markdown 文件。\n\n## 功能说明\n- 支持 Markdown 语法\n- 可以预览和编辑\n\n",
			baseName)

	case ".json":
		return fmt.Sprintf(`{
  "filename": "%s",
  "type": "example",
  "description": "示例 JSON 文件",
  "created_at": "2024-02-12"
}`, filename)

	case ".html":
		return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>%s</title>
</head>
<body>
    <h1>示例 HTML 文件</h1>
    <p>这是一个示例页面。</p>
</body>
</html>`, baseName)

	default:
		return fmt.Sprintf("示例文件: %s\n内容类型: %s\n", filename, ext)
	}
}

// getContentType 根据文件扩展名获取 Content-Type
func (h *DownloadHandler) getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	contentTypes := map[string]string{
		".txt":      "text/plain; charset=utf-8",
		".text":     "text/plain; charset=utf-8",
		".md":       "text/markdown; charset=utf-8",
		".markdown": "text/markdown; charset=utf-8",
		".html":     "text/html; charset=utf-8",
		".htm":      "text/html; charset=utf-8",
		".json":     "application/json; charset=utf-8",
		".pdf":      "application/pdf",
		".doc":      "application/msword",
		".docx":     "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":      "application/vnd.ms-excel",
		".xlsx":     "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".zip":      "application/zip",
		".xml":      "text/xml; charset=utf-8",
	}

	if ct, ok := contentTypes[ext]; ok {
		return ct
	}

	return "application/octet-stream"
}

// ExportKnowledgeToText 导出知识库内容为文本文件
// QueryParam:
//   - format: 导出格式 (txt, json, markdown)
//
// Body: {"entities": [...], "relations": [...]}
func (h *DownloadHandler) ExportKnowledgeToText(c *gin.Context) {
	format := c.DefaultQuery("format", "txt")

	var req struct {
		Entities  []types.GraphNode     `json:"entities"`
		Relations []types.GraphRelation `json:"relations"`
		Title     string                `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if req.Title == "" {
		req.Title = "knowledge_export"
	}

	downloadDir := "D:\\link\\download"
	var filename string
	var content string

	switch format {
	case "json":
		filename = req.Title + ".json"
		// 简化实现，实际应该序列化为 JSON
		content = fmt.Sprintf(`{"title": "%s", "entities": %d, "relations": %d}`,
			req.Title, len(req.Entities), len(req.Relations))

	case "markdown", "md":
		filename = req.Title + ".md"
		content = fmt.Sprintf("# %s\n\n## Entities\n%d entities found\n\n## Relations\n%d relations found\n",
			req.Title, len(req.Entities), len(req.Relations))

	default: // txt
		filename = req.Title + ".txt"
		content = fmt.Sprintf("%s\n\nEntities: %d\nRelations: %d\n",
			req.Title, len(req.Entities), len(req.Relations))
	}

	destPath := filepath.Join(downloadDir, filename)
	if err := os.WriteFile(destPath, []byte(content), 0644); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to create export file: %v", err)})
		return
	}

	c.JSON(200, gin.H{
		"message":   "knowledge exported successfully",
		"filename":  filename,
		"path":      destPath,
		"format":    format,
		"entities":  len(req.Entities),
		"relations": len(req.Relations),
	})
}
