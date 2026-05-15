import { requestClient } from '#/api/request';

export namespace MessageApi {
  export interface MessageDTO {
    id: number;
    source_id: number;
    source_name?: string;
    channel_id: number;
    channel_name?: string;
    channel_type?: string;
    title: string;
    content: string;
    status: number;
    created_at_ts: number;
    updated_at_ts: number;
  }

  export interface ListResult {
    list: MessageDTO[];
    total: number;
    page: number;
    page_size: number;
  }

  export interface FilterParams {
    page?: number;
    page_size?: number;
    source_id?: number;
    channel_id?: number;
    status?: number;
    start_date?: number;
    end_date?: number;
    keyword?: string;
  }
}

/**
 * 获取消息列表
 */
export async function getMessageListApi(params?: { page?: number; page_size?: number }) {
  return requestClient.get<MessageApi.ListResult>('/api/messages', { params });
}

/**
 * 筛选消息
 */
export async function filterMessagesApi(params?: MessageApi.FilterParams) {
  return requestClient.get<MessageApi.ListResult>('/api/messages/filter', { params });
}

/**
 * 获取消息详情
 */
export async function getMessageDetailApi(id: number) {
  return requestClient.get<MessageApi.MessageDTO>(`/api/messages/${id}`);
}

/**
 * 删除消息
 */
export async function deleteMessageApi(id: number) {
  return requestClient.delete<null>(`/api/messages/${id}`);
}
