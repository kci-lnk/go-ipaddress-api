package handler

import (
	"log/slog"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kci-lnk/ipaddress-api/internal/cache"
	"github.com/kci-lnk/ipaddress-api/internal/ipdata"
	"github.com/kci-lnk/ipaddress-api/internal/ratelimit"
	"github.com/kci-lnk/ipaddress-api/pkg/response"
)

type Handler struct {
	ipData *ipdata.IPData
	cache  *cache.Cache
	rl     *ratelimit.RateLimiter
}

func New(ip *ipdata.IPData, c *cache.Cache, rl *ratelimit.RateLimiter) *Handler {
	return &Handler{
		ipData: ip,
		cache:  c,
		rl:     rl,
	}
}

func (h *Handler) Lookup(c *gin.Context) {
	ip := c.Query("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, response.InvalidIP(ip))
		return
	}

	// Validate IP format
	if net.ParseIP(ip) == nil {
		c.JSON(http.StatusBadRequest, response.InvalidIP(ip))
		return
	}

	// Check cache first
	if h.cache != nil {
		cached, err := h.cache.Get(c.Request.Context(), ip)
		if err == nil && cached != nil {
			slog.Debug("cache hit", "ip", ip)
			c.Data(http.StatusOK, "application/json", cached)
			return
		}
	}

	// Lookup IP
	result, err := h.ipData.Lookup(ip)
	if err != nil {
		slog.Error("IP lookup failed", "ip", ip, "error", err)
		c.JSON(http.StatusInternalServerError, response.InternalError(ip))
		return
	}

	if result == nil {
		c.JSON(http.StatusNotFound, response.NotFound(ip))
		return
	}

	ipResult := &response.IpResult{
		Version:     result.Version,
		Continent:   result.Continent,
		Country:     result.Country,
		Province:    result.Province,
		City:        result.City,
		District:    result.District,
		Isp:         result.Isp,
		CountryCode: result.CountryCode,
		Fields:      result.Fields,
		Raw:         result.Raw,
	}

	resp := response.Success(ip, ipResult)

	// Cache the result
	if h.cache != nil {
		h.cache.Set(c.Request.Context(), ip, resp)
	}

	slog.Info("IP lookup success", "ip", ip, "country", result.Country, "city", result.City)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Health(c *gin.Context) {
	status := gin.H{"status": "ok"}

	if h.cache != nil {
		if err := h.cache.Ping(c.Request.Context()); err != nil {
			status["status"] = "degraded"
			status["redis"] = "unavailable"
			slog.Warn("health check: redis unavailable", "error", err)
		}
	}

	c.JSON(http.StatusOK, status)
}
