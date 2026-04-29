# 配置说明

## 配置文件结构

MystiSql 使用 YAML 格式的配置文件，支持以下配置项：

### server - 服务器配置

```yaml
server:
  host: 0.0.0.0        # API 服务器监听地址（默认：0.0.0.0）
  port: 8080           # API 服务器监听端口（默认：8080）
  mode: release        # 运行模式：debug（调试）或 release（生产）
```

**字段说明：**

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|-----|------|------|--------|------|
| host | string | 否 | 0.0.0.0 | API 服务器监听地址，0.0.0.0 表示监听所有网络接口 |
| port | int | 否 | 8080 | API 服务器监听端口，建议使用 1024 以上的端口 |
| mode | string | 否 | release | 运行模式，debug 会输出详细日志，release 为生产模式 |

### discovery - 服务发现配置

```yaml
discovery:
  type: static         # 发现类型：static、k8s、consul
```

**字段说明：**

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|-----|------|------|--------|------|
| type | string | 否 | static | 服务发现类型 |
|     |       |      |        | • static: 静态配置（从配置文件读取） |
|     |       |      |        | • k8s: Kubernetes API 动态发现（Phase 2） |
|     |       |      |        | • consul: Consul 服务发现（Phase 4） |

### instances - 数据库实例列表

```yaml
instances:
  - name: local-mysql              # 实例名称（唯一标识）
    type: mysql                    # 数据库类型
    host: localhost                # 主机地址
    port: 3306                     # 端口号
    username: root                 # 用户名
    password: root                 # 密码
    database: test                 # 数据库名
    labels:                        # 标签（可选）
      environment: development
      team: backend
```

**字段说明：**

| 字段 | 类型 | 必填 | 默认值 | 说明 |
|-----|------|------|--------|------|
| name | string | 是 | - | 实例名称，必须唯一，用于查询时指定实例 |
| type | string | 是 | - | 数据库类型，支持：mysql、postgresql、oracle、redis |
| host | string | 是 | - | 数据库主机地址，可以是 IP 或域名 |
| port | int | 是 | - | 数据库端口号 |
| username | string | 否 | - | 数据库用户名 |
| password | string | 否 | - | 数据库密码，建议使用环境变量 |
| database | string | 否 | - | 默认连接的数据库名 |
| labels | map | 否 | - | 实例标签，K8s 风格的键值对 |

**支持的数据库类型：**

- `mysql`: MySQL 数据库（Phase 1 支持）
- `postgresql`: PostgreSQL 数据库（Phase 2 支持）
- `oracle`: Oracle 数据库（Phase 2 支持）
- `redis`: Redis 数据库（Phase 3 支持）

## 环境变量覆盖

可以使用环境变量覆盖配置文件中的值：

```bash
# 服务器配置
export MYSTISQL_SERVER_HOST=0.0.0.0
export MYSTISQL_SERVER_PORT=8080
export MYSTISQL_SERVER_MODE=release

# 服务发现配置
export MYSTISQL_DISCOVERY_TYPE=static

# 实例配置（数组索引从 0 开始）
export MYSTISQL_INSTANCES_0_NAME=local-mysql
export MYSTISQL_INSTANCES_0_HOST=localhost
export MYSTISQL_INSTANCES_0_PORT=3306
export MYSTISQL_INSTANCES_0_USERNAME=root
export MYSTISQL_INSTANCES_0_PASSWORD=secret
```

## 配置示例

### 开发环境配置

```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: debug

discovery:
  type: static

instances:
  - name: dev-mysql
    type: mysql
    host: localhost
    port: 3306
    username: root
    password: root
    database: dev_db
    labels:
      environment: development
```

### 高可用配置示例

```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: release

# 高可用配置
ha:
  enabled: true
  readWriteSplitting:
    enabled: true
    readStrategy: round-robin  # round-robin, weighted, least-conn
    readAfterWrite: default
  healthCheck:
    enabled: true
    interval: 10s
    timeout: 5s
    failureThreshold: 3
    recoveryThreshold: 2
  failover:
    enabled: true
    autoFailover: false
    timeout: 30s
    maxDelay: 1s

# 多集群配置
multiCluster:
  enabled: false
  clusters:
    - name: cluster-1
      region: us-east-1
      instances:
        - name: mysql-master-us
          type: mysql
          host: mysql-master.us-east-1.example.com
          port: 3306
          role: master
        - name: mysql-slave-us-1
          type: mysql
          host: mysql-slave-1.us-east-1.example.com
          port: 3306
          role: slave
        - name: mysql-slave-us-2
          type: mysql
          host: mysql-slave-2.us-east-1.example.com
          port: 3306
          role: slave

instances:
  - name: production-mysql-master
    type: mysql
    host: mysql-master.production.svc.cluster.local
    port: 3306
    username: mystisql_user
    password: ${MYSQL_MASTER_PASSWORD}
    database: production_db
    role: master
    labels:
      environment: production
      role: master
      cluster: production
      region: us-east-1
      availabilityZone: us-east-1a
      version: "8.0"
      tier: primary
      owner: database-team
      purpose: production
      service: mystisql
      managedBy: kubernetes
      lifecycle: production
      backup: daily
      monitoring: enabled
      alerting: enabled
      securityGroup: sg-123456
      subnet: subnet-123456
      vpc: vpc-123456
      account: 123456789012
      region: us-east-1
      cloud: aws

  - name: production-mysql-slave-1
    type: mysql
    host: mysql-slave-1.production.svc.cluster.local
    port: 3306
    username: mystisql_user
    password: ${MYSQL_SLAVE_PASSWORD}
    database: production_db
    role: slave
    labels:
      environment: production
      role: slave
      cluster: production
      region: us-east-1
      availabilityZone: us-east-1b
      version: "8.0"
      tier: primary
      owner: database-team
      purpose: production
      service: mystisql
      managedBy: kubernetes
      lifecycle: production
      backup: daily
      monitoring: enabled
      alerting: enabled
      securityGroup: sg-123456
      subnet: subnet-123456
      vpc: vpc-123456
      account: 123456789012
      region: us-east-1
      cloud: aws

  - name: production-mysql-slave-2
    type: mysql
    host: mysql-slave-2.production.svc.cluster.local
    port: 3306
    username: mystisql_user
    password: ${MYSQL_SLAVE_PASSWORD}
    database: production_db
    role: slave
    labels:
      environment: production
      role: slave
      cluster: production
      region: us-east-1
      availabilityZone: us-east-1c
      version: "8.0"
      tier: primary
      owner: database-team
      purpose: production
      service: mystisql
      managedBy: kubernetes
      lifecycle: production
      backup: daily
      monitoring: enabled
      alerting: enabled
      securityGroup: sg-123456
      subnet: subnet-123456
      vpc: vpc-123456
      account: 123456789012
      region: us-east-1
      cloud: aws
```

### 生产环境配置

```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: release

discovery:
  type: static

instances:
  - name: production-mysql-master
    type: mysql
    host: mysql-master.production.svc.cluster.local
    port: 3306
    username: mystisql_user
    password: ${MYSQL_MASTER_PASSWORD}  # 从环境变量读取
    database: production_db
    labels:
      environment: production
      role: master
      
  - name: production-mysql-slave
    type: mysql
    host: mysql-slave.production.svc.cluster.local
    port: 3306
    username: mystisql_user
    password: ${MYSQL_SLAVE_PASSWORD}
    database: production_db
    labels:
      environment: production
      role: slave
```

### Phase 3 安全配置示例

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"

# Token 认证配置
auth:
  enabled: true
  token:
    secret: "change-this-secret-in-production"
    expire: "24h"
  whitelist:
    - "/health"
    - "/api/v1/auth/login"

# 审计日志配置
audit:
  enabled: true
  logFile: "/var/log/mystisql/audit.log"
  retentionDays: 30

# SQL 验证器配置
validator:
  enabled: true
  dangerousOperations:
    - "DROP"
    - "TRUNCATE"
    - "DELETE_WITHOUT_WHERE"
    - "UPDATE_WITHOUT_WHERE"
  whitelist: []
  blacklist: []

# WebSocket 配置
websocket:
  enabled: true
  maxConnections: 1000
  idleTimeout: "10m"
  maxConcurrentQueries: 5
```

## 配置文件位置

MystiSql 按以下顺序查找配置文件：

1. 命令行参数 `--config` 或 `-c` 指定的路径
2. 环境变量 `MYSTISQL_CONFIG` 指定的路径
3. 当前目录下的 `config.yaml`
4. `/etc/mystisql/config.yaml`
