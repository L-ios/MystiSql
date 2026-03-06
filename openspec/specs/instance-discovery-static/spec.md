# 静态实例发现规范

## Purpose

定义 MystiSql 的静态实例发现功能，包括从配置文件加载数据库实例、实例注册、查询和状态管理等功能，确保系统能够正确发现和管理数据库实例。

## Requirements

### Requirement: 支持静态配置文件

系统必须支持从 YAML 配置文件加载数据库实例定义。

#### Scenario: 加载有效的配置文件

- **WHEN** 提供了包含数据库实例定义的有效 YAML 配置文件
- **THEN** 系统必须成功解析并加载所有实例
- **AND** 每个实例必须包含 name、type、host、port 字段
- **AND** 可选字段包括 username、password、database

#### Scenario: 处理配置文件不存在的情况

- **WHEN** 指定的配置文件不存在
- **THEN** 系统必须返回清晰的错误消息："配置文件未找到"
- **AND** 系统不得崩溃或异常退出

#### Scenario: 处理无效的配置格式

- **WHEN** 配置文件包含无效的 YAML 语法
- **THEN** 系统必须返回验证错误并显示行号
- **AND** 系统不得加载任何实例

---

### Requirement: 实例注册功能

系统必须提供接口来注册数据库实例。

#### Scenario: 注册新实例

- **WHEN** 向注册中心提供有效的 DatabaseInstance 对象
- **THEN** 实例必须被添加到实例存储中
- **AND** 实例必须可以通过名称查询

#### Scenario: 注册重复实例

- **WHEN** 尝试注册已存在的同名实例
- **THEN** 系统必须返回 ErrInstanceAlreadyExists 错误
- **AND** 现有实例必须保持不变

---

### Requirement: 实例发现接口

系统必须为所有发现实现提供标准的 InstanceDiscoverer 接口。

#### Scenario: 实现 InstanceDiscoverer 接口

- **WHEN** 发现提供者实现 InstanceDiscoverer 接口
- **THEN** 它必须提供 Name() string 方法
- **AND** 必须提供 Discover(ctx context.Context) ([]*DatabaseInstance, error) 方法

#### Scenario: 从静态配置发现实例

- **WHEN** 调用 StaticDiscoverer.Discover() 方法
- **THEN** 必须返回配置文件中加载的所有实例
- **AND** 如果配置未加载，必须返回错误

---

### Requirement: 实例查询功能

系统必须提供查询已注册实例的能力。

#### Scenario: 按名称获取实例

- **WHEN** 使用有效实例名调用 GetInstance(name) 方法
- **THEN** 系统必须返回对应的 DatabaseInstance 对象
- **AND** 如果名称不存在，必须返回 ErrInstanceNotFound 错误

#### Scenario: 列出所有实例

- **WHEN** 调用 ListInstances() 方法
- **THEN** 系统必须返回所有已注册的实例
- **AND** 如果没有注册实例，必须返回空列表（不是 nil）

---

### Requirement: 实例状态管理

系统必须跟踪和管理每个实例的状态。

#### Scenario: 检查实例状态

- **WHEN** 实例注册到系统中
- **THEN** 每个实例必须有状态字段（Status）
- **AND** 状态必须是以下之一：Unknown、Healthy、Unhealthy

#### Scenario: 更新实例状态

- **WHEN** 实例的连接状态发生变化
- **THEN** 系统必须能够更新实例的状态
- **AND** 状态变更必须对后续查询可见