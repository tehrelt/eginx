package app

import (
	"context"

	"github.com/tehrelt/eginx/internal/config"
	"github.com/tehrelt/eginx/internal/pool"
	"github.com/tehrelt/eginx/internal/router"
)

type App struct {
	cfg    *config.Manager
	router *router.Router
	pool   *pool.ServerPool
}

func (a *App) Shutdown(ctx context.Context) any {
	return a.router.Shutdown(ctx)
}

type AppOptFn func(a *App)

func New(cfg *config.Manager, pool *pool.ServerPool, opts ...AppOptFn) *App {

	router := router.New(router.Config{
		Port: cfg.Config().Port,
	})

	app := &App{
		cfg:    cfg,
		router: router,
		pool:   pool,
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

func (a *App) setup(ctx context.Context) {
	a.router.Use(a.pool.Serve(ctx))
}

func (a *App) Run(ctx context.Context) {
	a.setup(ctx)

	a.router.Run(ctx)
}
