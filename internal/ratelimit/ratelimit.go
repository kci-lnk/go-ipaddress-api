package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// GlobalRedisKeyPrefix is the fixed prefix for all Redis keys
	GlobalRedisKeyPrefix = "ipaddress:"
	RateLimitKeySuffix  = "ratelimit:"
)

type RateLimiter struct {
	client  *redis.Client
	enabled bool
	rps     int
	burst   int
	prefix  string
}

func New(client *redis.Client, enabled bool, rps, burst int) *RateLimiter {
	return &RateLimiter{
		client:  client,
		enabled: enabled,
		rps:     rps,
		burst:   burst,
		prefix:  GlobalRedisKeyPrefix + RateLimitKeySuffix,
	}
}

func (rl *RateLimiter) key(ip string) string {
	return fmt.Sprintf("%s%s", rl.prefix, ip)
}

func (rl *RateLimiter) Allow(ctx context.Context, ip string) (bool, error) {
	if !rl.enabled || rl.rps <= 0 {
		return true, nil
	}

	key := rl.key(ip)

	luaScript := `
		local key = KEYS[1]
		local rps = tonumber(ARGV[1])
		local burst = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local window = 1.0 / rps

		local current = redis.call('GET', key)
		if current then
			local last = tonumber(current)
			local tokens = redis.call('GET', key .. ':tokens')
			if not tokens then
				tokens = burst
			else
				tokens = tonumber(tokens)
			end

			local elapsed = now - last
			local refill = elapsed / window
			tokens = math.min(burst, tokens + refill)

			if tokens >= 1 then
				tokens = tokens - 1
				redis.call('SET', key, now, 'EX', 10)
				redis.call('SET', key .. ':tokens', tokens, 'EX', 10)
				return 1
			else
				redis.call('SET', key .. ':tokens', tokens, 'EX', 10)
				return 0
			end
		else
			redis.call('SET', key, now, 'EX', 10)
			redis.call('SET', key .. ':tokens', burst - 1, 'EX', 10)
			return 1
		end
	`

	now := float64(time.Now().UnixNano()) / 1e9
	result, err := rl.client.Eval(ctx, luaScript, []string{key}, rl.rps, rl.burst, now).Int()
	if err != nil {
		return true, err
	}

	return result == 1, nil
}
