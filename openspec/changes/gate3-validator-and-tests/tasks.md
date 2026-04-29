## 1. SQL Parser POC 验证

- [x] 1.1 引入 `github.com/xwb1989/sqlparser` 依赖，创建 POC 测试文件 `internal/service/validator/sqlparser_poc_test.go`
- [x] 1.2 POC 验证：中文表名（`SELECT * FROM 用户表`）、子查询、多语句分割、SET 语句
- [x] 1.3 POC 验证：PG 特有语法降级（`ON CONFLICT`、`RETURNING` 应解析失败并返回 error）
- [x] 1.4 POC 报告：记录 sqlparser 覆盖度，决定是否继续 AST 方案

## 2. AST 验证器实现

- [x] 2.1 新增 `internal/service/validator/ast_validator.go`：`ASTValidator` 结构体 + `Validate()` + `GetQueryType()` + 降级逻辑
- [x] 2.2 新增 `internal/service/validator/ast_validator_test.go`：覆盖所有 spec 场景（注释绕过、子查询、多语句、降级、默认关闭）
- [x] 2.3 在 `pkg/types/config.go` 的 `ValidatorConfig` 中添加 `UseParser bool` 字段
- [x] 2.4 在 `internal/service/validator/service.go` 中根据 `UseParser` 配置选择验证器实现

## 3. 核心模块测试补充

- [x] 3.1 新增 `internal/service/health/enhanced_checker_test.go`：健康检查流程、状态缓存、事件通知、并发安全
- [x] 3.2 JDBC WebSocket Transport 接口测试已存在（`WebSocketTransportTest.java`，8 个测试覆盖连接/关闭/协议/错误处理）

## 4. 验证

- [x] 4.1 `go build ./...` 编译通过
- [x] 4.2 `make test` 全部测试通过（含新增测试）
- [x] 4.3 AST 验证器性能基准测试：实际场景（~90% SELECT）比率 ~5x，全危险语句 ~8.6x，阈值 10x 通过
