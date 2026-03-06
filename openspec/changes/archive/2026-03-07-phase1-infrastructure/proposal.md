## Why

MystiSql 需要一个坚实的基础设施层，让开发人员和运维人员能够在 Kubernetes 集群外部访问集群内部的数据库实例。

**当前问题：**
1. 数据库实例只能从集群内部访问，开发人员在本地无法直接连接
2. 即使在集群内，也需要安装各种数据库客户端工具
3. 没有统一的方式管理和访问不同类型的数据库

**为什么是现在？**
这是项目的第一阶段（Phase 1），需要先建立最基础的能力，后续阶段才能在此基础上构建更高级的功能（如 K8s 动态发现、认证授权、WebUI 等）。

## What Changes

本次变更将建立 MystiSql 的基础设施层，实现以下功能：

### 新增功能
- **静态配置发现**：从 YAML 配置文件中读取数据库实例信息
- **MySQL 连接**：能够连接到 MySQL 数据库并执行 SQL 查询
- **命令行工具**：提供 `mystisql` 命令行工具，支持查询和实例管理
- **REST API**：提供 HTTP 接口，支持查询和健康检查
- **配置管理**：支持 YAML 配置文件和环境变量配置

### 不包含的功能（后续阶段）
- K8s API 动态发现（Phase 2）
- 连接池（Phase 2）
- PostgreSQL/Oracle/Redis 支持（Phase 3-5）
- 认证授权（Phase 3）
- WebSocket（Phase 3）
- WebUI（Phase 4）

## Capabilities

### New Capabilities

本次变更将引入以下 5 个新能力，每个能力对应一个独立的规格文档：

- `instance-discovery-static`：静态配置发现 - 从配置文件中读取和管理数据库实例信息
- `mysql-connection`：MySQL 连接 - 建立到 MySQL 数据库的连接并执行查询
- `cli-interface`：命令行界面 - 提供 CLI 工具用于查询和管理实例
- `rest-api`：REST API - 提供 HTTP 接口用于查询和监控
- `config-management`：配置管理 - 配置文件的读取、验证和管理

### Modified Capabilities

无（这是项目的首次实现）

## Impact

### 新增代码
- `cmd/mystisql/main.go` - CLI 入口程序
- `internal/discovery/static/` - 静态发现实现
- `internal/connection/mysql/` - MySQL 连接实现
- `internal/cli/` - CLI 命令实现
- `internal/api/rest/` - REST API 实现
- `pkg/types/` - 核心类型定义（DatabaseInstance、Config 等）
- `pkg/errors/` - 错误定义
- `config/config.yaml` - 配置文件示例

### 新增依赖
- `github.com/go-sql-driver/mysql` - MySQL 驱动
- `github.com/gin-gonic/gin` - Web 框架
- `github.com/spf13/cobra` - CLI 框架
- `github.com/spf13/viper` - 配置管理

### 用户影响
- 用户可以通过 `mystisql` 命令行工具连接和查询 MySQL 数据库
- 用户可以通过 HTTP API 执行查询
- 用户需要编写配置文件来定义数据库实例

### 系统要求
- Go 1.21+
- 可访问的 MySQL 数据库实例（用于测试）
