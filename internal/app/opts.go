package app

import (
	"context"
	"errors"
)

type Limiter interface {
	Allow(context.Context, string) bool
}

type limiterHeader struct {
	Limiter
	header string
}

var errTooManyRequests = errors.New("too many requests")

func WithLimiter(limiter Limiter, header string) AppOptFn {
	return func(a *App) {
		a.limiter = &limiterHeader{
			Limiter: limiter,
			header:  header,
		}
	}
}
