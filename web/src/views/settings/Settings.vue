<template>
  <div class="settings-container">
    <el-page-header @back="goBack" :content="t('settings.title')" />

    <el-tabs v-model="activeTab" class="settings-tabs">
      <el-tab-pane :name="tab.key" :label="tab.label" v-for="tab in tabs" :key="tab.key">
        <component :is="tab.component" v-if="tab.component" />
        <div v-else class="placeholder">
          <el-icon :size="64"><Tools /></el-icon>
          <p>{{ tab.label }}功能开发中...</p>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { Tools } from '@element-plus/icons-vue'
import { useI18n } from 'vue-i18n'
import GeneralSettings from './GeneralSettings.vue'

const { t } = useI18n()
const router = useRouter()

const activeTab = ref('general')

const tabs = ref([
  { key: 'general', label: t('settings.general'), component: GeneralSettings },
  { key: 'model', label: t('settings.model'), component: null },
  { key: 'agent', label: t('settings.agent'), component: null },
  { key: 'mcp', label: t('settings.mcp'), component: null },
  { key: 'webSearch', label: t('settings.webSearch'), component: null },
  { key: 'apiInfo', label: t('settings.apiInfo'), component: null },
  { key: 'systemInfo', label: t('settings.systemInfo'), component: null },
  { key: 'tenantInfo', label: t('settings.tenantInfo'), component: null }
])

function goBack() {
  router.back()
}
</script>

<style scoped>
.settings-container {
  padding: 24px;
  background: white;
  border-radius: 8px;
}

.settings-tabs {
  margin-top: 24px;
}

.placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 0;
  color: #909399;
}

.placeholder p {
  margin-top: 16px;
  font-size: 14px;
}
</style>
