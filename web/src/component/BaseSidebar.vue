<template>
  <aside class="base-sidebar" :class="classes" :style="sidebarStyle">
    <!-- 头部 -->
    <div v-if="!hideHeader" class="sidebar-header">
      <slot name="header">
        <div class="header-content">
          <div v-if="logo" class="sidebar-logo">
            <component :is="logo" v-if="typeof logo === 'object'" />
            <span v-else class="logo-text">{{ logo }}</span>
          </div>
          <h1 v-if="title" class="sidebar-title">{{ title }}</h1>
        </div>
      </slot>
      <button v-if="collapsible && !collapsed" class="collapse-toggle" @click="toggleCollapse">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M11 19l-7-7 7-7M18 19l-7-7 7-7"/>
        </svg>
      </button>
    </div>

    <!-- 导航菜单 -->
    <nav class="sidebar-nav">
      <div v-if="$slots.prepend" class="nav-prepend">
        <slot name="prepend" />
      </div>

      <div class="nav-list">
        <slot name="default">
          <template v-for="item in items" :key="item.key">
            <!-- 导航分组 -->
            <div v-if="item.group" class="nav-group">
              <div class="group-title" v-if="!collapsed">{{ item.group }}</div>
            </div>

            <!-- 导航项 -->
            <component
              :is="item.to ? 'router-link' : 'button'"
              v-else
              class="nav-item"
              :class="{
                'is-active': isActive(item),
                'is-disabled': item.disabled,
                'has-icon': item.icon,
                'has-badge': item.badge
              }"
              :to="item.to"
              :href="item.href"
              @click="handleItemClick(item)"
            >
              <div v-if="item.icon" class="nav-icon">
                <component :is="item.icon" v-if="typeof item.icon === 'object'" />
                <span v-else class="icon-string">{{ item.icon }}</span>
              </div>
              <span v-if="!collapsed" class="nav-label">{{ item.label }}</span>
              <span v-if="item.badge && !collapsed" class="nav-badge">{{ item.badge }}</span>
            </component>
          </template>
        </slot>
      </div>

      <div v-if="$slots.append" class="nav-append">
        <slot name="append" />
      </div>
    </nav>

    <!-- 底部 -->
    <div v-if="!hideFooter" class="sidebar-footer">
      <slot name="footer" />
      <button v-if="collapsible && collapsed" class="expand-toggle" @click="toggleCollapse" :title="expandText">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M15 19l-7-7 7-7"/>
        </svg>
      </button>
    </div>

    <!-- 装饰光效 -->
    <div v-if="glow" class="sidebar-glow"></div>
  </aside>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'
import { useRoute } from 'vue-router'

interface NavItem {
  key: string
  label?: string
  icon?: string | Component
  to?: string | object
  href?: string
  disabled?: boolean
  badge?: string | number
  group?: string
}

interface Props {
  collapsed?: boolean
  collapsible?: boolean
  position?: 'left' | 'right'
  variant?: 'default' | 'glass' | 'solid' | 'transparent'
  width?: string
  collapsedWidth?: string
  logo?: string | Component
  title?: string
  hideHeader?: boolean
  hideFooter?: boolean
  glow?: boolean
  expandText?: string
  items?: NavItem[]
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false,
  collapsible: true,
  position: 'left',
  variant: 'glass',
  width: '260px',
  collapsedWidth: '70px',
  hideHeader: false,
  hideFooter: false,
  glow: true,
  expandText: '展开',
  items: () => []
})

const emit = defineEmits<{
  'update:collapsed': [value: boolean]
  'item-click': [item: NavItem]
}>()

const route = useRoute()

const classes = computed(() => [
  `position-${props.position}`,
  `variant-${props.variant}`,
  {
    'is-collapsed': props.collapsed
  }
])

const sidebarStyle = computed(() => ({
  '--sidebar-width': props.collapsed ? props.collapsedWidth : props.width
}))

function isActive(item: NavItem): boolean {
  if (item.to) {
    const path = typeof item.to === 'string' ? item.to : item.to.path
    return route.path === path || route.path.startsWith(path + '/')
  }
  return false
}

function toggleCollapse() {
  emit('update:collapsed', !props.collapsed)
}

function handleItemClick(item: NavItem) {
  if (!item.disabled) {
    emit('item-click', item)
  }
}
</script>

<style scoped>
/* ==================== 基础样式 ==================== */
.base-sidebar {
  position: relative;
  width: var(--sidebar-width);
  height: 100%;
  display: flex;
  flex-direction: column;
  transition: width var(--transition-base);
  z-index: 100;
}

/* ==================== 位置变体 ==================== */
.position-left {
  border-right: 1px solid var(--color-border-primary);
}

.position-right {
  border-left: 1px solid var(--color-border-primary);
}

/* ==================== 变体样式 ==================== */

/* Glass - 玻璃态（默认） */
.variant-glass {
  background: var(--color-bg-secondary);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
}

/* Solid - 实色 */
.variant-solid {
  background: var(--color-bg-secondary);
}

/* Transparent - 透明 */
.variant-transparent {
  background: transparent;
}

.variant-transparent .nav-item:hover {
  background: var(--color-bg-elevated);
}

/* ==================== 折叠状态 ==================== */
.is-collapsed .sidebar-header .header-content,
.is-collapsed .nav-label,
.is-collapsed .nav-badge,
.is-collapsed .group-title {
  opacity: 0;
  width: 0;
  overflow: hidden;
}

.is-collapsed .collapse-toggle {
  display: none;
}

/* ==================== 头部 ==================== */
.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-secondary);
  min-height: 64px;
}

.header-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  overflow: hidden;
}

.sidebar-logo {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  border-radius: var(--radius-md);
  color: white;
  font-size: 18px;
  flex-shrink: 0;
}

.sidebar-logo svg {
  width: 20px;
  height: 20px;
}

.logo-text {
  font-weight: 700;
  font-size: 18px;
}

.sidebar-title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
  white-space: nowrap;
}

.collapse-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.collapse-toggle:hover {
  background: var(--color-bg-tertiary);
  color: var(--color-text-primary);
}

.collapse-toggle svg {
  width: 18px;
  height: 18px;
}

/* ==================== 导航区域 ==================== */
.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: var(--spacing-sm);
}

.nav-prepend,
.nav-append {
  padding: var(--spacing-sm) 0;
}

.nav-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

/* 导航分组 */
.nav-group {
  padding: var(--spacing-md) var(--spacing-sm) var(--spacing-xs);
}

.group-title {
  font-size: var(--text-xs);
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  transition: opacity var(--transition-fast);
}

/* 导航项 */
.nav-item {
  position: relative;
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: 10px var(--spacing-md);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  text-decoration: none;
  font-size: var(--text-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  border: none;
  background: transparent;
  width: 100%;
  text-align: left;
}

.nav-item:hover:not(.is-disabled) {
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
}

.nav-item.is-active {
  background: linear-gradient(135deg,
    rgba(99, 102, 241, 0.2),
    rgba(168, 85, 247, 0.2)
  );
  color: var(--color-primary-light);
}

.nav-item.is-active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 60%;
  background: linear-gradient(180deg, var(--color-primary), var(--color-secondary));
  border-radius: 0 2px 2px 0;
}

.nav-item.is-disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.nav-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.nav-icon svg {
  width: 18px;
  height: 18px;
}

.icon-string {
  font-size: 16px;
}

.nav-label {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: opacity var(--transition-fast);
}

.nav-badge {
  padding: 2px 8px;
  background: var(--color-danger);
  color: white;
  font-size: var(--text-xs);
  font-weight: 600;
  border-radius: var(--radius-full);
  transition: opacity var(--transition-fast);
}

/* ==================== 底部 ==================== */
.sidebar-footer {
  padding: var(--spacing-md);
  border-top: 1px solid var(--color-border-secondary);
  position: relative;
}

.expand-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 10px;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-primary);
  border-radius: var(--radius-md);
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.expand-toggle:hover {
  background: var(--color-bg-tertiary);
  color: var(--color-text-primary);
}

.expand-toggle svg {
  width: 18px;
  height: 18px;
}

/* ==================== 装饰光效 ==================== */
.sidebar-glow {
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(
    circle at center,
    rgba(99, 102, 241, 0.1) 0%,
    transparent 50%
  );
  pointer-events: none;
  z-index: -1;
}

/* ==================== 滚动条 ==================== */
.sidebar-nav::-webkit-scrollbar {
  width: 4px;
}

.sidebar-nav::-webkit-scrollbar-track {
  background: transparent;
}

.sidebar-nav::-webkit-scrollbar-thumb {
  background: var(--color-bg-tertiary);
  border-radius: 2px;
}

.sidebar-nav::-webkit-scrollbar-thumb:hover {
  background: var(--color-bg-elevated);
}
</style>
