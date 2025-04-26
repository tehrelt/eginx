package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"
)

type Config struct {
	Port    int      `json:"port"`
	Targets []string `json:"targets"`
	Version string   `json:"version"`
	Limiter struct {
		Enabled    bool `json:"enabled"`
		DefaultRPS int  `json:"defaultRPS"`
	} `json:"limiter"`
}

func (c *Config) reload(path string) error {
	cfg, err := parse(path)
	if err != nil {
		return err
	}

	*c = *cfg
	return nil
}

func Parse(ctx context.Context, path string) (*Manager, error) {
	cfg, err := parse(path)
	if err != nil {
		return nil, err
	}

	return newManager(ctx, cfg, time.Second, path), nil
}

func parse(filePath string) (cfg *Config, err error) {

	f, err := os.Open(filePath)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, err
	}

	for _, target := range cfg.Targets {
		if _, err := url.Parse(target); err != nil {
			return cfg, fmt.Errorf("failed to parse url %q: %w", target, err)
		}
	}

	return cfg, nil
}
