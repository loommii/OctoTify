import { defineStore } from 'pinia'
import { ref } from 'vue'

import { Sdk } from '@/api/generated/sdk.gen'

import type {
  DtoMessageDto,
  DtoMessageDetailDto,
} from '@/api/generated/types.gen'

export const useMessageStore = defineStore('messageStore', () => {
  // 创建 SDK 实例
  const sdk = new Sdk()

  // 消息列表
  const list = ref<DtoMessageDto[]>([])

  // 当前查看的消息详情
  const current = ref<DtoMessageDetailDto | null>(null)

  // 分页信息
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)

  // 加载状态
  const loading = ref(false)

  // 获取消息列表（分页）
  async function fetchList(newPage = 1, newPageSize = 20) {
    try {
      loading.value = true
      page.value = newPage
      pageSize.value = newPageSize

      const response = await sdk.listMessages({
        query: {
          page: newPage,
          page_size: newPageSize,
        },
      })

      // 响应拦截器已解包 data 字段，response.data 为业务数据
      const data = response.data as unknown as {
        list: DtoMessageDto[]
        total: number
      }

      list.value = data.list ?? []
      total.value = data.total ?? 0
    } finally {
      loading.value = false
    }
  }

  // 筛选消息（多条件组合查询）
  async function filter(params: {
    sourceId?: number
    channelId?: number
    status?: number
    keyword?: string
    startDate?: number
    endDate?: number
    newPage?: number
    newPageSize?: number
  }) {
    try {
      loading.value = true
      const newPage = params.newPage ?? 1
      const newPageSize = params.newPageSize ?? 20
      page.value = newPage
      pageSize.value = newPageSize

      const response = await sdk.filterMessages({
        query: {
          source_id: params.sourceId,
          channel_id: params.channelId,
          status: params.status,
          keyword: params.keyword,
          start_date: params.startDate,
          end_date: params.endDate,
          page: newPage,
          page_size: newPageSize,
        },
      })

      const data = response.data as unknown as {
        list: DtoMessageDto[]
        total: number
      }

      list.value = data.list ?? []
      total.value = data.total ?? 0
    } finally {
      loading.value = false
    }
  }

  // 获取消息详情
  async function fetchDetail(id: number) {
    try {
      loading.value = true
      const response = await sdk.getMessageDetail({
        path: { id },
      })
      current.value = response.data as unknown as DtoMessageDetailDto
      return current.value
    } finally {
      loading.value = false
    }
  }

  // 删除消息
  async function deleteMessage(id: number) {
    const response = await sdk.deleteMessage({
      path: { id },
    })
    // 删除成功后刷新列表
    await fetchList(page.value, pageSize.value)
    return response.data
  }

  // 重置 Store 到初始状态
  function $reset() {
    list.value = []
    current.value = null
    total.value = 0
    page.value = 1
    pageSize.value = 20
    loading.value = false
  }

  return {
    list,
    current,
    total,
    page,
    pageSize,
    loading,
    fetchList,
    filter,
    fetchDetail,
    deleteMessage,
    $reset,
  }
})
