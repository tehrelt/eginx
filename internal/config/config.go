package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
)

type Config struct {
	Port       int      `json:"port"`
	Targets    []string `json:"targets"`
	Version    string   `json:"version"`
	WorkerPool struct {
		Count int `json:"count"`
	} `json:"workerPool"`
}

func Parse(filePath string) (cfg Config, err error) {

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
