<template>
  <!-- 全局背景层 -->
  <div class="app-background">
    <div class="bg-image"></div>
    <div class="bg-overlay"></div>
  </div>

  <!-- 应用内容 -->
  <router-view />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

onMounted(() => {
  // 初始化时检查本地存储的Token
  authStore.checkAuth()
})
</script>

<style>
/* ==================== 全局重置 ==================== */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  width: 100%;
  height: 100%;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
  color: #ffffff;
  background: transparent;
  overflow-x: hidden;
}

body {
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* ==================== CSS 变量 ==================== */
:root {
  /* 主色调 - 基于背景图片的粉蓝色调 */
  --color-primary: #ff7eb3;
  --color-primary-light: #ffb3d9;
  --color-primary-dark: #ff5a9f;

  /* 辅助色 */
  --color-secondary: #7afcff;
  --color-accent: #b8b8ff;
  --color-success: #7affc3;
  --color-warning: #ffd97d;
  --color-danger: #ff7e7e;

  /* 中性色 - 玻璃态半透明 */
  --color-bg-primary: rgba(0, 0, 0, 0.4);
  --color-bg-secondary: rgba(0, 0, 0, 0.3);
  --color-bg-tertiary: rgba(0, 0, 0, 0.2);
  --color-bg-elevated: rgba(255, 255, 255, 0.1);
  --color-bg-hover: rgba(255, 255, 255, 0.15);

  /* 文字颜色 */
  --color-text-primary: #ffffff;
  --color-text-secondary: rgba(255, 255, 255, 0.85);
  --color-text-muted: rgba(255, 255, 255, 0.6);
  --color-text-disabled: rgba(255, 255, 255, 0.4);

  /* 边框颜色 */
  --color-border-primary: rgba(255, 255, 255, 0.2);
  --color-border-secondary: rgba(255, 255, 255, 0.1);
  --color-border-focus: var(--color-primary);

  /* 阴影 */
  --shadow-sm: 0 2px 8px rgba(0, 0, 0, 0.1);
  --shadow-md: 0 4px 12px rgba(0, 0, 0, 0.15);
  --shadow-lg: 0 8px 24px rgba(0, 0, 0, 0.2);
  --shadow-xl: 0 12px 32px rgba(0, 0, 0, 0.25);
  --shadow-glow: 0 0 20px rgba(255, 126, 179, 0.4);

  /* 圆角 */
  --radius-sm: 8px;
  --radius-md: 12px;
  --radius-lg: 18px;
  --radius-xl: 24px;
  --radius-full: 9999px;

  /* 间距 */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 12px;
  --spacing-lg: 16px;
  --spacing-xl: 24px;
  --spacing-2xl: 32px;

  /* 过渡 */
  --transition-fast: 150ms ease;
  --transition-base: 250ms ease;
  --transition-slow: 350ms ease;

  /* 字体 */
  --font-sans: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  --font-mono: 'JetBrains Mono', 'Fira Code', Consolas, monospace;

  /* 字号 */
  --text-xs: 0.75rem;
  --text-sm: 0.875rem;
  --text-base: 1rem;
  --text-lg: 1.125rem;
  --text-xl: 1.25rem;
  --text-2xl: 1.5rem;
  --text-3xl: 1.875rem;
}

/* ==================== 全局背景 ==================== */
.app-background {
  position: fixed;
  inset: 0;
  z-index: -1;
  overflow: hidden;
}

.bg-image {
  position: absolute;
  inset: 0;
  background-image: url('@/wallhaven-d83m63_1920x1080.png');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
}

.bg-overlay {
  position: absolute;
  inset: 0;
  /* 轻微暗色遮罩，让背景图可见 */
  background: rgba(0, 0, 0, 0.15);
}

/* ==================== 滚动条样式 ==================== */
* {
  scrollbar-width: thin;
  scrollbar-color: rgba(255, 255, 255, 0.2) transparent;
}

::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.3);
}

/* ==================== 选中文本样式 ==================== */
::selection {
  background: var(--color-primary);
  color: white;
}

::-moz-selection {
  background: var(--color-primary);
  color: white;
}

/* ==================== 链接样式 ==================== */
a {
  color: var(--color-primary-light);
  text-decoration: none;
  transition: color var(--transition-fast);
}

a:hover {
  color: var(--color-primary);
}

/* ==================== 焦点样式 ==================== */
:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

/* ==================== 工具类 ==================== */
.glass {
  background: var(--color-bg-elevated);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid var(--color-border-primary);
}

.glass-dark {
  background: var(--color-bg-secondary);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  border: 1px solid var(--color-border-primary);
}

.text-gradient {
  background: linear-gradient(135deg, var(--color-primary-light), var(--color-secondary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* ==================== 动画 ==================== */
@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.8; }
}

@keyframes glow {
  0%, 100% {
    box-shadow: 0 0 5px var(--color-primary), 0 0 10px var(--color-primary);
  }
  50% {
    box-shadow: 0 0 15px var(--color-primary), 0 0 25px var(--color-primary);
  }
}

/* ==================== 响应式 ==================== */
@media (max-width: 768px) {
  .bg-image {
    background-position: center center;
  }
}
</style>
