package app

import (
	"fmt"
	"net/http"

	"github.com/tehrelt/eginx/internal/router"
)

func serverNameMiddleware(version string) router.HandlerFn {
	val := fmt.Sprintf("%s/%s", serverName, version)
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Server", val)
		return nil
	}
}

func limiterMiddleware(cfg *limiterHeader) router.HandlerFn {
	return func(w http.ResponseWriter, r *http.Request) error {
		key := r.Header.Get(cfg.header)
		if !cfg.Allow(r.Context(), key) {
			return router.NewError(errTooManyRequests, http.StatusTooManyRequests)
		}

		return nil
	}
}
