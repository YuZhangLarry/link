<template>
  <div class="graph-list-container">
    <div class="header">
      <h2>知识图谱</h2>
      <el-button type="primary" :icon="Plus" @click="showSelectKBDialog">
        选择知识库查看图谱
      </el-button>
    </div>

    <el-card class="intro-card" shadow="never">
      <div class="intro-content">
        <el-icon :size="48" color="#409eff"><Share /></el-icon>
        <div class="intro-text">
          <h3>什么是知识图谱？</h3>
          <p>知识图谱是从文档中自动提取的实体（人名、地名、概念等）和它们之间的关系（包含、关联、依赖等）构成的可视化网络。</p>
        </div>
      </div>
    </el-card>

    <el-divider />

    <div class="kb-list-section">
      <h3>选择知识库查看图谱</h3>
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
            <el-button type="primary" @click="goToGraph(row.id)">
              <el-icon><Share /></el-icon>
              查看图谱
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && knowledgeBases.length === 0" description="暂无知识库">
        <el-button type="primary" @click="goToKnowledgeList">
          创建知识库
        </el-button>
      </el-empty>
    </div>

    <!-- 选择知识库对话框 -->
    <el-dialog v-model="selectKBDialogVisible" title="选择知识库查看图谱" width="600px">
      <el-table
        :data="knowledgeBases"
        max-height="400"
        @row-click="handleSelectKB"
      >
        <el-table-column prop="name" label="知识库名称" />
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
      </el-table>
      <template #footer>
        <el-button @click="selectKBDialogVisible = false">取消</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Share } from '@element-plus/icons-vue'
import { knowledgeApi } from '@/api/knowledge'
import type { KnowledgeBase } from '@/types'

const router = useRouter()
const loading = ref(false)
const knowledgeBases = ref<KnowledgeBase[]>([])
const selectKBDialogVisible = ref(false)

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

function goToGraph(kbId: string) {
  router.push(`/graphs/${kbId}`)
}

function showSelectKBDialog() {
  if (knowledgeBases.value.length === 0) {
    loadKnowledgeBases()
  } else {
    selectKBDialogVisible.value = true
  }
}

function handleSelectKB(row: KnowledgeBase) {
  selectKBDialogVisible.value = false
  goToGraph(row.id)
}

function goToKnowledgeList() {
  router.push('/knowledge')
}

onMounted(() => {
  loadKnowledgeBases()
})
</script>

<style scoped>
.graph-list-container {
  padding: 24px;
  background: white;
  border-radius: 8px;
  min-height: calc(100vh - 120px);
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

.intro-card {
  margin-bottom: 24px;
}

.intro-content {
  display: flex;
  gap: 24px;
  align-items: flex-start;
}

.intro-text h3 {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px 0;
}

.intro-text p {
  color: #606266;
  line-height: 1.6;
  margin: 0;
}

.kb-list-section h3 {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
}
</style>
