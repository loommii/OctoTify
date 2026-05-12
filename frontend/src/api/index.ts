import { createClient } from './generated/client'
import type { AxiosError } from 'axios'

import { useAuth } from '@/composables/useAuth'
import { useAuthStore } from '@/stores/authStore'

// 创建 API 客户端实例
export const apiClient = createClient({
  baseURL: import.meta.env.VITE_API_BASE_URL, // 从环境变量读取 API 基础地址
  timeout: 15000, // 请求超时时间 15 秒
})

// Token 刷新状态标记
let isRefreshing = false

// 请求失败队列，用于在刷新 Token 期间暂存请求
let failedQueue: Array<{
  resolve: (token: string) => void
  reject: (error: unknown) => void
}> = []

// 处理等待队列中的请求
const processQueue = (error: unknown, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (token) {
      // Token 刷新成功，用新 Token 重试请求
      prom.resolve(token)
    } else {
      // Token 刷新失败，拒绝请求
      prom.reject(error)
    }
  })
  failedQueue = []
}

// 请求拦截器：自动添加 Authorization Header
apiClient.instance.interceptors.request.use(
  async (config) => {
    const { accessToken } = useAuth()
    if (accessToken.value) {
      config.headers.Authorization = `Bearer ${accessToken.value}`
    }
    return config
  },
  (error) => Promise.reject(error),
)

// 响应拦截器：统一错误处理 + 响应解包 + 401 自动刷新 Token
apiClient.instance.interceptors.response.use(
  // 响应成功：检查业务状态码并解包 data 字段
  // 后端统一返回格式：{ code: 0, msg: "成功", data: { ... } }
  // 解包后 Store 中直接用 response.data 即可获取业务数据
  (response) => {
    const res = response.data
    if (res && typeof res === 'object' && 'code' in res) {
      // 业务状态码不为 0，表示业务逻辑失败，提取 msg 作为错误信息
      if (res.code !== 0) {
        const error = new Error((res as { msg?: string }).msg || '请求失败')
        ;(error as any).code = (res as { code?: number }).code
        ;(error as any).response = res
        return Promise.reject(error)
      }
      // 业务状态码为 0，自动解包 data 字段
      // 使 Store 中代码更简洁：const data = response.data
      response.data = (res as { data?: unknown }).data ?? res
    }
    return response
  },
  // 响应错误：处理 HTTP 错误和 401 Token 过期
  async (error: AxiosError) => {
    const originalRequest = error.config as any

    // 处理 401 未授权错误，尝试自动刷新 Token
    if (error.response?.status === 401 && !originalRequest.__isRetryRequest) {
      // 如果正在刷新 Token，将请求加入队列等待
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        })
          .then((token: unknown) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            return apiClient.instance(originalRequest)
          })
          .catch((err: unknown) => Promise.reject(err))
      }

      // 标记当前请求为重试请求，避免无限循环
      originalRequest.__isRetryRequest = true
      isRefreshing = true

      try {
        // 调用刷新 Token 接口
        const authStore = useAuthStore()
        const newToken = await authStore.refreshAccessToken()

        // 刷新成功，处理队列中的请求
        processQueue(null, newToken)
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return apiClient.instance(originalRequest)
      } catch (refreshError) {
        // 刷新 Token 失败，清空队列并跳转登录页
        processQueue(refreshError, null)
        const { clearAuth } = useAuth()
        clearAuth()
        window.location.href = '/login'
        return Promise.reject(refreshError)
      } finally {
        isRefreshing = false
      }
    }

    // 其他错误处理，提取后端返回的 msg 字段
    const responseData = error.response?.data as { msg?: string } | undefined
    const errorMsg = responseData?.msg || error.message || '请求失败'

    console.error('API Error:', errorMsg)

    return Promise.reject(error)
  },
)
