package middlewares

import (
	"sync"
	"time"

	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type loginRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewLoginRateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := &loginRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		limiter.mu.Lock()
		defer limiter.mu.Unlock()

		requests := limiter.requests[ip]
		validRequests := requests[:0]
		for _, t := range requests {
			if now.Sub(t) < limiter.window {
				validRequests = append(validRequests, t)
			}
		}

		if len(validRequests) >= limiter.limit {
			response.Error(c, 429, "LOGIN_RATE_LIMIT_EXCEEDED", "登录尝试过于频繁，请稍后再试")
			c.Abort()
			return
		}

		validRequests = append(validRequests, now)
		limiter.requests[ip] = validRequests
		c.Next()
	}
}

func LoginRateLimitMiddleware() gin.HandlerFunc {
	return NewLoginRateLimitMiddleware(10, time.Minute)
}
