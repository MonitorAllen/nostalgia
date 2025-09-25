import axiosInstance from "@/config/axios";
import type { Admin } from "@/stores/auth";

export interface UpdateAdminParams {
    id: number
    username?: string
    password?: string
    is_active?: boolean
    role_id?: number
    old_password?: string
}

export const updateAdmin = async (params: UpdateAdminParams): Promise<Admin> => {
    const res = await axiosInstance.patch('/admin', params)
    return res.data.admin as Admin
}