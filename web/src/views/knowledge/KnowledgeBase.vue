<template>
  <div class="knowledge-detail-container">
    <el-page-header @back="goBack" :content="knowledgeBase?.name || '知识库详情'" />

    <el-tabs v-model="activeTab" class="detail-tabs" @tab-change="handleTabChange">
      <!-- 概览 -->
      <el-tab-pane label="概览" name="overview">
        <div class="overview-section" v-loading="statsLoading">
          <el-row :gutter="20" v-if="stats">
            <el-col :span="6">
              <el-statistic title="文档数量" :value="stats.knowledge_count">
                <template #suffix>个</template>
              </el-statistic>
            </el-col>
            <el-col :span="6">
              <el-statistic title="分块数量" :value="stats.chunk_count">
                <template #suffix>个</template>
              </el-statistic>
            </el-col>
            <el-col :span="6">
              <div class="statistic-item">
                <div class="statistic-title">总存储</div>
                <div class="statistic-value">{{ formatFileSize(stats.total_size) }}</div>
              </div>
            </el-col>
            <el-col :span="6">
              <el-statistic title="状态">
                <template #default>
                  <el-tag :type="knowledgeBase?.status === 1 ? 'success' : 'info'">
                    {{ knowledgeBase?.status === 1 ? '启用' : '禁用' }}
                  </el-tag>
                </template>
              </el-statistic>
            </el-col>
          </el-row>

          <el-divider />

          <el-descriptions :column="2" border>
            <el-descriptions-item label="知识库名称">{{ knowledgeBase?.name }}</el-descriptions-item>
            <el-descriptions-item label="类型">{{ knowledgeBase?.type || '通用' }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatDateTime(knowledgeBase?.created_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="更新时间">
              {{ formatDateTime(knowledgeBase?.updated_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="描述" :span="2">
              {{ knowledgeBase?.description || '-' }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </el-tab-pane>

      <!-- 文档管理 -->
      <el-tab-pane label="文档" name="documents">
        <div class="documents-section">
          <div class="toolbar">
            <el-upload
              :auto-upload="false"
              :show-file-list="false"
              :on-change="handleFileSelect"
              accept=".txt,.md,.pdf,.doc,.docx"
            >
              <el-button type="primary" :icon="Upload">上传文档</el-button>
            </el-upload>
            <el-text class="hint">支持 txt, md, pdf, doc, docx 格式</el-text>
          </div>

          <el-divider />

          <el-table
            :data="knowledges"
            v-loading="knowledgesLoading"
            stripe
            style="width: 100%"
          >
            <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
            <el-table-column prop="file_type" label="类型" width="100" />
            <el-table-column prop="file_size" label="大小" width="120">
              <template #default="{ row }">
                {{ formatFileSize(row.file_size) }}
              </template>
            </el-table-column>
            <el-table-column prop="chunk_count" label="分块数" width="100">
              <template #default="{ row }">
                {{ row.chunk_count || 0 }}
              </template>
            </el-table-column>
            <el-table-column prop="parse_status" label="处理状态" width="120">
              <template #default="{ row }">
                <el-tag :type="getParseStatusType(row.parse_status)">
                  {{ getParseStatusText(row.parse_status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="上传时间" width="180">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="{ row }">
                <el-button
                  link
                  type="primary"
                  @click="viewKnowledgeChunks(row)"
                  :disabled="row.chunk_count === 0"
                >
                  查看分块
                </el-button>
                <el-button link type="danger" @click="deleteKnowledge(row.id)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-empty v-if="!knowledgesLoading && knowledges.length === 0" description="暂无文档" />
        </div>
      </el-tab-pane>

      <!-- 知识检索 -->
      <el-tab-pane label="检索" name="search">
        <div class="search-section">
          <el-input
            v-model="searchQuery"
            placeholder="输入要搜索的内容..."
            class="search-input"
            @keyup.enter="handleSearch"
          >
            <template #append>
              <el-button :icon="Search" @click="handleSearch" :loading="searchLoading">
                搜索
              </el-button>
            </template>
          </el-input>

          <div class="search-options">
            <el-text>Top K:</el-text>
            <el-input-number v-model="searchTopK" :min="1" :max="20" size="small" />
            <el-text>相似度阈值:</el-text>
            <el-input-number
              v-model="searchThreshold"
              :min="0"
              :max="1"
              :step="0.1"
              :precision="2"
              size="small"
            />
          </div>

          <el-divider />

          <div v-loading="searchLoading" class="search-results">
            <div v-if="searchResults.length > 0">
              <div
                v-for="(result, index) in searchResults"
                :key="result.chunk_id"
                class="search-result-item"
              >
                <div class="result-header">
                  <el-text type="primary" tag="b">结果 {{ index + 1 }}</el-text>
                  <el-tag size="small">相似度: {{ (result.score * 100).toFixed(1) }}%</el-tag>
                </div>
                <div class="result-content">
                  <el-text>{{ result.content }}</el-text>
                </div>
                <div class="result-footer">
                  <el-text size="small" type="info">来源: {{ result.knowledge_title }}</el-text>
                  <el-tag
                    v-for="tag in result.tags"
                    :key="tag"
                    size="small"
                    style="margin-left: 8px"
                  >
                    {{ tag }}
                  </el-tag>
                </div>
              </div>
            </div>
            <el-empty
              v-else-if="!searchLoading && hasSearched"
              description="未找到相关内容"
            />
            <el-empty v-else description="输入关键词搜索知识库内容" />
          </div>
        </div>
      </el-tab-pane>

      <!-- 分块列表 -->
      <el-tab-pane label="分块" name="chunks">
        <div class="chunks-section">
          <div class="toolbar">
            <el-select
              v-model="selectedKnowledgeId"
              placeholder="选择文档"
              clearable
              @change="loadChunks"
              style="width: 300px"
            >
              <el-option
                v-for="kb in knowledges"
                :key="kb.id"
                :label="kb.title"
                :value="kb.id"
              />
            </el-select>
          </div>

          <el-divider />

          <el-table :data="chunks" v-loading="chunksLoading" stripe style="width: 100%">
            <el-table-column prop="chunk_index" label="序号" width="80" />
            <el-table-column prop="content" label="内容" min-width="400" show-overflow-tooltip />
            <el-table-column prop="token_count" label="Token数" width="100" />
            <el-table-column label="标签" width="200">
              <template #default="{ row }">
                <el-tag v-for="tag in row.tags" :key="tag" size="small" style="margin-right: 4px">
                  {{ tag }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button link type="primary" @click="viewChunkDetail(row)">
                  详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-empty v-if="!chunksLoading && chunks.length === 0" description="暂无分块数据" />
        </div>
      </el-tab-pane>

      <!-- 设置 -->
      <el-tab-pane label="设置" name="settings">
        <div class="settings-section" v-loading="settingsLoading">
          <el-form :model="settingsForm" label-width="120px" style="max-width: 600px">
            <el-form-item label="知识库名称">
              <el-input v-model="settingsForm.name" placeholder="请输入知识库名称" />
            </el-form-item>
            <el-form-item label="描述">
              <el-input
                v-model="settingsForm.description"
                type="textarea"
                :rows="3"
                placeholder="请输入描述"
              />
            </el-form-item>
            <el-form-item label="状态">
              <el-switch
                v-model="settingsForm.status"
                :active-value="1"
                :inactive-value="0"
                active-text="启用"
                inactive-text="禁用"
              />
            </el-form-item>

            <el-divider content-position="left">策略配置</el-divider>

            <el-form-item label="向量模型">
              <el-select v-model="settingsForm.embedding_model" style="width: 100%">
                <el-option label="BGE-M3" value="bge-m3" />
                <el-option label="BGE-Large" value="bge-large" />
                <el-option label="Text2Vec" value="text2vec" />
              </el-select>
            </el-form-item>

            <el-form-item label="分块大小">
              <el-input-number
                v-model="settingsForm.chunk_size"
                :min="128"
                :max="2048"
                :step="64"
              />
            </el-form-item>

            <el-form-item label="分块重叠">
              <el-input-number
                v-model="settingsForm.chunk_overlap"
                :min="0"
                :max="512"
                :step="32"
              />
            </el-form-item>

            <el-form-item label="启用图谱">
              <el-switch v-model="settingsForm.enable_graph" />
            </el-form-item>

            <el-form-item label="启用标签">
              <el-switch v-model="settingsForm.enable_tag" />
            </el-form-item>

            <el-form-item>
              <el-button type="primary" @click="saveSettings" :loading="settingsSaving">
                保存设置
              </el-button>
            </el-form-item>
          </el-form>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 上传文档对话框 -->
    <el-dialog v-model="showUploadDialog" title="上传文档" width="500px">
      <el-form :model="uploadForm" label-width="100px">
        <el-form-item label="文件">
          <el-text>{{ uploadForm.file?.name }}</el-text>
        </el-form-item>
        <el-form-item label="标题">
          <el-input v-model="uploadForm.title" placeholder="不填写则使用文件名" />
        </el-form-item>
        <el-form-item label="分块大小">
          <el-input-number
            v-model="uploadForm.chunk_size"
            :min="128"
            :max="2048"
            :step="64"
          />
        </el-form-item>
        <el-form-item label="分块重叠">
          <el-input-number
            v-model="uploadForm.chunk_overlap"
            :min="0"
            :max="512"
            :step="32"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="primary" @click="uploadFile" :loading="uploading">
          上传
        </el-button>
      </template>
    </el-dialog>

    <!-- 分块详情对话框 -->
    <el-dialog v-model="showChunkDialog" title="分块详情" width="700px">
      <el-descriptions :column="2" border v-if="currentChunk">
        <el-descriptions-item label="序号">{{ currentChunk.chunk_index }}</el-descriptions-item>
        <el-descriptions-item label="Token数">{{ currentChunk.token_count }}</el-descriptions-item>
        <el-descriptions-item label="标签" :span="2">
          <el-tag v-for="tag in currentChunk.tags" :key="tag" style="margin-right: 4px">
            {{ tag }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="内容" :span="2">
          <el-text style="white-space: pre-wrap">{{ currentChunk.content }}</el-text>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import axios from 'axios'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Upload, Search } from '@element-plus/icons-vue'
import { knowledgeApi } from '@/api/knowledge'
import type {
  KnowledgeBase,
  KnowledgeBaseStats,
  Knowledge,
  Chunk,
  UpdateKnowledgeBaseRequest
} from '@/types'

const router = useRouter()
const route = useRoute()

const kbId = ref<string>(route.params.id as string)
const activeTab = ref('overview')
const knowledgeBase = ref<KnowledgeBase | null>(null)

// 概览相关
const statsLoading = ref(false)
const stats = ref<KnowledgeBaseStats | null>(null)

// 文档相关
const knowledgesLoading = ref(false)
const knowledges = ref<Knowledge[]>([])
const showUploadDialog = ref(false)
const uploading = ref(false)
const uploadForm = reactive({
  file: null as File | null,
  title: '',
  chunk_size: 512,
  chunk_overlap: 100
})
const statusPolling = ref<Record<string, NodeJS.Timeout>>({})

// 检索相关
const searchQuery = ref('')
const searchTopK = ref(5)
const searchThreshold = ref(0.6)
const searchLoading = ref(false)
const searchResults = ref<any[]>([])
const hasSearched = ref(false)

// 分块相关
const selectedKnowledgeId = ref<string>('')
const chunksLoading = ref(false)
const chunks = ref<Chunk[]>([])
const showChunkDialog = ref(false)
const currentChunk = ref<Chunk | null>(null)

// 设置相关
const settingsLoading = ref(false)
const settingsSaving = ref(false)
const settingsForm = reactive<UpdateKnowledgeBaseRequest>({
  name: '',
  description: '',
  status: 1,
  embedding_model: 'bge-m3',
  chunk_size: 512,
  chunk_overlap: 100,
  enable_graph: true,
  enable_tag: true
})

function formatDateTime(date?: string) {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

function formatFileSize(bytes: number) {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return (bytes / Math.pow(k, i)).toFixed(2) + ' ' + sizes[i]
}

function getParseStatusType(status: string) {
  const map: Record<string, any> = {
    unprocessed: 'info',
    processing: 'warning',
    completed: 'success',
    failed: 'danger'
  }
  return map[status] || 'info'
}

function getParseStatusText(status: string) {
  const map: Record<string, string> = {
    unprocessed: '未处理',
    processing: '处理中',
    completed: '已完成',
    failed: '失败'
  }
  return map[status] || status
}

function goBack() {
  router.push('/knowledge')
}

// 加载知识库详情
async function loadKnowledgeBase() {
  try {
    const res = await knowledgeApi.getDetail(kbId.value)
    if (res.data) {
      knowledgeBase.value = res.data
      // 初始化设置表单
      settingsForm.name = res.data.name
      settingsForm.description = res.data.description || ''
      settingsForm.status = res.data.status
    }
  } catch (error) {
    console.error('Failed to load knowledge base:', error)
  }
}

// 加载统计信息
async function loadStats() {
  statsLoading.value = true
  try {
    const res = await knowledgeApi.getStats(kbId.value)
    if (res.data) {
      stats.value = res.data
    }
  } catch (error) {
    console.error('Failed to load stats:', error)
  } finally {
    statsLoading.value = false
  }
}

// 加载文档列表
async function loadKnowledges() {
  knowledgesLoading.value = true
  try {
    const res = await knowledgeApi.getKnowledgeList(kbId.value)
    if (res.data) {
      knowledges.value = res.data
      // 为处理中的文档启动状态轮询
      knowledges.value.forEach((kb) => {
        if (kb.parse_status === 'processing' || kb.parse_status === 'pending') {
          startStatusPolling(kb.id)
        }
      })
    }
  } catch (error) {
    console.error('Failed to load knowledges:', error)
  } finally {
    knowledgesLoading.value = false
  }
}

// 启动状态轮询
function startStatusPolling(knowledgeId: string) {
  // 清除旧的轮询
  if (statusPolling.value[knowledgeId]) {
    clearInterval(statusPolling.value[knowledgeId])
  }

  statusPolling.value[knowledgeId] = setInterval(async () => {
    try {
      const res = await knowledgeApi.getKnowledgeStatus(kbId.value, knowledgeId)
      if (res.data) {
        const status = res.data
        // 如果处理完成或失败，停止轮询并刷新列表
        if (status.status === 'completed' || status.status === 'failed') {
          stopStatusPolling(knowledgeId)
          await loadKnowledges()
        }
      }
    } catch (error) {
      console.error('Failed to poll status:', error)
      stopStatusPolling(knowledgeId)
    }
  }, 3000) // 每3秒轮询一次
}

// 停止状态轮询
function stopStatusPolling(knowledgeId: string) {
  if (statusPolling.value[knowledgeId]) {
    clearInterval(statusPolling.value[knowledgeId])
    delete statusPolling.value[knowledgeId]
  }
}

// 处理文件选择
function handleFileSelect(file: any) {
  uploadForm.file = file.raw
  uploadForm.title = file.name
  showUploadDialog.value = true
}

// 上传文件
async function uploadFile() {
  if (!uploadForm.file) {
    ElMessage.warning('请选择文件')
    return
  }

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', uploadForm.file)
    if (uploadForm.title) {
      formData.append('title', uploadForm.title)
    }
    formData.append('chunk_size', uploadForm.chunk_size.toString())
    formData.append('chunk_overlap', uploadForm.chunk_overlap.toString())

    const res = await knowledgeApi.uploadFile(kbId.value, formData)
    if (res.data) {
      ElMessage.success('文件上传成功，正在处理中...')
      showUploadDialog.value = false
      // 启动状态轮询
      startStatusPolling(res.data.knowledge_id)
      // 刷新列表
      await loadKnowledges()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '上传失败')
  } finally {
    uploading.value = false
  }
}

// 删除文档
async function deleteKnowledge(id: string) {
  try {
    await ElMessageBox.confirm('确定要删除这个文档吗？此操作不可恢复。', '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await knowledgeApi.deleteKnowledge(kbId.value, id)
    ElMessage.success('文档删除成功')
    await loadKnowledges()
    await loadStats()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

// 查看文档分块
function viewKnowledgeChunks(knowledge: Knowledge) {
  selectedKnowledgeId.value = knowledge.id
  activeTab.value = 'chunks'
  loadChunks()
}

// 搜索知识
async function handleSearch() {
  if (!searchQuery.value.trim()) {
    ElMessage.warning('请输入搜索内容')
    return
  }

  searchLoading.value = true
  hasSearched.value = true
  try {
    const res = await knowledgeApi.search({
      query: searchQuery.value,
      kb_ids: [kbId.value],
      top_k: searchTopK.value,
      score_threshold: searchThreshold.value,
      include_graph: false
    })
    if (res.data) {
      searchResults.value = res.data.results || []
      if (searchResults.value.length === 0) {
        ElMessage.info('未找到相关内容')
      }
    }
  } catch (error: any) {
    ElMessage.error(error.message || '搜索失败')
  } finally {
    searchLoading.value = false
  }
}

// 加载分块
async function loadChunks() {
  if (!selectedKnowledgeId.value) {
    chunks.value = []
    return
  }

  chunksLoading.value = true
  try {
    const res = await knowledgeApi.getChunks(kbId.value, {
      knowledge_id: selectedKnowledgeId.value
    })
    if (res.data) {
      chunks.value = res.data.items || []
    }
  } catch (error) {
    console.error('Failed to load chunks:', error)
  } finally {
    chunksLoading.value = false
  }
}

// 查看分块详情
function viewChunkDetail(chunk: Chunk) {
  currentChunk.value = chunk
  showChunkDialog.value = true
}

// 保存设置
async function saveSettings() {
  if (!settingsForm.name) {
    ElMessage.warning('请输入知识库名称')
    return
  }

  settingsSaving.value = true
  try {
    await knowledgeApi.update(kbId.value, settingsForm)
    ElMessage.success('设置保存成功')
    await loadKnowledgeBase()
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    settingsSaving.value = false
  }
}

// 标签切换处理
function handleTabChange(tabName: string) {
  if (tabName === 'overview' && !stats.value) {
    loadStats()
  } else if (tabName === 'documents' && knowledges.value.length === 0) {
    loadKnowledges()
  }
}

onMounted(async () => {
  await loadKnowledgeBase()
  await loadStats()
})

onUnmounted(() => {
  // 清理所有轮询
  Object.keys(statusPolling.value).forEach((key) => {
    stopStatusPolling(key)
  })
})
</script>

<style scoped>
.knowledge-detail-container {
  padding: 24px;
  background: white;
  border-radius: 8px;
}

.detail-tabs {
  margin-top: 24px;
}

.overview-section,
.documents-section,
.search-section,
.chunks-section,
.settings-section {
  padding: 16px 0;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
}

.hint {
  font-size: 12px;
  color: #909399;
}

.search-input {
  margin-bottom: 16px;
}

.search-options {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.search-results {
  min-height: 200px;
}

.search-result-item {
  padding: 16px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  margin-bottom: 12px;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.result-content {
  margin-bottom: 12px;
  line-height: 1.6;
}

.result-footer {
  display: flex;
  align-items: center;
}

.statistic-item {
  text-align: center;
}

.statistic-title {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.statistic-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}
</style>
