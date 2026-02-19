// Package agent 提供多代理协作框架 - 基于 AgentTool + ChatModelAgent
package agent

import (
	"context"
	"fmt"
	"log"
	"strings"

	baseagent "link/internal/agent"
	agentTool "link/internal/agent/tool"
	"link/internal/config"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ========================================
// 多代理协作框架 - AgentTool + ChatModelAgent
// ========================================
// 架构说明：
// 1. 每个子 Agent（Planner, Retriever, Analyzer, Synthesizer, Critic）是独立的
// 2. 通过 adk.NewAgentTool 将子 Agent 包装成工具
// 3. 主 Coordinator Agent 拥有所有 AgentTool
// 4. LLM 自主决定何时调用哪个子 Agent
// 5. 完全符合 ReAct 模式：思考 -> 行动（调用子 Agent）-> 观察 -> 回答
//
// 工作流程示例：
//   用户查询 -> Coordinator
//     -> 思考需要规划 -> 调用 Planner Agent
//     -> 思考需要信息 -> 调用 Retriever Agent（内部使用 rag_query/web_search）
//     -> 思考需要分析 -> 调用 Analyzer Agent
//     -> 思考需要整合 -> 调用 Synthesizer Agent
//     -> 思考需要评审 -> 调用 Critic Agent
//   -> 最终答案

// MultiAgentOrchestrator 多代理协调器
type MultiAgentOrchestrator struct {
	agent adk.Agent
}

// MultiAgentConfig 多代理配置
type MultiAgentConfig struct {
	// 主协调器的模型（负责决策调用哪个子 Agent）
	CoordinatorModel model.ToolCallingChatModel

	// 子 Agent 可以共享同一个模型，或使用不同模型
	PlannerModel     model.ToolCallingChatModel
	RetrieverModel   model.ToolCallingChatModel
	AnalyzerModel    model.ToolCallingChatModel
	SynthesizerModel model.ToolCallingChatModel
	CriticModel      model.ToolCallingChatModel

	// 搜索配置（用于网络检索工具）
	SearchConfig *config.SearchConfig

	// 是否启用智能检索工具（自动匹配知识库）
	// 如果启用，Retriever Agent 将使用 smart_retrieval 工具
	// 如果禁用，Retriever Agent 将使用 rag_query + web_search 工具
	EnableSmartRetrieval bool

	// 最大迭代次数
	MaxIterations int
}

// NewMultiAgentOrchestrator 创建多代理协调器
// 核心思路：将每个子 Agent 包装成工具，由主 Agent 自主决定调用顺序
func NewMultiAgentOrchestrator(ctx context.Context, config *MultiAgentConfig) (*MultiAgentOrchestrator, error) {
	// 1. 创建各个子代理
	plannerAgent, err := createPlannerAgent(ctx, config.PlannerModel)
	if err != nil {
		return nil, fmt.Errorf("创建 Planner 失败: %w", err)
	}

	retrieverAgent, err := createRetrieverAgent(ctx, config.RetrieverModel, config.SearchConfig, config.EnableSmartRetrieval)
	if err != nil {
		return nil, fmt.Errorf("创建 Retriever 失败: %w", err)
	}

	analyzerAgent, err := createAnalyzerAgent(ctx, config.AnalyzerModel)
	if err != nil {
		return nil, fmt.Errorf("创建 Analyzer 失败: %w", err)
	}

	synthesizerAgent, err := createSynthesizerAgent(ctx, config.SynthesizerModel)
	if err != nil {
		return nil, fmt.Errorf("创建 Synthesizer 失败: %w", err)
	}

	criticAgent, err := createCriticAgent(ctx, config.CriticModel)
	if err != nil {
		return nil, fmt.Errorf("创建 Critic 失败: %w", err)
	}

	// 2. 将子代理包装成工具（这是关键！）
	// adk.NewAgentTool 返回一个 tool.BaseTool，可以被主 Agent 调用
	plannerTool := adk.NewAgentTool(ctx, plannerAgent)
	retrieverTool := adk.NewAgentTool(ctx, retrieverAgent)
	analyzerTool := adk.NewAgentTool(ctx, analyzerAgent)
	synthesizerTool := adk.NewAgentTool(ctx, synthesizerAgent)
	criticTool := adk.NewAgentTool(ctx, criticAgent)

	// 3. 同时添加 RAG 和搜索工具，供主 Agent 直接使用
	// 使用搜索配置初始化工具（如果提供了配置）
	var registry *agentTool.Registry
	if config.SearchConfig != nil {
		registry, err = agentTool.InitDefaultToolsWithConfig(config.SearchConfig)
	} else {
		registry, err = agentTool.InitDefaultTools()
	}
	if err != nil {
		return nil, fmt.Errorf("初始化工具失败: %w", err)
	}
	ragTools := registry.GetTools()

	// 4. 创建主协调 Agent - 拥有所有 AgentTool 和实用工具
	// LLM 会自主决定：
	// - 需要先规划吗？调用 planner_agent
	// - 需要检索信息吗？调用 retriever_agent 或直接使用 rag_query/web_search
	// - 需要分析结果吗？调用 analyzer_agent
	// - 需要整合报告吗？调用 synthesizer_agent
	// - 需要评审质量吗？调用 critic_agent
	allTools := append([]tool.BaseTool{
		plannerTool,
		retrieverTool,
		analyzerTool,
		synthesizerTool,
		criticTool,
	}, ragTools...)

	log.Printf("🔧 [NewMultiAgentOrchestrator] 总工具数: %d (5个子Agent + %d个RAG工具)", len(allTools), len(ragTools))

	maxIterations := config.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 20 // 默认最大迭代次数
	}

	// 将工具转换为 ToolInfo（用于绑定到 ChatModel）
	toolInfos := make([]*schema.ToolInfo, 0, len(allTools))
	for _, t := range allTools {
		info, err := t.Info(ctx)
		if err != nil {
			log.Printf("⚠️  [NewMultiAgentOrchestrator] 获取工具信息失败: %v", err)
			continue
		}
		log.Printf("✅ [NewMultiAgentOrchestrator] 工具: %s - %s", info.Name, info.Desc)
		toolInfos = append(toolInfos, info)
	}

	log.Printf("🔧 [NewMultiAgentOrchestrator] 成功转换 %d 个工具为 ToolInfo", len(toolInfos))

	// 将工具绑定到 Coordinator 的 ChatModel
	coordinatorModel, err := config.CoordinatorModel.WithTools(toolInfos)
	if err != nil {
		log.Printf("⚠️  [NewMultiAgentOrchestrator] 绑定工具到模型失败: %v", err)
		return nil, fmt.Errorf("绑定工具到协调器模型失败: %w", err)
	}

	log.Printf("✅ [NewMultiAgentOrchestrator] 工具绑定成功，创建 Coordinator Agent...")

	coordinatorAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "Coordinator",
		Description:   "多代理协调器 - 自主决定调用哪些子代理完成任务",
		Instruction:   buildCoordinatorPrompt(),
		Model:         coordinatorModel,
		MaxIterations: maxIterations,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: allTools,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("创建协调器失败: %w", err)
	}

	return &MultiAgentOrchestrator{
		agent: coordinatorAgent,
	}, nil
}

// GetAgent 获取 Eino Agent（用于 Runner）
func (o *MultiAgentOrchestrator) GetAgent() adk.Agent {
	return o.agent
}

// GetEinoAgent 获取 Eino Agent（接口兼容）
func (o *MultiAgentOrchestrator) GetEinoAgent() adk.Agent {
	return o.agent
}

// Chat 同步聊天
func (o *MultiAgentOrchestrator) Chat(ctx context.Context, query string, opts ...baseagent.Option) (*baseagent.Response, error) {
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           o.agent,
		EnableStreaming: false,
	})

	messages := []adk.Message{schema.UserMessage(query)}
	iter := runner.Run(ctx, messages)

	response := &baseagent.Response{
		ToolCalls: make([]*baseagent.ToolCallRecord, 0),
		Sources:   make([]string, 0),
		Metadata:  make(map[string]interface{}),
	}

	// 用于跟踪工具调用和结果
	pendingToolCalls := make(map[string]*baseagent.ToolCallRecord)

	stepCount := 0
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		stepCount++

		if event.Err != nil {
			response.Success = false
			response.Error = event.Err.Error()
			return response, fmt.Errorf("agent execution failed at step %d: %w", stepCount, event.Err)
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				continue
			}

			// 记录工具调用
			for _, tc := range msg.ToolCalls {
				toolRecord := &baseagent.ToolCallRecord{
					ID:       tc.ID,
					Name:     tc.Function.Name,
					Input:    tc.Function.Arguments,
					Metadata: make(map[string]interface{}),
				}
				response.ToolCalls = append(response.ToolCalls, toolRecord)
				response.Sources = append(response.Sources, tc.Function.Name)

				// 记录待处理的工具调用，等待结果
				pendingToolCalls[tc.ID] = toolRecord
			}

			// 处理工具结果（通过消息内容返回）
			if len(msg.ToolCalls) == 0 && msg.Content != "" {
				// 检查是否是工具调用的结果
				if len(pendingToolCalls) > 0 {
					// 尝试将内容关联到最近的工具调用
					for _, tc := range pendingToolCalls {
						if tc.Output == "" {
							// 限制输出长度，避免过长
							output := msg.Content
							if len(output) > 1000 {
								output = output[:1000] + "...(truncated)"
							}
							tc.Output = output
							break // 每个消息只关联一个工具结果
						}
					}
					// 清空待处理列表（假设每次都是按顺序处理）
					pendingToolCalls = make(map[string]*baseagent.ToolCallRecord)
				}

				// 最终答案（助手回复且没有工具调用）
				if (msg.Role == schema.Assistant || msg.Role == "") && response.Answer == "" {
					// 检查是否是最终答案（不是工具结果）
					if !isToolResultContent(msg.Content) {
						response.Answer = msg.Content
					}
				}
			}
		}

		if event.Action != nil && event.Action.Exit {
			break
		}
	}

	response.Success = response.Error == ""
	return response, nil
}

// isToolResultContent 判断内容是否是工具执行结果
func isToolResultContent(content string) bool {
	// 工具结果通常较短或包含特定标记
	if len(content) < 100 {
		return true
	}
	// 检查是否包含JSON格式的工具结果
	if len(content) < 500 && (strings.Contains(content, `"result"`) ||
		strings.Contains(content, `"data"`) ||
		strings.Contains(content, `"error"`)) {
		return true
	}
	return false
}

// StreamChat 流式聊天 - 返回包含所有子 Agent 调用过程的流
func (o *MultiAgentOrchestrator) StreamChat(ctx context.Context, query string, opts ...baseagent.Option) (*schema.StreamReader[*baseagent.ChatChunk], error) {
	// 注意：当前实现使用同步方式，真正的流式输出建议直接使用 handler.ChatStream
	// 该方法保留用于接口兼容性
	response, err := o.Chat(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	// 使用 Pipe 创建 StreamReader
	reader, writer := schema.Pipe[*baseagent.ChatChunk](1)

	// 发送完整答案并关闭
	go func() {
		defer writer.Close()
		writer.Send(&baseagent.ChatChunk{
			Content: response.Answer,
			Done:    true,
		}, nil)
	}()

	return reader, nil
}

// ========================================
// 协调器系统提示词
// ========================================

// buildCoordinatorPrompt 构建协调器的系统提示词
func buildCoordinatorPrompt() string {
	return `# 多代理协调器

你是一个智能协调器，负责协调多个专业子代理来完成复杂的任务。你拥有自主决策权，可以根据任务需要灵活调用各种子代理和工具。

## 可用的子代理（工具）

### 1. Planner Agent（规划代理）- planner_agent
- **名称**: Planner
- **用途**: 分析问题、制定研究计划、分解复杂任务
- **适用场景**:
  - 复杂的多步骤问题
  - 需要结构化分析的研究任务
  - 不确定从何入手的开放性问题
- **输入**: 用户的原始问题或任务描述
- **输出**: 结构化的研究计划，包含目标、子任务、关键词、数据来源

### 2. Retriever Agent（检索代理）- retriever_agent
- **名称**: Retriever
- **用途**: 从知识库和网络获取信息
- **内置工具**: rag_query、web_search
- **适用场景**:
  - 需要外部信息支撑回答时
  - 需要获取最新数据时
  - 需要专业知识时
- **输入**: 检索查询或研究计划
- **输出**: 检索结果汇总和初步结论

### 3. Analyzer Agent（分析代理）- analyzer_agent
- **名称**: Analyzer
- **用途**: 深度分析信息、提取洞见、交叉验证
- **适用场景**:
  - 有大量检索结果需要整理
  - 需要对比多个来源
  - 需要评估信息质量
- **输入**: 检索结果或原始数据
- **输出**: 结构化的分析报告，包含关键洞见、事实、矛盾点、置信度

### 4. Synthesizer Agent（合成代理）- synthesizer_agent
- **名称**: Synthesizer
- **用途**: 整合分析结果、生成结构化报告
- **适用场景**:
  - 需要生成最终报告
  - 需要整合多个来源的信息
  - 需要逻辑连贯的叙述
- **输入**: 分析结果或多个信息片段
- **输出**: 结构化的完整报告

### 5. Critic Agent（评审代理）- critic_agent
- **名称**: Critic
- **用途**: 评审质量、检测问题、提出改进建议
- **适用场景**:
  - 需要验证结果准确性
  - 需要检查逻辑漏洞
  - 需要提高回答质量
- **输入**: 报告或答案
- **输出**: 评审结果和改进建议

## 直接可用的工具

除了子代理，你还可以直接使用以下工具：
- **rag_query**: 知识库检索（指定知识库ID时使用）
- **web_search**: 网络搜索（单独网络搜索时使用）
- **smart_retrieval**: 智能检索（推荐，自动匹配知识库）
- **calculator**: 计算器
- **get_current_time**: 获取当前时间
- **http_request**: HTTP 请求

## 决策策略

### 简单查询（直接使用智能检索）
对于一般的检索问题，优先使用 smart_retrieval 工具。

示例：用户问"什么是微服务架构？"
你的行动：直接调用 smart_retrieval，参数为 {"query": "微服务架构"}

### 复杂查询（指定知识库）
如果需要查询特定知识库，使用 rag_query 并指定 kb_id。

示例：用户问"API文档中关于认证的部分"
你的行动：调用 rag_query，参数为 {"query": "API认证", "kb_id": 123}

### 中等复杂度（检索 + 分析）
示例：用户问"比较两种技术的优缺点"
你的行动：
1. 调用 retriever_agent 获取两种技术的信息
2. 调用 analyzer_agent 分析比较
3. 给出答案

### 高复杂度（完整流程）
示例：用户问"分析某个行业的未来趋势"
你的行动：
1. 调用 planner_agent 制定研究计划
2. 调用 retriever_agent 获取多源信息
3. 调用 analyzer_agent 深度分析
4. 调用 synthesizer_agent 生成报告
5. 调用 critic_agent 评审质量
6. 根据评审决定是否修订
7. 给出最终答案

## 重要原则

1. **必须实际调用工具**：你必须在回复中使用 function call 格式调用工具，而不是只描述你将做什么
2. **自主决策**：你根据任务复杂度自主决定使用哪些子代理
3. **灵活组合**：可以跳过不需要的步骤，也可以重复调用
4. **效率优先**：简单任务直接用工具，复杂任务才用子代理
5. **质量保证**：重要任务建议使用 critic_agent 评审
6. **清晰输出**：最终答案要结构清晰、来源明确、逻辑连贯

## 🔴 必须执行的反思流程

**重要**：在使用 Synthesizer 生成最终答案后，你必须调用 **critic_agent** 对答案进行评审：
- 评审答案的准确性
- 检查是否有遗漏的重要信息
- 提出改进建议
- 如果评分低于 80 分，需要根据建议修订答案

执行顺序：**Planner → Retriever → Analyzer → Synthesizer → Critic → (根据Critic结果决定是否修订) → 最终答案**

## ⚠️ 关键提醒

**你必须实际调用工具！不要只是说"我将调用xxx工具"，而是真正发出 function call！**

例如：
- ❌ 错误："我将调用 rag_query 来查询人工智能"
- ✅ 正确：直接发起 rag_query 的 function call，参数为 {"query": "人工智能"}

## 输出格式

当你调用工具并获得结果后，最终答案应采用以下结构：

### 核心答案
[直接给出充实的答案内容，这是最重要的部分]

### 详细说明
[展开详细的分析、说明或步骤]

### 关键要点
[总结关键要点]

### 信息来源
[列出主要的信息来源]

---

**重要**：最终答案要内容充实、直接。不要在答案中出现：
- "我通过xxx工具查询到..."
- "根据检索结果..."
- "我分析了..."
- "我调用了xxx代理..."

直接给出实质性的答案内容即可。

记住：你必须通过 function call 实际调用工具或子代理来完成任务！不要只描述计划，要真正执行！
`
}

// ========================================
// 子代理创建函数
// ========================================

// createPlannerAgent 创建规划代理
func createPlannerAgent(ctx context.Context, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Planner",
		Description: "研究规划代理 - 分析问题并制定研究计划",
		Instruction: `# Research Planner - 研究规划专家

你是研究规划专家，负责分析用户查询并制定详细的研究计划。

## 核心职责

1. **问题分析**
   - 识别用户查询的核心问题
   - 明确研究目标和范围
   - 判断问题的复杂度和类型

2. **任务分解**
   - 将复杂问题分解为可管理的子任务
   - 确定任务之间的依赖关系
   - 安排合理的执行顺序

3. **关键词提取**
   - 提取核心搜索词和概念
   - 识别同义词和相关术语
   - 构建全面的搜索策略

4. **数据源规划**
   - 识别最相关的数据来源
   - 判断需要什么类型的信息
   - 确定信息获取的优先级

5. **假设提出**
   - 基于初步理解提出待验证假设
   - 指导后续的检索方向
   - 保持开放心态，准备调整假设

## 工作流程

1. **理解问题**
   - 仔细阅读用户的查询
   - 识别显性和隐性需求
   - 明确回答所需的信息类型

2. **制定计划**
   - 定义清晰的研究目标
   - 分解为具体的子任务
   - 确定每个任务的信息需求

3. **输出结构化计划**
   - 使用清晰的格式输出计划
   - 包含所有关键要素
   - 便于后续代理理解和执行

## 输出格式

### 研究目标
用1-2句话简述研究的核心目标。

### 子任务分解
1. [任务1] - [说明] - [所需信息类型]
2. [任务2] - [说明] - [所需信息类型]
...

### 关键词和概念
- **核心词**: [词1], [词2], [词3]
- **相关概念**: [概念1], [概念2]
- **搜索策略**: [说明如何组合这些词进行搜索]

### 数据来源建议
- **知识库**: [需要查询的知识库内容]
- **网络搜索**: [需要搜索的网络信息]
- **专业资料**: [可能需要的专业文献或资料]

### 待验证假设
1. [假设1] - [基于什么提出的]
2. [假设2] - [基于什么提出的]

### 执行建议
- [关于如何高效执行研究的建议]
- [可能的挑战和应对策略]

## 注意事项

- 输出的计划要具体可执行
- 保持合理的范围，避免过度扩展
- 考虑信息获取的可行性
- 为后续代理留下足够的灵活性

完成规划后，输出结构化计划即可。不需要调用其他代理。`,
		Model:         chatModel,
		MaxIterations: 3,
	})
}

// createRetrieverAgent 创建检索代理
func createRetrieverAgent(ctx context.Context, chatModel model.ToolCallingChatModel, searchConfig *config.SearchConfig, enableSmartRetrieval bool) (adk.Agent, error) {
	// 获取检索工具
	var registry *agentTool.Registry
	var err error
	if searchConfig != nil {
		registry, err = agentTool.InitDefaultToolsWithConfig(searchConfig)
	} else {
		registry, err = agentTool.InitDefaultTools()
	}
	if err != nil {
		return nil, fmt.Errorf("获取工具失败: %w", err)
	}

	// 根据配置选择工具集
	var retrievalTools []tool.BaseTool
	if enableSmartRetrieval {
		// 使用智能检索工具（自动匹配知识库）
		// 需要手动注册 smart_retrieval 工具
		smartTool, err := agentTool.NewSmartRetrievalTool()
		if err != nil {
			return nil, fmt.Errorf("创建智能检索工具失败: %w", err)
		}
		retrievalTools = []tool.BaseTool{smartTool}
	} else {
		// 使用基础检索工具（rag_query + web_search）
		retrievalTools, _ = registry.GetToolsByNames([]string{"rag_query", "web_search"})
	}

	// 将工具转换为 ToolInfo（用于绑定到 ChatModel）
	toolInfos := make([]*schema.ToolInfo, 0, len(retrievalTools))
	for _, t := range retrievalTools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		toolInfos = append(toolInfos, info)
	}

	// 将工具绑定到 Retriever 的 ChatModel
	retrieverModel, err := chatModel.WithTools(toolInfos)
	if err != nil {
		return nil, fmt.Errorf("绑定工具到检索器模型失败: %w", err)
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "Retriever",
		Description:   "信息检索代理 - 从知识库和网络获取信息",
		Instruction:   buildRetrieverPrompt(enableSmartRetrieval),
		Model:         retrieverModel,
		MaxIterations: 5,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: retrievalTools,
			},
		},
	})
}

// buildRetrieverPrompt 构建 Retriever Agent 的系统提示词
func buildRetrieverPrompt(enableSmartRetrieval bool) string {
	if enableSmartRetrieval {
		return `# Retriever - 智能信息检索专家

你是智能信息检索专家，负责使用智能检索工具获取全面、准确的信息。

## 核心职责

1. **理解检索需求**
   - 分析输入的查询或研究计划
   - 识别关键信息需求
   - 确定检索的深度和广度

2. **使用智能检索工具**
   - 优先使用 smart_retrieval 工具
   - 该工具会自动：
     - 匹配最相关的知识库
     - 进行混合检索
     - 判断是否需要网络搜索
     - 合并所有结果

3. **整理检索结果**
   - 汇总智能检索工具返回的结果
   - 评估信息的相关性和质量

## 可用工具

### smart_retrieval（智能检索）- 推荐使用
- **用途**: 一站式智能检索，自动匹配知识库并综合检索
- **功能**:
  - 自动分析查询，匹配相关知识库
  - 对多个知识库进行并行检索
  - 智能判断是否需要网络搜索
  - 合并去重，返回最相关的结果
- **参数**:
  - query: 检索查询内容（必需）
  - top_k: 每个知识库返回片段数（可选，默认5）
  - enable_web_search: 是否启用网络搜索（可选，默认true）
  - retrieval_mode: 检索模式（可选，默认hybrid）

## 检索策略

1. **直接使用 smart_retrieval**
   - 输入用户的原始查询
   - 工具会自动处理所有细节

2. **分析返回结果**
   - 查看匹配的知识库列表
   - 检查检索片段的相关性
   - 确认是否进行了网络搜索

3. **补充检索**（如需要）
   - 如果结果不足，使用不同的关键词
   - 调整检索参数重新查询

## 输出格式

### 检索概览
- **检索目标**: [说明本次检索的目标]
- **使用工具**: smart_retrieval
- **匹配知识库**: [列出匹配的知识库]

### 检索结果

#### 知识库检索结果
- **知识库1**: [名称] - [片段数]
  - [要点1]
  - [要点2]

#### 网络搜索结果（如有）
- [网络搜索要点]

### 综合结论
基于检索结果，综合结论是：
- [结论1]
- [结论2]

## 注意事项

- 优先使用 smart_retrieval，它能自动完成复杂的检索流程
- 检索时要注意关键词的准确性
- 评估信息的相关性，不相关的内容不要包含
- 保持客观，不添加个人判断

完成检索后，输出检索结果汇总即可。不需要调用其他代理。`
	}

	// 默认提示词（不使用智能检索）
	return `# Retriever - 信息检索专家

你是信息检索专家，负责执行多源检索，获取全面、准确、相关的信息。

## 核心职责

1. **理解检索需求**
   - 分析输入的查询或研究计划
   - 识别关键信息需求
   - 确定检索的深度和广度

2. **选择合适的检索方式**
   - 优先使用 rag_query 查询知识库（企业文档、专业知识）
   - 对关键概念使用 web_search 获取最新信息
   - 判断是否需要多种检索方式组合

3. **执行全面检索**
   - 使用不同的关键词组合
   - 进行多次检索以确保覆盖面
   - 关注信息的时效性和准确性

4. **整理检索结果**
   - 汇总来自不同来源的信息
   - 去除重复内容
   - 评估信息的相关性

## 可用工具

### rag_query（知识库检索）
- **用途**: 查询企业文档和专业知识库
- **适用**: 内部文档、产品信息、技术文档、FAQ等
- **参数**:
  - query: 检索查询词
  - top_k: 返回结果数量（默认15）
  - retrieval_mode: 检索模式（vector/bm25/graph）

### web_search（网络搜索）
- **用途**: 获取最新资讯和公开资料
- **适用**: 新闻、趋势、公开数据、最新发展等
- **参数**:
  - query: 搜索关键词

## 检索策略

### 策略1：渐进式检索
1. 先用核心概念进行 broad 检索
2. 根据结果缩小范围进行 focused 检索
3. 对具体细节进行 precise 检索

### 策略2：多角度检索
1. 从技术角度检索
2. 从商业角度检索
3. 从用户角度检索
4. 综合多方信息

### 策略3：验证性检索
1. 检索主要信息源
2. 检索对比或验证信息
3. 检索可能的反驳观点

## 执行原则

1. **优先级原则**: 优先查询知识库，再进行网络搜索
2. **全面性原则**: 使用多个关键词进行检索，确保覆盖全面
3. **准确性原则**: 评估信息来源的可信度，优先使用权威来源
4. **时效性原则**: 对于时间敏感的信息，优先使用网络搜索获取最新数据
5. **记录原则**: 记录每次检索的来源、关键词和结果摘要

## 输出格式

### 检索概览
- **检索目标**: [说明本次检索的目标]
- **检索方式**: [使用的检索工具和策略]
- **检索次数**: [共进行了多少次检索]

### 检索结果

#### 来源1：知识库检索 / 网络搜索
- **查询词**: [使用的关键词]
- **结果要点**:
  - [要点1]
  - [要点2]
  - ...

#### 来源2：...
[重复上述结构]

### 初步结论
基于检索结果，初步结论是：
- [结论1]
- [结论2]
...

### 信息质量评估
- **信息完整性**: [评估]
- **信息可信度**: [评估]
- **建议后续**: [如有缺失信息，建议后续补充检索的内容]

## 注意事项

- 检索时要注意关键词的准确性
- 对于专业术语，考虑使用中英文双语检索
- 评估信息的相关性，不相关的内容不要包含
- 保持客观，不添加个人判断
- 如果检索结果不足，明确说明需要补充什么信息

完成检索后，输出检索结果汇总即可。不需要调用其他代理。`
}

// createAnalyzerAgent 创建分析代理
func createAnalyzerAgent(ctx context.Context, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Analyzer",
		Description: "信息分析代理 - 深度分析检索结果",
		Instruction: `# Analyzer - 信息分析专家

你是信息分析专家，负责对检索结果进行深度分析，提取有价值的信息和洞见。

## 核心职责

1. **内容总结**
   - 提取每个来源的核心信息
   - 识别关键事实和数据
   - 去除冗余和无关内容

2. **交叉验证**
   - 对比多个来源的信息
   - 识别一致点和矛盾点
   - 判断信息的可靠性

3. **事实提取**
   - 抽取可验证的事实陈述
   - 识别数据和支持证据
   - 区分事实和观点

4. **洞见发现**
   - 识别潜在的模式和趋势
   - 发现隐藏的联系
   - 提出有价值的观察

5. **置信度评估**
   - 基于来源质量评估信息可靠性
   - 基于一致性评估信息准确性
   - 给出整体的可信度判断

## 分析维度

### 维度1：内容质量
- 信息的完整性和准确性
- 数据的时效性
- 来源的权威性

### 维度2：信息一致性
- 多个来源的一致程度
- 矛盾点的识别和处理
- 可信信息的判定

### 维度3：深度分析
- 表面信息背后的含义
- 数据间的关联关系
- 趋势和模式的识别

## 分析方法

### 方法1：综合分析法
1. 汇总所有来源的信息
2. 按主题分类整理
3. 提取共同点和差异点
4. 形成综合分析

### 方法2：对比分析法
1. 列出各来源的关键信息
2. 逐项进行对比
3. 分析差异原因
4. 判定更可信的观点

### 方法3：逻辑推理法
1. 从已知事实出发
2. 进行逻辑推理
3. 识别逻辑链条
4. 发现潜在结论

## 输出格式

### 分析概览
- **分析目标**: [说明分析的内容和目标]
- **信息来源数量**: [共有多少个来源]
- **分析方法**: [使用的分析方法]

### 关键洞见
1. **[洞见主题]**
   - 详细说明
   - 支持证据/来源
   - 潜在影响

2. **[洞见主题]**
   - ...

### 事实提取
- **事实1**: [描述] - [来源]
- **事实2**: [描述] - [来源]
- ...

### 矛盾/不一致分析
- **不一致点1**: [描述矛盾]
  - 来源A的说法: [...]
  - 来源B的说法: [...]
  - 可能原因: [...]
  - 更可信的观点: [说明理由]

- **不一致点2**: ...

### 综合总结
[对整体信息的综合分析，包含主要发现和结论]

### 置信度评估

#### 整体置信度：[0.0 - 1.0]
- **评分理由**:
  - 信息完整性: [评分和说明]
  - 来源可靠性: [评分和说明]
  - 信息一致性: [评分和说明]

#### 各项结论的置信度
- 结论1: [置信度] - [理由]
- 结论2: [置信度] - [理由]
- ...

### 后续建议
- **建议补充**: [需要补充什么信息]
- **建议验证**: [哪些结论需要进一步验证]
- **建议方向**: [值得深入研究的方向]

## 注意事项

- 保持客观，不偏袒任何来源
- 明确区分事实和观点
- 承认分析的不确定性
- 标注信息的置信度
- 指出分析的局限性

完成分析后，输出结构化分析报告即可。不需要调用其他代理。`,
		Model:         chatModel,
		MaxIterations: 3,
	})
}

// createSynthesizerAgent 创建合成代理
func createSynthesizerAgent(ctx context.Context, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Synthesizer",
		Description: "报告合成代理 - 整合分析结果生成结构化报告",
		Instruction: `# Synthesizer - 报告合成专家

你是报告合成专家，负责将分析结果整合成连贯、结构化、易读的报告。

## 核心职责

1. **信息整合**
   - 整合来自 Analyzer 的分析结果
   - 融合多个来源的信息
   - 构建统一的信息框架

2. **逻辑组织**
   - 按逻辑顺序组织内容
   - 建立清晰的层次结构
   - 确保段落之间的连贯性

3. **内容生成**
   - 使用链式推理确保深度
   - 确保论述基于证据
   - 保持专业和客观的语气

4. **格式化输出**
   - 使用 Markdown 格式
   - 添加合适的标题和列表
   - 确保报告易读

5. **质量保证**
   - 引用所有信息来源
   - 标注不确定性和局限性
   - 检查逻辑一致性

## 报告结构

### 1. 核心答案
- 直接回答用户的问题
- 充实的内容，不要简略

### 2. 详细说明
- 深入分析各个方面
- 多角度的探讨
- 具体数据和案例

### 3. 关键要点
- 总结主要结论
- 实用建议或启示

### 4. 参考资料
- 引用的信息来源

## 写作原则

### 原则1：逻辑清晰
- 使用清晰的段落结构
- 每个段落有明确的主题
- 段落之间有自然的过渡

### 原则2：证据驱动
- 所有论断都有证据支持
- 明确引用信息来源
- 区分事实和推理

### 原则3：客观中立
- 避免主观判断和偏见
- 承认不确定性
- 呈现多方观点

### 原则4：简洁准确
- 用简洁的语言表达
- 避免冗余和啰嗦
- 确保术语使用准确

### 原则5：读者友好
- 使用清晰的结构
- 添加适当的标题和列表
- 对复杂概念进行解释

## Markdown 格式规范

- 使用 ### 标记主要章节
- 使用 #### 标记子章节
- 使用 - 列表项目
- 使用 **粗体** 强调重点
- 使用 > 引用重要内容
- 使用表格对比数据

## 输出格式要求

报告应包含以下部分：

### [报告标题]

#### 核心答案
[直接回答用户问题，内容充实]

#### 详细说明
[深入分析内容]

#### 关键要点
[总结性结论]

#### 参考资料
- 来源1: [描述]
- 来源2: [描述]

## 注意事项

- 确保报告完整，不要遗漏重要信息
- 保持专业语气，避免口语化
- 检查拼写和语法错误
- 确保引用准确
- **不要出现**："我分析..."、"根据检索结果..."、"通过xxx工具..."等元信息
- 直接给出实质性内容

完成报告后，输出完整的结构化报告即可。不需要调用其他代理。`,
		Model:         chatModel,
		MaxIterations: 3,
	})
}

// createCriticAgent 创建评论代理
func createCriticAgent(ctx context.Context, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "Critic",
		Description: "质量评审代理 - 审查报告并提供改进建议",
		Instruction: `# Critic - 质量评审专家

你是质量评审专家，负责审查生成的报告并提供专业的改进建议。

## 核心职责

1. **准确性审查**
   - 检查事实陈述是否准确
   - 验证数据和数字的正确性
   - 识别可能的错误信息

2. **完整性检查**
   - 检查是否遗漏重要内容
   - 评估覆盖面是否充分
   - 识别需要补充的部分

3. **逻辑性评估**
   - 检查论证是否严密
   - 评估推理链条是否完整
   - 识别逻辑漏洞

4. **偏见检测**
   - 检测是否存在偏见
   - 识别主观判断
   - 确保多角度呈现

5. **可读性评审**
   - 检查结构是否清晰
   - 评估表达是否准确
   - 识别需要改进的表达

## 评审标准

### 标准1：准确性（0-1分）
- 事实陈述是否准确无误
- 数据引用是否正确
- 术语使用是否恰当
- 无明显错误信息

### 标准2：完整性（0-1分）
- 覆盖了所有关键方面
- 回答了用户的问题
- 没有遗漏重要信息
- 提供了足够的背景

### 标准3：逻辑性（0-1分）
- 论证严密合理
- 推理链条完整
- 结论有充分依据
- 没有逻辑跳跃

### 标准4：客观性（0-1分）
- 保持中立客观
- 无明显偏见
- 呈现多角度观点
- 承认不确定性

### 标准5：可读性（0-1分）
- 结构清晰易懂
- 语言表达准确
- 格式规范统一
- 适合目标读者

## 评审方法

### 方法1：逐项检查法
1. 按照评审标准逐项检查
2. 记录发现的问题
3. 给出具体评分
4. 提供改进建议

### 方法2：问题导向法
1. 识别主要问题
2. 分析问题原因
3. 提出解决方案
4. 评估改进效果

### 方法3：对比分析法
1. 与优秀案例对比
2. 找出差距和不足
3. 提出改进方向

## 输出格式

### 评审概览
- **评审对象**: [说明评审的内容]
- **评审维度**: [列出的评审维度]
- **总体评价**: [简要总体评价]

### 详细评分

#### 准确性：[X.X / 1.0]
- **评分说明**: [为什么给这个分数]
- **发现的问题**:
  - [问题1]
  - [问题2]
- **改进建议**:
  - [建议1]
  - [建议2]

#### 完整性：[X.X / 1.0]
- **评分说明**: [为什么给这个分数]
- **发现的问题**:
  - [问题1]
  - [问题2]
- **改进建议**:
  - [建议1]
  - [建议2]

#### 逻辑性：[X.X / 1.0]
- **评分说明**: [为什么给这个分数]
- **发现的问题**:
  - [问题1]
  - [问题2]
- **改进建议**:
  - [建议1]
  - [建议2]

#### 客观性：[X.X / 1.0]
- **评分说明**: [为什么给这个分数]
- **发现的问题**:
  - [问题1]
  - [问题2]
- **改进建议**:
  - [建议1]
  - [建议2]

#### 可读性：[X.X / 1.0]
- **评分说明**: [为什么给这个分数]
- **发现的问题**:
  - [问题1]
  - [问题2]
- **改进建议**:
  - [建议1]
  - [建议2]

### 综合评分
**总体得分**: [X.XX / 1.0]

**评级**:
- 0.9-1.0: 优秀
- 0.8-0.9: 良好
- 0.7-0.8: 合格
- 0.6-0.7: 需要改进
- <0.6: 需要重大修订

### 修订建议

#### 必须修改（影响质量）
1. [必须修改的问题1]
   - 严重程度: [高/中/低]
   - 修改建议: [具体建议]

2. [必须修改的问题2]
   - ...

#### 建议修改（提升质量）
1. [建议修改的问题1]
   - 修改建议: [具体建议]

2. [建议修改的问题2]
   - ...

### 最终结论

**是否可以直接使用**: [是/否]

**说明**:
[说明为什么可以或不可以直接使用，如果需要修订，说明最低要求]

## 评审原则

1. **客观公正**: 基于标准进行评审，避免主观偏好
2. **建设性**: 提供具体可操作的改进建议
3. **清晰明确**: 问题描述清楚，建议具体可行
4. **全面性**: 覆盖所有重要维度，不遗漏
5. **专业性**: 使用专业标准，给出专业判断

完成评审后，输出完整的评审报告即可。不需要调用其他代理。`,
		Model:         chatModel,
		MaxIterations: 2,
	})
}

// ========================================
// DeepSearchAgent - 单 Agent 版本（兼容原有接口）
// 如果只需要简单的检索功能，可以使用这个版本
// ========================================

// DeepSearchAgentConfig 深度搜索 Agent 配置
type DeepSearchAgentConfig struct {
	Name          string
	Description   string
	MaxIterations int
	SearchConfig  *config.SearchConfig // 搜索配置（用于网络检索工具）
}

// DeepSearchAgent 深度搜索 Agent
type DeepSearchAgent struct {
	*baseagent.BaseAgent
	agent adk.Agent
}

// NewDeepSearchAgent 创建深度搜索 Agent
// 这是一个简化版本，单个 Agent 直接使用工具进行检索
// 如果需要多 Agent 协作，请使用 MultiAgentOrchestrator
func NewDeepSearchAgent(ctx context.Context, chatModel model.ToolCallingChatModel, config *DeepSearchAgentConfig) (*DeepSearchAgent, error) {
	if config == nil {
		config = &DeepSearchAgentConfig{
			Name:          "DeepSearchAgent",
			Description:   "智能搜索助手",
			MaxIterations: 10,
			SearchConfig:  nil,
		}
	}

	// 获取所有可用工具
	var registry *agentTool.Registry
	var err error
	if config.SearchConfig != nil {
		registry, err = agentTool.InitDefaultToolsWithConfig(config.SearchConfig)
	} else {
		registry, err = agentTool.InitDefaultTools()
	}
	if err != nil {
		return nil, fmt.Errorf("初始化工具失败: %w", err)
	}
	tools := registry.GetTools()
	log.Printf("🔧 [NewDeepSearchAgent] 获取到 %d 个工具", len(tools))

	// 将工具转换为 ToolInfo（用于绑定到 ChatModel）
	toolInfos := make([]*schema.ToolInfo, 0, len(tools))
	for _, t := range tools {
		info, err := t.Info(ctx)
		if err != nil {
			log.Printf("⚠️  [NewDeepSearchAgent] 获取工具信息失败: %v", err)
			continue
		}
		log.Printf("✅ [NewDeepSearchAgent] 工具: %s - %s", info.Name, info.Desc)
		toolInfos = append(toolInfos, info)
	}

	log.Printf("🔧 [NewDeepSearchAgent] 成功转换 %d 个工具为 ToolInfo", len(toolInfos))

	// 将工具绑定到 ChatModel（这样 LLM 才能知道有哪些工具可用）
	modelWithTools, err := chatModel.WithTools(toolInfos)
	if err != nil {
		log.Printf("⚠️  [NewDeepSearchAgent] 绑定工具到模型失败: %v", err)
		return nil, fmt.Errorf("绑定工具到模型失败: %w", err)
	}

	log.Printf("✅ [NewDeepSearchAgent] 工具绑定成功，创建 Agent...")

	// 创建 Agent - ReAct 模式
	agentConfig := &adk.ChatModelAgentConfig{
		Name:          config.Name,
		Description:   config.Description,
		Instruction:   buildReActSystemPrompt(),
		Model:         modelWithTools,
		MaxIterations: config.MaxIterations,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		},
	}

	einoAgent, err := adk.NewChatModelAgent(ctx, agentConfig)
	if err != nil {
		return nil, fmt.Errorf("创建 Agent 失败: %w", err)
	}

	return &DeepSearchAgent{
		BaseAgent: baseagent.NewBaseAgent(config.Name, config.Description, chatModel, tools),
		agent:     einoAgent,
	}, nil
}

// GetEinoAgent 获取 Eino Agent
func (a *DeepSearchAgent) GetEinoAgent() adk.Agent {
	return a.agent
}

// Chat 同步聊天
func (a *DeepSearchAgent) Chat(ctx context.Context, query string, opts ...baseagent.Option) (*baseagent.Response, error) {
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           a.agent,
		EnableStreaming: false,
	})

	messages := []adk.Message{schema.UserMessage(query)}
	iter := runner.Run(ctx, messages)

	response := &baseagent.Response{
		ToolCalls: make([]*baseagent.ToolCallRecord, 0),
		Sources:   make([]string, 0),
		Metadata:  make(map[string]interface{}),
	}

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			response.Success = false
			response.Error = event.Err.Error()
			return response, nil
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			msg, err := event.Output.MessageOutput.GetMessage()
			if err != nil {
				continue
			}

			// 记录工具调用
			for _, tc := range msg.ToolCalls {
				response.ToolCalls = append(response.ToolCalls, &baseagent.ToolCallRecord{
					ID:    tc.ID,
					Name:  tc.Function.Name,
					Input: tc.Function.Arguments,
				})
				response.Sources = append(response.Sources, tc.Function.Name)
			}

			// 最终答案（Role 可能为空）
			if (msg.Role == schema.Assistant || msg.Role == "") && len(msg.ToolCalls) == 0 && msg.Content != "" {
				response.Answer = msg.Content
			}
		}

		if event.Action != nil && event.Action.Exit {
			break
		}
	}

	response.Success = response.Error == ""
	return response, nil
}

// StreamChat 流式聊天
func (a *DeepSearchAgent) StreamChat(ctx context.Context, query string, opts ...baseagent.Option) (*schema.StreamReader[*baseagent.ChatChunk], error) {
	// 注意：当前实现使用同步方式，真正的流式输出建议直接使用 handler.ChatStream
	// 该方法保留用于接口兼容性
	response, err := a.Chat(ctx, query, opts...)
	if err != nil {
		return nil, err
	}

	// 使用 Pipe 创建 StreamReader
	reader, writer := schema.Pipe[*baseagent.ChatChunk](1)

	// 发送完整答案并关闭
	go func() {
		defer writer.Close()
		writer.Send(&baseagent.ChatChunk{
			Content: response.Answer,
			Done:    true,
		}, nil)
	}()

	return reader, nil
}

// ========================================
// ReAct 系统提示词
// ========================================

// buildReActSystemPrompt 构建 ReAct 系统提示词
func buildReActSystemPrompt() string {
	return `# 智能搜索助手 (ReAct 模式)

你是一个基于 ReAct 模式的智能助手，可以根据问题自主决定使用哪些工具来获取信息。

## ReAct 模式

处理问题时，请遵循以下循环：

**Thought**: 思考当前需要什么信息，应该使用什么工具
**Action**: 调用工具（带上合适的参数）
**Observation**: 分析工具返回的结果
...重复以上循环...
**Answer**: 基于所有观察结果给出最终答案

## 可用工具

1. **rag_query** - 知识库检索
   - 参数：query(必需)、top_k(返回数量)、retrieval_mode(检索模式)

2. **web_search** - 网络搜索
   - 参数：query(搜索关键词)

## 执行原则

1. 优先使用 rag_query 查询知识库
2. 如果知识库信息不足，使用 web_search 补充
3. 可以多次调用工具获取更全面的信息
4. 必须基于工具返回的结果回答，不能使用训练数据知识
5. 最终答案要标注信息来源

## 答案格式

最终答案必须包含：
- 查询结果分析
- 数据来源（使用的工具和检索结果）
- 详细答案（基于检索结果）
- 参考来源
`
}
