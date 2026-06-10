import adminHttp from './adminHttp'
import type { AdminLoginRequest, AdminLoginResponse, AdminTokens, AdminUser } from '../types'

export function loginAdmin(data: AdminLoginRequest) {
  return adminHttp.post<AdminLoginResponse>('/admin/login', data, {
    skipAuth: true,
    skipErrorHandler: true,
  })
}

export function renewAdminAccessToken(refreshToken: string) {
  return adminHttp.post<Pick<AdminTokens, 'access_token' | 'access_token_expires_at'>>(
    '/admin/renew_access',
    { refresh_token: refreshToken },
    { skipAuth: true, skipErrorHandler: true },
  )
}

export function getAdminInfo() {
  return adminHttp.get<{ admin: AdminUser }>('/admin/info')
}
