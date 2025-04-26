package storage

import (
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

func NewRedis(ctx context.Context, connString string) (*redis.Client, error) {
	config, err := redis.ParseURL(connString)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(config)

	if err := client.Ping(ctx).Err(); err != nil {
		slog.Error("failed to connect to redis", slog.String("connString", connString))
		return nil, err
	}

	slog.Info("connected to redis")

	return client, nil
}
