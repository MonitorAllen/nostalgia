import http from '@/util/http'
import type { User } from '@/types/user'

export interface SetupStatus {
  initialized: boolean
  setup_available: boolean
}

export interface CreateSetupAdminPayload {
  setup_token: string
  username: string
  password: string
  full_name: string
  email: string
}

export function getSetupStatus() {
  return http.get<SetupStatus>('/setup/status', {
    skipAuth: true,
    skipErrorHandler: true,
  })
}

export function createSetupAdmin(payload: CreateSetupAdminPayload) {
  return http.post<User>('/setup/admin', payload, {
    skipAuth: true,
    skipErrorHandler: true,
  })
}
