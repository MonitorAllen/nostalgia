import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginUser from '@/components/LoginUser.vue'
import EditorView from '@/views/article/EditorView.vue'
import ArticleView from '@/views/article/ArticleView.vue'
import RegisterUser from '@/components/RegisterUser.vue'
import VerifyEmail from '@/views/auth/VerifyEmail.vue'
import NotFound from '@/views/NotFound.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView
    },
    {
      path: '/login',
      name: 'login',
      component: LoginUser,
      meta: { hideNavbar: true } // 隐藏导航栏
    },
    {
      path: '/register',
      name: 'register',
      component: RegisterUser,
      meta: { hideNavbar: true } // 隐藏导航栏
    },
    {
      path: '/article/edit/:id?',
      name: 'editor',
      component: EditorView,
      props: true,
    },
    {
      path: '/article/:id?',
      name: 'articleView',
      component: ArticleView,
      props: true,
    },
    {
      path: '/auth/verifyEmail/:token',
      name: 'verifyEmail',
      component: VerifyEmail,
      props: true,
      meta: { hideNavbar: true }
    },
    { 
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: NotFound,
      meta: { hideNavbar: true }
    },
  ]
})

export default router
