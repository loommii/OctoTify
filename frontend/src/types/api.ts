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
  data: T
}

export interface PageResult<T = unknown> {
  list: T[]
  total: number
  page: number
  page_size: number
}
