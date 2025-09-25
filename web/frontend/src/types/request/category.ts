import type {Category} from "@/types/category";

export interface GetCategoryRequest {
    id: number
}

export interface GetCategoryResponse {
    category: Category
}