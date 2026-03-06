## 1. 基础设施 - 类型和错误定义

- [x] 1.1 定义核心类型 pkg/types/instance.go（DatabaseInstance、DatabaseType、InstanceStatus）
- [x] 1.2 定义核心类型 pkg/types/config.go（Config、ServerConfig、DiscoveryConfig、InstanceConfig）
- [x] 1.3 定义核心类型 pkg/types/result.go（QueryResult、ColumnInfo、Row）
- [x] 1.4 定义哨兵错误 pkg/errors/errors.go（所有错误常量）
- [x] 1.5 更新 go.mod 添加依赖（gin、cobra、viper、go-sql-driver/mysql、zap）

## 2. 配置管理

- [x] 2.1 创建配置文件示例 config/config.yaml（包含实例配置示例）
- [x] 2.2 实现配置加载 internal/config/loader.go（Load、Validate 方法）
- [x] 2.3 添加配置验证逻辑（必填字段检查、格式验证）
- [x] 2.4 支持多配置文件路径查找（./config.yaml、./config/config.yaml、/etc/mystisql/config.yaml）
- [x] 2.5 支持环境变量覆盖配置（Viper 自动支持）
- [x] 2.6 测试配置加载功能（创建 loader_test.go）

## 3. 服务发现（静态）

- [x] 3.1 定义 InstanceDiscoverer 接口 internal/discovery/types.go
- [x] 3.2 定义 InstanceRegistry 接口 internal/discovery/types.go
- [x] 3.3 实现 InstanceRegistry internal/discovery/registry.go（Register、GetInstance、ListInstances 方法）
- [x] 3.4 实现 StaticDiscoverer internal/discovery/static/discoverer.go（Name、Discover 方法）
- [x] 3.5 连接 StaticDiscoverer 到配置加载器
- [x] 3.6 测试实例注册和发现功能（创建 registry_test.go、discoverer_test.go）

## 4. MySQL 连接

- [x] 4.1 定义 Connection 接口 internal/connection/types.go
- [x] 4.2 实现 MySQLConnection internal/connection/mysql/connection.go
- [x] 4.3 实现 Connect() 方法（支持超时、错误处理）
- [x] 4.4 实现 Query() 方法（支持 context、结果集处理）
- [x] 4.5 实现 Exec() 方法（INSERT/UPDATE/DELETE）
- [x] 4.6 实现 Ping() 方法（连接健康检查）
- [x] 4.7 实现 Close() 方法（资源清理）
- [x] 4.8 添加连接字符串构建逻辑（从 InstanceConfig 构建 Dsn）
- [x] 4.9 测试 MySQL 连接功能（创建 connection_test.go）

## 5. CLI 基础框架

- [x] 5.1 创建 CLI 入口 cmd/mystisql/main.go（初始化 Cobra、Viper、zap logger）
- [x] 5.2 实现 root 命令 internal/cli/root.go（全局标志、配置加载）
- [x] 5.3 实现 version 命令 internal/cli/version.go（显示版本信息）
- [x] 5.4 初始化日志系统（支持 verbose 标志）
- [x] 5.5 测试 CLI 基础命令（创建 version_test.go）

## 6. CLI - Instances 命令

- [x] 6.1 实现 instances list 命令 internal/cli/instances.go
- [x] 6.2 支持表格格式输出（默认）
- [x] 6.3 支持 JSON 格式输出（--format json）
- [x] 6.4 支持 CSV 格式输出（--format csv）
- [x] 6.5 实现 instances get 命令（获取单个实例详情）
- [x] 6.6 测试 instances 命令（创建 instances_test.go）

## 7. CLI - Query 命令

- [x] 7.1 实现 query 命令 internal/cli/query.go
- [x] 7.2 集成 MySQL 连接执行查询
- [x] 7.3 实现表格格式输出（默认）
- [x] 7.4 实现 JSON 格式输出
- [x] 7.5 实现 CSV 格式输出
- [x] 7.6 添加 verbose 日志支持
- [x] 7.7 处理错误和退出状态码
- [x] 7.8 测试 query 命令（创建 query_test.go）

## 8. REST API - 服务器设置

- [x] 8.1 创建 API 服务器 internal/api/rest/server.go（Gin 初始化、路由设置）
- [x] 8.2 实现优雅关闭逻辑（SIGTERM/SIGINT 处理）
- [x] 8.3 添加 CORS 中间件
- [x] 8.4 添加日志中间件（记录请求信息）
- [x] 8.5 添加错误恢复中间件
- [x] 8.6 配置 Gin 模式（debug/release）
- [x] 8.7 测试服务器启动和关闭（创建 server_test.go）

## 9. REST API - 健康检查端点

- [x] 9.1 实现 GET /health 端点 internal/api/rest/handlers.go
- [x] 9.2 实现健康检查响应格式
- [x] 9.3 实现 check-instances 参数支持
- [x] 9.4 测试健康检查端点（创建 handlers_test.go）

## 10. REST API - Instances 端点

- [x] 10.1 实现 GET /api/v1/instances 端点
- [x] 10.2 实现实例列表响应格式
- [x] 10.3 实现密码脱敏逻辑
- [x] 10.4 测试 instances 端点（包含在 handlers_test.go）

## 11. REST API - Query 端点

- [x] 11.1 实现 POST /api/v1/query 端点
- [x] 11.2 定义请求和响应结构（QueryRequest、QueryResponse）
- [x] 11.3 集成 MySQL 连接执行查询
- [x] 11.4 实现 context 超时控制
- [x] 11.5 实现错误响应格式
- [x] 11.6 测试 query 端点（包含在 handlers_test.go）

## 12. 集成和端到端测试

- [x] 12.1 在 main.go 中组装所有组件（配置、发现、连接、CLI、API）
- [x] 12.2 测试完整流程：配置加载 → 实例发现 → 连接建立 → 查询执行
- [x] 12.3 测试 CLI 完整工作流（mystisql instances list、mystisql query）
- [x] 12.4 测试 API 完整工作流（health、instances、query 端点）
- [x] 12.5 测试错误处理和边界情况（创建 complete_flow_test.go）

## 13. 代码质量和文档

- [x] 13.1 运行 go fmt ./... 格式化代码
- [x] 13.2 运行 go vet ./... 检查代码
- [x] 13.3 运行 golangci-lint run 并修复问题
- [x] 13.4 为所有导出函数和类型添加文档注释
- [x] 13.5 检查代码是否遵循 AGENTS.md 规范
- [x] 13.6 确保 Context 传递正确
- [x] 13.7 确保错误处理包含上下文信息
- [x] 13.8 确保不记录敏感信息（密码、token）

## 14. 最终验证

- [x] 14.1 更新 README.md 添加 Phase 1 使用示例
- [x] 14.2 创建详细的配置文件注释示例
- [x] 14.3 手动测试：编译成功（go build）
- [x] 14.4 手动测试：CLI 命令正常工作
- [x] 14.5 手动测试：API 端点正常工作
- [x] 14.6 手动测试：能够连接真实 MySQL 数据库
- [x] 14.7 手动测试：查询执行正确
- [x] 14.8 验证所有 Phase 1 目标已完成
