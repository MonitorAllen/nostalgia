export const AUTH_STORAGE_KEYS = {
    TOKEN: 'nostalgia_user_token',
    TOKEN_EXPIRES: 'nostalgia_user_token_expires_at',
    REFRESH_TOKEN: 'nostalgia_user_refresh_token',
    REFRESH_TOKEN_EXPIRES: 'nostalgia_user_refresh_token_expires_at',
    USER: 'nostalgia_user_info',
} as const

export const LEGACY_AUTH_STORAGE_KEYS = [
    'nostalgia_access_token',
    'nostalgia_token_expires_at',
    'nostalgia_refresh_token',
    'nostalgia_refresh_token_expires_at',
    'nostalgia_admin_access_token',
    'nostalgia_admin_access_token_expires_at',
    'nostalgia_admin_refresh_token',
    'nostalgia_admin_refresh_token_expires_at',
    'nostalgia_admin_user',
] as const

export type AuthStorage = Pick<Storage, 'removeItem'>

export function cleanupLegacyAuthStorage(storage: AuthStorage = localStorage) {
    LEGACY_AUTH_STORAGE_KEYS.forEach((key) => storage.removeItem(key))
}
