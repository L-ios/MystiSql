## MODIFIED Requirements

### Requirement: Query 命令

系统必须提供 query 命令来执行 SQL 查询。

#### Scenario: 执行查询命令

- **WHEN** 用户运行 `mystisql query --instance <instance-name> "<sql>"`
- **THEN** 系统必须连接到指定实例
- **AND** 必须执行 SQL 查询
- **AND** 必须在格式化表格中显示结果
- **AND** 必须显示查询执行时间

#### Scenario: 查询命令 - 实例不存在

- **WHEN** 用户运行查询命令时使用了不存在的实例名
- **THEN** 系统必须显示错误："实例未找到：<name>"
- **AND** 必须以非零状态码退出（1）
- **AND** 必须列出可用实例

#### Scenario: 查询命令 - 无效 SQL

- **WHEN** 用户运行查询命令时使用了无效的 SQL 语法
- **THEN** 系统必须显示 MySQL 错误消息
- **AND** 必须以非零状态码退出（1）
- **AND** 必须显示 SQL 语句（便于调试）

#### Scenario: 查询命令 - 输出格式选项

- **WHEN** 用户运行查询命令时带有 `--format json` 标志
- **THEN** 结果必须格式化为 JSON
- **AND** JSON 必须是有效的且格式化良好的
- **AND** 默认格式必须是 table

#### Scenario: 查询命令 - CSV 输出

- **WHEN** 用户运行查询命令时带有 `--format csv` 标志
- **THEN** 结果必须格式化为 CSV
- **AND** CSV 必须包含标题行
- **AND** 特殊字符必须正确转义

#### Scenario: 查询命令 - 使用 Token 认证

- **WHEN** 用户运行 `mystisql query --instance <name> "<sql>" --token <jwt_token>`
- **THEN** 系统必须使用 Token 进行认证
- **AND** 认证成功后执行查询

#### Scenario: 查询命令 - 认证失败

- **WHEN** 用户运行查询命令但 Token 无效或已过期
- **THEN** 系统必须显示错误："认证失败：Token 无效或已过期"
- **AND** 必须以状态码 1 退出

#### Scenario: 查询命令 - 未提供 Token

- **WHEN** 用户运行查询命令但未提供 Token（且未在配置文件或环境变量中设置）
- **THEN** 系统必须显示错误："未提供认证 Token，请使用 --token 参数或配置环境变量 MYSTISQL_TOKEN"
- **AND** 必须以状态码 1 退出

---

### Requirement: Instances 命令

系统必须提供 instances 命令来管理数据库实例。

#### Scenario: 列出所有实例

- **WHEN** 用户运行 `mystisql instances list`
- **THEN** 系统必须显示所有已注册的实例
- **AND** 每个实例必须显示：名称、类型、主机、端口、状态
- **AND** 输出必须是格式化的表格

#### Scenario: 列出实例 - 无实例配置

- **WHEN** 用户运行 instances list 但没有配置实例
- **THEN** 系统必须显示："未配置实例"
- **AND** 必须以状态码 0 退出
- **AND** 必须提示用户如何添加实例

#### Scenario: 列出实例 - JSON 格式

- **WHEN** 用户运行 `mystisql instances list --format json`
- **THEN** 结果必须格式化为 JSON 数组
- **AND** JSON 必须包含所有实例字段

#### Scenario: 获取单个实例详情

- **WHEN** 用户运行 `mystisql instances get <instance-name>`
- **THEN** 系统必须显示指定实例的详细信息
- **AND** 必须包括所有配置字段（敏感信息脱敏）
- **AND** 如果实例不存在，必须返回错误

#### Scenario: 列出实例 - 使用 Token 认证

- **WHEN** 用户运行 `mystisql instances list --token <jwt_token>`
- **THEN** 系统必须使用 Token 进行认证
- **AND** 认证成功后返回实例列表

---

### Requirement: 全局标志支持

系统必须支持全局配置标志。

#### Scenario: 使用自定义配置文件

- **WHEN** 用户运行命令时带有 `--config /path/to/config.yaml` 标志
- **THEN** 系统必须从指定文件加载配置
- **AND** 必须验证配置文件格式
- **AND** 如果文件不存在，必须返回错误："配置文件未找到"

#### Scenario: 指定多个配置位置

- **WHEN** 未指定 --config 标志
- **THEN** 系统必须按顺序查找配置文件：
  1. ./config.yaml（当前目录）
  2. ./config/config.yaml
  3. /etc/mystisql/config.yaml（系统范围）
- **AND** 必须使用第一个找到的文件

#### Scenario: 配置文件优先级

- **WHEN** 同时存在多个配置文件位置
- **THEN** 当前目录的配置文件优先级最高
- **AND** 系统范围的配置文件优先级最低

#### Scenario: Token 传递方式优先级

- **WHEN** 同时存在多种 Token 配置方式
- **THEN** 优先级为：命令行参数 --token > 环境变量 MYSTISQL_TOKEN > 配置文件 token 字段

## ADDED Requirements

### Requirement: Auth 子命令

系统必须提供 auth 子命令管理 Token。

#### Scenario: 生成 Token

- **WHEN** 用户运行 `mystisql auth token --user-id <user_id> --role <role>`
- **THEN** 系统必须生成并返回 JWT Token
- **AND** 必须显示 Token 的过期时间

#### Scenario: 撤销 Token

- **WHEN** 管理员运行 `mystisql auth revoke --token <jwt_token>`
- **THEN** 系统必须撤销该 Token
- **AND** 必须显示成功消息

#### Scenario: 查看当前 Token 信息

- **WHEN** 用户运行 `mystisql auth info --token <jwt_token>`
- **THEN** 系统必须显示 Token 的详细信息（user_id、role、过期时间）
