package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

type TrustProxyConfig struct {
	Enabled       bool
	RealIPHeader  string
	RealIPHeaders []string
}

func TrustProxy(cfg TrustProxyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		ip := getRealIP(c, cfg)
		if ip != "" {
			c.Request.RemoteAddr = ip + ":0"
		}

		c.Next()
	}
}

func getRealIP(c *gin.Context, cfg TrustProxyConfig) string {
	for _, header := range cfg.RealIPHeaders {
		ip := c.GetHeader(header)
		if ip == "" {
			continue
		}

		if header == "X-Forwarded-For" {
			parts := strings.Split(ip, ",")
			ip = strings.TrimSpace(parts[0])
		}

		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	if cfg.RealIPHeader != "" {
		ip := c.GetHeader(cfg.RealIPHeader)
		if ip != "" && net.ParseIP(ip) != nil {
			return ip
		}
	}

	return ""
}
