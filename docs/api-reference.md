# MystiSql API 参考文档

## 概述

MystiSql 提供 RESTful API 用于数据库访问和管理。所有 API 端点（除健康检查外）都需要 JWT Token 认证。

**基础 URL**: `http://localhost:8080/api/v1`

**认证方式**: Bearer Token
```
Authorization: Bearer <jwt_token>
```

---

## 健康检查

### 检查服务状态

检查服务健康状态，无需认证。

**请求**
```http
GET /health
```

**请求示例：**
```bash
curl http://localhost:8080/health
```

**响应**
```json
{
  "status": "healthy",
  "version": "0.1.0",
  "timestamp": "2026-03-06T10:00:00Z"
}
```

**字段说明：**

| 字段 | 类型 | 说明 |
|-----|------|------|
| status | string | 服务状态：healthy（健康）、unhealthy（不健康） |
| version | string | MystiSql 版本号 |
| timestamp | string | 响应时间戳（ISO 8601 格式） |

---

## 认证 API

### 生成 Token

生成新的 JWT Token 用于 API 认证。

**请求**
```http
POST /api/v1/auth/token
Content-Type: application/json

{
  "user_id": "admin",
  "role": "admin"
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "tokenId": "token-abc123",
    "expiresAt": "2026-03-10T10:00:00Z",
    "issuedAt": "2026-03-09T10:00:00Z",
    "userId": "admin",
    "role": "admin"
  }
}
```

**错误响应**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "user_id and role are required"
  }
}
```

---

### 撤销 Token

撤销指定的 JWT Token，使其立即失效。

**请求**
```http
DELETE /api/v1/auth/token
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**响应**
```json
{
  "success": true,
  "message": "Token revoked successfully"
}
```

---

### 查询 Token 列表

查询已撤销的 Token 列表。

**请求**
```http
GET /api/v1/auth/tokens
Authorization: Bearer <jwt_token>
```

**响应**
```json
{
  "success": true,
  "revokedTokens": [
    {
      "token": "eyJh...（脱敏）",
      "reason": "User logout",
      "revokedAt": "2026-03-09T10:30:00Z"
    }
  ]
}
```

---

### 查看 Token 信息

查看当前 Token 的详细信息。

**请求**
```http
GET /api/v1/auth/token/info?token=<jwt_token>
Authorization: Bearer <jwt_token>
```

**响应**
```json
{
  "success": true,
  "userId": "admin",
  "role": "admin",
  "tokenId": "token-abc123",
  "expiresAt": "2026-03-10T10:00:00Z",
  "issuedAt": "2026-03-09T10:00:00Z"
}
```

---

## 审计日志 API

### 查询审计日志

查询 SQL 执行审计日志。

**请求**
```http
GET /api/v1/audit/logs?start_time=2026-03-01&end_time=2026-03-09&user_id=admin&page=1&page_size=20
Authorization: Bearer <jwt_token>
```

**查询参数**
- `start_time`: 开始时间（可选）
- `end_time`: 结束时间（可选）
- `user_id`: 用户 ID（可选）
- `instance`: 实例名称（可选）
- `page`: 页码（默认 1）
- `page_size`: 每页大小（默认 20）

**响应**
```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "id": "log-123",
        "timestamp": "2026-03-09T10:00:00Z",
        "userId": "admin",
        "instance": "local-mysql",
        "queryType": "SELECT",
        "query": "SELECT * FROM users LIMIT 10",
        "rowsAffected": 10,
        "executionTime": 50,
        "clientIp": "192.168.1.100",
        "success": true
      }
    ],
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```

---

## SQL 验证器 API

### 更新 SQL 白名单

更新允许执行的 SQL 模式白名单。

**请求**
```http
PUT /api/v1/validator/whitelist
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "patterns": [
    "SELECT * FROM system_config",
    "SELECT id, name FROM users WHERE status = ?"
  ]
}
```

**响应**
```json
{
  "success": true,
  "message": "Whitelist updated successfully",
  "count": 2
}
```

---

### 更新 SQL 黑名单

更新禁止执行的 SQL 模式黑名单。

**请求**
```http
PUT /api/v1/validator/blacklist
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "patterns": [
    "DELETE FROM audit_log",
    "DROP TABLE .*"
  ]
}
```

**响应**
```json
{
  "success": true,
  "message": "Blacklist updated successfully",
  "count": 2
}
```

---

## 事务管理 API

### 开始事务

开始一个新的事务。

**请求**
```http
POST /api/v1/transaction/begin
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "instance": "local-mysql",
  "isolation_level": "READ_COMMITTED"
}
```

**隔离级别**
- `DEFAULT`: 使用数据库默认隔离级别
- `READ_UNCOMMITTED`: 读未提交
- `READ_COMMITTED`: 读已提交
- `REPEATABLE_READ`: 可重复读
- `SERIALIZABLE`: 序列化

**响应**
```json
{
  "transaction_id": "tx-abc123-def456",
  "connection_id": "conn-xyz789",
  "instance": "local-mysql",
  "created_at": "2026-03-09T10:00:00Z",
  "expires_at": "2026-03-09T10:05:00Z"
}
```

---

### 在事务中执行查询

在事务中执行 SQL 查询。

**请求**
```http
POST /api/v1/query
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "instance": "local-mysql",
  "sql": "SELECT * FROM users WHERE id = 1",
  "transaction_id": "tx-abc123-def456"
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "columns": [
      {"name": "id", "type": "INT"},
      {"name": "name", "type": "VARCHAR"}
    ],
    "rows": [
      [1, "Alice"]
    ],
    "rowCount": 1
  },
  "executionTime": 5
}
```

---

### 在事务中执行更新

在事务中执行 SQL 更新语句。

**请求**
```http
POST /api/v1/exec
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "instance": "local-mysql",
  "sql": "UPDATE users SET name = 'Bob' WHERE id = 1",
  "transaction_id": "tx-abc123-def456"
}
```

**响应**
```json
{
  "success": true,
  "data": {
    "affectedRows": 1,
    "lastInsertId": 0
  },
  "executionTime": 3
}
```

---

### 提交事务

提交事务并释放连接。

**请求**
```http
POST /api/v1/transaction/commit
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "transaction_id": "tx-abc123-def456"
}
```

**响应**
```json
{
  "message": "Transaction committed successfully",
  "transaction_id": "tx-abc123-def456"
}
```

---

### 回滚事务

回滚事务并释放连接。

**请求**
```http
POST /api/v1/transaction/rollback
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "transaction_id": "tx-abc123-def456"
}
```

**响应**
```json
{
  "message": "Transaction rolled back successfully",
  "transaction_id": "tx-abc123-def456"
}
```

---

### 查询事务状态

查询指定事务的详细信息。

**请求**
```http
GET /api/v1/transaction/tx-abc123-def456
Authorization: Bearer <jwt_token>
```

**响应**
```json
{
  "transaction_id": "tx-abc123-def456",
  "connection_id": "conn-xyz789",
  "instance": "local-mysql",
  "state": "active",
  "isolation_level": "READ_COMMITTED",
  "created_at": "2026-03-09T10:00:00Z",
  "expires_at": "2026-03-09T10:05:00Z",
  "last_activity_at": "2026-03-09T10:02:30Z",
  "user_id": "admin"
}
```

---

### 列出所有事务

列出当前所有活动的事务。

**请求**
```http
GET /api/v1/transaction
Authorization: Bearer <jwt_token>
```

**响应**
```json
{
  "transactions": [
    {
      "transaction_id": "tx-abc123",
      "connection_id": "conn-xyz",
      "instance": "local-mysql",
      "state": "active",
      "created_at": "2026-03-09T10:00:00Z"
    }
  ],
  "count": 1
}
```

---

### 延长事务

延长事务的超时时间。

**请求**
```http
POST /api/v1/transaction/tx-abc123-def456/extend
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "duration": "10m"
}
```

**响应**
```json
{
  "message": "Transaction extended successfully",
  "transaction_id": "tx-abc123-def456",
  "new_expires_at": "2026-03-09T10:12:00Z"
}
```

---

## 批量操作 API

### 批量执行 SQL

批量执行多个 SQL 语句。

**请求**
```http
POST /api/v1/batch
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "instance": "local-mysql",
  "queries": [
    "INSERT INTO users (name) VALUES ('Alice')",
    "INSERT INTO users (name) VALUES ('Bob')",
    "UPDATE stats SET count = count + 2"
  ],
  "stopOnError": false,
  "useTransaction": true
}
```

**参数**
- `instance`: 数据库实例名称（必填）
- `queries`: SQL 语句数组（必填）
- `transactionId`: 事务 ID（可选，用于在已有事务中执行）
- `stopOnError`: 遇到错误是否停止（默认 false）
- `useTransaction`: 是否在新事务中执行（默认 false）

**响应**
```json
{
  "message": "Batch execution completed successfully",
  "result": {
    "results": [
      {
        "index": 0,
        "sql": "INSERT INTO users (name) VALUES ('Alice')",
        "rowsAffected": 1,
        "lastInsertId": 1,
        "success": true,
        "executionTimeMs": 2
      },
      {
        "index": 1,
        "sql": "INSERT INTO users (name) VALUES ('Bob')",
        "rowsAffected": 1,
        "lastInsertId": 2,
        "success": true,
        "executionTimeMs": 1
      }
    ],
    "totalRowsAffected": 2,
    "successCount": 2,
    "failureCount": 0,
    "totalExecutionTimeMs": 5
  }
}
```

**错误响应（部分失败）**
```json
{
  "message": "Batch execution completed with errors",
  "result": {
    "results": [
      {
        "index": 0,
        "sql": "INSERT INTO users (name) VALUES ('Alice')",
        "rowsAffected": 1,
        "success": true
      },
      {
        "index": 1,
        "sql": "INSERT INTO users (name) VALUES ('Bob')",
        "success": false,
        "error": "duplicate key violation"
      }
    ],
    "successCount": 1,
    "failureCount": 1
  },
  "error": "batch execution failed with 1 errors"
}
```

---

## 数据库查询 API

### 执行查询

执行 SQL 查询（SELECT、SHOW 等）。

**请求**
```http
POST /api/v1/query
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "instance": "local-mysql",
  "sql": "SELECT * FROM users LIMIT 10",
  "timeout": 30
}
```

**参数**
- `instance`: 数据库实例名称（必填）
- `sql`: SQL 查询语句（必填）
- `timeout`: 超时时间（秒，可选）
- `transaction_id`: 事务 ID（可选）

**响应**
```json
{
  "success": true,
  "executionTime": 5,
  "data": {
    "columns": [
      {"name": "id", "type": "INT"},
      {"name": "name", "type": "VARCHAR"},
      {"name": "email", "type": "VARCHAR"}
    ],
    "rows": [
      [1, "Alice", "alice@example.com"],
      [2, "Bob", "bob@example.com"]
    ],
    "rowCount": 2
  }
}
```

---

### 执行更新

执行 SQL 更新语句（INSERT、UPDATE、DELETE）。

**请求**
```http
POST /api/v1/exec
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "instance": "local-mysql",
  "sql": "INSERT INTO users (name, email) VALUES ('Charlie', 'charlie@example.com')",
  "timeout": 30
}
```

**响应**
```json
{
  "success": true,
  "executionTime": 3,
  "data": {
    "affectedRows": 1,
    "lastInsertId": 3
  }
}
```

---

## 实例管理 API

### 列出实例

列出所有可用的数据库实例。

**请求**
```http
GET /api/v1/instances
Authorization: Bearer <jwt_token>
```

**响应**
```json
{
  "total": 2,
  "instances": [
    {
      "name": "local-mysql",
      "type": "mysql",
      "host": "localhost",
      "port": 3306,
      "database": "testdb",
      "username": "root",
      "status": "healthy",
      "labels": {
        "env": "dev"
      }
    },
    {
      "name": "local-postgres",
      "type": "postgresql",
      "host": "localhost",
      "port": 5432,
      "database": "testdb",
      "username": "postgres",
      "status": "healthy"
    }
  ]
}
```

---

### 检查实例健康状态

检查指定实例的健康状态。

**请求**
```http
GET /api/v1/instances/local-mysql/health
Authorization: Bearer <jwt_token>
```

**响应**
```json
{
  "instance": "local-mysql",
  "status": "healthy",
  "timestamp": "2026-03-09T10:00:00Z"
}
```

---

### 查询连接池状态

查询指定实例的连接池统计信息。

**请求**
```http
GET /api/v1/instances/local-mysql/pool
Authorization: Bearer <jwt_token>
```

**响应**
```json
{
  "instance": "local-mysql",
  "stats": {
    "totalConnections": 10,
    "activeConnections": 3,
    "idleConnections": 7,
    "maxConnections": 20,
    "waitCount": 0,
    "waitDuration": 0
  },
  "timestamp": "2026-03-09T10:00:00Z"
}
```

---

## WebSocket API

### WebSocket 连接

通过 WebSocket 进行实时查询。

**连接 URL**
```
ws://localhost:8080/ws?token=<jwt_token>
```

**消息格式**

发送查询：
```json
{
  "action": "query",
  "instance": "local-mysql",
  "query": "SELECT * FROM users LIMIT 10",
  "requestId": "req-123"
}
```

响应：
```json
{
  "action": "query",
  "requestId": "req-123",
  "success": true,
  "data": {
    "columns": [...],
    "rows": [...],
    "rowCount": 10
  },
  "executionTime": 5
}
```

**心跳**

客户端发送：
```json
{
  "action": "ping"
}
```

服务端响应：
```json
{
  "action": "pong",
  "timestamp": "2026-03-09T10:00:00Z"
}
```

**错误响应**
```json
{
  "action": "query",
  "requestId": "req-123",
  "success": false,
  "error": {
    "code": "QUERY_FAILED",
    "message": "SQL syntax error"
  }
}
```

---

## CORS 支持

API 支持 CORS（跨域资源共享），允许从 Web 前端调用：

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

---

## 错误响应

当请求失败时，API 返回统一错误格式：

**响应格式：**
```json
{
  "error": "错误描述",
  "code": "ERROR_CODE",
  "details": "详细错误信息"
}
```

### 通用错误

| HTTP 状态码 | 错误码 | 说明 |
|-----------|--------|------|
| 400 | INVALID_REQUEST | 请求参数无效 |
| 401 | UNAUTHORIZED | 未认证或 Token 无效 |
| 403 | FORBIDDEN | 权限不足 |
| 404 | NOT_FOUND | 资源不存在 |
| 404 | INSTANCE_NOT_FOUND | 实例不存在 |
| 429 | TOO_MANY_REQUESTS | 请求频率超限 |
| 500 | INTERNAL_ERROR | 内部错误 |
| 503 | SERVICE_UNAVAILABLE | 服务不可用 |

### 认证错误

| 错误代码 | 说明 |
|---------|------|
| `INVALID_TOKEN` | Token 格式无效 |
| `TOKEN_EXPIRED` | Token 已过期 |
| `TOKEN_REVOKED` | Token 已被撤销 |
| `MISSING_TOKEN` | 缺少认证 Token |

### 查询错误

| 错误代码 | 说明 |
|---------|------|
| `QUERY_FAILED` | 查询执行失败 |
| `EXEC_FAILED` | 执行语句失败 |
| `INSTANCE_NOT_FOUND` | 数据库实例不存在 |
| `CONNECTION_FAILED` | 数据库连接失败 |
| `SQL_VALIDATION_FAILED` | SQL 验证失败 |

### 事务错误

| 错误代码 | 说明 |
|---------|------|
| `TRANSACTION_NOT_FOUND` | 事务不存在 |
| `TRANSACTION_EXPIRED` | 事务已过期 |
| `TRANSACTION_NOT_ACTIVE` | 事务非活动状态 |
| `MAX_TRANSACTIONS_REACHED` | 达到最大事务数 |

---

## 速率限制

API 请求受以下限制：

- **每秒请求数**: 100 requests/second
- **每分钟请求数**: 5000 requests/minute
- **并发连接数**: 1000 connections

超过限制将返回 `429 Too Many Requests`。

---

## SDK 示例

### Go SDK

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/go-resty/resty/v2"
)

type MystiSqlClient struct {
    baseURL string
    token   string
    client  *resty.Client
}

func NewMystiSqlClient(baseURL, token string) *MystiSqlClient {
    return &MystiSqlClient{
        baseURL: baseURL,
        token:   token,
        client:  resty.New(),
    }
}

func (c *MystiSqlClient) Query(ctx context.Context, instance, sql string) (*QueryResponse, error) {
    resp, err := c.client.R().
        SetContext(ctx).
        SetAuthToken(c.token).
        SetBody(map[string]interface{}{
            "instance": instance,
            "sql":      sql,
        }).
        SetResult(&QueryResponse{}).
        Post(c.baseURL + "/api/v1/query")
    
    if err != nil {
        return nil, err
    }
    
    return resp.Result().(*QueryResponse), nil
}

type QueryResponse struct {
    Success       bool         `json:"success"`
    ExecutionTime int64        `json:"executionTime"`
    Data          *QueryResult `json:"data"`
}

type QueryResult struct {
    Columns  []ColumnInfo `json:"columns"`
    Rows     [][]interface{} `json:"rows"`
    RowCount int          `json:"rowCount"`
}

func main() {
    client := NewMystiSqlClient("http://localhost:8080", "your-jwt-token")
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    result, err := client.Query(ctx, "local-mysql", "SELECT * FROM users LIMIT 10")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Query returned %d rows\n", result.Data.RowCount)
}
```

### Python SDK

```python
import requests

class MystiSqlClient:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.token = token
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
    
    def query(self, instance, sql, timeout=30):
        """执行 SQL 查询"""
        response = requests.post(
            f'{self.base_url}/api/v1/query',
            headers=self.headers,
            json={
                'instance': instance,
                'sql': sql,
                'timeout': timeout
            }
        )
        response.raise_for_status()
        return response.json()
    
    def begin_transaction(self, instance, isolation_level='DEFAULT'):
        """开始事务"""
        response = requests.post(
            f'{self.base_url}/api/v1/transaction/begin',
            headers=self.headers,
            json={
                'instance': instance,
                'isolation_level': isolation_level
            }
        )
        response.raise_for_status()
        return response.json()
    
    def commit_transaction(self, transaction_id):
        """提交事务"""
        response = requests.post(
            f'{self.base_url}/api/v1/transaction/commit',
            headers=self.headers,
            json={'transaction_id': transaction_id}
        )
        response.raise_for_status()
        return response.json()

# 使用示例
client = MystiSqlClient('http://localhost:8080', 'your-jwt-token')

# 执行查询
result = client.query('local-mysql', 'SELECT * FROM users LIMIT 10')
print(f"查询返回 {result['data']['rowCount']} 行")

# 使用事务
tx = client.begin_transaction('local-mysql')
print(f"事务 ID: {tx['transaction_id']}")

# 提交事务
client.commit_transaction(tx['transaction_id'])
```

---

## 最佳实践

### 1. Token 管理
- Token 有效期为 24 小时，建议在过期前刷新
- 使用环境变量存储 Token，避免硬编码
- 及时撤销不再使用的 Token

### 2. 事务使用
- 事务默认超时为 5 分钟，长时间操作请延长超时
- 事务完成后立即提交或回滚，释放连接
- 避免在事务中执行耗时操作

### 3. 批量操作
- 批量大小建议不超过 1000
- 对于大量数据，分批执行
- 使用事务保证原子性

### 4. 错误处理
- 始终检查响应中的 `success` 字段
- 根据错误代码进行重试或回退
- 记录错误日志便于排查

### 5. 性能优化
- 使用连接池减少连接开销
- 合理设置查询超时
- 对于频繁查询，考虑使用缓存

---

## 版本信息

- **API 版本**: v1
- **文档版本**: 1.0.0
- **最后更新**: 2026-03-09

---

## 支持

- **GitHub Issues**: https://github.com/your-org/mystisql/issues
- **文档**: https://mystisql.dev/docs
- **社区**: https://discord.gg/mystisql
