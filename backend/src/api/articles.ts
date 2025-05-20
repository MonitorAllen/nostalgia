import axiosInstance from "@/config/axios";
import type { Article } from "@/stores/article";

export const fetchArticleById = async (id: string): Promise<Article> => {
    const res = await axiosInstance.get(`/articles/${id}`)
    return res.data.article as Article
}

export type UpdateArticleParams = {
    id: string
    title: string
    summary: string
    is_publish: boolean
}

export const updateArticle = async (params: UpdateArticleParams): Promise<Article> => {
    const res = await axiosInstance.patch('/articles', params)
    return res.data.article as Article
}