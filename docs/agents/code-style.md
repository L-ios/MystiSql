# Code Style Guidelines

## 核心原则

- 遵循 [Effective Go](https://golang.org/doc/effective_go) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 函数 < 50 行，文件 < 500 行，优先组合而非继承
- Imports 分三组：标准库 / 第三方 / 本地包，组间空行分隔
- 命名：包名小写单词，CamelCase 变量/函数，缩写全大写（`HTTPServer` 非 `HttpServer`），接口用名词或动词

## 错误处理

- 所有 error 必须显式处理并附加上下文（`fmt.Errorf("xxx: %w", err)`）
- 定义哨兵错误（`ErrInstanceNotFound` 等）

## Context 使用

- `context.Context` 始终作为第一个参数
- 不存储 context 到 struct
- 用 context 控制超时和取消

## 测试

- 测试文件 `xxx_test.go`，函数名 `TestXxx`
- AAA 模式（Arrange, Act, Assert），优先 table-driven tests

## 接口与类型

- 优先用接口做抽象，类型别名增强可读性（`type InstanceID string`）

## 文档

- 导出函数/类型必须有文档注释，注释 **why** 而非 what
- 使用结构化日志（zap）

## 并发

- goroutine 必须有受控并发（semaphore/channel 限制数量）
- goroutine 内用 `defer` 清理资源
- 闭包中正确捕获循环变量
- 不生成无界 goroutine

---

# Security & Best Practices

- 不记录凭证、密码或敏感数据
- 使用参数化查询防止 SQL 注入
- 验证所有用户输入
- 日志前脱敏处理
- 使用 TLS 加密网络通信
- 使用连接池，`defer` 关闭连接
- 实现查询超时和结果集大小限制
- 大结果集使用流式处理

---

# Development Workflow

1. **开始前**: 检查 README.md 了解当前阶段和路线图
2. **编码**: 遵循代码风格，保持函数精简
3. **测试**: 编写单元测试，追求高覆盖率
4. **检查**: 提交前运行 `make lint`
5. **文档**: 更新导出函数的注释
6. **审查**: 确保编译通过、测试通过、无 linter 警告

---

# Pull Request Checklist

- [ ] 代码编译无错误
- [ ] 所有测试通过 (`make test`)
- [ ] 新代码有测试
- [ ] 遵循代码风格
- [ ] 无 linter 警告 (`make lint`)
- [ ] 文档已更新
- [ ] 无敏感数据泄露
- [ ] Breaking changes 已记录
