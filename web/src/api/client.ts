import axios, { AxiosInstance, AxiosError } from 'axios'
import { useAuthStore } from '../stores/authStore'
import type {
  InstancesListResponse,
  QueryRequest,
  QueryResponse,
  ExecRequest,
  ExecResponse,
  TokenInfoResponse,
  InstanceHealthResponse,
  PoolStatsResponse,
  AuditLogsResponse,
  AuditStatsResponse,
} from './types'

const API_BASE_URL = '/api/v1'

class ApiClient {
  private client: AxiosInstance

  constructor() {
    this.client = axios.create({
      baseURL: API_BASE_URL,
      timeout: 60000,
      headers: {
        'Content-Type': 'application/json',
      },
    })

    this.client.interceptors.request.use(
      (config) => {
        const token = useAuthStore.getState().token
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => Promise.reject(error)
    )

    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => {
        if (error.response?.status === 401) {
          useAuthStore.getState().clearAuth()
          window.location.href = '/login'
        }
        return Promise.reject(error)
      }
    )
  }

  async getTokenInfo(token: string): Promise<TokenInfoResponse> {
    const response = await this.client.post<TokenInfoResponse>(
      '/auth/token/info',
      { token }
    )
    return response.data
  }

  async verifyAndLogin(token: string): Promise<{
    success: boolean
    userId?: string
    role?: string
    expiresAt?: string
    error?: string
  }> {
    try {
      const response = await this.client.post<TokenInfoResponse>(
        '/auth/token/info',
        { token }
      )
      if (response.data.success) {
        return {
          success: true,
          userId: response.data.userId,
          role: response.data.role,
          expiresAt: response.data.expiresAt,
        }
      }
      return {
        success: false,
        error: response.data.error?.message || 'Token 验证失败',
      }
    } catch (error: unknown) {
      const axiosError = error as AxiosError<{ error?: { message?: string } }>
      const msg =
        axiosError.response?.data?.error?.message || 'Token 验证失败'
      return { success: false, error: msg }
    }
  }

  async getInstances(): Promise<InstancesListResponse> {
    const response = await this.client.get<InstancesListResponse>('/instances')
    return response.data
  }

  async getInstanceHealth(name: string): Promise<InstanceHealthResponse> {
    const response = await this.client.get<InstanceHealthResponse>(
      `/instances/${encodeURIComponent(name)}/health`
    )
    return response.data
  }

  async getPoolStats(name: string): Promise<PoolStatsResponse> {
    const response = await this.client.get<PoolStatsResponse>(
      `/instances/${encodeURIComponent(name)}/pool`
    )
    return response.data
  }

  async query(request: QueryRequest): Promise<QueryResponse> {
    const response = await this.client.post<QueryResponse>('/query', request)
    return response.data
  }

  async exec(request: ExecRequest): Promise<ExecResponse> {
    const response = await this.client.post<ExecResponse>('/exec', request)
    return response.data
  }

  async getAuditLogs(params: {
    userId?: string
    instance?: string
    startTime?: string
    endTime?: string
    page?: number
    pageSize?: number
    sensitive?: boolean
    queryType?: string
  }): Promise<AuditLogsResponse> {
    const response = await this.client.get<AuditLogsResponse>('/audit/logs', {
      params,
    })
    return response.data
  }

  async getAuditStats(params?: {
    startTime?: string
    endTime?: string
  }): Promise<AuditStatsResponse> {
    const response = await this.client.get<AuditStatsResponse>('/audit/stats', {
      params,
    })
    return response.data
  }
}

export const apiClient = new ApiClient()
