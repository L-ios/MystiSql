# MystiSql E2E 测试目录

本目录集中管理所有端到端（E2E）测试，支持本地开发和 GitHub Actions 自动化测试。

## 📁 目录结构

```
e2e-test/
├── backend/          # 后端 E2E 测试（Go）
│   ├── auth_e2e_test.go
│   ├── config.go
│   ├── helper.go
│   ├── fixture.go
│   └── README.md
├── frontend/         # 前端 E2E 测试（Playwright）
│   ├── login.spec.ts
│   ├── dashboard.spec.ts
│   ├── instances.spec.ts
│   ├── query.spec.ts
│   ├── audit-logs.spec.ts
│   └── README.md
├── jdbc/             # JDBC E2E 测试（Java）
│   ├── JdbcE2EConnectionTest.java
│   ├── JdbcE2EPreparedStatementTest.java
│   └── JdbcE2EQueryTest.java
└── README.md         # 本文档
```

## 🚀 快速开始

### 方式一：使用统一测试脚本（推荐）

```bash
# 运行所有 E2E 测试
make e2e-run

# 只运行前端测试
make e2e-run-frontend

# 只运行后端测试
make e2e-run-backend

# 只运行 JDBC 测试
make e2e-run-jdbc

# 查看测试报告
make e2e-report
```

### 方式二：使用单独测试命令

```bash
# 后端测试
cd e2e-test/backend
go test -v -tags=e2e ./...

# 前端测试
cd web
npm run test:e2e

# JDBC 测试
cd jdbc
mvn clean test
```

## 📊 测试报告

### 本地测试报告

运行测试后，报告会自动生成在 `test-reports/` 目录：

```
test-reports/
├── index.html              # 统一报告入口
├── frontend/               # 前端测试报告
│   └── index.html
├── backend/                # 后端测试报告
│   └── coverage.html
├── jdbc/                   # JDBC 测试报告
│   └── surefire-reports/
└── e2e-test-report.tar.gz  # 压缩包（用于 GitHub Actions）
```

打开 `test-reports/index.html` 查看完整的测试报告。

### GitHub Actions 测试报告

GitHub Actions 会自动运行 E2E 测试，并上传测试报告：

1. 进入 GitHub Actions 页面
2. 选择对应的 workflow run
3. 在 Artifacts 部分下载 `e2e-test-reports`
4. 解压后打开 `index.html` 查看报告

**注意**：无论测试成功或失败，测试报告都会上传。

## 🔧 测试类型详解

### 后端 E2E 测试（Go）

**位置**：`e2e-test/backend/`

**测试内容**：
- Token 生成与验证
- Token 撤销机制
- 健康检查 API
- 认证中间件

**运行方式**：
```bash
# 使用统一脚本
make e2e-run-backend

# 或直接运行
cd e2e-test/backend
go test -v -tags=e2e -cover ./...
```

**详细文档**：查看 [backend/README.md](backend/README.md)

### 前端 E2E 测试（Playwright）

**位置**：`e2e-test/frontend/`

**测试内容**：
- 登录页面（表单验证、认证流程）
- 仪表盘（统计卡片、实例状态）
- 实例管理（实例列表、健康检查）
- SQL 查询（查询执行、结果展示）
- 审计日志（日志查询、过滤）

**运行方式**：
```bash
# 使用统一脚本
make e2e-run-frontend

# 或直接运行
cd web
npm run test:e2e
```

**详细文档**：查看 [frontend/README.md](frontend/README.md)

### JDBC E2E 测试（Java）

**位置**：`e2e-test/jdbc/`

**测试内容**：
- JDBC 连接管理
- SQL 查询执行
- PreparedStatement 测试
- 事务管理

**运行方式**：
```bash
# 使用统一脚本
make e2e-run-jdbc

# 或使用 Maven
cd jdbc
mvn clean test
```

## 🔄 GitHub Actions 集成

### 自动触发

GitHub Actions 会在以下情况自动运行 E2E 测试：

- Push 到 `main` 或 `develop` 分支
- Pull Request 到 `main` 或 `develop` 分支

### 手动触发

在 GitHub Actions 页面，可以手动触发测试：

1. 进入 Actions 页面
2. 选择 "E2E Tests" workflow
3. 点击 "Run workflow"
4. 选择测试类型（all/frontend/backend/jdbc）
5. 点击 "Run workflow"

### Workflow 配置

Workflow 配置文件：`.github/workflows/e2e-tests.yml`

**关键特性**：
- ✅ 自动设置 Go、Node.js、Java 环境
- ✅ 自动安装 Playwright 浏览器
- ✅ 自动构建 Gateway 服务
- ✅ 运行测试并生成报告
- ✅ 压缩测试报告为 tar.gz
- ✅ 上传报告（无论成功或失败）
- ✅ 测试失败时 workflow 标记为失败

### 测试报告下载

1. 进入 GitHub Actions 页面
2. 选择对应的 workflow run
3. 在页面底部找到 "Artifacts" 部分
4. 下载 `e2e-test-reports`
5. 解压后查看 `index.html`

## 🛠️ 开发指南

### 添加新的后端测试

1. 在 `e2e-test/backend/` 创建新测试文件
2. 文件名格式：`<功能>_e2e_test.go`
3. 添加 `//go:build e2e` 标签
4. 使用 helper 函数简化测试代码

```go
//go:build e2e

package e2e

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestNewFeature(t *testing.T) {
    SkipIfShort(t)
    
    config, err := LoadConfig()
    require.NoError(t, err)
    
    // 测试逻辑
}
```

### 添加新的前端测试

1. 在 `e2e-test/frontend/` 创建新测试文件
2. 文件名格式：`<功能>.spec.ts`
3. 使用 Playwright API 编写测试

```typescript
import { test, expect } from '@playwright/test';

test.describe('新功能测试', () => {
  test('测试用例', async ({ page }) => {
    await page.goto('/');
    // 测试逻辑
  });
});
```

### 添加新的 JDBC 测试

1. 在 `e2e-test/jdbc/` 创建新测试文件
2. 文件名格式：`JdbcE2E<功能>Test.java`
3. 使用 JUnit 5 编写测试

```java
package io.github.mystisql.jdbc.e2e;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

public class JdbcE2ENewFeatureTest {
    @Test
    void testNewFeature() throws SQLException {
        // 测试逻辑
    }
}
```

## 📝 最佳实践

### 1. 测试隔离

- 每个测试应该独立运行
- 不依赖其他测试的结果
- 清理测试数据

### 2. Mock vs 真实环境

- **前端测试**：使用 Mock API 隔离测试
- **后端测试**：使用真实数据库连接
- **JDBC 测试**：使用真实 Gateway 服务

### 3. 测试数据管理

- 使用 fixture 生成测试数据
- 测试完成后清理数据
- 避免硬编码测试数据

### 4. 错误处理

- 测试各种错误场景
- 验证错误消息
- 确保错误处理逻辑正确

### 5. 性能考虑

- 避免不必要的等待
- 使用并行测试（如果可能）
- 控制测试数据量

## 🔍 故障排查

### 后端测试失败

```bash
# 检查 Gateway 服务是否启动
curl http://localhost:8080/health

# 查看服务日志
tail -f /tmp/gateway.log

# 检查配置文件
cat config/config.yaml
```

### 前端测试失败

```bash
# 查看 Playwright 报告
cd web && npm run test:e2e:report

# 调试模式运行
cd web && npm run test:e2e:debug

# UI 模式运行
cd web && npm run test:e2e:ui
```

### JDBC 测试失败

```bash
# 检查 Java 版本
java -version  # 应该是 21+

# 查看 Maven 测试报告
cat jdbc/target/surefire-reports/*.txt

# 检查 Gateway 服务
curl http://localhost:8080/health
```

### GitHub Actions 失败

1. 查看 workflow 日志
2. 下载测试报告 artifact
3. 检查错误消息
4. 本地重现问题

## 📚 相关文档

- [后端测试详细文档](backend/README.md)
- [前端测试详细文档](frontend/README.md)
- [E2E 测试指南](../E2E_TESTING_GUIDE.md)
- [项目规则](../.trae/rules/project_rules.md)

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支
3. 添加测试用例
4. 确保所有测试通过
5. 提交 Pull Request

## 📄 许可证

本项目采用 Apache 2.0 许可证。详见 [LICENSE](../LICENSE) 文件。
