# Capability: Data Masking

## Purpose

提供数据脱敏能力，保护敏感数据（如手机号、身份证、邮箱、银行卡）在查询结果中的隐私安全，支持基于角色的差异化脱敏策略。

## Requirements

### Requirement: 数据脱敏规则
系统 SHALL 支持配置不同类型的数据脱敏规则。

#### Scenario: 手机号脱敏
- **WHEN** 字段类型为手机号
- **THEN** 显示为 `138****1234` 格式

#### Scenario: 身份证脱敏
- **WHEN** 字段类型为身份证
- **THEN** 显示为 `110108****1234` 格式

#### Scenario: 邮箱脱敏
- **WHEN** 字段类型为邮箱
- **THEN** 显示为 `ali***@example.com` 格式

#### Scenario: 银行卡脱敏
- **WHEN** 字段类型为银行卡
- **THEN** 显示为 `************1234` 格式

#### Scenario: 完全脱敏
- **WHEN** 配置完全脱敏
- **THEN** 显示为 `******`

### Requirement: 脱敏规则配置
系统 SHALL 支持灵活的脱敏规则配置。

#### Scenario: 配置脱敏规则
- **WHEN** 配置 `masking.rules[].type = "phone"` 和 `pattern = "prefix:3,suffix:4"`
- **THEN** 手机号保留前 3 后 4 位

#### Scenario: 配置字段级脱敏
- **WHEN** 配置 `masking.fields["users.phone"].enabled = true`
- **THEN** users 表的 phone 字段应用脱敏

### Requirement: 基于角色的脱敏
系统 SHALL 支持基于用户角色的脱敏策略。

#### Scenario: 管理员查看原始数据
- **WHEN** 用户角色为 admin
- **THEN** 不应用脱敏，显示原始数据

#### Scenario: 普通用户查看脱敏数据
- **WHEN** 用户角色为 analyst
- **THEN** 应用脱敏规则

### Requirement: 脱敏审计
系统 SHALL 记录脱敏操作日志。

#### Scenario: 记录脱敏事件
- **WHEN** 数据被脱敏
- **THEN** 记录脱敏字段、用户、时间
