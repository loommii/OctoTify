import { ref } from 'vue'

export interface PasswordConfirmOptions {
  title: string
  description: string
  confirmText?: string
}

export function usePasswordConfirm() {
  const showDialog = ref(false)
  const dialogOptions = ref<PasswordConfirmOptions>({
    title: '',
    description: '',
    confirmText: '确认',
  })
  const actionLoading = ref(false)
  const pendingAction = ref<((password: string) => Promise<void>) | null>(null)

  function requestPassword(options: PasswordConfirmOptions, action: (password: string) => Promise<void>) {
    // 如果已有对话框打开，拒绝新请求，防止 pendingAction 被覆盖
    if (showDialog.value) {
      return
    }
    dialogOptions.value = {
      title: options.title,
      description: options.description,
      confirmText: options.confirmText || '确认',
    }
    pendingAction.value = action
    showDialog.value = true
  }

  async function handleConfirm(password: string) {
    if (!pendingAction.value) return
    actionLoading.value = true
    try {
      await pendingAction.value(password)
      showDialog.value = false
      pendingAction.value = null
    } catch {
      // 失败时保留 pendingAction，允许用户重试
    } finally {
      actionLoading.value = false
    }
  }

  return {
    showDialog,
    dialogOptions,
    actionLoading,
    requestPassword,
    handleConfirm,
  }
}
