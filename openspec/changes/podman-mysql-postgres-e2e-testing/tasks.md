## 1. 环境检查和准备

- [x] 1.1 验证 Podman 是否安装并可用（podman --version）
- [x] 1.2 拉取 MySQL 8 官方镜像（podman pull mysql:8）
- [x] 1.3 拉取 PostgreSQL 14 官方镜像（podman pull postgres:14）
- [x] 1.4 创建测试所需的目录结构（scripts/e2e/, test/e2e/, config/）

## 2. Podman 容器管理脚本

- [x] 2.1 创建 scripts/e2e/setup-test-env.sh 启动脚本
  - 启动 MySQL 8 容器（端口映射、环境变量、volume mount）
  - 启动 PostgreSQL 14 容器（端口映射、环境变量、volume mount）
  - 等待数据库健康检查通过
  - 输出连接信息
- [x] 2.2 创建 scripts/e2e/teardown-test-env.sh 停止脚本
  - 停止所有测试容器
  - 清理容器网络
  - 可选清理数据卷
- [x] 2.3 创建 scripts/e2e/check-env.sh 环境检查脚本
  - 检查 Podman 是否可用
  - 检查容器运行状态
  - 检查数据库连接状态
- [x] 2.4 创建 scripts/e2e/reset-db.sh 数据库重置脚本
  - 清空测试数据库
  - 重新执行初始化脚本

## 3. 数据库初始化脚本

- [x] 3.1 创建 test/e2e/init-mysql.sql
  - 创建测试数据库 mystisql_test
  - 创建测试表：users, orders, products, audit_log
  - 插入初始测试数据（至少 10 条记录/表）
- [x] 3.2 创建 test/e2e/init-postgres.sql
  - 创建测试数据库 mystisql_test
  - 创建测试表：users, orders, products, audit_log
  - 插入初始测试数据（至少 10 条记录/表）
  - 使用 PostgreSQL 特有类型（SERIAL, JSONB 等）

## 4. 测试配置文件

- [x] 4.1 创建 config/e2e-test.yaml
  - MySQL 连接配置（host, port, username, password, database）
  - PostgreSQL 连接配置（host, port, username, password, database, sslmode）
  - 连接池配置（maxOpen, maxIdle, maxLifetime）
  - 支持环境变量覆盖
- [x] 4.2 创建测试配置加载工具 test/e2e/config.go
  - 从 config/e2e-test.yaml 加载配置
  - 支持环境变量覆盖
  - 提供默认值

## 5. MySQL 连接池 e2e 测试

- [x] 5.1 创建 test/e2e/mysql_e2e_test.go
  - 测试连接建立和健康检查
  - 测试连接池管理（并发连接、连接复用）
  - 测试查询执行（SELECT, INSERT, UPDATE, DELETE）
  - 测试结果处理（NULL 值、各种数据类型）
  - 测试连接重连（模拟数据库重启）
  - 测试超时处理
  - 测试错误处理（无效 SQL、约束冲突）
- [x] 5.2 使用 build tag 标记为 e2e 测试（//go:build e2e）
- [x] 5.3 添加测试前置条件检查（检查环境是否就绪）
- [x] 5.4 添加测试后清理逻辑

## 6. PostgreSQL 驱动 e2e 测试

- [x] 6.1 创建 test/e2e/postgres_e2e_test.go
  - 测试连接建立和健康检查
  - 测试连接池管理
  - 测试查询执行（SELECT, INSERT, UPDATE, DELETE）
  - 测试 PostgreSQL 特有功能（RETURNING, CTE, 窗口函数）
  - 测试 SSL 模式配置
  - 测试超时配置
  - 测试错误处理（唯一约束、外键约束）
- [x] 6.2 使用 build tag 标记为 e2e 测试（//go:build e2e）
- [x] 6.3 添加测试前置条件检查
- [x] 6.4 添加测试后清理逻辑

## 7. 事务管理 e2e 测试

- [x] 7.1 创建 test/e2e/transaction_e2e_test.go
  - 测试事务提交（验证数据持久化）
  - 测试事务回滚（验证数据撤销）
  - 测试事务隔离性（READ COMMITTED）
  - 测试事务超时自动回滚
  - 测试并发事务（独立执行、互不干扰）
  - 测试事务中执行混合操作（SELECT, INSERT, UPDATE, DELETE）
  - 测试事务错误处理
  - 测试跨数据库事务（MySQL 和 PostgreSQL）
- [x] 7.2 使用 build tag 标记为 e2e 测试
- [x] 7.3 添加并发测试用例

## 8. 批量操作 e2e 测试

- [x] 8.1 创建 test/e2e/batch_e2e_test.go
  - 测试批量插入（100 条记录）
  - 测试批量更新（50 条记录）
  - 测试批量删除（30 条记录）
  - 测试混合批处理（INSERT, UPDATE, DELETE）
  - 测试批量操作性能（对比逐条执行）
  - 测试批量操作错误处理（部分失败）
  - 测试批量操作大小限制（1000 条）
  - 测试事务中的批量操作
  - 测试大批量数据插入（1000 条记录）
- [x] 8.2 使用 build tag 标记为 e2e 测试
- [x] 8.3 添加性能基准测试（Benchmark 函数）

## 10. 测试辅助工具

- [x] 10.1 创建 test/e2e/helper.go
  - 提供测试数据库连接获取函数
  - 提供测试数据清理函数
  - 提供测试断言辅助函数
  - 提供测试日志记录函数
- [x] 10.2 创建 test/e2e/fixture.go
  - 提供测试数据 fixture
  - 提供测试数据生成函数

## 9. Makefile 集成

- [x] 9.1 添加 e2e-setup 目标
  - 调用 scripts/e2e/setup-test-env.sh
  - 验证环境就绪
- [x] 9.2 添加 e2e-test 目标
  - 验证测试环境是否运行
  - 执行 go test -tags=e2e ./test/e2e/...
  - 生成测试报告
- [x] 9.3 添加 e2e-teardown 目标
  - 调用 scripts/e2e/teardown-test-env.sh
  - 清理测试环境
- [x] 9.4 添加 e2e-reset 目标
  - 调用 scripts/e2e/reset-db.sh
  - 重置测试数据库

## 11. 文档和示例

- [x] 11.1 更新 README.md
  - 添加 e2e 测试章节
  - 说明前置条件（Podman 安装）
  - 说明如何运行 e2e 测试
- [x] 11.2 创建 test/e2e/README.md
  - 详细说明 e2e 测试架构
  - 说明如何添加新的 e2e 测试
  - 说明如何调试失败的测试
  - 提供 CI/CD 集成示例
- [x] 11.3 更新 AGENTS.md
  - 添加 e2e 测试相关指南
  - 说明 e2e 测试执行命令

## 12. 验证和优化

- [x] 12.1 运行所有 e2e 测试并确保通过
  - make e2e-test
  - 修复所有失败的测试
- [x] 12.2 验证测试覆盖率
  - 确保核心功能都有 e2e 测试覆盖
  - 补充缺失的测试用例
- [x] 12.3 优化测试执行时间
  - 并行化独立的测试用例
  - 减少不必要的等待时间
- [x] 12.4 优化容器启动时间
  - 使用轻量级镜像
  - 优化初始化脚本
