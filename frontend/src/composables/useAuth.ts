import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import type { LoginReq, RegisterReq, ChangePasswordReq } from '@/types/api'
import { login, register, changePassword } from '@/api/auth'

export function useAuth() {
  const router = useRouter()
  const authStore = useAuthStore()

  const isLoading = ref(false)
  const errorMessage = ref('')

  const user = computed(() => authStore.user)
  const isAuthenticated = computed(() => authStore.isAuthenticated)

  const handleLogin = async (data: LoginReq) => {
    isLoading.value = true
    errorMessage.value = ''

    try {
      const response = await login(data)
      authStore.setTokens(response.data.access_token, response.data.refresh_token)
      authStore.setUser(response.data.user)
      router.push({ name: 'Dashboard' })
    } catch (error) {
      errorMessage.value = error instanceof Error ? error.message : '登录失败'
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const handleRegister = async (data: RegisterReq) => {
    isLoading.value = true
    errorMessage.value = ''

    try {
      const response = await register(data)
      authStore.setTokens(response.data.access_token, response.data.refresh_token)
      authStore.setUser(response.data.user)
      router.push({ name: 'Dashboard' })
    } catch (error) {
      errorMessage.value = error instanceof Error ? error.message : '注册失败'
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const handleChangePassword = async (data: ChangePasswordReq) => {
    isLoading.value = true
    errorMessage.value = ''

    try {
      await changePassword(data)
      authStore.clearTokens()
      router.push({ name: 'Login' })
    } catch (error) {
      errorMessage.value = error instanceof Error ? error.message : '修改密码失败'
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const handleLogout = async () => {
    await authStore.logout()
    router.push({ name: 'Login' })
  }

  const fetchProfile = async () => {
    try {
      await authStore.fetchProfile()
    } catch {
      router.push({ name: 'Login' })
    }
  }

  return {
    user,
    isLoading,
    errorMessage,
    isAuthenticated,
    handleLogin,
    handleRegister,
    handleChangePassword,
    handleLogout,
    fetchProfile,
  }
}
