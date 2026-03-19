export interface AuthUser {
  id: number
  email: string
  avatar_url?: string
  role: string
  credits: number
  created_at?: string
  updated_at?: string
}

export interface LoginRequest {
  email: string
  password: string
  captcha_id: string
  captcha_code: string
}

export interface RegisterRequest {
  email: string
  password: string
  captcha_id: string
  captcha_code: string
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface UpdateProfileRequest {
  avatar_url?: string
}

export interface AuthResponse {
  token: string
  user: AuthUser
}

export interface CaptchaResponse {
  captcha_id: string
  image_data: string
}
