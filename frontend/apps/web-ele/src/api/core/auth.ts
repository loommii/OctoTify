import { requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    password?: string;
    username?: string;
  }

  /** 登录接口返回值 */
  export interface LoginResult {
    access_token: string;
    refresh_token: string;
  }

  export interface RefreshTokenResult {
    access_token: string;
    refresh_token: string;
  }
}

/**
 * 注册
 */
export async function registerApi(data: { username: string; password: string }) {
  return requestClient.post('/api/user/register', data);
}

/**
 * 登录
 */
export async function loginApi(data: AuthApi.LoginParams) {
  return requestClient.post<AuthApi.LoginResult>('/api/auth/login', data);
}

/**
 * 刷新 accessToken
 */
export async function refreshTokenApi(data?: { refresh_token: string }) {
  return requestClient.post<AuthApi.RefreshTokenResult>('/api/auth/refresh', data);
}

/**
 * 退出登录
 */
export async function logoutApi() {
  return requestClient.post('/api/auth/logout');
}
