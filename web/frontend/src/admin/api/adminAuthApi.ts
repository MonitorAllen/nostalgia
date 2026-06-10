import http from '@/util/http'
import { useAuthStore } from '@/store/module/auth'
import type { User } from '@/types/user'
import type { AdminLoginRequest, AdminTokens } from '../types'

export function loginAdmin(data: AdminLoginRequest) {
  return http.post<AdminTokens & { user: User }>('/users/login', data, {
    skipAuth: true,
    skipErrorHandler: true,
  })
}

export function renewAdminAccessToken(refreshToken: string) {
  return http.post<Pick<AdminTokens, 'access_token' | 'access_token_expires_at'>>(
    '/tokens/renew_access',
    { refresh_token: refreshToken },
    { skipAuth: true, skipErrorHandler: true },
  )
}

export function getAdminInfo() {
  const authStore = useAuthStore()
  return Promise.resolve({ data: { admin: authStore.currentUser } })
}
