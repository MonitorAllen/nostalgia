import { computed, readonly, ref } from 'vue'
import { defineStore } from 'pinia'
import { loginAdmin, renewAdminAccessToken } from '../api/adminAuthApi'
import type { AdminLoginRequest, AdminTokens, AdminUser } from '../types'

const STORAGE_KEYS = {
  TOKEN: 'nostalgia_admin_access_token',
  TOKEN_EXPIRES: 'nostalgia_admin_access_token_expires_at',
  REFRESH_TOKEN: 'nostalgia_admin_refresh_token',
  REFRESH_TOKEN_EXPIRES: 'nostalgia_admin_refresh_token_expires_at',
  ADMIN: 'nostalgia_admin_user',
} as const

export function readJson<T>(key: string, fallback: T): T {
  const raw = localStorage.getItem(key)
  if (!raw) return fallback

  try {
    return JSON.parse(raw) as T
  } catch {
    localStorage.removeItem(key)
    return fallback
  }
}

function isFutureDate(value: string) {
  const timestamp = new Date(value).getTime()
  return Number.isFinite(timestamp) && timestamp > Date.now()
}

export const useAdminAuthStore = defineStore('admin-auth', () => {
  const token = ref(localStorage.getItem(STORAGE_KEYS.TOKEN) || '')
  const tokenExpiresAt = ref(localStorage.getItem(STORAGE_KEYS.TOKEN_EXPIRES) || '')
  const refreshToken = ref(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN) || '')
  const refreshTokenExpiresAt = ref(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN_EXPIRES) || '')
  const admin = ref<AdminUser | null>(readJson<AdminUser | null>(STORAGE_KEYS.ADMIN, null))

  const hasValidAccessToken = () => {
    return Boolean(token.value && tokenExpiresAt.value && isFutureDate(tokenExpiresAt.value))
  }

  const hasValidRefreshToken = () => {
    return Boolean(
      refreshToken.value &&
        refreshTokenExpiresAt.value &&
        isFutureDate(refreshTokenExpiresAt.value),
    )
  }

  const isAuthenticated = computed(hasValidAccessToken)

  const isRefreshTokenValid = computed(hasValidRefreshToken)

  const setTokens = (tokens: AdminTokens) => {
    token.value = tokens.access_token
    tokenExpiresAt.value = tokens.access_token_expires_at
    refreshToken.value = tokens.refresh_token
    refreshTokenExpiresAt.value = tokens.refresh_token_expires_at

    localStorage.setItem(STORAGE_KEYS.TOKEN, tokens.access_token)
    localStorage.setItem(STORAGE_KEYS.TOKEN_EXPIRES, tokens.access_token_expires_at)
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, tokens.refresh_token)
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN_EXPIRES, tokens.refresh_token_expires_at)
  }

  const updateAccessToken = (accessToken: string, expiresAt: string) => {
    token.value = accessToken
    tokenExpiresAt.value = expiresAt

    localStorage.setItem(STORAGE_KEYS.TOKEN, accessToken)
    localStorage.setItem(STORAGE_KEYS.TOKEN_EXPIRES, expiresAt)
  }

  const setAdmin = (value: AdminUser | null) => {
    admin.value = value

    if (value) {
      localStorage.setItem(STORAGE_KEYS.ADMIN, JSON.stringify(value))
      return
    }

    localStorage.removeItem(STORAGE_KEYS.ADMIN)
  }

  const clear = () => {
    token.value = ''
    tokenExpiresAt.value = ''
    refreshToken.value = ''
    refreshTokenExpiresAt.value = ''
    admin.value = null

    Object.values(STORAGE_KEYS).forEach((key) => {
      localStorage.removeItem(key)
    })
  }

  const login = async (credentials: AdminLoginRequest) => {
    try {
      const response = await loginAdmin(credentials)
      const { admin: adminUser, ...tokens } = response.data

      setTokens(tokens)
      setAdmin(adminUser)

      return response
    } catch (error) {
      clear()
      throw error
    }
  }

  const refreshAccessToken = async () => {
    if (!hasValidRefreshToken()) {
      clear()
      throw new Error('No valid admin refresh token available')
    }

    try {
      const response = await renewAdminAccessToken(refreshToken.value)
      const { access_token, access_token_expires_at } = response.data

      updateAccessToken(access_token, access_token_expires_at)

      return access_token
    } catch (error) {
      clear()
      throw error
    }
  }

  const ensureAuthenticated = async () => {
    if (hasValidAccessToken()) {
      return true
    }

    if (!hasValidRefreshToken()) {
      clear()
      return false
    }

    try {
      await refreshAccessToken()
      return true
    } catch {
      clear()
      return false
    }
  }

  const logout = () => {
    clear()
    window.location.href = '/admin/login'
  }

  return {
    token: readonly(token),
    tokenExpiresAt: readonly(tokenExpiresAt),
    refreshToken: readonly(refreshToken),
    refreshTokenExpiresAt: readonly(refreshTokenExpiresAt),
    admin: readonly(admin),
    isAuthenticated,
    isRefreshTokenValid,
    hasValidAccessToken,
    hasValidRefreshToken,
    setTokens,
    setAdmin,
    clear,
    login,
    refreshAccessToken,
    ensureAuthenticated,
    logout,
  }
})
