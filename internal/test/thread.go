package test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prichrd/distributed-rate-limiting/internal/ratelimit"
)

// CheckThread runs fake queries against a `ratelimit.RateLimiter`.
func CheckThread(ctx context.Context, wg *sync.WaitGroup, bc ratelimit.RateLimiter, tid int, sleep time.Duration) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			now := time.Now()
			rid := time.Now().Format(time.RFC3339)
			msg := "[thread %d] %s check: %s\n"
			if err := bc.Check(ctx, fmt.Sprintf("%d", tid), rid, now); err != nil {
				fmt.Printf(msg, tid, now.Format(time.RFC3339), err)
			} else {
				fmt.Printf(msg, tid, now.Format(time.RFC3339), "successful")
			}
			time.Sleep(sleep)
		}
	}
}
