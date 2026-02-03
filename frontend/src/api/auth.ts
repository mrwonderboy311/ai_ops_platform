import { apiClient } from './client'
import type {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  RefreshTokenResponse,
  ApiErrorResponse,
} from '../types/auth'

class AuthAPI {
  /**
   * User login
   */
  async login(data: LoginRequest): Promise<LoginResponse> {
    const response = await apiClient.post<{ data: LoginResponse }>('/api/v1/auth/login', data)
    return response.data.data
  }

  /**
   * User registration
   */
  async register(data: RegisterRequest): Promise<RegisterResponse> {
    const response = await apiClient.post<{ data: RegisterResponse }>('/api/v1/auth/register', data)
    return response.data.data
  }

  /**
   * Refresh access token
   */
  async refreshToken(refreshToken: string): Promise<RefreshTokenResponse> {
    // Use direct axios call to avoid interceptor loop
    const response = await apiClient.post<{ data: RefreshTokenResponse }>(
      '/api/v1/auth/refresh',
      { refreshToken }
    )
    return response.data.data
  }

  /**
   * Extract error information from API error response
   */
  getErrorMessage(error: unknown): string {
    const axiosError = error as { response?: { data?: ApiErrorResponse } }
    if (axiosError.response?.data?.error) {
      return axiosError.response.data.error.message
    }
    return '操作失败，请稍后重试'
  }

  /**
   * Extract error code from API error response
   */
  getErrorCode(error: unknown): string {
    const axiosError = error as { response?: { data?: ApiErrorResponse } }
    if (axiosError.response?.data?.error) {
      return axiosError.response.data.error.code
    }
    return 'UNKNOWN_ERROR'
  }
}

export const authAPI = new AuthAPI()
