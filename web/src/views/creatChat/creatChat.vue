<template>
  <div class="create-chat-container">
    <el-page-header @back="goBack" content="创建对话" />

    <div class="form-container">
      <el-form :model="formData" label-width="100px" style="max-width: 600px">
        <el-form-item label="对话标题">
          <el-input v-model="formData.title" placeholder="请输入对话标题" />
        </el-form-item>

        <el-form-item label="描述">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入对话描述（可选）"
          />
        </el-form-item>

        <el-form-item label="最大轮次">
          <el-input-number v-model="formData.maxRounds" :min="1" :max="100" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleCreate" :loading="creating">
            创建
          </el-button>
          <el-button @click="goBack">取消</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { sessionApi } from '@/api/session'

const router = useRouter()

const creating = ref(false)
const formData = reactive({
  title: '',
  description: '',
  maxRounds: 50
})

function goBack() {
  router.back()
}

async function handleCreate() {
  if (!formData.title) {
    ElMessage.warning('请输入对话标题')
    return
  }

  creating.value = true
  try {
    const res = await sessionApi.create({
      title: formData.title,
      description: formData.description,
      max_rounds: formData.maxRounds
    })

    if (res.data) {
      ElMessage.success('创建成功')
      router.push(`/chat`)
    }
  } catch (error) {
    console.error('Failed to create session:', error)
  } finally {
    creating.value = false
  }
}
</script>

<style scoped>
.create-chat-container {
  padding: 24px;
  background: white;
  border-radius: 8px;
}

.form-container {
  margin-top: 24px;
}
</style>
