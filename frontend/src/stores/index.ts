export { useAuthStore } from './authStore'
export { useUserStore } from './userStore'
export { useSourceStore } from './sourceStore'
export { useChannelStore } from './channelStore'
export { useMessageStore } from './messageStore'

import { useAuthStore } from './authStore'
import { useUserStore } from './userStore'
import { useSourceStore } from './sourceStore'
import { useChannelStore } from './channelStore'
import { useMessageStore } from './messageStore'

// 登出时调用，重置所有 Store 到初始状态
// 注意：channelTypes 是低频变动的元数据，reset 时保留不重置
export function resetAllStores() {
  useAuthStore().$reset()
  useUserStore().$reset()
  useSourceStore().$reset()
  useChannelStore().$reset()
  useMessageStore().$reset()
}
