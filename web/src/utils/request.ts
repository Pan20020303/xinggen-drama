import type { AxiosError, AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import axios from 'axios'

interface CustomAxiosInstance extends Omit<AxiosInstance, 'get' | 'post' | 'put' | 'patch' | 'delete'> {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<T>
}

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 600000, // 10分钟超时，匹配后端AI生成接口
  headers: {
    'Content-Type': 'application/json'
  }
}) as CustomAxiosInstance

const USER_TOKEN_KEY = 'token'
const USER_KEY = 'user'
const ADMIN_TOKEN_KEY = 'admin_token'
const ADMIN_USER_KEY = 'admin_user'
const USER_AUTH_PATHS = new Set(['/auth/login', '/auth/register', '/auth/refresh'])
let userRefreshPromise: Promise<string | null> | null = null

function getRequestPath(url?: string): string {
  if (!url) return ''
  if (url.startsWith('http://') || url.startsWith('https://')) {
    try {
      return new URL(url).pathname
    } catch {
      return url
    }
  }
  return url
}

function isAdminRequest(url?: string): boolean {
  const path = getRequestPath(url)
  return path.startsWith('/admin/')
}

function isUserAuthRequest(url?: string): boolean {
  const path = getRequestPath(url)
  return USER_AUTH_PATHS.has(path)
}

async function refreshUserToken(): Promise<string | null> {
  if (userRefreshPromise) {
    return userRefreshPromise
  }

  const token = localStorage.getItem(USER_TOKEN_KEY)
  if (!token) {
    return null
  }

  userRefreshPromise = axios.post('/api/v1/auth/refresh', {}, {
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    timeout: 600000
  })
    .then((resp) => {
      const body = resp.data
      if (!body?.success || !body?.data?.token) {
        return null
      }
      const newToken = body.data.token as string
      localStorage.setItem(USER_TOKEN_KEY, newToken)
      if (body.data.user) {
        localStorage.setItem(USER_KEY, JSON.stringify(body.data.user))
      }
      return newToken
    })
    .catch(() => null)
    .finally(() => {
      userRefreshPromise = null
    })

  return userRefreshPromise
}

function redirectToUserLogin(pathname: string, search: string) {
  localStorage.removeItem(USER_TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
  if (pathname !== '/login' && pathname !== '/register') {
    const redirect = encodeURIComponent(`${pathname}${search}`)
    window.location.href = `/login?redirect=${redirect}`
  }
}

request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = isAdminRequest(config.url)
      ? localStorage.getItem(ADMIN_TOKEN_KEY)
      : localStorage.getItem(USER_TOKEN_KEY)
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => {
    const res = response.data
    if (res.success) {
      return res.data
    } else {
      // 不在这里显示错误提示，让业务代码自行处理
      return Promise.reject(new Error(res.error?.message || '请求失败'))
    }
  },
  async (error: AxiosError<any>) => {
    if (error.response?.status === 401) {
      const originalConfig = (error.config || {}) as (InternalAxiosRequestConfig & { _retry?: boolean })
      const pathname = window.location.pathname
      const search = window.location.search

      if (pathname.startsWith('/admin')) {
        localStorage.removeItem(ADMIN_TOKEN_KEY)
        localStorage.removeItem(ADMIN_USER_KEY)
        if (pathname !== '/admin/login') {
          const redirect = encodeURIComponent(`${pathname}${search}`)
          window.location.href = `/admin/login?redirect=${redirect}`
        }
      } else {
        const canRetry =
          !originalConfig._retry &&
          !!originalConfig.url &&
          !isUserAuthRequest(originalConfig.url) &&
          !!localStorage.getItem(USER_TOKEN_KEY)

        if (canRetry) {
          originalConfig._retry = true
          const newToken = await refreshUserToken()
          if (newToken) {
            originalConfig.headers = originalConfig.headers || {}
            originalConfig.headers.Authorization = `Bearer ${newToken}`
            return request(originalConfig as AxiosRequestConfig)
          }
        }

        redirectToUserLogin(pathname, search)
      }
    }

    return Promise.reject(error)
  }
)

export default request
