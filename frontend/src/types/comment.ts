export interface Comment {
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
    child: Comment[]
}