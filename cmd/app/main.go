package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/tehrelt/eginx/internal/app"
	"github.com/tehrelt/eginx/internal/config"
	"github.com/tehrelt/eginx/internal/limiter/tokenbucket"
	"github.com/tehrelt/eginx/internal/pool"
	"github.com/tehrelt/eginx/internal/storage"
	"github.com/tehrelt/eginx/internal/storage/clientstorage"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "./config.json", "path to config file")
}

type mockstorage struct {
	db map[string]int
	m  sync.Mutex
}

func (s *mockstorage) Get(_ context.Context, key string) (int, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if rate, ok := s.db[key]; ok {
		return rate, nil
	}

	return 0, errors.New("not found")
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
		log.Fatal(err)
	}

	ctx, _ = signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	pool := pool.New(cfg, slog.With(slog.Int("worker", 1)))
	db, err := storage.NewRedis(ctx, cfg.Config().Redis.ConnectionString())
	if err != nil {
		log.Fatal(err)
	}

	storage := clientstorage.New(db)

	opts := make([]app.AppOptFn, 0, 2)
	if cfg.Config().Limiter.Enabled {

		slog.Info("limiter enabled")

		limiterOpts := make([]tokenbucket.TokenBucketOpt, 0, 2)

		defaultRpm := cfg.Config().Limiter.DefaultRPM
		if defaultRpm != 0 {
			slog.Info("default limiter enabled", slog.Int("rpm", defaultRpm))

			limiterOpts = append(
				limiterOpts,
				tokenbucket.WithDefaultBucket(ctx, defaultRpm),
			)
		}

		bucketpool := tokenbucket.NewTokenBucket(ctx, storage, limiterOpts...)
		limiterOpt := app.WithLimiter(bucketpool, "X-API-Key")
		opts = append(opts, limiterOpt)
	}

	app := app.New(cfg, pool, opts...)

	go func() {
		app.Run(ctx)
	}()
	<-ctx.Done()

	ctx = context.Background()
	if err := app.Shutdown(ctx); err != nil {
		panic(err)
	}
}
