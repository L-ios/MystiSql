package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupRateLimitRouter(config RateLimitConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit(config))
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"pong": true})
	})
	return r
}

func TestRateLimit_AllowsUnderLimit(t *testing.T) {
	router := setupRateLimitRouter(RateLimitConfig{RequestsPerMinute: 5})

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ping", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, w.Code)
		}
	}
}

func TestRateLimit_BlocksOverLimit(t *testing.T) {
	router := setupRateLimitRouter(RateLimitConfig{RequestsPerMinute: 3})

	for i := 0; i < 3; i++ {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/ping", nil))
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", w.Code)
	}

	var body map[string]map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if body["error"]["code"] != "RATE_LIMITED" {
		t.Errorf("error code = %v, want RATE_LIMITED", body["error"]["code"])
	}
}

func TestRateLimit_WindowSlides(t *testing.T) {
	limiter := newRateLimiter(2, 50*time.Millisecond)

	now := time.Now()
	if !limiter.allow("k", now) || !limiter.allow("k", now) {
		t.Fatal("first two requests should be allowed")
	}
	if limiter.allow("k", now) {
		t.Fatal("third request should be blocked")
	}

	// 窗口滑过后应重新放行
	later := now.Add(60 * time.Millisecond)
	if !limiter.allow("k", later) {
		t.Error("request after window slide should be allowed")
	}
}

func TestRateLimit_DefaultConfig(t *testing.T) {
	// RequestsPerMinute<=0 时应回退到默认 60
	router := setupRateLimitRouter(RateLimitConfig{})

	for i := 0; i < 60; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ping", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, w.Code)
		}
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/ping", nil))
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", w.Code)
	}
}
