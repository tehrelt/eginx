package limiter

import (
	"context"
	"sync"
)

type KeyStorage interface {
	Get(context.Context, string) (int, error)
}

type LimiterCreatorFn func(ctx context.Context, ratePerSec int) Limiter

type limiterPool struct {
	ctx            context.Context
	pool           map[string]Limiter
	storage        KeyStorage
	create         LimiterCreatorFn
	defaultLimiter Limiter
	m              sync.Mutex
}

type LimiterPoolOpt func(*limiterPool)

func NewLimiterPool(ctx context.Context, storage KeyStorage, create LimiterCreatorFn, opts ...LimiterPoolOpt) *limiterPool {
	lp := &limiterPool{
		ctx:     ctx,
		pool:    make(map[string]Limiter),
		storage: storage,
		create:  create,
	}
	for _, opt := range opts {
		opt(lp)
	}
	return lp
}

func (lp *limiterPool) Allow(ctx context.Context, key string) bool {
	limiter := lp.get(ctx, key)
	if limiter == nil {
		return false
	}

	return limiter.Allow()
}

func (lp *limiterPool) get(ctx context.Context, key string) Limiter {
	lp.m.Lock()
	limiter, ok := lp.pool[key]
	lp.m.Unlock()

	if !ok {
		rate, err := lp.storage.Get(ctx, key)
		if err != nil {
			if lp.defaultLimiter == nil {
				return nil
			}

			return lp.defaultLimiter
		}

		limiter = lp.create(lp.ctx, rate)
		lp.m.Lock()
		lp.pool[key] = limiter
		lp.m.Unlock()
	} else {
		rate, err := lp.storage.Get(ctx, key)
		if err != nil {
			return lp.defaultLimiter
		}

		if rate != limiter.Capacity() {
			limiter = lp.create(lp.ctx, rate)
			lp.m.Lock()
			lp.pool[key] = limiter
			lp.m.Unlock()
		}
	}

	return limiter
}
