## Purpose

定义 MystiSql REPL 中数据库实例切换功能的规范，支持用户通过 SQL 风格命令在多个数据库实例间灵活切换。

## REMOVED Requirements

### Requirement: 实例列表展示

**Reason**: REPL 模式下实例列表通过 `SHOW INSTANCES` 命令实现，不需要专门的列表视图。

**Migration**: 使用 `SHOW INSTANCES` 命令查看所有实例。

### Requirement: 实例分组管理

**Reason**: 简化初始实现，分组功能可通过配置文件管理。

**Migration**: 在配置文件中组织实例配置。

### Requirement: 切换历史记录

**Reason**: REPL 模式下历史记录通过命令历史实现。

**Migration**: 使用上下箭头浏览历史命令，包括 USE 命令。

## MODIFIED Requirements

### Requirement: 实例切换操作

系统必须支持用户通过命令切换当前连接的数据库实例。

#### Scenario: USE 命令切换

- **WHEN** 用户执行 `USE <instance-name>` 命令
- **THEN** 系统必须切换到指定实例
- **AND** 必须显示 "Database changed" 消息
- **AND** 提示符必须更新为新实例名称

#### Scenario: 切换前连接测试

- **WHEN** 用户切换到新实例
- **THEN** 系统必须先测试连接是否可用
- **AND** 如果连接失败，必须显示错误信息
- **AND** 必须保持当前实例连接不变

#### Scenario: 显示当前实例

- **WHEN** 用户执行 `SELECT INSTANCE()` 或 `STATUS` 命令
- **THEN** 系统必须显示当前实例名称
- **AND** 必须显示实例类型和连接状态

---

### Requirement: 自动重连机制

系统必须支持自动重连功能。

#### Scenario: 连接断开检测

- **WHEN** 检测到连接断开
- **THEN** 系统必须显示 "Connection lost" 警告
- **AND** 必须禁止执行新的 SQL 命令

#### Scenario: 自动重连

- **WHEN** 连接断开后
- **THEN** 系统必须自动尝试重连
- **AND** 重连间隔必须逐步增加
- **AND** 最大重连间隔不超过 30 秒

#### Scenario: 重连成功

- **WHEN** 自动重连成功
- **THEN** 系统必须显示 "Connection restored" 消息
- **AND** 用户可以继续执行 SQL 命令

---

### Requirement: 实例详情查看

系统必须支持查看实例的详细信息。

#### Scenario: 显示实例详情

- **WHEN** 用户执行 `SHOW INSTANCE <name>` 命令
- **THEN** 系统必须显示实例的详细配置信息
- **AND** 必须包括：名称、类型、地址、端口
- **AND** 敏感信息必须脱敏显示

#### Scenario: 显示所有实例

- **WHEN** 用户执行 `SHOW INSTANCES` 命令
- **THEN** 系统必须显示所有已配置的实例列表
- **AND** 必须以表格形式显示
- **AND** 必须包含名称、类型、状态列

---

### Requirement: 连接配置管理

系统必须支持查看连接参数。

#### Scenario: 查看连接参数

- **WHEN** 用户执行 `SHOW INSTANCE STATUS` 命令
- **THEN** 系统必须显示当前实例的连接参数
- **AND** 必须包括：超时时间、连接池状态
- **AND** 敏感参数必须脱敏

---

### Requirement: 错误处理

系统必须清晰显示实例相关的错误信息。

#### Scenario: 连接错误显示

- **WHEN** 连接实例失败
- **THEN** 系统必须显示详细的错误信息
- **AND** 错误信息必须包括：错误类型、错误描述
- **AND** 必须提供可能的解决方案

#### Scenario: 实例不存在

- **WHEN** 用户尝试切换到不存在的实例
- **THEN** 系统必须显示 "Unknown instance" 错误
- **AND** 必须提示用户使用 `SHOW INSTANCES` 查看可用实例
