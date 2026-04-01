## Why

`internal/service/router/sql_parser.go` 已实现基础的 SQL 类型识别（SELECT/INSERT/UPDATE/DELETE）和读写判断，但从未被任何代码调用。配置中支持多实例，但所有查询都路由到同一实例，无法利用主从架构实现读写分离。需要将 sql_parser 集成到查询路由层，按 SQL 类型路由到不同的连接池（写→主库，读→从库）。

## What Changes

### 配置模型
- **实例角色标记**：配置中每个实例新增 `role` 字段（`primary`/`replica`/`readwrite`）
- **主从关联**：新增 `replicaOf` 字段，从库指向其主库名称
- **配置示例**：
  ```yaml
  instances:
    - name: "mysql-primary"
      type: "mysql"
      role: "primary"
    - name: "mysql-replica-1"
      type: "mysql"
      role: "replica"
      replicaOf: "mysql-primary"
  ```

### 路由逻辑
- **写操作**（INSERT/UPDATE/DELETE/DDL）→ 路由到 primary 实例的连接池
- **读操作**（SELECT）→ 路由到 replica 实例的连接池（轮询或随机负载均衡）
- **事务内查询**→ 始终路由到 primary（保证一致性）
- **未配置 replica**→ 所有查询走 primary（向后兼容）

### sql_parser.go 升级
- **正则缓存**：编译一次复用，不再每次调用重新编译
- **多语句处理**：取第一条语句类型决定路由
- **集成到 Engine**：`getConnectionPool` 根据解析出的 SQL 类型选择目标实例

## Capabilities

### Modified Capabilities
- `read-write-splitting`: 从孤立代码改为集成到 Engine 路由层
- `rest-api`: 查询请求可指定目标实例或自动路由

## Impact

### 受影响的代码
- `internal/service/router/sql_parser.go` — 正则缓存、多语句处理
- `internal/service/query/engine.go` — 路由逻辑集成
- `internal/config/` — 新增 role/replicaOf 配置字段

### 前置条件
- Gate 0 完成（可编译）
- Gate 1 完成（Engine 工厂注册改造完成，路由逻辑基于改造后的 Engine）

### Done 标准
- 配置主从后，SELECT 自动路由到 replica，INSERT 自动路由到 primary
- 未配置主从时行为不变
- 事务内查询始终走 primary

### 信心
**50%** — 需要设计主从关联的数据模型和负载均衡策略；事务内查询一致性需要仔细处理；sql_parser.go 的多语句处理可能引入误判
