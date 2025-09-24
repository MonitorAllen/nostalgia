import type { AxiosResponse } from "axios";

export type ApiSuccessResponse<T> = AxiosResponse<T>

export interface ApiErrorResponse {
    error: string
}