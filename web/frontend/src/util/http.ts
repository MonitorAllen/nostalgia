// utils/http.ts - 简化版本
import axios, {type AxiosResponse, type InternalAxiosRequestConfig} from 'axios'
import {useAuthStore} from '@/store/module/auth'
import {useToast} from 'primevue/usetoast'

// 扩展请求配置
declare module 'axios' {
    interface AxiosRequestConfig {
        skipAuth?: boolean
        skipErrorHandler?: boolean
    }
}

class HttpClient {
    private instance = axios.create({
        baseURL: import.meta.env.VITE_APP_BASE_URL,
        timeout: 10000,
        skipAuth: true,
        skipErrorHandler: true
    })

    private refreshPromise: Promise<string> | null = null

    constructor() {
        this.setupInterceptors()
    }

    private setupInterceptors() {
        // 请求拦截器
        this.instance.interceptors.request.use(
            this.handleRequest.bind(this),
            this.handleRequestError.bind(this)
        )

        // 响应拦截器
        this.instance.interceptors.response.use(
            this.handleResponse.bind(this),
            this.handleResponseError.bind(this)
        )
    }

    private async handleRequest(config: InternalAxiosRequestConfig) {
        // 跳过认证的请求
        if (config.skipAuth) {
            return config
        }

        // 获取有效 token
        const token = await this.getValidToken()
        if (token) {
            config.headers.Authorization = `Bearer ${token}`
        }

        return config
    }

    private handleRequestError(error: any) {
        return Promise.reject(error)
    }

    private handleResponse(response: AxiosResponse) {
        return response
    }

    private async handleResponseError(error: any) {
        const { config, response } = error

        if (response?.status === 403) {
            window.location.href = '/403'
        }

        // 401 错误且不是刷新 token 的请求
        if (response?.status === 401 && !config.skipAuth) {
            try {
                const token = await this.refreshToken()
                // 重试原请求
                config.headers.Authorization = `Bearer ${token}`
                return this.instance.request(config)
            } catch {
                this.handleLogout()
                return Promise.reject(error)
            }
        }

        // 统一错误处理
        if (!config.skipErrorHandler) {
            this.handleError(error)
        }

        return Promise.reject(error)
    }

    private async getValidToken(): Promise<string | null> {
        const authStore = useAuthStore()

        // 检查 token 是否需要刷新
        if (this.shouldRefreshToken()) {
            return await this.refreshToken()
        }

        return authStore.token
    }

    // 是否需要刷新 token
    private shouldRefreshToken(): boolean {
        const authStore = useAuthStore()
        const expiresAt = authStore.tokenExpiresAt

        if (!expiresAt) return false

        // 提前 2 分钟刷新
        const fiveMinutesFromNow = Date.now() + 2 * 60 * 1000
        return new Date(expiresAt).getTime() < fiveMinutesFromNow
    }

    private async refreshToken(): Promise<string> {
        // 防止并发刷新
        if (this.refreshPromise) {
            return this.refreshPromise
        }

        this.refreshPromise = this.performTokenRefresh()

        try {
            return await this.refreshPromise
        } finally {
            this.refreshPromise = null
        }
    }

    private async performTokenRefresh(): Promise<string> {
        const authStore = useAuthStore()

        return await authStore.refreshAccessToken()
    }

    private handleError(error: any) {
        const toast = useToast()
        const message = error.response?.data?.error || '请求失败'

        toast.add({
            severity: 'error',
            summary: '错误',
            detail: message,
            life: 3000
        })
    }

    private handleLogout() {
        const authStore = useAuthStore()
        authStore.logout()

        // 跳转到登录页
        window.location.href = '/login'
    }

    // 公共方法
    get<T = any>(url: string, config?: any) {
        return this.instance.get<T>(url, config)
    }

    post<T = any>(url: string, data?: any, config?: any) {
        return this.instance.post<T>(url, data, config)
    }

    put<T = any>(url: string, data?: any, config?: any) {
        return this.instance.put<T>(url, data, config)
    }

    patch<T = any>(url: string, data?: any, config?: any) {
        return this.instance.patch<T>(url, data, config)
    }

    delete<T = any>(url: string, config?: any) {
        return this.instance.delete<T>(url, config)
    }
}

export default new HttpClient()