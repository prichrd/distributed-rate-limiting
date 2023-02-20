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

	// We initialize the rate limiter with 5 requests per second buckets.
	bc := ratelimit.NewFixedWindow(&ratelimit.FixedWindowConfig{
		Duration: time.Second,
		Capacity: 5,
	}, rdb)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Thread 1 should not trip rate limiting, as it is doing 1 request per second.
	wg.Add(1)
	go test.CheckThread(ctx, &wg, bc, 1, time.Second)

	// Thread 2 should trip rate limiting, as it is doing 10 request per second.
	wg.Add(1)
	go test.CheckThread(ctx, &wg, bc, 2, 100*time.Millisecond)

	procutil.WaitForInterrupt()
	cancel()
	wg.Wait()
}
