package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/tehrelt/eginx/internal/config"
	"github.com/tehrelt/eginx/internal/storage"
	"github.com/tehrelt/eginx/internal/storage/clientstorage"
)

type ClientStorage interface {
	Create(ctx context.Context, key string, rate int) error
	Delete(ctx context.Context, key string) error
}

type CreateClientRequest struct {
	Key  string `json:"key"`
	Rate int    `json:"rate"`
}

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "", "path to config file")
}

func main() {
	flag.Parse()
	if configPath == "" {
		log.Fatal("config path is required")
	}

	ctx := context.Background()
	cfg, err := config.Parse(ctx, configPath)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := storage.NewRedis(ctx, cfg.Config().Redis.ConnectionString())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	storage := clientstorage.New(pool)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req CreateClientRequest

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if req.Key == "" || req.Rate == 0 {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}

			slog.Info("creating client", slog.String("key", req.Key), slog.Int("rate", req.Rate))
			if err := storage.Create(r.Context(), req.Key, req.Rate); err != nil {
				slog.Error("failed to create client", slog.String("key", req.Key), slog.Int("rate", req.Rate))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

		case http.MethodDelete:
			key := strings.TrimPrefix(r.URL.Path, "/")
			if err := storage.Delete(r.Context(), key); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}

	})

	log.Fatal(http.ListenAndServe(":7000", nil))
}
