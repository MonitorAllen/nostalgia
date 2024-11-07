// 本地缓存服务

const PREFIX = 'nostalgia_'

// user 模块
const USER_PREFIX = `${PREFIX}user_`
const USER_TOKEN = `${USER_PREFIX}token`
const USER_TOKEN_EXPIRES_AT = `${USER_PREFIX}token_expires_at`
const USER_REFRESH_TOKEN = `${USER_PREFIX}refresh_token`
const USER_REFRESH_TOKEN_EXPIRES_AT = `${USER_PREFIX}refresh_token_expires_at`
const USER_INFO = `${USER_PREFIX}info`

// 储存
const set = (key: string, val: any) => {
  localStorage.setItem(key, val)
}

// 读取
const get = (key: string): string => localStorage.getItem(key) ?? ''

export default {
  set,
  get,
  USER_TOKEN,
  USER_TOKEN_EXPIRES_AT,
  USER_REFRESH_TOKEN,
  USER_REFRESH_TOKEN_EXPIRES_AT,
  USER_INFO,
}
