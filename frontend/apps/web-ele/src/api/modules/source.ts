import { requestClient } from '#/api/request';

export namespace SourceApi {
  export interface SourceDTO {
    id: number;
    user_id: number;
    name: string;
    token?: string;
    description: string;
    status: number;
    created_at_ts: number;
  }

  /** 来源详情（含渠道列表和最后使用时间） */
  export interface SourceDetailDTO {
    id: number;
    user_id: number;
    name: string;
    token: string;
    description: string;
    status: number;
    created_at_ts: number;
    updated_at_ts: number;
    last_used_at_ts: number;
  }

  /** 来源详情响应（包含已绑定渠道列表） */
  export interface SourceDetailResponse {
    source: SourceDetailDTO;
    channels: {
      id: number;
      name: string;
      type: string;
    }[];
  }

  export interface ListResult {
    list: SourceDTO[];
    total: number;
    page: number;
    page_size: number;
  }

  /** 创建来源请求 */
  export interface CreateSourceReq {
    name: string;
    description?: string;
    channel_ids?: number[];
  }

  /** 编辑来源请求 */
  export interface UpdateSourceReq {
    name: string;
    description?: string;
    channel_ids?: number[];
  }

  /** Token 响应 */
  export interface TokenResponse {
    token: string;
  }

  /** 二次验证密码请求 */
  export interface PasswordRequest {
    password: string;
  }
}

/**
 * 获取消息来源列表
 */
export async function getSourceListApi(params?: { page?: number; page_size?: number }) {
  return requestClient.get<SourceApi.ListResult>('/api/sources', { params });
}

/**
 * 获取所有来源（用于消息筛选）
 */
export async function getAllSourcesApi() {
  return requestClient.get<SourceApi.ListResult>('/api/sources', {
    params: { page: 1, page_size: 100 },
  });
}

/**
 * 创建消息来源
 */
export async function createSourceApi(data: SourceApi.CreateSourceReq) {
  return requestClient.post<SourceApi.SourceDTO>('/api/sources', data);
}

/**
 * 获取来源详情（含已绑定渠道列表）
 */
export async function getSourceDetailApi(id: number) {
  return requestClient.get<SourceApi.SourceDetailResponse>(`/api/sources/${id}`);
}

/**
 * 编辑消息来源
 */
export async function updateSourceApi(id: number, data: SourceApi.UpdateSourceReq) {
  return requestClient.put<null>(`/api/sources/${id}`, data);
}

/**
 * 删除消息来源（需密码二次验证）
 */
export async function deleteSourceApi(id: number, data: SourceApi.PasswordRequest) {
  return requestClient.delete<null>(`/api/sources/${id}`, { data });
}

/**
 * 启用消息来源（需密码二次验证）
 */
export async function enableSourceApi(id: number, data: SourceApi.PasswordRequest) {
  return requestClient.put<null>(`/api/sources/${id}/enable`, data);
}

/**
 * 停用消息来源（需密码二次验证）
 */
export async function disableSourceApi(id: number, data: SourceApi.PasswordRequest) {
  return requestClient.put<null>(`/api/sources/${id}/disable`, data);
}

/**
 * 查看来源令牌（需密码二次验证）
 */
export async function getSourceTokenApi(id: number, data: SourceApi.PasswordRequest) {
  return requestClient.post<SourceApi.TokenResponse>(`/api/sources/${id}/token`, data);
}

/**
 * 重置来源令牌（需密码二次验证）
 */
export async function resetSourceTokenApi(id: number, data: SourceApi.PasswordRequest) {
  return requestClient.put<SourceApi.TokenResponse>(`/api/sources/${id}/token`, data);
}
