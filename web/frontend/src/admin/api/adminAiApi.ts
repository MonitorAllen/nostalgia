import adminHttp from './adminHttp'
import type {
  AdminAIConfigResponse,
  AdminAIConfigUpdateRequest,
  AdminAIModelsRequest,
  AdminAIModelsResponse,
  AdminAIPolishRequest,
  AdminAIPolishResponse
} from '../types'

export const AI_POLISH_REQUEST_TIMEOUT_MS = 310000

export function polishAdminText(data: AdminAIPolishRequest) {
  return adminHttp.post<AdminAIPolishResponse>('/ai/polish', data, {
    skipErrorHandler: true,
    timeout: AI_POLISH_REQUEST_TIMEOUT_MS
  })
}

export function getAdminAIConfig() {
  return adminHttp.get<AdminAIConfigResponse>('/ai/config')
}

export function updateAdminAIConfig(data: AdminAIConfigUpdateRequest) {
  return adminHttp.patch<AdminAIConfigResponse>('/ai/config', data)
}

export function listAdminAIModels(data: AdminAIModelsRequest) {
  return adminHttp.post<AdminAIModelsResponse>('/ai/models', data)
}
