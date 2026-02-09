export interface AuthUser {
  id: number
  email: string
  role: string
  credits: number
  created_at?: string
  updated_at?: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
}

export interface AuthResponse {
  token: string
  user: AuthUser
}
