package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"InstanceNotFound", ErrInstanceNotFound, "实例未找到"},
		{"InstanceAlreadyExists", ErrInstanceAlreadyExists, "实例已存在"},
		{"DiscoveryFailed", ErrDiscoveryFailed, "发现失败"},
		{"InvalidInstanceConfig", ErrInvalidInstanceConfig, "无效的实例配置"},
		{"ConnectionFailed", ErrConnectionFailed, "连接失败"},
		{"ConnectionClosed", ErrConnectionClosed, "连接已关闭"},
		{"ConnectionTimeout", ErrConnectionTimeout, "连接超时"},
		{"QueryFailed", ErrQueryFailed, "查询失败"},
		{"QueryTimeout", ErrQueryTimeout, "查询超时"},
		{"ResultSetClosed", ErrResultSetClosed, "结果集已关闭"},
		{"ConfigNotFound", ErrConfigNotFound, "配置文件未找到"},
		{"ConfigInvalid", ErrConfigInvalid, "配置文件无效"},
		{"ConfigParseFailed", ErrConfigParseFailed, "配置文件解析失败"},
		{"ValidationFailed", ErrValidationFailed, "验证失败"},
		{"MissingRequiredField", ErrMissingRequiredField, "缺少必填字段"},
		{"InvalidFieldValue", ErrInvalidFieldValue, "无效的字段值"},
		{"InvalidRequest", ErrInvalidRequest, "无效的请求"},
		{"Unauthorized", ErrUnauthorized, "未授权"},
		{"Forbidden", ErrForbidden, "禁止访问"},
		{"InternalServer", ErrInternalServer, "内部服务器错误"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("sentinel error is nil")
			}
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestErrorUniqueness(t *testing.T) {
	all := []error{
		ErrInstanceNotFound, ErrInstanceAlreadyExists, ErrDiscoveryFailed, ErrInvalidInstanceConfig,
		ErrConnectionFailed, ErrConnectionClosed, ErrConnectionTimeout, ErrQueryFailed,
		ErrQueryTimeout, ErrResultSetClosed,
		ErrConfigNotFound, ErrConfigInvalid, ErrConfigParseFailed,
		ErrValidationFailed, ErrMissingRequiredField, ErrInvalidFieldValue,
		ErrInvalidRequest, ErrUnauthorized, ErrForbidden, ErrInternalServer,
	}
	for i, a := range all {
		for j, b := range all {
			if i != j && a == b {
				t.Errorf("errors[%d] (%q) == errors[%d] (%q), should be unique", i, a, j, b)
			}
		}
	}
}

func TestErrorsIs(t *testing.T) {
	wrapped := fmt.Errorf("context: %w", ErrInstanceNotFound)
	if !errors.Is(wrapped, ErrInstanceNotFound) {
		t.Error("errors.Is should match wrapped error")
	}
	if errors.Is(wrapped, ErrConnectionFailed) {
		t.Error("errors.Is should not match different error")
	}
}
