package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ipLimiter struct {
	tokens     float64
	lastCheck  time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*ipLimiter
	rate     float64
	burst    float64
}

func newRateLimiter(requestsPerMinute int, burst int) *rateLimiter {
	return &rateLimiter{
		limiters: make(map[string]*ipLimiter),
		rate:     float64(requestsPerMinute) / 60.0,
		burst:    float64(burst),
	}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	lim, ok := rl.limiters[key]
	if !ok {
		lim = &ipLimiter{tokens: rl.burst, lastCheck: now}
		rl.limiters[key] = lim
	}

	elapsed := now.Sub(lim.lastCheck).Seconds()
	lim.lastCheck = now
	lim.tokens += elapsed * rl.rate
	if lim.tokens > rl.burst {
		lim.tokens = rl.burst
	}
	if lim.tokens < 1 {
		return false
	}
	lim.tokens--
	return true
}

// AuthRateLimitMiddleware limits auth endpoints per client IP.
func AuthRateLimitMiddleware() gin.HandlerFunc {
	limiter := newRateLimiter(5, 5)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.allow("auth:" + ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Try again later."})
			c.Abort()
			return
		}
		c.Next()
	}
}

// PublicRateLimitMiddleware limits general public endpoints per IP.
func PublicRateLimitMiddleware() gin.HandlerFunc {
	limiter := newRateLimiter(60, 60)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.allow("public:" + ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Try again later."})
			c.Abort()
			return
		}
		c.Next()
	}
}
