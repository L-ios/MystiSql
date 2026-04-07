# Playwright E2E 测试

本目录包含 MystiSql WebUI 的端到端测试。

## 测试文件

- `login.spec.ts` - 登录页面测试
- `dashboard.spec.ts` - 仪表盘页面测试
- `instances.spec.ts` - 实例列表页面测试
- `query.spec.ts` - SQL 查询页面测试
- `audit-logs.spec.ts` - 审计日志页面测试

## 运行测试

```bash
# 运行所有测试
npm run test:e2e

# 运行特定测试文件
npx playwright test login.spec.ts

# 运行特定浏览器
npx playwright test --project=chromium

# 运行并查看浏览器
npx playwright test --headed

# 运行并打开 UI 模式
npm run test:e2e:ui

# 运行并调试
npm run test:e2e:debug

# 查看测试报告
npm run test:e2e:report
```

## 测试覆盖

### 登录页面
- ✅ 显示登录表单
- ✅ 表单验证
- ✅ 成功登录
- ✅ 错误处理
- ✅ 认证保护
- ✅ 无障碍访问
- ✅ 网络错误处理

### 仪表盘页面
- ✅ 显示标题
- ✅ 显示统计卡片
- ✅ 显示实例状态概览
- ✅ 加载状态
- ✅ 实例列表显示
- ✅ 健康状态标签
- ✅ 空状态处理
- ✅ API 错误处理

### 实例列表页面
- ✅ 显示页面标题
- ✅ 显示刷新按钮
- ✅ 显示实例表格
- ✅ 加载实例列表
- ✅ 实例详情模态框
- ✅ 刷新实例列表
- ✅ 健康状态徽章
- ✅ API 错误处理
- ✅ 标签显示
- ✅ 空状态处理

### SQL 查询页面
- ✅ 显示页面元素
- ✅ 实例选择验证
- ✅ SQL 输入验证
- ✅ 成功执行查询
- ✅ 查询错误显示
- ✅ 执行时间显示
- ✅ 表格/JSON 视图切换
- ✅ 复制结果
- ✅ 大结果集分页
- ✅ 历史记录按钮
- ✅ 空状态显示

### 审计日志页面
- ✅ 显示页面标题
- ✅ 显示搜索表单
- ✅ 显示审计日志表格
- ✅ 加载审计日志
- ✅ 按用户 ID 搜索
- ✅ 按实例搜索
- ✅ 按 SQL 类型过滤
- ✅ 按敏感操作过滤
- ✅ 高亮敏感操作
- ✅ 查询类型标签颜色
- ✅ 执行时间显示
- ✅ 成功/失败状态
- ✅ 重置搜索表单
- ✅ 分页处理
- ✅ 空状态处理
- ✅ API 错误处理

## 测试策略

1. **隔离性**: 每个测试独立运行，使用 beforeEach 钩子设置初始状态
2. **Mock API**: 使用 Playwright 的 route 功能 mock API 响应
3. **等待策略**: 使用 Playwright 的自动等待机制，避免显式 sleep
4. **断言**: 使用 Playwright 的 expect 断言，支持自动重试
5. **错误处理**: 测试各种错误场景，确保 UI 正确处理错误

## 最佳实践

1. 使用 `data-testid` 属性定位元素（推荐）
2. 使用语义化的选择器（如 `getByRole`, `getByText`）
3. 避免使用 CSS 类名作为选择器
4. 使用 `test.beforeEach` 设置测试前置条件
5. 使用 `test.describe` 组织相关测试
6. 为每个测试添加清晰的描述
7. 使用 Mock API 隔离前端测试
8. 测试用户交互流程，而不是实现细节

## CI/CD 集成

在 CI/CD 环境中运行测试：

```yaml
- name: Install dependencies
  run: cd web && npm ci

- name: Install Playwright browsers
  run: cd web && npx playwright install --with-deps

- name: Run Playwright tests
  run: cd web && npm run test:e2e

- name: Upload test results
  if: always()
  uses: actions/upload-artifact@v3
  with:
    name: playwright-report
    path: web/playwright-report/
```

## 故障排查

### 测试超时
- 增加超时时间：`test.setTimeout(60000)`
- 检查网络请求是否正确 mock
- 检查元素是否正确渲染

### 元素找不到
- 检查选择器是否正确
- 使用 `page.pause()` 调试
- 检查元素是否在 iframe 中

### 测试不稳定
- 使用 `test.slow()` 标记慢测试
- 检查异步操作是否正确等待
- 避免使用固定延迟（`waitForTimeout`）
