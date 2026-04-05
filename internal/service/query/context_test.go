package query

import (
	"context"
	"testing"
)

func TestContextWithUserID(t *testing.T) {
	ctx := context.Background()

	// 未设置时返回 unknown
	if got := getUserIDFromContext(ctx); got != "unknown" {
		t.Errorf("getUserIDFromContext(empty) = %q, want %q", got, "unknown")
	}

	// 设置后能正确取出
	ctx = ContextWithUserID(ctx, "user-123")
	if got := getUserIDFromContext(ctx); got != "user-123" {
		t.Errorf("getUserIDFromContext = %q, want %q", got, "user-123")
	}
}

func TestContextWithClientIP(t *testing.T) {
	ctx := context.Background()

	// 未设置时返回 unknown
	if got := getClientIPFromContext(ctx); got != "unknown" {
		t.Errorf("getClientIPFromContext(empty) = %q, want %q", got, "unknown")
	}

	// 设置后能正确取出
	ctx = ContextWithClientIP(ctx, "192.168.1.1")
	if got := getClientIPFromContext(ctx); got != "192.168.1.1" {
		t.Errorf("getClientIPFromContext = %q, want %q", got, "192.168.1.1")
	}
}

func TestContextWithRole(t *testing.T) {
	ctx := context.Background()

	// 未设置时返回空字符串
	if got := getRoleFromContext(ctx); got != "" {
		t.Errorf("getRoleFromContext(empty) = %q, want empty string", got)
	}

	// 设置后能正确取出
	ctx = ContextWithRole(ctx, "admin")
	if got := getRoleFromContext(ctx); got != "admin" {
		t.Errorf("getRoleFromContext = %q, want %q", got, "admin")
	}

	// 覆盖已有值
	ctx = ContextWithRole(ctx, "readonly")
	if got := getRoleFromContext(ctx); got != "readonly" {
		t.Errorf("getRoleFromContext after override = %q, want %q", got, "readonly")
	}
}

func TestContextWithUserIDWrongType(t *testing.T) {
	// 在 context 中存入非 string 类型的值
	ctx := context.WithValue(context.Background(), contextKey("user_id"), 12345)
	if got := getUserIDFromContext(ctx); got != "unknown" {
		t.Errorf("getUserIDFromContext(wrong type) = %q, want %q", got, "unknown")
	}
}

func TestContextWithClientIPWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey("client_ip"), []string{"1.2.3.4"})
	if got := getClientIPFromContext(ctx); got != "unknown" {
		t.Errorf("getClientIPFromContext(wrong type) = %q, want %q", got, "unknown")
	}
}

func TestContextWithRoleWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey("role"), 42)
	if got := getRoleFromContext(ctx); got != "" {
		t.Errorf("getRoleFromContext(wrong type) = %q, want empty string", got)
	}
}
