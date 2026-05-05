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
  return request<ChannelDTO>({
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
  return request<{ status: string; message: string }>({
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

export function createSource(data: CreateSourceReq) {
  return request<SourceDTO>({
    url: '/sources',
    method: 'POST',
    data,
  })
}

export function updateSource(id: number, data: UpdateSourceReq) {
  return request<SourceDTO>({
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

export async function getSourceDetail(id: number) {
  const res = await request<{ source: SourceDetailDTO; channels: ChannelDTO[] }>({
    url: `/sources/${id}`,
    method: 'GET',
  })
  
  // 在 API 层做数据转换，将后端嵌套结构扁平化
  if (res.data) {
    return {
      ...res,
      data: {
        ...res.data.source,
        channels: res.data.channels,
      } as SourceDetailDTO,
    }
  }
  
  return res as unknown as ReturnType<typeof request<SourceDetailDTO>>
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
