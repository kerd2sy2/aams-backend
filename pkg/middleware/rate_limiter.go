package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count    int
	lastSeen time.Time
}

func isLoopback(ip string) bool {
	return ip == "127.0.0.1" || ip == "::1" || ip == "localhost" || ip == "0.0.0.0" || ip == ""
}

// RateLimiter returns a Gin middleware that limits requests per IP address.
// maxRequests: maximum number of requests allowed within the window duration.
// window: the sliding time window for the rate limit.
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

		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
			mu.Unlock()
			c.Next()
			return
		}

		// Reset count if window has passed
		if time.Since(v.lastSeen) > window {
			v.count = 1
			v.lastSeen = time.Now()
			mu.Unlock()
			c.Next()
			return
		}

		v.count++
		v.lastSeen = time.Now()
		mu.Unlock()

		if v.count > maxRequests {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "تم تجاوز الحد المسموح من الطلبات. حاول مرة أخرى لاحقاً",
			})
			return
		}

		c.Next()
	}
}

// StrictLoginLimiter is a rate limiter for the login endpoint
func StrictLoginLimiter() gin.HandlerFunc {
	var (
		mu       sync.Mutex
		attempts = make(map[string]*visitor)
	)

	go func() {
		for {
			time.Sleep(5 * time.Minute)
			mu.Lock()
			for ip, v := range attempts {
				if time.Since(v.lastSeen) > 15*time.Minute {
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

		mu.Lock()
		v, exists := attempts[ip]

		if !exists {
			attempts[ip] = &visitor{count: 1, lastSeen: time.Now()}
			mu.Unlock()
			c.Next()
			return
		}

		if time.Since(v.lastSeen) > 15*time.Minute {
			v.count = 1
			v.lastSeen = time.Now()
			mu.Unlock()
			c.Next()
			return
		}

		v.count++
		v.lastSeen = time.Now()
		mu.Unlock()

		// Block after 50 attempts in 15 minutes
		if v.count > 50 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "تم تجاوز عدد محاولات تسجيل الدخول. حاول مرة أخرى بعد 15 دقيقة",
			})
			return
		}

		c.Next()
	}
}
