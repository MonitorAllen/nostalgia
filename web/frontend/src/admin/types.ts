export interface AdminUser {
  id: number
  username: string
  is_active?: boolean
  role_id?: number
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
  category_id?: number
  category_name?: string
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
  id: number
  name: string
  article_count?: number
  created_at?: string
  updated_at?: string
}

export interface AdminCategoryListResponse {
  categories: AdminCategory[]
  count: string | number
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
