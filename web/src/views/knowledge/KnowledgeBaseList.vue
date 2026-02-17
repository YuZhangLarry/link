<template>
  <div class="knowledge-list-container">
    <div class="header">
      <h2>知识库管理</h2>
      <el-button type="primary" :icon="Plus" @click="showCreateDialog = true">
        新建知识库
      </el-button>
    </div>

    <el-table
      :data="knowledgeBases"
      v-loading="loading"
      stripe
      style="width: 100%"
    >
      <el-table-column prop="name" label="知识库名称" min-width="200" />
      <el-table-column prop="description" label="描述" show-overflow-tooltip />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.status === 1" type="success">启用</el-tag>
          <el-tag v-else type="info">禁用</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="goToDetail(row.id)">查看</el-button>
          <el-button link type="primary" @click="editKnowledgeBase(row)">编辑</el-button>
          <el-button link type="danger" @click="deleteKnowledgeBase(row.id)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && knowledgeBases.length === 0" description="暂无知识库">
      <el-button type="primary" @click="showCreateDialog = true">创建第一个知识库</el-button>
    </el-empty>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingId ? '编辑知识库' : '新建知识库'"
      width="700px"
      :close-on-click-modal="false"
      @open="handleDialogOpen"
    >
      <el-tabs v-model="activeTab" class="dialog-tabs">
        <!-- 基本信息 -->
        <el-tab-pane label="基本信息" name="basic">
          <el-form :model="formData" :rules="formRules" ref="formRef" label-width="120px">
            <el-form-item label="知识库名称" prop="name">
              <el-input v-model="formData.name" placeholder="请输入知识库名称" clearable />
            </el-form-item>

            <el-form-item label="描述" prop="description">
              <el-input
                v-model="formData.description"
                type="textarea"
                :rows="2"
                placeholder="请输入知识库描述"
                clearable />
            </el-form-item>

            <el-form-item label="图标URL" prop="avatar">
              <el-input v-model="formData.avatar" placeholder="请输入图标URL" clearable />
            </el-form-item>

            <el-form-item label="可见性" prop="is_public">
              <el-switch v-model="formData.is_public" active-text="公开" inactive-text="私有" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 数据处理配置 -->
        <el-tab-pane label="数据处理配置" name="processing">
          <el-form :model="formData" label-width="140px">
            <el-divider content-position="left">分块配置</el-divider>

            <el-form-item label="分块大小">
              <el-input-number
                v-model="formData.chunk_size"
                :min="128"
                :max="2048"
                :step="64"
              />
              <div class="form-tip">建议值：512-1024，值越小分块越精细</div>
            </el-form-item>

            <el-form-item label="分块重叠">
              <el-input-number
                v-model="formData.chunk_overlap"
                :min="0"
                :max="512"
                :step="32"
              />
              <div class="form-tip">建议值：50-200，重叠可以保持上下文连贯性</div>
            </el-form-item>

            <el-divider content-position="left">索引构建</el-divider>

            <el-form-item label="构建知识图谱">
              <el-switch v-model="formData.graph_enabled" />
              <div class="form-tip">为文档构建知识图谱，支持图谱检索</div>
            </el-form-item>

            <el-form-item label="构建BM25索引">
              <el-switch v-model="formData.bm25_enabled" />
              <div class="form-tip">构建BM25关键词索引，支持关键词检索</div>
            </el-form-item>

          </el-form>
        </el-tab-pane>

        <!-- 检索配置说明 -->
        <el-tab-pane label="检索配置说明" name="retrieval-info">
          <div class="info-panel">
            <el-alert type="info" :closable="false">
              <template #title>
                <span style="display: flex; align-items: center; gap: 8px;">
                  <el-icon><InfoFilled /></el-icon>
                  检索配置已迁移至会话级别
                </span>
              </template>
              <p>知识库的检索配置（如相似度阈值、返回数量、重排序等）现在可以在创建会话时进行调整，实现跨知识库的统一检索配置。</p>
              <p>这样设计的好处：</p>
              <ul>
                <li>支持跨多个知识库的统一检索配置</li>
                <li>不同的会话可以使用不同的检索策略</li>
                <li>知识库专注于数据处理，检索策略按需配置</li>
              </ul>
            </el-alert>

            <el-divider />

            <h4>默认检索配置</h4>
            <el-descriptions :column="1" border>
              <el-descriptions-item label="检索模式">向量检索（必选）+ 可选 BM25/图谱检索</el-descriptions-item>
              <el-descriptions-item label="相似度阈值">70%</el-descriptions-item>
              <el-descriptions-item label="返回数量">5条</el-descriptions-item>
              <el-descriptions-item label="重排序">默认关闭</el-descriptions-item>
            </el-descriptions>
          </div>
        </el-tab-pane>
      </el-tabs>

      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveKnowledgeBase" :loading="saving">
          {{ editingId ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, InfoFilled } from '@element-plus/icons-vue'
import { knowledgeApi } from '@/api/knowledge'
import type {
  KnowledgeBase,
  CreateKnowledgeBaseRequest
} from '@/types'

const router = useRouter()

const loading = ref(false)
const saving = ref(false)
const showCreateDialog = ref(false)
const editingId = ref<string>('')
const knowledgeBases = ref<KnowledgeBase[]>([])
const activeTab = ref('basic')
const formRef = ref()

// 表单数据
const formData = reactive<CreateKnowledgeBaseRequest>({
  name: '',
  description: '',
  avatar: '',
  is_public: false,
  chunk_size: 512,
  chunk_overlap: 100,
  graph_enabled: false,
  bm25_enabled: false
})

const formRules = {
  name: [
    { required: true, message: '请输入知识库名称', trigger: 'blur' }
  ]
}

function goToDetail(id: string) {
  router.push(`/knowledge/${id}`)
}

async function loadKnowledgeBases() {
  loading.value = true
  try {
    const res = await knowledgeApi.getList()
    if (res.data) {
      knowledgeBases.value = (res.data as any).items || []
    }
  } catch (error) {
    console.error('Failed to load knowledge bases:', error)
  } finally {
    loading.value = false
  }
}

function handleDialogOpen() {
  if (editingId.value) {
    // 编辑模式：加载现有数据
    loadKnowledgeBaseData(editingId.value)
  } else {
    // 创建模式：重置表单
    resetForm()
  }
}

async function loadKnowledgeBaseData(id: string) {
  try {
    const res = await knowledgeApi.getDetail(id)
    if (res.data) {
      const data = res.data as KnowledgeBase

      // 基本信息
      formData.name = data.name || ''
      formData.description = data.description || ''
      formData.avatar = data.avatar || ''
      formData.is_public = data.is_public || false

      // 数据处理配置
      if (data.setting) {
        formData.chunk_size = data.setting.chunk_size ?? 512
        formData.chunk_overlap = data.setting.chunk_overlap ?? 100
        formData.graph_enabled = data.setting.graph_enabled ?? false
        formData.bm25_enabled = data.setting.bm25_enabled ?? false
        formData.image_processing_mode = data.setting.image_processing_mode || 'none'
        formData.extract_mode = data.setting.extract_mode || 'none'
      }
    }
  } catch (error) {
    console.error('Failed to load knowledge base:', error)
  }
}

function resetForm() {
  Object.assign(formData, {
    name: '',
    description: '',
    avatar: '',
    is_public: false,
    chunk_size: 512,
    chunk_overlap: 100,
    graph_enabled: false,
    bm25_enabled: false
  })
}

async function saveKnowledgeBase() {
  if (!formData.name) {
    ElMessage.warning('请输入知识库名称')
    return
  }

  saving.value = true
  try {
    const request: CreateKnowledgeBaseRequest = {
      name: formData.name,
      description: formData.description,
      avatar: formData.avatar,
      is_public: formData.is_public,
      chunk_size: formData.chunk_size,
      chunk_overlap: formData.chunk_overlap,
      graph_enabled: formData.graph_enabled,
      bm25_enabled: formData.bm25_enabled
    }

    if (editingId.value) {
      await knowledgeApi.update(editingId.value, request)
      ElMessage.success('知识库更新成功')
    } else {
      await knowledgeApi.create(request)
      ElMessage.success('知识库创建成功')
    }

    showCreateDialog.value = false
    await loadKnowledgeBases()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    saving.value = false
  }
}

async function editKnowledgeBase(kb: KnowledgeBase) {
  editingId.value = kb.id
  showCreateDialog.value = true
}

async function deleteKnowledgeBase(id: string) {
  try {
    await ElMessageBox.confirm('删除知识库后，所有相关文档和分块也将被删除，此操作不可恢复。确定要删除吗？', '删除确认', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await knowledgeApi.delete(id)
    ElMessage.success('知识库删除成功')
    await loadKnowledgeBases()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

onMounted(() => {
  loadKnowledgeBases()
})
</script>

<style scoped>
.knowledge-list-container {
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

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}

.tag-tip {
  font-size: 12px;
  color: #909399;
  margin-left: 8px;
}

.dialog-tabs {
  margin-bottom: 20px;
}

.info-panel {
  padding: 16px;
}

.info-panel h4 {
  margin-bottom: 16px;
  color: #303133;
}

.info-panel p {
  margin-bottom: 12px;
  color: #606266;
}

.info-panel ul {
  padding-left: 20px;
  color: #606266;
}

.info-panel li {
  margin-bottom: 8px;
}

:deep(.el-divider__text) {
  font-size: 14px;
  font-weight: 500;
}
</style>
