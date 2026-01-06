import type { ApiSuccessResponse } from "@/types/api.ts";
import type { Category } from "@/types/category.ts";
import http from "@/util/http.ts";
import type { AxiosResponse } from "axios";

export interface CreateCategoryRequest {
  name: string
}

export interface CreateCategoryResponse {
  category: Category
}

export async function createCategory(req: CreateCategoryRequest): Promise<AxiosResponse<CreateCategoryResponse>> {
  return http.post('/category', req)
}

export interface ListAllCategoriesResponse {
  categories: Category[]
}

export async function listAllCategories(): Promise<ApiSuccessResponse<ListAllCategoriesResponse>> {
  return http.get('/category/all')
}

export interface ListCategoriesResponse {
  categories: Category[]
  count: string
}

export async function listCategories(): Promise<AxiosResponse<ListCategoriesResponse>> {
  return http.get(`/category`)
}

export interface UpdateCategoryRequest {
  id: number
  name: string
}

export interface UpdateCategoryResponse {
  category: Category
}

export async function updateCategory(req: UpdateCategoryRequest): Promise<AxiosResponse<UpdateCategoryResponse>> {
  return http.put(`/category`, req)
}

export interface DeleteCategoryRequest {
  id: number
}

export async function deleteCategory(req: DeleteCategoryRequest): Promise<AxiosResponse> {
  return http.delete(`/category/${req.id}`)
}
