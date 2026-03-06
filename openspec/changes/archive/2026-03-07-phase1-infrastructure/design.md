## Context（背景）

MystiSql 是一个数据库访问网关项目，目前处于早期开发阶段（Phase 1）。

**当前状态：**
- ✅ 项目目录结构已创建（cmd、internal、pkg、config、test）
- ✅ README.md 定义了完整的架构和路线图
- ✅ AGENTS.md 提供了开发规范和指南
- ❌ 还没有任何实现代码
- ❌ 没有测试代码

**约束条件：**
- 必须使用 Go 语言（项目已初始化为 Go 项目）
- 必须遵循 AGENTS.md 中的编码规范
- 必须遵循 Go 最佳实践（Effective Go）
- Phase 1 必须保持最小化和专注
- 暂不使用 K8s client-go（Phase 2 才引入）

**技术栈：**
- 语言：Go 1.21+
- 架构：分层架构（数据库层 → 发现层 → 连接层 → 服务层 → 接入层）

## Goals / Non-Goals（目标 / 非目标）

### Goals（本次设计要达成的目标）

**核心目标：**
1. ✅ 实现静态配置发现机制（从 YAML 文件读取数据库实例）
2. ✅ 集成 MySQL 驱动，实现基本的连接和查询功能
3. ✅ 搭建 CLI 框架，提供基础的命令行工具
4. ✅ 搭建 REST API 框架，提供基础的 HTTP 接口
5. ✅ 建立配置管理模式，支持 YAML 配置和环境变量

**质量目标：**
- 代码简洁、可读性强
- 遵循 Go 最佳实践
- 为后续阶段打好基础（接口设计要考虑扩展性）
- 基本的错误处理和日志记录

### Non-Goals（明确不在本次范围内）

**不在 Phase 1 范围：**
- ❌ K8s API 动态发现（client-go 集成）
- ❌ 连接池管理（database/sql 基础连接即可）
- ❌ PostgreSQL、Oracle、Redis 支持
- ❌ 认证和授权机制
- ❌ WebSocket 支持
- ❌ WebUI 界面
- ❌ 读写分离
- ❌ SQL 解析和智能路由
- ❌ 审计日志
- ❌ 数据脱敏

## Decisions（关键技术决策）

### 决策 1：配置文件格式

**选择：** YAML + Viper

**为什么选择 YAML：**
- ✅ 人类可读性强，支持注释
- ✅ 支持复杂嵌套结构
- ✅ Kubernetes 生态系统的标准格式（ConfigMap）
- ✅ 容易映射到 Go 结构体

**为什么选择 Viper：**
- ✅ Go 生态中最流行的配置管理库
- ✅ 支持多种格式（YAML、JSON、TOML）
- ✅ 支持环境变量覆盖
- ✅ 支持热重载
- ✅ 与 Cobra 无缝集成

**考虑过的替代方案：**
- ❌ JSON：不支持注释，可读性差
- ❌ TOML：在 Kubernetes 生态中不太常见
- ❌ 纯环境变量：难以管理复杂配置

**配置结构示例：**
```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: release

discovery:
  type: static

instances:
  - name: production-mysql
    type: mysql
    host: mysql.production.svc.cluster.local
    port: 3306
    username: root
    password: secret
    database: myapp
```

---

### 决策 2：CLI 框架选择

**选择：** Cobra + Viper 组合

**为什么选择 Cobra：**
- ✅ Go 生态中 CLI 工具的事实标准
- ✅ 被广泛使用（Docker、Kubernetes、Hugo、Etcd 等）
- ✅ 自动生成帮助文档
- ✅ 支持子命令和嵌套命令
- ✅ 支持命令自动补全
- ✅ 与 Viper 无缝集成

**CLI 命令结构设计：**
```
mystisql
├── version              # 显示版本信息
├── instances            # 实例管理
│   └── list            # 列出所有实例
└── query <instance> <sql>  # 执行查询
    --format (table|json|csv)  # 输出格式
    --config <file>           # 指定配置文件
    --verbose                 # 详细日志
```

**考虑过的替代方案：**
- ❌ urfave/cli：功能较少，社区较小
- ❌ 自己实现 flag 解析：工作量大，缺少标准功能

---

### 决策 3：REST API 框架选择

**选择：** Gin

**为什么选择 Gin：**
- ✅ 高性能（最快的 Go HTTP 路由器之一）
- ✅ 轻量级，API 简洁
- ✅ 内置 JSON 验证和绑定
- ✅ 中间件支持完善
- ✅ 活跃的社区和完善的文档
- ✅ README.md 中推荐的框架

**API 端点设计：**
```
GET  /health                  # 健康检查
GET  /api/v1/instances        # 列出所有实例
POST /api/v1/query            # 执行查询
  Body: {
    "instance": "production-mysql",
    "sql": "SELECT * FROM users LIMIT 10"
  }
```

**考虑过的替代方案：**
- ❌ Fiber：性能更好但不够成熟
- ❌ Echo：不错但社区比 Gin 小
- ❌ net/http 标准库：太底层，需要更多样板代码

---

### 决策 4：实例发现架构设计

**选择：** 基于接口的设计，先实现静态发现

**为什么选择接口设计：**
- ✅ 定义清晰的契约（InstanceDiscoverer 接口）
- ✅ 静态实现简单，适合 Phase 1
- ✅ 容易在 Phase 2 添加 K8s/Consul 发现器
- ✅ 便于测试（可以使用 Mock）
- ✅ 遵循 Go 接口最佳实践

**接口定义：**
```go
type InstanceDiscoverer interface {
    // Name 返回发现器的名称
    Name() string
    
    // Discover 发现并返回数据库实例列表
    Discover(ctx context.Context) ([]*DatabaseInstance, error)
}
```

**实例注册中心（InstanceRegistry）：**
```go
type InstanceRegistry interface {
    // Register 注册一个实例
    Register(instance *DatabaseInstance) error
    
    // GetInstance 根据名称获取实例
    GetInstance(name string) (*DatabaseInstance, error)
    
    // ListInstances 列出所有实例
    ListInstances() ([]*DatabaseInstance, error)
}
```

**实现策略：**
- Phase 1：实现 `StaticDiscoverer`（从配置文件读取）
- Phase 2：添加 `K8sDiscoverer`（监听 K8s API）
- Phase 4：添加 `ConsulDiscoverer`（服务注册中心）

---

### 决策 5：数据库连接管理

**选择：** 使用 database/sql 标准库 + go-sql-driver/mysql，不使用连接池（Phase 2）

**为什么选择 database/sql：**
- ✅ Go 标准库，稳定可靠
- ✅ 提供基本的连接管理
- ✅ 支持连接池（可以在 Phase 2 启用）
- ✅ 所有数据库驱动都遵循这个接口

**为什么 Phase 1 不使用连接池：**
- ✅ 保持 Phase 1 简单
- ✅ 专注核心功能验证
- ✅ Phase 2 再优化性能

**连接接口设计：**
```go
type Connection interface {
    // Connect 建立连接
    Connect(ctx context.Context) error
    
    // Query 执行查询
    Query(ctx context.Context, sql string) (*QueryResult, error)
    
    // Ping 检查连接健康
    Ping(ctx context.Context) error
    
    // Close 关闭连接
    Close() error
}
```

**考虑过的替代方案：**
- ❌ GORM 等 ORM：太重，Phase 1 只需要基本查询
- ❌ sqlx：功能不错但 database/sql 已足够
- ❌ 直接使用驱动 API：不标准，database/sql 是标准接口

---

### 决策 6：错误处理策略

**选择：** 定义哨兵错误（Sentinel Errors）+ 结构化错误包装

**为什么选择这种方式：**
- ✅ 遵循 Go 1.13+ 的错误处理最佳实践
- ✅ 支持使用 `errors.Is()` 检查错误类型
- ✅ 使用 `fmt.Errorf("上下文信息: %w", err)` 包装错误
- ✅ 为每种组件定义明确的错误类型

**错误定义模式：**
```go
// pkg/errors/errors.go
var (
    // 发现相关错误
    ErrInstanceNotFound      = errors.New("实例未找到")
    ErrInstanceAlreadyExists = errors.New("实例已存在")
    ErrDiscoveryFailed       = errors.New("发现失败")
    
    // 连接相关错误
    ErrConnectionFailed = errors.New("连接失败")
    ErrConnectionClosed = errors.New("连接已关闭")
    ErrQueryFailed      = errors.New("查询失败")
    
    // 配置相关错误
    ErrConfigNotFound   = errors.New("配置文件未找到")
    ErrConfigInvalid    = errors.New("配置文件无效")
)
```

**错误处理示例：**
```go
func (d *StaticDiscoverer) Discover(ctx context.Context) ([]*DatabaseInstance, error) {
    instances, err := d.loadFromConfig()
    if err != nil {
        return nil, fmt.Errorf("从配置加载实例失败: %w", err)
    }
    
    if len(instances) == 0 {
        return nil, fmt.Errorf("%w: 配置文件中没有定义任何实例", ErrDiscoveryFailed)
    }
    
    return instances, nil
}
```

---

### 决策 7：日志记录策略

**选择：** 使用结构化日志库 Zap

**为什么选择 Zap：**
- ✅ 高性能（比标准库 log 快很多）
- ✅ 结构化日志（JSON 格式）
- ✅ 支持日志级别（Debug、Info、Warn、Error）
- ✅ AGENTS.md 中推荐使用

**日志使用规范：**
- ❌ 不记录敏感信息（密码、token 等）
- ✅ 记录关键操作（连接、查询、错误）
- ✅ 使用结构化字段（key-value）
- ✅ 不同环境使用不同日志级别

**示例：**
```go
logger.Info("成功连接到数据库实例",
    zap.String("instance", instanceName),
    zap.String("host", host),
    zap.Int("port", port),
    // 注意：不记录 username 和 password
)
```

---

### 决策 8：项目结构组织

**选择：** 遵循 README.md 和 AGENTS.md 定义的结构

**目录结构：**
```
MystiSql/
├── cmd/mystisql/              # CLI 入口
│   └── main.go
├── internal/                  # 私有代码
│   ├── config/               # 配置加载
│   │   └── loader.go
│   ├── discovery/            # 服务发现
│   │   ├── types.go         # 接口定义
│   │   ├── registry.go      # 实例注册中心
│   │   └── static/          # 静态发现实现
│   │       └── discoverer.go
│   ├── connection/           # 数据库连接
│   │   ├── types.go         # 接口定义
│   │   └── mysql/           # MySQL 实现
│   │       └── connection.go
│   ├── cli/                  # CLI 命令
│   │   ├── root.go
│   │   ├── query.go
│   │   └── instances.go
│   └── api/rest/             # REST API
│       ├── server.go
│       └── handlers.go
├── pkg/                      # 公共库
│   ├── types/               # 核心类型
│   │   ├── instance.go
│   │   └── config.go
│   └── errors/              # 错误定义
│       └── errors.go
├── config/                   # 配置文件
│   └── config.yaml
└── test/                     # 集成测试
```

**设计原则：**
- `internal/` 放私有实现代码
- `pkg/` 放可复用的公共代码
- 按功能分层组织（discovery、connection、cli、api）
- 每个包职责清晰，高内聚低耦合

## Risks / Trade-offs（风险与权衡）

### 风险 1：静态发现不适用于动态环境

**风险描述：**
静态配置发现需要手动更新配置文件，在 K8s 环境中实例经常变化时不方便。

**缓解措施：**
- ✅ 这是 Phase 1 的有意设计，保持简单
- ✅ Phase 2 会实现 K8s 动态发现
- ✅ 提供 SIGHUP 信号支持配置热重载

---

### 风险 2：无连接池可能导致性能问题

**风险描述：**
每次查询都可能创建新连接，频繁连接/断开会影响性能。

**缓解措施：**
- ✅ Phase 1 专注功能验证，性能优化在 Phase 2
- ✅ database/sql 内部有基本连接管理
- ✅ 文档中明确说明这是 Phase 1 的限制

---

### 风险 3：无认证机制存在安全隐患

**风险描述：**
任何人都可以访问和查询数据库，没有权限控制。

**缓解措施：**
- ✅ 文档中明确说明仅用于开发/测试环境
- ✅ 建议通过网络安全策略限制访问
- ✅ Phase 3 会实现完整的认证授权机制

**权衡（Trade-off）：**
- 🔹 牺牲安全性 → 换取快速实现和验证核心功能
- 🔹 牺牲性能 → 换取代码简洁和易于理解

---

### 风险 4：只支持 MySQL 限制了适用范围

**风险描述：**
Phase 1 只支持 MySQL，用户无法访问 PostgreSQL、Oracle、Redis。

**缓解措施：**
- ✅ 接口设计考虑了多数据库支持
- ✅ Phase 3 会添加 PostgreSQL 支持
- ✅ Phase 4 会添加 Oracle 和 Redis 支持

---

### 风险 5：缺少单元测试可能影响代码质量

**风险描述：**
Phase 1 任务中没有明确要求写单元测试，可能影响代码质量。

**缓解措施：**
- ✅ 代码会遵循 Go 最佳实践
- ✅ 会在实施过程中添加基础单元测试
- ✅ 使用 table-driven tests 提高测试覆盖率
- ✅ Phase 2 会完善测试体系

## Migration Plan（迁移计划）

不适用 - 这是初始实现，没有需要迁移的旧系统。

## Open Questions（待解决问题）

**暂无待解决问题。**

如果在实施过程中发现需要澄清的技术决策，会更新此文档。
