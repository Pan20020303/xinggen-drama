import request from '@/utils/request'
import type { AuthResponse, AuthUser, ChangePasswordRequest, LoginRequest, RegisterRequest } from '@/types/auth'

export const authAPI = {
  login(data: LoginRequest) {
    return request.post<AuthResponse>('/auth/login', data)
  },

  register(data: RegisterRequest) {
    return request.post<AuthResponse>('/auth/register', data)
  },

  refresh() {
    return request.post<AuthResponse>('/auth/refresh')
  },

  changePassword(data: ChangePasswordRequest) {
    return request.put<{ updated: boolean }>('/auth/password', data)
  },

  me() {
    return request.get<AuthUser>('/auth/me')
  }
}
