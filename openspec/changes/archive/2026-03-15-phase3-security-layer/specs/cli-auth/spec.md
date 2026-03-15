## ADDED Requirements

### Requirement: CLI Token 配置
系统 SHALL 支持在 CLI 中配置和管理 Token。

#### Scenario: 通过配置文件设置 Token
- **WHEN** 用户在 `~/.mystisql/config.yaml` 中设置 `token: <jwt_token>`
- **THEN** CLI 使用该 Token 进行认证

#### Scenario: 通过环境变量设置 Token
- **WHEN** 用户设置环境变量 `MYSTISQL_TOKEN=<jwt_token>`
- **THEN** CLI 使用该 Token 进行认证

#### Scenario: 通过命令行参数传递 Token
- **WHEN** 用户执行命令时添加 `--token <jwt_token>` 参数
- **THEN** CLI 使用该 Token 进行认证

### Requirement: CLI Token 优先级
系统 SHALL 定义 Token 传递方式的优先级。

#### Scenario: 优先级顺序
- **WHEN** 同时存在多种 Token 配置方式
- **THEN** 优先级为：命令行参数 > 环境变量 > 配置文件

### Requirement: CLI 认证失败处理
系统 SHALL 在认证失败时返回明确的错误信息。

#### Scenario: Token 无效
- **WHEN** CLI 使用的 Token 无效或过期
- **THEN** 系统返回错误 "认证失败：Token 无效或已过期" 并退出

#### Scenario: 未提供 Token
- **WHEN** CLI 命令需要认证但未提供 Token
- **THEN** 系统返回错误 "未提供认证 Token，请使用 --token 参数或配置环境变量 MYSTISQL_TOKEN" 并退出

### Requirement: CLI Token 管理命令
系统 SHALL 提供 CLI 命令管理 Token。

#### Scenario: 生成 Token
- **WHEN** 用户执行 `mystisql auth token --user-id admin --role admin`
- **THEN** 系统生成并返回 JWT Token

#### Scenario: 撤销 Token
- **WHEN** 管理员执行 `mystisql auth revoke --token <jwt_token>`
- **THEN** 系统撤销该 Token 并返回成功信息

### Requirement: CLI 查询命令认证
系统 SHALL 为查询命令添加认证支持。

#### Scenario: 查询命令使用 Token
- **WHEN** 用户执行 `mystisql query --instance local-mysql "SELECT * FROM users" --token <jwt_token>`
- **THEN** CLI 使用 Token 认证并执行查询

#### Scenario: 实例列表命令使用 Token
- **WHEN** 用户执行 `mystisql instances list --token <jwt_token>`
- **THEN** CLI 使用 Token 认证并返回实例列表
