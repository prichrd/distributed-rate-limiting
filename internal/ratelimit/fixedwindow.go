package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// FixedWindow implements a fixed window rate limiting pattern. It creates and increments
// request counts for fixed time windows. If for instance the configuration specifies one
// window per minute, requests will get bucketed by minutes. If we burst the capacity for
// a given minute, the request is rejected.
//
// This approach is very simple to implement. Because it doesn't use a sliding window, this
// algorithm can become problematic when dealing with bursts of traffic. It can for instance
// allow a user to do 20 consecutive requests if they are made close to the configured
// duration gap.
type FixedWindow struct {
	windowConfig *FixedWindowConfig
	rdb          *redis.Client
}

// FixedWindowConfig represents the configuration of a fixed window.
type FixedWindowConfig struct {
	Duration time.Duration
	Capacity int64
}

// NewFixedWindow returns a configured instance of FixedWindow.
func NewFixedWindow(windowConfig *FixedWindowConfig, rdb *redis.Client) *FixedWindow {
	return &FixedWindow{
		windowConfig: windowConfig,
		rdb:          rdb,
	}
}

// Check updates the state of the rate limiter and checks if the request is rate
// limited or not. In cases where the request is rate limited, the function will
// return a `ratelimit.ErrRateLimited` error.
func (r *FixedWindow) Check(ctx context.Context, key string, reqID string, at time.Time) error {
	tt := at.Truncate(r.windowConfig.Duration)
	nkey := fmt.Sprintf("%s_%d", key, tt.Unix())

	v, _ := r.rdb.Get(ctx, nkey).Result()
	count, _ := strconv.ParseInt(v, 10, 64)
	if count == 0 {
		// If the count doesn't exist, we create it with an expiration period that's 5 times longer
		// than the configured lifetime of a bucket.
		r.rdb.Set(ctx, nkey, 1, 5*r.windowConfig.Duration)
		return nil
	}

	if count >= r.windowConfig.Capacity {
		return ErrRateLimited
	}

	if err := r.rdb.Incr(ctx, nkey).Err(); err != nil {
		return fmt.Errorf("incrementing time bucket value: %w", err)
	}

	return nil
}
