import type { Article } from "@/stores/article";
import type {ApiSuccessResponse} from "@/types/api.ts";
import http from "@/util/http.ts";
import type {AxiosResponse} from "axios";

export interface FetchArticleByIdRequest {
  id: string
  needContent: boolean
}

export interface FetchArticleByIdResponse {
  article: Article
}

export async function fetchArticleById (req: FetchArticleByIdRequest): Promise<ApiSuccessResponse<FetchArticleByIdResponse>> {
    return http.get(`/articles/${req.id}/${req.needContent}`, {skipAuth: false})
}

export interface UpdateArticleRequest {
    id: string
    title: string
    summary: string
    content?: string
    is_publish: boolean
    category_id: number
}

export interface UpdateArticleResponse {
  article: Article
}

export async function updateArticle   (req: UpdateArticleRequest): Promise<ApiSuccessResponse<UpdateArticleResponse>> {
    return  http.patch('/articles', req)
}

export interface CreateArticleRequest {
    title: string
    summary: string
    is_publish: boolean
}

export interface CreateArticleResponse {
  article: Article
}

export async function createArticle(req: CreateArticleRequest): Promise<AxiosResponse<CreateArticleResponse>> {
    return http.post('/articles', req)
}

export interface DeleteArticleRequest {
  id: string
}

export async function deleteArticle (req: DeleteArticleRequest): Promise<AxiosResponse> {
    return http.delete(`/articles/${req.id}`)
}

export interface ListAllArticleRequest {
  page: number
  limit: number
}

export interface ListAllArticleResponse {
  articles: Article[]
  count: string
}

export async function listAllArticles(req: ListAllArticleRequest): Promise<AxiosResponse<ListAllArticleResponse>> {
  return http.get(`/articles?page=${req.page}&limit=${req.limit}`)
}
