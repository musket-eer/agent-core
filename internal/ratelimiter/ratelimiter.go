package ratelimiter

import (
	"context"
	"net/http"

	"github.com/ulule/limiter/v3"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

// RedisLimiter holds the shared limiter instance
type RedisLimiter struct {
	Limiter *limiter.Limiter
}

// NewRedisLimiter initializes a new Redis-backed rate limiter using the global Redis client
func NewRedisLimiter(rateStr string) (*RedisLimiter, error) {
	// e.g., "10-S" for 10 requests per second
	rateParsed, err := limiter.NewRateFromFormatted(rateStr)
	if err != nil {
		return nil, err
	}

	redisClient := GetClient() // use your central Redis client

	store, err := redisstore.NewStoreWithOptions(redisClient, limiter.StoreOptions{
		Prefix:   "rate_limiter",
		MaxRetry: 3,
	})
	if err != nil {
		return nil, err
	}

	return &RedisLimiter{
		Limiter: limiter.New(store, rateParsed),
	}, nil
}

// Allow checks if the request should be allowed under rate limits
func (rl *RedisLimiter) Allow(r *http.Request, key string) (bool, limiter.Context, error) {
	ctx := context.Background()
	limiterCtx, err := rl.Limiter.Get(ctx, key)
	if err != nil {
		return false, limiterCtx, err
	}
	return limiterCtx.Reached == false, limiterCtx, nil
}
