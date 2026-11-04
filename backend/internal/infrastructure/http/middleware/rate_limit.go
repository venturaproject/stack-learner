package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"boilerplate-learner/internal/infrastructure/cache"
)

type ipEntry struct {
	count     int
	windowEnd time.Time
}

// RateLimiter is a fixed-window per-client limiter. When a Redis instance is
// provided the window is shared across processes; otherwise it falls back to an
// in-process map (also used transparently if Redis errors, so a cache outage
// never disables limiting).
type RateLimiter struct {
	redis       *cache.Redis
	name        string
	maxRequests int
	window      time.Duration

	mu    sync.Mutex
	store map[string]*ipEntry
}

func NewRateLimiter(redis *cache.Redis, name string, maxRequests int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		redis:       redis,
		name:        name,
		maxRequests: maxRequests,
		window:      window,
		store:       make(map[string]*ipEntry),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if rl.allow(c, clientIP(c)) {
			c.Next()
			return
		}
		c.Header("Retry-After", strconv.Itoa(int(rl.window.Seconds())))
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Demasiados intentos, espera un momento"})
	}
}

func (rl *RateLimiter) allow(c *gin.Context, ip string) bool {
	if rl.redis != nil {
		count, err := rl.redis.Incr(c.Request.Context(), "ratelimit:"+rl.name+":"+ip, rl.window)
		if err == nil {
			return count <= int64(rl.maxRequests)
		}
		// Redis hiccup — fall through to the in-process limiter.
	}
	return rl.allowInMemory(ip)
}

func (rl *RateLimiter) allowInMemory(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, ok := rl.store[ip]
	if !ok || now.After(entry.windowEnd) {
		rl.store[ip] = &ipEntry{count: 1, windowEnd: now.Add(rl.window)}
		return true
	}
	entry.count++
	return entry.count <= rl.maxRequests
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		now := time.Now()
		rl.mu.Lock()
		for ip, e := range rl.store {
			if now.After(e.windowEnd) {
				delete(rl.store, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func clientIP(c *gin.Context) string {
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	return c.ClientIP()
}
