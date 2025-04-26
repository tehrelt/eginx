package limiter

import (
	"context"
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
	limiter, ok := lp.pool[key]
	if !ok {
		rate, err := lp.storage.Get(ctx, key)
		if err != nil {
			if lp.defaultLimiter == nil {
				return nil
			}
			return lp.defaultLimiter
		}

		limiter = lp.create(lp.ctx, rate)
		lp.pool[key] = limiter
	}

	return limiter
}
