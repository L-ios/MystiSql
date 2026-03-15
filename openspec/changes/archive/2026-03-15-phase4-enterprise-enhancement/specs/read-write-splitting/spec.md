## ADDED Requirements

### Requirement: 主从实例识别
系统 SHALL 根据实例标签识别主库和从库。

#### Scenario: 通过 role 标签识别主从
- **WHEN** 实例标签包含 `role=master` 或 `role=slave`
- **THEN** 系统正确识别实例角色

#### Scenario: 无标签实例默认为主库
- **WHEN** 实例没有 role 标签
- **THEN** 系统将其视为主库处理

### Requirement: SQL 语句类型识别
系统 SHALL 自动识别 SQL 语句是读操作还是写操作。

#### Scenario: SELECT 语句识别
- **WHEN** 执行 `SELECT * FROM users`
- **THEN** 系统识别为读操作

#### Scenario: INSERT/UPDATE/DELETE 语句识别
- **WHEN** 执行 `INSERT INTO users ...`
- **THEN** 系统识别为写操作

#### Scenario: 事务内语句识别
- **WHEN** 在事务中执行 SELECT
- **THEN** 系统识别为写操作（需要走主库）

### Requirement: 读请求路由到从库
系统 SHALL 将读请求优先路由到可用的从库。

#### Scenario: 路由到从库
- **WHEN** 执行 SELECT 且存在可用从库
- **THEN** 请求被路由到从库

#### Scenario: 无从库时路由到主库
- **WHEN** 执行 SELECT 但无从库可用
- **THEN** 请求被路由到主库

#### Scenario: 从库延迟过高时切主库
- **WHEN** 从库复制延迟超过阈值（默认 1s）
- **THEN** 请求被路由到主库

### Requirement: 写请求路由到主库
系统 SHALL 将写请求路由到主库。

#### Scenario: 路由到主库
- **WHEN** 执行 INSERT/UPDATE/DELETE
- **THEN** 请求被路由到主库

### Requirement: 读写分离配置
系统 SHALL 支持配置读写分离策略。

#### Scenario: 启用读写分离
- **WHEN** 配置 `readWriteSplit.enabled = true`
- **THEN** 系统启用读写分离功能

#### Scenario: 配置延迟阈值
- **WHEN** 配置 `readWriteSplit.maxReplicaLag = "2s"`
- **THEN** 延迟超过 2s 的从库不接收读请求
