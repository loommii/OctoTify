export interface UserDTO {
  id: number
  username: string
  created_at: number
}

export interface AuthResp {
  access_token: string
  refresh_token: string
  user: UserDTO
}

export interface LoginReq {
  username: string
  password: string
}

export interface RegisterReq {
  username: string
  password: string
}

export interface RefreshReq {
  refresh_token: string
}

export interface ChangePasswordReq {
  old_password: string
  new_password: string
}

export interface Response<T = unknown> {
  code: number
  msg: string
  data?: T
}

export interface PageResult<T = unknown> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface ConfigField {
  name: string
  label: string
  type: string
  required: boolean
  placeholder: string
}

export interface ChannelTypeMeta {
  type: string
  name: string
  description: string
  config_fields: ConfigField[]
}

export interface CreateChannelReq {
  type: string
  name: string
  config: Record<string, unknown>
}

export interface UpdateChannelReq {
  name?: string
  config?: Record<string, unknown>
}

export interface ChannelDTO {
  id: number
  user_id: number
  type: string
  name: string
  config: Record<string, unknown>
  status: number
  created_at: number
  updated_at: number
  last_used_at: number
}

export interface CreateSourceReq {
  name: string
  description?: string
  channel_ids: number[]
}

export interface UpdateSourceReq {
  name?: string
  description?: string
  channel_ids?: number[]
}

export interface SourceDTO {
  id: number
  user_id: number
  name: string
  token: string
  description: string
  status: number
  created_at: number
  updated_at: number
  last_used_at: number
  channels?: ChannelDTO[]
}

export interface SourceTokenResponse {
  token: string
}

export interface SourceDetailDTO extends SourceDTO {
  channels: ChannelDTO[]
}

export interface VerifyPasswordReq {
  password: string
}

export interface MessageDTO {
  id: number
  source_id: number
  source_name: string
  title: string
  message: string
  status: number
  created_at: number
}

export interface MessageDetailDTO extends MessageDTO {
  push_results: PushResult[]
}

export interface PushResult {
  channel_id: number
  channel_name: string
  channel_type: string
  status: number
  error_message: string
  pushed_at: number
}

export interface PushMessageReq {
  title: string
  message: string
}

export interface PushResponse {
  source_id: number
  source_name: string
  total_channels: number
  success_channels: number
  failed_channels: number
  results: PushResult[]
}

export interface PageQuery {
  page?: number
  page_size?: number
}

export interface MessageFilter {
  source_id?: number
  status?: number
  start_time?: number
  end_time?: number
}
