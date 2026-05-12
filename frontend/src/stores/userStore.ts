import { defineStore } from 'pinia'
import { ref } from 'vue'

import { Sdk } from '@/api/generated/sdk.gen'

import type { DtoUserDto } from '@/api/generated/types.gen'

export const useUserStore = defineStore('userStore', () => {
  // 创建 SDK 实例
  const sdk = new Sdk()

  // 用户信息
  const user = ref<DtoUserDto | null>(null)

  // 加载状态
  const loading = ref(false)

  // 获取用户资料
  async function fetchUserProfile() {
    try {
      loading.value = true
      const response = await sdk.getUserProfile()
      // 响应拦截器已解包 data 字段，response.data 为业务数据
      const data = response.data as unknown as { user: DtoUserDto }
      user.value = data.user
      return data.user
    } finally {
      loading.value = false
    }
  }

  // 修改密码
  async function changePassword(oldPassword: string, newPassword: string) {
    const response = await sdk.changePassword({
      body: {
        old_password: oldPassword,
        new_password: newPassword,
      },
    })
    return response.data
  }

  // 清空用户信息（登出时调用）
  function clearUser() {
    user.value = null
  }

  // 重置 Store 到初始状态
  function $reset() {
    user.value = null
    loading.value = false
  }

  return {
    user,
    loading,
    fetchUserProfile,
    changePassword,
    clearUser,
    $reset,
  }
})
