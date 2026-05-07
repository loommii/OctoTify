import { ref } from 'vue'

export interface ConfirmOptions {
  title: string
  description: string
  confirmText?: string
  confirmType?: 'danger' | 'warning' | 'primary'
}

export function useConfirm() {
  const showDialog = ref(false)
  const dialogOptions = ref<ConfirmOptions>({
    title: '',
    description: '',
    confirmText: '确认',
    confirmType: 'primary',
  })
  const actionLoading = ref(false)
  const pendingAction = ref<(() => Promise<void>) | null>(null)

  function requestConfirm(options: ConfirmOptions, action: () => Promise<void>) {
    dialogOptions.value = {
      title: options.title,
      description: options.description,
      confirmText: options.confirmText || '确认',
      confirmType: options.confirmType || 'primary',
    }
    pendingAction.value = action
    showDialog.value = true
  }

  async function handleConfirm() {
    if (!pendingAction.value) return
    actionLoading.value = true
    try {
      await pendingAction.value()
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
    requestConfirm,
    handleConfirm,
  }
}
