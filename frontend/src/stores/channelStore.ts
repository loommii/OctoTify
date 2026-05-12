import { defineStore } from 'pinia'
import { ref } from 'vue'

import { Sdk } from '@/api/generated/sdk.gen'

import type {
  DtoChannelDto,
  DtoChannelTypeMeta,
} from '@/api/generated/types.gen'

export const useChannelStore = defineStore('channelStore', () => {
  // 创建 SDK 实例
  const sdk = new Sdk()

  // Channel 列表
  const list = ref<DtoChannelDto[]>([])

  // 当前查看的 Channel 详情
  const current = ref<DtoChannelDto | null>(null)

  // 渠道类型元数据（低频变动，缓存不重置）
  const channelTypes = ref<DtoChannelTypeMeta[]>([])

  // 分页信息
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)

  // 加载状态
  const loading = ref(false)

  // 获取 Channel 列表（分页）
  async function fetchList(newPage = 1, newPageSize = 20) {
    try {
      loading.value = true
      page.value = newPage
      pageSize.value = newPageSize

      const response = await sdk.listChannels({
        query: {
          page: newPage,
          page_size: newPageSize,
        },
      })

      // 响应拦截器已解包 data 字段，response.data 为业务数据
      const data = response.data as unknown as {
        list: DtoChannelDto[]
        total: number
      }

      list.value = data.list ?? []
      total.value = data.total ?? 0
    } finally {
      loading.value = false
    }
  }

  // 获取 Channel 详情
  async function fetchDetail(id: number) {
    try {
      loading.value = true
      const response = await sdk.getChannelDetail({
        path: { id },
      })
      current.value = response.data as unknown as DtoChannelDto
      return current.value
    } finally {
      loading.value = false
    }
  }

  // 获取渠道类型元数据（带缓存，只加载一次）
  async function fetchChannelTypes() {
    // 已有缓存则直接返回
    if (channelTypes.value.length > 0) {
      return channelTypes.value
    }

    const response = await sdk.getChannelTypes()
    channelTypes.value = (response.data as unknown as DtoChannelTypeMeta[]) ?? []
    return channelTypes.value
  }

  // 创建 Channel
  async function create(
    name: string,
    type: string,
    config: Record<string, Record<string, unknown>>,
  ) {
    const response = await sdk.createChannel({
      body: { name, type, config },
    })
    // 创建成功后刷新列表
    await fetchList(page.value, pageSize.value)
    return response.data
  }

  // 编辑 Channel
  async function update(
    id: number,
    name: string,
    config: Record<string, Record<string, unknown>>,
  ) {
    const response = await sdk.updateChannel({
      path: { id },
      body: { name, config },
    })
    // 编辑成功后刷新列表
    await fetchList(page.value, pageSize.value)
    return response.data
  }

  // 删除 Channel
  async function deleteChannel(id: number) {
    const response = await sdk.deleteChannel({
      path: { id },
    })
    // 删除成功后刷新列表
    await fetchList(page.value, pageSize.value)
    return response.data
  }

  // 启用 Channel
  async function enable(id: number) {
    const response = await sdk.enableChannel({
      path: { id },
    })
    // 启用成功后刷新列表
    await fetchList(page.value, pageSize.value)
    return response.data
  }

  // 禁用 Channel
  async function disable(id: number) {
    const response = await sdk.disableChannel({
      path: { id },
    })
    // 禁用成功后刷新列表
    await fetchList(page.value, pageSize.value)
    return response.data
  }

  // 测试 Channel 连接
  async function test(id: number) {
    const response = await sdk.testChannel({
      path: { id },
    })
    return response.data
  }

  // 重置 Store 到初始状态（保留 channelTypes 缓存）
  function $reset() {
    list.value = []
    current.value = null
    total.value = 0
    page.value = 1
    pageSize.value = 20
    loading.value = false
    // channelTypes 不重置，它是低频变动的元数据
  }

  return {
    list,
    current,
    channelTypes,
    total,
    page,
    pageSize,
    loading,
    fetchList,
    fetchDetail,
    fetchChannelTypes,
    create,
    update,
    deleteChannel,
    enable,
    disable,
    test,
    $reset,
  }
})
