import axiosInstance from "@/config/axios";
import type { Article } from "@/stores/article";

export const fetchArticleById = async (id: string, neetContent: boolean): Promise<Article> => {
    const res = await axiosInstance.get(`/articles/${id}/${neetContent}`)
    return res.data.article as Article
}

export interface UpdateArticleParams {
    id: string
    title: string
    summary: string
    content: string
    is_publish: boolean
}

export const updateArticle = async (params: UpdateArticleParams): Promise<Article> => {
    const res = await axiosInstance.patch('/articles', params)
    return res.data.article as Article
}

export interface CreateArticleParams {
    title: string
    summary: string
    is_publish: boolean
}

export const createArticle = async (params: CreateArticleParams): Promise<Article> => {
    const res = await axiosInstance.post('/articles', params)
    return res.data.article as Article
}

export const deleteArticle = async (id: string): Promise<any> => {
    await axiosInstance.delete(`/articles/${id}`)
}