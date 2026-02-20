<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="modelValue" class="modal-overlay" :class="{ 'has-backdrop': showBackdrop }" @click="handleOverlayClick">
        <Transition name="modal-content">
          <div
            v-if="modelValue"
            class="modal-container"
            :class="[
              `size-${size}`,
              { 'is-fullscreen': fullscreen }
            ]"
            :style="containerStyle"
            @click.stop
          >
            <!-- 头部 -->
            <div v-if="!hideHeader" class="modal-header">
              <div class="header-left">
                <div v-if="icon" class="modal-icon">
                  <component :is="icon" v-if="typeof icon === 'object'" />
                  <span v-else>{{ icon }}</span>
                </div>
                <div class="header-text">
                  <h3 class="modal-title">{{ title }}</h3>
                  <p v-if="subtitle" class="modal-subtitle">{{ subtitle }}</p>
                </div>
              </div>
              <div class="header-actions">
                <slot name="actions" />
                <button v-if="closable" class="close-button" @click="close" :title="closeText">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M18 6L6 18M6 6l12 12"/>
                  </svg>
                </button>
              </div>
            </div>

            <!-- 内容 -->
            <div class="modal-body" :class="{ 'no-padding': noPadding }">
              <slot />
            </div>

            <!-- 底部 -->
            <div v-if="!hideFooter && ($slots.footer || showDefaultFooter)" class="modal-footer">
              <slot name="footer">
                <BaseButton variant="ghost" @click="handleCancel">{{ cancelText }}</BaseButton>
                <BaseButton variant="primary" :loading="loading" @click="handleConfirm">{{ confirmText }}</BaseButton>
              </slot>
            </div>

            <!-- 装饰光效 -->
            <div v-if="glow" class="modal-glow"></div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, watch, type Component } from 'vue'
import BaseButton from './BaseButton.vue'

interface Props {
  modelValue: boolean
  title?: string
  subtitle?: string
  icon?: string | Component
  size?: 'sm' | 'md' | 'lg' | 'xl' | 'full'
  closable?: boolean
  closeOnClickModal?: boolean
  closeOnPressEscape?: boolean
  showBackdrop?: boolean
  hideHeader?: boolean
  hideFooter?: boolean
  showDefaultFooter?: boolean
  noPadding?: boolean
  fullscreen?: boolean
  loading?: boolean
  confirmText?: string
  cancelText?: string
  closeText?: string
  glow?: boolean
  destroyOnClose?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  size: 'md',
  closable: true,
  closeOnClickModal: true,
  closeOnPressEscape: true,
  showBackdrop: true,
  hideHeader: false,
  hideFooter: false,
  showDefaultFooter: true,
  noPadding: false,
  fullscreen: false,
  loading: false,
  confirmText: '确定',
  cancelText: '取消',
  closeText: '关闭',
  glow: true,
  destroyOnClose: false
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'open': []
  'close': []
  'confirm': []
  'cancel': []
}>()

const containerStyle = computed(() => ({}))

function close() {
  emit('update:modelValue', false)
  emit('close')
}

function handleConfirm() {
  emit('confirm')
}

function handleCancel() {
  emit('cancel')
  close()
}

function handleOverlayClick() {
  if (props.closeOnClickModal) {
    close()
  }
}

// 键盘事件处理
watch(() => props.modelValue, (isOpen) => {
  if (isOpen) {
    emit('open')
    document.addEventListener('keydown', handleKeydown)
  } else {
    document.removeEventListener('keydown', handleKeydown)
  }
})

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && props.closeOnPressEscape) {
    close()
  }
}
</script>

<style scoped>
/* ==================== 遮罩层 ==================== */
.modal-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-lg);
}

.modal-overlay.has-backdrop {
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
}

/* ==================== 容器 ==================== */
.modal-container {
  position: relative;
  width: 100%;
  max-height: 90vh;
  background: var(--color-bg-secondary);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  border: 1px solid var(--color-border-primary);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 尺寸变体 */
.size-sm {
  max-width: 400px;
}

.size-md {
  max-width: 540px;
}

.size-lg {
  max-width: 720px;
}

.size-xl {
  max-width: 900px;
}

.size-full {
  max-width: 100%;
  height: 100%;
  max-height: 100%;
  border-radius: 0;
}

.is-fullscreen {
  width: 100vw;
  height: 100vh;
  max-width: 100vw;
  max-height: 100vh;
  border-radius: 0;
}

/* ==================== 头部 ==================== */
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-lg) var(--spacing-xl);
  border-bottom: 1px solid var(--color-border-secondary);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex: 1;
  min-width: 0;
}

.modal-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  background: linear-gradient(135deg, var(--color-primary), var(--color-secondary));
  border-radius: var(--radius-md);
  color: white;
  font-size: 20px;
  flex-shrink: 0;
}

.modal-icon svg {
  width: 22px;
  height: 22px;
}

.header-text {
  flex: 1;
  min-width: 0;
}

.modal-title {
  font-size: var(--text-xl);
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.modal-subtitle {
  font-size: var(--text-sm);
  color: var(--color-text-muted);
  margin: 4px 0 0 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  flex-shrink: 0;
}

.close-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: transparent;
  border: none;
  border-radius: var(--radius-md);
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.close-button:hover {
  background: var(--color-bg-tertiary);
  color: var(--color-text-primary);
}

.close-button svg {
  width: 20px;
  height: 20px;
}

/* ==================== 内容 ==================== */
.modal-body {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-xl);
}

.modal-body.no-padding {
  padding: 0;
}

/* ==================== 底部 ==================== */
.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-md);
  padding: var(--spacing-lg) var(--spacing-xl);
  border-top: 1px solid var(--color-border-secondary);
  flex-shrink: 0;
}

/* ==================== 光效 ==================== */
.modal-glow {
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(
    circle at center,
    rgba(99, 102, 241, 0.15) 0%,
    transparent 50%
  );
  pointer-events: none;
  z-index: -1;
  animation: modalGlow 8s ease-in-out infinite;
}

@keyframes modalGlow {
  0%, 100% {
    transform: translate(0, 0) scale(1);
  }
  33% {
    transform: translate(10%, -10%) scale(1.1);
  }
  66% {
    transform: translate(-10%, 10%) scale(0.9);
  }
}

/* ==================== 过渡动画 ==================== */
.modal-enter-active,
.modal-leave-active {
  transition: all var(--transition-base);
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-active .modal-container,
.modal-leave-active .modal-container {
  transition: all var(--transition-base);
}

.modal-enter-from .modal-container,
.modal-leave-to .modal-container {
  opacity: 0;
  transform: scale(0.95) translateY(20px);
}

/* 内容过渡 */
.modal-content-enter-active {
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.modal-content-leave-active {
  transition: all 0.25s ease-out;
}

.modal-content-enter-from {
  opacity: 0;
  transform: scale(0.9) translateY(-30px);
}

.modal-content-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(10px);
}

/* ==================== 响应式 ==================== */
@media (max-width: 640px) {
  .modal-overlay {
    padding: 0;
  }

  .modal-container {
    max-width: 100vw;
    max-height: 100vh;
    border-radius: 0;
    height: 100%;
  }

  .modal-header {
    padding: var(--spacing-md) var(--spacing-lg);
  }

  .modal-body {
    padding: var(--spacing-lg);
  }

  .modal-footer {
    padding: var(--spacing-md) var(--spacing-lg);
    flex-direction: column-reverse;
  }

  .modal-footer :deep(.base-button) {
    width: 100%;
  }
}
</style>
