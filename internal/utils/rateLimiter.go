package utils

import (
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/gin-gonic/gin"
)

var (
	limiterMap = make(map[string]*rate.Limiter)
	mu         sync.Mutex 
)

const rateLimit = 1000

func UnauthenticatedRateLimiterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		limiter := getLimiter(clientIP)

		if !limiter.Allow() {
			c.JSON(429, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getLimiter(clientIP string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if limiter, exists := limiterMap[clientIP]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(rateLimit)), rateLimit)
	limiterMap[clientIP] = limiter
	return limiter
}
