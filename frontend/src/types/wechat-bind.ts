/**
 * 微信 ClawBot 绑定相关类型
 *
 * 所有类型从 OpenAPI 自动生成，此处仅做类型别名导出。
 */

import type { components } from './openapi'

// ============================================================
// 从 OpenAPI 导出的类型别名
// ============================================================

export type GetBindStatusReq = components['schemas']['octotify_internal_handler_dto.GetBindStatusReq']
export type StartBindResp = components['schemas']['octotify_internal_handler_dto.StartBindResp']
export type BindStatusResp = components['schemas']['octotify_internal_handler_dto.BindStatusResp']
export type BindCredentialsDTO = components['schemas']['octotify_internal_handler_dto.BindCredentialsDTO']

// ============================================================
// 前端业务扩展类型
// ============================================================

/** 绑定状态枚举值（iLink 原始状态值 + 空字符串） */
export type BindStatus = 'wait' | 'pending' | 'scanned' | 'confirmed' | 'expired' | ''
