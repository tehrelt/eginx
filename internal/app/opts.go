package app

import (
	"context"
	"errors"
	"net/http"

	"github.com/tehrelt/eginx/internal/router"
)

type LimiterPool interface {
	Allow(context.Context, string) bool
}

type GetKeyFn func(r *http.Request) string

var errTooManyRequests = errors.New("too many requests")

func WithLimiter(limiter LimiterPool, getKey GetKeyFn) AppOptFn {
	return func(a *App) {
		a.router.Use(
			func(w http.ResponseWriter, r *http.Request) error {
				key := getKey(r)

				if !limiter.Allow(r.Context(), key) {
					return router.NewError(errTooManyRequests, http.StatusTooManyRequests)
				}

				return nil
			},
		)
	}
}
