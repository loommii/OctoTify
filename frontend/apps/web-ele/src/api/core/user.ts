import { requestClient } from '#/api/request';

/**
 * 后端返回的原始用户信息
 */
interface UserProfileResp {
  user: {
    id: number;
    username: string;
    created_at_ts: number;
  };
}

/**
 * 自定义用户信息（Vben 必需字段 + 后端特有字段）
 */
export interface UserInfo {
  /** 用户ID */
  userId: string;
  /** 用户名 */
  username: string;
  /** 真实姓名/昵称 */
  realName: string;
  /** 头像地址 */
  avatar: string;
  /** 用户角色 */
  roles: string[];
  /** 用户描述 */
  desc: string;
  /** 首页地址 */
  homePath: string;
  /** 访问令牌 */
  token: string;
  /** 创建时间（Unix 毫秒时间戳） */
  createdAtTs: number;
}

/**
 * 获取用户信息（返回自定义 UserInfo 格式）
 */
export async function getUserInfoApi(): Promise<UserInfo> {
  const result = await requestClient.get<UserProfileResp>('/api/user/profile');
  return {
    userId: String(result.user.id),
    username: result.user.username,
    realName: result.user.username,
    avatar: '',
    roles: [],
    desc: '',
    homePath: '',
    token: '',
    createdAtTs: result.user.created_at_ts,
  };
}

/**
 * 修改密码请求体
 */
export interface ChangePasswordReq {
  old_password: string;
  new_password: string;
}

/**
 * 修改密码
 */
export async function changePasswordApi(data: ChangePasswordReq) {
  return requestClient.put<null>('/api/user/password', data);
}
