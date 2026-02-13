<template>
  <div class="agent-list-container">
    <div class="header">
      <h2>Agent 管理</h2>
      <el-button type="primary" :icon="Plus">创建 Agent</el-button>
    </div>

    <el-empty description="暂无 Agent" />
  </div>
</template>

<script setup lang="ts">
import { Plus } from '@element-plus/icons-vue'
</script>

<style scoped>
.agent-list-container {
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
}
</style>
