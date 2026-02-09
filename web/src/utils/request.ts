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

request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('token')
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
  (error: AxiosError<any>) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')

      const pathname = window.location.pathname
      const search = window.location.search
      if (pathname !== '/login' && pathname !== '/register') {
        const redirect = encodeURIComponent(`${pathname}${search}`)
        window.location.href = `/login?redirect=${redirect}`
      }
    }

    return Promise.reject(error)
  }
)

export default request
