# MystiSql JDBC E2E 测试指南

## 概述

本文档介绍如何运行 JDBC E2E 测试，这些测试会连接到真实的 MystiSql Gateway 服务。

## 前置条件

1. **MystiSql Gateway 服务已启动**
   ```bash
   # 启动 Gateway 服务
   ./bin/mystisql --config config.yaml
   ```

2. **MySQL 实例已配置并运行**
   ```bash
   # 确保 MySQL 实例已配置在 config.yaml 中
   # 例如：
   # instances:
   #   - name: local-mysql
   #     type: mysql
   #     host: localhost
   #     port: 3306
   #     username: root
   #     password: root
   #     database: test
   ```

3. **JDBC 驱动已构建**
   ```bash
   cd jdbc
   ./gradlew build
   ```

## 配置环境变量

### 必需的环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `GATEWAY_HOST` | Gateway 服务主机地址 | `localhost` |
| `GATEWAY_PORT` | Gateway 服务端口 | `8080` |
| `INSTANCE_NAME` | 要连接的数据库实例名称 | `local-mysql` |
| `AUTH_TOKEN` | 认证令牌（如果 Gateway 启用了认证） | 空 |

### 配置示例

```bash
# 基本配置（无认证）
export GATEWAY_HOST=localhost
export GATEWAY_PORT=8080
export INSTANCE_NAME=local-mysql

# 带认证的配置
export GATEWAY_HOST=localhost
export GATEWAY_PORT=8080
export INSTANCE_NAME=local-mysql
export AUTH_TOKEN=your-jwt-token-here
```

## 运行 E2E 测试

### 运行所有 E2E 测试

```bash
cd jdbc
./gradlew e2e
```

### 运行特定测试类

```bash
# 运行连接测试
./gradlew e2e --tests "io.github.mystisql.jdbc.e2e.JdbcE2EConnectionTest"

# 运行查询测试
./gradlew e2e --tests "io.github.mystisql.jdbc.e2e.JdbcE2EQueryTest"

# 运行 PreparedStatement 测试
./gradlew e2e --tests "io.github.mystisql.jdbc.e2e.JdbcE2EPreparedStatementTest"
```

### 运行特定测试方法

```bash
# 运行单个测试方法
./gradlew e2e --tests "io.github.mystisql.jdbc.e2e.JdbcE2EConnectionTest.testConnectToGateway"
```

## 测试覆盖范围

### 1. 连接测试 (JdbcE2EConnectionTest)

- ✅ 连接到 Gateway 并验证连接
- ✅ 获取连接元数据
- ✅ 测试自动提交模式
- ✅ 创建 Statement 对象
- ✅ 创建 PreparedStatement 对象
- ✅ 测试连接关闭
- ✅ 测试无效连接参数
- ✅ 测试连接超时设置

### 2. 查询测试 (JdbcE2EQueryTest)

- ✅ 执行简单 SELECT 查询
- ✅ 执行 SELECT 1 查询
- ✅ 执行多列 SELECT 查询
- ✅ 测试 ResultSet 元数据
- ✅ 测试空结果集
- ✅ 测试 NULL 值处理
- ✅ 测试不同数据类型
- ✅ 测试 executeUpdate 方法
- ✅ 测试 execute 方法
- ✅ 测试查询超时

### 3. PreparedStatement 测试 (JdbcE2EPreparedStatementTest)

- ✅ PreparedStatement 基本查询
- ✅ 设置字符串参数
- ✅ 设置整数参数
- ✅ 设置 NULL 参数
- ✅ 清除参数
- ✅ 设置不同类型的参数
- ✅ PreparedStatement executeUpdate
- ✅ PreparedStatement 查询数据
- ✅ PreparedStatement 更新数据
- ✅ PreparedStatement 删除数据

## 测试结果

### 成功示例

```
JDBC E2E Test Configuration:
  Gateway: localhost:8080
  Instance: local-mysql
  JDBC URL: jdbc:mystisql://localhost:8080/local-mysql

✅ 成功连接到 MystiSql Gateway
✅ 数据库产品: MystiSql Gateway
✅ 数据库版本: 1.0.0
✅ 驱动名称: MystiSql JDBC Driver
✅ 驱动版本: 1.1.0
✅ 自动提交模式切换正常
✅ Statement 创建和关闭正常
✅ PreparedStatement 创建和关闭正常
✅ 连接关闭正常
✅ 无效连接参数正确抛出异常
✅ 连接超时设置正常
```

### 失败排查

#### 连接失败

**错误信息**: `Connection refused` 或 `Timeout`

**解决方案**:
1. 检查 Gateway 服务是否运行
2. 检查端口是否正确
3. 检查防火墙设置

#### 认证失败

**错误信息**: `401 Unauthorized`

**解决方案**:
1. 检查 AUTH_TOKEN 环境变量是否设置
2. 检查 token 是否有效
3. 检查 token 是否过期

#### 实例不存在

**错误信息**: `Instance not found`

**解决方案**:
1. 检查 INSTANCE_NAME 环境变量是否正确
2. 检查 config.yaml 中是否配置了该实例
3. 检查实例名称拼写是否正确

## CI/CD 集成

### GitHub Actions 示例

```yaml
name: JDBC E2E Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: root
          MYSQL_DATABASE: test
        ports:
          - 3306:3306
        options: >-
          --health-cmd="mysqladmin ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=5
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up JDK 11
      uses: actions/setup-java@v3
      with:
        java-version: '11'
        distribution: 'temurin'
    
    - name: Build JDBC Driver
      run: |
        cd jdbc
        ./gradlew build
    
    - name: Start MystiSql Gateway
      run: |
        ./bin/mystisql --config config.yaml &
        sleep 5
    
    - name: Run E2E Tests
      env:
        GATEWAY_HOST: localhost
        GATEWAY_PORT: 8080
        INSTANCE_NAME: local-mysql
      run: |
        cd jdbc
        ./gradlew e2e
```

## 最佳实践

### 1. 测试隔离

每个测试方法都是独立的，会创建新的连接和资源，测试结束后自动清理。

### 2. 资源管理

使用 try-with-resources 确保所有资源（Connection, Statement, ResultSet）都能正确关闭。

### 3. 错误处理

测试会验证各种错误场景，包括：
- 无效连接参数
- SQL 语法错误
- 超时场景
- NULL 值处理

### 4. 数据清理

测试中创建的临时表会在测试结束后自动清理。

## 故障排查

### 启用详细日志

```bash
# 启用 JDBC 驱动日志
export JAVA_OPTS="-Djava.util.logging.config.file=logging.properties"

# 运行测试
./gradlew e2e --info
```

### 检查 Gateway 日志

```bash
# 查看 Gateway 日志
tail -f /var/log/mystisql/gateway.log
```

### 检查网络连接

```bash
# 测试 Gateway 连接
curl http://localhost:8080/health

# 测试 API 端点
curl http://localhost:8080/api/v1/instances
```

## 总结

JDBC E2E 测试提供了完整的端到端测试覆盖，确保 JDBC 驱动能够正确连接到 MystiSql Gateway 并执行各种数据库操作。通过这些测试，可以验证：

1. ✅ 连接管理
2. ✅ 查询执行
3. ✅ 参数化查询
4. ✅ 数据类型处理
5. ✅ 错误处理
6. ✅ 资源管理

运行这些测试是验证 JDBC 驱动功能的重要步骤。
