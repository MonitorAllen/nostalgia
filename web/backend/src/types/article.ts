export interface Article {
  id: string
  title: string
  summary: string
  content?: string
  likes: number
  views: number
  is_publish: boolean
  created_at: string
  updated_at: string
  owner: string
  category_id: number
  category_name: string
  cover: string
  slug: string
  check_outdated: boolean
}
