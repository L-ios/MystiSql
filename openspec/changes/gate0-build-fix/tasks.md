## 1. WebUI Build Tag 拆分

- [x] 1.1 创建 `web/web_handler_stub.go`（`//go:build !webembed`），`NewHandler()` 返回 `(nil, nil)`
- [x] 1.2 创建 `web/web_handler_embed.go`（`//go:build webembed`），保留 `//go:embed all:dist` 和 SPA 路由逻辑
- [x] 1.3 删除原 `web/handler.go`

## 2. 验证

- [x] 2.1 `go build ./...` 零错误（无 webembed tag）
- [x] 2.2 `go vet ./...` 零错误
- [x] 2.3 `go build -tags webembed ./...` 在无 dist 时按预期失败（可选验证）
