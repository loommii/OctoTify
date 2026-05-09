/**
 * Toast 通知 Composable（单例模式）
 *
 * 职责：
 * - 管理 Toast 通知的显示状态
 * - 提供成功/错误通知的快捷方法
 * - 自动隐藏机制
 *
 * 设计原则：
 * - 单例模式：所有组件共享同一份响应式状态
 * - 全局统一：由 GlobalToast.vue 负责唯一渲染
 * - 类型安全：严格定义 Toast 类型
 *
 * 为什么用单例而不是实例化？
 * - 项目中有 GlobalToast.vue 作为全局 Toast 渲染器
 * - 如果每次 useToast() 创建独立实例，各组件的状态互不关联
 * - 导致 GlobalToast.vue 无法感知其他组件触发的 Toast
 * - 单例模式确保所有组件操作同一份状态，GlobalToast.vue 始终能正确响应
 */
import { ref, type Ref } from 'vue'

// ============================================================
// 类型定义
// ============================================================

/** Toast 类型枚举 */
export type ToastType = 'success' | 'error'

/** Composable 返回值类型 */
export interface UseToastReturn {
  /** Toast 是否可见 */
  visible: Ref<boolean>
  /** Toast 消息内容 */
  message: Ref<string>
  /** Toast 类型 */
  type: Ref<ToastType>
  /** 显示 Toast */
  show: (msg: string, toastType?: ToastType, duration?: number) => void
  /** 显示成功 Toast */
  success: (msg: string, duration?: number) => void
  /** 显示错误 Toast */
  error: (msg: string, duration?: number) => void
  /** 隐藏 Toast */
  hide: () => void
}

// ============================================================
// 单例状态（模块级别，所有调用共享）
// ============================================================

/** Toast 是否可见 */
const visible = ref(false)

/** Toast 消息内容 */
const message = ref('')

/** Toast 类型 */
const type = ref<ToastType>('success')

/** 自动隐藏定时器 */
let timer: ReturnType<typeof setTimeout> | null = null

// ============================================================
// Composable 实现
// ============================================================

/**
 * Toast 通知 Composable
 *
 * 单例模式：所有组件共享同一份响应式状态，
 * 由 GlobalToast.vue 统一渲染，无需各视图自行挂载 Toast 组件
 *
 * @returns Toast 状态和控制方法
 */
export function useToast(): UseToastReturn {
  /**
   * 显示 Toast 通知
   *
   * 先清除已有定时器，再设置新状态和定时器
   *
   * @param msg - 消息内容
   * @param toastType - Toast 类型（success/error）
   * @param duration - 显示时长（毫秒），默认 3000ms
   */
  function show(msg: string, toastType: ToastType = 'success', duration = 3000): void {
    hide()

    message.value = msg
    type.value = toastType
    visible.value = true

    timer = setTimeout(() => {
      hide()
    }, duration)
  }

  /**
   * 显示成功 Toast
   *
   * @param msg - 消息内容
   * @param duration - 显示时长（毫秒）
   */
  function success(msg: string, duration = 3000): void {
    show(msg, 'success', duration)
  }

  /**
   * 显示错误 Toast
   *
   * @param msg - 消息内容
   * @param duration - 显示时长（毫秒）
   */
  function error(msg: string, duration = 3000): void {
    show(msg, 'error', duration)
  }

  /**
   * 隐藏 Toast
   *
   * 清除定时器，关闭显示
   */
  function hide(): void {
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
