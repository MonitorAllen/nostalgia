import http from "@/util/http";
import type {ApiSuccessResponse} from "@/types/request/api";
import type {
    ListArticlesRequest,
    ListArticlesResponse,
    SearchArticlesRequest,
    SearchArticlesResponse
} from "@/types/request/article";
import type {Article} from "@/types/article";


export async function listArticle(req: ListArticlesRequest): Promise<ApiSuccessResponse<ListArticlesResponse>> {
    return http.get(`/articles`,
        {
            params: req,
            skipAuth: true
        }
    )
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

export async function searchArticles(req: SearchArticlesRequest): Promise<ApiSuccessResponse<SearchArticlesResponse>> {
    return http.get(`/articles/search`,
        {
            params: req,
            skipAuth: true
        }
    )
}