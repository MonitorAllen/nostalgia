import { computed } from 'vue'
import { defineStore } from 'pinia'
import { ADMIN_LOGIN_PATH } from '@/admin/adminRoutes'
import { useAuthStore } from '@/store/module/auth'
import type { AdminLoginRequest } from '../types'

export const useAdminAuthStore = defineStore('admin-auth', () => {
  const authStore = useAuthStore()
  const admin = computed(() => (authStore.isAdmin ? authStore.currentUser : null))

  const login = async (credentials: AdminLoginRequest) => {
    const response = await authStore.login(credentials)

    if (response.data.user.role !== 'admin') {
      authStore.clearTokens()
      throw new Error('Admin role required')
    }

    return response
  }

  const clear = () => {
    authStore.clearTokens()
  }

  const logout = () => {
    clear()
    window.location.href = ADMIN_LOGIN_PATH
  }

  return {
    token: computed(() => authStore.token),
    tokenExpiresAt: computed(() => authStore.tokenExpiresAt),
    refreshToken: computed(() => authStore.refreshToken),
    refreshTokenExpiresAt: computed(() => authStore.refreshTokenExpiresAt),
    admin,
    isAuthenticated: computed(() => authStore.isAdmin && authStore.isAuthenticated),
    isRefreshTokenValid: computed(() => authStore.isRefreshTokenValid),
    hasValidAccessToken: authStore.hasValidAccessToken,
    hasValidRefreshToken: authStore.hasValidRefreshToken,
    setTokens: authStore.setTokens,
    setAdmin: authStore.setCurrentUser,
    clear,
    refreshAccessToken: authStore.refreshAccessToken,
    ensureAuthenticated: authStore.ensureAdminAuthenticated,
    login,
    logout
  }
})
