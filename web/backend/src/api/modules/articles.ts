import { Nostalgia } from '@/api/interface/nostalgia'
import http from '@/api'

/**
 * @name 文章管理模块
 */
export const getArticleListApi = async (params: any) => {
  const res = await http.get<Nostalgia.ListAllArticleResponse>(`/articles`, params)

  return {
    ...res,
    data: {
      list: res.data.articles,
      total: parseInt(res.data.count as any) || 0,
      page: params.page,
      limit: params.limit,
    },
  } as any
}

export const getArticleByIdApi = (id: string, needContent: boolean = true) => {
  return http.get<{ article: Nostalgia.Article }>(`/articles/${id}/${needContent}`)
}

export const createArticleApi = (params: Partial<Nostalgia.Article>) => {
  return http.post<{ article: Nostalgia.Article }>(`/articles`, params)
}

export const updateArticleApi = (params: Partial<Nostalgia.Article>) => {
  return http.patch<{ article: Nostalgia.Article }>(`/articles`, params)
}

export const deleteArticleApi = (id: string) => {
  return http.delete(`/articles/${id}`)
}
