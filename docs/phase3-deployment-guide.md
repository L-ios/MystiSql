# MystiSql Phase 3: 安全控制层部署指南

> Phase 3 实现了企业级安全控制能力，> **适用版本**: MystiSql v0.3.0+

## 📋 目录

- [前置条件](#前置条件)
- [安装步骤](#安装步骤)
- [配置说明](#配置说明)
- [启用安全功能](#启用安全功能)
- [测试验证](#测试验证)
- [故障排查](#故障排查)
- [生产环境建议](#生产环境建议)

---

## 🔧 前置条件

### 系统要求

- **操作系统**: Linux / macOS / Windows
- **Go 版本**: 1.21+
- **Kubernetes**: 1.20+ (如使用 K8s 发现)
- **数据库**: MySQL 5.7+ / PostgreSQL 12+

### 依赖组件

- MySQL 驱动: `github.com/go-sql-driver/mysql` v1.7+
- PostgreSQL 驱动: `github.com/jackc/pgx/v5` v5.4+
- JWT 库: `github.com/golang-jwt/jwt/v5` v5.0+
- WebSocket: `github.com/gorilla/websocket` v1.5+

---

## 📦 安装步骤

### 1. 下载或编译

#### 方式一：从源码编译

```bash
# 克隆仓库
git clone https://github.com/your-org/MystiSql.git
cd MystiSql

# 编译
go build -o bin/mystisql ./cmd/mystisql

# 或编译 Linux 版本
GOOS=linux GOARCH=amd64 go build -o bin/mystisql-linux-amd64 ./cmd/mystisql
```

#### 方式二：下载预编译版本

```bash
# 下载最新版本
wget https://github.com/your-org/MystiSql/releases/download/v0.3.0/mystisql-linux-amd64
chmod +x mystisql-linux-amd64
sudo mv mystisql-linux-amd64 /usr/local/bin/mystisql
```

### 2. 创建配置文件

```bash
# 创建配置目录
sudo mkdir -p /etc/mystisql
sudo mkdir -p /var/log/mystisql

# 复制配置文件
sudo cp config/config.example.yaml /etc/mystisql/config.yaml

# 编辑配置文件
sudo vi /etc/mystisql/config.yaml
```

### 3. 设置环境变量

```bash
# 创建环境变量文件
cat << EOF | sudo tee /etc/mystisql/mystisql.env
# Token 密钥（生产环境必须更改！）
MYSTISQL_AUTH_TOKEN_SECRET=$(openssl rand -base64 32)

# 数据库密码
MYSQL_MASTER_PASSWORD=your-mysql-password
MYSQL_SLAVE_PASSWORD=your-mysql-password
POSTGRESQL_PASSWORD=your-postgresql-password
EOF

# 加载环境变量
source /etc/mystisql/mystisql.env
```

### 4. 创建 Systemd 服务

```bash
# 创建服务文件
cat << EOF | sudo tee /etc/systemd/system/mystisql.service
[Unit]
Description=MystiSql Database Gateway
After=network.target

[Service]
Type=simple
User=mystisql
Group=mystisql
EnvironmentFile=/etc/mystisql/mystisql.env
ExecStart=/usr/local/bin/mystisql serve --config /etc/mystisql/config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# 创建用户
sudo useradd -r -s /bin/false mystisql

# 设置权限
sudo chown -R mystisql:mystisql /etc/mystisql
sudo chown -R mystisql:mystisql /var/log/mystisql

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable mystisql
sudo systemctl start mystisql
```

---

## ⚙️ 配置说明

### Phase 3 新增配置项

#### 1. Token 认证配置

```yaml
auth:
  enabled: true                          # 是否启用认证
  token:
    secret: "${MYSTISQL_AUTH_TOKEN_SECRET}"  # JWT签名密钥（推荐使用环境变量）
    expire: "24h"                          # Token 过期时间
  whitelist:                               # 无需认证的路径
    - "/health"
    - "/metrics"
    - "/api/v1/auth/login"
```

**重要**:
- 生产环境**必须**更改 `secret`，不要使用默认值！
- 推荐使用强随机密钥：`openssl rand -base64 32`
- Token 过期时间根据安全要求调整（建议 1h-24h）

#### 2. 审计日志配置

```yaml
audit:
  enabled: true                           # 是否启用审计日志
  logFile: "/var/log/mystisql/audit.log" # 日志文件路径
  retentionDays: 30                       # 日志保留天数
```

**说明**:
- 审计日志按天轮转，文件名格式：`audit.log.YYYY-MM-DD`
- 自动删除超过 `retentionDays` 的日志文件
- 日志格式：JSON Lines，便于 ELK/Splunk 分析

#### 3. SQL 验证器配置

```yaml
validator:
  enabled: true                           # 是否启用 SQL 验证
  dangerousOperations:                    # 危险操作列表
    - "DROP"
    - "TRUNCATE"
    - "DELETE_WITHOUT_WHERE"
    - "UPDATE_WITHOUT_WHERE"
  whitelist: []                           # SQL 白名单（正则）
  blacklist: []                           # SQL 黑名单（正则）
```

**配置示例**:

```yaml
  # 允许特定的 DROP 操作
  whitelist:
    - "^DROP TABLE IF EXISTS temp_.*"
    - "^TRUNCATE TABLE cache_.*"
  
  # 禁止访问敏感表
  blacklist:
    - ".*audit_log.*"
    - ".*user_password.*"
```

#### 4. WebSocket 配置

```yaml
websocket:
  enabled: true                           # 是否启用 WebSocket
  maxConnections: 1000                    # 最大连接数
  idleTimeout: "10m"                      # 空闲超时时间
  maxConcurrentQueries: 5                 # 单连接最大并发查询数
```

#### 5. PostgreSQL 特有配置

```yaml
instances:
  - name: "production-postgresql"
    type: "postgresql"
    host: "postgresql.production.svc.cluster.local"
    port: 5432
    username: "mystisql_user"
    password: "${POSTGRESQL_PASSWORD}"
    database: "production_db"
    # PostgreSQL 特有配置
    sslmode: "require"           # SSL 模式: disable, require, verify-ca, verify-full
    connectTimeout: "10s"       # 连接超时
```

---

## 🔐 启用安全功能

### 渐进式启用（推荐）

#### 步骤 1: 仅启用审计日志（无破坏性）

```yaml
auth:
  enabled: false              # 暂时禁用认证

audit:
  enabled: true               # 启用审计日志
  logFile: "/var/log/mystisql/audit.log"
  retentionDays: 30

validator:
  enabled: false              # 暂时禁用 SQL 验证
```

**验证**:
```bash
# 执行一些 SQL 查询
mystisql query --instance production-mysql "SELECT * FROM users LIMIT 10"

# 检查审计日志
tail -f /var/log/mystisql/audit.log
```

#### 步骤 2: 启用 SQL 验证（观察模式）

```yaml
validator:
  enabled: true
  dangerousOperations:
    - "DROP"
    - "TRUNCATE"
  whitelist: []
  blacklist: []
```

**验证**:
```bash
# 尝试执行危险操作（应该被拦截）
mystisql query --instance production-mysql "DROP TABLE test_table"

# 检查审计日志中的拦截记录
grep "BLOCKED" /var/log/mystisql/audit.log
```

#### 步骤 3: 生成用户 Token

```bash
# 使用管理员账户生成 Token
curl -X POST http://localhost:8080/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{"user_id": "dev-user-001", "role": "developer"}'

# 返回示例:
# {
#   "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#   "expires_at": "2026-03-08T10:00:00Z"
# }

# 保存 Token
export MYSTISQL_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### 步骤 4: 启用全局认证

```yaml
auth:
  enabled: true              # 正式启用认证
  token:
    secret: "${MYSTISQL_AUTH_TOKEN_SECRET}"
    expire: "24h"
  whitelist:
    - "/health"
    - "/api/v1/auth/login"
```

**验证**:
```bash
# 无 Token 访问（应该返回 401）
curl http://localhost:8080/api/v1/instances

# 使用 Token 访问（应该成功）
curl -H "Authorization: Bearer $MYSTISQL_TOKEN" \
  http://localhost:8080/api/v1/instances

# CLI 使用 Token
mystisql instances list --token $MYSTISQL_TOKEN
```

---

## ✅ 测试验证

### 1. 功能测试

#### 测试 Token 认证

```bash
# 生成 Token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test-user", "role": "admin"}' | jq -r '.token')

# 验证 Token
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/auth/token/info

# 撤销 Token
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/auth/token
```

#### 测试审计日志

```bash
# 执行查询（会记录到审计日志）
curl -X POST http://localhost:8080/api/v1/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"instance": "production-mysql", "query": "SELECT * FROM users LIMIT 10"}'

# 查询审计日志
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/audit/logs?start=2026-03-08&end=2026-03-08"
```

#### 测试 SQL 验证

```bash
# 尝试执行危险操作（应该被拦截）
curl -X POST http://localhost:8080/api/v1/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"instance": "production-mysql", "query": "DROP TABLE users"}'

# 预期响应: 403 Forbidden
```

#### 测试 PostgreSQL 连接

```bash
# 查询 PostgreSQL 实例
curl -X POST http://localhost:8080/api/v1/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"instance": "production-postgresql", "query": "SELECT version()"}'
```

#### 测试 WebSocket

```javascript
// 使用 JavaScript 测试
const ws = new WebSocket('ws://localhost:8080/ws?token=' + TOKEN);

ws.onopen = () => {
  console.log('Connected');
  
  // 发送查询
  ws.send(JSON.stringify({
    action: 'query',
    instance: 'production-mysql',
    query: 'SELECT * FROM users LIMIT 5'
  }));
};

ws.onmessage = (event) => {
  console.log('Received:', JSON.parse(event.data));
};
```

### 2. 性能测试

```bash
# 认证中间件延迟测试（应 < 1ms）
ab -n 1000 -c 10 \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/instances

# 审计日志写入测试（应不阻塞）
ab -n 100 -c 10 \
  -H "Authorization: Bearer $TOKEN" \
  -p instance=test-mysql query='SELECT 1' \
  http://localhost:8080/api/v1/query
```

### 3. 安全测试

```bash
# 测试无 Token 访问（应返回 401）
curl http://localhost:8080/api/v1/instances

# 测试过期 Token（应返回 401）
curl -H "Authorization: Bearer expired-token" \
  http://localhost:8080/api/v1/instances

# 测试 SQL 注入（应被拦截）
curl -X POST http://localhost:8080/api/v1/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"instance": "production-mysql", "query": "SELECT * FROM users WHERE id = 1 OR 1=1"}'
```

---

## 🔧 故障排查

### 常见问题

#### 1. Token 认证失败

**症状**: 401 Unauthorized

**检查**:
```bash
# 检查 Token 是否有效
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/auth/token/info

# 检查密钥配置
grep "secret" /etc/mystisql/config.yaml

# 检查服务日志
journalctl -u mystisql -f
```

**解决**:
- 确保 `auth.token.secret` 配置正确
- 检查 Token 是否过期
- 确认请求路径不在白名单中

#### 2. 审计日志未记录

**症状**: 审计日志文件为空

**检查**:
```bash
# 检查配置
grep -A 5 "audit:" /etc/mystisql/config.yaml

# 检查目录权限
ls -ld /var/log/mystisql

# 检查服务状态
systemctl status mystisql
```

**解决**:
- 确认 `audit.enabled: true`
- 检查目录权限：`sudo chown -R mystisql:mystisql /var/log/mystisql`
- 重启服务：`sudo systemctl restart mystisql`

#### 3. SQL 验证误判

**症状**: 合法 SQL 被拦截

**解决**:
```yaml
# 添加到白名单
validator:
  whitelist:
    - "^your-allowed-sql-pattern.*"
```

#### 4. PostgreSQL 连接失败

**症状**: 连接超时或认证失败

**检查**:
```bash
# 测试网络连通性
telnet postgresql.production.svc.cluster.local 5432

# 测试认证
psql -h postgresql.production.svc.cluster.local \
  -U mystisql_user -d production_db
```

**解决**:
- 检查 `sslmode` 配置
- 确认用户名密码正确
- 检查网络策略/防火墙

#### 5. WebSocket 连接失败

**症状**: WebSocket 握手失败

**检查**:
```bash
# 检查 WebSocket 配置
grep -A 5 "websocket:" /etc/mystisql/config.yaml

# 测试 WebSocket 连接
wscat -c "ws://localhost:8080/ws?token=$TOKEN"
```

**解决**:
- 确认 `websocket.enabled: true`
- 检查 Token 是否有效
- 确认 URL 包含 Token 参数

---

## 🏭 生产环境建议

### 1. 安全加固

#### Token 密钥管理

```bash
# 生成强随机密钥
SECRET=$(openssl rand -base64 32)

# 使用 Kubernetes Secret
kubectl create secret generic mystisql-secret \
  --from-literal=token-secret=$SECRET

# 或使用 Vault 等 Secret 管理工具
```

#### 网络安全

```yaml
# 限制访问来源（config.yaml）
server:
  host: "0.0.0.0"
  port: 8080
  
# 配置防火墙规则
# iptables -A INPUT -p tcp --dport 8080 -s 10.0.0.0/8 -j ACCEPT
# iptables -A INPUT -p tcp --dport 8080 -j DROP
```

### 2. 高可用部署

#### 多实例部署

```yaml
# docker-compose.yml
version: '3'
services:
  mystisql:
    image: mystisql:latest
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1'
          memory: 512M
    environment:
      - MYSTISQL_AUTH_TOKEN_SECRET=${TOKEN_SECRET}
    volumes:
      - ./config.yaml:/etc/mystisql/config.yaml
      - ./logs:/var/log/mystisql
```

#### 负载均衡

```nginx
# nginx.conf
upstream mystisql_backend {
    server mystisql-1:8080;
    server mystisql-2:8080;
    server mystisql-3:8080;
}

server {
    listen 80;
    server_name mystisql.example.com;
    
    location / {
        proxy_pass http://mystisql_backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location /ws {
        proxy_pass http://mystisql_backend;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

### 3. 监控和告警

#### Prometheus 指标

```yaml
# config.yaml
metrics:
  enabled: true
  path: "/metrics"
```

#### 日志聚合

```yaml
# Filebeat 配置
filebeat.inputs:
- type: log
  paths:
    - /var/log/mystisql/audit.log
  fields:
    app: mystisql
    type: audit
  fields_under_root: true

output.elasticsearch:
  hosts: ["http://elasticsearch:9200"]
  index: "mystisql-audit-%{+yyyy.MM.dd}"
```

#### 告警规则

```yaml
# Prometheus 告警规则
groups:
- name: mystisql
  rules:
  - alert: MystiSqlAuthFailures
    expr: rate(mystisql_auth_failures_total[5m]) > 10
    for: 5m
    annotations:
      summary: "High authentication failure rate"
      
  - alert: MystiSqlSQLBlocked
    expr: rate(mystisql_sql_blocked_total[5m]) > 5
    for: 5m
    annotations:
      summary: "High SQL blocking rate"
```

### 4. 备份和恢复

#### 审计日志备份

```bash
# 定时备份审计日志
cat << EOF | sudo tee /etc/cron.d/mystisql-audit-backup
0 2 * * * mystisql /usr/local/bin/mystisql-audit-backup.sh
EOF

# 备份脚本
cat << 'EOF' | sudo tee /usr/local/bin/mystisql-audit-backup.sh
#!/bin/bash
DATE=$(date -d "yesterday" +%Y-%m-%d)
LOG_FILE="/var/log/mystisql/audit.log.$DATE"
BACKUP_DIR="/var/backups/mystisql/audit"

mkdir -p $BACKUP_DIR
if [ -f "$LOG_FILE" ]; then
    gzip -c $LOG_FILE > $BACKUP_DIR/audit.log.$DATE.gz
    # 上传到 S3/OSS 等
    # aws s3 cp $BACKUP_DIR/audit.log.$DATE.gz s3://bucket/audit/
fi
EOF

chmod +x /usr/local/bin/mystisql-audit-backup.sh
```

---

## 📚 相关文档

- [配置说明](/config/config.example.yaml)
- [API 文档](/docs/api.md)
- [安全最佳实践](/docs/security.md)
- [升级指南](/docs/upgrade.md)

---

## 💬 获取帮助

- **文档**: https://docs.mystisql.io
- **Issues**: https://github.com/your-org/MystiSql/issues
- **社区**: https://slack.mystisql.io

---

**最后更新**: 2026-03-08  
**版本**: v0.3.0
