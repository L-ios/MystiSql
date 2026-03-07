## Why

Phase 1 已完成基础设施层，实现了静态配置发现和 MySQL 连接基础功能。现在需要实现核心引擎能力和 JDBC 驱动，使开发者能在 IDE 中直接连接 K8s 数据库，同时支持 K8s 动态发现和连接池管理，提升系统可靠性和性能。

## What Changes

- **服务发现层**：实现 K8s API 动态发现，支持 ConfigMap 配置源，添加实例状态监控
- **连接层**：实现连接池管理，添加连接健康检查和自动重连机制
- **服务层**：实现 SQL 解析和路由，添加查询超时控制和结果集大小限制，实现表结构缓存
- **接入层**：完善 RESTful API，实现 SQL 执行接口和实例列表接口，增强 CLI 查询命令
- **JDBC 驱动**：实现完整的 JDBC 驱动，支持 IDE 工具集成

## Capabilities

### New Capabilities
- **k8s-discovery**: K8s API 动态发现能力，通过 client-go 监听 Service/Pod 变化
- **connection-pool**: 连接池管理，支持连接复用、健康检查和自动重连
- **sql-engine**: SQL 解析和路由引擎，支持查询超时控制和结果集限制
- **jdbc-driver**: JDBC 驱动实现，支持 IDE 工具集成
- **instance-health**: 实例健康状态监控和检查

### Modified Capabilities
- **mysql-connection**: 增强连接管理，集成连接池功能
- **rest-api**: 完善 API 端点，添加 SQL 执行和实例管理接口
- **cli-interface**: 增强 CLI 命令，支持更复杂的查询操作

## Impact

- **代码结构**：新增 internal/discovery/k8s、internal/connection/pool、internal/service/query、internal/jdbc 等包
- **依赖**：新增 client-go、pgx 等依赖
- **API**：新增 POST /api/v1/query 接口，完善 GET /api/v1/instances 接口
- **CLI**：增强 mystisql query 命令，支持更多选项
- **部署**：需要 K8s RBAC 权限以支持动态发现

## 交付物

- 完整的核心引擎功能，支持 K8s 动态发现和连接池管理
- JDBC 驱动发布，可在 DataGrip、DBeaver 等 IDE 工具中使用
- 完善的 RESTful API 和 CLI 命令
- 详细的技术文档和使用示例