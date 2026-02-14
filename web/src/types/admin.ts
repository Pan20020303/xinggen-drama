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
  reference_id?: string
  service_type?: string
  model?: string
  description?: string
  created_at?: string
}

export interface AdminRechargeResponse {
  user: AdminUser
  transaction: CreditTransaction
}

export type AdminAuthResponse = AuthResponse

export type AdminAIServiceType = 'text' | 'image' | 'video'

// Backend returns masked secrets for admin AI configs:
// - api_key is always empty string
// - api_key_set indicates whether a key exists in storage
export interface AdminAIServiceConfigView {
  id: number
  user_id?: number
  service_type: AdminAIServiceType
  provider: string
  name: string
  base_url: string
  api_key: string
  api_key_set: boolean
  model: string[]
  credit_cost: number
  endpoint: string
  query_endpoint: string
  priority: number
  is_default: boolean
  is_active: boolean
  settings?: string
  created_at?: string
  updated_at?: string
}

export interface AdminCreateAIConfigRequest {
  service_type: AdminAIServiceType
  name: string
  provider: string
  base_url: string
  api_key: string
  model: string[]
  credit_cost?: number
  endpoint?: string
  query_endpoint?: string
  priority?: number
  is_default?: boolean
  settings?: string
}

export interface AdminUpdateAIConfigRequest {
  name: string
  provider: string
  base_url: string
  api_key?: string
  model: string[]
  credit_cost?: number
  endpoint?: string
  query_endpoint: string
  priority: number
  is_default: boolean
  is_active: boolean
  settings?: string
}

export interface AdminTestAIConnectionRequest {
  base_url: string
  api_key: string
  model: string[]
  provider: string
  endpoint?: string
}
