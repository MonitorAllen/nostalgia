import axios from 'axios'
import { API_BASE_URL } from './index'
import { useRouter } from 'vue-router'
import {STORAGE_KEYS} from "@/stores/auth.ts";

const axiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 5000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 判断 token 是否即将过期（比如还有 5 分钟过期）
const isTokenNearExpiry = () => {
  const expiryTime = localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN_EXPIRES_AT)
  if (!expiryTime) return true

  const expiry = new Date(expiryTime).getTime()
  const currentTime = Date.now()
  const timeUntilExpiry = expiry - currentTime
  return timeUntilExpiry < 2 * 60 * 1000 // 小于 2 分钟就刷新
}

// 刷新 token
const refreshToken = async () => {
  const refresh_token = localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN)
  const refreshTokenExpiry = localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN_EXPIRES_AT)

  if (!refresh_token || !refreshTokenExpiry || new Date(refreshTokenExpiry).getTime() < Date.now()) {
    throw new Error('Refresh token expired')
  }

  try {
    const response = await axios.post(`${API_BASE_URL}/admin/renew_access`, {refresh_token: refresh_token})

    const {
      access_token,
      access_token_expires_at,
    } = response.data

    localStorage.setItem('access_token', access_token)
    localStorage.setItem('access_token_expires_at', access_token_expires_at)
    return access_token
  } catch {
    throw new Error('Failed to refresh token')
  }
}

let isRefreshing = false
let refreshSubscribers: ((token: string) => void)[] = []

// 订阅 token 刷新
const subscribeTokenRefresh = (cb: (token: string) => void) => {
  refreshSubscribers.push(cb)
}

// 执行所有订阅
const onRefreshed = (token: string) => {
  refreshSubscribers.forEach(cb => cb(token))
  refreshSubscribers = []
}

// 请求拦截器
axiosInstance.interceptors.request.use(
  async (config) => {
    const token = localStorage.getItem('access_token')
    if (token) {
      if (isTokenNearExpiry() && !config.url?.includes('/auth/refresh_token')) {
        try {
          const newToken = await refreshToken()
          config.headers.Authorization = `Bearer ${newToken}`
        } catch (error) {
          // 如果刷新失败，继续使用旧 token
          config.headers.Authorization = `Bearer ${token}`
        }
      } else {
        config.headers.Authorization = `Bearer ${token}`
      }
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
axiosInstance.interceptors.response.use(
  (response) => {
    return response
  },
  async (error) => {
    const originalRequest = error.config

    // 如果是未认证错误且不是刷新 token 的请求
    if (error.response?.status === 401 && !originalRequest._retry && !originalRequest.url.includes('/admin/renew_access')) {
      if (isRefreshing) {
        // 如果正在刷新，将请求添加到队列
        try {
          const token = await new Promise<string>((resolve, reject) => {
            subscribeTokenRefresh(resolve)
          })
          originalRequest.headers.Authorization = `Bearer ${token}`
          return axiosInstance(originalRequest)
        } catch (err) {
          const router = useRouter()
          router.push('/login')
          return Promise.reject(err)
        }
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        const newToken = await refreshToken()
        isRefreshing = false
        onRefreshed(newToken)
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return axiosInstance(originalRequest)
      } catch (refreshError) {
        isRefreshing = false
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        localStorage.removeItem('access_token_expires_at')
        localStorage.removeItem('refresh_token_expires_at')
        const router = useRouter()
        router.push('/login')
        return Promise.reject(refreshError)
      }
    }

    return Promise.reject(error)
  }
)

export default axiosInstance
