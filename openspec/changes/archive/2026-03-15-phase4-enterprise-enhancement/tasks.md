## Phase 4.1: WebUI 基础版（3 周，25 任务）

### 1.1 项目搭建（1 周）
- [x] 🔄 创建 `web/` 前端项目（Vite + React + TypeScript）
- [x] 配置 Ant Design 5.x
- [x] 配置路由（React Router）
- [x] 配置状态管理（Zustand）
- [x] 配置 API 客户端（Axios + 拦截器）
- [x] 配置构建脚本（Vite）
- [x] 添加 ESLint + Prettier

### 1.2 登录页（0.5 周）
- [x] 实现登录页组件
- [x] 实现 Token 登录（复用现有 API）
- [x] 实现登录状态持久化（localStorage）
- [x] 实现会话管理

### 1.3 实例列表页（0.5 周）
- [x] 实现实例列表组件
- [x] 实现实例状态显示（健康/不健康）
- [x] 实现实例切换

### 1.4 SQL 执行器（1 周）
- [x] 集成 Monaco Editor
- [x] 实现 SQL 语法高亮
- [x] 实现执行按钮 + 加载状态
- [x] 实现错误提示
- [x] 实现 SQL 历史记录（localStorage）
- [x] 实现快捷键（Ctrl+Enter 执行）

### 1.5 结果展示（0.5 周）
- [x] 实现结果表格组件（虚拟滚动）
- [x] 实现分页
- [x] 实现列排序
- [x] 实现导出 CSV/JSON

### 1.6 部署集成（0.5 周）
- [x] Go embed 集成静态资源
- [x] 实现 WebUI 路由（`/`、`/assets/*`）
- [x] 添加配置 `webui.enabled`
- [x] 更新 Dockerfile

---

## Phase 4.2: 读写分离（2 周，12 任务）

### 2.1 SQL 类型识别（0.5 周）
- [x] 实现 SQL 解析器（识别 SELECT/INSERT/UPDATE/DELETE）
- [x] 实现事务检测
- [x] 添加单元测试

### 2.2 路由逻辑（1 周）
- [x] 设计读写分离路由器接口
- [x] 实现主库选择器
- [x] 实现从库选择器（轮询）
- [x] 实现从库延迟检测
- [x] 实现路由逻辑（读→从库，写→主库，事务→主库）
- [x] 添加配置支持
- [x] 添加单元测试

### 2.3 集成测试（0.5 周）
- [x] 添加读写分离集成测试
- [x] 添加故障转移测试

---

## Phase 4.3: OIDC 认证（2 周，10 任务）

### 3.1 OIDC Provider（1.5 周）
- [x] 创建 `internal/service/auth/oidc/` 模块
- [x] 实现 OIDC 配置发现（.well-known）
- [x] 实现 Authorization Code Flow
- [x] 实现 ID Token 验证
- [x] 实现 UserInfo 端点调用
- [x] 实现角色映射（从 Token claims）
- [x] 添加配置支持
- [x] 添加单元测试

### 3.2 API 端点（0.5 周）
- [x] 添加 `GET /api/v1/auth/oidc/login`
- [x] 添加 `GET /api/v1/auth/oidc/callback`

---

## Phase 4.4: RBAC + 数据脱敏（2 周，15 任务）

### 4.1 简化 RBAC（1 周）
- [x] 创建 `internal/service/rbac/` 模块
- [x] 设计权限模型（`instance:database:action`）
- [x] 实现角色管理服务
- [x] 实现权限检查中间件
- [x] 实现 API: `GET/POST/DELETE /api/v1/rbac/roles`
- [x] 实现 API: `GET/POST /api/v1/rbac/users/{id}/roles`
- [x] 添加配置支持
- [x] 添加单元测试

### 4.2 数据脱敏（1 周）
- [x] 创建 `internal/service/masking/` 模块
- [x] 实现手机号脱敏规则
- [x] 实现身份证脱敏规则
- [x] 实现邮箱脱敏规则
- [x] 实现银行卡脱敏规则
- [x] 实现基于角色的脱敏策略
- [x] 添加配置支持

---

## Phase 4.5: Consul 发现（1 周，8 任务，可选）

### 5.1 Consul Discoverer（1 周）
- [x] 创建 `internal/discovery/consul/` 模块
- [x] 实现 ConsulDiscoverer 接口
- [x] 支持服务列表发现
- [x] 支持健康检查过滤
- [x] 支持服务标签解析
- [x] 添加配置支持
- [x] 添加单元测试
- [x] 添加集成测试

---

## 测试

### 单元测试
- [x] 所有新模块添加单元测试
- [x] 测试覆盖率 > 80%

### 集成测试
- [x] 添加 WebUI E2E 测试
- [x] 添加读写分离集成测试
- [x] 添加 OIDC 集成测试（Mock）

---

## 文档

- [x] 更新 README.md（Phase 4 完成）
- [x] 添加 WebUI 使用文档
- [x] 添加读写分离配置文档
- [x] 添加 OIDC 集成文档
- [x] 添加 RBAC 配置文档

---

## 任务统计

| 阶段 | 任务数 | 工期 |
|-----|-------|------|
| Phase 4.1: WebUI | 25 | 3 周 |
| Phase 4.2: 读写分离 | 12 | 2 周 |
| Phase 4.3: OIDC | 10 | 2 周 |
| Phase 4.4: RBAC + 脱敏 | 15 | 2 周 |
| Phase 4.5: Consul（可选）| 8 | 1 周 |
| 测试 + 文档 | 9 | - |
| **总计** | **79** | **10 周** |
