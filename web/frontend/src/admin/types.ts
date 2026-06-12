export type AdminInt64 = string | number

export interface AdminUser {
  id: AdminInt64
  username: string
  role?: string
  created_at?: string
}

export interface AdminTokens {
  access_token: string
  access_token_expires_at: string
  refresh_token: string
  refresh_token_expires_at: string
}

export interface AdminLoginRequest {
  username: string
  password: string
}

export interface AdminLoginResponse extends AdminTokens {
  admin: AdminUser
}

export interface AdminArticle {
  id: string
  title: string
  summary: string
  content: string
  likes: number
  views: number
  is_publish: boolean
  created_at: string
  updated_at: string
  owner: string
  category_id?: AdminInt64
  category_name?: string
  cover?: string
  slug?: string
  check_outdated?: boolean
  created_by_automation?: boolean
  automation_status?: string
}

export interface CreateAdminArticleRequest {
  title?: string
  summary?: string
  content?: string
  is_publish?: boolean
  category_id?: AdminInt64
}

export interface UpdateAdminArticleRequest extends CreateAdminArticleRequest {
  id: string
  cover?: string
  slug?: string
  check_outdated?: boolean
}

export interface AdminArticleListResponse {
  articles: AdminArticle[]
  count: string | number
}

export interface AdminArticleResponse {
  article: AdminArticle
}

export interface AdminCategory {
  id: AdminInt64
  name: string
  article_count?: AdminInt64
  created_at?: string
  updated_at?: string
}

export interface AdminCategoryAllResponse {
  categories: AdminCategory[]
}

export interface AdminUploadRequest {
  article_id: string
  content: string
  type: 'content' | 'cover'
}

export interface AdminUploadResponse {
  url: string
  filename: string
}

export type AdminAIPolishMode =
  | 'improve'
  | 'shorten'
  | 'expand'
  | 'title_candidates'
  | 'summary_candidates'

export type AdminAIPolishTarget = 'content_selection' | 'title' | 'summary'

export interface AdminAIPolishRequest {
  mode: AdminAIPolishMode
  target: AdminAIPolishTarget
  text: string
  article_id?: string
  article_title?: string
  article_summary?: string
  article_excerpt?: string
  locale?: string
}

export interface AdminAIPolishSuggestion {
  content: string
  reason?: string
}

export interface AdminAIPolishResponse {
  suggestions: AdminAIPolishSuggestion[]
  mode: AdminAIPolishMode
  target: AdminAIPolishTarget
  model: string
}
