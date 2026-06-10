import {createRouter, createWebHistory} from 'vue-router'
import HomeView from '../views/HomeView.vue'
import LoginUser from '@/components/LoginUser.vue'
import ArticleView from '@/views/article/ArticleView.vue'
import RegisterUser from '@/components/RegisterUser.vue'
import VerifyEmail from '@/views/auth/VerifyEmail.vue'
import NotFound from '@/views/NotFound.vue'
import MainLayout from "@/views/layout/MainLayout.vue";
import CategoryArticleView from "@/views/category/CategoryArticleView.vue";
import Forbidden from "@/views/Forbidden.vue";
import { useAdminAuthStore } from '@/admin/stores/adminAuth'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            component: MainLayout,
            children: [
                {
                    path: '',
                    name: 'home',
                    component: HomeView
                },
                {
                    path: 'category/:id',
                    name: 'categoryArticle',
                    component: CategoryArticleView,
                    props: (route) => ({
                        id: Number(route.params.id) // 在这里转换为 number
                    })
                },
                {
                    path: 'search',
                    name: 'search',
                    component: () => import('@/views/article/SearchArticleView.vue')
                }
            ]
        },
        {
            path: '/login',
            name: 'login',
            component: LoginUser,
            meta: {hideNavbar: true} // 隐藏导航栏
        },
        {
            path: '/register',
            name: 'register',
            component: RegisterUser,
            meta: {hideNavbar: true} // 隐藏导航栏
        },
        {
            path: '/article/:id?',
            name: 'articleView',
            component: ArticleView,
            props: true,
        },
        {
            path: '/auth/verifyEmail/:email_id/:secret_code',
            name: 'verifyEmail',
            component: VerifyEmail,
            props: true,
            meta: {hideNavbar: true}
        },
        {
            path: '/403',
            name: 'Forbidden',
            component: Forbidden,
            meta: {hideNavbar: true}
        },
        {
            path: '/admin/login',
            name: 'adminLogin',
            component: () => import('@/views/admin/AdminLoginView.vue'),
            meta: { hideNavbar: true, hideFooter: true },
        },
        {
            path: '/admin',
            component: () => import('@/views/admin/AdminLayout.vue'),
            meta: { hideNavbar: true, hideFooter: true, requiresAdmin: true },
            children: [
                { path: '', redirect: { name: 'adminArticles' } },
                {
                    path: 'articles',
                    name: 'adminArticles',
                    component: () => import('@/views/admin/AdminArticleListView.vue'),
                },
                {
                    path: 'articles/new',
                    name: 'adminArticleNew',
                    component: () => import('@/views/admin/AdminArticleEditorView.vue'),
                },
                {
                    path: 'articles/:id/edit',
                    name: 'adminArticleEdit',
                    component: () => import('@/views/admin/AdminArticleEditorView.vue'),
                    props: true,
                },
                {
                    path: 'categories',
                    name: 'adminCategories',
                    component: () => import('@/views/admin/AdminCategoryView.vue'),
                },
                {
                    path: ':pathMatch(.*)*',
                    name: 'adminNotFound',
                    redirect: { name: 'adminArticles' },
                },
            ],
        },
        {
            path: '/:pathMatch(.*)*',
            name: 'NotFound',
            component: NotFound,
            meta: {hideNavbar: true}
        },
    ]
})

router.beforeEach(async (to) => {
    if (!to.meta.requiresAdmin) return true

    const adminAuth = useAdminAuthStore()
    const authenticated = await adminAuth.ensureAuthenticated()

    if (authenticated) return true

    return {
        name: 'adminLogin',
        query: { redirect: to.fullPath },
    }
})

export default router
