package ratelimiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type LimiterConfig struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

var (
	limiters = make(map[string]*LimiterConfig)
	mu       sync.Mutex
)

// GetLimiter returns a rate limiter for a given key (IP, user ID, etc.)
// rateLimit is in requests per second, burst is the burst size
func GetLimiter(key string, rateLimit rate.Limit, burst int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	if config, exists := limiters[key]; exists {
		config.LastSeen = time.Now()
		return config.Limiter
	}

	limiter := rate.NewLimiter(rateLimit, burst)
	limiters[key] = &LimiterConfig{
		Limiter:  limiter,
		LastSeen: time.Now(),
	}

	return limiter
}
