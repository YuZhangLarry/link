<template>
  <div class="base-card" :class="classes" :style="cardStyle">
    <!-- 头部 -->
    <div v-if="$slots.header || title || subtitle" class="card-header">
      <div class="header-content">
        <div v-if="icon" class="header-icon">
          <component :is="icon" v-if="typeof icon === 'object'" />
          <span v-else class="icon-string">{{ icon }}</span>
        </div>
        <div class="header-text">
          <h3 v-if="title" class="card-title">{{ title }}</h3>
          <p v-if="subtitle" class="card-subtitle">{{ subtitle }}</p>
        </div>
      </div>
      <div v-if="$slots.extra" class="header-extra">
        <slot name="extra" />
      </div>
    </div>

    <!-- 内容 -->
    <div class="card-body" :class="{ 'no-padding': noPadding }">
      <slot />
    </div>

    <!-- 底部 -->
    <div v-if="$slots.footer" class="card-footer">
      <slot name="footer" />
    </div>

    <!-- 装饰元素 -->
    <div v-if="glow" class="card-glow" :style="{ '--glow-color': glowColor }"></div>
    <div v-if="border" class="card-border" :style="{ '--border-color': borderColor }"></div>
  </div>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'

interface Props {
  variant?: 'default' | 'glass' | 'gradient' | 'elevated' | 'outlined' | 'flat'
  size?: 'sm' | 'md' | 'lg'
  title?: string
  subtitle?: string
  icon?: string | Component
  noPadding?: boolean
  hoverable?: boolean
  clickable?: boolean
  glow?: boolean
  glowColor?: string
  border?: boolean
  borderColor?: string
  shadow?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'glass',
  size: 'md',
  noPadding: false,
  hoverable: true,
  clickable: false,
  glow: false,
  glowColor: 'var(--color-primary)',
  border: false,
  borderColor: 'var(--color-primary)',
  shadow: false
})

const emit = defineEmits<{
  click: [event: MouseEvent]
}>()

const classes = computed(() => [
  `variant-${props.variant}`,
  `size-${props.size}`,
  {
    'is-hoverable': props.hoverable,
    'is-clickable': props.clickable,
    'has-glow': props.glow,
    'has-border': props.border,
    'has-shadow': props.shadow
  }
])

const cardStyle = computed(() => ({}))

function handleClick(event: MouseEvent) {
  if (props.clickable) {
    emit('click', event)
  }
}
</script>

<style scoped>
/* ==================== 基础样式 ==================== */
.base-card {
  position: relative;
  overflow: hidden;
  border-radius: var(--radius-lg);
  transition: all var(--transition-base);
}

.base-card.is-clickable {
  cursor: pointer;
}

/* ==================== 尺寸变体 ==================== */
.size-sm {
  border-radius: var(--radius-sm);
}

.size-md {
  border-radius: var(--radius-lg);
}

.size-lg {
  border-radius: var(--radius-xl);
}

/* ==================== 变体样式 ==================== */

/* Glass - 玻璃态（默认） */
.variant-glass {
  background: var(--color-bg-elevated);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid var(--color-border-primary);
}

.variant-glass.is-hoverable:hover {
  background: rgba(255, 255, 255, 0.12);
  border-color: rgba(255, 255, 255, 0.2);
  box-shadow: var(--shadow-lg), 0 0 30px rgba(99, 102, 241, 0.1);
  transform: translateY(-2px);
}

.variant-glass.is-clickable:active {
  transform: translateY(0) scale(0.98);
}

/* Default - 默认 */
.variant-default {
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border-primary);
}

.variant-default.is-hoverable:hover {
  background: var(--color-bg-tertiary);
  box-shadow: var(--shadow-md);
}

/* Elevated - 浮起效果 */
.variant-elevated {
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border-primary);
  box-shadow: var(--shadow-lg);
}

.variant-elevated.is-hoverable:hover {
  box-shadow: var(--shadow-xl);
  transform: translateY(-4px);
}

/* Outlined - 轮廓样式 */
.variant-outlined {
  background: transparent;
  border: 1px solid var(--color-border-primary);
}

.variant-outlined.is-hoverable:hover {
  background: var(--color-bg-elevated);
  border-color: var(--color-border-focus);
}

/* Flat - 扁平样式 */
.variant-flat {
  background: transparent;
  border: none;
}

.variant-flat.is-hoverable:hover {
  background: var(--color-bg-elevated);
}

/* Gradient - 渐变样式 */
.variant-gradient {
  background: linear-gradient(135deg,
    rgba(99, 102, 241, 0.3),
    rgba(168, 85, 247, 0.3)
  );
  border: 1px solid rgba(255, 255, 255, 0.15);
}

.variant-gradient.is-hoverable:hover {
  background: linear-gradient(135deg,
    rgba(99, 102, 241, 0.4),
    rgba(168, 85, 247, 0.4)
  );
  border-color: rgba(168, 85, 247, 0.4);
}

/* ==================== 头部 ==================== */
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border-secondary);
}

.header-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.header-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  border-radius: var(--radius-md);
  color: white;
  font-size: 18px;
}

.header-icon svg {
  width: 20px;
  height: 20px;
}

.icon-string {
  font-size: 20px;
}

.header-text {
  flex: 1;
}

.card-title {
  font-size: var(--text-lg);
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.card-subtitle {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  margin: 2px 0 0 0;
}

.header-extra {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

/* ==================== 内容区域 ==================== */
.card-body {
  padding: var(--spacing-lg);
}

.card-body.no-padding {
  padding: 0;
}

/* ==================== 底部 ==================== */
.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md) var(--spacing-lg);
  border-top: 1px solid var(--color-border-secondary);
  background: rgba(0, 0, 0, 0.1);
}

/* ==================== 装饰效果 ==================== */
.card-glow {
  position: absolute;
  inset: -1px;
  border-radius: inherit;
  opacity: 0;
  transition: opacity var(--transition-base);
  pointer-events: none;
  z-index: -1;
  background: linear-gradient(135deg,
    var(--glow-color),
    transparent 50%,
    transparent
  );
  filter: blur(15px);
}

.has-glow .card-glow {
  opacity: 0.6;
}

.is-hoverable.has-glow:hover .card-glow {
  opacity: 1;
}

.card-border {
  position: absolute;
  inset: 0;
  border-radius: inherit;
  pointer-events: none;
  z-index: 1;
}

.card-border::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: inherit;
  padding: 1px;
  background: linear-gradient(135deg,
    var(--border-color),
    transparent 50%,
    var(--border-color)
  );
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
  opacity: 0.6;
}

.has-shadow {
  box-shadow: var(--shadow-lg);
}

/* ==================== 响应式 ==================== */
@media (max-width: 640px) {
  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }

  .header-extra {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
