import adminHttp from './adminHttp'
import type { AdminAIConfigResponse, AdminAIPolishRequest, AdminAIPolishResponse } from '../types'

export function polishAdminText(data: AdminAIPolishRequest) {
  return adminHttp.post<AdminAIPolishResponse>('/ai/polish', data, { skipErrorHandler: true })
}

export function getAdminAIConfig() {
  return adminHttp.get<AdminAIConfigResponse>('/ai/config')
}
