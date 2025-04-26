package config

import (
	"bytes"
	"crypto/sha256"
	"io"
	"log/slog"
	"os"
	"time"
)

type Manager struct {
	lastHash []byte
	path     string
	config   *Config
	timeout  time.Duration
	changed  chan struct{}
}

func (m *Manager) Changed() <-chan struct{} {
	return m.changed
}

func newManager(config *Config, timeout time.Duration, path string) *Manager {
	m := &Manager{
		config:  config,
		timeout: timeout,
		path:    path,
		changed: make(chan struct{}, 1),
	}
	m.lastHash, _ = m.hash()

	go m.HandleChange()

	return m
}

func (m *Manager) Config() *Config {
	return m.config
}

func (m *Manager) HandleChange() {
	t := time.NewTicker(m.timeout)

	for range t.C {
		hash, err := m.hash()
		if err != nil {
			slog.Error("failed to get hash", slog.String("err", err.Error()))
			continue
		}

		if !bytes.Equal(m.lastHash, hash) {
			slog.Debug("config changed, reloading...")
			m.lastHash = hash
			m.config.reload(m.path)
			m.changed <- struct{}{}
		}

	}
}

func (m *Manager) hash() ([]byte, error) {
	file, err := os.Open(m.path)
	if err != nil {
		slog.Error("failed to open file", slog.String("err", err.Error()))
		return nil, err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, err
	}

	return hasher.Sum(nil), nil
}
