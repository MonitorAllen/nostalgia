export const ADMIN_BASE_PATH = '/backend'
export const ADMIN_LOGIN_PATH = `${ADMIN_BASE_PATH}/login`
export const ADMIN_ARTICLES_PATH = `${ADMIN_BASE_PATH}/articles`

export const buildAdminLoginRedirect = (current: string) => {
  if (current.startsWith(ADMIN_LOGIN_PATH)) return ADMIN_LOGIN_PATH

  return `${ADMIN_LOGIN_PATH}?redirect=${encodeURIComponent(current)}`
}
