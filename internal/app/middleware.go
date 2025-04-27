package app

import (
	"net/http"

	"github.com/tehrelt/eginx/internal/router"
)

func limiterMiddleware(cfg *limiterHeader) router.HandlerFn {
	return func(w http.ResponseWriter, r *http.Request) error {
		key := r.Header.Get(cfg.header)
		if !cfg.Allow(r.Context(), key) {
			return router.NewError(errTooManyRequests, http.StatusTooManyRequests)
		}

		return nil
	}
}
