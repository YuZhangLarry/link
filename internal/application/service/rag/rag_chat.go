package rag

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/embedding"

	"link/internal/config"
	"link/internal/models/chat"
	"link/internal/types"
	"link/internal/types/interfaces"
)

// ChatServiceInterface 聊天服务接口（避免循环导入）
type ChatServiceInterface interface {
	Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error)
	ChatStream(ctx context.Context, req *types.ChatRequest) (<-chan types.StreamChatEvent, error)
}

// RAGChatService 集成 RAG 的聊天服务
type RAGChatService struct {
	chatService          ChatServiceInterface
	pipeline             *Pipeline
	enableRAG            bool
	retrievalSettingRepo interfaces.RetrievalSettingRepository
}

// NewRAGChatService 创建 RAG 聊天服务
func NewRAGChatService(
	chatConfig *config.ChatConfig,
	kbSettingRepo interfaces.KBSettingRepository,
	chunkRepo interfaces.ChunkRepository,
	embedder embedding.Embedder,
	milvusRetriever interface{},
	neo4jRepo interfaces.Neo4jGraphRepository,
	graphQueryRepo interfaces.GraphQueryRepository,
	retrievalSettingRepo interfaces.RetrievalSettingRepository,
	chatService ChatServiceInterface,
) (*RAGChatService, error) {
	// 转换配置类型
	chatConfigForPipeline := &chat.ChatConfig{
		Source:    chatConfig.Source,
		BaseURL:   chatConfig.BaseURL,
		ModelName: chatConfig.ModelName,
		APIKey:    chatConfig.APIKey,
		Provider:  chatConfig.Provider,
		ModelID:   fmt.Sprintf("rag_%d", time.Now().UnixNano()),
	}

	// 创建 RAG Pipeline
	pipeline, err := NewPipeline(
		chatConfigForPipeline,
		kbSettingRepo,
		chunkRepo,
		embedder,
		milvusRetriever,
		neo4jRepo,
		graphQueryRepo,
	)
	if err != nil {
		return nil, fmt.Errorf("创建 RAG Pipeline 失败: %w", err)
	}

	return &RAGChatService{
		chatService:          chatService,
		pipeline:             pipeline,
		enableRAG:            true,
		retrievalSettingRepo: retrievalSettingRepo,
	}, nil
}

// NewRAGChatServiceWithReranker 创建带重排模型的 RAG 聊天服务
func NewRAGChatServiceWithReranker(
	chatConfig *config.ChatConfig,
	kbSettingRepo interfaces.KBSettingRepository,
	chunkRepo interfaces.ChunkRepository,
	embedder embedding.Embedder,
	milvusRetriever interface{},
	neo4jRepo interfaces.Neo4jGraphRepository,
	graphQueryRepo interfaces.GraphQueryRepository,
	rerankEmbedder RerankEmbedder,
	chatService ChatServiceInterface,
) (*RAGChatService, error) {
	// 转换配置类型
	chatConfigForPipeline := &chat.ChatConfig{
		Source:    chatConfig.Source,
		BaseURL:   chatConfig.BaseURL,
		ModelName: chatConfig.ModelName,
		APIKey:    chatConfig.APIKey,
		Provider:  chatConfig.Provider,
		ModelID:   fmt.Sprintf("rag_%d", time.Now().UnixNano()),
	}

	// 创建带重排的 RAG Pipeline
	pipeline, err := NewPipelineWithReranker(
		chatConfigForPipeline,
		kbSettingRepo,
		chunkRepo,
		embedder,
		milvusRetriever,
		neo4jRepo,
		graphQueryRepo,
		rerankEmbedder,
	)
	if err != nil {
		return nil, fmt.Errorf("创建 RAG Pipeline 失败: %w", err)
	}

	return &RAGChatService{
		chatService: chatService,
		pipeline:    pipeline,
		enableRAG:   true,
	}, nil
}

// SetAgent 设置 Agent
func (s *RAGChatService) SetAgent(agent interface{}) {
	// 类型断言设置 Agent（需要外部实现）
	_ = agent
	log.Printf("[RAGChatService] SetAgent called")
}

// Chat 聊天（支持 RAG）
func (s *RAGChatService) Chat(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	// 获取租户 ID
	tenantID := getTenantID(ctx)

	// 执行 RAG 检索（如果启用）
	var ragContext *types.RAGContext
	if s.enableRAG && req.RAGConfig != nil && req.RAGConfig.Enabled {
		var err error
		ragContext, err = s.executeRAG(ctx, tenantID, req)
		if err != nil {
			log.Printf("[RAGChatService] RAG 检索失败: %v", err)
			// 失败时继续使用原始查询，不影响聊天
		}
	}

	// 构建增强后的消息内容
	content := s.buildContentWithContext(req.Content, ragContext)

	// 创建新的请求（使用增强后的内容）
	enhancedReq := *req
	enhancedReq.Content = content

	// 调用基础聊天服务
	resp, err := s.chatService.Chat(ctx, &enhancedReq)
	if err != nil {
		return nil, err
	}

	// 添加 RAG 上下文到响应
	resp.RAGContext = ragContext

	return resp, nil
}

// ChatStream 流式聊天（支持 RAG）
func (s *RAGChatService) ChatStream(ctx context.Context, req *types.ChatRequest) (<-chan types.StreamChatEvent, error) {
	// 获取租户 ID
	tenantID := getTenantID(ctx)

	// 执行 RAG 检索（如果启用）
	var ragContext *types.RAGContext
	if s.enableRAG && req.RAGConfig != nil && req.RAGConfig.Enabled {
		var err error
		ragContext, err = s.executeRAG(ctx, tenantID, req)
		if err != nil {
			log.Printf("[RAGChatService] RAG 检索失败: %v", err)
		}
	}

	// 构建增强后的消息内容
	content := s.buildContentWithContext(req.Content, ragContext)

	// 创建新的请求（使用增强后的内容）
	enhancedReq := *req
	enhancedReq.Content = content

	// 调用基础聊天服务的流式方法
	eventChan, err := s.chatService.ChatStream(ctx, &enhancedReq)
	if err != nil {
		return nil, err
	}

	// 包装事件流，添加 RAG 上下文
	resultChan := make(chan types.StreamChatEvent, 10)
	go func() {
		defer close(resultChan)
		firstEvent := true
		for event := range eventChan {
			// 在第一个事件中添加 RAG 上下文
			if firstEvent && ragContext != nil {
				event.RAGContext = ragContext
				firstEvent = false
			}
			resultChan <- event
		}
	}()

	return resultChan, nil
}

// ========================================
// RAG 执行方法
// ========================================

// executeRAG 执行 RAG 检索
func (s *RAGChatService) executeRAG(ctx context.Context, tenantID int64, req *types.ChatRequest) (*types.RAGContext, error) {
	if req.RAGConfig == nil || !req.RAGConfig.Enabled {
		return nil, nil
	}

	if req.RAGConfig.KBID == "" {
		return nil, fmt.Errorf("RAG 启用时必须指定知识库 ID (kb_id)")
	}

	log.Printf("[RAGChatService] 执行 RAG 检索: kbID=%s, query=%s", req.RAGConfig.KBID, req.Content)

	// 转换 RAGConfig 到 PipelineConfig
	pipelineConfig := s.convertToPipelineConfig(req.RAGConfig)

	// 获取对话历史（用于查询增强）
	conversationHistory := s.buildConversationHistory(req.History)

	// 执行 Pipeline
	result, err := s.pipeline.Execute(
		ctx,
		tenantID,
		req.RAGConfig.KBID,
		req.Content,
		conversationHistory,
		pipelineConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("Pipeline 执行失败: %w", err)
	}

	// 转换结果为 RAGContext
	ragContext := &types.RAGContext{
		Query:             result.Query,
		FinalQuery:        result.FinalQuery,
		Contexts:          result.GetContexts(),
		ContextsWithScore: result.GetContextsWithScore(),
		SourceTypes:       result.SourceTypes,
		RetrievedCount:    len(result.RetrievedDocs),
		Stages:            s.convertStages(result.Stages),
	}

	log.Printf("[RAGChatService] RAG 检索完成: 检索到 %d 个文档, 来源=%v",
		ragContext.RetrievedCount, ragContext.SourceTypes)

	return ragContext, nil
}

// buildContentWithContext 构建带上下文的消息内容
func (s *RAGChatService) buildContentWithContext(content string, ragContext *types.RAGContext) string {
	if ragContext == nil || len(ragContext.Contexts) == 0 {
		return content
	}

	// 构建上下文提示
	var contextBuilder strings.Builder
	contextBuilder.WriteString("以下是从知识库中检索到的相关内容，请基于这些信息回答用户问题：\n\n")

	for i, ctx := range ragContext.Contexts {
		contextBuilder.WriteString(fmt.Sprintf("[文档%d] %s\n", i+1, ctx))
	}

	contextBuilder.WriteString("\n用户问题：\n")
	contextBuilder.WriteString(content)

	return contextBuilder.String()
}

// buildConversationHistory 构建对话历史字符串
func (s *RAGChatService) buildConversationHistory(history []types.Message) string {
	if len(history) == 0 {
		return ""
	}

	var sb strings.Builder
	for _, msg := range history {
		sb.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
	}
	return sb.String()
}

// convertToPipelineConfig 转换 RAGConfig 到 PipelineConfig
func (s *RAGChatService) convertToPipelineConfig(ragConfig *types.RAGConfig) *PipelineConfig {
	// 确保 retrieval_modes 包含 vector
	modes := ragConfig.RetrievalModes
	hasVector := false
	for _, mode := range modes {
		if mode == "vector" {
			hasVector = true
			break
		}
	}
	if !hasVector && len(modes) == 0 {
		modes = []string{"vector"}
	} else if !hasVector {
		modes = append([]string{"vector"}, modes...)
	}

	return &PipelineConfig{
		RetrievalModes:      modes,
		VectorTopK:          ragConfig.VectorTopK,
		KeywordTopK:         ragConfig.KeywordTopK,
		GraphTopK:           ragConfig.GraphTopK,
		SimilarityThreshold: ragConfig.SimilarityThreshold,
		Alpha:               ragConfig.Alpha,
	}
}

// convertStages 转换阶段结果
func (s *RAGChatService) convertStages(stages map[string]*StageResult) map[string]interface{} {
	if stages == nil {
		return nil
	}

	result := make(map[string]interface{})
	for name, stage := range stages {
		result[name] = map[string]interface{}{
			"name":         stage.Name,
			"input":        stage.Input,
			"output":       stage.Output,
			"success":      stage.Success,
			"error":        stage.Error,
			"input_count":  stage.InputCount,
			"output_count": stage.OutputCount,
		}
	}
	return result
}

// ========================================
// 委托方法
// ========================================

// SetPipeline 设置 Pipeline
func (s *RAGChatService) SetPipeline(pipeline *Pipeline) {
	s.pipeline = pipeline
}

// EnableRAG 启用/禁用 RAG
func (s *RAGChatService) EnableRAG(enable bool) {
	s.enableRAG = enable
}

// GetPipeline 获取 Pipeline
func (s *RAGChatService) GetPipeline() *Pipeline {
	return s.pipeline
}

// ========================================
// 辅助方法
// ========================================

// getTenantID 从 context 获取租户 ID
func getTenantID(ctx context.Context) int64 {
	if tenantID, ok := ctx.Value("tenant_id").(int64); ok {
		return tenantID
	}
	// 从 context.Value 获取可能返回 float64 (JSON 数字)
	if v := ctx.Value("tenant_id"); v != nil {
		switch tid := v.(type) {
		case float64:
			return int64(tid)
		case int:
			return int64(tid)
		case int64:
			return tid
		}
	}
	return 1 // 默认租户 ID
}

// ========================================
// 向后兼容的聊天方法（不使用 RAG）
// ========================================

// ChatWithoutRAG 不使用 RAG 的聊天
func (s *RAGChatService) ChatWithoutRAG(ctx context.Context, req *types.ChatRequest) (*types.ChatResponse, error) {
	return s.chatService.Chat(ctx, req)
}

// StreamWithoutRAG 不使用 RAG 的流式聊天
func (s *RAGChatService) StreamWithoutRAG(ctx context.Context, req *types.ChatRequest) (<-chan types.StreamChatEvent, error) {
	return s.chatService.ChatStream(ctx, req)
}

// CreateChatInstance 创建聊天实例（向后兼容）
func (s *RAGChatService) CreateChatInstance() (chat.Chat, error) {
	// 通过反射或直接访问创建
	// 这里简化处理，实际可能需要调整
	return chat.NewChat(&chat.ChatConfig{})
}

// ========================================
// Session RAG 配置管理
// ========================================

// SaveRAGConfigToSession 保存 RAG 配置到 retrieval_settings 表
func (s *RAGChatService) SaveRAGConfigToSession(ctx context.Context, sessionID string, ragConfig *types.RAGConfig, tenantID int64) error {
	if ragConfig == nil {
		return nil
	}

	// 使用 retrieval_settings 仓储保存配置
	return s.retrievalSettingRepo.UpsertBySessionID(ctx, sessionID, tenantID, ragConfig)
}

// GetRAGConfigFromSession 从 retrieval_settings 表获取 RAG 配置
func (s *RAGChatService) GetRAGConfigFromSession(ctx context.Context, sessionID string) (*types.RAGConfig, error) {
	// 从 retrieval_settings 表查询
	setting, err := s.retrievalSettingRepo.FindBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("查询检索设置失败: %w", err)
	}

	// 辅助函数：安全获取指针值
	getInt := func(ptr *int, defaultVal int) int {
		if ptr != nil {
			return *ptr
		}
		return defaultVal
	}
	getFloat64 := func(ptr *float64, defaultVal float64) float64 {
		if ptr != nil {
			return *ptr
		}
		return defaultVal
	}
	getNumber := func(ptr *types.Number, defaultVal float64) float64 {
		if ptr != nil {
			return float64(*ptr)
		}
		return defaultVal
	}

	// 确定检索模式
	var retrievalModes []string
	// 默认使用向量检索
	retrievalModes = []string{"vector"}

	// 将 RetrievalSetting 转换为 RAGConfig
	ragConfig := &types.RAGConfig{
		Enabled:             true, // 如果有设置就认为启用
		RetrievalModes:      retrievalModes,
		VectorTopK:          getInt(setting.VectorTopK, 15),
		KeywordTopK:         getInt(setting.BM25TopK, 15),
		GraphTopK:           getInt(setting.GraphTopK, 10),
		SimilarityThreshold: getFloat64(setting.VectorThreshold, 0.0),
		Alpha:               float32(getNumber(setting.HybridAlpha, 0.6)),
	}

	return ragConfig, nil
}
