package ratelimit

import (
	"context"
	"time"
)

// RateLimiter is implemented by types able to run rate limiting checks.
type RateLimiter interface {
	// Check updates the state of the rate limiter and checks if the request is rate
	// limited or not. In cases where the request is rate limited, the function will
	// return a `ratelimit.ErrRateLimited` error.
	Check(ctx context.Context, key string, at time.Time) error
}
