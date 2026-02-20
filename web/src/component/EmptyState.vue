<template>
  <div class="empty-state" :class="classes">
    <!-- 图标/图片 -->
    <div class="empty-icon" :class="{ 'is-animated': animated }">
      <slot name="icon">
        <component :is="icon" v-if="typeof icon === 'object'" />
        <img v-else-if="image" :src="image" :alt="title" />
        <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M3 3v18h18M18 17V9M13 17V5M8 17v-3"/>
        </svg>
      </slot>
    </div>

    <!-- 标题 -->
    <h3 v-if="title" class="empty-title">{{ title }}</h3>

    <!-- 描述 -->
    <p v-if="description || $slots.description" class="empty-description">
      <slot name="description">
        {{ description }}
      </slot>
    </p>

    <!-- 操作按钮 -->
    <div v-if="$slots.action || actionText" class="empty-action">
      <slot name="action">
        <BaseButton :variant="actionVariant" @click="handleAction">
          {{ actionText }}
        </BaseButton>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, type Component } from 'vue'
import BaseButton from './BaseButton.vue'

interface Props {
  title?: string
  description?: string
  icon?: string | Component
  image?: string
  actionText?: string
  actionVariant?: 'primary' | 'secondary' | 'ghost' | 'gradient'
  size?: 'sm' | 'md' | 'lg'
  animated?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  actionVariant: 'primary',
  size: 'md',
  animated: true
})

const emit = defineEmits<{
  action: []
}>()

const classes = computed(() => `size-${props.size}`)

function handleAction() {
  emit('action')
}
</script>

<style scoped>
/* ==================== 基础样式 ==================== */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: var(--spacing-2xl);
  color: var(--color-text-muted);
}

/* ==================== 尺寸变体 ==================== */
.size-sm {
  padding: var(--spacing-lg);
}

.size-sm .empty-icon {
  width: 48px;
  height: 48px;
  margin-bottom: var(--spacing-md);
}

.size-sm .empty-title {
  font-size: var(--text-base);
}

.size-sm .empty-description {
  font-size: var(--text-sm);
}

.size-md {
  padding: var(--spacing-2xl);
}

.size-md .empty-icon {
  width: 80px;
  height: 80px;
  margin-bottom: var(--spacing-lg);
}

.size-md .empty-title {
  font-size: var(--text-lg);
}

.size-md .empty-description {
  font-size: var(--text-sm);
}

.size-lg {
  padding: var(--spacing-2xl) var(--spacing-2xl);
}

.size-lg .empty-icon {
  width: 120px;
  height: 120px;
  margin-bottom: var(--spacing-xl);
}

.size-lg .empty-title {
  font-size: var(--text-xl);
}

.size-lg .empty-description {
  font-size: var(--text-base);
}

/* ==================== 图标 ==================== */
.empty-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-disabled);
  opacity: 0.6;
  transition: all var(--transition-base);
}

.empty-icon svg {
  width: 100%;
  height: 100%;
}

.empty-icon img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.empty-icon.is-animated {
  animation: emptyFloat 3s ease-in-out infinite;
}

@keyframes emptyFloat {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

/* ==================== 标题 ==================== */
.empty-title {
  font-weight: 600;
  color: var(--color-text-secondary);
  margin: 0 0 var(--spacing-sm) 0;
}

/* ==================== 描述 ==================== */
.empty-description {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  max-width: 400px;
  margin: 0 0 var(--spacing-lg) 0;
  line-height: 1.6;
}

/* ==================== 操作按钮 ==================== */
.empty-action {
  display: flex;
  gap: var(--spacing-sm);
  margin-top: var(--spacing-md);
}
</style>
