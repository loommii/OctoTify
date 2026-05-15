/**
 * Source / Channel 状态码到 ElTag 类型的映射
 * 1: 正常 -> success
 * 2: 禁用 -> danger
 */
export const entityStatusTypeMap: Record<number, 'success' | 'danger'> = {
  1: 'success',
  2: 'danger',
};

/**
 * Source / Channel 状态码到中文标签的映射
 */
export const entityStatusLabelMap: Record<number, string> = {
  1: '正常',
  2: '禁用',
};

/**
 * Message 状态码到 ElTag 类型的映射
 * 100: 待推送 -> info
 * 200: 成功 -> success
 * 300: 失败 -> danger
 */
export const messageStatusTypeMap: Record<number, 'info' | 'success' | 'danger'> = {
  100: 'info',
  200: 'success',
  300: 'danger',
};

/**
 * Message 状态码到中文标签的映射
 */
export const messageStatusLabelMap: Record<number, string> = {
  100: '待推送',
  200: '成功',
  300: '失败',
};

/**
 * 获取 Source/Channel 状态的 ElTag 类型
 */
export function getEntityStatusTagType(status: number | null | undefined): 'success' | 'danger' {
  return entityStatusTypeMap[status ?? 0] ?? 'danger';
}

/**
 * 获取 Source/Channel 状态的中文标签
 */
export function getEntityStatusLabel(status: number | null | undefined): string {
  if (status === null || status === undefined) return '--';
  return entityStatusLabelMap[status] ?? `未知(${status})`;
}

/**
 * 获取 Message 状态的 ElTag 类型
 */
export function getMessageStatusTagType(status: number | null | undefined): 'info' | 'success' | 'danger' {
  return messageStatusTypeMap[status ?? 0] ?? 'info';
}

/**
 * 获取 Message 状态的中文标签
 */
export function getMessageStatusLabel(status: number | null | undefined): string {
  if (status === null || status === undefined) return '--';
  return messageStatusLabelMap[status] ?? `未知(${status})`;
}
