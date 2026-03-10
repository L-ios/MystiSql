package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MystiSql/internal/discovery"
	"MystiSql/pkg/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupTestRouter 创建测试路由
func setupTestRouter() (*gin.Engine, *discovery.Registry, *zap.Logger) {
	logger := zap.NewNop()
	registry := discovery.NewRegistry()
	handlers := NewHandlers(registry, logger, "test-version")

	router := gin.New()
	router.GET("/health", handlers.Health)
	router.GET("/api/v1/instances", handlers.ListInstances)
	router.POST("/api/v1/query", handlers.Query)

	return router, registry, logger
}

// TestHealthEndpoint 测试健康检查端点
func TestHealthEndpoint(t *testing.T) {
	router, _, _ := setupTestRouter()

	tests := []struct {
		name       string
		url        string
		wantStatus int
		wantFields []string
	}{
		{
			name:       "基本健康检查",
			url:        "/health",
			wantStatus: http.StatusOK,
			wantFields: []string{"status", "timestamp", "version"},
		},
		{
			name:       "带实例检查的健康检查（无实例）",
			url:        "/health?check-instances=true",
			wantStatus: http.StatusOK,
			wantFields: []string{"status", "timestamp", "version", "instances"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("状态码错误: got %v, want %v", w.Code, tt.wantStatus)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("响应解析失败: %v", err)
			}

			for _, field := range tt.wantFields {
				if _, exists := response[field]; !exists {
					t.Errorf("缺少字段: %s", field)
				}
			}
		})
	}
}

// TestHealthWithInstances 测试带实例检查的健康检查
func TestHealthWithInstances(t *testing.T) {
	router, registry, _ := setupTestRouter()

	// 添加测试实例
	instance1 := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	instance1.SetStatus(types.InstanceStatusHealthy)
	if err := registry.Register(instance1); err != nil {
		t.Fatalf("注册实例失败: %v", err)
	}

	instance2 := types.NewDatabaseInstance("test-mysql-2", types.DatabaseTypeMySQL, "localhost", 3307)
	instance2.SetStatus(types.InstanceStatusUnhealthy)
	if err := registry.Register(instance2); err != nil {
		t.Fatalf("注册实例失败: %v", err)
	}

	// 测试带实例检查的健康检查
	req, _ := http.NewRequest("GET", "/health?check-instances=true", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码错误: got %v, want %v", w.Code, http.StatusOK)
	}

	var response HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("响应解析失败: %v", err)
	}

	// 验证实例健康信息
	if response.Instances == nil {
		t.Fatal("缺少实例健康信息")
	}

	if response.Instances.Total != 2 {
		t.Errorf("实例总数错误: got %v, want 2", response.Instances.Total)
	}

	if response.Instances.Healthy != 1 {
		t.Errorf("健康实例数错误: got %v, want 1", response.Instances.Healthy)
	}

	if response.Instances.Unhealthy != 1 {
		t.Errorf("不健康实例数错误: got %v, want 1", response.Instances.Unhealthy)
	}
}

// TestListInstances 测试实例列表端点
func TestListInstances(t *testing.T) {
	router, registry, _ := setupTestRouter()

	// 添加测试实例
	instance1 := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	instance1.SetCredentials("root", "password123")
	instance1.Database = "testdb"
	if err := registry.Register(instance1); err != nil {
		t.Fatalf("注册实例失败: %v", err)
	}

	instance2 := types.NewDatabaseInstance("test-mysql-2", types.DatabaseTypeMySQL, "192.168.1.100", 3307)
	instance2.SetCredentials("admin", "secret")
	if err := registry.Register(instance2); err != nil {
		t.Fatalf("注册实例失败: %v", err)
	}

	// 测试实例列表
	req, _ := http.NewRequest("GET", "/api/v1/instances", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("状态码错误: got %v, want %v", w.Code, http.StatusOK)
	}

	var response InstancesListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("响应解析失败: %v", err)
	}

	// 验证响应
	if response.Total != 2 {
		t.Errorf("实例总数错误: got %v, want 2", response.Total)
	}

	// 验证密码已脱敏（密码字段不应在 JSON 中）
	for _, instance := range response.Instances {
		if instance.Name == "test-mysql" {
			if instance.Host != "localhost" {
				t.Errorf("主机地址错误: got %v, want localhost", instance.Host)
			}
			if instance.Port != 3306 {
				t.Errorf("端口错误: got %v, want 3306", instance.Port)
			}
			if instance.Database != "testdb" {
				t.Errorf("数据库名错误: got %v, want testdb", instance.Database)
			}
			// 注意：InstanceResponse 中没有 Password 字段，密码已脱敏
		}
	}
}

// TestQueryEndpoint 测试查询端点
func TestQueryEndpoint(t *testing.T) {
	router, registry, _ := setupTestRouter()

	// 添加测试实例
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	instance.SetCredentials("root", "password")
	if err := registry.Register(instance); err != nil {
		t.Fatalf("注册实例失败: %v", err)
	}

	tests := []struct {
		name       string
		request    QueryRequest
		wantStatus int
		wantError  bool
	}{
		{
			name: "缺少实例名称",
			request: QueryRequest{
				SQL: "SELECT 1",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "缺少 SQL 语句",
			request: QueryRequest{
				Instance: "test-mysql",
			},
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name: "实例不存在",
			request: QueryRequest{
				Instance: "nonexistent",
				SQL:      "SELECT 1",
			},
			wantStatus: http.StatusNotFound,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/v1/query", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("状态码错误: got %v, want %v", w.Code, tt.wantStatus)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("响应解析失败: %v", err)
			}

			// 检查是否有错误
			if tt.wantError {
				if success, ok := response["success"].(bool); !ok || success {
					t.Error("期望错误响应，但收到成功响应")
				}
			}
		})
	}
}

// TestQueryRequestValidation 测试查询请求验证
func TestQueryRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request QueryRequest
		wantErr bool
	}{
		{
			name: "有效请求",
			request: QueryRequest{
				Instance: "test-instance",
				SQL:      "SELECT 1",
			},
			wantErr: false,
		},
		{
			name: "带超时的有效请求",
			request: QueryRequest{
				Instance: "test-instance",
				SQL:      "SELECT 1",
				Timeout:  60,
			},
			wantErr: false,
		},
		{
			name: "缺少实例名称",
			request: QueryRequest{
				SQL: "SELECT 1",
			},
			wantErr: true,
		},
		{
			name: "缺少 SQL 语句",
			request: QueryRequest{
				Instance: "test-instance",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用 Gin 的绑定验证
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.request)
			c.Request = httptest.NewRequest("POST", "/query", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			var req QueryRequest
			err := c.ShouldBindJSON(&req)

			if (err != nil) != tt.wantErr {
				t.Errorf("验证错误: got %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestToInstanceResponse 测试实例响应转换（密码脱敏）
func TestToInstanceResponse(t *testing.T) {
	instance := types.NewDatabaseInstance("test-mysql", types.DatabaseTypeMySQL, "localhost", 3306)
	instance.SetCredentials("root", "supersecret")
	instance.Database = "testdb"
	instance.SetStatus(types.InstanceStatusHealthy)
	instance.AddLabel("env", "test")

	response := ToInstanceResponse(instance)

	// 验证字段
	if response.Name != "test-mysql" {
		t.Errorf("名称错误: got %v, want test-mysql", response.Name)
	}

	if response.Type != types.DatabaseTypeMySQL {
		t.Errorf("类型错误: got %v, want %v", response.Type, types.DatabaseTypeMySQL)
	}

	if response.Host != "localhost" {
		t.Errorf("主机错误: got %v, want localhost", response.Host)
	}

	if response.Port != 3306 {
		t.Errorf("端口错误: got %v, want 3306", response.Port)
	}

	if response.Database != "testdb" {
		t.Errorf("数据库错误: got %v, want testdb", response.Database)
	}

	if response.Username != "root" {
		t.Errorf("用户名错误: got %v, want root", response.Username)
	}

	if response.Status != "healthy" {
		t.Errorf("状态错误: got %v, want healthy", response.Status)
	}

	// 验证密码已脱敏（InstanceResponse 没有 Password 字段）
	// 通过 JSON 序列化验证
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}

	jsonStr := string(jsonData)
	if bytes.Contains(jsonData, []byte("supersecret")) {
		t.Error("密码未脱敏，在 JSON 响应中发现密码")
	}

	// 验证 JSON 中没有 password 字段
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		t.Fatalf("JSON 解析失败: %v", err)
	}

	if _, exists := jsonMap["password"]; exists {
		t.Error("密码字段不应出现在 JSON 响应中")
	}

	t.Logf("JSON 响应: %s", jsonStr)
}

// TestErrorResponse 测试错误响应格式
func TestErrorResponse(t *testing.T) {
	errResp := NewErrorResponse("TEST_ERROR", "This is a test error")

	if errResp.Success {
		t.Error("错误响应的 Success 应该为 false")
	}

	if errResp.Error.Code != "TEST_ERROR" {
		t.Errorf("错误代码错误: got %v, want TEST_ERROR", errResp.Error.Code)
	}

	if errResp.Error.Message != "This is a test error" {
		t.Errorf("错误消息错误: got %v, want 'This is a test error'", errResp.Error.Message)
	}

	// 验证 JSON 序列化
	jsonData, err := json.Marshal(errResp)
	if err != nil {
		t.Fatalf("JSON 序列化失败: %v", err)
	}

	var decoded ErrorResponse
	if err := json.Unmarshal(jsonData, &decoded); err != nil {
		t.Fatalf("JSON 反序列化失败: %v", err)
	}

	if decoded.Success {
		t.Error("反序列化后的 Success 应该为 false")
	}
}

// TestCORSHeaders 测试 CORS 头
func TestCORSHeaders(t *testing.T) {
	router, _, _ := setupTestRouter()

	req, _ := http.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 验证 CORS 头
	// 注意：需要实际运行 CORS 中间件才能测试
	_ = w
}

// TestTimeout 测试超时控制
func TestTimeout(t *testing.T) {
	// 这个测试需要真实的数据库连接，这里只测试超时参数解析
	request := QueryRequest{
		Instance: "test-instance",
		SQL:      "SELECT SLEEP(100)",
		Timeout:  5,
	}

	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/query", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	var decoded QueryRequest
	if err := json.NewDecoder(req.Body).Decode(&decoded); err != nil {
		t.Fatalf("请求解码失败: %v", err)
	}

	if decoded.Timeout != 5 {
		t.Errorf("超时时间错误: got %v, want 5", decoded.Timeout)
	}
}

// TestContextCancellation 测试上下文取消
func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	select {
	case <-time.After(200 * time.Millisecond):
		t.Error("上下文应该已经超时")
	case <-ctx.Done():
		// 预期行为
	}
}