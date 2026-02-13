<template>
  <div class="chat-container">
    <!-- 会话列表 -->
    <div class="session-list" :class="{ collapsed: sessionListCollapsed }">
      <div class="session-header">
        <el-button
          type="primary"
          :icon="Plus"
          @click="createNewSession"
          size="small"
        >
          {{ t('chat.newChat') }}
        </el-button>
        <el-button
          link
          :icon="sessionListCollapsed ? DArrowRight : DArrowLeft"
          @click="toggleSessionList"
        />
      </div>
      <div class="sessions" v-loading="loadingSessions">
        <div
          v-for="session in sessions"
          :key="session.id"
          class="session-item"
          :class="{ active: currentSessionId === session.id }"
          @click="selectSession(session.id)"
        >
          <el-icon class="session-icon"><ChatDotRound /></el-icon>
          <span class="session-title">{{ session.title }}</span>
          <el-button
            link
            :icon="Delete"
            size="small"
            class="delete-btn"
            @click.stop="deleteSession(session.id)"
          />
        </div>
      </div>
    </div>

    <!-- 聊天主区域 -->
    <div class="chat-main">
      <!-- 消息列表 -->
      <div class="messages-container" ref="messagesContainer">
        <div v-if="messages.length === 0" class="empty-state">
          <el-icon :size="64" color="#dcdfe6"><ChatLineRound /></el-icon>
          <p>{{ t('chat.emptyChat') }}</p>
        </div>
        <div v-else class="messages">
          <div
            v-for="message in messages"
            :key="message.id"
            class="message"
            :class="message.role"
          >
            <div class="message-avatar">
              <el-avatar v-if="message.role === 'user'" :size="36">
                {{ authStore.username.charAt(0).toUpperCase() }}
              </el-avatar>
              <el-icon v-else :size="36" color="#409eff"><Service /></el-icon>
            </div>
            <div class="message-content">
              <div class="message-text" v-html="renderMarkdown(message.content)"></div>
              <div class="message-meta">
                <span class="message-time">{{ formatTime(message.created_at) }}</span>
                <el-button
                  link
                  :icon="CopyDocument"
                  size="small"
                  @click="copyMessage(message.content)"
                >
                  {{ t('common.copy') }}
                </el-button>
              </div>
            </div>
          </div>
          <!-- 流式输出中的消息 -->
          <div v-if="streamingContent" class="message assistant streaming">
            <div class="message-avatar">
              <el-icon :size="36" color="#409eff"><Service /></el-icon>
            </div>
            <div class="message-content">
              <div class="message-text" v-html="renderMarkdown(streamingContent)"></div>
              <div class="message-meta">
                <span class="streaming-indicator">{{ t('chat.generating') }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 输入框 -->
      <div class="input-area">
        <div class="input-wrapper">
          <el-input
            v-model="inputMessage"
            type="textarea"
            :placeholder="t('chat.inputPlaceholder')"
            :rows="1"
            :autosize="{ minRows: 1, maxRows: 6 }"
            @keydown.enter.exact="sendMessage"
            @keydown.enter.shift.prevent
            :disabled="isStreaming"
          />
          <el-button
            type="primary"
            :icon="isStreaming ? VideoPause : Position"
            circle
            @click="isStreaming ? stopStreaming : sendMessage"
            :loading="isStreaming"
            :disabled="!inputMessage.trim()"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  DArrowLeft,
  DArrowRight,
  Delete,
  ChatDotRound,
  ChatLineRound,
  Service,
  CopyDocument,
  Position,
  VideoPause
} from '@element-plus/icons-vue'
import MarkdownIt from 'markdown-it'
import { useAuthStore } from '@/stores/auth'
import { sessionApi } from '@/api/session'
import { streamChatWithAuth } from '@/api/chat/stream'
import type { Session, Message } from '@/types'
import { formatTime } from '@/utils'
import { copyToClipboard } from '@/utils/security'

const { t } = useI18n()
const authStore = useAuthStore()

// 状态
const sessionListCollapsed = ref(false)
const loadingSessions = ref(false)
const sessions = ref<Session[]>([])
const currentSessionId = ref<string>('')
const messages = ref<Message[]>([])
const inputMessage = ref('')
const isStreaming = ref(false)
const streamingContent = ref('')
const messagesContainer = ref<HTMLElement>()

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

// 切换会话列表
function toggleSessionList() {
  sessionListCollapsed.value = !sessionListCollapsed.value
}

// 加载会话列表
async function loadSessions() {
  loadingSessions.value = true
  try {
    const res = await sessionApi.list({ page: 1, size: 50 })
    if (res.data) {
      // 确保 items 存在且是数组
      sessions.value = res.data.items || []
      if (sessions.value.length > 0 && !currentSessionId.value) {
        currentSessionId.value = sessions.value[0].id
        await loadMessages(sessions.value[0].id)
      }
    } else {
      // 如果没有 data，初始化为空数组
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
      messages.value = res.data.messages
      scrollToBottom()
    }
  } catch (error) {
    console.error('Failed to load messages:', error)
  }
}

// 创建新会话
async function createNewSession() {
  try {
    const res = await sessionApi.create({
      title: '新对话',
      description: '',
      max_rounds: 50
    })
    if (res.data) {
      // 确保 sessions 是数组
      if (!Array.isArray(sessions.value)) {
        sessions.value = []
      }
      sessions.value.unshift(res.data)
      currentSessionId.value = res.data.id
      messages.value = []
      inputMessage.value = ''
    }
  } catch (error) {
    console.error('Failed to create session:', error)
    ElMessage.error('创建会话失败')
  }
}

// 选择会话
async function selectSession(sessionId: string) {
  currentSessionId.value = sessionId
  await loadMessages(sessionId)
}

// 删除会话
async function deleteSession(sessionId: string) {
  try {
    await ElMessageBox.confirm(t('chat.deleteConfirm'), '提示', {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })

    await sessionApi.delete(sessionId)
    sessions.value = sessions.value.filter(s => s.id !== sessionId)

    if (currentSessionId.value === sessionId) {
      if (sessions.value.length > 0) {
        currentSessionId.value = sessions.value[0].id
        await loadMessages(currentSessionId.value)
      } else {
        currentSessionId.value = ''
        messages.value = []
      }
    }

    ElMessage.success(t('common.success'))
  } catch (error) {
    // 用户取消
  }
}

// 发送消息
async function sendMessage() {
  const content = inputMessage.value.trim()
  if (!content || isStreaming.value) return

  // 如果没有会话，先创建
  if (!currentSessionId.value) {
    await createNewSession()
  }

  // 添加用户消息
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
  scrollToBottom()

  // 开始流式聊天
  isStreaming.value = true
  streamingContent.value = ''

  try {
    const sessionId = currentSessionId.value
    for await (const event of streamChatWithAuth({
      content: userMessage.content,
      session_id: sessionId,
      stream: true
    })) {
      if (event.event === 'content') {
        streamingContent.value += event.content
        scrollToBottom()
      } else if (event.event === 'end') {
        // 添加助手消息
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
    ElMessage.error(t('chat.error'))
  } finally {
    isStreaming.value = false
    streamingContent.value = ''
  }
}

// 停止流式输出
function stopStreaming() {
  isStreaming.value = false
  if (streamingContent.value) {
    const assistantMessage: Message = {
      id: Date.now().toString(),
      session_id: currentSessionId.value,
      role: 'assistant',
      content: streamingContent.value,
      token_count: 0,
      created_at: new Date().toISOString()
    }
    messages.value.push(assistantMessage)
    streamingContent.value = ''
  }
}

// 复制消息
async function copyMessage(content: string) {
  const success = await copyToClipboard(content)
  if (success) {
    ElMessage.success(t('common.success'))
  } else {
    ElMessage.error(t('common.error'))
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

onMounted(() => {
  loadSessions()
})
</script>

<style scoped>
.chat-container {
  display: flex;
  height: 100%;
  background: #f5f7fa;
}

.session-list {
  width: 260px;
  background: white;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  transition: width 0.3s;
}

.session-list.collapsed {
  width: 0;
  overflow: hidden;
}

.session-header {
  padding: 16px;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  gap: 8px;
  align-items: center;
}

.sessions {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.session-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  margin-bottom: 4px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
}

.session-item:hover {
  background: #f5f7fa;
}

.session-item.active {
  background: #e6f7ff;
}

.session-icon {
  color: #909399;
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
  transition: opacity 0.3s;
}

.session-item:hover .delete-btn {
  opacity: 1;
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: white;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #909399;
}

.empty-state p {
  margin-top: 16px;
  font-size: 16px;
}

.messages {
  max-width: 800px;
  margin: 0 auto;
}

.message {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
}

.message.user {
  flex-direction: row-reverse;
}

.message-avatar {
  flex-shrink: 0;
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
  border-radius: 12px;
  font-size: 14px;
  line-height: 1.6;
}

.message.user .message-text {
  background: #409eff;
  color: white;
}

.message.assistant .message-text {
  background: #f5f7fa;
  color: #303133;
}

.message-text :deep(p) {
  margin: 0 0 8px 0;
}

.message-text :deep(p:last-child) {
  margin-bottom: 0;
}

.message-text :deep(pre) {
  background: #2d2d2d;
  color: #ccc;
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
}

.message-text :deep(code) {
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
}

.message-text :deep(pre code) {
  background: transparent;
  padding: 0;
}

.message.meta {
  margin-top: 8px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.message-time {
  font-size: 12px;
  color: #909399;
}

.streaming-indicator {
  font-size: 12px;
  color: #409eff;
  display: flex;
  align-items: center;
  gap: 4px;
}

.streaming-indicator::after {
  content: '...';
  animation: dots 1.5s infinite;
}

@keyframes dots {
  0%, 20% { content: '.'; }
  40% { content: '..'; }
  60%, 100% { content: '...'; }
}

.input-area {
  padding: 16px 24px;
  border-top: 1px solid #e4e7ed;
}

.input-wrapper {
  max-width: 800px;
  margin: 0 auto;
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.input-wrapper :deep(.el-textarea) {
  flex: 1;
}

.message.streaming .message-text {
  border-left: 3px solid #409eff;
}
</style>
