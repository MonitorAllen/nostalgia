// noinspection DuplicatedCode

import axios, {type AxiosResponse, type InternalAxiosRequestConfig} from 'axios'
import {useAuthStore} from '@/stores/auth'
import {useToast} from 'primevue/usetoast'
import {API_BASE_URL} from "@/config";

// 扩展请求配置
declare module 'axios' {
    interface AxiosRequestConfig {
        skipAuth?: boolean
        skipErrorHandler?: boolean
        _isRetry?: boolean // 标记是否为重试请求
    }
}

class HttpClient {
    private instance = axios.create({
        baseURL: API_BASE_URL,
        timeout: 10000,
        skipAuth: false,
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

        // 401 错误且不是刷新 token 的请求
        if (response?.status === 401 && !config.skipAuth && !config._isRetry) {
            try {
                const token = await this.refreshToken()
                // 重试原请求
                config.headers.Authorization = `Bearer ${token}`
                return this.instance.request(config)
            } catch (error) {
                // 刷新 token 失败，执行退出登录
                console.error('Token refresh failed:', error)
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

        // 如果没有 token，直接返回
        if (!authStore.accessToken) {
          return null
        }

        // 检查 token 是否需要刷新
        if (this.shouldRefreshToken()) {
          try {
            return await this.refreshToken()
          } catch (error) {
            console.error('Token refresh failed in getValidToken:', error)
            // 刷新失败，返回当前 token，让后续请求去处理 401
            return authStore.accessToken
          }
        }

        return authStore.accessToken
    }

    // 是否需要刷新 token
    private shouldRefreshToken(): boolean {
        const authStore = useAuthStore()
        const expiresAt = authStore.accessTokenExpiresAt

        if (!expiresAt || !authStore.accessToken) return false

        // 提前 2 分钟刷新
        const twoMinutesFromNow = Date.now() + 2 * 60 * 1000
        return new Date(expiresAt).getTime() < twoMinutesFromNow
    }

    private async refreshToken(): Promise<string> {
        // 防止并发刷新
        if (this.refreshPromise) {
            return this.refreshPromise
        }

        this.refreshPromise = this.performTokenRefresh()

        try {
            return await this.refreshPromise
        } catch (error) {
          throw error
        } finally {
            this.refreshPromise = null
        }
    }

    private async performTokenRefresh(): Promise<string> {
        const authStore = useAuthStore()

        try {
          await authStore.refreshAccessToken()
          if (!authStore.accessToken) {
            throw new Error('No access token after refresh')
          }
          return authStore.accessToken
        } catch (error) {
          console.error('Failed to refresh token:', error)
          // 刷新失败，清除本地认证状态
          authStore.clearAuth() // 假设你有这个方法
          throw error
        }


    }

    private handleError(error: any) {
        const toast = useToast()
        let message = error.response?.data?.error || '请求失败'

      // 根据不同的错误状态码提供更友好的提示
      if (error.response) {
        const status = error.response.status
        switch (status) {
          case 400:
            message = error.response.data?.error || '请求参数错误'
            break
          case 401:
            message = '未授权访问'
            break
          case 403:
            message = '没有权限访问此资源'
            break
          case 404:
            message = '请求的资源不存在'
            break
          case 422:
            message = error.response.data?.error || '数据验证失败'
            break
          case 500:
            message = '服务器内部错误'
            break
          default:
            message = error.response.data?.error || `请求失败 (${status})`
        }
      } else if (error.request) {
        message = '网络请求失败，请检查网络连接'
      }

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
