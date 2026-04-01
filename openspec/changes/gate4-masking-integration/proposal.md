## Why

当前 `MaskingService`（`internal/service/masking/`）实现了 4 种脱敏算法（手机/邮箱/身份证/银行卡），但存在三个阻断性问题：(1) 只能处理固定的 `MaskData` struct，不能处理动态查询结果列；(2) PolicyStore 无 mutex 线程不安全；(3) 从未被任何代码实例化或调用，是完全孤立的代码。需要重新设计数据模型和集成方式，使其真正融入查询管线。

## What Changes

### 重新设计数据模型
- **从固定 struct 改为列规则配置**：脱敏规则按 `实例名.表名.列名` 三级 key 配置，每条规则指定脱敏类型（phone/email/idcard/bankcard/custom）和参数
- **规则持久化**：YAML 文件存储脱敏规则，启动时加载，API 可动态增删改
- **PolicyStore 加 mutex**：修复线程安全问题

### 集成到查询管线
- **Engine.ExecuteQuery 后处理**：查询结果返回后，根据用户角色查找匹配的脱敏规则，对结果中的列值进行脱敏
- **列匹配策略**：按 `实例.表.列` 精确匹配优先，其次按 `列名` 全局匹配（如所有叫 `phone` 的列），支持通配符 `*`
- **性能考量**：脱敏在结果截断后执行，不影响查询性能；无匹配规则时跳过

### API 和配置
- **新增路由**：`/api/v1/masking/rules` CRUD 操作
- **配置示例**：
  ```yaml
  masking:
    enabled: true
    rules:
      - role: "readonly"
        pattern: "*.users.phone"
        type: "phone"
      - role: "readonly"
        pattern: "*.users.email"
        type: "email"
      - role: "readonly"
        pattern: "*.orders.bank_card"
        type: "bankcard"
  ```

## Capabilities

### Modified Capabilities
- `data-masking`: 从固定 struct 改为动态列规则匹配，集成到查询管线

## Impact

### 受影响的代码
- `internal/service/masking/masking.go` — 重构脱敏逻辑为动态列处理
- `internal/service/masking/policy.go` — 线程安全 + 持久化
- `internal/service/query/engine.go` — ExecuteQuery 后处理集成
- `internal/api/rest/server.go` — 注册 masking 路由
- `config/` — masking 配置段

### 前置条件
- Gate 0 完成（可编译）

### Done 标准
- readonly 角色查询 users 表，phone 列自动脱敏为 `138****56`
- admin 角色查询同一表，phone 列显示原文
- 规则通过 API 动态增删后立即生效

### 信心
**55%** — 核心难点是列名到规则的匹配策略设计：按列名匹配（`phone` → phone 脱敏）简单但可能误匹配；按 `实例.表.列` 精确匹配安全但配置繁琐。需要权衡后决定
