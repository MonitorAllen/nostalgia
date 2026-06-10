export type UserRole = 'admin' | 'visitor'

export interface User {
  id?: string | number
  username: string
  full_name: string
  email: string
  is_email_verified: boolean
  created_at?: string
  create_at?: string
  role?: UserRole
}
