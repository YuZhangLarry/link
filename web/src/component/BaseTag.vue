<template>
  <component
    :is="tag"
    class="base-tag"
    :class="classes"
    :style="tagStyle"
  >
    <!-- 图标 -->
    <span v-if="icon" class="tag-icon">
      <component :is="icon" v-if="typeof icon === 'object'" />
      <span v-else class="icon-string">{{ icon }}</span>
    </span>

    <!-- 文字内容 -->
    <span v-if="$slots.default" class="tag-content">
      <slot />
    </span>

    <!-- 关闭按钮 -->
    <button
      v-if="closable"
      class="tag-close"
      @click="handleClose"
      type="button"
    >
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M18 6L6 18M6 6l12 12"/>
      </svg>
    </button>

    <!-- 装饰点 -->
    <span v-if="dot" class="tag-dot" :style="{ background: dotColor }"></span>
  </component>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'

interface Props {
  variant?: 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'gradient'
  size?: 'xs' | 'sm' | 'md' | 'lg'
  shape?: 'square' | 'rounded' | 'pill'
  icon?: string | Component
  closable?: boolean
  dot?: boolean
  dotColor?: string
  tag?: 'span' | 'div' | 'label' | 'a'
  href?: string
  glow?: boolean
  bordered?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
  size: 'md',
  shape: 'rounded',
  tag: 'span',
  closable: false,
  dot: false,
  dotColor: 'currentColor',
  glow: false,
  bordered: false
})

const emit = defineEmits<{
  close: []
}>()

const classes = computed(() => [
  `variant-${props.variant}`,
  `size-${props.size}`,
  `shape-${props.shape}`,
  {
    'is-closable': props.closable,
    'has-icon': props.icon,
    'has-dot': props.dot,
    'has-glow': props.glow,
    'is-bordered': props.bordered
  }
])

const tagStyle = computed(() => ({}))

function handleClose(event: Event) {
  event.stopPropagation()
  emit('close')
}
</script>

<style scoped>
/* ==================== 基础样式 ==================== */
.base-tag {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-family: var(--font-sans);
  font-weight: 500;
  line-height: 1;
  white-space: nowrap;
  transition: all var(--transition-fast);
  cursor: default;
}

.base-tag:has(a) {
  cursor: pointer;
  text-decoration: none;
}

/* ==================== 形状变体 ==================== */
.shape-square {
  border-radius: var(--radius-sm);
}

.shape-rounded {
  border-radius: var(--radius-md);
}

.shape-pill {
  border-radius: var(--radius-full);
}

/* ==================== 尺寸变体 ==================== */
.size-xs {
  padding: 2px 8px;
  font-size: var(--text-xs);
  gap: 3px;
}

.size-sm {
  padding: 4px 10px;
  font-size: var(--text-sm);
  gap: 4px;
}

.size-md {
  padding: 6px 12px;
  font-size: var(--text-sm);
  gap: 4px;
}

.size-lg {
  padding: 8px 16px;
  font-size: var(--text-base);
  gap: 6px;
}

/* ==================== 变体样式 ==================== */

/* Default - 默认 */
.variant-default {
  background: var(--color-bg-elevated);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border-primary);
}

.variant-default:hover {
  background: var(--color-bg-tertiary);
  color: var(--color-text-primary);
}

/* Primary - 主色 */
.variant-primary {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.2), rgba(99, 102, 241, 0.3));
  color: var(--color-primary-light);
  border: 1px solid rgba(99, 102, 241, 0.3);
}

.variant-primary:hover {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.3), rgba(99, 102, 241, 0.4));
  border-color: rgba(99, 102, 241, 0.5);
}

/* Success - 成功 */
.variant-success {
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.2), rgba(34, 197, 94, 0.3));
  color: #4ade80;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.variant-success:hover {
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.3), rgba(34, 197, 94, 0.4));
}

/* Warning - 警告 */
.variant-warning {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.2), rgba(245, 158, 11, 0.3));
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

.variant-warning:hover {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.3), rgba(245, 158, 11, 0.4));
}

/* Danger - 危险 */
.variant-danger {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.2), rgba(239, 68, 68, 0.3));
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.variant-danger:hover {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.3), rgba(239, 68, 68, 0.4));
}

/* Info - 信息 */
.variant-info {
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.2), rgba(6, 182, 212, 0.3));
  color: #22d3ee;
  border: 1px solid rgba(6, 182, 212, 0.3);
}

.variant-info:hover {
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.3), rgba(6, 182, 212, 0.4));
}

/* Gradient - 渐变 */
.variant-gradient {
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  color: white;
  border: none;
}

.variant-gradient:hover {
  background: linear-gradient(135deg, var(--color-primary-light), var(--color-secondary));
  box-shadow: var(--shadow-glow);
}

/* ==================== 修饰样式 ==================== */
.is-bordered {
  border-width: 2px;
}

.has-glow {
  animation: tagGlow 2s ease-in-out infinite;
}

@keyframes tagGlow {
  0%, 100% {
    box-shadow: 0 0 5px currentColor;
  }
  50% {
    box-shadow: 0 0 15px currentColor, 0 0 25px currentColor;
  }
}

/* ==================== 图标 ==================== */
.tag-icon {
  display: flex;
  align-items: center;
  font-size: 1em;
}

.icon-string {
  font-size: 1.2em;
}

/* ==================== 关闭按钮 ==================== */
.tag-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: inherit;
  cursor: pointer;
  opacity: 0.7;
  transition: all var(--transition-fast);
}

.tag-close:hover {
  opacity: 1;
  background: rgba(0, 0, 0, 0.2);
}

.tag-close svg {
  width: 12px;
  height: 12px;
}

/* ==================== 装饰点 ==================== */
.tag-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  margin-right: -2px;
}

/* ==================== 内容 ==================== */
.tag-content {
  display: inline-flex;
  align-items: center;
}

/* ==================== 响应式 ==================== */
@media (max-width: 640px) {
  .size-lg {
    padding: 6px 12px;
    font-size: var(--text-sm);
  }
}
</style>
