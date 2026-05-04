import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosError } from 'axios'
import { v7 as uuidv7 } from 'uuid'
import type { Response } from '@/types/api'
import router from '@/router'

const apiClient: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use((config) => {
  config.headers['X-Request-ID'] = uuidv7()
  config.headers['Accept-Language'] = 'zh-CN,zh;q=0.9,en;q=0.8'

  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }

  return config
})

apiClient.interceptors.response.use(
  (response) => {
    const data = response.data as Response
    if (data.code !== 0) {
      return Promise.reject(new Error(data.msg))
    }
    return response
  },
  async (error: AxiosError) => {
    if (error.response?.status === 401) {
      const refreshToken = localStorage.getItem('refresh_token')
      if (refreshToken) {
        try {
          const refreshResponse = await apiClient.post<Response<{ access_token: string; refresh_token: string; user: unknown }>>('/auth/refresh', {
            refresh_token: refreshToken,
          })

          const refreshData = refreshResponse.data
          if (refreshData.code === 0 && refreshData.data) {
            localStorage.setItem('access_token', refreshData.data.access_token)
            localStorage.setItem('refresh_token', refreshData.data.refresh_token)

            if (error.config) {
              error.config.headers.Authorization = `Bearer ${refreshData.data.access_token}`
              return apiClient.request(error.config)
            }
          }
        } catch {
          // refresh 请求失败
        }

        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        router.push({ name: 'Login' })
        return Promise.reject(new Error('登录已过期，请重新登录'))
      }

      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      router.push({ name: 'Login' })
      return Promise.reject(new Error('请先登录'))
    } else if (error.code === 'ECONNABORTED') {
      return Promise.reject(new Error('请求超时，请检查网络连接'))
    } else if (!error.response) {
      return Promise.reject(new Error('网络连接异常，请检查网络设置'))
    }

    return Promise.reject(error)
  }
)

export const request = <T = unknown>(config: AxiosRequestConfig): Promise<Response<T>> => {
  return apiClient.request<Response<T>>(config).then((res) => res.data)
}

export default apiClient
