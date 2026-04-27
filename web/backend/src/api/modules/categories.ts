import { Nostalgia } from '@/api/interface/nostalgia'
import http from '@/api'

/**
 * @name 分类管理模块
 */
export const getCategoryListApi = async () => {
  const res = await http.get<Nostalgia.ListAllCategoryResponse>(`/categories/all`)

  // Translate Nostalgia's response to Geeker's ResPage format
  return {
    ...res,
    data: {
      list: res.data.categories,
      total: parseInt(res.data.count),
    },
  } as any
}

export const createCategoryApi = (params: { name: string }) => {
  return http.post(`/categories`, params)
}

export const updateCategoryApi = (params: { id: number; name: string }) => {
  return http.patch(`/categories`, params)
}

export const deleteCategoryApi = (id: number) => {
  return http.delete(`/categories/${id}`)
}
