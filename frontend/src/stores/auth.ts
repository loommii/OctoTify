import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserDTO } from '@/types/api'
import { getUserProfile, logout as apiLogout } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserDTO | null>(null)
  const accessToken = ref<string>('')
  const refreshToken = ref<string>('')

  const isAuthenticated = computed(() => !!accessToken.value)

  const setTokens = (access: string, refresh: string) => {
    accessToken.value = access
    refreshToken.value = refresh
  }

  const clearTokens = () => {
    accessToken.value = ''
    refreshToken.value = ''
    user.value = null
  }

  const setUser = (userData: UserDTO) => {
    user.value = userData
  }

  const fetchProfile = async () => {
    try {
      const response = await getUserProfile()
      if (response.data && response.data.user) {
        user.value = response.data.user
      }
    } catch {
      clearTokens()
      throw new Error('获取用户信息失败')
    }
  }

  const logout = async () => {
    try {
      await apiLogout()
    } finally {
      clearTokens()
    }
  }

  return {
    user,
    accessToken,
    refreshToken,
    isAuthenticated,
    setTokens,
    clearTokens,
    setUser,
    fetchProfile,
    logout,
  }
}, {
  persist: true,
})
