import { defineStore } from 'pinia'
import {ref, computed, readonly} from 'vue'
import http from '@/util/http'
import type { LoginRequest, RegisterRequest } from '@/types/request/user'
import type { User } from '@/types/user'
import {useUserStore} from "@/store/module/user";
import axios from "axios";

// 存储键名常量
const STORAGE_KEYS = {
    TOKEN: 'nostalgia_access_token',
    TOKEN_EXPIRES: 'nostalgia_token_expires_at',
    REFRESH_TOKEN: 'nostalgia_refresh_token',
    REFRESH_TOKEN_EXPIRES: 'nostalgia_refresh_token_expires_at',
} as const

export const useAuthStore = defineStore('auth', () => {
    // 状态
    const token = ref(localStorage.getItem(STORAGE_KEYS.TOKEN) || '')
    const tokenExpiresAt = ref(localStorage.getItem(STORAGE_KEYS.TOKEN_EXPIRES) || '')
    const refreshToken = ref(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN) || '')
    const refreshTokenExpiresAt = ref(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN_EXPIRES) || '')

    // 计算属性
    const isAuthenticated = computed(() => {
        if (!token.value || !tokenExpiresAt.value) return false

        const expiresAt = new Date(tokenExpiresAt.value)
        const now = new Date()

        return now < expiresAt
    })

    const isTokenExpiring = computed(() => {
        if (!tokenExpiresAt.value) return false

        const expiresAt = new Date(tokenExpiresAt.value)
        const now = new Date()
        const fiveMinutesFromNow = new Date(now.getTime() + 5 * 60 * 1000)

        return expiresAt < fiveMinutesFromNow
    })

    const isRefreshTokenValid = computed(() => {
        if (!refreshTokenExpiresAt.value) return false
        const expiresAt = new Date(refreshTokenExpiresAt.value)
        return new Date() < expiresAt
    })

    // 方法
    const setTokens = (tokens: {
        access_token: string
        access_token_expires_at: string
        refresh_token: string
        refresh_token_expires_at: string
    }) => {
        token.value = tokens.access_token
        tokenExpiresAt.value = tokens.access_token_expires_at
        refreshToken.value = tokens.refresh_token
        refreshTokenExpiresAt.value = tokens.refresh_token_expires_at

        // 同步到 localStorage
        localStorage.setItem(STORAGE_KEYS.TOKEN, tokens.access_token)
        localStorage.setItem(STORAGE_KEYS.TOKEN_EXPIRES, tokens.access_token_expires_at)
        localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, tokens.refresh_token)
        localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN_EXPIRES, tokens.refresh_token_expires_at)
    }

    const updateAccessToken = (accessToken: string, expiresAt: string) => {
        token.value = accessToken
        tokenExpiresAt.value = expiresAt

        localStorage.setItem(STORAGE_KEYS.TOKEN, accessToken)
        localStorage.setItem(STORAGE_KEYS.TOKEN_EXPIRES, expiresAt)
    }

    const clearTokens = () => {
        token.value = ''
        tokenExpiresAt.value = ''
        refreshToken.value = ''
        refreshTokenExpiresAt.value = ''

        // 清除 localStorage
        Object.values(STORAGE_KEYS).forEach(key => {
            localStorage.removeItem(key)
        })
    }

    const login = async (credentials: LoginRequest) => {
        try {
            const response = await http.post('/auth/login', credentials, { skipAuth: true })

            const { access_token, access_token_expires_at, refresh_token, refresh_token_expires_at, user } = response.data

            setTokens({
                access_token,
                access_token_expires_at,
                refresh_token,
                refresh_token_expires_at
            })

            // 触发用户信息更新
            const userStore = useUserStore()
            userStore.setUser(user)

            return response

        } catch (error) {
            clearTokens()
            throw error
        }
    }

    const register = async (data: RegisterRequest) => {
        try {
            // 先注册
            await http.post('/auth/register', data, { skipAuth: true })

            // 然后自动登录
            return await login({
                username: data.username,
                password: data.password
            })

        } catch (error) {
            clearTokens()
            throw error
        }
    }

    const logout = () => {
        clearTokens()

        // 清除用户信息
        const userStore = useUserStore()
        userStore.clearUser()

        // 重新加载页面或跳转到登录页
        window.location.href = '/login'
    }

    const refreshAccessToken = async () => {
        if (!refreshToken.value && isRefreshTokenValid.value) {
            throw new Error('No refresh token available')
        }

        try {
            const response = await axios.post('/tokens/renew_access', {
                refresh_token: refreshToken.value
            }, { skipAuth: true, skipErrorHandler: true })

            const { access_token, access_token_expires_at } = response.data
            updateAccessToken(access_token, access_token_expires_at)

            return access_token
        } catch (error) {
            // 刷新失败，清除所有认证信息
            logout()
            throw error
        }
    }

    return {
        // 状态
        token: readonly(token),
        tokenExpiresAt: readonly(tokenExpiresAt),
        refreshToken: readonly(refreshToken),
        refreshTokenExpiresAt: readonly(refreshTokenExpiresAt),

        // 计算属性
        isAuthenticated,
        isTokenExpiring,

        // 方法
        login,
        register,
        logout,
        refreshAccessToken,
        updateAccessToken,
        setTokens,
        clearTokens
    }
})