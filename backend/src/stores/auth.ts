import { defineStore } from 'pinia'
import { ref } from 'vue'
import axiosInstance from '../config/axios'

interface LoginForm {
  username: string
  password: string
}

interface Admin {
  id: number
  username: string
  email: string
  role: string
  avatar?: string
}

export const useAuthStore = defineStore('auth', () => {
  const admin = ref<Admin | null>(null)
  const access_token = ref<string | null>(null)
  const refresh_token = ref<string | null>(null)
  const isAuthenticated = ref(false)

  // 初始化认证状态
  const initAuth = () => {
    const storedAccessToken = localStorage.getItem('access_token')
    const storedAccessTokenExpiresAt = localStorage.getItem('access_token_expires_at')
    
    if (storedAccessToken && storedAccessTokenExpiresAt) {
      // 检查 token 是否过期
      const expiresAt = new Date(storedAccessTokenExpiresAt).getTime()
      const now = new Date().getTime()
      
      if (now < expiresAt) {
        // token 未过期，恢复认证状态
        access_token.value = storedAccessToken
        refresh_token.value = localStorage.getItem('refresh_token')
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
      const response = await axiosInstance.get('/admin/info')
      admin.value = response.data.admin
    } catch (error) {
      clearAuth()
    }
  }

  // 清除认证信息
  const clearAuth = () => {
    admin.value = null
    access_token.value = null
    refresh_token.value = null
    isAuthenticated.value = false
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('access_token_expires_at')
    localStorage.removeItem('refresh_token_expires_at')
    delete axiosInstance.defaults.headers.common['Authorization']
  }

  const login = async (form: LoginForm) => { 
    try {
      const response = await axiosInstance.post('/admin/login', form)
      const { data } = response
      access_token.value = data.access_token
      refresh_token.value = data.refresh_token
      admin.value = data.admin
      isAuthenticated.value = true
      // 保存token到localStorage
      localStorage.setItem('access_token', data.access_token)
      localStorage.setItem('refresh_token', data.refresh_token)
      localStorage.setItem('access_token_expires_at', data.access_token_expires_at)
      localStorage.setItem('refresh_token_expires_at', data.refresh_token_expires_at)
      // 设置axios默认请求头
      axiosInstance.defaults.headers.common['Authorization'] = `Bearer ${data.access_token}`
    } catch (error) {
      throw error
    }
  }

  const logout = () => {
    try {
      axiosInstance.post('/admin/logout')
      clearAuth()
    } catch (error) {
      throw error
    }
  }

  // 初始化认证状态
  initAuth()

  return {
    admin,
    access_token,
    isAuthenticated,
    login,
    logout,
  }
}) 