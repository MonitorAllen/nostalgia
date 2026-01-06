import type {Category} from "@/types/category";
import http from "@/util/http";
import type {GetCategoryRequest, GetCategoryResponse} from "@/types/request/category";
import type {ApiSuccessResponse} from "@/types/request/api";

export async function getCategory(data: GetCategoryRequest): Promise<ApiSuccessResponse<GetCategoryResponse>> {
    return http.get(`/categories/${data.id}`, {skipAuth: true})
}

export interface ListCategoriesResponse {
    categories: Category[]
    count: number
}

export async function listCategories(): Promise<ApiSuccessResponse<ListCategoriesResponse>> {
    return http.get('/categories', {skipAuth: true})
}