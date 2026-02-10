import request from '@/utils/request'
import type { AuthResponse, AuthUser, LoginRequest, RegisterRequest } from '@/types/auth'

export const authAPI = {
  login(data: LoginRequest) {
    return request.post<AuthResponse>('/auth/login', data)
  },

  register(data: RegisterRequest) {
    return request.post<AuthResponse>('/auth/register', data)
  },

  me() {
    return request.get<AuthUser>('/auth/me')
  }
}
