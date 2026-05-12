import { ref, computed } from 'vue'
import { apiClient } from '@/api'

// Token 存储 Key
const ACCESS_TOKEN_KEY = 'accessToken'
const REFRESH_TOKEN_KEY = 'refreshToken'

// 响应式 Token 状态
const accessToken = ref<string | null>(null)
const refreshToken = ref<string | null>(null)

// 是否已认证（计算属性）
const isAuthenticated = computed(() => !!accessToken.value)

export function useAuth() {
  // 设置 Access Token
  function setAccessToken(token: string) {
    accessToken.value = token
    localStorage.setItem(ACCESS_TOKEN_KEY, token)
    apiClient.setConfig({
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
  }

  // 设置 Refresh Token
  function setRefreshToken(token: string) {
    refreshToken.value = token
    localStorage.setItem(REFRESH_TOKEN_KEY, token)
  }

  // 清除认证信息（登出时调用）
  function clearAuth() {
    accessToken.value = null
    refreshToken.value = null
    localStorage.removeItem(ACCESS_TOKEN_KEY)
    localStorage.removeItem(REFRESH_TOKEN_KEY)
    apiClient.setConfig({
      headers: {
        Authorization: undefined,
      },
    })
  }

  // 初始化认证状态（应用启动时调用）
  function initAuth() {
    const storedAccessToken = localStorage.getItem(ACCESS_TOKEN_KEY)
    const storedRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)

    if (storedAccessToken) {
      accessToken.value = storedAccessToken
      apiClient.setConfig({
        headers: {
          Authorization: `Bearer ${storedAccessToken}`,
        },
      })
    }

    if (storedRefreshToken) {
      refreshToken.value = storedRefreshToken
    }
  }

  return {
    accessToken,
    refreshToken,
    isAuthenticated,
    setAccessToken,
    setRefreshToken,
    clearAuth,
    initAuth,
  }
}
