/**
 * 微信 ClawBot 绑定 Composable
 *
 * 职责：
 * - 管理绑定流程的状态机
 * - 执行长轮询查询绑定状态
 * - 管理绑定凭证
 *
 * 设计原则：
 * - 单一职责：仅管理绑定状态，不处理 UI 提示
 * - 状态机模式：清晰的状态流转，避免状态不一致
 * - 类型安全：完整的 TypeScript 类型定义
 * - 资源安全：自动清理异步任务和定时器
 *
 * 状态流转：
 * '' → pending → scanned → confirmed (成功)
 * '' → pending → expired (过期)
 * '' → pending → wait (等待扫码)
 */
import { ref, onScopeDispose, type Ref } from 'vue'
import { startWechatClawbotBind, checkBindStatus } from '@/api/channels'
import type { BindCredentialsDTO, BindStatus } from '@/types/api'
import type { AxiosError } from 'axios'

// ============================================================
// 类型定义
// ============================================================

/** Composable 返回值接口 */
export interface UseWechatBindReturn {
  /** 二维码 URL（用于渲染二维码图片） */
  bindQRCodeURL: Ref<string>
  /** 二维码原始值（用于轮询请求） */
  qrcode: Ref<string>
  /** 当前绑定状态 */
  bindStatus: Ref<BindStatus>
  /** 二维码图片是否加载失败 */
  qrcodeLoadError: Ref<boolean>
  /** 长轮询失败重试次数 */
  longPollRetryCount: Ref<number>
  /** 发起绑定流程 */
  startBind: () => Promise<void>
  /** 取消绑定，清除所有状态 */
  cancelBind: () => void
  /** 停止长轮询 */
  stopLongPolling: () => void
  /** 处理二维码图片加载失败 */
  handleQRCodeError: () => void
  /** 获取绑定凭证 */
  getCredentials: () => BindCredentialsDTO | null
}

// ============================================================
// 常量配置
// ============================================================

/** 长轮询超时时间（毫秒） */
const POLLING_TIMEOUT_MS = 65000

/** 长轮询最大重试次数 */
const MAX_RETRY_COUNT = 10

/** 重试基础延迟时间（毫秒） */
const BASE_RETRY_DELAY_MS = 1000

/** 最大重试延迟时间（毫秒） */
const MAX_RETRY_DELAY_MS = 10000

/** 轮询间隔（毫秒） */
const POLLING_INTERVAL_MS = 500

/** 终止状态：轮询遇到这些状态时停止 */
const TERMINAL_STATUSES: readonly BindStatus[] = ['confirmed', 'expired'] as const

// ============================================================
// 工具函数
// ============================================================

/**
 * 延迟指定毫秒数
 *
 * @param ms - 延迟时间（毫秒）
 */
function delay(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

/**
 * 计算指数退避延迟时间
 *
 * 退避策略：1s → 2s → 4s → 8s → 10s（上限）
 *
 * @param retryCount - 当前重试次数（从 1 开始）
 * @returns 延迟时间（毫秒）
 */
function calculateExponentialBackoff(retryCount: number): number {
  const exponentialDelay = BASE_RETRY_DELAY_MS * Math.pow(2, retryCount - 1)
  return Math.min(exponentialDelay, MAX_RETRY_DELAY_MS)
}

/**
 * 校验后端返回的状态值是否为合法的 BindStatus
 *
 * @param value - 后端返回的状态字符串
 * @returns 是否为合法的绑定状态
 */
function isValidBindStatus(value: string): value is BindStatus {
  const validStatuses: readonly string[] = ['wait', 'pending', 'scanned', 'confirmed', 'expired', '']
  return validStatuses.includes(value)
}

// ============================================================
// Composable 实现
// ============================================================

/**
 * 微信 ClawBot 绑定 Composable
 *
 * @returns 绑定状态和控制方法
 */
export function useWechatBind(): UseWechatBindReturn {
  // ==================== 响应式状态 ====================

  /** 二维码 URL（用于渲染二维码图片） */
  const bindQRCodeURL = ref('')

  /** 二维码原始值（用于轮询请求） */
  const qrcode = ref('')

  /** 当前绑定状态 */
  const bindStatus = ref<BindStatus>('')

  /** 二维码图片是否加载失败 */
  const qrcodeLoadError = ref(false)

  /** 长轮询失败重试次数 */
  const longPollRetryCount = ref(0)

  // ==================== 内部状态 ====================

  /** 轮询活动标志 */
  let isPollingActive = false

  /** 当前轮询的 AbortController */
  let currentAbortController: AbortController | null = null

  /** 绑定成功后保存的凭证 */
  let savedCredentials: BindCredentialsDTO | null = null

  // ==================== 核心方法 ====================

  /**
   * 停止轮询
   *
   * 取消进行中的请求并重置轮询标志
   */
  function stopPolling(): void {
    isPollingActive = false

    if (currentAbortController) {
      currentAbortController.abort()
      currentAbortController = null
    }
  }

  /**
   * 重置绑定状态（保留轮询清理）
   *
   * 清除所有响应式状态和内部状态
   */
  function resetState(): void {
    stopPolling()

    bindQRCodeURL.value = ''
    qrcode.value = ''
    bindStatus.value = ''
    qrcodeLoadError.value = false
    longPollRetryCount.value = 0
    savedCredentials = null
  }

  /**
   * 发起绑定流程
   *
   * 步骤：
   * 1. 停止旧轮询（不重置 UI 状态，避免闪烁）
   * 2. 调用后端接口获取二维码
   * 3. 更新状态为 pending
   * 4. 启动长轮询
   *
   * @throws 如果获取二维码失败，抛出错误
   */
  async function startBind(): Promise<void> {
    // 先停止旧轮询，但不重置 UI 状态
    // 避免从 expired → '' → pending 的闪烁
    stopPolling()
    qrcodeLoadError.value = false
    savedCredentials = null

    try {
      const response = await startWechatClawbotBind()

      // 验证响应数据完整性
      if (!response.data?.qrcode || !response.data?.qrcode_url) {
        throw new Error('无效的二维码响应')
      }

      // 更新状态（此时才更新 UI，避免中间态闪烁）
      qrcode.value = response.data.qrcode
      bindQRCodeURL.value = response.data.qrcode_url
      bindStatus.value = 'pending'

      // 启动轮询
      startPolling()
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error('[useWechatBind] 发起绑定失败:', error)
      }
      throw error
    }
  }

  /**
   * 启动长轮询
   *
   * 如果轮询已在运行，则忽略本次调用
   */
  function startPolling(): void {
    if (isPollingActive) {
      return
    }

    isPollingActive = true
    longPollRetryCount.value = 0

    // 不使用 await，轮询在后台运行
    // 错误由 pollingLoop 内部处理，不会冒泡
    pollingLoop()
  }

  /**
   * 长轮询主循环
   *
   * 持续查询绑定状态直到遇到终止状态或用户取消
   *
   * 终止条件：
   * - 绑定成功（confirmed）
   * - 二维码过期（expired）
   * - 用户主动取消
   * - 达到最大重试次数
   */
  async function pollingLoop(): Promise<void> {
    while (isPollingActive && !TERMINAL_STATUSES.includes(bindStatus.value)) {
      try {
        // 创建新的 AbortController
        currentAbortController = new AbortController()

        // 执行轮询请求
        const response = await checkBindStatus(
          qrcode.value,
          POLLING_TIMEOUT_MS,
          currentAbortController.signal
        )

        // 检查轮询是否已被取消
        if (!isPollingActive) {
          return
        }

        // 更新绑定状态
        if (response.data?.status) {
          const status = response.data.status

          // 校验后端返回的状态值
          if (isValidBindStatus(status)) {
            bindStatus.value = status
          } else if (import.meta.env.DEV) {
            console.warn(`[useWechatBind] 收到未知绑定状态: ${status}`)
          }

          // 保存凭证（如果有）
          if (response.data.credentials) {
            savedCredentials = response.data.credentials
          }
        }

        // 如果不是终止状态，等待后继续轮询
        if (!TERMINAL_STATUSES.includes(bindStatus.value)) {
          await delay(POLLING_INTERVAL_MS)
        }
      } catch (error: unknown) {
        // 请求被主动取消，直接退出
        if ((error as AxiosError)?.code === 'ERR_CANCELED') {
          return
        }

        // 轮询已被停止，直接退出
        if (!isPollingActive) {
          return
        }

        // 处理轮询失败（内部消化错误，不冒泡）
        const shouldContinue = await handlePollingError(error)
        if (!shouldContinue) {
          return
        }
      }
    }

    // 轮询结束，重置活动标志
    isPollingActive = false
  }

  /**
   * 处理轮询请求失败
   *
   * 使用指数退避策略重试，达到最大次数后停止轮询
   *
   * @param error - 错误对象
   * @returns 是否应该继续轮询
   */
  async function handlePollingError(error: unknown): Promise<boolean> {
    longPollRetryCount.value++

    if (import.meta.env.DEV) {
      console.error(
        `[useWechatBind] 轮询请求失败 (第 ${longPollRetryCount.value} 次):`,
        error
      )
    }

    // 达到最大重试次数，停止轮询
    if (longPollRetryCount.value >= MAX_RETRY_COUNT) {
      stopPolling()
      bindStatus.value = ''
      return false
    }

    // 指数退避延迟后继续
    const backoffDelay = calculateExponentialBackoff(longPollRetryCount.value)
    await delay(backoffDelay)

    return true
  }

  /**
   * 停止长轮询（公开方法）
   *
   * 取消进行中的请求，重置轮询标志
   */
  function stopLongPolling(): void {
    stopPolling()
  }

  /**
   * 取消绑定
   *
   * 停止轮询并清除所有状态
   */
  function cancelBind(): void {
    resetState()
  }

  /**
   * 处理二维码图片加载失败
   *
   * 防止重复触发错误状态
   */
  function handleQRCodeError(): void {
    if (!qrcodeLoadError.value) {
      qrcodeLoadError.value = true

      if (import.meta.env.DEV) {
        console.error('[useWechatBind] 二维码图片加载失败')
      }
    }
  }

  /**
   * 获取绑定凭证
   *
   * @returns 绑定凭证，如果未绑定成功则返回 null
   */
  function getCredentials(): BindCredentialsDTO | null {
    return savedCredentials
  }

  // ==================== 生命周期 ====================

  /** 组件卸载时自动清理资源 */
  onScopeDispose(() => {
    stopPolling()
  })

  // ==================== 返回值 ====================

  return {
    bindQRCodeURL,
    qrcode,
    bindStatus,
    qrcodeLoadError,
    longPollRetryCount,
    startBind,
    cancelBind,
    stopLongPolling,
    handleQRCodeError,
    getCredentials,
  }
}
