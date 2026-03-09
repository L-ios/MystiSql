## ADDED Requirements

### Requirement: 批量操作 e2e 测试
系统 SHALL 在真实数据库上进行端到端测试，验证批量操作的性能和正确性。

#### Scenario: e2e 测试 - 批量插入
- **WHEN** e2e 测试执行批量 INSERT 100 条数据
- **THEN** 系统 SHALL 成功插入所有数据
- **AND** 返回的影响行数 SHALL 等于 100
- **AND** 数据库 SHALL 包含所有插入的记录

#### Scenario: e2e 测试 - 批量更新
- **WHEN** e2e 测试执行批量 UPDATE 50 条数据
- **THEN** 系统 SHALL 成功更新所有符合条件的记录
- **AND** 返回的影响行数 SHALL 等于实际更新的行数
- **AND** 数据库 SHALL 反映更新后的值

#### Scenario: e2e 测试 - 批量删除
- **WHEN** e2e 测试执行批量 DELETE 30 条数据
- **THEN** 系统 SHALL 成功删除所有符合条件的记录
- **AND** 返回的影响行数 SHALL 等于实际删除的行数
- **AND** 数据库 SHALL 不包含被删除的记录

#### Scenario: e2e 测试 - 混合批处理
- **WHEN** e2e 测试执行混合批处理（INSERT 20 条，UPDATE 10 条，DELETE 5 条）
- **THEN** 系统 SHALL 按顺序执行所有操作
- **AND** 返回的结果 SHALL 包含每个操作的影响行数
- **AND** 数据库 SHALL 反映所有修改

#### Scenario: e2e 测试 - 批量操作性能
- **WHEN** e2e 测试比较批量插入 100 条与逐条插入 100 条的执行时间
- **THEN** 批量插入的执行时间 SHALL 显著低于逐条插入
- **AND** 性能提升 SHALL 至少 50%

#### Scenario: e2e 测试 - 批量操作错误处理
- **WHEN** e2e 测试执行批量操作，其中第 3 条 SQL 违反约束
- **THEN** 系统 SHALL 返回详细结果，标记第 3 条失败
- **AND** 成功的 SQL SHALL 执行完成
- **AND** 失败的 SQL SHALL 返回错误信息

#### Scenario: e2e 测试 - 批量操作大小限制
- **WHEN** e2e 测试执行超过最大批量大小（1000 条）的批处理
- **THEN** 系统 SHALL 返回错误
- **AND** 错误信息 SHALL 说明批量大小超过限制

#### Scenario: e2e 测试 - 事务中的批量操作
- **WHEN** e2e 测试在事务中执行批量操作
- **THEN** 所有操作 SHALL 在同一事务中执行
- **AND** 提交后所有修改 SHALL 生效
- **AND** 回滚后所有修改 SHALL 撤销

#### Scenario: e2e 测试 - 大批量数据插入
- **WHEN** e2e 测试执行大批量数据插入（1000 条记录）
- **THEN** 系统 SHALL 成功插入所有数据
- **AND** 执行时间 SHALL 在合理范围内（< 5 秒）
- **AND** 内存使用 SHALL 保持稳定
