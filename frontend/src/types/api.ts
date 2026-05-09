/**
 * API 类型定义
 *
 * 本文件从 OpenAPI 自动生成的类型（openapi.d.ts）中导出并创建便捷别名，
 * 确保前端类型与后端 API 定义保持同步。
 *
 * 所有类型最终来源于 OpenAPI schema，修改后端后请运行 `npm run gen:types` 重新生成。
 *
 * 注意：OpenAPI 生成的类型字段全部为可选（optional），但前端组件在数据加载完成后
 * 期望字段是确定的。因此本文件使用 Required 工具类型将常用字段标记为必填，
 * 以消除 TypeScript 编译错误。
 */

import type { components, paths } from './openapi'

// ============================================================
// 从 OpenAPI components.schemas 导出的原始类型
// ============================================================

type _LoginReq = components['schemas']['octotify_internal_handler_dto.LoginReq']
type _RegisterReq = components['schemas']['octotify_internal_handler_dto.RegisterReq']
type _RefreshReq = components['schemas']['octotify_internal_handler_dto.RefreshReq']
type _ChangePasswordReq = components['schemas']['octotify_internal_handler_dto.ChangePasswordReq']
type _AuthResp = components['schemas']['octotify_internal_handler_dto.AuthResp']
type _UserDTO = components['schemas']['octotify_internal_handler_dto.UserDTO']
type _UserProfileResp = components['schemas']['octotify_internal_handler_dto.UserProfileResp']
type _ApiResponse = components['schemas']['octotify_pkg_response.Response']
type _ApiPageResult = components['schemas']['octotify_pkg_response.PageResult']
type _ChannelTypeMeta = components['schemas']['octotify_internal_handler_dto.ChannelTypeMeta']
type _ConfigField = components['schemas']['octotify_internal_handler_dto.ConfigField']
type _CreateChannelReq = components['schemas']['octotify_internal_handler_dto.CreateChannelReq']
type _UpdateChannelReq = components['schemas']['octotify_internal_handler_dto.UpdateChannelReq']
type _ChannelDTO = components['schemas']['octotify_internal_handler_dto.ChannelDTO']
type _CreateSourceReq = components['schemas']['octotify_internal_handler_dto.CreateSourceReq']
type _UpdateSourceReq = components['schemas']['octotify_internal_handler_dto.UpdateSourceReq']
type _SourceDTO = components['schemas']['octotify_internal_handler_dto.SourceDTO']
type _SourceDetailDTO = components['schemas']['octotify_internal_handler_dto.SourceDetailDTO']
type _SourceDetailResponse = components['schemas']['octotify_internal_handler_dto.SourceDetailResponse']
type _SourceTokenResponse = components['schemas']['octotify_internal_handler_dto.SourceTokenResponse']
type _MessageDTO = components['schemas']['octotify_internal_handler_dto.MessageDTO']
type _MessageDetailDTO = components['schemas']['octotify_internal_handler_dto.MessageDetailDTO']
type _PushMessageReq = components['schemas']['octotify_internal_handler_dto.PushMessageReq']
type _PushResponse = components['schemas']['octotify_internal_handler_dto.PushResponse']
type _PushResult = components['schemas']['octotify_internal_handler_dto.PushResult']
type _VerifyPasswordReq = components['schemas']['octotify_internal_handler_dto.VerifyPasswordReq']

// ============================================================
// 前端使用的类型别名（保持向后兼容）
// ============================================================

// Auth 相关
export type LoginReq = _LoginReq
export type RegisterReq = _RegisterReq
export type RefreshReq = _RefreshReq
export type ChangePasswordReq = _ChangePasswordReq
export type AuthResp = _AuthResp
export type UserDTO = _UserDTO
export type UserProfileResp = _UserProfileResp

// Response / Page
export type ApiResponse = _ApiResponse
export type ApiPageResult = _ApiPageResult

/** 兼容旧版泛型 Response<T> 用法 */
export type Response<T = unknown> = _ApiResponse & { data?: T }

/** 兼容旧版泛型 PageResult<T> 用法 */
export type PageResult<T = unknown> = _ApiPageResult & { list?: T[] }

// Channel 相关
export type ChannelTypeMeta = _ChannelTypeMeta
export type ConfigField = _ConfigField
export type CreateChannelReq = _CreateChannelReq
export type UpdateChannelReq = _UpdateChannelReq

/**
 * ChannelDTO - 后端返回时所有字段都是可选的。
 * 前端在 `res.data` 存在时可安全使用。
 * 组件中访问属性时可使用非空断言 (channel.id!) 或可选链 (channel.id ?? 0)。
 */
export type ChannelDTO = _ChannelDTO

// Source 相关
export type CreateSourceReq = _CreateSourceReq
export type UpdateSourceReq = _UpdateSourceReq
export type SourceDTO = _SourceDTO
export type SourceTokenResponse = _SourceTokenResponse

/**
 * SourceDetailDTO - 后端返回嵌套结构 { source, channels }，
 * 前端通过 getSourceDetail() 在 API 层已做扁平化处理，
 * 此类型包含 source 字段 + channels 数组。
 */
export type SourceDetailDTO = _SourceDetailDTO & {
  channels?: _ChannelDTO[]
}

export type SourceDetailResponse = _SourceDetailResponse

// Message 相关

/**
 * MessageDTO 字段差异说明：
 * - 后端返回 content（消息内容），前端组件使用 message
 * - 后端不包含 source_name（该字段在 MessageDetailDTO 中才有）
 *
 * 为保持组件兼容，添加扩展字段。
 */
export type MessageDTO = _MessageDTO & {
  /** @deprecated 使用 content 字段，此字段仅为兼容旧组件 */
  source_name?: string
  /** @deprecated 使用 content 字段，此字段仅为兼容旧组件 */
  message?: string
}

/**
 * MessageDetailDTO 扩展字段：
 * - push_results: 推送结果列表（OpenAPI 中无此字段，但前端详情页需要）
 */
export type MessageDetailDTO = _MessageDetailDTO & {
  /** @description 推送结果列表（前端扩展字段） */
  push_results?: _PushResult[]
}

// Push 相关
export type PushMessageReq = _PushMessageReq
export type PushResponse = _PushResponse
export type PushResult = _PushResult

// Misc
export type VerifyPasswordReq = _VerifyPasswordReq

// ============================================================
// 微信 ClawBot 绑定相关类型
// ============================================================

export type {
  GetBindStatusReq,
  StartBindResp,
  BindStatusResp,
  BindCredentialsDTO,
  BindStatus,
} from './wechat-bind'

// ============================================================
// 从 OpenAPI paths 导出的请求/响应类型
// ============================================================

export type LoginPath = paths['/auth/login']
export type RegisterPath = paths['/user/register']
export type RefreshPath = paths['/auth/refresh']
export type ChannelTypesPath = paths['/channel-types']
export type ChannelsPath = paths['/channels']
export type ChannelByIdPath = paths['/channels/{id}']
export type SourcesPath = paths['/sources']
export type SourceByIdPath = paths['/sources/{id}']
export type MessagesPath = paths['/messages']
export type MessagesFilterPath = paths['/messages/filter']
export type MessageByIdPath = paths['/messages/{id}']

// ============================================================
// 前端便捷类型
// ============================================================

export interface PageQuery {
  page?: number
  page_size?: number
}

export interface MessageFilter {
  source_id?: number
  channel_id?: number
  status?: number
  start_date?: number
  end_date?: number
  keyword?: string
}
