/**
 * 渠道创建表单 Composable
 *
 * 职责：
 * - 管理渠道类型选择和表单数据
 * - 表单验证和提交
 * - 确认对话框和 Toast 通知
 *
 * 设计原则：
 * - 单一职责：仅管理表单逻辑
 * - 类型安全：完整的 TypeScript 类型定义
 * - 可组合：与 useWechatBind 配合使用
 */
import { ref, computed, type Ref, type ComputedRef } from 'vue'
import { useRouter } from 'vue-router'
import { getChannelTypes, createChannel } from '@/api/channels'
import { useToast } from './useToast'
import { useConfirm } from './useConfirm'
import type { ChannelTypeMeta, ConfigField, BindCredentialsDTO } from '@/types/api'

// ============================================================
// 类型定义
// ============================================================

/** 渠道表单数据 */
export interface ChannelFormData {
  /** 渠道类型 */
  type: string
  /** 渠道名称 */
  name: string
  /** 配置字段键值对 */
  config: Record<string, string>
}

/** Composable 选项 */
export interface UseChannelFormOptions {
  /** 进入绑定模式的回调 */
  onEnterBindMode?: () => void
  /** 退出绑定模式的回调 */
  onExitBindMode?: () => void
  /** 获取当前绑定状态的回调 */
  getBindStatus?: () => string
}

/** Composable 返回值接口 */
export interface UseChannelFormReturn {
  /** 渠道类型列表 */
  channelTypes: Ref<ChannelTypeMeta[]>
  /** 表单数据 */
  form: Ref<ChannelFormData>
  /** 是否正在提交 */
  submitting: Ref<boolean>
  /** 当前选中的渠道类型 */
  selectedType: ComputedRef<ChannelTypeMeta | undefined>
  /** 当前类型可见的配置字段 */
  visibleConfigFields: ComputedRef<ConfigField[]>
  /** 确认对话框是否可见 */
  showDialog: Ref<boolean>
  /** 确认对话框选项 */
  dialogOptions: Ref<{ title: string; description: string; confirmText?: string; confirmType?: 'danger' | 'warning' | 'primary' }>
  /** 确认按钮是否加载中 */
  actionLoading: Ref<boolean>
  /** 处理确认操作 */
  handleConfirm: () => void
  /** 判断字段是否为凭证字段 */
  isCredentialField: (fieldName: string) => boolean
  /** 对敏感值进行脱敏处理（前3后3，中间 ***） */
  maskValue: (value: string) => string
  /** 选择渠道类型 */
  selectType: (type: string) => void
  /** 提交表单 */
  handleSubmit: (credentials: BindCredentialsDTO | null) => void
  /** 加载渠道类型列表 */
  loadChannelTypes: () => Promise<void>
  /** 返回上一页 */
  goBack: () => void
}

// ============================================================
// 常量配置
// ============================================================

/** 微信 ClawBot 渠道类型标识 */
const WECHAT_CLAWBOT_TYPE = 'wechat_clawbot'

/** 凭证字段名称列表（绑定成功后脱敏展示，不可编辑） */
const CREDENTIAL_FIELD_NAMES = ['bot_token_ciphertext', 'ilink_bot_id', 'ilink_user_id']

/** 创建成功后跳转延迟（毫秒） */
const REDIRECT_DELAY_MS = 1500

// ============================================================
// Composable 实现
// ============================================================

/**
 * 渠道创建表单 Composable
 *
 * @param options - 配置选项
 * @returns 表单状态和控制方法
 */
export function useChannelForm(options: UseChannelFormOptions): UseChannelFormReturn {
  const router = useRouter()
  const toast = useToast()
  const { success: showSuccess, error: showError } = toast
  const { showDialog, dialogOptions, actionLoading, requestConfirm, handleConfirm } = useConfirm()

  // ==================== 响应式状态 ====================

  /** 渠道类型列表 */
  const channelTypes = ref<ChannelTypeMeta[]>([])

  /** 是否正在提交 */
  const submitting = ref(false)

  /** 表单数据 */
  const form = ref<ChannelFormData>({
    type: '',
    name: '',
    config: {},
  })

  // ==================== 计算属性 ====================

  /** 当前选中的渠道类型元数据 */
  const selectedType = computed(() => {
    return channelTypes.value.find((t) => t.type === form.value.type)
  })

  /** 当前类型可见的配置字段（过滤掉绑定流程自动填充的字段） */
  const visibleConfigFields = computed<ConfigField[]>(() => {
    if (!selectedType.value) return []

    const fields = selectedType.value.config_fields ?? []

    // 微信 ClawBot 类型：隐藏 bot_token_nonce（由绑定流程自动填充）
    if (form.value.type === WECHAT_CLAWBOT_TYPE) {
      return fields.filter((field) => field.name !== 'bot_token_nonce')
    }

    return fields
  })

  // ==================== 业务方法 ====================

  /**
   * 判断字段是否为凭证字段
   *
   * 凭证字段在绑定成功后脱敏展示，不可编辑
   *
   * @param fieldName - 字段名称
   * @returns 是否为凭证字段
   */
  function isCredentialField(fieldName: string): boolean {
    const bindStatus = options.getBindStatus?.() || ''
    return bindStatus === 'confirmed'
      && form.value.type === WECHAT_CLAWBOT_TYPE
      && CREDENTIAL_FIELD_NAMES.includes(fieldName)
  }

  /**
   * 对敏感值进行脱敏处理
   *
   * 保留前 3 个和后 3 个字符，中间用 *** 替代
   * 长度不足 7 的值直接返回 ***
   *
   * @param value - 原始值
   * @returns 脱敏后的字符串
   */
  function maskValue(value: string): string {
    if (!value || value.length <= 6) return '***'
    return value.slice(0, 3) + '***' + value.slice(-3)
  }

  /**
   * 选择渠道类型
   *
   * 切换类型时重置配置数据，并触发绑定模式回调
   *
   * @param type - 渠道类型标识
   */
  function selectType(type: string): void {
    if (form.value.type === type) return

    form.value.type = type
    form.value.config = {}

    if (type === WECHAT_CLAWBOT_TYPE) {
      options.onEnterBindMode?.()
    } else {
      options.onExitBindMode?.()
    }
  }

  /**
   * 验证表单数据
   *
   * @returns 表单是否有效
   */
  function validateForm(): boolean {
    if (!form.value.type) {
      showError('请选择渠道类型')
      return false
    }

    if (!form.value.name) {
      showError('请输入渠道名称')
      return false
    }

    // 微信 ClawBot 类型：必须完成绑定
    if (form.value.type === WECHAT_CLAWBOT_TYPE) {
      const bindStatus = options.getBindStatus?.() || ''
      if (bindStatus !== 'confirmed') {
        showError('请先完成微信绑定')
        return false
      }
      return true
    }

    // 其他类型：校验必填字段
    if (!selectedType.value) return false

    const fields = selectedType.value.config_fields ?? []
    for (const field of fields) {
      if (field.required && !form.value.config[field.name ?? '']) {
        showError(`请填写 ${field.label}`)
        return false
      }
    }

    return true
  }

  /**
   * 标准化配置数据
   *
   * 将字符串值转换为对应类型（如数字字段转为 Number）
   *
   * @param config - 原始配置键值对
   * @param fields - 字段定义列表
   * @returns 标准化后的配置
   */
  function normalizeConfig(
    config: Record<string, string>,
    fields: ConfigField[]
  ): Record<string, unknown> {
    const normalized: Record<string, unknown> = {}

    for (const field of fields) {
      const fieldName = field.name ?? ''
      const value = config[fieldName]

      if (field.type === 'number' && value !== '' && value !== undefined) {
        normalized[fieldName] = Number(value)
      } else {
        normalized[fieldName] = value
      }
    }

    return normalized
  }

  /**
   * 提交表单
   *
   * 验证表单后弹出确认对话框，确认后创建渠道
   *
   * @param credentials - 微信绑定凭证（仅 wechat_clawbot 类型需要）
   */
  function handleSubmit(credentials: BindCredentialsDTO | null): void {
    if (!validateForm()) return
    if (!selectedType.value) return

    // 合并凭证到配置中
    const configToUse = credentials
      ? { ...form.value.config, ...credentials }
      : form.value.config

    const fields = selectedType.value.config_fields ?? []
    const normalizedConfig = normalizeConfig(configToUse, fields)

    requestConfirm(
      {
        title: '创建推送渠道',
        description: `确定要创建渠道 "${form.value.name}" 吗？`,
        confirmText: '创建',
        confirmType: 'primary',
      },
      async () => {
        submitting.value = true
        try {
          await createChannel({
            type: form.value.type,
            name: form.value.name,
            config: normalizedConfig,
          })
          showSuccess('创建成功', REDIRECT_DELAY_MS)
          setTimeout(() => {
            router.push({ name: 'ChannelList' })
          }, REDIRECT_DELAY_MS)
        } catch (err) {
          if (import.meta.env.DEV) {
            console.error('[useChannelForm] 创建渠道失败:', err)
          }
          showError('创建失败，请重试')
        } finally {
          submitting.value = false
        }
      }
    )
  }

  /**
   * 加载渠道类型列表
   */
  async function loadChannelTypes(): Promise<void> {
    try {
      const res = await getChannelTypes()
      if (res.data) {
        channelTypes.value = res.data
      }
    } catch (err) {
      if (import.meta.env.DEV) {
        console.error('[useChannelForm] 加载渠道类型失败:', err)
      }
    }
  }

  /**
   * 返回上一页
   */
  function goBack(): void {
    router.back()
  }

  // ==================== 返回值 ====================

  return {
    channelTypes,
    form,
    submitting,
    selectedType,
    visibleConfigFields,
    showDialog,
    dialogOptions,
    actionLoading,
    handleConfirm,
    isCredentialField,
    maskValue,
    selectType,
    handleSubmit,
    loadChannelTypes,
    goBack,
  }
}
