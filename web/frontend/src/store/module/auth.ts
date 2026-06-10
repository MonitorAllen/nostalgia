import { defineStore } from 'pinia'
import { computed, readonly, ref } from 'vue'
import http from '@/util/http'
import type { LoginRequest, RegisterRequest } from '@/types/request/user'
import type { User } from '@/types/user'
import { useUserStore } from '@/store/module/user'
import {
    AUTH_STORAGE_KEYS as STORAGE_KEYS,
    LEGACY_AUTH_STORAGE_KEYS,
    cleanupLegacyAuthStorage,
} from '@/store/module/authStorage'

interface AuthTokens {
    access_token: string
    access_token_expires_at: string
    refresh_token: string
    refresh_token_expires_at: string
}

interface LoginResponse extends AuthTokens {
    user: User
}

function isFutureDate(value: string) {
    const timestamp = new Date(value).getTime()
    return Number.isFinite(timestamp) && timestamp > Date.now()
}

function readStoredUser() {
    const raw = localStorage.getItem(STORAGE_KEYS.USER)
    if (!raw) return null

    try {
        const parsed = JSON.parse(raw) as User | string | null
        return parsed && typeof parsed === 'object' ? parsed : null
    } catch {
        localStorage.removeItem(STORAGE_KEYS.USER)
        return null
    }
}

export const useAuthStore = defineStore('auth', () => {
    cleanupLegacyAuthStorage()

    const token = ref(localStorage.getItem(STORAGE_KEYS.TOKEN) || '')
    const tokenExpiresAt = ref(localStorage.getItem(STORAGE_KEYS.TOKEN_EXPIRES) || '')
    const refreshToken = ref(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN) || '')
    const refreshTokenExpiresAt = ref(localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN_EXPIRES) || '')
    const currentUser = ref<User | null>(readStoredUser())

    const hasValidAccessToken = () => {
        return Boolean(token.value && tokenExpiresAt.value && isFutureDate(tokenExpiresAt.value))
    }

    const hasValidRefreshToken = () => {
        return Boolean(
            refreshToken.value &&
            refreshTokenExpiresAt.value &&
            isFutureDate(refreshTokenExpiresAt.value),
        )
    }

    const isAuthenticated = computed(hasValidAccessToken)

    const isTokenExpiring = computed(() => {
        if (!tokenExpiresAt.value) return false

        const expiresAt = new Date(tokenExpiresAt.value)
        const now = new Date()
        const twoMinutesFromNow = new Date(now.getTime() + 2 * 60 * 1000)

        return expiresAt < twoMinutesFromNow
    })

    const isRefreshTokenValid = computed(hasValidRefreshToken)

    const role = computed(() => currentUser.value?.role)

    const isAdmin = computed(() => role.value === 'admin')

    const syncUserStore = (user: User | null) => {
        const userStore = useUserStore()
        userStore.SET_USERINFO(user || '')
    }

    const setCurrentUser = (user: User | null) => {
        currentUser.value = user

        if (user) {
            localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(user))
        } else {
            localStorage.removeItem(STORAGE_KEYS.USER)
        }

        syncUserStore(user)
    }

    const setTokens = (tokens: AuthTokens) => {
        token.value = tokens.access_token
        tokenExpiresAt.value = tokens.access_token_expires_at
        refreshToken.value = tokens.refresh_token
        refreshTokenExpiresAt.value = tokens.refresh_token_expires_at

        localStorage.setItem(STORAGE_KEYS.TOKEN, tokens.access_token)
        localStorage.setItem(STORAGE_KEYS.TOKEN_EXPIRES, tokens.access_token_expires_at)
        localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, tokens.refresh_token)
        localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN_EXPIRES, tokens.refresh_token_expires_at)

        LEGACY_AUTH_STORAGE_KEYS.forEach((key) => localStorage.removeItem(key))
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
        currentUser.value = null

        Object.values(STORAGE_KEYS).forEach(key => {
            localStorage.removeItem(key)
        })

        LEGACY_AUTH_STORAGE_KEYS.forEach((key) => localStorage.removeItem(key))
        syncUserStore(null)
    }

    const login = async (credentials: LoginRequest) => {
        try {
            const response = await http.post<LoginResponse>('/users/login', credentials, {
                skipAuth: true,
                skipErrorHandler: true,
            })

            const { access_token, access_token_expires_at, refresh_token, refresh_token_expires_at, user } = response.data

            setTokens({
                access_token,
                access_token_expires_at,
                refresh_token,
                refresh_token_expires_at
            })

            setCurrentUser(user)

            return response

        } catch (error) {
            clearTokens()
            throw error
        }
    }

    const register = async (data: RegisterRequest) => {
        try {
            await http.post('/users', data, {
                skipAuth: true,
                skipErrorHandler: true,
            })

            return await login({
                username: data.username,
                password: data.password
            })

        } catch (error) {
            clearTokens()
            throw error
        }
    }

    const logout = (redirectTo: string | false = '/login') => {
        clearTokens()

        if (redirectTo) {
            window.location.href = redirectTo
        }
    }

    const refreshAccessToken = async () => {
        if (!hasValidRefreshToken()) {
            clearTokens()
            throw new Error('No refresh token available')
        }

        try {
            const response = await http.post<Pick<AuthTokens, 'access_token' | 'access_token_expires_at'>>('/tokens/renew_access', {
                refresh_token: refreshToken.value
            }, {
                skipAuth: true,
                skipErrorHandler: true,
            })

            const { access_token, access_token_expires_at } = response.data
            updateAccessToken(access_token, access_token_expires_at)

            return access_token
        } catch (error) {
            clearTokens()
            throw error
        }
    }

    const ensureAuthenticated = async () => {
        if (hasValidAccessToken()) {
            return true
        }

        if (!hasValidRefreshToken()) {
            clearTokens()
            return false
        }

        try {
            await refreshAccessToken()
            return true
        } catch {
            clearTokens()
            return false
        }
    }

    const ensureAdminAuthenticated = async () => {
        const authenticated = await ensureAuthenticated()

        if (authenticated && isAdmin.value) {
            return true
        }

        clearTokens()
        return false
    }

    return {
        token: readonly(token),
        tokenExpiresAt: readonly(tokenExpiresAt),
        refreshToken: readonly(refreshToken),
        refreshTokenExpiresAt: readonly(refreshTokenExpiresAt),
        currentUser: readonly(currentUser),

        isAuthenticated,
        isTokenExpiring,
        isRefreshTokenValid,
        role,
        isAdmin,

        hasValidAccessToken,
        hasValidRefreshToken,
        login,
        register,
        logout,
        refreshAccessToken,
        updateAccessToken,
        setTokens,
        setCurrentUser,
        clearTokens,
        ensureAuthenticated,
        ensureAdminAuthenticated,
    }
})
