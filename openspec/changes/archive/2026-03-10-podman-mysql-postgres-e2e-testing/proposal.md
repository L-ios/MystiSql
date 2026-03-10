## Why

当前项目缺乏端到端（e2e）测试环境，无法在真实的 MySQL 8 和 PostgreSQL 数据库上验证连接、查询、事务等功能。开发人员需要快速搭建本地测试环境，以验证 Phase 3 的认证、审计、SQL 验证等特性在真实数据库场景下的行为。

## What Changes

- 使用 Podman 容器化部署 MySQL 8 和 PostgreSQL 数据库实例
- 创建自动化脚本启动和管理测试数据库环境
- 为 MySQL 和 PostgreSQL 连接池编写 e2e 测试用例
- 测试 Phase 3 功能：认证中间件、审计日志、SQL 验证器、事务管理、批量操作
- 提供 Makefile 命令简化测试环境管理
- 创建测试配置文件和初始化 SQL 脚本

## Capabilities

### New Capabilities

- `e2e-testing-infrastructure`: Podman 容器化的 MySQL 8 和 PostgreSQL 测试环境，包含自动化启动/停止脚本、测试数据库初始化、连接配置管理

### Modified Capabilities

- `mysql-connection`: 增强 MySQL 连接池的 e2e 测试覆盖
- `postgresql-driver`: 增强 PostgreSQL 驱动的 e2e 测试覆盖
- `jdbc-transaction`: 在真实数据库上验证事务管理功能
- `jdbc-batch-operations`: 在真实数据库上验证批量操作功能

## Impact

- **新增文件**:
  - `scripts/e2e/setup-test-env.sh`: 启动测试数据库环境
  - `scripts/e2e/teardown-test-env.sh`: 停止测试数据库环境
  - `test/e2e/mysql_e2e_test.go`: MySQL e2e 测试用例
  - `test/e2e/postgres_e2e_test.go`: PostgreSQL e2e 测试用例
  - `test/e2e/transaction_e2e_test.go`: 事务 e2e 测试用例
  - `test/e2e/batch_e2e_test.go`: 批量操作 e2e 测试用例
  - `test/e2e/init-mysql.sql`: MySQL 初始化脚本
  - `test/e2e/init-postgres.sql`: PostgreSQL 初始化脚本
  - `config/e2e-test.yaml`: e2e 测试配置文件

- **依赖变更**: 
  - 需要 Podman 运行环境
  - 需要 MySQL 8.x 容器镜像
  - 需要 PostgreSQL 14+ 容器镜像

- **测试流程**: 
  - CI/CD 流程可选执行 e2e 测试（需要 Podman 环境）
  - 本地开发可通过 `make e2e-test` 运行
