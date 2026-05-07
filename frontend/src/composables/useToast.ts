import { ref } from 'vue'

const visible = ref(false)
const message = ref('')
const type = ref<'success' | 'error'>('success')

let timer: number | null = null

export function useToast() {
  function show(msg: string, toastType: 'success' | 'error' = 'success', duration = 3000) {
    if (timer) {
      clearTimeout(timer)
    }

    message.value = msg
    type.value = toastType
    visible.value = true

    timer = window.setTimeout(() => {
      visible.value = false
      timer = null
    }, duration)
  }

  function success(msg: string, duration = 3000) {
    show(msg, 'success', duration)
  }

  function error(msg: string, duration = 3000) {
    show(msg, 'error', duration)
  }

  function hide() {
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
    visible.value = false
  }

  return {
    visible,
    message,
    type,
    show,
    success,
    error,
    hide,
  }
}
