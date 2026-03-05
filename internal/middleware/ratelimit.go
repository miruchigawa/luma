package middleware

import (
	"fmt"
	"sync"
	"time"

	"luma/internal/command"
)

type tokenBucket struct {
	tokens     int
	lastRefill time.Time
}

// RateLimiter implements a simple sliding-window token bucket per user.
type RateLimiter struct {
	buckets    map[string]*tokenBucket
	mu         sync.Mutex
	maxTokens  int
	refillTime time.Duration
}

// NewRateLimiter creates a rate limiter (e.g., config: max 5 requests per 10 seconds).
func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets:    make(map[string]*tokenBucket),
		maxTokens:  maxRequests,
		refillTime: window,
	}
}

// Middleware returns the middleware function to be chained.
func (rl *RateLimiter) Middleware() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *command.Context) error {
			if !ctx.Msg.IsCommand {
				return next(ctx) // don't rate limit normal messages
			}

			sender := ctx.Msg.From.ToNonAD().String()

			rl.mu.Lock()
			bucket, exists := rl.buckets[sender]
			now := time.Now()

			if !exists {
				bucket = &tokenBucket{
					tokens:     rl.maxTokens,
					lastRefill: now,
				}
				rl.buckets[sender] = bucket
			}

			// Refill tokens logic
			if now.Sub(bucket.lastRefill) >= rl.refillTime {
				bucket.tokens = rl.maxTokens
				bucket.lastRefill = now
			}

			if bucket.tokens > 0 {
				bucket.tokens--
				rl.mu.Unlock()
				return next(ctx)
			}

			timeLeft := rl.refillTime - now.Sub(bucket.lastRefill)
			rl.mu.Unlock()

			seconds := int(timeLeft.Seconds())
			if seconds < 1 {
				seconds = 1
			}

			msg := fmt.Sprintf("⏳ Slow down! Try again in %ds.", seconds)
			_ = ctx.Reply(msg)
			return nil
		}
	}
}
