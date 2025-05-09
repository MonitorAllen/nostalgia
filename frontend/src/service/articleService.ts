import request from '@/util/request.js'
import type {CreateArticleRequest, UpdateArticleRequest} from '@/types/request/article'

// 根据ID获取文章
const getArticle = (id: string) => request.get(`articles/${id}`, {
    // 设置不使用拦截器
    headers: {},
    skipAuth: true, // 自定义配置，用于标记该请求不需要经过拦截器
})

const listArticle = (page: number, limit: number) => {
    return request.get(`/articles?page=${page}&limit=${limit}`, {
        // 设置不使用拦截器
        headers: {},
        skipAuth: true, // 自定义配置，用于标记该请求不需要经过拦截器
    })
}

const createArticle = (req: CreateArticleRequest) =>
    request.post('/articles', req)

const updateArticle = (req: UpdateArticleRequest) =>
    request.put('/articles', req)

export default {
    getArticle,
    listArticle,
    createArticle,
    updateArticle,
}
