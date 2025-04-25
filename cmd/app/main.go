package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

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

	l := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slog.LevelDebug},
		),
	)
	slog.SetDefault(l)

	ctx := context.Background()

	cfg, err := config.Parse(configPath)
	if err != nil {
		panic(err)
	}

	pool := pool.New(cfg, slog.With(slog.Int("worker", 1)))

	pool.Run(ctx)
}
