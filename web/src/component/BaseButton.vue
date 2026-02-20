<template>
  <component
    :is="tag"
    :type="tag === 'button' ? nativeType : undefined"
    :to="tag === 'router-link' ? to : undefined"
    :href="tag === 'a' ? to : undefined"
    :disabled="disabled || loading"
    class="base-button"
    :class="classes"
    :style="customStyle"
    @click="handleClick"
  >
    <span v-if="loading" class="button-loader">
      <svg class="spinner" viewBox="0 0 24 24">
        <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" stroke-width="3" stroke-dasharray="32" stroke-dashoffset="32" />
      </svg>
    </span>

    <span v-if="icon && !loading" class="button-icon" :class="{ 'icon-only': !$slots.default }">
      <component :is="icon" v-if="typeof icon === 'object'" />
      <span v-else class="icon-string">{{ icon }}</span>
    </span>

      <span v-if="$slots.default" class="button-content">
      <slot />
    </span>

    <span v-if="badge" class="button-badge">{{ badge }}</span>
  </component>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'

interface Props {
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger' | 'success' | 'gradient' | 'glass'
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
  tag?: 'button' | 'a' | 'router-link'
  nativeType?: 'button' | 'submit' | 'reset'
  icon?: string | Component
  badge?: string | number
  disabled?: boolean
  loading?: boolean
  block?: boolean
  rounded?: boolean
  to?: string | object
  glow?: boolean
  shadow?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary',
  size: 'md',
  tag: 'button',
  nativeType: 'button',
  disabled: false,
  loading: false,
  block: false,
  rounded: false,
  glow: false,
  shadow: false
})

const emit = defineEmits<{
  click: [event: Event]
}>()

const classes = computed(() => [
  `variant-${props.variant}`,
  `size-${props.size}`,
  {
    'is-disabled': props.disabled,
    'is-loading': props.loading,
    'is-block': props.block,
    'is-rounded': props.rounded,
    'has-glow': props.glow,
    'has-shadow': props.shadow
  }
])

const customStyle = computed(() => {
  if (props.glow && !props.disabled) {
    return {
      '--button-glow-color': 'var(--color-primary)'
    }
  }
  return {}
})

function handleClick(event: Event) {
  if (!props.disabled && !props.loading) {
    emit('click', event)
  }
}
</script>

<style scoped>
/* ==================== 基础样式 ==================== */
.base-button {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  font-family: var(--font-sans);
  font-weight: 500;
  line-height: 1;
  white-space: nowrap;
  cursor: pointer;
  user-select: none;
  transition: all var(--transition-base);
  border: none;
  outline: none;
  text-decoration: none;
}

.base-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

/* ==================== 尺寸变体 ==================== */
.size-xs {
  padding: 4px 10px;
  font-size: var(--text-xs);
  border-radius: var(--radius-sm);
  gap: 4px;
}

.size-sm {
  padding: 6px 14px;
  font-size: var(--text-sm);
  border-radius: var(--radius-sm);
  gap: 6px;
}

.size-md {
  padding: 10px 18px;
  font-size: var(--text-base);
  border-radius: var(--radius-md);
  gap: var(--spacing-sm);
}

.size-lg {
  padding: 14px 24px;
  font-size: var(--text-lg);
  border-radius: var(--radius-lg);
  gap: var(--spacing-md);
}

.size-xl {
  padding: 18px 32px;
  font-size: var(--text-xl);
  border-radius: var(--radius-xl);
  gap: var(--spacing-lg);
}

/* ==================== 变体样式 ==================== */

/* Primary - 主按钮 */
.variant-primary {
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dark));
  color: white;
  box-shadow: var(--shadow-md), 0 0 0 1px rgba(255, 255, 255, 0.1) inset;
}

.variant-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, var(--color-primary-light), var(--color-primary));
  box-shadow: var(--shadow-lg), 0 0 0 1px rgba(255, 255, 255, 0.15) inset;
  transform: translateY(-1px);
}

.variant-primary:active:not(:disabled) {
  transform: translateY(0);
  box-shadow: var(--shadow-sm);
}

/* Secondary - 次要按钮 */
.variant-secondary {
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
  border: 1px solid var(--color-border-primary);
  backdrop-filter: blur(10px);
}

.variant-secondary:hover:not(:disabled) {
  background: var(--color-bg-tertiary);
  border-color: var(--color-border-focus);
}

.variant-secondary:active:not(:disabled) {
  transform: scale(0.98);
}

/* Ghost - 幽灵按钮 */
.variant-ghost {
  background: transparent;
  color: var(--color-text-secondary);
}

.variant-ghost:hover:not(:disabled) {
  background: var(--color-bg-elevated);
  color: var(--color-text-primary);
}

/* Danger - 危险按钮 */
.variant-danger {
  background: linear-gradient(135deg, var(--color-danger), #dc2626);
  color: white;
  box-shadow: var(--shadow-md);
}

.variant-danger:hover:not(:disabled) {
  background: linear-gradient(135deg, #f87171, var(--color-danger));
  box-shadow: var(--shadow-lg);
  transform: translateY(-1px);
}

/* Success - 成功按钮 */
.variant-success {
  background: linear-gradient(135deg, var(--color-success), #16a34a);
  color: white;
  box-shadow: var(--shadow-md);
}

.variant-success:hover:not(:disabled) {
  background: linear-gradient(135deg, #4ade80, var(--color-success));
  box-shadow: var(--shadow-lg);
  transform: translateY(-1px);
}

/* Gradient - 渐变按钮 */
.variant-gradient {
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary), var(--color-accent));
  background-size: 200% 200%;
  color: white;
  animation: gradientShift 3s ease infinite;
  box-shadow: var(--shadow-md);
}

.variant-gradient:hover:not(:disabled) {
  background-position: right center;
  box-shadow: var(--shadow-glow);
  transform: translateY(-1px);
}

@keyframes gradientShift {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}

/* Glass - 玻璃态按钮 */
.variant-glass {
  background: rgba(255, 255, 255, 0.1);
  color: var(--color-text-primary);
  border: 1px solid rgba(255, 255, 255, 0.2);
  backdrop-filter: blur(10px) saturate(150%);
  -webkit-backdrop-filter: blur(10px) saturate(150%);
}

.variant-glass:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.15);
  border-color: rgba(255, 255, 255, 0.3);
  box-shadow: 0 8px 32px rgba(99, 102, 241, 0.2);
}

/* ==================== 修饰类 ==================== */
.is-block {
  width: 100%;
}

.is-rounded {
  border-radius: var(--radius-full);
}

.has-glow:not(:disabled) {
  animation: buttonGlow 2s ease-in-out infinite;
}

@keyframes buttonGlow {
  0%, 100% {
    box-shadow: 0 0 5px var(--button-glow-color, var(--color-primary)),
                0 0 10px var(--button-glow-color, var(--color-primary));
  }
  50% {
    box-shadow: 0 0 15px var(--button-glow-color, var(--color-primary)),
                0 0 25px var(--button-glow-color, var(--color-primary)),
                0 0 35px var(--button-glow-color, var(--color-primary));
  }
}

.has-shadow {
  box-shadow: var(--shadow-lg);
}

/* ==================== 图标 ==================== */
.button-icon {
  display: flex;
  align-items: center;
  justify-content: center;
}

.button-icon svg {
  width: 1em;
  height: 1em;
}

.icon-string {
  font-size: 1.2em;
}

.icon-only {
  margin: 0;
}

/* ==================== 加载状态 ==================== */
.button-loader {
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.spinner {
  width: 1em;
  height: 1em;
  animation: spin 0.8s linear infinite;
}

.spinner circle {
  animation: dash 1.5s ease-in-out infinite;
}

@keyframes dash {
  0% {
    stroke-dasharray: 1, 150;
    stroke-dashoffset: 0;
  }
  50% {
    stroke-dasharray: 90, 150;
    stroke-dashoffset: -35;
  }
  100% {
    stroke-dasharray: 90, 150;
    stroke-dashoffset: -124;
  }
}

/* ==================== 徽章 ==================== */
.button-badge {
  position: absolute;
  top: -6px;
  right: -6px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  background: var(--color-danger);
  color: white;
  font-size: 10px;
  font-weight: 600;
  line-height: 18px;
  text-align: center;
  border-radius: var(--radius-full);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

/* ==================== 内容 ==================== */
.button-content {
  display: inline-flex;
  align-items: center;
}
</style>
