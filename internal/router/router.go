package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

const (
	defaultPort = 5000
)

type ErrorHandlerFn func(http.ResponseWriter, *http.Request, *ServerError)

var (
	defaultHandler func(*Router) http.Handler = func(router *Router) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := router.handle(w, r); err != nil {
				var serr *ServerError
				if !errors.As(err, &serr) {
					serr = NewError(err, http.StatusInternalServerError)
				}

				router.errorHandler(w, r, serr)
			}
		})
	}

	defaultErrorHandler ErrorHandlerFn = func(w http.ResponseWriter, r *http.Request, err *ServerError) {
		err.Write(w)
	}
)

type Router struct {
	middlewares  []HandlerFn
	server       *http.Server
	errorHandler ErrorHandlerFn
}

type Config struct {
	Host         string
	Port         int
	ErrorHandler ErrorHandlerFn
}

func (rc Config) Address() string {
	if rc.Port == 0 {
		rc.Port = defaultPort
	}

	return fmt.Sprintf("%s:%d", rc.Host, rc.Port)
}

func New(cfg Config) *Router {
	r := &Router{

		middlewares: []HandlerFn{},
	}

	r.errorHandler = defaultErrorHandler
	if cfg.ErrorHandler != nil {
		r.errorHandler = cfg.ErrorHandler
	}

	r.server = &http.Server{
		Addr:    cfg.Address(),
		Handler: defaultHandler(r),
	}

	return r
}

func (r *Router) Use(middlewares ...HandlerFn) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (router *Router) handle(w http.ResponseWriter, r *http.Request) error {
	for _, mw := range router.middlewares {
		if err := mw(w, r); err != nil {
			return err
		}
	}

	return nil
}

func (r *Router) Run(ctx context.Context) {
	if err := r.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return
		}

		panic(err)
	}
}

func (r *Router) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}
