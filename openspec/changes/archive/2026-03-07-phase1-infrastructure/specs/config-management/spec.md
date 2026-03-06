## ADDED Requirements

### Requirement: 支持 YAML 配置文件

系统必须支持通过 YAML 文件加载配置。

#### Scenario: 加载有效的 YAML 配置

- **WHEN** 提供了有效的 YAML 配置文件
- **THEN** 系统必须成功解析文件内容
- **AND** 必须将配置映射到 Config 结构体
- **AND** 必须支持嵌套结构

#### Scenario: 处理不存在的配置文件

- **WHEN** 指定的配置文件路径不存在
- **THEN** 系统必须返回 ErrConfigNotFound 错误
- **AND** 错误消息必须包含文件路径

#### Scenario: 处理无效的 YAML 格式

- **WHEN** 配置文件包含无效的 YAML 语法
- **THEN** 系统必须返回 ErrConfigParseFailed 错误
- **AND** 错误消息必须包含语法错误的位置信息

---

### Requirement: 多配置文件路径支持

系统必须支持从多个标准位置查找配置文件。

#### Scenario: 按优先级查找配置文件

- **WHEN** 未指定配置文件路径
- **THEN** 系统必须按以下顺序查找：
  1. ./config.yaml（当前目录，最高优先级）
  2. ./config/config.yaml
  3. /etc/mystisql/config.yaml（系统范围，最低优先级）
- **AND** 必须使用第一个找到的文件
- **AND** 如果都找不到，必须返回 ErrConfigNotFound 错误

#### Scenario: 显式指定配置文件

- **WHEN** 使用 --config 标志指定配置文件路径
- **THEN** 系统必须只使用指定的文件
- **AND** 必须不查找其他路径
- **AND** 如果文件不存在，必须返回错误

#### Scenario: 配置文件优先级

- **WHEN** 同时存在多个配置文件位置
- **THEN** 当前目录的配置文件优先级最高
- **AND** 必须忽略其他位置的配置文件

---

### Requirement: 数据库实例配置

系统必须支持配置多个数据库实例。

#### Scenario: 配置单个 MySQL 实例

- **WHEN** 配置文件包含一个 MySQL 实例定义
- **THEN** 必须包含必填字段：name、type、host、port
- **AND** 可选字段：username、password、database
- **AND** type 必须是支持的数据库类型之一

#### Scenario: 配置多个数据库实例

- **WHEN** 配置文件包含多个实例定义
- **THEN** 系统必须加载所有实例
- **AND** 每个实例的 name 必须唯一
- **AND** 如果有重复名称，必须返回验证错误

#### Scenario: 实例配置验证

- **WHEN** 加载实例配置
- **THEN** 必须验证 name 不为空
- **AND** 必须验证 host 不为空
- **AND** 必须验证 port 在有效范围内（1-65535）
- **AND** 必须验证 type 是支持的数据库类型

---

### Requirement: 服务器配置

系统必须支持 API 服务器的配置。

#### Scenario: 配置服务器监听地址

- **WHEN** 配置文件包含 server.host 字段
- **THEN** 系统必须使用指定的主机地址
- **AND** 默认值必须是 "0.0.0.0"（监听所有接口）

#### Scenario: 配置服务器监听端口

- **WHEN** 配置文件包含 server.port 字段
- **THEN** 系统必须使用指定的端口号
- **AND** 默认值必须是 8080
- **AND** 端口号必须在有效范围内（1-65535）

#### Scenario: 配置运行模式

- **WHEN** 配置文件包含 server.mode 字段
- **THEN** 模式必须是 "debug" 或 "release" 之一
- **AND** 默认值必须是 "release"
- **AND** Gin 框架必须运行在指定模式

---

### Requirement: 发现配置

系统必须支持服务发现方法的配置。

#### Scenario: 配置静态发现

- **WHEN** discovery.type 设置为 "static"
- **THEN** 系统必须从配置文件的 instances 字段加载实例
- **AND** 必须不执行外部发现

#### Scenario: 不支持的发现类型

- **WHEN** discovery.type 设置为不支持的值
- **THEN** 系统必须返回 ErrConfigInvalid 错误
- **AND** 错误消息必须列出支持的发现类型

---

### Requirement: 环境变量覆盖

系统必须支持通过环境变量覆盖配置。

#### Scenario: 环境变量覆盖服务器配置

- **WHEN** 设置了环境变量 MYSTISQL_SERVER_PORT
- **THEN** 该值必须覆盖配置文件中的 server.port
- **AND** 系统必须记录覆盖信息（如果启用了 verbose）

#### Scenario: 环境变量命名规范

- **WHEN** 使用环境变量覆盖配置
- **THEN** 环境变量必须使用格式：MYSTISQL_<SECTION>_<FIELD>
- **AND** 必须支持嵌套结构（如 MYSTISQL_SERVER_PORT）
- **AND** 字段名必须大写

#### Scenario: 环境变量优先级

- **WHEN** 同时存在配置文件和环境变量
- **THEN** 环境变量的优先级必须高于配置文件
- **AND** 未设置的环境变量不得影响配置文件的值

---

### Requirement: 配置验证

系统必须在加载配置时进行验证。

#### Scenario: 验证必填字段

- **WHEN** 配置文件缺少必填字段
- **THEN** 系统必须返回 ErrConfigInvalid 错误
- **AND** 错误消息必须明确指出缺少的字段
- **AND** 必须在启动时失败，而不是运行时

#### Scenario: 验证字段格式

- **WHEN** 配置字段不符合格式要求
- **THEN** 系统必须返回验证错误
- **AND** 必须提供字段名和期望格式
- **AND** 端口号必须是整数且在 1-65535 范围内

#### Scenario: 验证数据库类型

- **WHEN** 实例的 type 字段不是支持的数据库类型
- **THEN** 系统必须返回验证错误
- **AND** 必须列出支持的类型：mysql、postgresql、oracle、redis

---

### Requirement: 配置热重载

系统必须支持配置的热重载（通过信号）。

#### Scenario: 接收 SIGHUP 信号重载配置

- **WHEN** 进程收到 SIGHUP 信号
- **THEN** 系统必须重新加载配置文件
- **AND** 必须应用新的配置（如果有效）
- **AND** 必须记录重载事件

#### Scenario: 重载失败保持旧配置

- **WHEN** 配置重载时发生错误（如文件不存在、格式错误）
- **THEN** 系统必须继续使用旧配置
- **AND** 必须记录错误信息
- **AND** 必须不影响正在运行的服务

#### Scenario: 配置变更生效

- **WHEN** 成功重载配置
- **THEN** 新增的实例必须立即可用
- **AND** 删除的实例必须不再可用
- **AND** 修改的实例配置必须立即生效

---

### Requirement: 配置文件示例

系统必须提供配置文件示例。

#### Scenario: 提供完整的配置示例

- **WHEN** 用户查看配置文件示例
- **THEN** 必须包含所有配置字段的示例
- **AND** 必须包含注释说明每个字段的作用
- **AND** 必须包含至少一个 MySQL 实例示例

#### Scenario: 配置示例包含默认值

- **WHEN** 用户参考配置示例
- **THEN** 必须显示所有字段的默认值
- **AND** 必须标明哪些字段是必填的
- **AND** 必须标明哪些字段是可选的
