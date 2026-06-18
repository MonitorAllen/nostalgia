import adminHttp from './adminHttp'
import type {
  AdminUserListResponse,
  AdminUserStatusFilter,
  DisableAdminUserRequest,
  ManagedAdminUser,
  UpdateAdminUserRequest
} from '../types'

export interface ListAdminUsersParams {
  q?: string
  status: AdminUserStatusFilter
  page: number
  limit: number
}

export function listAdminUsers(params: ListAdminUsersParams) {
  return adminHttp.get<AdminUserListResponse>('/users', { params })
}

export function updateAdminUser(data: UpdateAdminUserRequest) {
  return adminHttp.patch<{ user: ManagedAdminUser }>(`/users/${data.id}`, data)
}

export function disableAdminUser(id: string, data: DisableAdminUserRequest) {
  return adminHttp.post<{ user: ManagedAdminUser }>(`/users/${id}/disable`, data)
}

export function enableAdminUser(id: string) {
  return adminHttp.post<{ user: ManagedAdminUser }>(`/users/${id}/enable`)
}
