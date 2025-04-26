package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tehrelt/eginx/internal/config"
	"github.com/tehrelt/eginx/internal/pool"
	"github.com/tehrelt/eginx/internal/router"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "./config.json", "path to config file")
}

func main() {
	flag.Parse()
	ctx := context.Background()

	l := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug},
		),
	)
	slog.SetDefault(l)

	cfg, err := config.Parse(ctx, configPath)
	if err != nil {
		panic(err)
	}

	ctx, _ = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	router := router.New(router.Config{
		Port: cfg.Config().Port,
	})

	pool := pool.New(cfg, slog.With(slog.Int("worker", 1)))

	router.Use(func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Add("Through-Middleware", "true")
		return nil
	})

	router.Use(pool.Serve(ctx))

	go func() {
		router.Run(ctx)
	}()

	<-ctx.Done()
	ctx = context.Background()

	if err := router.Shutdown(ctx); err != nil {
		panic(err)
	}

}
