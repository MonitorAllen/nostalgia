import adminHttp from './adminHttp'
import type { AdminCategory, AdminCategoryListResponse } from '../types'

export function listAllAdminCategories() {
  return adminHttp.get<AdminCategoryListResponse>('/categories/all')
}

export function createAdminCategory(data: { name: string }) {
  return adminHttp.post<{ category: AdminCategory }>('/categories', data)
}

export function updateAdminCategory(data: { id: number; name: string }) {
  return adminHttp.patch<{ category: AdminCategory }>('/categories', data)
}

export function deleteAdminCategory(id: number) {
  return adminHttp.delete<void>(`/categories/${id}`)
}
