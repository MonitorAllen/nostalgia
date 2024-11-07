import axios from 'axios'
import type { AxiosRequestConfig } from 'axios'
import storageService from '@/service/storageService'
import { ref } from 'vue'

// 扩展 AxiosRequestConfig 接口，添加自定义属性 skipAuth
declare module 'axios' {
  export interface AxiosRequestConfig {
    skipAuth?: boolean;
  }
}

// 锁，用来避免并发请求同时触发刷新
const isRefreshing = ref(false)
let subscribers: any[] = []

const service = axios.create({
  baseURL: import.meta.env.VITE_APP_BASE_URL,
  timeout: 1000 * 5,
  headers: { Authorization: `Bearer ${storageService.get(storageService.USER_TOKEN)}` }
})

// Add a request interceptor
service.interceptors.request.use(
  async (config) => {
    if ((config as AxiosRequestConfig).skipAuth) {
      return config
    }

    // 检查 access_token 是否即将过期
    if (isTokenExpiring()) {
      if (!isRefreshing.value) {
        // 如果没有在刷新 token，则刷新
        isRefreshing.value = true
        await refreshToken() // 刷新 token 的逻辑
      } else {
        // 如果正在刷新，等待刷新完成
        await new Promise((resolve) => {
          subscribers.push(resolve)
        })
      }
    }

    // await refreshTokenIfNeeded();
    Object.assign(config.headers, {
      Authorization: `Bearer ${storageService.get(storageService.USER_TOKEN)}`
    })
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 检查 token 是否即将过期
function isTokenExpiring() {
  const tokenExpiresAt = new Date(storageService.get(storageService.USER_TOKEN_EXPIRES_AT))

  const now = new Date()

  return ((tokenExpiresAt.getTime() - now.getTime()) / 1000) < 300 // 如果 access_token 剩余时间少于 5 分钟
}

// 使用 refresh_token 刷新 token
async function refreshToken() {
  const refreshTokenExpiresAt = new Date(storageService.get(storageService.USER_REFRESH_TOKEN_EXPIRES_AT))
  const now = new Date()

  // 判断 refresh_token 是否过期
  if (now >= refreshTokenExpiresAt) {
    console.error('refresh_token 过期')
    logout()
    return
  }

  try {
    const controller = new AbortController()
    const timeoutId = setTimeout(() => controller.abort(), 5000) // 5秒超时

    console.log('刷新token')
    const response = await axios.post('/tokens/renew_access', {
      refresh_token: storageService.get(storageService.USER_REFRESH_TOKEN)
    }, { signal: controller.signal })

    clearTimeout(timeoutId)  // 请求成功，清除超时

    // 刷新成功，更新 token 和过期时间
    storageService.set(storageService.USER_TOKEN, response.data.access_token)
    storageService.set(storageService.USER_TOKEN_EXPIRES_AT, response.data.access_token_expires_at)
    // 通知所有等待请求 token 刷新完毕
    subscribers.forEach((callback) => callback());
    subscribers = [];
    console.log('刷新成功')
  } catch (error) {
    console.error('刷新 token 失败，登出' + error)
    logout()
  } finally {
    isRefreshing.value = false;
  }

}

function logout() {
  // 清除 token
  storageService.set(storageService.USER_TOKEN, '')
  storageService.set(storageService.USER_TOKEN_EXPIRES_AT, '')
  storageService.set(storageService.USER_TOKEN_EXPIRES_AT, '')
  storageService.set(storageService.USER_REFRESH_TOKEN_EXPIRES_AT, '')
  storageService.set(storageService.USER_INFO, null)

  window.location.reload()
}

export default service
