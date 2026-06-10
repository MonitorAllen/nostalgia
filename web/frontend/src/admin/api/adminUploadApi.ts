import adminHttp from './adminHttp'
import type { AdminUploadRequest, AdminUploadResponse } from '../types'

export function uploadAdminFile(data: AdminUploadRequest, signal?: AbortSignal) {
  return adminHttp.post<AdminUploadResponse>('/util/upload_file', data, {
    signal,
    skipErrorHandler: true,
  })
}
