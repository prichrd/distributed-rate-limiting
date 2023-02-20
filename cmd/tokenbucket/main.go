package main

import (
	"context"
	"sync"
	"time"

	"github.com/prichrd/distributed-rate-limiting/internal/procutil"
	"github.com/prichrd/distributed-rate-limiting/internal/ratelimit"
	"github.com/prichrd/distributed-rate-limiting/internal/test"
	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// We initialize the rate limiter with 10 req per 10 sec, effectively
	// allowing 1 req/s with a refresh every 10 seconds.
	bc := ratelimit.NewTokenBucket(&ratelimit.TokenBucketConfig{
		Capacity:  10,
		ResetRate: 10 * time.Second,
	}, rdb)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Thread 1 should not trip rate limiting, as it is doing 1 request per second.
	wg.Add(1)
	go test.CheckThread(ctx, &wg, bc, 1, time.Second)

	// Thread 2 should trip rate limiting, as it is doing 2 request per second.
	wg.Add(1)
	go test.CheckThread(ctx, &wg, bc, 2, 500*time.Millisecond)

	procutil.WaitForInterrupt()
	cancel()
	wg.Wait()
}
