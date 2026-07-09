import { requestClient } from '#/api/request';

export namespace ChannelApi {
  export interface ChannelDTO {
    id: number;
    user_id: number;
    type: string;
    name: string;
    config: Record<string, any>;
    status: number;
    created_at_ts: number;
    updated_at_ts: number;
    last_used_at_ts: number;
  }

  export interface ListResult {
    list: ChannelDTO[];
    total: number;
    page: number;
    page_size: number;
  }

  /** 渠道类型元数据 */
  export interface ConfigField {
    name: string;
    label: string;
    type: string;
    required: boolean;
    placeholder: string;
  }

  export interface ChannelTypeMeta {
    type: string;
    name: string;
    description: string;
    config_fields: ConfigField[];
  }

  /** 创建渠道请求 */
  export interface CreateChannelReq {
    type: string;
    name: string;
    config: Record<string, any>;
  }

  /** 编辑渠道请求 */
  export interface UpdateChannelReq {
    name: string;
    config: Record<string, any>;
  }
}

/**
 * 获取渠道类型元数据
 */
export async function getChannelTypesApi() {
  return requestClient.get<ChannelApi.ChannelTypeMeta[]>('/api/channel-types');
}

/**
 * 获取推送渠道列表
 */
export async function getChannelListApi(params?: { page?: number; page_size?: number }) {
  return requestClient.get<ChannelApi.ListResult>('/api/channels', { params });
}

/**
 * 获取所有渠道（用于来源绑定）
 */
export async function getAllChannelsApi() {
  return requestClient.get<ChannelApi.ListResult>('/api/channels', {
    params: { page: 1, page_size: 100 },
  });
}

/**
 * 创建渠道
 */
export async function createChannelApi(data: ChannelApi.CreateChannelReq) {
  return requestClient.post<ChannelApi.ChannelDTO>('/api/channels', data);
}

/**
 * 获取渠道详情
 */
export async function getChannelDetailApi(id: number) {
  return requestClient.get<ChannelApi.ChannelDTO>(`/api/channels/${id}`);
}

/**
 * 编辑渠道
 */
export async function updateChannelApi(id: number, data: ChannelApi.UpdateChannelReq) {
  return requestClient.put<null>(`/api/channels/${id}`, data);
}

/**
 * 删除渠道
 */
export async function deleteChannelApi(id: number) {
  return requestClient.delete<null>(`/api/channels/${id}`);
}

/**
 * 启用渠道
 */
export async function enableChannelApi(id: number) {
  return requestClient.put(`/api/channels/${id}/enable`);
}

/**
 * 停用渠道
 */
export async function disableChannelApi(id: number) {
  return requestClient.put(`/api/channels/${id}/disable`);
}

/**
 * 测试连接
 */
export async function testChannelApi(id: number) {
  return requestClient.post(`/api/channels/${id}/test`);
}
