import type { AuthResponse } from './auth'

export type AdminUserRole = 'user' | 'vip' | 'admin' | 'platform_admin'
export type AdminUserStatus = 'active' | 'disabled'

export interface AdminUser {
  id: number
  email: string
  role: AdminUserRole
  status: AdminUserStatus
  credits: number
  created_at?: string
  updated_at?: string
}

export interface PaginationResult<T> {
  items: T[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface AdminLoginRequest {
  email: string
  password: string
}

export interface AdminUpdateUserStatusRequest {
  status: AdminUserStatus
}

export interface AdminUpdateUserRoleRequest {
  role: AdminUserRole
}

export interface AdminRechargeRequest {
  user_id: number
  amount: number
  note?: string
}

export interface CreditTransaction {
  id: number
  user_id: number
  amount: number
  type: string
  description?: string
  created_at?: string
}

export interface AdminRechargeResponse {
  user: AdminUser
  transaction: CreditTransaction
}

export type AdminAuthResponse = AuthResponse
