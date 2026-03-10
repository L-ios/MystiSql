## ADDED Requirements

### Requirement: MySQL 连接池 e2e 测试
系统 SHALL 在真实的 MySQL 8 数据库上进行端到端测试，验证连接池的所有核心功能。

#### Scenario: e2e 测试 - 成功建立连接
- **WHEN** e2e 测试使用有效配置连接到 MySQL 8 容器
- **THEN** 系统 SHALL 成功建立连接
- **AND** 连接 SHALL 通过健康检查
- **AND** 连接 SHALL 可以立即执行查询

#### Scenario: e2e 测试 - 连接池管理
- **WHEN** e2e 测试创建连接池并执行多个并发查询
- **THEN** 系统 SHALL 正确管理连接池中的连接
- **AND** 系统 SHALL 复用空闲连接
- **AND** 系统 SHALL 在达到最大连接数时等待可用连接
- **AND** 系统 SHALL 在测试结束后正确关闭所有连接

#### Scenario: e2e 测试 - 查询执行和结果处理
- **WHEN** e2e 测试在真实 MySQL 8 数据库上执行 SELECT、INSERT、UPDATE、DELETE 查询
- **THEN** 系统 SHALL 正确执行所有查询类型
- **AND** 系统 SHALL 正确返回结果集和受影响行数
- **AND** 系统 SHALL 正确处理 NULL 值
- **AND** 系统 SHALL 正确处理各种数据类型（INT, VARCHAR, TEXT, DATETIME, DECIMAL）

#### Scenario: e2e 测试 - 连接重连
- **WHEN** e2e 测试模拟数据库重启或连接中断
- **THEN** 系统 SHALL 检测到连接断开
- **AND** 系统 SHALL 尝试重新建立连接
- **AND** 系统 SHALL 在重连成功后继续执行查询

#### Scenario: e2e 测试 - 健康检查
- **WHEN** e2e 测试调用连接健康检查（Ping）
- **THEN** 系统 SHALL 返回连接状态
- **AND** 系统 SHALL 在连接断开时返回错误
- **AND** 系统 SHALL 使用 `SELECT 1` 验证连接

#### Scenario: e2e 测试 - 超时处理
- **WHEN** e2e 测试执行长时间运行的查询并设置超时
- **THEN** 系统 SHALL 在超时后取消查询
- **AND** 系统 SHALL 返回 context.DeadlineExceeded 错误
- **AND** 系统 SHALL 正确清理资源

#### Scenario: e2e 测试 - 错误处理
- **WHEN** e2e 测试执行各种错误场景（无效 SQL、约束冲突、权限不足）
- **THEN** 系统 SHALL 返回明确的错误信息
- **AND** 错误信息 SHALL 包含足够的上下文
- **AND** 系统 SHALL 不暴露敏感信息（如密码）
