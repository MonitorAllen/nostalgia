export const ADMIN_IMAGE_MAX_BYTES = 5 * 1024 * 1024
export const ADMIN_IMAGE_ALLOWED_TYPES = ['image/jpeg', 'image/png'] as const
export const ADMIN_IMAGE_ACCEPT = ADMIN_IMAGE_ALLOWED_TYPES.join(',')

export type AdminImageFileLike = Pick<File, 'size' | 'type'>

export function validateAdminImageFile(file?: AdminImageFileLike | null) {
  if (!file) return '请选择要上传的图片'
  if (!ADMIN_IMAGE_ALLOWED_TYPES.includes(file.type as (typeof ADMIN_IMAGE_ALLOWED_TYPES)[number])) {
    return '仅支持 JPG 或 PNG 图片'
  }
  if (file.size > ADMIN_IMAGE_MAX_BYTES) return '图片不能超过 5 MB'
  return ''
}

export function getAdminUploadErrorMessage(error: unknown, fallback = '请稍后再试') {
  if (typeof error === 'object' && error) {
    const candidate = error as {
      code?: unknown
      message?: unknown
      name?: unknown
      response?: { data?: { error?: string; message?: string } | string }
    }

    if (
      candidate.name === 'AbortError' ||
      candidate.name === 'CanceledError' ||
      candidate.code === 'ERR_CANCELED'
    ) {
      return '上传已取消'
    }

    const data = candidate.response?.data
    if (typeof data === 'string' && data) return data
    if (data && typeof data === 'object') {
      if (data.error) return data.error
      if (data.message) return data.message
    }

    if (typeof candidate.message === 'string' && candidate.message) {
      return candidate.message
    }
  }

  return fallback
}
