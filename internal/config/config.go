package config

import (
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
}

func (c *Config) reload(path string) error {
	cfg, err := parse(path)
	if err != nil {
		return err
	}

	*c = *cfg
	return nil
}

func Parse(path string) (*Manager, error) {
	cfg, err := parse(path)
	if err != nil {
		return nil, err
	}

	return newManager(cfg, time.Second, path), nil
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
