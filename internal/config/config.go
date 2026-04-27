package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
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
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Mode           string `mapstructure:"mode"`
	ReadTimeout    int    `mapstructure:"read_timeout"`
	WriteTimeout   int    `mapstructure:"write_timeout"`
	IdleTimeout    int    `mapstructure:"idle_timeout"`
	MaxHeaderBytes int    `mapstructure:"max_header_bytes"`
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	DialTimeout  int    `mapstructure:"dial_timeout"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type RateLimitConfig struct {
	Enabled          bool   `mapstructure:"enabled"`
	RequestsPerSecond int   `mapstructure:"requests_per_second"`
	Burst            int    `mapstructure:"burst"`
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

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
