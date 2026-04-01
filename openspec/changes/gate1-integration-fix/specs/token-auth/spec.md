## MODIFIED Requirements

### Requirement: Token 生成接口
系统 SHALL 提供 RESTful API 接口生成 Token，CLI 命令 SHALL 正确发送请求 body。

#### Scenario: CLI 生成 Token 发送 JSON body
- **WHEN** 执行 `mystisql auth token --user-id admin --role admin --server http://localhost:8080`
- **THEN** CLI 发送 POST 请求，body 为 JSON {"user_id":"admin","role":"admin"}（非 nil body）
- **AND** 服务端返回有效的 JWT Token
