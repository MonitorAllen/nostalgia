import type { User } from '@/types/user'

export interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
}