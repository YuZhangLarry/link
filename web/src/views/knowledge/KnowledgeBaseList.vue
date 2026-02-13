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
      <el-table-column prop="description" label="描述" min-width="300" show-overflow-tooltip />
      <el-table-column prop="type" label="类型" width="120">
        <template #default="{ row }">
          <el-tag>{{ row.type || '通用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="goToDetail(row.id)">
            查看
          </el-button>
          <el-button link type="primary" @click="editKnowledgeBase(row)">
            编辑
          </el-button>
          <el-button link type="danger" @click="deleteKnowledgeBase(row.id)">
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!loading && knowledgeBases.length === 0" description="暂无知识库">
      <el-button type="primary" @click="showCreateDialog = true">
        创建第一个知识库
      </el-button>
    </el-empty>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingId ? '编辑知识库' : '新建知识库'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="100px">
        <el-form-item label="知识库名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入知识库名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入知识库描述"
          />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="formData.type" placeholder="请选择类型" style="width: 100%">
            <el-option label="通用" value="general" />
            <el-option label="技术文档" value="technical" />
            <el-option label="产品手册" value="product" />
            <el-option label="FAQ" value="faq" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>

        <el-divider content-position="left">策略配置</el-divider>

        <el-form-item label="向量模型">
          <el-select v-model="formData.embedding_model" placeholder="请选择向量模型" style="width: 100%">
            <el-option label="BGE-M3" value="bge-m3" />
            <el-option label="BGE-Large" value="bge-large" />
            <el-option label="Text2Vec" value="text2vec" />
          </el-select>
        </el-form-item>

        <el-form-item label="分块大小">
          <el-input-number
            v-model="formData.chunk_size"
            :min="128"
            :max="2048"
            :step="64"
            style="width: 100%"
          />
          <span class="form-tip">建议值：512-1024，越小分块越精细，越大语义越完整</span>
        </el-form-item>

        <el-form-item label="分块重叠">
          <el-input-number
            v-model="formData.chunk_overlap"
            :min="0"
            :max="512"
            :step="32"
            style="width: 100%"
          />
          <span class="form-tip">建议值：50-200，重叠可以保持上下文连贯性</span>
        </el-form-item>

        <el-form-item label="启用图谱">
          <el-switch v-model="formData.enable_graph" />
          <span class="form-tip">启用后自动构建知识图谱，提取实体和关系</span>
        </el-form-item>

        <el-form-item label="启用标签">
          <el-switch v-model="formData.enable_tag" />
          <span class="form-tip">启用后自动为文档分块打标签</span>
        </el-form-item>
      </el-form>
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
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { knowledgeApi } from '@/api/knowledge'
import type { KnowledgeBase, CreateKnowledgeBaseRequest } from '@/types'

const router = useRouter()

const loading = ref(false)
const saving = ref(false)
const showCreateDialog = ref(false)
const editingId = ref<string>('')
const formRef = ref<FormInstance>()

const formData = reactive<CreateKnowledgeBaseRequest & { status?: number }>({
  name: '',
  description: '',
  type: 'general',
  embedding_model: 'bge-m3',
  chunk_size: 512,
  chunk_overlap: 100,
  enable_graph: true,
  enable_tag: true
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入知识库名称', trigger: 'blur' },
    { min: 2, max: 50, message: '名称长度在2-50个字符之间', trigger: 'blur' }
  ]
}

const knowledgeBases = ref<KnowledgeBase[]>([])

function getStatusType(status: number) {
  if (status === 1) return 'success'
  if (status === 0) return 'info'
  return 'warning'
}

function getStatusText(status: number) {
  if (status === 1) return '启用'
  if (status === 0) return '禁用'
  return '未知'
}

async function loadKnowledgeBases() {
  loading.value = true
  try {
    const res = await knowledgeApi.getList()
    if (res.data) {
      // 后端返回的是 {items: [], page: 1, page_size: 10, total: 0}
      knowledgeBases.value = (res.data as any).items || []
    }
  } catch (error) {
    console.error('Failed to load knowledge bases:', error)
  } finally {
    loading.value = false
  }
}

function goToDetail(id: string) {
  router.push(`/knowledge/${id}`)
}

function editKnowledgeBase(kb: KnowledgeBase) {
  editingId.value = kb.id
  formData.name = kb.name
  formData.description = kb.description || ''
  formData.type = kb.type
  formData.status = kb.status
  showCreateDialog.value = true
}

async function saveKnowledgeBase() {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    saving.value = true
    try {
      if (editingId.value) {
        // 编辑模式
        await knowledgeApi.update(editingId.value, formData)
        ElMessage.success('知识库更新成功')
      } else {
        // 创建模式
        await knowledgeApi.create(formData)
        ElMessage.success('知识库创建成功')
      }
      showCreateDialog.value = false
      resetForm()
      await loadKnowledgeBases()
    } catch (error: any) {
      ElMessage.error(error.message || '操作失败')
    } finally {
      saving.value = false
    }
  })
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

function resetForm() {
  editingId.value = ''
  formData.name = ''
  formData.description = ''
  formData.type = 'general'
  formData.embedding_model = 'bge-m3'
  formData.chunk_size = 512
  formData.chunk_overlap = 100
  formData.enable_graph = true
  formData.enable_tag = true
  formRef.value?.resetFields()
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
  margin-left: 12px;
  font-size: 12px;
  color: #909399;
}
</style>
