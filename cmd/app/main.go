package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/tehrelt/eginx/internal/app"
	"github.com/tehrelt/eginx/internal/config"
	"github.com/tehrelt/eginx/internal/limiter"
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

		limiterOpts := make([]limiter.LimiterPoolOpt, 0, 2)
		if cfg.Config().Limiter.DefaultRPS != 0 {
			slog.Info("default limiter enabled", slog.Int("requests_per_second", cfg.Config().Limiter.DefaultRPS))
			limiterOpts = append(
				limiterOpts,
				limiter.WithDefaultLimiter(
					tokenbucket.New(ctx, cfg.Config().Limiter.DefaultRPS),
				),
			)
		}
		limiterpool := limiter.NewLimiterPool(ctx, storage, func(ctx context.Context, ratePerSec int) limiter.Limiter {
			return tokenbucket.New(ctx, ratePerSec)
		}, limiterOpts...)

		limiterOpt := app.WithLimiter(limiterpool, func(r *http.Request) string {
			key := r.Header.Get("X-API-Key")
			slog.Debug("api key from request", slog.String("key", key))
			return key
		})
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
