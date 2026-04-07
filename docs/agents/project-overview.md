# Project Overview

MystiSql 是一个面向 Kubernetes 集群的数据库访问网关，支持多种数据库类型，并提供多种接入方式。

## 核心定位

- 统一管理和访问集群内数据库实例
- 屏蔽不同数据库驱动差异，提供统一查询入口
- 提供 CLI / REPL、REST API、WebSocket、JDBC Driver 等多种接入方式

## 当前支持的接入层

- **CLI / REPL**：终端交互式查询
- **REST API**：标准 HTTP 查询与管理接口
- **WebSocket**：实时查询与长连接交互
- **JDBC Driver**：支持 Java 生态与常见数据库工具接入
- **WebUI**：浏览器管理与查询界面

## 当前支持的数据库类型

- MySQL
- PostgreSQL
- Oracle
- Redis

此外，代码中还存在 SQLite、MSSQL、ClickHouse、Elasticsearch、etcd 等驱动实现或实验性支持，具体完成度需按代码实际审查判断。

## 主要能力

### 实例发现
- 静态配置发现
- Kubernetes 服务发现

### 连接管理
- 连接池管理
- 健康检查
- 多数据库驱动封装

### 核心服务
- SQL 查询与执行
- 认证与 Token 管理
- 审计日志
- SQL 安全校验
- 事务管理
- 批量执行

### 安全相关能力
- JWT 认证
- 审计日志
- SQL 危险操作拦截
- WebSocket 鉴权

### JDBC 相关能力
- JDBC Driver 基础接入
- 事务支持
- 批量操作支持
- 元数据支持
- PreparedStatement 支持

## 项目阶段

README 中当前阶段标记为 **Phase 3（安全控制层）**。但 AI Agent 不应仅依据 README 判断完成度，必须以代码实际实现为准，逐模块核实功能可用性。

## 目录结构（简版）

```text
cmd/mystisql/          # CLI 入口
internal/
  connection/          # 驱动与连接池
  discovery/           # 实例发现
  service/             # 核心服务
  api/                 # REST / WebSocket
  cli/                 # CLI / REPL
pkg/                   # 公共类型与错误
config/                # 配置文件
test/                  # 集成与 E2E 测试
web/                   # WebUI
jdbc/                  # Java JDBC Driver
```

## Agent 注意事项

- 不要把 README 的阶段描述直接当作真实完成度
- 做评估时必须覆盖 Go 后端、JDBC、WebUI 三个层面
- 如果涉及阶段性能力（如安全、审计、事务），优先查看对应服务实现和接入层注册情况
