export interface Article {
  id: string
  title: string
  summary: string
  content: string
  views: number
  likes: number
  is_publish: boolean
  owner: string
  created_at: string
  updated_at: string
  deleted_at: string
  username: string
  category_name: string
}

export interface ArticleComments {
  id: number
  content: string
  article_id: string
  parent_id: number
  likes: number
  from_user_id: string
  to_user_id: string
  created_at: string
  deleted_at: string
  from_user_name: string
  to_user_name: string
  child: ArticleComments[]
}