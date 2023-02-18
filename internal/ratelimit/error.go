package ratelimit

import "errors"

var ErrRateLimited = errors.New("rate limited")
