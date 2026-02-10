import request from '@/utils/request'
import type {
  AdminAuthResponse,
  AdminAIServiceConfigView,
  AdminCreateAIConfigRequest,
  AdminLoginRequest,
  AdminRechargeRequest,
  AdminRechargeResponse,
  AdminTestAIConnectionRequest,
  AdminUpdateAIConfigRequest,
  AdminUpdateUserRoleRequest,
  AdminUpdateUserStatusRequest,
  AdminUser,
  CreditTransaction,
  PaginationResult
} from '@/types/admin'

export const adminAPI = {
  login(data: AdminLoginRequest) {
    return request.post<AdminAuthResponse>('/admin/auth/login', data)
  },

  listUsers(params?: { page?: number; page_size?: number }) {
    return request.get<PaginationResult<AdminUser>>('/admin/users', { params })
  },

  updateUserStatus(userId: number, data: AdminUpdateUserStatusRequest) {
    return request.patch<AdminUser>(`/admin/users/${userId}/status`, data)
  },

  updateUserRole(userId: number, data: AdminUpdateUserRoleRequest) {
    return request.patch<AdminUser>(`/admin/users/${userId}/role`, data)
  },

  recharge(data: AdminRechargeRequest) {
    return request.post<AdminRechargeResponse>('/admin/billing/recharge', data)
  },

  listTransactions(params?: { user_id?: number; page?: number; page_size?: number }) {
    return request.get<PaginationResult<CreditTransaction>>('/admin/billing/transactions', { params })
  },

  listAIConfigs(params?: { service_type?: 'text' | 'image' | 'video' }) {
    return request.get<AdminAIServiceConfigView[]>('/admin/ai-configs', { params })
  },

  createAIConfig(data: AdminCreateAIConfigRequest) {
    return request.post<AdminAIServiceConfigView>('/admin/ai-configs', data)
  },

  updateAIConfig(id: number, data: AdminUpdateAIConfigRequest) {
    return request.put<AdminAIServiceConfigView>(`/admin/ai-configs/${id}`, data)
  },

  deleteAIConfig(id: number) {
    return request.delete(`/admin/ai-configs/${id}`)
  },

  testAIConfig(data: AdminTestAIConnectionRequest) {
    return request.post('/admin/ai-configs/test', data)
  }
}
