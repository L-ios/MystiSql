package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	// RequestsPerMinute 每分钟允许的最大请求数
	RequestsPerMinute int
	// ClientKeyFn 从请求中提取限流键，默认使用客户端 IP
	ClientKeyFn func(c *gin.Context) string
}

// slidingWindow 滑动窗口计数器
type slidingWindow struct {
	mu       sync.Mutex
	requests []time.Time
	limit    int
	window   time.Duration
}

func (w *slidingWindow) allow(now time.Time) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 清理窗口外的请求记录
	cutoff := now.Add(-w.window)
	keep := w.requests[:0]
	for _, t := range w.requests {
		if t.After(cutoff) {
			keep = append(keep, t)
		}
	}
	w.requests = keep

	if len(w.requests) >= w.limit {
		return false
	}
	w.requests = append(w.requests, now)
	return true
}

// rateLimiter 基于客户端键的滑动窗口限流器
type rateLimiter struct {
	mu      sync.Mutex
	windows map[string]*slidingWindow
	limit   int
	window  time.Duration

	// lastCleanup 上次全局清理时间，避免 map 无限增长
	lastCleanup  time.Time
	cleanupEvery time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		windows:      make(map[string]*slidingWindow),
		limit:        limit,
		window:       window,
		lastCleanup:  time.Now(),
		cleanupEvery: 10 * time.Minute,
	}
}

func (r *rateLimiter) allow(key string, now time.Time) bool {
	r.mu.Lock()
	if now.Sub(r.lastCleanup) > r.cleanupEvery {
		r.cleanupLocked(now)
		r.lastCleanup = now
	}

	w, exists := r.windows[key]
	if !exists {
		w = &slidingWindow{limit: r.limit, window: r.window}
		r.windows[key] = w
	}
	r.mu.Unlock()

	return w.allow(now)
}

// cleanupLocked 清理长时间无活动的窗口（需持有 r.mu）
func (r *rateLimiter) cleanupLocked(now time.Time) {
	cutoff := now.Add(-r.cleanupEvery)
	for key, w := range r.windows {
		w.mu.Lock()
		if len(w.requests) == 0 || !w.requests[len(w.requests)-1].After(cutoff) {
			delete(r.windows, key)
		}
		w.mu.Unlock()
	}
}

// RateLimit 返回滑动窗口速率限制中间件。
// 超出限制的请求返回 429 Too Many Requests。
func RateLimit(config RateLimitConfig) gin.HandlerFunc {
	if config.RequestsPerMinute <= 0 {
		config.RequestsPerMinute = 60
	}
	if config.ClientKeyFn == nil {
		config.ClientKeyFn = func(c *gin.Context) string { return c.ClientIP() }
	}

	limiter := newRateLimiter(config.RequestsPerMinute, time.Minute)

	return func(c *gin.Context) {
		key := config.ClientKeyFn(c)
		if !limiter.allow(key, time.Now()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": gin.H{
					"code":    "RATE_LIMITED",
					"message": "请求频率超出限制，请稍后重试",
				},
			})
			return
		}
		c.Next()
	}
}
