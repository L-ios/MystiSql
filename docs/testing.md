# E2E 测试

MystiSql 提供了完整的端到端（E2E）测试环境，使用 Podman 容器化部署 MySQL 8 和 PostgreSQL 14 数据库实例。

## 前置条件

- **Podman**: 需要 Podman 5.0+ 版本
- **Go**: 需要 Go 1.21+ 版本
- **磁盘空间**: 至少 2GB 可用空间（用于数据库镜像）

## 快速开始

### 1. 设置测试环境

```bash
# 启动 MySQL 8 和 PostgreSQL 14 容器
make e2e-setup

# 或使用脚本
./scripts/e2e/setup-test-env.sh
```

输出示例：
```
=== MystiSql E2E Test Environment Setup ===
MySQL Port: 13306
PostgreSQL Port: 15432

Step 3: Starting MySQL 8 container...
Step 4: Starting PostgreSQL 14 container...
Step 5: Waiting for databases to be ready...
MySQL is ready!
PostgreSQL is ready!
Step 6: Initializing test data...

=== Test Environment Ready ===

MySQL Connection:
  Host: localhost
  Port: 13306
  User: root
  Password: test123456
  Database: mystisql_test

PostgreSQL Connection:
  Host: localhost
  Port: 15432
  User: postgres
  Password: test123456
  Database: mystisql_test
```

### 2. 运行 E2E 测试

```bash
# 运行所有 E2E 测试
make e2e-test

# 或直接使用 go test
go test -v -tags=e2e ./test/e2e/...

# 运行特定测试
go test -v -tags=e2e -run TestMySQLBasic ./test/e2e/...
```

### 3. 清理测试环境

```bash
# 停止并删除测试容器
make e2e-teardown

# 或使用脚本
./scripts/e2e/teardown-test-env.sh
```

## 测试环境管理

### 检查环境状态

```bash
make e2e-check

# 或
./scripts/e2e/check-env.sh
```

输出示例：
```
=== MystiSql E2E Test Environment Check ===

1. Checking Podman...
   ✓ Podman installed: podman version 5.7.1

2. Checking container images...
   ✓ MySQL 8 image available
   ✓ PostgreSQL 14 image available

3. Checking container status...
   ✓ MySQL container is running
   ✓ PostgreSQL container is running

5. Checking ports...
   ✓ MySQL port 13306 is listening
   ✓ PostgreSQL port 15432 is listening

=== Summary ===
✓ All checks passed! Environment is ready for e2e testing.
```

### 重置测试数据

```bash
# 重置所有测试数据库
make e2e-reset

# 或重置特定数据库
./scripts/e2e/reset-db.sh mysql      # 仅重置 MySQL
./scripts/e2e/reset-db.sh postgres   # 仅重置 PostgreSQL
```

## 测试内容

E2E 测试覆盖以下核心功能：

### 1. 基础连接测试 (`test/e2e/basic_test.go`)
- MySQL 连接建立和健康检查
- PostgreSQL 连接建立和健康检查
- 基础查询执行（SELECT）

### 2. MySQL 连接池测试
- 连接池管理（并发连接、连接复用）
- 查询执行（SELECT, INSERT, UPDATE, DELETE）
- 结果处理（NULL 值、各种数据类型）
- 超时处理
- 错误处理（无效 SQL、约束冲突）

### 3. PostgreSQL 驱动测试
- 连接和查询执行
- PostgreSQL 特有功能（RETURNING, CTE, JSONB, 数组类型）
- SSL 模式配置
- 错误处理（唯一约束、外键约束）

### 4. 事务管理测试
- 事务提交和回滚
- 事务隔离性（READ COMMITTED）
- 并发事务
- 事务超时

### 5. 批量操作测试
- 批量插入（100 条记录）
- 批量更新和删除
- 混合批处理
- 性能对比测试

## 测试配置

E2E 测试配置文件位于 `config/e2e-test.yaml`：

```yaml
# MySQL 测试实例
instances:
  - name: "test-mysql"
    type: "mysql"
    host: "127.0.0.1"
    port: 13306
    username: "root"
    password: "test123456"
    database: "mystisql_test"

  - name: "test-postgres"
    type: "postgresql"
    host: "127.0.0.1"
    port: 15432
    username: "postgres"
    password: "test123456"
    database: "mystisql_test"
```

可以通过环境变量覆盖配置：

```bash
export MYSQL_HOST=127.0.0.1
export MYSQL_PORT=13306
export POSTGRES_HOST=127.0.0.1
export POSTGRES_PORT=15432
```

## 测试数据

测试数据初始化脚本：

- `test/e2e/init-mysql.sql` - MySQL 测试数据（10 用户、12 产品、10 订单）
- `test/e2e/init-postgres.sql` - PostgreSQL 测试数据（含 JSONB、数组类型）

## 自定义端口

如果默认端口被占用，可以通过环境变量修改：

```bash
export MYSQL_PORT=23306
export POSTGRES_PORT=25432

./scripts/e2e/setup-test-env.sh
```

## CI/CD 集成

在 CI/CD 流程中集成 E2E 测试：

```yaml
# GitHub Actions 示例
name: E2E Tests

on: [push, pull_request]

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
      
      - name: Cleanup
        if: always()
        run: make e2e-teardown
```

## 跳过 E2E 测试

在短模式下运行测试会跳过 E2E 测试：

```bash
# 跳过 E2E 测试
go test -short ./...

# 或者不添加 -tags=e2e
go test ./test/e2e/...  # 会被跳过
```

## 故障排查

### 容器启动失败

```bash
# 检查 Podman 状态
podman info

# 查看容器日志
podman logs mystisql-test-mysql
podman logs mystisql-test-postgres

# 检查端口占用
netstat -tuln | grep 13306
netstat -tuln | grep 15432
```

### 镜像拉取失败

如果遇到镜像拉取问题，配置 Podman 镜像加速器：

```bash
# 编辑 ~/.config/containers/registries.conf
cat > ~/.config/containers/registries.conf <<EOF
unqualified-search-registries = ["docker.io"]

[[registry]]
prefix = "docker.io"
location = "docker.io"

[[registry.mirror]]
location = "docker.1panel.live"
EOF
```

### 数据库连接失败

```bash
# 检查容器状态
podman ps

# 测试数据库连接
podman exec mystisql-test-mysql mysql -uroot -ptest123456 -e "SELECT 1"
podman exec mystisql-test-postgres psql -U postgres -c "SELECT 1"
```

## Makefile 命令

完整的 E2E 测试相关命令：

```bash
make e2e-check          # 检查环境状态
make e2e-setup          # 启动测试环境
make e2e-test           # 运行 E2E 测试
make e2e-test-coverage  # 运行测试并生成覆盖率报告
make e2e-teardown       # 清理测试环境
make e2e-reset          # 重置测试数据库
make e2e-reset-mysql    # 仅重置 MySQL
make e2e-reset-postgres # 仅重置 PostgreSQL
```

## 性能基准测试

运行性能基准测试：

```bash
# 运行批量插入性能测试
go test -v -tags=e2e -bench=BenchmarkBatchInsert -benchmem ./test/e2e/...
```
