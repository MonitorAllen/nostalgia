import axios from 'axios'
import type {
  AxiosError,
  AxiosRequestConfig,
  AxiosResponse,
  InternalAxiosRequestConfig
} from 'axios'
import { buildAdminLoginRedirect } from '@/admin/adminRoutes'
import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/store/module/auth'

declare module 'axios' {
  interface AxiosRequestConfig {
    skipAuth?: boolean
    skipErrorHandler?: boolean
  }
}

interface AdminInternalRequestConfig extends InternalAxiosRequestConfig {
  skipAuth?: boolean
  skipErrorHandler?: boolean
  _adminRetry?: boolean
}

class AdminHttpClient {
  private instance = axios.create({
    baseURL: '/v1',
    timeout: 10000
  })

  private refreshPromise: Promise<string> | null = null

  constructor() {
    this.setupInterceptors()
  }

  private setupInterceptors() {
    this.instance.interceptors.request.use(
      this.handleRequest.bind(this),
      this.handleRequestError.bind(this)
    )

    this.instance.interceptors.response.use(
      this.handleResponse.bind(this),
      this.handleResponseError.bind(this)
    )
  }

  private async handleRequest(config: InternalAxiosRequestConfig) {
    const adminConfig = config as AdminInternalRequestConfig

    if (adminConfig.skipAuth) {
      return adminConfig
    }

    const authStore = useAuthStore()
    const authenticated = await authStore.ensureAdminAuthenticated()

    if (!authenticated) {
      this.redirectToLogin()
      return Promise.reject(new Error('Admin authentication required'))
    }

    if (authStore.token) {
      adminConfig.headers.Authorization = `Bearer ${authStore.token}`
    }

    return adminConfig
  }

  private handleRequestError(error: unknown) {
    return Promise.reject(error)
  }

  private handleResponse(response: AxiosResponse) {
    return response
  }

  private async handleResponseError(error: AxiosError) {
    const config = error.config as AdminInternalRequestConfig | undefined

    if (error.response?.status === 401 && config && !config.skipAuth && !config._adminRetry) {
      config._adminRetry = true

      try {
        const token = await this.refreshToken()
        config.headers.Authorization = `Bearer ${token}`
        return this.instance.request(config)
      } catch (refreshError) {
        const authStore = useAuthStore()
        authStore.clearTokens()
        this.redirectToLogin()
        return Promise.reject(refreshError)
      }
    }

    if (!config?.skipErrorHandler) {
      this.handleError(error)
    }

    return Promise.reject(error)
  }

  private async refreshToken() {
    if (this.refreshPromise) {
      return this.refreshPromise
    }

    const authStore = useAuthStore()
    this.refreshPromise = authStore.refreshAccessToken()

    try {
      return await this.refreshPromise
    } finally {
      this.refreshPromise = null
    }
  }

  private handleError(error: AxiosError) {
    const toast = useToast()
    const data = error.response?.data as { error?: string; message?: string } | undefined
    const message = data?.error || data?.message || error.message || '请求失败'

    toast.add({
      severity: 'error',
      summary: '错误',
      detail: message,
      life: 3000
    })
  }

  private redirectToLogin() {
    const current = `${window.location.pathname}${window.location.search}${window.location.hash}`
    window.location.href = buildAdminLoginRedirect(current)
  }

  get<T = unknown>(url: string, config?: AxiosRequestConfig) {
    return this.instance.get<T>(url, config)
  }

  post<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) {
    return this.instance.post<T>(url, data, config)
  }

  patch<T = unknown>(url: string, data?: unknown, config?: AxiosRequestConfig) {
    return this.instance.patch<T>(url, data, config)
  }

  delete<T = unknown>(url: string, config?: AxiosRequestConfig) {
    return this.instance.delete<T>(url, config)
  }
}

export default new AdminHttpClient()
