package main

import (
	"context"
	"fmt"
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
		DB:   1,
	})

	// We initialize the rate limiter with a maximum capacity of 10
	// queued requests.
	bc := ratelimit.NewLeakyBucket(&ratelimit.LeakyBucketConfig{
		Capacity: 10,
	}, rdb)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	// Thread 1 should not trip rate limiting, as it is doing 1 request per second,
	// and we are dequeueing at the same rate.
	wg.Add(2)
	go test.CheckThread(ctx, &wg, bc, 1, time.Second)
	go dequeueElements(ctx, &wg, rdb, 1, time.Second)

	// Thread 2 should trip rate limiting, as it is doing 2 request per second, and
	// we are dequeueing at a rate of 1 per second.
	wg.Add(2)
	go test.CheckThread(ctx, &wg, bc, 2, 500*time.Millisecond)
	go dequeueElements(ctx, &wg, rdb, 2, time.Second)

	procutil.WaitForInterrupt()
	cancel()
	wg.Wait()
}

func dequeueElements(ctx context.Context, wg *sync.WaitGroup, rdb *redis.Client, tid int, sleep time.Duration) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			rdb.RPop(ctx, fmt.Sprintf("%d", tid))
			time.Sleep(sleep)
		}
	}
}
