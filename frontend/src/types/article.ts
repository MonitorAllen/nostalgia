import type { Tag } from '@/types/tag'

export interface Article {
  id: string
  title: string
  summary: string
  content: string
  views: number
  likes: number
  is_publish: boolean
  owner: string
  create_at: string
  update_at: string
  delete_at: string
  username: string
  tags: Tag[]
}