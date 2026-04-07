package webui

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- Stub 模式测试（默认构建，无 -tags webembed）---

func TestNewHandler_StubMode_ReturnsNil(t *testing.T) {
	handler, err := NewHandler()

	assert.Nil(t, err, "NewHandler 不应返回错误")
	assert.Nil(t, handler, "stub 模式下 NewHandler 应返回 nil")
}

func TestStubHandler_ServeHTTP_ReturnsNotFound(t *testing.T) {
	// stub 模式下 NewHandler 返回 nil，但如果有人直接构造 Handler{} 调用 ServeHTTP
	// 应返回 404
	handler := &Handler{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code, "stub Handler 应返回 404")
	assert.Contains(t, rec.Body.String(), "WebUI not available", "响应体应包含不可用提示")
}

func TestStubHandler_ServeHTTP_ContentType(t *testing.T) {
	handler := &Handler{}
	req := httptest.NewRequest(http.MethodGet, "/any/path", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	contentType := rec.Header().Get("Content-Type")
	assert.Equal(t, "text/plain; charset=utf-8", contentType, "stub 响应 Content-Type 应为 text/plain")
}
