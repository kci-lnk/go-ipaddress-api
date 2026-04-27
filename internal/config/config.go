package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Redis      RedisConfig      `mapstructure:"redis"`
	RateLimit  RateLimitConfig  `mapstructure:"ratelimit"`
	Cache      CacheConfig      `mapstructure:"cache"`
	TrustProxy TrustProxyConfig `mapstructure:"trust_proxy"`
	Log        LogConfig        `mapstructure:"log"`
}

type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	Mode           string        `mapstructure:"mode"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type RateLimitConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerSecond int  `mapstructure:"requests_per_second"`
	Burst             int  `mapstructure:"burst"`
}

type CacheConfig struct {
	TTL int `mapstructure:"ttl"`
}

type TrustProxyConfig struct {
	Enabled       bool     `mapstructure:"enabled"`
	RealIPHeader  string   `mapstructure:"real_ip_header"`
	RealIPHeaders []string `mapstructure:"real_ip_headers"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			return duration
		}
	}
	return defaultVal
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:           getEnv("SERVER__HOST", "0.0.0.0"),
			Port:           getEnvInt("SERVER__PORT", 30661),
			Mode:           getEnv("GIN_MODE", "release"),
			ReadTimeout:    getEnvDuration("SERVER__READ_TIMEOUT", 10*time.Second),
			WriteTimeout:   getEnvDuration("SERVER__WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:    getEnvDuration("SERVER__IDLE_TIMEOUT", 60*time.Second),
			MaxHeaderBytes: getEnvInt("SERVER__MAX_HEADER_BYTES", 4096),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS__HOST", "localhost"),
			Port:         getEnvInt("REDIS__PORT", 6379),
			Password:     getEnv("REDIS__PASSWORD", ""),
			DB:           getEnvInt("REDIS__DB", 0),
			PoolSize:     getEnvInt("REDIS__POOL_SIZE", 10),
			DialTimeout:  getEnvDuration("REDIS__DIAL_TIMEOUT", 5*time.Second),
			ReadTimeout:  getEnvDuration("REDIS__READ_TIMEOUT", 3*time.Second),
			WriteTimeout: getEnvDuration("REDIS__WRITE_TIMEOUT", 3*time.Second),
		},
		RateLimit: RateLimitConfig{
			Enabled:           getEnv("RATELIMIT__ENABLED", "true") == "true",
			RequestsPerSecond: getEnvInt("RATELIMIT__REQUESTS_PER_SECOND", 100),
			Burst:             getEnvInt("RATELIMIT__BURST", 200),
		},
		Cache: CacheConfig{
			TTL: getEnvInt("CACHE__TTL", 3600),
		},
		TrustProxy: TrustProxyConfig{
			Enabled:       getEnv("TRUST_PROXY__ENABLED", "true") == "true",
			RealIPHeader:  getEnv("TRUST_PROXY__REAL_IP_HEADER", "X-Real-IP"),
			RealIPHeaders: []string{"X-Real-IP", "X-Forwarded-For", "CF-Connecting-IP"},
		},
		Log: LogConfig{
			Level:  getEnv("LOG__LEVEL", "info"),
			Format: getEnv("LOG__FORMAT", "json"),
		},
	}

	return cfg, nil
}
