package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prichrd/distributed-rate-limiting/internal/ratelimit"
	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// We initialize the rate limiter with 10 req per 10 sec, effectively
	// allowing 1 req/s with a refresh every 10 seconds.
	bc := ratelimit.NewTokenBucket(&ratelimit.BucketConfig{
		Capacity:  10,
		ResetRate: 10 * time.Second,
	}, rdb)

	ctx := context.Background()

	var wg sync.WaitGroup

	// Thread 1 should not trip rate limiting, as it is doing 1 request per second.
	wg.Add(1)
	go tokenCheckThread(ctx, &wg, bc, 1, time.Second)

	// Thread 2 should trip rate limiting, as it is doing 2 request per second.
	wg.Add(1)
	go tokenCheckThread(ctx, &wg, bc, 2, 500*time.Millisecond)

	wg.Wait()
}

func tokenCheckThread(ctx context.Context, wg *sync.WaitGroup, bc ratelimit.RateLimiter, tid int, sleep time.Duration) {
	defer wg.Done()
	for {
		now := time.Now()
		msg := "[thread %d] %s check: %s\n"
		if err := bc.Check(ctx, fmt.Sprintf("%d", tid), now); err != nil {
			fmt.Printf(msg, tid, now.Format(time.RFC3339), err)
		} else {
			fmt.Printf(msg, tid, now.Format(time.RFC3339), "successful")
		}
		time.Sleep(sleep)
	}
}
