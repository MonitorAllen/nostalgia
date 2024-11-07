import storageService from '@/service/storageService'
import userService from '@/service/userService'
import { defineStore } from 'pinia'
import type { User } from '@/types/user'
import type { LoginRequest, RegisterRequest } from '@/types/request/user'

export const useUserStore = defineStore('user', {
  state: () => ({
    token: storageService.get(storageService.USER_TOKEN),
    token_expires_at: storageService.get(storageService.USER_TOKEN_EXPIRES_AT),
    refresh_token: storageService.get(storageService.USER_REFRESH_TOKEN),
    refresh_token_expires_at: storageService.get(storageService.USER_REFRESH_TOKEN_EXPIRES_AT),
    userInfo: storageService.get(storageService.USER_INFO)
      ? JSON.parse(storageService.get(storageService.USER_INFO))
      : null,
  }),
  actions: {
    SET_TOKEN(token: string) {
      // 更新本地缓存
      storageService.set(storageService.USER_TOKEN, token)
      // 更新 state
      this.token = token
    },
    SET_TOKEN_EXPIRES(token_expires_at: string) {
      // 更新本地缓存
      storageService.set(storageService.USER_TOKEN_EXPIRES_AT, token_expires_at)
      // 更新 state
      this.token_expires_at = token_expires_at
    },
    SET_REFRESH_TOKEN(refresh_token: string) {
      // 更新本地缓存
      storageService.set(storageService.USER_REFRESH_TOKEN, refresh_token)
      // 更新 state
      this.refresh_token = refresh_token
    },
    SET_REFRESH_TOKEN_EXPIRES(refresh_token_expires_at: string) {
      // 更新本地缓存
      storageService.set(storageService.USER_REFRESH_TOKEN_EXPIRES_AT, refresh_token_expires_at)
      // 更新 state
      this.refresh_token_expires_at = refresh_token_expires_at
    },
    SET_USERINFO(userInfo: User | '') {
      // 更新本地缓存
      storageService.set(storageService.USER_INFO, JSON.stringify(userInfo))
      // 更新 state
      this.userInfo = userInfo
    },
    register({ username, email, full_name, password }: RegisterRequest) {
      return new Promise((resolve, reject) => {
        userService
          .register({
            username,
            email,
            full_name,
            password,
          })
          .then(() => {
            return userService.login({ username, password })
          })
          .then((res) => {
            // 保存用户信息
            const { access_token, access_token_expires_at, refresh_token, refresh_token_expires_at } = res.data
            this.SET_TOKEN(access_token)
            this.SET_TOKEN_EXPIRES(access_token_expires_at)
            this.SET_REFRESH_TOKEN(refresh_token)
            this.SET_REFRESH_TOKEN_EXPIRES(refresh_token_expires_at)

            this.SET_USERINFO(res.data.user)
            resolve(res)
          })
          .catch((err) => {
            reject(err)
          })
      })
    },
    login({ username, password }: LoginRequest) {
      return new Promise((resolve, reject) => {
        userService
          .login({
            username,
            password,
          })
          .then((res) => {
            // 保存token
            const { access_token, access_token_expires_at, refresh_token, refresh_token_expires_at } = res.data
            this.SET_TOKEN(access_token)
            this.SET_TOKEN_EXPIRES(access_token_expires_at)
            this.SET_REFRESH_TOKEN(refresh_token)
            this.SET_REFRESH_TOKEN_EXPIRES(refresh_token_expires_at)
            // 保存用户信息
            this.SET_USERINFO(res.data.user)
            resolve(res)
          })
          .catch((err) => {
            reject(err)
          })
      })
    },
    logout() {
      // 清除 token
      this.SET_TOKEN('')
      this.SET_TOKEN_EXPIRES('')
      this.SET_REFRESH_TOKEN('')
      this.SET_REFRESH_TOKEN_EXPIRES('')
      // 清除用户信息
      this.SET_USERINFO('')

      window.location.reload()
    },
  },
})
