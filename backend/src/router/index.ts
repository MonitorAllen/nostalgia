import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import LoginView from '@/views/LoginView.vue'
import AdminLayout from '@/layouts/AdminLayout.vue'
import EditorView from '@/views/EditorView.vue'

const router = createRouter({
  history: createWebHistory('/backend'),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView
    },
    {
      path: '/article/edit/:id?',
      name: 'editor',
      component: EditorView,
      props: true,
      meta: { requiresAuth: true }
    },
    {
      path: '/',
      component: AdminLayout,
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('@/views/HomeView.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'manage/articles',
          name: 'articles-management',
          component: () => import('@/views/ArticlesManagementView.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'admin/update',
          name: 'update-admin',
          component: () => import('@/views/UpdateAdminView.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'manage/categories',
          name: 'categories-management',
          component: () => import('@/views/CategoryManagementView.vue'),
          meta: { requiresAuth: true }
        },
      ]
    }
  ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else {
    next()
  }
})

export default router
