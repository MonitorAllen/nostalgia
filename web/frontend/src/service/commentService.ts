import request from '@/util/request'
import type {CreateCommentRequest, DeleteCommentRequest} from "@/types/request/comment";


const listComments = (articleId: string) =>
    request.get('/comments/' + articleId)

const createComment = (req: CreateCommentRequest) =>
    request.post('/comments/', req)

const deleteComment = (id: number) => 
    request.delete(`/comments/${id}`)

export default {
    listComments,
    createComment,
    deleteComment,
}
