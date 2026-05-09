import axios from 'axios'
import type { AxiosInstance, AxiosRequestConfig, AxiosError, InternalAxiosRequestConfig } from 'axios'
import { v7 as uuidv7 } from 'uuid'
import type { Response } from '@/types/api'
import router from '@/router'
import { useAuthStore } from '@/stores/auth'

// 扩展 AxiosRequestConfig 以支持重试计数
declare module 'axios' {
  interface InternalAxiosRequestConfig {
    _retryCount?: number
  }
}

// 最大重试次数，防止 token 刷新后仍返回 401 时无限重试
const MAX_RETRIES = 1

const apiClient: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Token 刷新并发控制
let isRefreshing = false
let failedQueue: Array<{
  resolve: (value: void) => void
  reject: (error: Error) => void
  config: InternalAxiosRequestConfig
}> = []

function processQueue(error: Error | null) {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve()
    }
  })
  failedQueue = []
}

apiClient.interceptors.request.use((config) => {
  config.headers['X-Request-ID'] = uuidv7()
  config.headers['Accept-Language'] = 'zh-CN,zh;q=0.9,en;q=0.8'

  const authStore = useAuthStore()
  if (authStore.accessToken) {
    config.headers.Authorization = `Bearer ${authStore.accessToken}`
  }

  // 初始化重试次数计数器
  if (config._retryCount === undefined) {
    config._retryCount = 0
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
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retryCount?: number } | undefined

    if (error.response?.status === 401 && originalRequest) {
      // 避免对 refresh 接口本身进行重试
      if (originalRequest.url === '/auth/refresh') {
        const authStore = useAuthStore()
        authStore.clearTokens()
        router.push({ name: 'Login' })
        return Promise.reject(new Error('登录已过期，请重新登录'))
      }

      // 检查是否超过最大重试次数
      const retryCount = originalRequest._retryCount ?? 0
      if (retryCount >= MAX_RETRIES) {
        const authStore = useAuthStore()
        authStore.clearTokens()
        router.push({ name: 'Login' })
        return Promise.reject(new Error('登录已过期，请重新登录'))
      }

      if (isRefreshing) {
        // 正在刷新 token，将请求加入队列
        return new Promise<void>((resolve, reject) => {
          failedQueue.push({ resolve, reject, config: originalRequest })
        })
          .then(() => {
            // token 已刷新，请求拦截器会自动从 authStore 获取新 token
            originalRequest._retryCount = (originalRequest._retryCount ?? 0) + 1
            return apiClient.request(originalRequest)
          })
          .catch((err: unknown) => {
            return Promise.reject(err)
          })
      }

      isRefreshing = true

      const authStore = useAuthStore()
      const refreshToken = authStore.refreshToken
      if (refreshToken) {
        try {
          const refreshResponse = await apiClient.post<Response<{ access_token: string; refresh_token: string; user: unknown }>>('/auth/refresh', {
            refresh_token: refreshToken,
          })

          const refreshData = refreshResponse.data
          if (refreshData.code === 0 && refreshData.data) {
            authStore.setTokens(refreshData.data.access_token, refreshData.data.refresh_token)

            // 更新原始请求头并重发
            originalRequest.headers.Authorization = `Bearer ${refreshData.data.access_token}`

            // 处理队列中的等待请求
            processQueue(null)

            originalRequest._retryCount = retryCount + 1
            return apiClient.request(originalRequest)
          }
        } catch {
          // refresh 请求失败，拒绝队列中所有等待请求
          processQueue(new Error('Token 刷新失败'))
        } finally {
          isRefreshing = false
        }

        authStore.clearTokens()
        router.push({ name: 'Login' })
        return Promise.reject(new Error('登录已过期，请重新登录'))
      }

      isRefreshing = false
      authStore.clearTokens()
      router.push({ name: 'Login' })
      return Promise.reject(new Error('请先登录'))
    } else if (error.code === 'ECONNABORTED') {
      return Promise.reject(new Error('请求超时，请检查网络连接'))
    } else if (error.code === 'ERR_CANCELED') {
      // 请求被主动取消（如 AbortController.abort），透传原始错误供调用方处理
      return Promise.reject(error)
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
