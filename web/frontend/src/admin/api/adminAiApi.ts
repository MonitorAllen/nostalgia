import adminHttp from './adminHttp'
import type { AdminAIPolishRequest, AdminAIPolishResponse } from '../types'

export function polishAdminText(data: AdminAIPolishRequest) {
  return adminHttp.post<AdminAIPolishResponse>('/ai/polish', data)
}
