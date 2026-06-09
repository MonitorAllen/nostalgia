import adminHttp from './adminHttp'
import type { AdminCategory, AdminCategoryAllResponse, AdminInt64 } from '../types'

export function listAllAdminCategories() {
  return adminHttp.get<AdminCategoryAllResponse>('/categories/all')
}

export function createAdminCategory(data: { name: string }) {
  return adminHttp.post<{ category: AdminCategory }>('/categories', data)
}

export function updateAdminCategory(data: { id: AdminInt64; name: string }) {
  return adminHttp.patch<{ category: AdminCategory }>('/categories', data)
}

export function deleteAdminCategory(id: AdminInt64) {
  return adminHttp.delete<void>(`/categories/${id}`)
}
