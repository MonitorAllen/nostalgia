import { Login } from '@/api/interface/index'
import authMenuList from '@/assets/json/authMenuList.json'
import authButtonList from '@/assets/json/authButtonList.json'
import http from '@/api'

/**
 * @name 登录模块
 */
// 用户登录
export const loginApi = (params: Login.ReqLoginForm) => {
  return http.post<any>(`/admin/login`, params, { loading: false })
}

// 续期 token
export const renewAccessTokenApi = (params: { refresh_token: string }) => {
  return http.post<any>(`/admin/renew_access`, params, { loading: false })
}

// 获取菜单列表
export const getAuthMenuListApi = () => {
  return authMenuList
}

// 获取按钮权限
export const getAuthButtonListApi = () => {
  return authButtonList
}

// 用户退出登录
export const logoutApi = () => {
  return Promise.resolve() // Nostalgia might not have a logout endpoint or just client-side clear
}
