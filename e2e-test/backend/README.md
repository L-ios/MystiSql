# MystiSql E2E 测试文档

本文档介绍如何使用和扩展 MystiSql 的端到端（E2E）测试环境。

## 目录

- [概述](#概述)
- [测试架构](#测试架构)
- [快速开始](#快速开始)
- [测试环境详解](#测试环境详解)
- [如何添加新测试](#如何添加新测试)
- [调试失败的测试](#调试失败的测试)
- [CI/CD 集成](#cicd-集成)

## 概述

E2E 测试环境使用 Podman 容器化部署 MySQL 8 和 PostgreSQL 14 数据库， 提供真实的数据库环境进行集成测试。

### 测试覆盖范围

- **MySQL 连接池测试**：连接、查询、连接池管理、 超时处理
- **PostgreSQL 驱动测试**：连接、查询、特有功能（RETURNING, CTE, JSONB）
- **事务管理测试**：提交、回滚、隔离性、 并发事务
- **批量操作测试**：批量插入、更新、删除
 性能测试

### 测试工具

- **测试框架**：Go 标准测试框架 + testify
- **容器技术**：Podman（兼容 Docker）
- **数据库驱动**：
  - MySQL: github.com/go-sql-driver/mysql
  - PostgreSQL: github.com/lib/pq

## 测试架构

```
test/e2e/
├── config.go              # 测试配置加载
├── helper.go              # 测试辅助函数
├── fixture.go             # 测试数据生成器
├── basic_test.go          # 基础连接测试
├── mysql_e2e_test.go.bak   # MySQL 测试（备份）
├── postgres_e2e_test.go.bak # PostgreSQL 测试（备份）
├── transaction_e2e_test.go.bak # 事务测试（备份）
├── batch_e2e_test.go.bak     # 批量操作测试（备份）
├── init-mysql.sql         # MySQL 初始化脚本
├── init-postgres.sql      # PostgreSQL 初始化脚本
└── README.md              # 本文档
```

## 快速开始

### 1. 前置条件

- Podman 已安装
- Go 1.21+
- Make（可选）

### 2. 启动测试环境

```bash
# 方法 1: 使用 Makefile（推荐）
make e2e-setup

# 方法 2: 直接使用脚本
./scripts/e2e/setup-test-env.sh
```

### 3. 运行测试

```bash
# 运行所有 E2E 测试
make e2e-test

# 或直接使用 go test
go test -v -tags=e2e ./test/e2e/...

# 运行特定测试
go test -v -tags=e2e -run TestMySQLBasic ./test/e2e/...

# 运行并生成覆盖率报告
make e2e-test-coverage
```

### 4. 清理环境

```bash
# 停止并删除测试容器
make e2e-teardown

# 可选：清理数据卷
CLEAN_VOLUMES=true make e2e-teardown
```

## 测试环境详解

### 容器配置

| 数据库 | 容器名称 | 端口 | 用户名 | 密码 | 数据库 |
|--------|----------|------|--------|------|--------|
| MySQL 8 | mystisql-test-mysql | 13306 | root | test123456 | mystisql_test |
| PostgreSQL 14 | mystisql-test-postgres | 15432 | postgres | test123456 | mystisql_test |

### 测试数据

测试环境初始化时会创建以下测试表：

#### MySQL 表

- `users` - 用户信息
- `products` - 产品信息
- `orders` - 订单信息
- `audit_log` - 审计日志

#### PostgreSQL 表

- `users` - 用户信息
- `products` - 产品信息（含数组字段）
- `orders` - 订单信息
- `audit_log` - 审计日志

每个表至少插入 10 条测试数据，### 环境变量

可通过环境变量覆盖默认配置：

```bash
# MySQL 配置
export MYSQL_HOST=127.0.0.1
export MYSQL_PORT=13306
export MYSQL_USER=root
export MYSQL_PASSWORD=test123456
export MYSQL_DATABASE=mystisql_test

# PostgreSQL 配置
export POSTGRES_HOST=127.0.0.1
export POSTGRES_PORT=15432
export POSTGRES_USER=postgres
export POSTGRES_PASSWORD=test123456
export POSTGRES_DATABASE=mystisql_test
```

## 如何添加新测试

### 1. 创建测试文件

在 `test/e2e/` 目录下创建新的测试文件， 文件名格式：`<功能>_e2e_test.go`

```go
//go:build e2e

package e2e

import (
    "context"
    "database/sql"
    "testing"
    "time"

    _ "github.com/go-sql-driver/mysql"
    _ "github.com/lib/pq"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestYourFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping in short mode")
    }

    // 加载配置
    config, err := LoadConfig()
    require.NoError(t, err, "Failed to load config")

    // 建立连接
    db, err := sql.Open("mysql", config.MySQL.DSN())
    require.NoError(t, err, "Failed to connect to MySQL")
    defer db.Close()

    // 编写测试逻辑
    // ...
}
```

### 2. 测试模板

**基础测试模板**：

```go
func TestBasicExample(t *testing.T) {
    SkipIfShort(t)
    
    config, err := LoadConfig()
    require.NoError(t, err)
    
    db := NewMySQLConnection(t, &config.MySQL)
    
    // 测试逻辑
    ctx := context.Background()
    var result int
    err = db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
    assert.NoError(t, err)
    assert.Equal(t, 1, result)
}
```

**事务测试模板**：

```go
func TestTransaction(t *testing.T) {
    SkipIfShort(t)
    
    config, err := LoadConfig()
    require.NoError(t, err)
    
    db := NewMySQLConnection(t, &config.MySQL)
    
    ctx := context.Background()
    tx, err := db.BeginTx(ctx, nil)
    require.NoError(t, err)
    defer tx.Rollback()
    
    // 事务操作
    // ...
    
    err = tx.Commit()
    require.NoError(t, err)
}
```

### 3. 测试最佳实践

1. **使用 SkipIfShort**：在短测试模式下跳过 E2E 测试
2. **使用 helper 函数**：如 `NewMySQLConnection`, `CleanupTable`
3. **清理测试数据**：测试完成后清理插入的数据
4. **使用 fixture**：使用 `GenerateTestUser` 等函数生成测试数据
5. **添加注释**：说明测试的目的和步骤

## 调试失败的测试

### 1. 检查环境状态

```bash
# 检查测试环境是否就绪
make e2e-check

# 或使用脚本
./scripts/e2e/check-env.sh
```

### 2. 查看容器日志

```bash
# 查看 MySQL 容器日志
podman logs mystisql-test-mysql

# 查看 PostgreSQL 容器日志
podman logs mystisql-test-postgres
```

### 3. 手动测试数据库连接

```bash
# 测试 MySQL 连接
podman exec -it mystisql-test-mysql \
  mysql -uroot -ptest123456 mystisql_test -e "SELECT * FROM users LIMIT 5"

# 测试 PostgreSQL 连接
podman exec -it mystisql-test-postgres \
  psql -U postgres -d mystisql_test -c "SELECT * FROM users LIMIT 5"
```

### 4. 重置测试数据

```bash
# 重置所有数据库
make e2e-reset

# 只重置 MySQL
make e2e-reset-mysql

# 只重置 PostgreSQL
make e2e-reset-postgres
```

### 5. 常见问题

**问题 1: 端口被占用**
```bash
# 检查端口占用
netstat -tuln | grep 13306
netstat -tuln | grep 15432

# 解决方法：停止占用端口的进程或修改配置
```

**问题 2: 容器无法启动**
```bash
# 查看容器状态
podman ps -a

# 删除并重新创建
podman rm -f mystisql-test-mysql mystisql-test-postgres
make e2e-setup
```

**问题 3: 数据库连接失败**
```bash
# 检查容器网络
podman network ls

# 重启容器
podman restart mystisql-test-mysql mystisql-test-postgres
```

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: E2E Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  e2e-test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Install Podman
      run: |
        sudo apt-get update
        sudo apt-get install -y podman
    
    - name: Setup E2E Environment
      run: make e2e-setup
    
    - name: Run E2E Tests
      run: make e2e-test
    
    - name: Upload Coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./e2e-coverage.out
    
    - name: Cleanup
      if: always()
      run: make e2e-teardown
```

### GitLab CI 示例

```yaml
e2e-test:
  image: golang:1.21
  services:
    - name: mysql:8
      alias: mysql
    - name: postgres:14
      alias: postgres
  variables:
    MYSQL_ROOT_PASSWORD: test123456
    MYSQL_DATABASE: mystisql_test
    POSTGRES_PASSWORD: test123456
    POSTGRES_DB: mystisql_test
  before_script:
    - apt-get update && apt-get install -y podman
    - make e2e-setup
  script:
    - make e2e-test
  after_script:
    - make e2e-teardown
```

### Jenkins Pipeline 示例

```groovy
pipeline {
    agent any
    
    stages {
        stage('Setup') {
            steps {
                sh 'make e2e-setup'
            }
        }
        
        stage('Test') {
            steps {
                sh 'make e2e-test'
            }
            post {
                always {
                    junit 'test-results/*.xml'
                }
            }
        }
    }
    
    post {
        always {
            sh 'make e2e-teardown'
        }
    }
}
```

### 跳过 E2E 测试

在 CI/CD 中可选跳过 E2E 测试：

```bash
# 使用 -short 标志
go test -short ./...

# 或使用环境变量
export SKIP_E2E=true
```

## 参考资料

- [Podman 官方文档](https://podman.io/docs)
- [Go 测试最佳实践](https://golang.org/doc/tutorial/add-a-test)
- [testify 文档](https://github.com/stretchr/testify)
