import { defineStore } from 'pinia'
import { UserState } from '@/stores/interface'
import piniaPersistConfig from '@/stores/helper/persist'

export const useUserStore = defineStore({
  id: 'geeker-user',
  state: (): UserState => ({
    token: '',
    refreshToken: '',
    expiresAt: '',
    userInfo: { name: 'Geeker' },
  }),
  getters: {},
  actions: {
    // Set Token
    setToken(token: string) {
      this.token = token
    },
    // Set Refresh Token
    setRefreshToken(token: string) {
      this.refreshToken = token
    },
    // Set ExpiresAt
    setExpiresAt(expiresAt: string) {
      this.expiresAt = expiresAt
    },
    // Set setUserInfo
    setUserInfo(userInfo: UserState['userInfo']) {
      this.userInfo = userInfo
    },
  },
  persist: piniaPersistConfig('geeker-user'),
})
