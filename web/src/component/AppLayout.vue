<template>
  <div class="app-layout" :class="layoutClasses">
    <!-- 背景层 -->
    <AppBackground
      :image="backgroundImage"
      :variant="backgroundVariant"
      :noise="backgroundNoise"
      :particles="backgroundParticles"
      :glow="backgroundGlow"
    />

    <!-- 侧边栏 -->
    <BaseSidebar
      v-if="!hideSidebar"
      v-model:collapsed="sidebarCollapsed"
      :items="sidebarItems"
      :logo="logo"
      :title="title"
      :position="sidebarPosition"
      @item-click="handleSidebarClick"
    >
      <template #header>
        <slot name="sidebar-header" />
      </template>
      <template #default>
        <slot name="sidebar" />
      </template>
      <template #footer>
        <slot name="sidebar-footer" />
      </template>
    </BaseSidebar>

    <!-- 主内容区 -->
    <main class="layout-main" :class="{ 'no-sidebar': hideSidebar }">
      <!-- 顶部导航栏 -->
      <header v-if="!hideHeader" class="layout-header">
        <div class="header-left">
          <button
            v-if="!hideSidebar && showToggle"
            class="sidebar-toggle"
            @click="sidebarCollapsed = !sidebarCollapsed"
          >
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 12h18M3 6h18M3 18h18"/>
            </svg>
          </button>
          <slot name="header-left" />
        </div>
        <div class="header-center">
          <slot name="header-center" />
        </div>
        <div class="header-right">
          <slot name="header-right">
            <!-- 用户信息 -->
            <div v-if="user" class="user-dropdown">
              <BaseButton variant="ghost" size="sm" class="user-button">
                <div class="user-avatar">
                  {{ user.name?.charAt(0)?.toUpperCase() || 'U' }}
                </div>
                <span class="user-name">{{ user.name }}</span>
              </BaseButton>
            </div>
          </slot>
        </div>
      </header>

      <!-- 内容区域 -->
      <div class="layout-content" :class="contentClasses">
        <slot />
      </div>

      <!-- 底部 -->
      <footer v-if="!hideFooter" class="layout-footer">
        <slot name="footer">
          <p>&copy; {{ currentYear }} {{ title || 'Link' }}. All rights reserved.</p>
        </slot>
      </footer>
    </main>

    <!-- 全局弹窗容器 -->
    <Teleport to="body">
      <div class="modal-container">
        <slot name="modal" />
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, type Component } from 'vue'
import AppBackground from './AppBackground.vue'
import BaseSidebar from './BaseSidebar.vue'
import BaseButton from './BaseButton.vue'

interface SidebarItem {
  key: string
  label: string
  icon?: string | Component
  to?: string | object
  badge?: string | number
}

interface User {
  name: string
  email?: string
  avatar?: string
}

interface Props {
  // 布局
  layout?: 'default' | 'fluid' | 'boxed'
  hideSidebar?: boolean
  hideHeader?: boolean
  hideFooter?: boolean
  sidebarPosition?: 'left' | 'right'
  showToggle?: boolean

  // 侧边栏
  sidebarItems?: SidebarItem[]
  logo?: string | Component
  title?: string

  // 背景
  backgroundImage?: string
  backgroundVariant?: 'dark' | 'darker' | 'light' | 'gradient' | 'glass'
  backgroundNoise?: boolean
  backgroundParticles?: boolean
  backgroundGlow?: boolean

  // 内容
  contentPadding?: boolean
  contentFullHeight?: boolean

  // 用户
  user?: User
}

const props = withDefaults(defineProps<Props>(), {
  layout: 'default',
  hideSidebar: false,
  hideHeader: false,
  hideFooter: false,
  sidebarPosition: 'left',
  showToggle: true,
  sidebarItems: () => [],
  backgroundVariant: 'dark',
  backgroundNoise: true,
  backgroundParticles: false,
  backgroundGlow: true,
  contentPadding: true,
  contentFullHeight: false
})

const emit = defineEmits<{
  'sidebar-click': [item: SidebarItem]
}>()

const sidebarCollapsed = ref(false)
const currentYear = new Date().getFullYear()

const layoutClasses = computed(() => [
  `layout-${props.layout}`,
  {
    'sidebar-collapsed': sidebarCollapsed.value,
    'sidebar-right': props.sidebarPosition === 'right'
  }
])

const contentClasses = computed(() => [
  {
    'has-padding': props.contentPadding,
    'has-full-height': props.contentFullHeight
  }
])

function handleSidebarClick(item: SidebarItem) {
  emit('sidebar-click', item)
}
</script>

<style scoped>
/* ==================== 布局容器 ==================== */
.app-layout {
  display: flex;
  width: 100%;
  min-height: 100vh;
  position: relative;
}

/* ==================== 侧边栏状态 ==================== */
.sidebar-collapsed .layout-main {
  margin-left: 70px;
}

.sidebar-right .layout-main {
  margin-left: 0;
  margin-right: 260px;
}

.sidebar-right.sidebar-collapsed .layout-main {
  margin-right: 70px;
}

/* ==================== 主内容区 ==================== */
.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  transition: margin var(--transition-base);
}

.layout-main.no-sidebar {
  margin: 0;
}

/* ==================== 顶部导航 ==================== */
.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-xl);
  height: 64px;
  background: var(--color-bg-secondary);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-bottom: 1px solid var(--color-border-primary);
  flex-shrink: 0;
}

.header-left,
.header-center,
.header-right {
  display: flex;
  align-items: center;
}

.header-left {
  flex: 0 0 auto;
  gap: var(--spacing-md);
}

.header-center {
  flex: 1;
  justify-content: center;
}

.header-right {
  flex: 0 0 auto;
  gap: var(--spacing-sm);
}

.sidebar-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.sidebar-toggle:hover {
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
}

.sidebar-toggle svg {
  width: 20px;
  height: 20px;
}

/* 用户下拉 */
.user-dropdown {
  display: flex;
  align-items: center;
}

.user-button {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.user-avatar {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  border-radius: 50%;
  color: white;
  font-weight: 600;
  font-size: var(--text-sm);
}

.user-name {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

/* ==================== 内容区域 ==================== */
.layout-content {
  flex: 1;
  overflow-y: auto;
}

.layout-content.has-padding {
  padding: var(--spacing-xl);
}

.layout-content.has-full-height {
  overflow: hidden;
}

/* ==================== 底部 ==================== */
.layout-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-lg);
  border-top: 1px solid var(--color-border-secondary);
  background: var(--color-bg-secondary);
  flex-shrink: 0;
}

.layout-footer p {
  margin: 0;
  font-size: var(--text-sm);
  color: var(--color-text-muted);
}

/* ==================== 布局变体 ==================== */

/* Default - 默认布局 */
.layout-default .layout-content {
  max-width: 100%;
}

/* Fluid - 流体布局 */
.layout-fluid .layout-content {
  max-width: 100%;
}

/* Boxed - 盒式布局 */
.layout-boxed .layout-content {
  max-width: 1400px;
  margin: 0 auto;
}

/* ==================== 响应式 ==================== */
@media (max-width: 1024px) {
  .sidebar-right .layout-main {
    margin-right: 0;
  }

  .sidebar-right.sidebar-collapsed .layout-main {
    margin-right: 0;
  }
}

@media (max-width: 768px) {
  .layout-header {
    padding: 0 var(--spacing-md);
    height: 56px;
  }

  .layout-content.has-padding {
    padding: var(--spacing-md);
  }

  .user-name {
    display: none;
  }

  .header-center {
    display: none;
  }
}

/* ==================== 模态容器 ==================== */
.modal-container {
  position: relative;
  z-index: 1000;
}
</style>
