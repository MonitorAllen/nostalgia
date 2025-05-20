import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import LoginView from '@/views/LoginView.vue'
import AdminLayout from '@/layouts/AdminLayout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView
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
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('@/views/DashboardView.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'manage/articles',
          name: 'articles-management',
          component: () => import('@/views/ArticlesManagementView.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'articles/new',
          name: 'new-article',
          component: () => import('@/views/NewArticleView.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'categories',
          name: 'categories',
          component: () => import('@/views/CategoriesView.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'comments',
          name: 'comments',
          component: () => import('@/views/CommentsView.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'settings/basic',
          name: 'basic-settings',
          component: () => import('@/views/BasicSettingsView.vue'),
          meta: { requiresAuth: true }
        },
        {
          path: 'manage/users',
          name: 'user-management',
          component: () => import('@/views/UserManagementView.vue'),
          meta: { requiresAuth: true }
        }
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
