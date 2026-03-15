## Context

MystiSql 是一个数据库访问网关，已完成 Phase 1-3：
- **Phase 1**: 基础设施（MySQL、静态发现、REST API）
- **Phase 2**: 核心引擎（K8s 发现、连接池、JDBC 驱动）
- **Phase 3**: 安全控制（Token 认证、审计日志、PostgreSQL、WebSocket）

**Phase 4 目标：企业级生产就绪，优先用户体验。**

### 当前架构现状

```
┌─────────────────────────────────────────────────────────────┐
│                     Access Layer                             │
│  REST API │ WebSocket │ JDBC Driver │ CLI (REPL)            │
├─────────────────────────────────────────────────────────────┤
│                     Service Layer                            │
│  Query Engine │ Auth (JWT) │ Audit │ Validator │ Cache      │
│  Connection Pool Monitor (已实现)                            │
├─────────────────────────────────────────────────────────────┤
│                   Connection Layer                           │
│  MySQL Pool │ PostgreSQL Pool                               │
├─────────────────────────────────────────────────────────────┤
│                    Discovery Layer                           │
│  Static Config │ K8s API (已支持 Watch)                      │
├─────────────────────────────────────────────────────────────┤
│                   Database Layer                             │
│  MySQL │ PostgreSQL                                         │
└─────────────────────────────────────────────────────────────┘
```

### 设计原则

1. **用户体验优先**: WebUI 第一时间交付用户价值
2. **功能最小化**: 只做企业必需功能，拒绝功能蔓延
3. **分期交付**: 5 个子阶段，每个可独立上线
4. **复用优先**: 最大化利用已有代码（如 Token 认证）

## Goals / Non-Goals

**Goals (聚焦 6 个核心)**:
1. ✅ **WebUI 基础版** - SQL 执行 + 结果展示（Phase 4.1 优先）
2. ✅ **读写分离** - 高频需求，MySQL 主从架构标配
3. ✅ **OIDC 认证** - 企业标准，Keycloak/Dex 集成
4. ✅ **简化 RBAC** - 库级别权限控制（够用就好）
5. ✅ **数据脱敏** - 合规必需（手机号/身份证）
6. ✅ **Consul 发现** - 补充 K8s 之外的发现方式

**Non-Goals (明确排除)**:
- ❌ Oracle 驱动 → Phase 5（需求少）
- ❌ Nacos 发现 → 只实现 Consul（二选一）
- ❌ LDAP 认证 → OIDC 覆盖大部分场景（LDAP 可通过 Dex 代理）
- ❌ 复杂 RBAC (Casbin) → 简化版足够
- ❌ SQL 注入检测 → 效果存疑，建议 WAF 层处理
- ❌ 多集群管理 → 增加复杂度，推迟
- ❌ K8s ServiceAccount 认证 → OIDC 可替代
- ❌ WebUI 高级功能 → Phase 5
- ❌ 敏感字段自动识别 → 合并到数据脱敏配置

## Decisions

### 1. WebUI 技术选型（Phase 4.1 优先）

**决策**: React 18 + TypeScript + Ant Design 5 + Vite

**理由**:
- React 生态成熟，组件丰富
- TypeScript 类型安全，减少运行时错误
- Ant Design 企业级 UI 组件库，开箱即用
- Vite 快速构建，开发体验好

**页面清单**:
| 页面 | 功能 | 复用 |
|-----|------|------|
| 登录 | Token 登录 | 复用现有 API |
| 实例列表 | 显示实例状态 | GET /api/v1/instances |
| SQL 执行 | 编辑器 + 执行 | POST /api/v1/query |
| 审计日志 | 日志查询 | GET /api/v1/audit/logs |

**部署方式**: 嵌入静态资源（Go embed）

**配置**:
```yaml
webui:
  enabled: true
  mode: embedded  # embedded | external
```

---

### 2. 读写分离（Phase 4.2）

**决策**: 基于标签的路由策略

**配置**:
```yaml
instances:
  - name: mysql-master
    type: mysql
    host: mysql-master.svc
    labels:
      role: master
      cluster: prod
      
  - name: mysql-slave-1
    type: mysql
    host: mysql-slave-1.svc
    labels:
      role: slave
      cluster: prod

readWriteSplit:
  enabled: true
  cluster: prod
  maxReplicaLag: 1s
```

**路由规则**:
| SQL 类型 | 目标 | 说明 |
|---------|------|------|
| SELECT | 从库（轮询） | 无从库走主库 |
| INSERT/UPDATE/DELETE | 主库 | 写操作 |
| 事务内 | 主库 | 保证一致性 |

---

### 3. OIDC 认证（Phase 4.3）

**决策**: 仅实现 OIDC，覆盖 Keycloak/Dex/Azure AD

**流程**:
```
用户 → OIDC 登录 → IdP 认证 → 回调 → MystiSql Token → 后续请求
```

**配置**:
```yaml
auth:
  providers:
    - type: oidc
      name: keycloak
      issuerUrl: https://keycloak.example.com/realms/myrealm
      clientId: mystisql
      clientSecret: ${OIDC_SECRET}
      roleClaim: "roles"
      
    - type: token  # 保留原有 Token 认证
      enabled: true
```

---

### 4. 简化 RBAC（Phase 4.4）

**决策**: 不用 Casbin，实现简化版

**权限格式**: `<instance>:<database>:<action>`

**配置**:
```yaml
rbac:
  roles:
    admin:
      permissions: ["*:*:*"]
    developer:
      permissions: ["mysql-prod:*:SELECT", "mysql-prod:*:INSERT"]
    readonly:
      permissions: ["*:*:SELECT"]
      
  users:
    alice:
      roles: [admin]
    bob:
      roles: [developer]
```

---

### 5. 数据脱敏（Phase 4.4）

**决策**: 基于规则的实时脱敏

**内置规则**:
| 字段名模式 | 类型 | 脱敏规则 |
|-----------|------|---------|
| `*phone*`, `*mobile*` | 手机号 | `138****1234` |
| `*idcard*`, `*id_card*` | 身份证 | `110108****1234` |
| `*email*`, `*mail*` | 邮箱 | `ali***@example.com` |
| `*bank*`, `*card*` | 银行卡 | `************1234` |

**配置**:
```yaml
masking:
  enabled: true
  rolePolicy:
    admin: none      # 不脱敏
    developer: all   # 全脱敏
    readonly: all
```

---

### 6. Consul 发现（Phase 4.5）

**决策**: 只实现 Consul，覆盖非 K8s 场景

**配置**:
```yaml
discovery:
  consul:
    enabled: true
    address: "consul:8500"
    serviceTag: "database"
    healthCheck: true
```

## Risks / Trade-offs

| 风险 | 概率 | 缓解措施 |
|-----|------|---------|
| WebUI 工期超期 | 中 | MVP 功能优先，高级功能推迟 |
| 读写分离延迟 | 中 | 可配置阈值 + 强制主库 |
| OIDC 集成复杂 | 中 | 使用成熟库 `go-oidc` |
| RBAC 功能不足 | 低 | 预留扩展点 |
| WebUI 安全 | 中 | CSP + XSS 防护 |

## Migration Plan

### Phase 4.1: WebUI 基础版（3 周）
**交付物**: 可访问的 Web 界面，SQL 执行和结果查看

| 模块 | 工期 |
|-----|------|
| WebUI 框架 + 登录 | 1 周 |
| 实例列表 + SQL 编辑器 | 1.5 周 |
| 结果展示 + 部署集成 | 0.5 周 |

---

### Phase 4.2: 读写分离（2 周）
**交付物**: 自动路由读写请求

| 模块 | 工期 |
|-----|------|
| SQL 类型识别 + 路由逻辑 | 1 周 |
| 测试 + 文档 | 1 周 |

---

### Phase 4.3: OIDC 认证（2 周）
**交付物**: 企业 SSO 登录

| 模块 | 工期 |
|-----|------|
| OIDC Provider 实现 | 1.5 周 |
| WebUI 集成 + 测试 | 0.5 周 |

---

### Phase 4.4: RBAC + 脱敏（2 周）
**交付物**: 权限控制 + 数据脱敏

| 模块 | 工期 |
|-----|------|
| 简化 RBAC | 1 周 |
| 数据脱敏 | 1 周 |

---

### Phase 4.5: Consul 发现（1 周，可选）
**交付物**: 非 K8s 环境的服务发现

| 模块 | 工期 |
|-----|------|
| Consul Discoverer | 0.5 周 |
| 测试 + 文档 | 0.5 周 |

---

## Open Questions

| 问题 | 决策 |
|-----|------|
| WebUI 部署方式？ | 嵌入静态资源，同进程 |
| RBAC 用 Casbin？ | ❌ 简化版，预留扩展点 |
| LDAP 要实现？ | ❌ OIDC 覆盖，LDAP 通过 Dex 代理 |
| 多集群支持？ | ❌ Phase 5 |
| SQL 注入检测？ | ❌ 建议 WAF 处理 |
