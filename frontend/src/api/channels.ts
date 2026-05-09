import { request } from '@/lib/request'
import type {
  PageResult,
  ChannelTypeMeta,
  ChannelDTO,
  CreateChannelReq,
  UpdateChannelReq,
  SourceDTO,
  SourceDetailDTO,
  CreateSourceReq,
  UpdateSourceReq,
  SourceTokenResponse,
  MessageDTO,
  MessageDetailDTO,
  VerifyPasswordReq,
  SourceDetailResponse,
  GetBindStatusReq,
  BindStatusResp,
  StartBindResp,
} from '@/types/api'

export function getChannelTypes() {
  return request<ChannelTypeMeta[]>({
    url: '/channel-types',
    method: 'GET',
  })
}

export function createChannel(data: CreateChannelReq) {
  return request<ChannelDTO>({
    url: '/channels',
    method: 'POST',
    data,
  })
}

export function updateChannel(id: number, data: UpdateChannelReq) {
  return request<null>({
    url: `/channels/${id}`,
    method: 'PUT',
    data,
  })
}

export function listChannels(page = 1, pageSize = 20) {
  return request<PageResult<ChannelDTO>>({
    url: '/channels',
    method: 'GET',
    params: { page, page_size: pageSize },
  })
}

export function getChannelDetail(id: number) {
  return request<ChannelDTO>({
    url: `/channels/${id}`,
    method: 'GET',
  })
}

export function testChannel(id: number) {
  return request<null>({
    url: `/channels/${id}/test`,
    method: 'POST',
  })
}

export function disableChannel(id: number) {
  return request<null>({
    url: `/channels/${id}/disable`,
    method: 'PUT',
  })
}

export function enableChannel(id: number) {
  return request<null>({
    url: `/channels/${id}/enable`,
    method: 'PUT',
  })
}

export function deleteChannel(id: number) {
  return request<null>({
    url: `/channels/${id}`,
    method: 'DELETE',
  })
}

export function startWechatClawbotBind() {
  return request<StartBindResp>({
    url: '/channels/wechat-clawbot/bind',
    method: 'POST',
  })
}

// 长轮询查询绑定状态（最长等待 60 秒）
// 注意：此处的 timeout 配置会覆盖 request.ts 中 Axios 实例的全局默认超时（30s）
// Axios 的请求级配置优先级高于实例级配置，因此 65s 超时能正确生效
export function checkBindStatus(qrcode: string, timeoutMs: number = 65000, signal?: AbortSignal) {
  const data: GetBindStatusReq = { qrcode }
  return request<BindStatusResp>({
    url: '/channels/wechat-clawbot/bind/status',
    method: 'POST',
    data,
    // 设置更长的超时时间以匹配后端长轮询超时
    timeout: timeoutMs,
    signal,
  })
}

export function createSource(data: CreateSourceReq) {
  return request<SourceDTO>({
    url: '/sources',
    method: 'POST',
    data,
  })
}

export function updateSource(id: number, data: UpdateSourceReq) {
  return request<null>({
    url: `/sources/${id}`,
    method: 'PUT',
    data,
  })
}

export function listSources(page = 1, pageSize = 20) {
  return request<PageResult<SourceDTO>>({
    url: '/sources',
    method: 'GET',
    params: { page, page_size: pageSize },
  })
}

/**
 * 获取来源详情，将后端返回的嵌套结构（source + channels）
 * 扁平化为 SourceDetailDTO 供前端使用
 */
export async function getSourceDetail(id: number) {
  const res = await request<SourceDetailResponse>({
    url: `/sources/${id}`,
    method: 'GET',
  })

  if (res.data) {
    const { source, channels } = res.data
    return {
      ...res,
      data: {
        ...source,
        channels: channels ?? [],
      } as SourceDetailDTO,
    }
  }

  return {
    ...res,
    data: null as SourceDetailDTO | null,
  }
}

export function getSourceToken(id: number, password: string) {
  return request<SourceTokenResponse>({
    url: `/sources/${id}/token`,
    method: 'POST',
    data: { password } satisfies VerifyPasswordReq,
  })
}

export function resetSourceToken(id: number, password: string) {
  return request<SourceTokenResponse>({
    url: `/sources/${id}/token`,
    method: 'PUT',
    data: { password } satisfies VerifyPasswordReq,
  })
}

export function disableSource(id: number, password: string) {
  return request<null>({
    url: `/sources/${id}/disable`,
    method: 'PUT',
    data: { password } satisfies VerifyPasswordReq,
  })
}

export function enableSource(id: number, password: string) {
  return request<null>({
    url: `/sources/${id}/enable`,
    method: 'PUT',
    data: { password } satisfies VerifyPasswordReq,
  })
}

export function deleteSource(id: number, password: string) {
  return request<null>({
    url: `/sources/${id}`,
    method: 'DELETE',
    data: { password } satisfies VerifyPasswordReq,
  })
}

export function listMessages(page = 1, pageSize = 20) {
  return request<PageResult<MessageDTO>>({
    url: '/messages',
    method: 'GET',
    params: { page, page_size: pageSize },
  })
}

export function filterMessages(params: Record<string, unknown>) {
  return request<PageResult<MessageDTO>>({
    url: '/messages/filter',
    method: 'GET',
    params,
  })
}

export function getMessageDetail(id: number) {
  return request<MessageDetailDTO>({
    url: `/messages/${id}`,
    method: 'GET',
  })
}

export function deleteMessage(id: number) {
  return request<null>({
    url: `/messages/${id}`,
    method: 'DELETE',
  })
}
