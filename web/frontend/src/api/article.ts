import http from "@/util/http";
import type {ApiSuccessResponse} from "@/types/request/api";
import type {ListArticlesRequest, ListArticlesResponse} from "@/types/request/article";
import type {Article} from "@/types/article";


export async function listArticle(req: ListArticlesRequest): Promise<ApiSuccessResponse<ListArticlesResponse>> {
    return http.get(`/articles?category_id=${req.categoryId}&page=${req.page}&limit=${req.limit}`, {skipAuth: true})
}

export interface GetArticleRequest {
    id: string
}

export interface GetArticleResponse {
    article: Article
}

export async function getArticle(req: GetArticleRequest): Promise<ApiSuccessResponse<GetArticleResponse>> {
    return http.get(`/articles/${req.id}`, {skipAuth: true})
}

export interface IncrementArticleLikesRequest {
    id: string
}

export async function incrementArticleLikes(req: IncrementArticleLikesRequest): Promise<ApiSuccessResponse<any>> {
    return http.patch('/articles/increment_likes', req, {skipAuth: true})
}

export interface IncrementArticleViewsRequest {
    id: string
}

export async function incrementArticleViews(req: IncrementArticleViewsRequest): Promise<ApiSuccessResponse<any>> {
    return http.patch('/articles/increment_views', req, {skipAuth: true})
}