package config

import (
	"context"
	"log/slog"
	"os"
	"time"
)

type Manager struct {
	path    string
	config  *Config
	timeout time.Duration
	stat    os.FileInfo
}

func newManager(ctx context.Context, config *Config, timeout time.Duration, path string) *Manager {
	m := &Manager{
		config:  config,
		timeout: timeout,
		path:    path,
	}

	m.stat, _ = os.Stat(path)

	return m
}

func (m *Manager) Config() *Config {
	return m.config
}

func (m *Manager) Watch(ctx context.Context) <-chan struct{} {
	t := time.NewTicker(m.timeout)
	ch := make(chan struct{}, 1)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				slog.Info("stopping file watcher")
				return

			case <-t.C:
				stat, err := os.Stat(m.path)
				if err != nil {
					slog.Warn("error checking config file", slog.String("error", err.Error()))
					continue
				}

				if !stat.ModTime().Equal(m.stat.ModTime()) || stat.Size() != m.stat.Size() {
					m.stat = stat
					slog.Debug("config file changed, reloading...")
					m.config.reload(m.path)
					select {
					case <-ctx.Done():
						slog.Info("stopping file watcher")
						return
					case ch <- struct{}{}:
						slog.Debug("config changed")
					}
				}
			}
		}
	}()

	return ch
}
