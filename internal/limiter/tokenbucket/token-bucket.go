package tokenbucket

import (
	"context"
	"log/slog"
	"time"
)

type tokenBucket struct {
	bucket   chan struct{}
	capacity int
}

func New(ctx context.Context, ratePerSec int) *tokenBucket {

	capacity := ratePerSec
	l := &tokenBucket{
		bucket:   make(chan struct{}, capacity),
		capacity: capacity,
	}

	for range capacity {
		l.bucket <- struct{}{}
	}

	go l.replenishCycle(ctx, time.Duration(time.Second.Nanoseconds()/int64(ratePerSec)))

	return l
}

func (l *tokenBucket) replenishCycle(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	slog.Debug("replenish cycle started", slog.Int64("interval_ms", interval.Milliseconds()))

	for {
		select {
		case <-ctx.Done():
			slog.Debug("limiter closed")
			return
		case <-t.C:
			select {
			case l.bucket <- struct{}{}:
				slog.Debug("replenish bucket")
			default:
			}
		}
	}
}

func (l *tokenBucket) Allow() bool {
	select {
	case <-l.bucket:
		slog.Debug("acquiring token")
		return true
	default:
		return false
	}
}
func (l *tokenBucket) Capacity() int {
	return l.capacity
}
