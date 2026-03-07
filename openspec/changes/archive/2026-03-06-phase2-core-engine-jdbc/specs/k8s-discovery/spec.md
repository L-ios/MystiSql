## ADDED Requirements

### Requirement: K8s API 动态发现

系统必须支持通过 K8s API 动态发现数据库实例。

#### Scenario: 启动 K8s 发现

- **WHEN** 配置 discovery.type 设置为 "k8s"
- **THEN** 系统必须初始化 K8s 客户端
- **AND** 必须使用 client-go 库连接 K8s API
- **AND** 必须根据配置的命名空间和标签选择器发现实例

#### Scenario: 监听资源变化

- **WHEN** K8s 资源（Service/Pod）发生变化
- **THEN** 系统必须通过 Watch 机制捕获变化
- **AND** 必须自动更新实例列表
- **AND** 必须处理资源添加、更新、删除事件

#### Scenario: 标签选择器配置

- **WHEN** 配置包含 discovery.k8s.selectors
- **THEN** 系统必须根据标签选择器过滤资源
- **AND** 必须为每个选择器指定数据库类型
- **AND** 必须支持多个标签选择器

#### Scenario: 端口映射配置

- **WHEN** 配置包含 discovery.k8s.portMapping
- **THEN** 系统必须根据数据库类型映射默认端口
- **AND** 必须支持自定义端口配置

#### Scenario: K8s API 权限

- **WHEN** 系统连接 K8s API
- **THEN** 必须使用 ServiceAccount 或 kubeconfig 进行认证
- **AND** 必须具有 list 和 watch 权限
- **AND** 必须处理权限不足的情况

---

### Requirement: ConfigMap 配置源

系统必须支持从 ConfigMap 读取数据库实例配置。

#### Scenario: 配置 ConfigMap 源

- **WHEN** 配置包含 discovery.config.sources.configMap
- **THEN** 系统必须从指定的 ConfigMap 读取实例配置
- **AND** 必须支持指定命名空间和 ConfigMap 名称
- **AND** 必须监控 ConfigMap 变化

#### Scenario: 多配置源优先级

- **WHEN** 同时配置多个发现源
- **THEN** 系统必须按优先级合并实例
- **AND** 静态配置优先级最高
- **AND** K8s 动态发现次之
- **AND** ConfigMap 配置次之

---

### Requirement: 实例事件通知

系统必须提供实例变化的事件通知机制。

#### Scenario: 实例添加事件

- **WHEN** 发现新的数据库实例
- **THEN** 系统必须触发 InstanceAdded 事件
- **AND** 必须更新实例注册表
- **AND** 必须记录实例添加日志

#### Scenario: 实例更新事件

- **WHEN** 实例配置发生变化
- **THEN** 系统必须触发 InstanceUpdated 事件
- **AND** 必须更新实例注册表
- **AND** 必须记录实例更新日志

#### Scenario: 实例删除事件

- **WHEN** 实例从 K8s 中删除
- **THEN** 系统必须触发 InstanceDeleted 事件
- **AND** 必须从实例注册表中移除
- **AND** 必须记录实例删除日志