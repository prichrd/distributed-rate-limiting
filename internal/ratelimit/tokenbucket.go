package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBucket implements the token bucket rate limiting pattern. The concept
// is simple, a bucket contains tokens and every request costs a token. On a
// regular cadence, the tokens get refilled. If a bucket runs out of token,
// requests are rate-limited. If we want to rate-limit by user ID, we need one
// bucket per user ID.
//
// This approach is a very simple one to implement and is relatively efficient
// on memory. The only downside to it is that it doesn't prevent race conditions.
// If there would be only one token left in a bucket and two requests are made
// simultaneously, we could allow 2 calls, when in fact, we should have only
// accepted one. To make transactions atomic, we could use redis locks, but it
// would impact performance.
type TokenBucket struct {
	bucketConfig *TokenBucketConfig
	rdb          *redis.Client
}

// TokenBucketConfig represents the configuration of a token bucket.
type TokenBucketConfig struct {
	Capacity  int64
	ResetRate time.Duration
}

// NewTokenBucket returns a configured instance of TokenBucket.
func NewTokenBucket(bucketConfig *TokenBucketConfig, rdb *redis.Client) *TokenBucket {
	return &TokenBucket{
		bucketConfig: bucketConfig,
		rdb:          rdb,
	}
}

// Check updates the state of the rate limiter and checks if the request is rate
// limited or not. In cases where the request is rate limited, the function will
// return a `ratelimit.ErrRateLimited` error.
func (r *TokenBucket) Check(ctx context.Context, key string, reqID string, at time.Time) error {
	v, _ := r.rdb.Get(ctx, fmt.Sprintf("%s_last_reset", key)).Result()
	lastResetSec, _ := strconv.ParseInt(v, 10, 64)
	lastResetTs := time.Unix(lastResetSec, 0)

	if at.Sub(lastResetTs) >= r.bucketConfig.ResetRate {
		// If the delta between the last reset and the current time is greater than the
		// configured refill reate, we need to reset the bucket.
		_, err := r.rdb.Set(ctx, fmt.Sprintf("%s_counter", key), strconv.FormatInt(r.bucketConfig.Capacity, 10), 0).Result()
		if err != nil {
			return fmt.Errorf("reseting bucket: %w", err)
		}

		_, err = r.rdb.Set(ctx, fmt.Sprintf("%s_last_reset", key), strconv.FormatInt(int64(at.Unix()), 10), 0).Result()
		if err != nil {
			return fmt.Errorf("reseting bucket: %w", err)
		}
	} else {
		// If we are reseting the counter, we validate that it is not at 0 capacity. If so,
		// we are rate limiting.
		v, _ := r.rdb.Get(ctx, fmt.Sprintf("%s_counter", key)).Result()
		tokens, _ := strconv.ParseInt(v, 10, 64)
		if tokens <= 0 {
			return ErrRateLimited
		}
	}

	if _, err := r.rdb.Decr(ctx, fmt.Sprintf("%s_counter", key)).Result(); err != nil {
		return fmt.Errorf("decrementing bucket counter: %w", err)
	}
	return nil
}
