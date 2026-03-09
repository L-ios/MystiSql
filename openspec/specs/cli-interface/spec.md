# CLI 接口规范

## Purpose

定义 MystiSql 命令行界面的功能和行为，确保用户可以通过命令行工具管理数据库实例和执行 SQL 查询。

## Requirements

### Requirement: CLI 框架设置

系统必须使用 Cobra 框架提供命令行界面。

#### Scenario: CLI 初始化

- **WHEN** 执行 mystisql 命令
- **THEN** CLI 必须使用 Cobra 框架初始化
- **AND** 必须支持 --help 标志显示帮助信息
- **AND** 必须支持 --version 标志显示版本信息

#### Scenario: 子命令支持

- **WHEN** CLI 初始化完成
- **THEN** 必须支持子命令进行不同操作
- **AND** 每个子命令必须有独立的帮助文档
- **AND** 子命令必须自动补全（如果 shell 支持）

#### Scenario: 处理无效命令

- **WHEN** 用户输入无效的子命令
- **THEN** 系统必须显示错误消息："未知命令"
- **AND** 必须显示可用命令列表
- **AND** 必须以非零状态码退出

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

---

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

### Requirement: Version 命令

系统必须提供 version 命令显示版本信息。

#### Scenario: 显示版本信息

- **WHEN** 用户运行 `mystisql version`
- **THEN** 系统必须显示版本号（如 v0.1.0）
- **AND** 必须显示构建日期
- **AND** 必须显示 Git 提交哈希
- **AND** 必须显示 Go 版本

#### Scenario: 使用 --version 标志

- **WHEN** 用户运行任何命令时带有 `--version` 标志
- **THEN** 系统必须显示版本信息
- **AND** 必须立即退出（不执行其他操作）

---

### Requirement: Verbose 日志

系统必须支持详细日志用于调试。

#### Scenario: 启用 verbose 模式

- **WHEN** 用户运行命令时带有 `--verbose` 或 `-v` 标志
- **THEN** 系统必须输出详细日志信息
- **AND** 必须包括连接详情（不含密码）
- **AND** 必须包括查询执行时间
- **AND** 必须包括配置加载详情

#### Scenario: 日志级别控制

- **WHEN** 用户使用 `-v` 标志多次（如 `-vv` 或 `-vvv`）
- **THEN** 日志详细程度必须相应增加
- **AND** `-v` = Info 级别
- **AND** `-vv` = Debug 级别
- **AND** `-vvv` = Trace 级别

---

### Requirement: 错误处理

系统必须提供清晰的用户友好的错误消息。

#### Scenario: 显示错误上下文

- **WHEN** 发生错误
- **THEN** 错误消息必须清晰明了
- **AND** 必须包含建议的解决方案（如果适用）
- **AND** 必须以红色显示错误（如果终端支持）

#### Scenario: 退出状态码

- **WHEN** 命令执行成功
- **THEN** 必须以状态码 0 退出

#### Scenario: 错误退出状态码

- **WHEN** 命令执行失败
- **THEN** 必须以非零状态码退出
- **AND** 状态码 1 = 一般错误
- **AND** 状态码 2 = 配置错误
- **AND** 状态码 3 = 连接错误

---

### Requirement: 帮助文档

系统必须提供全面的帮助文档。

#### Scenario: 显示命令帮助

- **WHEN** 用户运行 `mystisql --help` 或 `mystisql -h`
- **THEN** 系统必须显示主帮助信息
- **AND** 必须列出所有可用命令
- **AND** 必须显示全局标志

#### Scenario: 显示子命令帮助

- **WHEN** 用户运行 `mystisql query --help`
- **THEN** 系统必须显示 query 命令的详细帮助
- **AND** 必须显示用法示例
- **AND** 必须显示所有可用标志

#### Scenario: 显示使用示例

- **WHEN** 显示帮助信息
- **THEN** 必须包含实际使用示例
- **AND** 示例必须是可执行的
- **AND** 示例必须覆盖常见用例

---

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