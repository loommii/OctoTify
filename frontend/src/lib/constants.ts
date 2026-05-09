/**
 * 渠道类型中文名称映射
 */
export const CHANNEL_TYPE_NAMES: Record<string, string> = {
  wechat: '企业微信',
  telegram: 'Telegram',
  dingtalk: '钉钉',
  email: '邮件',
  webhook: 'Webhook',
  feishu: '飞书',
}

/**
 * 来源/渠道状态文本映射
 */
export const STATUS_TEXT_MAP: Record<number, string> = {
  1: '正常',
  2: '已停用',
  [-1]: '已删除',
}

/**
 * 来源/渠道状态样式类映射
 */
export const STATUS_CLASS_MAP: Record<number, string> = {
  1: 'status-active',
  2: 'status-disabled',
  [-1]: 'status-deleted',
}

/**
 * 消息状态文本映射
 */
export const MESSAGE_STATUS_TEXT_MAP: Record<number, string> = {
  100: '待推送',
  200: '成功',
  300: '失败',
  [-1]: '已删除',
}

/**
 * 消息状态样式类映射
 */
export const MESSAGE_STATUS_CLASS_MAP: Record<number, string> = {
  100: 'status-partial',
  200: 'status-success',
  300: 'status-failed',
  [-1]: '',
}

/**
 * 获取渠道类型中文名称
 */
export function getChannelTypeName(type: string): string {
  return CHANNEL_TYPE_NAMES[type] || type
}

/**
 * 获取来源/渠道状态文本
 */
export function getStatusText(status: number): string {
  return STATUS_TEXT_MAP[status] || '未知'
}

/**
 * 获取来源/渠道状态样式类
 */
export function getStatusClass(status: number): string {
  return STATUS_CLASS_MAP[status] || ''
}

/**
 * 获取消息状态文本
 */
export function getMessageStatusText(status: number): string {
  return MESSAGE_STATUS_TEXT_MAP[status] || '未知'
}

/**
 * 获取消息状态样式类
 */
export function getMessageStatusClass(status: number): string {
  return MESSAGE_STATUS_CLASS_MAP[status] || ''
}
