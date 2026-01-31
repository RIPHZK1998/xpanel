package middleware

import (
	"context"
	"fmt"
	"time"

	"xpanel/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter creates a rate limiting middleware using Redis.
type RateLimiter struct {
	redisClient *redis.Client
	maxRequests int
	windowSec   int
}

// NewRateLimiter creates a new rate limiter middleware.
func NewRateLimiter(redisClient *redis.Client, maxRequests, windowSec int) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		maxRequests: maxRequests,
		windowSec:   windowSec,
	}
}

// Middleware returns the rate limiting middleware handler.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		ip := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s", ip)

		ctx := context.Background()

		// Get current count
		count, err := rl.redisClient.Incr(ctx, key).Result()
		if err != nil {
			// Fail open - allow request if Redis is unavailable
			c.Next()
			return
		}

		// Set expiration on first request in window
		if count == 1 {
			rl.redisClient.Expire(ctx, key, time.Duration(rl.windowSec)*time.Second)
		}

		// Check if limit exceeded
		if int(count) > rl.maxRequests {
			// Get TTL for Retry-After header
			ttl, _ := rl.redisClient.TTL(ctx, key).Result()
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
			response.TooManyRequests(c, "rate limit exceeded")
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.maxRequests-int(count)))

		c.Next()
	}
}
