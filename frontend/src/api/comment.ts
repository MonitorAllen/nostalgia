import type {ApiSuccessResponse} from "@/types/request/api";
import http from "@/util/http";
import type {Comment} from "@/types/comment";

export interface listCommentsRequest {
    articleId: string
}

export interface listCommentsResponse {
    comments: Comment[]
}

export async function listComments(req: listCommentsRequest): Promise<ApiSuccessResponse<listCommentsResponse>> {
    return http.get(`/comments/${req.articleId}`, {skipAuth: true})
}