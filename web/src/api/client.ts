import axios, { AxiosInstance, AxiosError } from 'axios'
import { useAuthStore } from '../stores/authStore'
import type {
  InstancesListResponse,
  QueryRequest,
  QueryResponse,
  ExecRequest,
  ExecResponse,
  GenerateTokenResponse,
  TokenInfoResponse,
  InstanceHealthResponse,
  PoolStatsResponse,
  AuditLogsResponse,
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

  async login(userId: string, role: string): Promise<GenerateTokenResponse> {
    const response = await this.client.post<GenerateTokenResponse>(
      '/auth/token',
      { user_id: userId, role }
    )
    return response.data
  }

  async getTokenInfo(token: string): Promise<TokenInfoResponse> {
    const response = await this.client.get<TokenInfoResponse>(
      '/auth/token/info',
      { params: { token } }
    )
    return response.data
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
  }): Promise<AuditLogsResponse> {
    const response = await this.client.get<AuditLogsResponse>('/audit/logs', {
      params,
    })
    return response.data
  }
}

export const apiClient = new ApiClient()
