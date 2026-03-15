export interface ErrorResponse {
  success: false
  error: {
    code: string
    message: string
  }
}

export interface InstanceResponse {
  name: string
  type: string
  host: string
  port: number
  database?: string
  username?: string
  status: string
  labels?: Record<string, string>
}

export interface InstancesListResponse {
  total: number
  instances: InstanceResponse[]
}

export interface QueryRequest {
  instance: string
  sql: string
  timeout?: number
  transaction_id?: string
}

export interface ColumnInfo {
  name: string
  type: string
  nullable: boolean
}

export interface QueryResultData {
  columns: ColumnInfo[]
  rows: Record<string, unknown>[]
  rowCount: number
}

export interface QueryResponse {
  success: boolean
  data?: QueryResultData
  executionTime: number
  error?: {
    code: string
    message: string
  }
}

export interface ExecRequest {
  instance: string
  sql: string
  timeout?: number
  transaction_id?: string
}

export interface ExecResultData {
  affectedRows: number
  lastInsertId: number
}

export interface ExecResponse {
  success: boolean
  data?: ExecResultData
  executionTime: number
  error?: {
    code: string
    message: string
  }
}

export interface GenerateTokenRequest {
  user_id: string
  role: string
}

export interface TokenData {
  token: string
  tokenId: string
  expiresAt: string
  issuedAt: string
  userId: string
  role: string
}

export interface GenerateTokenResponse {
  success: boolean
  data?: TokenData
  error?: {
    code: string
    message: string
  }
}

export interface TokenInfoResponse {
  success: boolean
  userId?: string
  role?: string
  tokenId?: string
  expiresAt?: string
  issuedAt?: string
  error?: {
    code: string
    message: string
  }
}

export interface InstanceHealthResponse {
  instance: string
  status: string
  timestamp: string
}

export interface PoolStats {
  maxOpenConnections: number
  openConnections: number
  inUse: number
  idle: number
  waitCount: number
  waitDuration: number
  maxIdleClosed: number
  maxLifetimeClosed: number
}

export interface PoolStatsResponse {
  instance: string
  stats: PoolStats
  timestamp: string
}

export interface AuditLog {
  id: string
  timestamp: string
  userId: string
  instance: string
  sql: string
  executionTime: number
  rowsAffected: number
  success: boolean
  errorMessage?: string
}

export interface AuditLogsResponse {
  success: boolean
  data?: {
    total: number
    logs: AuditLog[]
  }
  error?: {
    code: string
    message: string
  }
}
