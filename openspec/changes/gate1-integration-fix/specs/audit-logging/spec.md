## MODIFIED Requirements

### Requirement: 审计日志记录
系统 SHALL 记录所有 SQL 执行操作的审计日志。

#### Scenario: 记录 SELECT 查询
- **WHEN** 用户执行 SELECT 查询
- **THEN** 系统记录审计日志，包含 SQL 语句、用户、实例、执行时间、返回行数

#### Scenario: 记录 INSERT/UPDATE/DELETE 操作
- **WHEN** 用户执行数据修改操作（通过 ExecuteExec）
- **THEN** 系统记录审计日志，包含 SQL 语句、用户、实例、执行时间、影响行数

#### Scenario: ExecuteExec 与 ExecuteQuery 审计格式一致
- **WHEN** 通过 ExecuteExec 执行 INSERT/UPDATE/DELETE
- **THEN** 审计日志格式与 ExecuteQuery 完全一致（user_id, client_ip, statement_type, rows_affected, error）

### Requirement: 日志轮转
系统 SHALL 支持审计日志文件轮转。

#### Scenario: 按天轮转无数据丢失
- **WHEN** 审计日志文件跨天触发轮转
- **THEN** Rotator 先 os.Rename 旧文件，再调用 Writer.Rotate() 刷出缓冲并重新打开新文件
- **AND** 轮转过程中的日志不丢失

#### Scenario: Rotator 优雅停止
- **WHEN** AuditService.Close() 被调用
- **THEN** Rotator.Stop() 等待轮转 goroutine 退出（使用 sync.WaitGroup）
- **AND** 然后才调用 Writer.Close()
