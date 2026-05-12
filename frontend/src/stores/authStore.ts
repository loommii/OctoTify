import { defineStore } from 'pinia'
import { ref } from 'vue'

import { useAuth } from '@/composables/useAuth'
import { Sdk } from '@/api/generated/sdk.gen'

export const useAuthStore = defineStore('authStore', () => {
  // 创建 SDK 实例用于调用认证 API
  const sdk = new Sdk()
  const { setAccessToken, setRefreshToken, clearAuth } = useAuth()

  // 登录加载状态
  const loginLoading = ref(false)

  // 用户登录
  async function login(username: string, password: string) {
    try {
      loginLoading.value = true

      const response = await sdk.login({
        body: {
          AuthCredentials: {
            username,
            password,
          },
        },
      })

      // 响应拦截器已解包 data 字段，response.data 为业务数据
      const data = response.data as unknown as {
        access_token: string
        refresh_token: string
        user: { id: number; username: string }
      }

      // 保存 Token
      setAccessToken(data.access_token)
      setRefreshToken(data.refresh_token)

      return {
        success: true,
        user: data.user,
      }
    } catch (error) {
      // 网络异常或业务错误
      return {
        success: false,
        message: error instanceof Error ? error.message : '网络错误，请稍后重试',
      }
    } finally {
      loginLoading.value = false
    }
  }

  // 用户登出
  async function logout() {
    try {
      await sdk.logout()
    } catch {
      // 忽略登出接口错误，确保本地 Token 被清除
    }
    // 清除 Token
    clearAuth()
    // 重置所有 Store（动态导入避免循环依赖）
    const { resetAllStores } = await import('./index')
    resetAllStores()
  }

  // 刷新 Access Token（由 api/index.ts 的 401 拦截器调用）
  async function refreshAccessToken() {
    const { refreshToken } = useAuth()

    if (!refreshToken.value) {
      throw new Error('No refresh token available')
    }

    const response = await sdk.refreshToken({
      body: {
        refresh_token: refreshToken.value,
      },
    })

    // 响应拦截器已解包 data 字段，response.data 为业务数据
    const data = response.data as unknown as {
      access_token: string
      refresh_token: string
    }

    // 保存新 Token
    setAccessToken(data.access_token)
    setRefreshToken(data.refresh_token)

    return data.access_token
  }

  // 重置 Store 到初始状态
  function $reset() {
    loginLoading.value = false
  }

  return {
    login,
    logout,
    refreshAccessToken,
    loginLoading,
    $reset,
  }
})
