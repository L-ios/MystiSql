## ADDED Requirements

### Requirement: PostgreSQL 驱动 e2e 测试
系统 SHALL 在真实的 PostgreSQL 数据库上进行端到端测试，验证驱动的所有核心功能。

#### Scenario: e2e 测试 - 成功建立 PostgreSQL 连接
- **WHEN** e2e 测试使用有效配置连接到 PostgreSQL 容器
- **THEN** 系统 SHALL 成功建立连接
- **AND** 连接 SHALL 通过健康检查
- **AND** 连接 SHALL 可以立即执行查询

#### Scenario: e2e 测试 - 连接池管理
- **WHEN** e2e 测试创建 PostgreSQL 连接池并执行多个并发查询
- **THEN** 系统 SHALL 正确管理连接池中的连接
- **AND** 系统 SHALL 复用空闲连接
- **AND** 系统 SHALL 在达到最大连接数时等待可用连接
- **AND** 系统 SHALL 在测试结束后正确关闭所有连接

#### Scenario: e2e 测试 - 查询执行和结果处理
- **WHEN** e2e 测试在真实 PostgreSQL 数据库上执行 SELECT、INSERT、UPDATE、DELETE 查询
- **THEN** 系统 SHALL 正确执行所有查询类型
- **AND** 系统 SHALL 正确返回结果集和受影响行数
- **AND** 系统 SHALL 正确处理 NULL 值
- **AND** 系统 SHALL 正确处理 PostgreSQL 特有数据类型（SERIAL, JSONB, ARRAY）

#### Scenario: e2e 测试 - SSL 模式配置
- **WHEN** e2e 测试配置不同的 SSL 模式（disable, require）
- **THEN** 系统 SHALL 按照配置建立连接
- **AND** 系统 SHALL 在 require 模式下验证 SSL 连接
- **AND** 系统 SHALL 在 SSL 配置错误时返回明确错误

#### Scenario: e2e 测试 - 超时配置
- **WHEN** e2e 测试配置连接超时和查询超时
- **THEN** 系统 SHALL 在超时后取消操作
- **AND** 系统 SHALL 返回超时错误
- **AND** 系统 SHALL 正确清理资源

#### Scenario: e2e 测试 - PostgreSQL 特有语法
- **WHEN** e2e 测试执行 PostgreSQL 特有语法（RETURNING, CTE, 窗口函数）
- **THEN** 系统 SHALL 正确执行并返回结果
- **AND** 系统 SHALL 正确处理 RETURNING 子句返回的数据

#### Scenario: e2e 测试 - 错误处理
- **WHEN** e2e 测试执行各种错误场景（唯一约束冲突、外键约束错误）
- **THEN** 系统 SHALL 返回明确的错误信息
- **AND** 错误信息 SHALL 包含冲突的字段名或约束名
- **AND** 系统 SHALL 不暴露敏感信息
