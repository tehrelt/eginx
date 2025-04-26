package tokenbucket

import (
	"context"
	"time"
)

type tokenBucketLimiter struct {
	bucket chan struct{}
}

func NewTokenBucket(ctx context.Context, capacity int, period time.Duration) *tokenBucketLimiter {
	l := &tokenBucketLimiter{
		bucket: make(chan struct{}, capacity),
	}

	for range capacity {
		l.bucket <- struct{}{}
	}

	interval := time.Duration(period.Nanoseconds() / int64(capacity))
	go l.replenishCycle(ctx, interval)

	return l
}

func (l *tokenBucketLimiter) replenishCycle(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			select {
			case l.bucket <- struct{}{}:
			default:
			}
		}
	}
}

func (l *tokenBucketLimiter) Allow() bool {
	select {
	case <-l.bucket:
		return true
	default:
		return false
	}
}
