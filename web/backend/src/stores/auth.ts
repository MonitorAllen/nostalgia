// noinspection DuplicatedCode

import { defineStore } from 'pinia'
import {computed, readonly, ref} from 'vue'
import axiosInstance from '../config/axios'
import http from "@/util/http.ts";

interface LoginForm {
  username: string
  password: string
}

export interface Admin {
  id: number
  username: string
  email: string
  role: string
  avatar?: string
}

export const STORAGE_KEYS = {
  ACCESS_TOKEN: 'nostalgia_backend_access_token',
  ACCESS_TOKEN_EXPIRES_AT: 'nostalgia_backend_token_expires_at',
  REFRESH_TOKEN: 'nostalgia_backend_refresh_token',
  REFRESH_TOKEN_EXPIRES_AT: 'nostalgia_backend_refresh_token_expires_at',
} as const

export const useAuthStore = defineStore('auth', () => {
  const admin = ref<Admin | null>(null)
  const accessToken = ref(localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN) || '')
  const accessTokenExpiresAt = ref(localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN) || '')
  const refreshToken = ref(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN) || '')
  const refreshTokenExpiresAt = ref(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN_EXPIRES_AT) || '')
  const isAuthenticated = ref(false)

  const isRefreshTokenValid = computed(() => {
    if (!refreshTokenExpiresAt.value) return false
    const expiresAt = new Date(refreshTokenExpiresAt.value)
    return new Date() < expiresAt
  })

  // 初始化认证状态
  const initAuth = () => {
    const storedAccessToken = localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN)
    const storedAccessTokenExpiresAt = localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN_EXPIRES_AT)

    if (storedAccessToken && storedAccessTokenExpiresAt) {
      // 检查 token 是否过期
      const expiresAt = new Date(storedAccessTokenExpiresAt).getTime()
      const now = new Date().getTime()

      if (now < expiresAt) {
        // token 未过期，恢复认证状态
        accessToken.value = storedAccessToken
        refreshToken.value = localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN) || ''
        isAuthenticated.value = true
        // 设置 axios 默认请求头
        axiosInstance.defaults.headers.common['Authorization'] = `Bearer ${storedAccessToken}`

        // 获取用户信息
        fetchAdminInfo()
      } else {
        // token 已过期，清除存储的信息
        clearAuth()
      }
    }
  }

  // 获取管理员信息
  const fetchAdminInfo = async () => {
    try {
      const response = await http.get('/admin/info')
      admin.value = response.data.admin
    } catch (error: any) {
      clearAuth()
    }
  }

  // 清除认证信息
  const clearAuth = () => {
    admin.value = null
    accessToken.value = ""
    accessTokenExpiresAt.value = ''
    refreshToken.value = ""
    refreshTokenExpiresAt.value = ''
    isAuthenticated.value = false
    // 清除 localStorage
    Object.values(STORAGE_KEYS).forEach(key => {
      localStorage.removeItem(key)
    })
    delete axiosInstance.defaults.headers.common['Authorization']
  }

  const login = async (form: LoginForm) => {
    try {
      const response = await http.post('/admin/login', form, {skipAuth: true})
      const { data } = response
      accessToken.value = data.access_token
      accessTokenExpiresAt.value = data.access_token_expires_at
      refreshToken.value = data.refresh_token
      refreshTokenExpiresAt.value = data.refresh_token_expires_at
      admin.value = data.admin
      isAuthenticated.value = true
      // 保存token到localStorage
      localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, data.access_token)
      localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, data.refresh_token)
      localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN_EXPIRES_AT, data.access_token_expires_at)
      localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN_EXPIRES_AT, data.refresh_token_expires_at)
      // 设置axios默认请求头
      axiosInstance.defaults.headers.common['Authorization'] = `Bearer ${data.access_token}`
    } catch (error) {
      throw error
    }
  }

  const logout = () => {
    try {
      http.post('/admin/logout')
      clearAuth()
    } catch (error) {
      throw error
    }
  }

  const refreshAccessToken = async () => {
    if (!refreshToken.value && isRefreshTokenValid) {
      throw new Error('No refresh token available')
    }

    try {
      const response = await http.post('/admin/renew_access', {
        refresh_token: refreshToken.value
      }, {skipAuth: true})

      const { access_token, access_token_expires_at } = response.data
      updateAccessToken(access_token, access_token_expires_at)

      return access_token

    } catch (error) {
      // 刷新失败，清除所有认证信息
      logout()
      throw error
    }
  }

  const updateAccessToken = (newAccessToken: string, newExpiresAt: string) => {
    accessToken.value = newAccessToken
    accessTokenExpiresAt.value = newExpiresAt

    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, newAccessToken)
    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN_EXPIRES_AT, newExpiresAt)
  }

  // 初始化认证状态
  initAuth()

  return {
    admin: readonly(admin),
    accessToken: readonly(accessToken),
    accessTokenExpiresAt: readonly(accessTokenExpiresAt),
    isAuthenticated: readonly(isAuthenticated),
    login,
    logout,
    clearAuth,
    refreshAccessToken,
    updateAccessToken,
  }
})
