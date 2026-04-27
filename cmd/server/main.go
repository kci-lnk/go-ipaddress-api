package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"github.com/kci-lnk/ipaddress-api/internal/cache"
	"github.com/kci-lnk/ipaddress-api/internal/config"
	"github.com/kci-lnk/ipaddress-api/internal/handler"
	"github.com/kci-lnk/ipaddress-api/internal/ipdata"
	"github.com/kci-lnk/ipaddress-api/internal/middleware"
	"github.com/kci-lnk/ipaddress-api/internal/ratelimit"
)

const Version = "1.0.0"

var BuildTime string
var GitCommit string

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("BuildTime: %s\n", BuildTime)
		fmt.Printf("GitCommit: %s\n", GitCommit)
		os.Exit(0)
	}

	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	if err := Run(*configPath); err != nil {
		log.Fatal(err)
	}
}

func Run(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize structured logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(logHandler))

	// Initialize IP data
	ipData, err := ipdata.New("ipdata")
	if err != nil {
		return fmt.Errorf("failed to initialize IP data: %w", err)
	}
	ipdata.Set(ipData)

	// Initialize Redis with timeouts
	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		DialTimeout:  time.Duration(cfg.Redis.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.Redis.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Redis.WriteTimeout) * time.Second,
	})

	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Initialize cache
	redisCache := cache.New(redisClient, cfg.Cache.TTL)

	// Initialize rate limiter
	rateLimiter := ratelimit.New(
		redisClient,
		cfg.RateLimit.Enabled,
		cfg.RateLimit.RequestsPerSecond,
		cfg.RateLimit.Burst,
	)

	// Initialize handler
	h := handler.New(ipData, redisCache, rateLimiter)

	// Setup Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Body size limit (4KB for query params)
	const maxQueryParamsSize = 4 * 1024
	r.Use(middleware.MaxBodySize(maxQueryParamsSize))

	// Trust proxy middleware
	r.Use(middleware.TrustProxy(middleware.TrustProxyConfig{
		Enabled:       cfg.TrustProxy.Enabled,
		RealIPHeader:  cfg.TrustProxy.RealIPHeader,
		RealIPHeaders: cfg.TrustProxy.RealIPHeaders,
	}))

	// Rate limit middleware
	r.Use(middleware.RateLimit(rateLimiter))

	// Routes
	r.GET("/health", h.Health)
	r.GET("/api/v1/ip/lookup", h.Lookup)

	// Create HTTP server with timeouts
	srv := &http.Server{
		Addr:           cfg.Server.Address(),
		Handler:        r,
		ReadTimeout:    time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:    time.Duration(cfg.Server.IdleTimeout) * time.Second,
		MaxHeaderBytes: cfg.Server.MaxHeaderBytes,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("server started", "address", cfg.Server.Address())

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("server exited")

	return nil
}
