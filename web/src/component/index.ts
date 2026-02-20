// ==================== 基础组件 ====================
export { default as BaseButton } from './BaseButton.vue'
export { default as BaseCard } from './BaseCard.vue'
export { default as BaseInput } from './BaseInput.vue'
export { default as BaseModal } from './BaseModal.vue'
export { default as BaseTag } from './BaseTag.vue'

// ==================== 数据组件 ====================
export { default as BaseTable } from './BaseTable.vue'
export { default as BaseLoader } from './BaseLoader.vue'
export { default as EmptyState } from './EmptyState.vue'

// ==================== 布局组件 ====================
export { default as BaseSidebar } from './BaseSidebar.vue'
export { default as AppBackground } from './AppBackground.vue'
export { default as AppLayout } from './AppLayout.vue'

// ==================== 样式 ====================
export { default as GlobalStyles } from './styles/global.css'

// ==================== 类型 ====================
export type * from './types'

// ==================== 便捷插件 ====================
import type { App } from 'vue'

// 自动注册所有组件
export function installComponents(app: App) {
  // 基础组件
  app.component('BaseButton', BaseButton)
  app.component('BaseCard', BaseCard)
  app.component('BaseInput', BaseInput)
  app.component('BaseModal', BaseModal)
  app.component('BaseTag', BaseTag)

  // 数据组件
  app.component('BaseTable', BaseTable)
  app.component('BaseLoader', BaseLoader)
  app.component('EmptyState', EmptyState)

  // 布局组件
  app.component('BaseSidebar', BaseSidebar)
  app.component('AppBackground', AppBackground)
  app.component('AppLayout', AppLayout)
}

// 默认导出用于 app.use()
export default {
  install: installComponents
}
