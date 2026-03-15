## Why

MystiSql 已完成 Phase 1-3，具备了基础的数据库访问、连接池管理、安全认证和审计日志能力。为了满足企业级生产环境需求并提供良好的用户体验，需要增强 WebUI 界面、读写分离、企业认证和权限控制能力。

**用户价值驱动**:
1. ❌ 无 WebUI → 用户只能通过 CLI/API 访问，学习成本高
2. ❌ 不支持读写分离 → 主从架构下查询全部走主库，性能瓶颈
3. ❌ 仅支持 Token 认证 → 无法集成企业 SSO
4. ❌ 无权限控制 → 所有用户权限相同
5. ❌ 敏感数据明文 → 不符合合规要求

## What Changes

### Phase 4.1: WebUI 基础版（3 周）
- **登录页**: 复用现有 Token 认证
- **实例列表**: 显示实例状态
- **SQL 编辑器**: Monaco Editor + 执行按钮
- **结果展示**: 表格 + 分页 + 导出

### Phase 4.2: 读写分离（2 周）
- 基于标签的主从路由
- 自动识别 SQL 类型（读/写）
- 事务内强制走主库

### Phase 4.3: OIDC 认证（2 周）
- OIDC Provider 实现（Keycloak/Dex/Azure AD）
- 角色/组映射
- 保留 Token 认证作为备选

### Phase 4.4: RBAC + 脱敏（2 周）
- 简化 RBAC（库/表/操作级别）
- 数据脱敏（手机号/身份证/邮箱/银行卡）

### Phase 4.5: Consul 发现（1 周，可选）
- Consul 服务注册中心发现
- 补充 K8s 之外的发现方式

## Capabilities

### New Capabilities

#### WebUI（Phase 4.1 优先）
- `webui-interface`: WebUI 基础界面
- `sql-editor`: SQL 编辑器（Monaco）
- `result-display`: 结果集展示

#### 连接层
- `read-write-splitting`: 读写分离

#### 安全增强
- `oidc-auth`: OIDC/OAuth2 认证
- `rbac-permissions`: 简化 RBAC
- `data-masking`: 数据脱敏

#### 服务发现
- `consul-discovery`: Consul 发现

### Modified Capabilities

- `auth-service`: 扩展支持 OIDC Provider
- `connection-pool`: 集成读写分离路由

## Impact

### 代码结构
```
internal/
├── api/
│   └── webui/           # 新增：WebUI 路由
├── connection/
│   └── router/          # 新增：读写分离路由器
├── service/
│   ├── auth/
│   │   └── oidc/        # 新增：OIDC 认证
│   ├── rbac/            # 新增：简化 RBAC
│   └── masking/         # 新增：数据脱敏
└── discovery/
    └── consul/          # 新增：Consul 发现

web/                     # 新增：前端项目
├── src/
│   ├── pages/
│   ├── components/
│   └── api/
└── package.json
```

### API 变更
```
# WebUI
GET  /                        # WebUI 入口
GET  /assets/*                # 静态资源

# OIDC
GET  /api/v1/auth/oidc/login    # OIDC 登录
GET  /api/v1/auth/oidc/callback # OIDC 回调

# RBAC
GET    /api/v1/rbac/roles              # 角色列表
POST   /api/v1/rbac/roles              # 创建角色
DELETE /api/v1/rbac/roles/{name}       # 删除角色
GET    /api/v1/rbac/users/{id}/roles   # 用户角色
POST   /api/v1/rbac/users/{id}/roles   # 分配角色

# 审计日志（扩展现有）
GET    /api/v1/audit/logs              # 日志查询
```

### 配置变更
```yaml
# 新增配置节
webui:
  enabled: true
  mode: embedded

readWriteSplit:
  enabled: true
  cluster: production

auth:
  providers:
    - type: oidc
      name: keycloak
      issuerUrl: https://keycloak.example.com/realms/myrealm
      clientId: mystisql
      clientSecret: ${OIDC_SECRET}

rbac:
  enabled: true
  roles:
    admin:
      permissions: ["*:*:*"]

masking:
  enabled: true
  rolePolicy:
    admin: none
    developer: all

discovery:
  consul:
    enabled: true
    address: "consul:8500"
```

### 依赖
```
# Go 依赖
github.com/coreos/go-oidc/v3  # OIDC 客户端
github.com/hashicorp/consul/api  # Consul 客户端

# 前端依赖（web/package.json）
react: ^18.2.0
antd: ^5.12.0
@monaco-editor/react: ^4.6.0
typescript: ^5.3.0
```
