export interface CreateArticleRequest {
  title: string
  summary: string
  content: string
  is_publish: boolean
}

export interface UpdateArticleRequest {
  id: string
  title: string
  summary: string
  content: string
  is_publish: boolean
  owner: string
}

