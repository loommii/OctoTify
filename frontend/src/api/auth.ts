import { request } from '@/lib/request'
import type {
  LoginReq,
  RegisterReq,
  RefreshReq,
  ChangePasswordReq,
  AuthResp,
  UserProfileResp,
} from '@/types/api'

export const login = (data: LoginReq) => {
  return request<AuthResp>({
    url: '/auth/login',
    method: 'post',
    data,
  })
}

export const register = (data: RegisterReq) => {
  return request<AuthResp>({
    url: '/user/register',
    method: 'post',
    data,
  })
}

export const refreshToken = (data: RefreshReq) => {
  return request<AuthResp>({
    url: '/auth/refresh',
    method: 'post',
    data,
  })
}

export const logout = () => {
  return request<null>({
    url: '/auth/logout',
    method: 'post',
  })
}

export const getUserProfile = () => {
  return request<UserProfileResp>({
    url: '/user/profile',
    method: 'get',
  })
}

export const changePassword = (data: ChangePasswordReq) => {
  return request<null>({
    url: '/user/password',
    method: 'put',
    data,
  })
}
