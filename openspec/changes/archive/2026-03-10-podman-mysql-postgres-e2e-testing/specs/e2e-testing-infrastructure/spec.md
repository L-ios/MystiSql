## ADDED Requirements

### Requirement: Podman 容器化的测试环境管理
系统 SHALL 提供 Podman 容器化的 MySQL 8 和 PostgreSQL 测试环境，支持一键启动和停止。

#### Scenario: 启动测试环境
- **WHEN** 执行启动脚本 `scripts/e2e/setup-test-env.sh`
- **THEN** 系统 SHALL 启动 MySQL 8 和 PostgreSQL 容器
- **AND** 系统 SHALL 等待数据库健康检查通过
- **AND** 系统 SHALL 输出连接信息（主机、端口、用户名、密码）

#### Scenario: 停止测试环境
- **WHEN** 执行停止脚本 `scripts/e2e/teardown-test-env.sh`
- **THEN** 系统 SHALL 停止所有测试容器
- **AND** 系统 SHALL 清理容器网络
- **AND** 系统 SHALL 删除临时数据卷（可选）

#### Scenario: 环境状态检查
- **WHEN** 执行环境检查命令
- **THEN** 系统 SHALL 报告 Podman 是否可用
- **AND** 系统 SHALL 报告容器运行状态
- **AND** 系统 SHALL 报告数据库连接状态

### Requirement: 测试数据库自动初始化
系统 SHALL 在测试环境启动时自动初始化测试数据库，包括创建数据库、表结构和测试数据。

#### Scenario: MySQL 数据库初始化
- **WHEN** MySQL 容器首次启动
- **THEN** 系统 SHALL 创建测试数据库 `mystisql_test`
- **AND** 系统 SHALL 执行 `test/e2e/init-mysql.sql` 初始化脚本
- **AND** 系统 SHALL 创建测试表：users, orders, products, audit_log
- **AND** 系统 SHALL 插入初始测试数据

#### Scenario: PostgreSQL 数据库初始化
- **WHEN** PostgreSQL 容器首次启动
- **THEN** 系统 SHALL 创建测试数据库 `mystisql_test`
- **AND** 系统 SHALL 执行 `test/e2e/init-postgres.sql` 初始化脚本
- **AND** 系统 SHALL 创建测试表：users, orders, products, audit_log
- **AND** 系统 SHALL 插入初始测试数据

#### Scenario: 数据库重置
- **WHEN** 执行数据库重置命令
- **THEN** 系统 SHALL 清空所有测试表数据
- **AND** 系统 SHALL 重新执行初始化脚本
- **AND** 系统 SHALL 确认数据一致性

### Requirement: 测试配置管理
系统 SHALL 提供独立的 e2e 测试配置文件，支持灵活的数据库连接配置。

#### Scenario: 加载测试配置
- **WHEN** 运行 e2e 测试
- **THEN** 系统 SHALL 加载 `config/e2e-test.yaml` 配置文件
- **AND** 系统 SHALL 读取 MySQL 和 PostgreSQL 连接参数
- **AND** 系统 SHALL 支持环境变量覆盖配置

#### Scenario: 动态端口分配
- **WHEN** 默认端口被占用
- **THEN** 系统 SHALL 支持通过配置文件指定端口
- **AND** 系统 SHALL 自动检测端口可用性
- **AND** 系统 SHALL 避免端口冲突

### Requirement: Makefile 命令集成
系统 SHALL 提供 Makefile 命令简化测试环境管理。

#### Scenario: 启动测试环境命令
- **WHEN** 执行 `make e2e-setup`
- **THEN** 系统 SHALL 调用 `scripts/e2e/setup-test-env.sh`
- **AND** 系统 SHALL 输出启动日志
- **AND** 系统 SHALL 验证环境就绪

#### Scenario: 运行 e2e 测试命令
- **WHEN** 执行 `make e2e-test`
- **THEN** 系统 SHALL 验证测试环境是否运行
- **AND** 系统 SHALL 执行 `go test ./test/e2e/...`
- **AND** 系统 SHALL 生成测试报告

#### Scenario: 清理测试环境命令
- **WHEN** 执行 `make e2e-teardown`
- **THEN** 系统 SHALL 调用 `scripts/e2e/teardown-test-env.sh`
- **AND** 系统 SHALL 清理所有测试容器

### Requirement: CI/CD 集成支持
系统 SHALL 支持 CI/CD 流程集成，提供可选的 e2e 测试执行。

#### Scenario: CI 环境检测
- **WHEN** 在 CI 环境中运行
- **THEN** 系统 SHALL 检测 CI 环境变量（CI, GITHUB_ACTIONS 等）
- **AND** 系统 SHALL 使用 CI 友好的配置
- **AND** 系统 SHALL 输出结构化日志

#### Scenario: 跳过 e2e 测试
- **WHEN** 使用 `go test -short` 或无 Podman 环境
- **THEN** 系统 SHALL 跳过 e2e 测试
- **AND** 系统 SHALL 输出跳过原因
- **AND** 系统 SHALL 不报错

#### Scenario: 失败诊断
- **WHEN** e2e 测试失败
- **THEN** 系统 SHALL 输出详细错误信息
- **AND** 系统 SHALL 包含数据库连接信息
- **AND** 系统 SHALL 包含失败的 SQL 语句和错误消息
