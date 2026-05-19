/**
 * 将 Unix 毫秒时间戳转换为可读日期时间格式
 * @param ts - Unix 毫秒时间戳，0 或 null 返回 '--'
 * @param format - 可选的自定义格式
 * @returns 格式化后的日期时间字符串
 */
export function formatTimestamp(ts: number | null | undefined): string {
  if (!ts) return '--';
  const date = new Date(ts);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

/**
 * 将 Unix 毫秒时间戳转换为日期格式（仅日期，无时间）
 */
export function formatDate(ts: number | null | undefined): string {
  if (!ts) return '--';
  const date = new Date(ts);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });
}

/**
 * 将 Unix 毫秒时间戳转换为相对时间（如 "2小时前"）
 */
export function formatRelativeTime(ts: number | null | undefined): string {
  if (!ts) return '--';
  const now = Date.now();
  const diff = now - ts;

  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  const months = Math.floor(days / 30);
  const years = Math.floor(months / 12);

  if (years > 0) return `${years}年前`;
  if (months > 0) return `${months}个月前`;
  if (days > 0) return `${days}天前`;
  if (hours > 0) return `${hours}小时前`;
  if (minutes > 0) return `${minutes}分钟前`;
  return '刚刚';
}
