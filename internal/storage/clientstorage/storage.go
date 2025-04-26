package clientstorage

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	redis "github.com/redis/go-redis/v9"
	"github.com/tehrelt/eginx/internal/storage"
)

type Storage struct {
	logger *slog.Logger
	db     *redis.Client
}

func New(client *redis.Client) *Storage {
	return &Storage{
		db:     client,
		logger: slog.With(slog.String("struct", "clientstorage.Storage")),
	}
}

func (s *Storage) Create(ctx context.Context, key string, rate int) error {
	fn := "clientstorage.Create"
	log := s.logger.With(slog.String("fn", fn))

	log.Info("creating client", slog.String("key", key), slog.Int("rate", rate))
	if err := s.db.Set(ctx, key, rate, 0).Err(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) Get(ctx context.Context, key string) (int, error) {
	fn := "clientstorage.Get"
	log := s.logger.With(slog.String("fn", fn))

	log.Info("getting client", slog.String("key", key))
	cmd := s.db.Get(ctx, key)
	if errors.Is(cmd.Err(), redis.Nil) {
		return 0, storage.ErrClientNotFound
	}

	rate, err := strconv.Atoi(cmd.Val())
	if err != nil {
		return 0, err
	}

	return rate, nil
}

func (s *Storage) Delete(ctx context.Context, key string) error {
	fn := "clientstorage.Delete"
	log := s.logger.With(slog.String("fn", fn))

	log.Info("deleting client", slog.String("key", key))
	cmd := s.db.Del(ctx, key)
	err := cmd.Err()
	if err != nil {
		log.Error("failed to delete client", slog.String("key", key), slog.String("err", err.Error()))
		return err
	}

	return nil
}
