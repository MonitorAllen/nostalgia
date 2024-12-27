export interface CreateCommentRequest {
    article_id: string
    content: string
    parent_id: number
    from_user_id: string
    to_user_id: string
}

export interface DeleteCommentRequest {
    id: number
}