import { defineStore } from 'pinia'
import { authAPI } from '@/api/auth'
import type { AuthResponse, AuthUser, ChangePasswordRequest, LoginRequest, RegisterRequest } from '@/types/auth'

const TOKEN_KEY = 'token'
const USER_KEY = 'user'
let refreshPromise: Promise<string | null> | null = null

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: '' as string,
    user: null as AuthUser | null
  }),

  getters: {
    isAuthenticated: (state) => Boolean(state.token)
  },

  actions: {
    initFromStorage() {
      this.token = localStorage.getItem(TOKEN_KEY) || ''
      const rawUser = localStorage.getItem(USER_KEY)
      this.user = rawUser ? JSON.parse(rawUser) : null
    },

    setAuth(payload: AuthResponse) {
      this.token = payload.token
      this.user = payload.user
      localStorage.setItem(TOKEN_KEY, payload.token)
      localStorage.setItem(USER_KEY, JSON.stringify(payload.user))
    },

    async login(payload: LoginRequest) {
      const data = await authAPI.login(payload)
      this.setAuth(data)
      return data
    },

    async register(payload: RegisterRequest) {
      const data = await authAPI.register(payload)
      this.setAuth(data)
      return data
    },

    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(USER_KEY)
    },

    updateCredits(credits: number) {
      if (!this.user) return
      this.user = {
        ...this.user,
        credits
      }
      localStorage.setItem(USER_KEY, JSON.stringify(this.user))
    },

    async refreshMe() {
      if (!this.token) return
      const me = await authAPI.me()
      this.user = me
      localStorage.setItem(USER_KEY, JSON.stringify(me))
    },

    async refreshToken(): Promise<string | null> {
      if (!this.token) return null
      if (!refreshPromise) {
        refreshPromise = authAPI.refresh()
          .then((data) => {
            this.setAuth(data)
            return data.token
          })
          .catch(() => null)
          .finally(() => {
            refreshPromise = null
          })
      }
      return refreshPromise
    },

    async changePassword(payload: ChangePasswordRequest) {
      return authAPI.changePassword(payload)
    }
  }
})
