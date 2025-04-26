package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/tehrelt/eginx/internal/config"
	"github.com/tehrelt/eginx/internal/pool"
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
	pool := pool.New(cfg, slog.With(slog.Int("worker", 1)))
	go func() {
		pool.Run(ctx)
	}()

	<-ctx.Done()
	ctx = context.Background()

	if err := pool.Shutdown(ctx); err != nil {
		panic(err)
	}

}
