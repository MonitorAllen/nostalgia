import axiosInstance from '@/config/axios'
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type {Article} from "@/types/article";

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
