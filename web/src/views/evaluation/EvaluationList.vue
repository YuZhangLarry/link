<template>
  <div class="evaluation-list-container">
    <div class="header">
      <h2>大模型测评</h2>
      <el-button type="primary" :icon="Plus" @click="showCreateDialog = true">
        创建测评任务
      </el-button>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-value">{{ stats.total }}</div>
        <div class="stat-label">总任务数</div>
      </div>
      <div class="stat-card">
        <div class="stat-value success">{{ stats.success }}</div>
        <div class="stat-label">已完成</div>
      </div>
      <div class="stat-card">
        <div class="stat-value warning">{{ stats.running }}</div>
        <div class="stat-label">执行中</div>
      </div>
      <div class="stat-card">
        <div class="stat-value danger">{{ stats.failed }}</div>
        <div class="stat-label">失败</div>
      </div>
    </div>

    <!-- 任务列表 -->
    <el-table
      :data="evaluations"
      v-loading="loading"
      stripe
      style="width: 100%"
      @row-click="viewDetail"
    >
      <el-table-column prop="id" label="任务ID" width="200" show-overflow-tooltip />
      <el-table-column prop="dataset_id" label="数据集" width="150" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="EvaluationStatusType[row.status as EvaluationStatus]">
            {{ EvaluationStatusText[row.status as EvaluationStatus] }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="进度" width="200">
        <template #default="{ row }">
          <el-progress
            :percentage="getProgress(row)"
            :status="row.status === EvaluationStatus.Failed ? 'exception' : (row.status === EvaluationStatus.Success ? 'success' : '')"
          />
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click.stop="viewDetail(row)">查看</el-button>
          <el-button
            v-if="row.status === EvaluationStatus.Running"
            link
            type="primary"
            @click.stop="refreshDetail(row)"
          >
            刷新
          </el-button>
          <el-button
            link
            type="danger"
            @click.stop="deleteEvaluation(row.id)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && evaluations.length === 0" description="暂无测评任务">
      <el-button type="primary" @click="showCreateDialog = true">创建第一个测评任务</el-button>
    </el-empty>

    <!-- 创建测评对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      title="创建测评任务"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="120px">
        <el-form-item label="数据集ID" prop="dataset_id">
          <el-select v-model="formData.dataset_id" placeholder="请选择数据集" style="width: 100%">
            <el-option label="default (默认数据集)" value="default" />
            <el-option
              v-for="dataset in datasets"
              :key="dataset"
              :label="dataset"
              :value="dataset"
            />
          </el-select>
          <div class="form-tip">可选择现有数据集或使用默认数据集</div>
        </el-form-item>

        <el-form-item label="知识库ID" prop="knowledge_base_id">
          <el-select
            v-model="formData.knowledge_base_id"
            placeholder="请选择知识库"
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="kb in knowledgeBases"
              :key="kb.id"
              :label="kb.name"
              :value="kb.id"
            />
          </el-select>
          <div class="form-tip">留空则自动创建临时知识库</div>
        </el-form-item>

        <el-form-item label="对话模型" prop="chat_id">
          <el-select
            v-model="formData.chat_id"
            placeholder="请选择对话模型"
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="model in chatModels"
              :key="model.id"
              :label="model.name"
              :value="model.id"
            />
          </el-select>
          <div class="form-tip">留空使用默认模型</div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="createEvaluation" :loading="creating">
          创建任务
        </el-button>
      </template>
    </el-dialog>

    <!-- 测评结果对话框 -->
    <el-dialog
      v-model="showDetailDialog"
      title="测评结果"
      width="900px"
      :close-on-click-modal="false"
    >
      <div v-if="currentDetail">
        <!-- 任务信息 -->
        <div class="detail-section">
          <h4>任务信息</h4>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="任务ID">{{ currentDetail.task.id }}</el-descriptions-item>
            <el-descriptions-item label="数据集">{{ currentDetail.task.dataset_id }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="EvaluationStatusType[currentDetail.task.status]">
                {{ EvaluationStatusText[currentDetail.task.status] }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="进度">
              {{ currentDetail.task.finished }} / {{ currentDetail.task.total }}
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatTime(currentDetail.task.created_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="结束时间">
              {{ currentDetail.task.end_time ? formatTime(currentDetail.task.end_time) : '-' }}
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <!-- 检索指标 -->
        <div v-if="currentDetail.metric?.retrieval_metrics" class="detail-section">
          <h4>检索指标</h4>
          <div class="metrics-grid">
            <div class="metric-item">
              <div class="metric-label">Precision</div>
              <div class="metric-value">{{ formatPercent(currentDetail.metric.retrieval_metrics.precision) }}</div>
            </div>
            <div class="metric-item">
              <div class="metric-label">Recall</div>
              <div class="metric-value">{{ formatPercent(currentDetail.metric.retrieval_metrics.recall) }}</div>
            </div>
            <div class="metric-item">
              <div class="metric-label">NDCG@3</div>
              <div class="metric-value">{{ formatPercent(currentDetail.metric.retrieval_metrics.ndcg3) }}</div>
            </div>
            <div class="metric-item">
              <div class="metric-label">NDCG@10</div>
              <div class="metric-value">{{ formatPercent(currentDetail.metric.retrieval_metrics.ndcg10) }}</div>
            </div>
            <div class="metric-item">
              <div class="metric-label">MRR</div>
              <div class="metric-value">{{ formatPercent(currentDetail.metric.retrieval_metrics.mrr) }}</div>
            </div>
            <div class="metric-item">
              <div class="metric-label">MAP</div>
              <div class="metric-value">{{ formatPercent(currentDetail.metric.retrieval_metrics.map) }}</div>
            </div>
          </div>
        </div>

        <!-- 生成指标 -->
        <div v-if="currentDetail.metric?.generation_metrics" class="detail-section">
          <h4>生成指标</h4>
          <div class="metrics-grid">
            <div class="metric-item">
              <div class="metric-label">BLEU-1</div>
              <div class="metric-value">{{ formatPercent(currentDetail.metric.generation_metrics.bleu1) }}</div>
            </div>
            <div class="metric-item">
              <div class="metric-label">BLEU-2</div>
              <div class="metric-value">{{ formatPercent(currentDetail.metric.generation_metrics.bleu2) }}</div>
            </div>
            <div class="metric-item">
              <div class="metric-label">BLEU-4</div>
              <div class="metric-value">{{ formatPercent(currentDetail.metric.generation_metrics.bleu4) }}</div>
            </div>
            <div class="metric-item">
              <div class="metric-label">ROUGE-1</div>
              <div class="metric-value">{{ formatPercent(currentDetail.metric.generation_metrics.rouge1) }}</div>
            </div>
            <div class="metric-item">
              <div class="metric-label">ROUGE-2</div>
              <div class="metric-value">{{ formatPercent(currentDetail.metric.generation_metrics.rouge2) }}</div>
            </div>
            <div class="metric-item">
              <div class="metric-label">ROUGE-L</div>
              <div class="metric-value">{{ formatPercent(currentDetail.metric.generation_metrics.rougeL) }}</div>
            </div>
          </div>
        </div>

        <!-- 加载中状态 -->
        <div v-if="currentDetail.task.status === EvaluationStatus.Running" class="loading-state">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>测评执行中，请稍候...</span>
        </div>

        <!-- 失败信息 -->
        <div v-if="currentDetail.task.status === EvaluationStatus.Failed && currentDetail.task.err_msg" class="error-state">
          <el-alert type="error" :closable="false">
            {{ currentDetail.task.err_msg }}
          </el-alert>
        </div>
      </div>

      <template #footer>
        <el-button @click="showDetailDialog = false">关闭</el-button>
        <el-button
          v-if="currentDetail?.task.status === EvaluationStatus.Running"
          type="primary"
          @click="refreshCurrentDetail"
        >
          刷新
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Loading } from '@element-plus/icons-vue'
import { evaluationApi } from '@/api/evaluation'
import { knowledgeApi } from '@/api/knowledge'
import { modelApi } from '@/api/model'
import {
  EvaluationStatus,
  EvaluationStatusText,
  EvaluationStatusType,
  type EvaluationTask,
  type EvaluationDetail,
  type CreateEvaluationRequest
} from '@/types'

const loading = ref(false)
const creating = ref(false)
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const evaluations = ref<EvaluationTask[]>([])
const datasets = ref<string[]>([])
const knowledgeBases = ref<any[]>([])
const chatModels = ref<any[]>([])
const currentDetail = ref<EvaluationDetail | null>(null)
const formRef = ref()
let refreshTimer: number | null = null

// 统计数据
const stats = reactive({
  total: 0,
  success: 0,
  running: 0,
  failed: 0
})

// 表单数据
const formData = reactive<CreateEvaluationRequest>({
  dataset_id: 'default',
  knowledge_base_id: '',
  chat_id: ''
})

const formRules = {
  dataset_id: [
    { required: true, message: '请选择数据集', trigger: 'change' }
  ]
}

function getProgress(task: EvaluationTask): number {
  if (task.total === 0) return 0
  return Math.round((task.finished / task.total) * 100)
}

function formatTime(timeStr: string): string {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN')
}

function formatPercent(value: number): string {
  return (value * 100).toFixed(2) + '%'
}

async function loadEvaluations() {
  loading.value = true
  try {
    const res = await evaluationApi.list({ page: 1, page_size: 100 })
    if (res.data) {
      evaluations.value = res.data.tasks || []
      // 更新统计
      stats.total = evaluations.value.length
      stats.success = evaluations.value.filter(t => t.status === EvaluationStatus.Success).length
      stats.running = evaluations.value.filter(t => t.status === EvaluationStatus.Running).length
      stats.failed = evaluations.value.filter(t => t.status === EvaluationStatus.Failed).length
    }
  } catch (error) {
    console.error('Failed to load evaluations:', error)
  } finally {
    loading.value = false
  }
}

async function loadDatasets() {
  try {
    const res = await evaluationApi.listDatasets()
    if (res.data) {
      datasets.value = res.data as any
    }
  } catch (error) {
    console.error('Failed to load datasets:', error)
  }
}

async function loadKnowledgeBases() {
  try {
    const res = await knowledgeApi.getList()
    if (res.data) {
      knowledgeBases.value = (res.data as any).items || []
    }
  } catch (error) {
    console.error('Failed to load knowledge bases:', error)
  }
}

async function loadModels() {
  try {
    const res = await modelApi.getList('chat')
    if (res.data) {
      chatModels.value = res.data as any || []
    }
  } catch (error) {
    console.error('Failed to load models:', error)
  }
}

async function createEvaluation() {
  if (!formData.dataset_id) {
    ElMessage.warning('请选择数据集')
    return
  }

  creating.value = true
  try {
    const res = await evaluationApi.create(formData)
    if (res.data) {
      ElMessage.success('测评任务创建成功')
      showCreateDialog.value = false
      await loadEvaluations()
      // 自动打开详情
      if (res.data.task) {
        await viewDetail(res.data.task)
      }
    }
  } catch (error: any) {
    ElMessage.error(error.message || '创建失败')
  } finally {
    creating.value = false
  }
}

async function viewDetail(task: EvaluationTask | EvaluationDetail) {
  const taskId = 'id' in task ? task.id : task.task.id
  showDetailDialog.value = true

  try {
    const res = await evaluationApi.getResult(taskId)
    if (res.data) {
      currentDetail.value = res.data
    }
  } catch (error) {
    console.error('Failed to load detail:', error)
  }
}

async function refreshDetail(task: EvaluationTask) {
  try {
    const res = await evaluationApi.getResult(task.id)
    if (res.data) {
      const index = evaluations.value.findIndex(e => e.id === task.id)
      if (index !== -1) {
        evaluations.value[index] = res.data.task
      }
    }
  } catch (error) {
    console.error('Failed to refresh:', error)
  }
}

async function refreshCurrentDetail() {
  if (currentDetail.value) {
    await refreshDetail(currentDetail.value.task)
    await viewDetail(currentDetail.value.task)
    await loadEvaluations()
  }
}

async function deleteEvaluation(id: string) {
  try {
    await ElMessageBox.confirm('删除测评任务后，相关数据也将被删除，此操作不可恢复。确定要删除吗？', '删除确认', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await evaluationApi.delete(id)
    ElMessage.success('删除成功')
    await loadEvaluations()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

// 自动刷新运行中的任务
function startAutoRefresh() {
  refreshTimer = window.setInterval(() => {
    const hasRunning = evaluations.value.some(e => e.status === EvaluationStatus.Running)
    if (hasRunning) {
      loadEvaluations()
      if (currentDetail.value?.task.status === EvaluationStatus.Running) {
        refreshCurrentDetail()
      }
    }
  }, 3000)
}

function stopAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

onMounted(() => {
  loadEvaluations()
  loadDatasets()
  loadKnowledgeBases()
  loadModels()
  startAutoRefresh()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.evaluation-list-container {
  padding: 24px;
  background: white;
  border-radius: 8px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.header h2 {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.stat-card {
  padding: 20px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  color: white;
  text-align: center;
}

.stat-card:nth-child(2) {
  background: linear-gradient(135deg, #84fab0 0%, #8fd3f4 100%);
}

.stat-card:nth-child(3) {
  background: linear-gradient(135deg, #ffecd2 0%, #fcb69f 100%);
}

.stat-card:nth-child(4) {
  background: linear-gradient(135deg, #ff9a9e 0%, #fecfef 100%);
}

.stat-value {
  font-size: 32px;
  font-weight: 700;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  opacity: 0.9;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}

.detail-section {
  margin-bottom: 24px;
}

.detail-section h4 {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.metric-item {
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
  text-align: center;
}

.metric-label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
}

.metric-value {
  font-size: 24px;
  font-weight: 600;
  color: #409eff;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #909399;
}

.loading-state .el-icon {
  margin-right: 8px;
  font-size: 24px;
}

.error-state {
  margin-top: 16px;
}

:deep(.el-table) {
  cursor: pointer;
}

:deep(.el-table__row):hover {
  background-color: #f5f7fa;
}
</style>
