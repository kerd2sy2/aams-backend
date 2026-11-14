package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count       int
	windowStart time.Time
	lastSeen    time.Time
}

func isLoopback(ip string) bool {
	return ip == "127.0.0.1" || ip == "::1" || ip == "localhost" || ip == "0.0.0.0" || ip == ""
}

func allowVisitor(mu *sync.Mutex, store map[string]*visitor, ip string, maxReq int, window time.Duration) bool {
	now := time.Now()
	mu.Lock()
	defer mu.Unlock()

	v, exists := store[ip]
	if !exists {
		store[ip] = &visitor{count: 1, windowStart: now, lastSeen: now}
		return true
	}

	// Reset counter strictly when the fixed window starting at windowStart expires.
	if now.Sub(v.windowStart) > window {
		v.count = 1
		v.windowStart = now
		v.lastSeen = now
		return true
	}

	v.count++
	v.lastSeen = now
	return v.count <= maxReq
}

// RateLimiter returns a Gin middleware that limits requests per IP address.
// maxRequests: maximum number of requests allowed within the window duration.
// window: the fixed time window before counters reset.
func RateLimiter(maxRequests int, window time.Duration) gin.HandlerFunc {
	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitor)
	)

	// Clean up expired visitors every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > window {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if isLoopback(ip) {
			c.Next()
			return
		}

		if !allowVisitor(&mu, visitors, ip, maxRequests, window) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "تم تجاوز الحد المسموح من الطلبات. حاول مرة أخرى لاحقاً",
			})
			return
		}

		c.Next()
	}
}

// StrictLoginLimiter is a rate limiter for the login endpoint.
// Tuned for shared carrier CGNAT / branch Wi-Fi across high-concurrency rollouts.
func StrictLoginLimiter() gin.HandlerFunc {
	var (
		mu       sync.Mutex
		attempts = make(map[string]*visitor)
	)
	const window = 5 * time.Minute
	const maxAttempts = 120

	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, v := range attempts {
				if time.Since(v.lastSeen) > window {
					delete(attempts, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if isLoopback(ip) {
			c.Next()
			return
		}

		if !allowVisitor(&mu, attempts, ip, maxAttempts, window) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "تم تجاوز عدد محاولات تسجيل الدخول المسموحة. يرجى الانتظار قليلاً والمحاولة لاحقاً.",
			})
			return
		}

		c.Next()
	}
}
