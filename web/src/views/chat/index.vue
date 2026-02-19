<template>
  <div class="chat-page">
    <!-- 侧边栏：会话列表 -->
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header">
        <button v-if="!sidebarCollapsed" class="new-chat-btn" @click="createNewSession">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 5v14M5 12h14"/>
          </svg>
          <span>新对话</span>
        </button>
        <button class="collapse-btn" @click="toggleSidebar" :title="sidebarCollapsed ? '展开' : '收起'">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path v-if="!sidebarCollapsed" d="M11 19l-7-7 7-7M18 19l-7-7 7-7"/>
            <path v-else d="M15 19l-7-7 7-7M3 12h18"/>
          </svg>
        </button>
      </div>

      <div class="sessions-list" v-loading="loadingSessions">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="session-item"
          :class="{ active: currentSessionId === session.id }"
          @click="selectSession(session.id)"
        >
          <div class="session-icon">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
          </div>
          <span class="session-title">{{ session.title }}</span>
          <button class="delete-btn" @click.stop="deleteSession(session.id)" title="删除">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
            </svg>
          </button>
        </div>
      </div>
    </aside>

    <!-- 主区域 -->
    <main class="main-content">
      <!-- 顶部工具栏 -->
      <header class="chat-header">
        <div class="header-left">
          <!-- 模式切换 Tab -->
          <div class="mode-tabs">
            <button
              class="mode-tab"
              :class="{ active: chatMode === 'normal' }"
              @click="switchMode('normal')"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
              </svg>
              <span>普通对话</span>
            </button>
            <button
              class="mode-tab"
              :class="{ active: chatMode === 'agent' }"
              @click="switchMode('agent')"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <path d="M12 16v-4M12 8h.01"/>
              </svg>
              <span>Agent 模式</span>
            </button>
          </div>
        </div>
        <div class="header-right" v-if="chatMode === 'normal'">
          <!-- RAG 开关和设置 -->
          <button
            class="rag-toggle-btn"
            :class="{ active: ragConfig.enabled }"
            @click="toggleRAG"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 2L2 7l10 5 10-5-10 5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
            </svg>
            <span>RAG</span>
          </button>
          <button class="settings-btn" @click="showRAGSettings = true" title="检索设置">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="3"/>
              <path d="M12 1v6m0 6v6m9-9l-4.5 4.5M15 15l-4.5 4.5M9 9l4.5-4.5M15 9l-4.5 4.5"/>
            </svg>
          </button>
        </div>
        <div class="header-right" v-else>
          <!-- Agent 模式说明 -->
          <div class="agent-mode-info">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 2a2 2 0 0 1 2 2c0 .74-.4 1.39-1 1.73V7h1a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1H9a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1V5.73A2 2 0 0 1 12 2Z"/>
            </svg>
            <span>智能 Agent - 自动选择工具进行深度搜索</span>
          </div>
        </div>
      </header>

      <!-- 消息区域 -->
      <div class="messages-area" ref="messagesContainer">
        <!-- 空状态 -->
        <div v-if="(chatMode === 'normal' && displayMessages.length === 0 && !isStreaming) ||
                    (chatMode === 'agent' && agentMessages.length === 0 && !isAgentStreaming)" class="empty-state">
          <div class="empty-icon" :class="{ 'agent-icon': chatMode === 'agent' }">
            <svg v-if="chatMode === 'agent'" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="12" cy="12" r="10"/>
              <path d="M12 16v-4M12 8h.01"/>
            </svg>
            <svg v-else width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M8 12h.01M12 16h.01M16 12h.01M21 12c0 4.97-4.03 9-9 9s-9-4.03-9-9 4.03-9 9-9 9 4.03 9 9z"/>
            </svg>
          </div>
          <p>{{ chatMode === 'agent' ? 'Agent 智能助手' : '开始一段新对话' }}</p>
          <p class="hint">{{ chatMode === 'agent' ? '我可以使用多种工具来帮助你获取信息' : '按 Enter 发送，Shift + Enter 换行' }}</p>
        </div>

        <!-- 统一消息列表 -->
        <div class="messages-list" :class="{ 'agent-mode': chatMode === 'agent' }">
          <!-- RAG 上下文提示 (仅普通模式) -->
          <div v-if="chatMode === 'normal' && currentRAGContext" class="rag-context-banner">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 12l2 2 4-4"/>
              <circle cx="12" cy="12" r="9"/>
            </svg>
            <span>已检索到 {{ currentRAGContext.retrieved_count }} 个相关文档</span>
            <span class="sources">{{ currentRAGContext.source_types?.join(', ') }}</span>
          </div>

          <!-- 消息列表 -->
          <div v-for="message in displayMessages" :key="message.id" style="margin-bottom: 24px;">
            <!-- 用户消息 -->
            <div v-if="message.role === 'user'" class="message user">
              <div class="message-avatar">
                <span>{{ userInitial }}</span>
              </div>
              <div class="message-content">
                <div class="message-text">{{ message.content }}</div>
              </div>
            </div>

            <!-- 助手消息 -->
            <div v-else class="message-group">
              <!-- 普通消息显示 -->
              <div class="message assistant">
                <div class="message-avatar">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M12 2a2 2 0 0 1 2 2c0 .74-.4 1.39-1 1.73V7h1a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1H9a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1V5.73A2 2 0 0 1 12 2Z"/>
                    <path d="M8 14h1v4H8v-4Zm6 0h1v4h-1v-4Z"/>
                  </svg>
                </div>
                <div class="message-content">
                  <div class="message-text" v-html="renderMarkdown(message.content)"></div>
                  <div class="message-actions">
                    <span class="message-time">{{ formatTime(message.created_at) }}</span>
                    <button class="action-btn" @click="copyMessage(message.content)" title="复制">
                      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
                      </svg>
                    </button>
                  </div>
                </div>
              </div>

              <!-- Agent 步骤（如果有） -->
              <div v-if="message.agent_steps" class="agent-steps">
                <div class="agent-response">
                  <div class="thinking-steps">
                    <div
                      v-for="step in parseAgentSteps(message.agent_steps)"
                      :key="step.id"
                      class="step-item"
                      :class="step.type"
                    >
                      <!-- 工具调用步骤 -->
                      <template v-if="step.type === 'action'">
                        <div class="step-header">
                          <span class="step-number">步骤 {{ step.step }}</span>
                          <span class="step-type">调用工具</span>
                        </div>
                        <div class="step-content">
                          <div v-if="step.thought" class="step-thought">
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                              <path d="M9.5 2A2.5 2.5 0 0 1 12 4.5v15a2.5 2.5 0 0 1-4.96.44 2.5 2.5 0 0 1-2.96-3.08 3 3 0 0 1-.34-5.55 2.5 2.5 0 0 1 1.32-4.24 2.5 2.5 0 0 1 4.44-4A2.5 2.5 0 0 1 9.5 2Z"/>
                              <path d="M14.5 2A2.5 2.5 0 0 0 12 4.5v15a2.5 2.5 0 0 0 4.96.44 2.5 2.5 0 0 0 2.96-3.08 3 3 0 0 0 .34-5.55 2.5 2.5 0 0 0-1.32-4.24 2.5 2.5 0 0 0-4.44-4A2.5 2.5 0 0 0 14.5 2Z"/>
                            </svg>
                            <span>{{ step.thought }}</span>
                          </div>
                          <div class="tool-call-card">
                            <div class="tool-name">
                              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
                              </svg>
                              <span>{{ step.tool_name }}</span>
                            </div>
                            <div v-if="step.tool_params" class="tool-params">
                              <span class="param-label">参数:</span>
                              <code class="param-value">{{ formatToolParams(step.tool_params) }}</code>
                            </div>
                            <div v-if="step.tool_output" class="tool-output">
                              <div class="output-header">
                                <span class="output-label">执行结果:</span>
                                <button class="toggle-btn" @click="toggleToolOutput(step.id)">
                                  {{ expandedToolOutputs[step.id] ? '收起' : '展开' }}
                                </button>
                              </div>
                              <div v-if="expandedToolOutputs[step.id]" class="output-value">
                                <pre>{{ formatToolOutput(step.tool_output) }}</pre>
                              </div>
                              <div v-else class="output-value output-truncated">
                                {{ truncateText(step.tool_output, 150) }}
                              </div>
                            </div>
                          </div>
                        </div>
                      </template>

                      <!-- 思考步骤 -->
                      <template v-else-if="step.type === 'thought'">
                        <div class="step-header">
                          <span class="step-number">步骤 {{ step.step }}</span>
                          <span class="step-type">思考</span>
                        </div>
                        <div class="step-content">
                          <div class="thought-content" v-html="renderMarkdown(step.content || '')"></div>
                        </div>
                      </template>

                      <!-- 搜索/检索步骤 -->
                      <template v-else-if="step.type === 'search' || step.type === 'retrieval'">
                        <div class="step-header search">
                          <span class="step-number">步骤 {{ step.step }}</span>
                          <span class="step-type">{{ step.stage || '信息检索' }}</span>
                        </div>
                        <div class="step-content">
                          <div class="tool-call-card">
                            <div class="tool-name">
                              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <circle cx="11" cy="11" r="8"/>
                                <path d="m21 21-4.35-4.35"/>
                              </svg>
                              <span>{{ step.tool_desc || step.tool_name || '检索工具' }}</span>
                            </div>
                            <div v-if="step.tool_params" class="tool-params">
                              <span class="param-label">参数:</span>
                              <code class="param-value">{{ formatToolParams(step.tool_params) }}</code>
                            </div>
                          </div>
                        </div>
                      </template>

                      <!-- 分析/综合步骤 -->
                      <template v-else-if="step.type === 'analysis' || step.type === 'synthesis' || step.type === 'plan'">
                        <div class="step-header analysis">
                          <span class="step-number">步骤 {{ step.step }}</span>
                          <span class="step-type">{{ step.stage || '分析' }}</span>
                        </div>
                        <div class="step-content">
                          <div class="thought-content" v-html="renderMarkdown(step.content || '')"></div>
                        </div>
                      </template>

                      <!-- 完成步骤 -->
                      <template v-else-if="step.type === 'complete'">
                        <div class="step-header complete">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
                            <path d="m9 11 3 3L22 4"/>
                          </svg>
                          <span>{{ step.reason || 'Agent 完成执行' }}</span>
                        </div>
                      </template>

                      <!-- 错误步骤 -->
                      <template v-else-if="step.type === 'error'">
                        <div class="step-header error">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="12" cy="12" r="10"/>
                            <path d="m15 9-6 6m0-6 6 6"/>
                          </svg>
                          <span>{{ step.content || '执行出错' }}</span>
                        </div>
                      </template>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 普通模式流式输出 -->
          <div v-if="chatMode === 'normal' && streamingContent" class="message assistant streaming">
            <div class="message-avatar">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M12 2a2 2 0 0 1 2 2c0 .74-.4 1.39-1 1.73V7h1a1 1 0 0 1 1 1v3a1 1 0 0 1-1 1H9a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1V5.73A2 2 0 0 1 12 2Z"/>
                <path d="M8 14h1v4H8v-4Zm6 0h1v4h-1v-4Z"/>
              </svg>
            </div>
            <div class="message-content">
              <div class="message-text" v-html="renderMarkdown(streamingContent)"></div>
              <div class="message-actions">
                <span class="streaming-indicator">正在输入</span>
              </div>
            </div>
          </div>

          <!-- Agent 模式流式输出 -->
          <div v-if="chatMode === 'agent' && isAgentStreaming" class="agent-message streaming">
            <div class="agent-response">
              <div class="agent-header">
                <div class="agent-avatar streaming">
                  <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <circle cx="12" cy="12" r="10"/>
                    <path d="M12 16v-4M12 8h.01"/>
                  </svg>
                </div>
                <span class="agent-label">Agent</span>
                <span class="streaming-indicator">思考中...</span>
              </div>

              <div class="thinking-steps">
                <div
                  v-for="step in currentAgentSteps"
                  :key="step.id"
                  class="step-item"
                  :class="[step.type, { active: step.isActive }]"
                >
                  <!-- 工具调用步骤 -->
                  <template v-if="step.type === 'action'">
                    <div class="step-header">
                      <span class="step-number">步骤 {{ step.step }}</span>
                      <span class="step-type">调用工具</span>
                      <span v-if="step.isActive" class="step-status">执行中...</span>
                    </div>
                    <div class="step-content">
                      <div v-if="step.thought" class="step-thought">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M9.5 2A2.5 2.5 0 0 1 12 4.5v15a2.5 2.5 0 0 1-4.96.44 2.5 2.5 0 0 1-2.96-3.08 3 3 0 0 1-.34-5.55 2.5 2.5 0 0 1 1.32-4.24 2.5 2.5 0 0 1 4.44-4A2.5 2.5 0 0 1 9.5 2Z"/>
                          <path d="M14.5 2A2.5 2.5 0 0 0 12 4.5v15a2.5 2.5 0 0 0 4.96.44 2.5 2.5 0 0 0 2.96-3.08 3 3 0 0 0 .34-5.55 2.5 2.5 0 0 0-1.32-4.24 2.5 2.5 0 0 0-4.44-4A2.5 2.5 0 0 0 14.5 2Z"/>
                        </svg>
                        <span>{{ step.thought }}</span>
                      </div>
                      <div class="tool-call-card">
                        <div class="tool-name">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
                          </svg>
                          <span>{{ step.tool_name }}</span>
                          <span v-if="step.isActive" class="tool-loading"></span>
                        </div>
                        <div v-if="step.tool_params" class="tool-params">
                          <span class="param-label">参数:</span>
                          <code class="param-value">{{ formatToolParams(step.tool_params) }}</code>
                        </div>
                        <div v-if="step.tool_output" class="tool-output">
                          <div class="output-header">
                            <span class="output-label">执行结果:</span>
                            <button class="toggle-btn" @click="toggleToolOutput(step.id)">
                              {{ expandedToolOutputs[step.id] ? '收起' : '展开' }}
                            </button>
                          </div>
                          <div v-if="expandedToolOutputs[step.id]" class="output-value">
                            <pre>{{ formatToolOutput(step.tool_output) }}</pre>
                          </div>
                          <div v-else class="output-value output-truncated">
                            {{ truncateText(step.tool_output, 150) }}
                          </div>
                        </div>
                      </div>
                    </div>
                  </template>

                  <!-- 思考步骤 -->
                  <template v-else-if="step.type === 'thought'">
                    <div class="step-header">
                      <span class="step-number">步骤 {{ step.step }}</span>
                      <span class="step-type">思考</span>
                    </div>
                    <div class="step-content">
                      <div class="thought-content" v-html="renderMarkdown(step.content || '')"></div>
                    </div>
                  </template>

                  <!-- 搜索/检索步骤 -->
                  <template v-else-if="step.type === 'search' || step.type === 'retrieval'">
                    <div class="step-header search">
                      <span class="step-number">步骤 {{ step.step }}</span>
                      <span class="step-type">{{ step.stage || '信息检索' }}</span>
                      <span v-if="step.isActive" class="step-status">执行中...</span>
                    </div>
                    <div class="step-content">
                      <div class="tool-call-card">
                        <div class="tool-name">
                          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="11" cy="11" r="8"/>
                            <path d="m21 21-4.35-4.35"/>
                          </svg>
                          <span>{{ step.tool_desc || step.tool_name || '检索工具' }}</span>
                          <span v-if="step.isActive" class="tool-loading"></span>
                        </div>
                        <div v-if="step.tool_params" class="tool-params">
                          <span class="param-label">参数:</span>
                          <code class="param-value">{{ formatToolParams(step.tool_params) }}</code>
                        </div>
                      </div>
                    </div>
                  </template>

                  <!-- 分析/综合步骤 -->
                  <template v-else-if="step.type === 'analysis' || step.type === 'synthesis' || step.type === 'plan'">
                    <div class="step-header analysis">
                      <span class="step-number">步骤 {{ step.step }}</span>
                      <span class="step-type">{{ step.stage || '分析' }}</span>
                    </div>
                    <div class="step-content">
                      <div class="thought-content" v-html="renderMarkdown(step.content || '')"></div>
                    </div>
                  </template>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 输入区域 -->
      <div class="input-area">
        <div class="input-container">
          <textarea
            v-model="inputMessage"
            class="message-input"
            :placeholder="chatMode === 'agent' ? '向 Agent 提问...' : '输入消息...'"
            rows="1"
            @keydown.enter.exact="sendMessage"
            @keydown.enter.shift.prevent
            @input="adjustTextareaHeight"
            ref="textareaRef"
            :disabled="isStreaming || isAgentStreaming"
          ></textarea>
          <div class="input-actions">
            <span class="char-count">{{ inputMessage.length }} / 4000</span>
            <button
              class="send-btn"
              :disabled="!inputMessage.trim() || isStreaming || isAgentStreaming"
              @click="sendMessage"
            >
              <svg v-if="!isStreaming && !isAgentStreaming" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 2L11 13M22 2l-7 20-4-9-9 9"/>
              </svg>
              <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="6" y="4" width="4" height="16"/>
                <rect x="14" y="4" width="4" height="16"/>
              </svg>
            </button>
          </div>
        </div>
      </div>
    </main>

    <!-- RAG 设置弹窗 -->
    <div v-if="showRAGSettings" class="modal-overlay" @click.self="showRAGSettings = false">
      <div class="modal-content">
        <div class="modal-header">
          <h2>检索设置</h2>
          <button class="close-btn" @click="showRAGSettings = false">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <!-- 基础设置 -->
          <div class="setting-section">
            <h3>基础设置</h3>
            <div class="setting-row">
              <label>知识库</label>
              <select v-model="ragConfig.kb_id" class="setting-select">
                <option value="">请选择知识库</option>
                <option v-for="kb in knowledgeBases" :key="kb.id" :value="kb.id">
                  {{ kb.name }}
                </option>
              </select>
            </div>
          </div>

          <!-- 检索模式 -->
          <div class="setting-section">
            <h3>检索模式（可多选）</h3>
            <div class="setting-row">
              <label>检索方式</label>
              <div class="checkbox-group">
                <label class="checkbox-item">
                  <input
                    type="checkbox"
                    :checked="ragConfig.retrieval_modes.includes('vector')"
                    disabled
                  >
                  <span>向量检索（必选）</span>
                </label>
                <label class="checkbox-item">
                  <input
                    type="checkbox"
                    :checked="ragConfig.retrieval_modes.includes('bm25')"
                    @change="toggleRetrievalMode('bm25')"
                  >
                  <span>关键词检索</span>
                </label>
                <label class="checkbox-item">
                  <input
                    type="checkbox"
                    :checked="ragConfig.retrieval_modes.includes('graph')"
                    @change="toggleRetrievalMode('graph')"
                  >
                  <span>图谱检索</span>
                </label>
              </div>
            </div>
            <div class="setting-row">
              <label>向量 TopK</label>
              <input
                type="range"
                v-model.number="ragConfig.vector_top_k"
                min="1"
                max="30"
                class="setting-slider"
              >
              <span class="setting-value">{{ ragConfig.vector_top_k }}</span>
            </div>
            <div class="setting-row" v-if="ragConfig.retrieval_modes.includes('bm25')">
              <label>关键词 TopK</label>
              <input
                type="range"
                v-model.number="ragConfig.keyword_top_k"
                min="1"
                max="30"
                class="setting-slider"
              >
              <span class="setting-value">{{ ragConfig.keyword_top_k }}</span>
            </div>
            <div class="setting-row" v-if="ragConfig.retrieval_modes.includes('graph')">
              <label>图谱 TopK</label>
              <input
                type="range"
                v-model.number="ragConfig.graph_top_k"
                min="1"
                max="20"
                class="setting-slider"
              >
              <span class="setting-value">{{ ragConfig.graph_top_k }}</span>
            </div>
            <div class="setting-row">
              <label>相似度阈值</label>
              <input
                type="range"
                v-model.number="ragConfig.similarity_threshold"
                min="0"
                max="1"
                step="0.05"
                class="setting-slider"
              >
              <span class="setting-value">{{ ragConfig.similarity_threshold.toFixed(2) }}</span>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button class="btn-secondary" @click="showRAGSettings = false">取消</button>
          <button class="btn-primary" @click="saveRAGConfig">保存配置</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import MarkdownIt from 'markdown-it'
import { useAuthStore } from '@/stores/auth'
import { sessionApi } from '@/api/session'
import { streamChatWithAuth } from '@/api/chat/stream'
import { streamAgentChat } from '@/api/agent'
import { knowledgeApi } from '@/api/knowledge'
import {
  defaultRAGConfig,
  type RAGConfig,
  type Session,
  type Message,
  type KnowledgeBase,
  type AgentStep,
  type ChatMode
} from '@/types'
import { formatTime } from '@/utils'
import { copyToClipboard } from '@/utils/security'

const authStore = useAuthStore()
const userInitial = computed(() => authStore.username?.charAt(0).toUpperCase() || 'U')

// 聊天模式: 'normal' | 'agent'
const chatMode = ref<ChatMode>('normal')

// 状态
const sidebarCollapsed = ref(false)
const loadingSessions = ref(false)
const sessions = ref<Session[]>([])
const currentSessionId = ref<string>('')
const currentSession = ref<Session | null>(null)
const messages = ref<Message[]>([])
const inputMessage = ref('')
const messagesContainer = ref<HTMLElement>()
const textareaRef = ref<HTMLTextAreaElement>()

// Agent 模式消息列表
const agentMessages = ref<Array<{
  role: 'user' | 'assistant'
  content?: string
  answer?: string
  steps: AgentStep[]
}>>([])

// 流式状态（普通模式）
const isStreaming = ref(false)
const streamingContent = ref('')
const currentRAGContext = ref<any>(null)

// Agent 流式状态
const isAgentStreaming = ref(false)
const currentAgentSteps = ref<AgentStep[]>([])
const expandedToolOutputs = ref<Record<string, boolean>>({})

// RAG 配置
const ragConfig = ref<RAGConfig>({ ...defaultRAGConfig })
const showRAGSettings = ref(false)
const knowledgeBases = ref<KnowledgeBase[]>([])

// Markdown渲染器
const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true
})

// 渲染Markdown
function renderMarkdown(content: string): string {
  return md.render(content)
}

// 统一显示的消息列表（根据是否有 agent_steps 决定显示方式）
const displayMessages = computed(() => messages.value)

// 切换聊天模式
async function switchMode(mode: ChatMode) {
  chatMode.value = mode

  // 切换到 Agent 模式时，从 messages 构建 agentMessages
  if (mode === 'agent' && messages.value.length > 0) {
    buildAgentMessages(messages.value)
  }

}

// 切换侧边栏
function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

// 加载知识库列表
async function loadKnowledgeBases() {
  try {
    const res = await knowledgeApi.getList()
    knowledgeBases.value = (res.data as any)?.items || []
  } catch (error) {
    console.error('Failed to load knowledge bases:', error)
  }
}

// 切换 RAG
function toggleRAG() {
  ragConfig.value.enabled = !ragConfig.value.enabled
  if (ragConfig.value.enabled && !ragConfig.value.kb_id) {
    showRAGSettings.value = true
  }
  saveRAGConfig()
}

// 切换检索模式
function toggleRetrievalMode(mode: 'bm25' | 'graph') {
  const index = ragConfig.value.retrieval_modes.indexOf(mode)
  if (index >= 0) {
    ragConfig.value.retrieval_modes.splice(index, 1)
  } else {
    ragConfig.value.retrieval_modes.push(mode)
  }
}

// 保存 RAG 配置到会话
async function saveRAGConfig() {
  if (currentSessionId.value) {
    try {
      await sessionApi.update(currentSessionId.value, {
        rag_config: ragConfig.value
      })
      const session = sessions.value.find(s => s.id === currentSessionId.value)
      if (session) {
        session.rag_config = ragConfig.value
      }
    } catch (error) {
      console.error('Failed to save RAG config:', error)
    }
  }
  showRAGSettings.value = false
  ElMessage.success('配置已保存')
}

// 加载会话列表
async function loadSessions() {
  loadingSessions.value = true
  try {
    const res = await sessionApi.list({ page: 1, size: 50 })
    if (res.data) {
      sessions.value = res.data.sessions || []
      if (sessions.value.length > 0 && !currentSessionId.value) {
        await selectSession(sessions.value[0].id)
      }
    } else {
      sessions.value = []
    }
  } catch (error) {
    console.error('Failed to load sessions:', error)
    sessions.value = []
  } finally {
    loadingSessions.value = false
  }
}

// 加载消息
async function loadMessages(sessionId: string) {
  try {
    const res = await sessionApi.getDetail(sessionId)
    if (res.data) {
      const loadedMessages = res.data.messages || []
      messages.value = loadedMessages

      // 根据当前模式构建相应的消息数组
      if (chatMode.value === 'agent') {
        buildAgentMessages(loadedMessages)
      }

      currentRAGContext.value = null
      scrollToBottom()
    }
  } catch (error) {
    console.error('Failed to load messages:', error)
  }
}

// 从 messages 构建 agentMessages（用于历史消息渲染）
function buildAgentMessages(msgs: Message[]) {
  agentMessages.value = []
  let currentAssistantMsg: {
    role: 'user' | 'assistant'
    content?: string
    answer?: string
    steps: AgentStep[]
  } | null = null

  for (const msg of msgs) {
    if (msg.role === 'user') {
      // 保存之前的 assistant 消息
      if (currentAssistantMsg) {
        agentMessages.value.push(currentAssistantMsg)
        currentAssistantMsg = null
      }
      // 添加用户消息
      agentMessages.value.push({
        role: 'user',
        content: msg.content,
        steps: []
      })
    } else if (msg.role === 'assistant') {
      // 解析 agent_steps
      const steps = parseAgentSteps(msg.agent_steps)

      // 如果已经有 assistant 消息，合并步骤
      if (currentAssistantMsg) {
        currentAssistantMsg.steps = [...currentAssistantMsg.steps, ...steps]
        if (msg.content && msg.content !== '执行完成') {
          currentAssistantMsg.answer = msg.content
        }
      } else {
        currentAssistantMsg = {
          role: 'assistant',
          answer: msg.content && msg.content !== '执行完成' ? msg.content : '',
          steps: steps
        }
      }
    }
  }

  // 保存最后一个 assistant 消息
  if (currentAssistantMsg) {
    agentMessages.value.push(currentAssistantMsg)
  }
}

// 解析 agent_steps JSON 字符串为 AgentStep 数组
function parseAgentSteps(agentStepsJson: string | undefined | null): AgentStep[] {
  if (!agentStepsJson) return []
  try {
    const parsed = JSON.parse(agentStepsJson)

    // 确保 parsed 是数组
    if (!Array.isArray(parsed)) {
      // 如果是空对象 {}，返回空数组（兼容旧数据）
      if (typeof parsed === 'object' && parsed !== null && Object.keys(parsed).length === 0) {
        return []
      }
      // 如果是包含数据的对象，包装成数组（数据异常情况）
      if (typeof parsed === 'object' && parsed !== null) {
        console.warn('agent_steps is not an array, wrapping as array:', parsed)
        return [parsed].map((s: any, index: number) => ({
          id: s.id || `step_${s.step}_${index}`,
          step: s.step,
          type: s.type || 'thought',
          stage: s.stage,
          content: s.content,
          thought: s.thought,
          tool_name: s.tool_name,
          tool_desc: s.tool_desc,
          tool_params: s.tool_params,
          tool_output: s.tool_output,
          tool_id: s.tool_id,
          is_agent: s.is_agent,
          agent_name: s.agent_name,
          agent_stage: s.agent_stage,
          related_tool: s.related_tool,
          related_step: s.related_step,
          reason: s.reason,
          timestamp: 0,
          isActive: false
        }))
      }
      // 如果是其他类型，返回空数组
      console.warn('agent_steps has unexpected type:', typeof parsed)
      return []
    }

    return parsed.map((s: any, index: number) => ({
      id: s.id || `step_${s.step}_${index}`,
      step: s.step,
      type: s.type || 'thought',
      stage: s.stage,
      content: s.content,
      thought: s.thought,
      tool_name: s.tool_name,
      tool_desc: s.tool_desc,
      tool_params: s.tool_params,
      tool_output: s.tool_output,
      tool_id: s.tool_id,
      is_agent: s.is_agent,
      agent_name: s.agent_name,
      agent_stage: s.agent_stage,
      related_tool: s.related_tool,
      related_step: s.related_step,
      reason: s.reason,
      timestamp: 0,
      isActive: false
    }))
  } catch (e) {
    console.error('Failed to parse agent_steps:', e)
    return []
  }
}

// 创建新会话
async function createNewSession() {
  try {
    const res = await sessionApi.create({
      title: '新对话',
      description: ''
    })
    if (res.data) {
      sessions.value.unshift(res.data)
      await selectSession(res.data.id)
    }
  } catch (error) {
    console.error('Failed to create session:', error)
    ElMessage.error('创建会话失败')
  }
}

// 选择会话
async function selectSession(sessionId: string) {
  currentSessionId.value = sessionId
  const session = sessions.value.find(s => s.id === sessionId)
  currentSession.value = session || null
  if (session?.rag_config) {
    ragConfig.value = { ...defaultRAGConfig, ...session.rag_config }
  }
  await loadMessages(sessionId)
}

// 删除会话
async function deleteSession(sessionId: string) {
  try {
    await ElMessageBox.confirm('确认删除此会话吗？', '提示', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning'
    })
    await sessionApi.delete(sessionId)
    sessions.value = sessions.value.filter(s => s.id !== sessionId)
    if (currentSessionId.value === sessionId) {
      if (sessions.value.length > 0) {
        await selectSession(sessions.value[0].id)
      } else {
        currentSessionId.value = ''
        currentSession.value = null
        messages.value = []
        agentMessages.value = []
      }
    }
    ElMessage.success('删除成功')
  } catch (error) {
    // 用户取消
  }
}

// 发送消息（根据模式选择）
async function sendMessage() {
  const content = inputMessage.value.trim()
  if (!content || (isStreaming.value && isAgentStreaming.value)) return

  if (chatMode.value === 'agent') {
    await sendAgentMessage(content)
  } else {
    await sendNormalMessage(content)
  }
}

// 普通模式发送消息
async function sendNormalMessage(content: string) {
  if (!currentSessionId.value) {
    await createNewSession()
  }

  const userMessage: Message = {
    id: Date.now().toString(),
    session_id: currentSessionId.value,
    role: 'user',
    content,
    token_count: content.length / 3,
    created_at: new Date().toISOString()
  }
  messages.value.push(userMessage)

  inputMessage.value = ''
  adjustTextareaHeight()
  scrollToBottom()

  isStreaming.value = true
  streamingContent.value = ''
  currentRAGContext.value = null

  try {
    const sessionId = currentSessionId.value
    const request = {
      content: userMessage.content,
      session_id: sessionId,
      stream: true,
      rag_config: ragConfig.value.enabled ? ragConfig.value : undefined
    }

    for await (const event of streamChatWithAuth(request)) {
      if (event.event === 'session') {
        if (event.session_id && !currentSessionId.value) {
          currentSessionId.value = event.session_id
        }
      } else if (event.event === 'rag_context') {
        currentRAGContext.value = event.rag_context
      } else if (event.event === 'content') {
        streamingContent.value += event.content
        scrollToBottom()
      } else if (event.event === 'end') {
        const assistantMessage: Message = {
          id: event.message_id || Date.now().toString(),
          session_id: sessionId,
          role: 'assistant',
          content: streamingContent.value,
          token_count: event.token_count || 0,
          created_at: new Date().toISOString()
        }
        messages.value.push(assistantMessage)
        streamingContent.value = ''
      }
    }
  } catch (error) {
    console.error('Chat error:', error)
    ElMessage.error('发送消息失败')
  } finally {
    isStreaming.value = false
    streamingContent.value = ''
  }
}

// Agent 模式发送消息
async function sendAgentMessage(content: string) {
  // 如果没有会话，先创建
  if (!currentSessionId.value) {
    await createNewSession()
  }

  // 添加用户消息到 messages 数组（用于普通模式显示）
  const userMessage: Message = {
    id: Date.now().toString(),
    session_id: currentSessionId.value,
    role: 'user',
    content,
    token_count: content.length / 3,
    created_at: new Date().toISOString()
  }
  messages.value.push(userMessage)

  // 添加用户消息到 agentMessages 数组（用于 Agent 模式显示）
  agentMessages.value.push({
    role: 'user',
    content,
    steps: []
  })

  inputMessage.value = ''
  adjustTextareaHeight()
  scrollToBottom()

  isAgentStreaming.value = true
  currentAgentSteps.value = []

  try {
    let currentAnswer = ''
    const startTime = Date.now()

    // 传递 session_id 给后端
    for await (const event of streamAgentChat({
      query: content,
      session_id: currentSessionId.value
    })) {
      if (event.event === 'session') {
        // 更新 session_id
        if (event.session_id && !currentSessionId.value) {
          currentSessionId.value = event.session_id
        }
      } else if (event.event === 'step') {
        const step: AgentStep = {
          id: `step_${event.step}_${Date.now()}`,
          step: event.step || currentAgentSteps.value.length + 1,
          type: event.type || 'thought',
          stage: event.stage,
          content: event.content,
          thought: event.thought,
          tool_name: event.tool_name,
          tool_desc: event.tool_desc,
          tool_params: event.tool_params,
          tool_output: event.tool_output,
          tool_id: event.tool_id,
          is_agent: event.is_agent,
          agent_name: event.agent_name,
          agent_stage: event.agent_stage,
          related_tool: event.related_tool,
          related_step: event.related_step,
          timestamp: Date.now() - startTime
        }

        // 标记之前活跃的步骤为非活跃
        currentAgentSteps.value.forEach(s => s.isActive = false)
        step.isActive = true
        currentAgentSteps.value.push(step)

        scrollToBottom()
      } else if (event.event === 'done') {
        // 标记所有步骤为非活跃
        currentAgentSteps.value.forEach(s => s.isActive = false)

        // 添加完成步骤
        currentAgentSteps.value.push({
          id: `step_complete_${Date.now()}`,
          step: currentAgentSteps.value.length + 1,
          type: 'complete',
          reason: event.reason || 'Agent 完成执行',
          timestamp: Date.now() - startTime
        })

        // 获取最终答案（优先使用事件中的answer）
        if (event.answer) {
          currentAnswer = event.answer
        }

        scrollToBottom()
      } else if (event.event === 'error') {
        currentAgentSteps.value.forEach(s => s.isActive = false)
        currentAgentSteps.value.push({
          id: `step_error_${Date.now()}`,
          step: currentAgentSteps.value.length + 1,
          type: 'error',
          content: event.content,
          timestamp: Date.now() - startTime
        })
      }
    }

    // 将流式步骤保存到 agentMessages，保留思考过程的显示
    await nextTick()
    const finalSteps = [...currentAgentSteps.value].map(s => ({ ...s, isActive: false }))
    agentMessages.value.push({
      role: 'assistant',
      answer: currentAnswer,
      steps: finalSteps
    })

    // 重新加载消息以同步后端数据（保持 messages 数组正确）
    if (currentSessionId.value) {
      const res = await sessionApi.getDetail(currentSessionId.value)
      if (res.data) {
        messages.value = res.data.messages || []
        // 在 Agent 模式下，构建 agentMessages
        if (chatMode.value === 'agent') {
          buildAgentMessages(messages.value)
        }
      }
    }
  } catch (error) {
    console.error('Agent chat error:', error)
    ElMessage.error('Agent 执行失败')
    const errorStep: AgentStep = {
      id: `step_error_${Date.now()}`,
      step: currentAgentSteps.value.length + 1,
      type: 'error',
      content: error instanceof Error ? error.message : '未知错误',
      timestamp: Date.now()
    }
    currentAgentSteps.value.push(errorStep)

    // 即使出错也保存步骤
    await nextTick()
    const finalSteps = [...currentAgentSteps.value].map(s => ({ ...s, isActive: false }))
    agentMessages.value.push({
      role: 'assistant',
      answer: '',
      steps: finalSteps
    })
  } finally {
    currentAgentSteps.value = []
    isAgentStreaming.value = false
  }
}

// 格式化工具参数
function formatToolParams(params: Record<string, any> | string): string {
  if (!params) return ''
  if (typeof params === 'string') {
    try {
      const parsed = JSON.parse(params)
      params = parsed
    } catch {
      return params as string
    }
  }
  if (typeof params !== 'object' || Array.isArray(params)) {
    return String(params)
  }
  try {
    // 格式化参数，只显示主要的参数
    const keys = Object.keys(params)
    if (keys.length === 0) return ''
    return JSON.stringify(params, null, 2)
  } catch {
    return String(params)
  }
}

// 截断文本
function truncateText(text: string, maxLength: number): string {
  if (!text) return ''
  if (text.length <= maxLength) return text
  return text.slice(0, maxLength) + '...'
}

// 格式化工具输出
function formatToolOutput(output: any): string {
  if (!output) return ''
  if (typeof output === 'string') {
    try {
      const parsed = JSON.parse(output)
      return JSON.stringify(parsed, null, 2)
    } catch {
      return output
    }
  }
  try {
    return JSON.stringify(output, null, 2)
  } catch {
    return String(output)
  }
}

// 切换工具输出展开状态
function toggleToolOutput(stepId: string) {
  expandedToolOutputs.value[stepId] = !expandedToolOutputs.value[stepId]
}

// 复制消息
async function copyMessage(content: string) {
  const success = await copyToClipboard(content)
  if (success) {
    ElMessage.success('已复制')
  }
}

// 滚动到底部
function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

// 调整输入框高度
function adjustTextareaHeight() {
  nextTick(() => {
    if (textareaRef.value) {
      textareaRef.value.style.height = 'auto'
      textareaRef.value.style.height = Math.min(textareaRef.value.scrollHeight, 200) + 'px'
    }
  })
}

onMounted(() => {
  loadSessions()
  loadKnowledgeBases()
})
</script>

<style scoped>
.chat-page {
  display: flex;
  height: 100vh;
  background: #0f0f0f;
  color: #e4e4e7;
}

/* ==================== 侧边栏 ==================== */
.sidebar {
  width: 260px;
  background: #1a1a1a;
  border-right: 1px solid #272727;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
}

.sidebar.collapsed {
  width: 0;
  overflow: hidden;
}

.sidebar-header {
  padding: 16px;
  border-bottom: 1px solid #272727;
  display: flex;
  gap: 8px;
  align-items: center;
}

.new-chat-btn {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.new-chat-btn:hover {
  background: #2563eb;
}

.collapse-btn {
  padding: 8px;
  background: transparent;
  border: none;
  color: #737373;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.2s;
}

.collapse-btn:hover {
  background: #272727;
  color: #a3a3a3;
}

.sessions-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.sessions-list::-webkit-scrollbar {
  width: 6px;
}

.sessions-list::-webkit-scrollbar-track {
  background: transparent;
}

.sessions-list::-webkit-scrollbar-thumb {
  background: #3f3f3f;
  border-radius: 3px;
}

.session-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  margin-bottom: 4px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  position: relative;
}

.session-item:hover {
  background: #272727;
}

.session-item.active {
  background: #1e3a5f;
}

.session-icon {
  color: #737373;
  flex-shrink: 0;
}

.session-title {
  flex: 1;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.delete-btn {
  opacity: 0;
  padding: 4px;
  background: transparent;
  border: none;
  color: #737373;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s;
}

.session-item:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  background: #ef4444;
  color: white;
}

/* ==================== 主内容区 ==================== */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #0f0f0f;
}

.chat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-bottom: 1px solid #272727;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.mode-tabs {
  display: flex;
  gap: 4px;
  background: #1a1a1a;
  padding: 4px;
  border-radius: 10px;
}

.mode-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: transparent;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  color: #737373;
  cursor: pointer;
  transition: all 0.2s;
}

.mode-tab:hover {
  color: #a3a3a3;
  background: #272727;
}

.mode-tab.active {
  background: #3b82f6;
  color: white;
}

.mode-tab svg {
  width: 16px;
  height: 16px;
}

.header-right {
  display: flex;
  gap: 8px;
}

.agent-mode-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 8px;
  color: #60a5fa;
  font-size: 13px;
}

.rag-toggle-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: #272727;
  border: 1px solid #3f3f3f;
  border-radius: 8px;
  color: #a3a3a3;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.rag-toggle-btn:hover {
  background: #3f3f3f;
}

.rag-toggle-btn.active {
  background: #3b82f6;
  border-color: #3b82f6;
  color: white;
}

.settings-btn {
  padding: 8px;
  background: #272727;
  border: 1px solid #3f3f3f;
  border-radius: 8px;
  color: #a3a3a3a;
  cursor: pointer;
  transition: all 0.2s;
}

.settings-btn:hover {
  background: #3f3f3f;
  color: #e4e4e7;
}

/* ==================== 消息区域 ==================== */
.messages-area {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.messages-area::-webkit-scrollbar {
  width: 6px;
}

.messages-area::-webkit-scrollbar-track {
  background: transparent;
}

.messages-area::-webkit-scrollbar-thumb {
  background: #3f3f3f;
  border-radius: 3px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #737373;
}

.empty-icon {
  color: #3f3f3f;
  margin-bottom: 16px;
}

.empty-icon.agent-icon {
  color: #3b82f6;
}

.empty-state p {
  margin: 4px 0;
  font-size: 16px;
}

.empty-state .hint {
  font-size: 13px;
  color: #525252;
}

.messages-list {
  max-width: 800px;
  margin: 0 auto;
}

.rag-context-banner {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  margin-bottom: 16px;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 10px;
  color: #60a5fa;
  font-size: 13px;
}

.rag-context-banner .sources {
  color: #737373;
  margin-left: auto;
}

/* ==================== 消息 ==================== */
.message {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message.user {
  flex-direction: row-reverse;
}

.message-avatar {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #3b82f6;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
}

.message.assistant .message-avatar {
  background: #272727;
}

.message-content {
  flex: 1;
  max-width: 70%;
}

.message.user .message-content {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
}

.message-text {
  padding: 12px 16px;
  border-radius: 16px;
  font-size: 15px;
  line-height: 1.6;
  word-wrap: break-word;
}

.message.user .message-text {
  background: #3b82f6;
  color: white;
  border-bottom-right-radius: 4px;
}

.message.assistant .message-text {
  background: #1a1a1a;
  border: 1px solid #272727;
  border-bottom-left-radius: 4px;
}

.message.streaming .message-text {
  border-left: 3px solid #3b82f6;
}

/* Markdown 样式 */
.message-text :deep(p) {
  margin: 0 0 8px 0;
}

.message-text :deep(p:last-child) {
  margin-bottom: 0;
}

.message-text :deep(pre) {
  background: #0d0d0d;
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 8px 0;
}

.message-text :deep(code) {
  background: rgba(59, 130, 246, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  color: #60a5fa;
}

.message-text :deep(pre code) {
  background: transparent;
  padding: 0;
  color: #a3a3a3;
}

.message-actions {
  margin-top: 8px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.message-time {
  font-size: 12px;
  color: #737373;
}

.action-btn {
  padding: 4px 8px;
  background: transparent;
  border: none;
  color: #737373;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn:hover {
  background: #272727;
  color: #e4e4e7;
}

.streaming-indicator {
  color: #3b82f6;
  font-size: 13px;
}

/* ==================== 消息组 ==================== */
.message-group {
  margin-bottom: 24px;
}

.agent-steps {
  margin-top: 12px;
}

/* ==================== Agent 模式 ==================== */
.messages-list.agent-mode {
  max-width: 900px;
}

.agent-message {
  margin-bottom: 24px;
}

.agent-message.streaming {
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.8; }
}

.agent-response {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.agent-header {
  display: flex;
  align-items: center;
  gap: 10px;
}

.agent-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.agent-avatar.streaming {
  animation: glow 1.5s ease-in-out infinite;
}

@keyframes glow {
  0%, 100% {
    box-shadow: 0 0 5px rgba(59, 130, 246, 0.5);
  }
  50% {
    box-shadow: 0 0 20px rgba(59, 130, 246, 0.8);
  }
}

.agent-label {
  font-weight: 600;
  font-size: 14px;
  color: #a3a3a3;
}

/* 思考步骤 */
.thinking-steps {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.step-item {
  background: #1a1a1a;
  border: 1px solid #272727;
  border-radius: 12px;
  overflow: hidden;
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateX(-10px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.step-item.action {
  border-left: 3px solid #f59e0b;
}

.step-item.thought {
  border-left: 3px solid #8b5cf6;
}

.step-item.complete {
  border: 1px solid #22c55e;
  background: rgba(34, 197, 94, 0.05);
}

.step-item.error {
  border: 1px solid #ef4444;
  background: rgba(239, 68, 68, 0.05);
}

.step-item.active {
  border-style: dashed;
  animation: borderPulse 1.5s infinite;
}

@keyframes borderPulse {
  0%, 100% { border-color: #3b82f6; }
  50% { border-color: #60a5fa; }
}

.step-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: #272727;
  font-size: 12px;
}

.step-header.complete {
  background: transparent;
  color: #22c55e;
}

.step-header.error {
  background: transparent;
  color: #ef4444;
}

.step-header.search {
  background: rgba(6, 182, 212, 0.1);
  color: #06b6d4;
}

.step-header.analysis {
  background: rgba(249, 115, 22, 0.1);
  color: #f97316;
}

.step-number {
  font-weight: 600;
  color: #e4e4e7;
}

.step-type {
  color: #737373;
}

.step-time {
  margin-left: auto;
  color: #525252;
}

.step-status {
  margin-left: auto;
  color: #3b82f6;
}

.step-content {
  padding: 12px 14px;
}

.step-thought {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  color: #a3a3a3;
  font-size: 13px;
}

.step-thought svg {
  flex-shrink: 0;
  color: #737373;
}

.tool-call-card {
  background: #0f0f0f;
  border-radius: 8px;
  padding: 12px;
}

.tool-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #f59e0b;
  margin-bottom: 10px;
}

.tool-params {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
  font-size: 12px;
}

.param-label {
  color: #737373;
}

.param-value {
  background: #272727;
  padding: 4px 8px;
  border-radius: 4px;
  color: #60a5fa;
  font-family: 'Consolas', monospace;
  font-size: 11px;
}

.tool-output {
  font-size: 12px;
}

.output-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.output-label {
  color: #737373;
  margin-right: 6px;
}

.output-value {
  color: #a3a3a3;
}

.output-value.output-truncated {
  color: #a3a3a3;
  font-style: italic;
}

.output-value pre {
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #a3a3a3;
}

.toggle-btn {
  padding: 4px 8px;
  background: transparent;
  border: 1px solid #3f3f3f;
  border-radius: 4px;
  color: #737373;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-btn:hover {
  background: #272727;
  color: #a3a3a3;
  border-color: #525252;
}

.tool-loading {
  width: 14px;
  height: 14px;
  border: 2px solid #3f3f3f;
  border-top-color: #f59e0b;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-left: auto;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.thought-content {
  color: #e4e4e7;
  font-size: 14px;
  line-height: 1.6;
}

.thought-content :deep(p) {
  margin: 0 0 8px 0;
}

.thought-content :deep(p:last-child) {
  margin-bottom: 0;
}

/* ==================== Agent 特殊步骤类型样式 ==================== */
.step-stage {
  margin-left: auto;
  color: #3b82f6;
  font-size: 11px;
  padding: 2px 6px;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 4px;
}

.step-type.agent-call {
  color: #8b5cf6;
}

.step-type.search {
  color: #06b6d4;
}

.step-type.plan {
  color: #eab308;
}

.step-type.analysis {
  color: #f97316;
}

.step-type.review {
  color: #ef4444;
}

.step-type.synthesis {
  color: #22c55e;
}

.step-type.retrieval {
  color: #14b8a6;
}

.agent-call-card,
.search-card {
  background: #0f0f0f;
  border-radius: 8px;
  padding: 12px;
}

.agent-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.agent-details {
  flex: 1;
}

.agent-name {
  font-weight: 600;
  color: #8b5cf6;
  margin-bottom: 4px;
}

.search-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-name {
  font-weight: 600;
  color: #06b6d4;
}

.plan-content,
.analysis-content,
.review-content,
.synthesis-content,
.retrieval-content,
.default-content {
  color: #e4e4e7;
  font-size: 14px;
  line-height: 1.6;
}

.plan-content :deep(p),
.analysis-content :deep(p),
.review-content :deep(p),
.synthesis-content :deep(p),
.retrieval-content :deep(p),
.default-content :deep(p) {
  margin: 0 0 8px 0;
}

.plan-content :deep(p:last-child),
.analysis-content :deep(p:last-child),
.review-content :deep(p:last-child),
.synthesis-content :deep(p:last-child),
.retrieval-content :deep(p:last-child),
.default-content :deep(p:last-child) {
  margin-bottom: 0;
}

.related-tool {
  color: #737373;
  font-size: 11px;
  font-style: italic;
}

/* 最终答案 */
.agent-answer {
  background: #1a1a1a;
  border: 1px solid #272727;
  border-radius: 12px;
  padding: 16px;
  margin-top: 8px;
}

.answer-label {
  font-size: 12px;
  font-weight: 600;
  color: #737373;
  margin-bottom: 10px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.answer-content {
  color: #e4e4e7;
  font-size: 14px;
  line-height: 1.7;
}

.answer-content :deep(p) {
  margin: 0 0 8px 0;
}

.answer-content :deep(p:last-child) {
  margin-bottom: 0;
}

.answer-content :deep(pre) {
  background: #0d0d0d;
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 8px 0;
}

.answer-content :deep(code) {
  background: rgba(59, 130, 246, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  color: #60a5fa;
}

.answer-content :deep(pre code) {
  background: transparent;
  padding: 0;
  color: #a3a3a3;
}

/* ==================== 输入区域 ==================== */
.input-area {
  padding: 16px 24px;
  border-top: 1px solid #272727;
}

.input-container {
  max-width: 800px;
  margin: 0 auto;
  background: #1a1a1a;
  border: 1px solid #3f3f3f;
  border-radius: 16px;
  padding: 12px 16px;
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.message-input {
  flex: 1;
  background: transparent;
  border: none;
  color: #e4e4e7;
  font-size: 15px;
  line-height: 1.5;
  resize: none;
  outline: none;
  font-family: inherit;
}

.message-input::placeholder {
  color: #737373;
}

.message-input:disabled {
  opacity: 0.5;
}

.input-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.char-count {
  font-size: 12px;
  color: #737373;
}

.send-btn {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #3b82f6;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}

.send-btn:hover:not(:disabled) {
  background: #2563eb;
}

.send-btn:disabled {
  background: #272727;
  cursor: not-allowed;
}

/* ==================== 弹窗 ==================== */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #1a1a1a;
  border: 1px solid #3f3f3f;
  border-radius: 16px;
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #272727;
}

.modal-header h2 {
  font-size: 18px;
  font-weight: 600;
  margin: 0;
}

.close-btn {
  padding: 4px;
  background: transparent;
  border: none;
  color: #737373;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #272727;
  color: #e4e4e7;
}

.modal-body {
  padding: 20px;
  overflow-y: auto;
  max-height: 60vh;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid #272727;
}

.setting-section {
  margin-bottom: 24px;
}

.setting-section:last-child {
  margin-bottom: 0;
}

.setting-section h3 {
  font-size: 14px;
  font-weight: 600;
  color: #a3a3a3;
  margin: 0 0 12px 0;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.setting-row:last-child {
  margin-bottom: 0;
}

.setting-row > label {
  font-size: 14px;
  color: #a3a3a3;
}

.setting-select {
  padding: 8px 12px;
  background: #0f0f0f;
  border: 1px solid #3f3f3f;
  border-radius: 6px;
  color: #e4e4e7;
  font-size: 14px;
  cursor: pointer;
}

.setting-select:focus {
  outline: none;
  border-color: #3b82f6;
}

.setting-slider {
  flex: 1;
  margin: 0 12px;
  -webkit-appearance: none;
  background: transparent;
}

.setting-slider::-webkit-slider-runnable-track {
  height: 4px;
  background: #3f3f3f;
  border-radius: 2px;
}

.setting-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 16px;
  height: 16px;
  background: #3b82f6;
  border-radius: 50%;
  cursor: pointer;
  margin-top: -6px;
}

.setting-value {
  min-width: 32px;
  text-align: right;
  font-size: 13px;
  color: #a3a3a3;
}

.checkbox-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.checkbox-item {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  padding: 8px 12px;
  background: #1a1a1a;
  border-radius: 6px;
  transition: background 0.2s;
}

.checkbox-item:hover {
  background: #272727;
}

.checkbox-item input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.checkbox-item input[type="checkbox"]:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.checkbox-item span {
  font-size: 14px;
  color: #e4e4e7;
}

.btn-primary, .btn-secondary {
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #3b82f6;
  border: none;
  color: white;
}

.btn-primary:hover {
  background: #2563eb;
}

.btn-secondary {
  background: transparent;
  border: 1px solid #3f3f3f;
  color: #a3a3a3;
}

.btn-secondary:hover {
  background: #272727;
  color: #e4e4e7;
}
</style>
