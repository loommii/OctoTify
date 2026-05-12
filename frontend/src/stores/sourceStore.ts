import { defineStore } from 'pinia'
import { ref } from 'vue'

import { Sdk } from '@/api/generated/sdk.gen'

import type {
  DtoSourceDto,
  DtoSourceDetailResponse,
} from '@/api/generated/types.gen'

export const useSourceStore = defineStore('sourceStore', () => {
  // 创建 SDK 实例
  const sdk = new Sdk()

  // Source 列表
  const list = ref<DtoSourceDto[]>([])

  // 当前查看的 Source 详情
  const current = ref<DtoSourceDetailResponse | null>(null)

  // 分页信息
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)

  // 加载状态
  const loading = ref(false)

  // 获取 Source 列表（分页）
  async function fetchList(newPage = 1, newPageSize = 20) {
    try {
      loading.value = true
      page.value = newPage
      pageSize.value = newPageSize

      const response = await sdk.listSources({
        query: {
          page: newPage,
          page_size: newPageSize,
        },
      })

      // 响应拦截器已解包 data 字段，response.data 为业务数据
      const data = response.data as unknown as {
        list: DtoSourceDto[]
        total: number
      }

      list.value = data.list ?? []
      total.value = data.total ?? 0
    } finally {
      loading.value = false
    }
  }

  // 获取 Source 详情（包含已绑定的渠道列表）
  async function fetchDetail(id: number) {
    try {
      loading.value = true
      const response = await sdk.getSourceDetail({
        path: { id },
      })
      current.value = response.data as unknown as DtoSourceDetailResponse
      return current.value
    } finally {
      loading.value = false
    }
  }

  // 创建 Source
  async function create(name: string, description: string, channelIds: number[] = []) {
    const response = await sdk.createSource({
      body: {
        SourceBaseReq: { name, description },
        channel_ids: channelIds,
      },
    })
    // 创建成功后刷新列表
    await fetchList(page.value, pageSize.value)
    return response.data
  }

  // 编辑 Source
  async function update(id: number, name: string, description: string, channelIds: number[] = []) {
    const response = await sdk.updateSource({
      path: { id },
      body: {
        SourceBaseReq: { name, description },
        channel_ids: channelIds,
      },
    })
    // 编辑成功后刷新列表
    await fetchList(page.value, pageSize.value)
    return response.data
  }

  // 删除 Source（需要密码二次验证，由页面传入）
  async function deleteSource(id: number, password: string) {
    const response = await sdk.deleteSource({
      path: { id },
      body: { password },
    })
    // 删除成功后刷新列表
    await fetchList(page.value, pageSize.value)
    return response.data
  }

  // 启用 Source（需要密码二次验证，由页面传入）
  async function enable(id: number, password: string) {
    const response = await sdk.enableSource({
      path: { id },
      body: { password },
    })
    // 启用成功后刷新列表
    await fetchList(page.value, pageSize.value)
    return response.data
  }

  // 禁用 Source（需要密码二次验证，由页面传入）
  async function disable(id: number, password: string) {
    const response = await sdk.disableSource({
      path: { id },
      body: { password },
    })
    // 禁用成功后刷新列表
    await fetchList(page.value, pageSize.value)
    return response.data
  }

  // 获取 Source Token（需要密码二次验证，由页面传入）
  async function fetchToken(id: number, password: string) {
    const response = await sdk.getSourceToken({
      path: { id },
      body: { password },
    })
    return (response.data as unknown as { token: string }).token
  }

  // 重置 Source Token（需要密码二次验证，由页面传入）
  async function resetToken(id: number, password: string) {
    const response = await sdk.resetSourceToken({
      path: { id },
      body: { password },
    })
    return (response.data as unknown as { token: string }).token
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
    fetchDetail,
    create,
    update,
    deleteSource,
    enable,
    disable,
    fetchToken,
    resetToken,
    $reset,
  }
})
