## ADDED Requirements

### Requirement: 事务管理 e2e 测试
系统 SHALL 在真实数据库上进行端到端测试，验证事务管理功能的数据一致性和隔离性。

#### Scenario: e2e 测试 - 事务提交
- **WHEN** e2e 测试在事务中执行多个 INSERT 操作并提交
- **THEN** 系统 SHALL 成功提交事务
- **AND** 所有 INSERT 操作的数据 SHALL 持久化到数据库
- **AND** 后续查询 SHALL 能看到提交的数据

#### Scenario: e2e 测试 - 事务回滚
- **WHEN** e2e 测试在事务中执行多个 INSERT 操作后回滚
- **THEN** 系统 SHALL 成功回滚事务
- **AND** 所有 INSERT 操作的数据 SHALL 不持久化
- **AND** 数据库 SHALL 恢复到事务开始前的状态

#### Scenario: e2e 测试 - 事务隔离性
- **WHEN** e2e 测试启动两个并发事务，事务 A 插入数据但未提交，事务 B 查询同一张表
- **THEN** 事务 B SHALL 看不到事务 A 未提交的数据（READ COMMITTED 隔离级别）
- **AND** 事务 A 提交后，事务 B 的后续查询 SHALL 能看到数据

#### Scenario: e2e 测试 - 事务超时
- **WHEN** e2e 测试启动一个长时间运行的事务并等待超时
- **THEN** 系统 SHALL 在超时后自动回滚事务
- **AND** 系统 SHALL 释放数据库连接
- **AND** 后续使用该 connectionId SHALL 返回错误

#### Scenario: e2e 测试 - 并发事务
- **WHEN** e2e 测试同时启动多个并发事务并在每个事务中执行操作
- **THEN** 每个事务 SHALL 独立执行
- **AND** 事务之间 SHALL 不相互干扰
- **AND** 所有事务 SHALL 能正确提交或回滚

#### Scenario: e2e 测试 - 事务中执行混合操作
- **WHEN** e2e 测试在事务中执行 SELECT、INSERT、UPDATE、DELETE 混合操作
- **THEN** 系统 SHALL 在同一连接上按顺序执行所有操作
- **AND** 提交后所有修改 SHALL 生效
- **AND** 回滚后所有修改 SHALL 撤销

#### Scenario: e2e 测试 - 事务错误处理
- **WHEN** e2e 测试在事务中执行违反约束的 SQL
- **THEN** 系统 SHALL 返回错误
- **AND** 事务 SHALL 保持在打开状态
- **AND** 可以选择提交或回滚事务

#### Scenario: e2e 测试 - 跨数据库事务
- **WHEN** e2e 测试在 MySQL 和 PostgreSQL 上分别启动事务
- **THEN** 每个数据库实例的事务 SHALL 独立管理
- **AND** 不同数据库的事务 SHALL 不相互影响
- **AND** 每个 connectionId SHALL 绑定到对应的数据库连接
