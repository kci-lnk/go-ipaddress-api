package middleware

import (
	"context"
	"net"

	"github.com/gin-gonic/gin"
)

type RateLimiter interface {
	Allow(ctx context.Context, ip string) (bool, error)
}

func RateLimit(rl RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := GetClientIP(c)

		allowed, err := rl.Allow(c.Request.Context(), ip)
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			c.JSON(429, gin.H{
				"code":   4290,
				"msg":    "rate limit exceeded",
				"ip":     ip,
				"result": nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetClientIP(c *gin.Context) string {
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}
