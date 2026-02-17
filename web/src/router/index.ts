import { createRouter, createWebHistory, RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { requiresAuth: false, title: '登录' }
  },
  {
    path: '/',
    name: 'Platform',
    component: () => import('@/views/platform/index.vue'),
    meta: { requiresAuth: true },
    redirect: '/chat',
    children: [
      {
        path: '/chat',
        name: 'Chat',
        component: () => import('@/views/chat/ChatView.vue'),
        meta: { title: '聊天' }
      },
      {
        path: '/chat/create',
        name: 'CreateChat',
        component: () => import('@/views/creatChat/creatChat.vue'),
        meta: { title: '创建对话' }
      },
      {
        path: '/knowledge',
        name: 'KnowledgeList',
        component: () => import('@/views/knowledge/KnowledgeBaseList.vue'),
        meta: { title: '知识库列表' }
      },
      {
        path: '/knowledge/:id',
        name: 'KnowledgeDetail',
        component: () => import('@/views/knowledge/KnowledgeBase.vue'),
        meta: { title: '知识库详情' }
      },
      {
        path: '/graphs',
        name: 'GraphList',
        component: () => import('@/views/knowledge/GraphList.vue'),
        meta: { title: '知识图谱列表' }
      },
      {
        path: '/graphs/:kbId',
        name: 'GraphView',
        component: () => import('@/views/knowledge/GraphView.vue'),
        meta: { title: '知识图谱视图' }
      },
      {
        path: '/settings',
        name: 'Settings',
        component: () => import('@/views/settings/Settings.vue'),
        meta: { title: '设置' }
      },
      {
        path: '/agent',
        name: 'AgentList',
        component: () => import('@/views/agent/AgentList.vue'),
        meta: { title: 'Agent管理' }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    redirect: '/chat'
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth !== false)

  // 设置页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - Link`
  }

  if (requiresAuth && !authStore.isAuthenticated) {
    // 需要认证但未登录，跳转到登录页
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if (to.name === 'Login' && authStore.isAuthenticated) {
    // 已登录用户访问登录页，跳转到平台首页
    next({ path: '/' })
  } else {
    next()
  }
})

export default router
