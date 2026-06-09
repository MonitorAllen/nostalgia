import adminHttp from './adminHttp'
import type { AdminArticle, AdminArticleListResponse, AdminArticleResponse } from '../types'

export interface ListAdminArticlesParams {
  title?: string
  page: number
  limit: number
}

export function listAdminArticles(params: ListAdminArticlesParams) {
  return adminHttp.get<AdminArticleListResponse>('/articles', { params })
}

export function getAdminArticle(id: string, needContent = true) {
  return adminHttp.get<AdminArticleResponse>(`/articles/${id}/${needContent}`)
}

export function createAdminArticle(data: Partial<AdminArticle>) {
  return adminHttp.post<AdminArticleResponse>('/articles', data)
}

export function updateAdminArticle(data: Partial<AdminArticle>) {
  return adminHttp.patch<AdminArticleResponse>('/articles', data)
}

export function deleteAdminArticle(id: string) {
  return adminHttp.delete<void>(`/articles/${id}`)
}
