import axiosInstance from '@/config/axios'
import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Article {
    id: string
    title: string
    summary: string
    content?: string
    likes: number
    views: number
    is_publish: boolean
    created_at: string
    updated_at: string
    owner: string
    category_id: string
    category_name: string
}

export const useArticleStore = defineStore('article', () => {
    const articles = ref<Article[]>([])
    const count = ref(0)

    const listAllArticles = async (page: number, limit: number) => {
        try {
            const res = await axiosInstance.get('/articles', {
                params: {
                    page, limit
                    }
                })
                articles.value = res.data.articles
                count.value = Number.parseInt(res.data.count)
        } catch (error) {
            throw error
        }
    }

    return {
        articles,
        count,
        listAllArticles
    }
})
