package tokenbucket

import (
	"context"
	"log/slog"
	"time"
)

type bucket struct {
	bucket chan struct{}
	rpm    int
}

func NewBucket(ctx context.Context, rpm int) *bucket {
	l := &bucket{
		bucket: make(chan struct{}, rpm),
		rpm:    rpm,
	}

	for range rpm {
		l.bucket <- struct{}{}
	}

	go l.replenishCycle(ctx, time.Duration(time.Minute.Nanoseconds()/int64(rpm)))

	return l
}

func (l *bucket) replenishCycle(ctx context.Context, interval time.Duration) {
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

func (l *bucket) Allow() bool {
	select {
	case <-l.bucket:
		slog.Debug("acquiring token")
		return true
	default:
		return false
	}
}
func (l *bucket) Capacity() int {
	return l.rpm
}
