export type UserRole = 'admin' | 'learner'
export type UserStatus = 'active' | 'disabled'

export interface User {
  id: string
  email: string
  username: string
  role: UserRole
  status: UserStatus
  timezone: string
  last_login_at?: string
  created_at: string
}

export type InvitationStatus = 'pending' | 'accepted' | 'expired' | 'revoked'

export interface Invitation {
  id: string
  email: string
  status: InvitationStatus
  invited_by: string
  invited_by_username?: string
  expires_at: string
  accepted_at?: string
  revoked_at?: string
  created_at: string
  invitation_url?: string
  invitation_token?: string
}

export interface AuthResponse {
  access_token: string
  expires_in: number
  user: User
}

export interface ApiErrorDetail {
  field?: string
  message: string
}

export interface ApiErrorResponse {
  error: {
    code: string
    message: string
    details?: ApiErrorDetail[]
  }
}
