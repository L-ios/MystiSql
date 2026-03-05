package errors

import "errors"

// 发现相关错误
var (
	// ErrInstanceNotFound 实例未找到
	ErrInstanceNotFound = errors.New("实例未找到")

	// ErrInstanceAlreadyExists 实例已存在
	ErrInstanceAlreadyExists = errors.New("实例已存在")

	// ErrDiscoveryFailed 发现失败
	ErrDiscoveryFailed = errors.New("发现失败")

	// ErrInvalidInstanceConfig 无效的实例配置
	ErrInvalidInstanceConfig = errors.New("无效的实例配置")
)

// 连接相关错误
var (
	// ErrConnectionFailed 连接失败
	ErrConnectionFailed = errors.New("连接失败")

	// ErrConnectionClosed 连接已关闭
	ErrConnectionClosed = errors.New("连接已关闭")

	// ErrConnectionTimeout 连接超时
	ErrConnectionTimeout = errors.New("连接超时")

	// ErrQueryFailed 查询失败
	ErrQueryFailed = errors.New("查询失败")

	// ErrQueryTimeout 查询超时
	ErrQueryTimeout = errors.New("查询超时")
)

// 配置相关错误
var (
	// ErrConfigNotFound 配置文件未找到
	ErrConfigNotFound = errors.New("配置文件未找到")

	// ErrConfigInvalid 配置文件无效
	ErrConfigInvalid = errors.New("配置文件无效")

	// ErrConfigParseFailed 配置文件解析失败
	ErrConfigParseFailed = errors.New("配置文件解析失败")
)

// 验证相关错误
var (
	// ErrValidationFailed 验证失败
	ErrValidationFailed = errors.New("验证失败")

	// ErrMissingRequiredField 缺少必填字段
	ErrMissingRequiredField = errors.New("缺少必填字段")

	// ErrInvalidFieldValue 无效的字段值
	ErrInvalidFieldValue = errors.New("无效的字段值")
)

// API 相关错误
var (
	// ErrInvalidRequest 无效的请求
	ErrInvalidRequest = errors.New("无效的请求")

	// ErrUnauthorized 未授权
	ErrUnauthorized = errors.New("未授权")

	// ErrForbidden 禁止访问
	ErrForbidden = errors.New("禁止访问")

	// ErrInternalServer 内部服务器错误
	ErrInternalServer = errors.New("内部服务器错误")
)
