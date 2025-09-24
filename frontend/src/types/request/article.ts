import type {Article} from "@/types/article";

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

export interface ListArticlesRequest {
  categoryId: number
  page: number
  limit: number
}

// 文章列表响应
export interface ListArticlesResponse {
  articles: Article[]
  count: number
}

export interface ArticleResponse {
  article: Article
}
