package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// LeakyBucket implements a leaky bucket rate limiting pattern. Just like a token
// bucket, the leaky bucket has a capacity. The capacity represents the queue of
// requests to process. If the process rate of messages isn't fast enough, messages
// leak from the bucket.
//
// This approach can be good for requests expected at a fixed rate. It is memory
// efficient as it only requires bookkeeping for the queue of requests. The approach
// can become problematic when dealing with bursts of requests, the bucket can fill
// up quickly with requests and drop newer ones.
type LeakyBucket struct {
	bucketConfig *LeakyBucketConfig
	rdb          *redis.Client
}

// LeakyBucketConfig represents the configuration of a leaky bucket.
type LeakyBucketConfig struct {
	Capacity int64
}

// NewLeakyBucket returns a configured instance of LeakyBucket.
func NewLeakyBucket(bucketConfig *LeakyBucketConfig, rdb *redis.Client) *LeakyBucket {
	return &LeakyBucket{
		bucketConfig: bucketConfig,
		rdb:          rdb,
	}
}

// Check checks the length of the request queue. If the number of elements is bigger
// than the capacity of the leaky bucket, we drop the request by returning a
// `ratelimit.ErrRateLimited` error.
func (r *LeakyBucket) Check(ctx context.Context, key string, reqID string, at time.Time) error {
	rc := r.rdb.LLen(ctx, key).Val()
	if rc >= r.bucketConfig.Capacity {
		return ErrRateLimited
	}

	r.rdb.RPush(ctx, key, reqID)
	return nil
}
