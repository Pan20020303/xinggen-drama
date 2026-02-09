import { defineStore } from 'pinia'
import { adminAPI } from '@/api/admin'
import type { AuthUser } from '@/types/auth'
import type { AdminLoginRequest } from '@/types/admin'

const ADMIN_TOKEN_KEY = 'admin_token'
const ADMIN_USER_KEY = 'admin_user'

export const useAdminAuthStore = defineStore('adminAuth', {
  state: () => ({
    token: '' as string,
    user: null as AuthUser | null
  }),

  getters: {
    isAuthenticated: (state) => Boolean(state.token)
  },

  actions: {
    initFromStorage() {
      this.token = localStorage.getItem(ADMIN_TOKEN_KEY) || ''
      const rawUser = localStorage.getItem(ADMIN_USER_KEY)
      this.user = rawUser ? JSON.parse(rawUser) : null
    },

    setAuth(payload: { token: string; user: AuthUser }) {
      this.token = payload.token
      this.user = payload.user
      localStorage.setItem(ADMIN_TOKEN_KEY, payload.token)
      localStorage.setItem(ADMIN_USER_KEY, JSON.stringify(payload.user))
    },

    async login(payload: AdminLoginRequest) {
      const data = await adminAPI.login(payload)
      this.setAuth(data)
      return data
    },

    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem(ADMIN_TOKEN_KEY)
      localStorage.removeItem(ADMIN_USER_KEY)
    }
  }
})
