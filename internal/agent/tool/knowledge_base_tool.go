package tool

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"link/internal/types"
	"link/internal/types/interfaces"
)

// ========================================
// 包级依赖管理
// ========================================

var (
	kbRepoInstance interfaces.KnowledgeBaseRepository
	kbRepoOnce     sync.Once
)

// InitKnowledgeBaseTool 初始化知识库工具的依赖
// 应该在应用启动时调用，传入实际的 repository 实现
func InitKnowledgeBaseTool(repo interfaces.KnowledgeBaseRepository) {
	kbRepoOnce.Do(func() {
		kbRepoInstance = repo
	})
}

// ========================================
// 请求/响应类型定义
// ========================================

// KbListRequestV2 知识库列表请求
type KbListRequestV2 struct {
	Status string `json:"status" jsonschema:"description=状态筛选：all(全部)/enabled(启用)/disabled(禁用)，默认all"`
}

// KbListResultV2 知识库列表结果
type KbListResultV2 struct {
	KnowledgeBases []KbInfoV2 `json:"knowledge_bases"`
	Count          int        `json:"count"`
	LatencyMs      int64      `json:"latency_ms"`
}

// KbInfoV2 知识库详细信息
type KbInfoV2 struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	DocumentCount int64  `json:"document_count"`
	ChunkCount    int64  `json:"chunk_count"`
	StorageSize   int64  `json:"storage_size"`
	IsPublic      bool   `json:"is_public"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// ========================================
// 工具创建函数
// ========================================

// NewKnowledgeBaseListTool 创建知识库列表工具
// 该工具可以查询当前租户下的所有知识库信息
func NewKnowledgeBaseListTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"kb_list",
		`获取当前租户下的所有知识库信息。

功能：
- 列出所有可用的知识库及其统计信息
- 获取知识库的配置详情
- 查看知识库中的文档和分块数量
- 支持按状态筛选

适用场景：
- 用户询问"有哪些知识库"
- 用户询问知识库统计信息
- 需要确定查询目标知识库时使用

参数：
- status: 状态筛选，可选值: all/enabled/disabled，默认 all`,
		listKnowledgeBases,
	)
}

// ========================================
// 工具执行逻辑
// ========================================

// listKnowledgeBases 查询知识库列表
func listKnowledgeBases(ctx context.Context, req *KbListRequestV2) (*KbListResultV2, error) {
	startTime := time.Now()

	// 1. 参数处理
	status := req.Status
	if status == "" {
		status = "all"
	}

	// 验证参数
	if status != "all" && status != "enabled" && status != "disabled" {
		return nil, fmt.Errorf("invalid status: %s, must be one of: all/enabled/disabled", status)
	}

	// 2. 检查依赖是否已初始化
	if kbRepoInstance == nil {
		return nil, fmt.Errorf("knowledge base repository not initialized, call InitKnowledgeBaseTool first")
	}

	// 3. 从上下文获取租户ID
	tenantID, err := getTenantIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取租户信息失败: %w", err)
	}

	// 4. 查询知识库列表
	knowledgeBases, err := queryKnowledgeBases(ctx, kbRepoInstance, tenantID, status)
	if err != nil {
		return nil, fmt.Errorf("查询知识库列表失败: %w", err)
	}

	// 5. 构建返回结果
	result := &KbListResultV2{
		KnowledgeBases: knowledgeBases,
		Count:          len(knowledgeBases),
		LatencyMs:      time.Since(startTime).Milliseconds(),
	}

	return result, nil
}

// queryKnowledgeBases 查询知识库列表（实际实现）
func queryKnowledgeBases(ctx context.Context, kbRepo interfaces.KnowledgeBaseRepository, tenantID int64, status string) ([]KbInfoV2, error) {
	// 设置分页参数（获取所有数据）
	page := 1
	pageSize := 1000

	// 查询知识库列表
	kbs, total, err := kbRepo.FindByTenantID(ctx, tenantID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("查询知识库失败: %w", err)
	}

	// 如果没有数据，返回空列表
	if total == 0 {
		return []KbInfoV2{}, nil
	}

	// 根据 status 筛选并转换为返回格式
	var results []KbInfoV2

	for _, kb := range kbs {
		// 跳过已删除的知识库
		if kb.DeletedAt != nil {
			continue
		}

		// 状态筛选
		kbStatus := "enabled"
		if kb.Status == 0 {
			kbStatus = "disabled"
		}

		if status != "all" && status != kbStatus {
			continue
		}

		// 转换为返回格式
		kbInfo := KbInfoV2{
			ID:            kb.ID,
			Name:          kb.Name,
			Description:   kb.Description,
			DocumentCount: int64(kb.DocumentCount),
			ChunkCount:    int64(kb.ChunkCount),
			StorageSize:   kb.StorageSize,
			IsPublic:      kb.IsPublic,
			Status:        kbStatus,
			CreatedAt:     kb.CreatedAt.Format(time.RFC3339),
			UpdatedAt:     kb.UpdatedAt.Format(time.RFC3339),
		}

		results = append(results, kbInfo)
	}

	return results, nil
}

// ========================================
// 上下文辅助函数
// ========================================

// getTenantIDFromContext 从上下文中获取租户ID
func getTenantIDFromContext(ctx context.Context) (int64, error) {
	// 尝试从 context.Value 中获取 tenant_id
	// 这是通过 middleware.ContextToRequest() 中间件设置的
	if tid, ok := ctx.Value("tenant_id").(int64); ok && tid > 0 {
		return tid, nil
	}

	// 如果没有找到租户ID，返回错误
	// 注意：在 Agent 工具调用场景下，需要在调用前确保上下文包含租户信息
	return 0, fmt.Errorf("tenant_id not found in context")
}

// ========================================
// 类型转换辅助函数
// ========================================

// typesToKbInfo 将 types.KnowledgeBase 转换为 KbInfoV2
func typesToKbInfo(kb *types.KnowledgeBase) KbInfoV2 {
	status := "enabled"
	if kb.Status == 0 {
		status = "disabled"
	}

	return KbInfoV2{
		ID:            kb.ID,
		Name:          kb.Name,
		Description:   kb.Description,
		DocumentCount: int64(kb.DocumentCount),
		ChunkCount:    int64(kb.ChunkCount),
		StorageSize:   kb.StorageSize,
		IsPublic:      kb.IsPublic,
		Status:        status,
		CreatedAt:     kb.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     kb.UpdatedAt.Format(time.RFC3339),
	}
}

// ========================================
// 工具工厂（支持依赖注入）
// ========================================

// KnowledgeBaseToolFactory 工具工厂
type KnowledgeBaseToolFactory struct {
	kbRepo interfaces.KnowledgeBaseRepository
}

// NewKnowledgeBaseToolFactory 创建工具工厂
func NewKnowledgeBaseToolFactory(kbRepo interfaces.KnowledgeBaseRepository) *KnowledgeBaseToolFactory {
	return &KnowledgeBaseToolFactory{
		kbRepo: kbRepo,
	}
}

// CreateToolUsingFactory 使用工厂创建工具
// 此方法会设置包级变量，然后返回使用 utils.InferTool 创建的工具
func (f *KnowledgeBaseToolFactory) CreateToolUsingFactory() (tool.InvokableTool, error) {
	InitKnowledgeBaseTool(f.kbRepo)
	return NewKnowledgeBaseListTool()
}
