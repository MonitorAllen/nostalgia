import adminHttp from './adminHttp'
import type {
  AdminAIConfigResponse,
  AdminAIConfigUpdateRequest,
  AdminAIModelsRequest,
  AdminAIModelsResponse,
  AdminAIPolishRequest,
  AdminAIPolishResponse
} from '../types'

export function polishAdminText(data: AdminAIPolishRequest) {
  return adminHttp.post<AdminAIPolishResponse>('/ai/polish', data, { skipErrorHandler: true })
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
