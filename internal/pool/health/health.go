package health

import (
	"log/slog"
	"net"
	"net/url"
	"sync"
	"time"
)

var (
	defaultTimeout = 5 * time.Second
	defaultPeriod  = 10 * time.Second
)

type HealthChecker struct {
	target  *url.URL
	period  time.Duration
	timeout time.Duration
	cancel  chan struct{}
	alive   bool
	m       sync.Mutex
	logger  *slog.Logger
}

func New(target *url.URL, opts ...HealthCheckerOption) *HealthChecker {
	hc := &HealthChecker{
		target:  target,
		period:  defaultPeriod,
		timeout: defaultTimeout,
		cancel:  make(chan struct{}),
		logger:  slog.With(slog.String("target", target.Host)),
	}

	for _, opt := range opts {
		opt(hc)
	}

	hc.check()

	go hc.run()

	return hc
}

func (hc *HealthChecker) check() {
	hc.m.Lock()
	defer hc.m.Unlock()

	defer func() {
		hc.logger.Debug("checked health", slog.Bool("alive", hc.alive))
	}()
	conn, err := net.DialTimeout("tcp", hc.target.Host, hc.timeout)
	if err != nil {
		hc.alive = false
		return
	}
	defer conn.Close()

	hc.alive = true
}

func (hc *HealthChecker) run() {
	ticker := time.NewTicker(hc.period)
	hc.logger.Info("starting health checker", slog.Float64("period", hc.period.Seconds()))

	for {
		select {
		case <-ticker.C:
			hc.check()
		case <-hc.cancel:
			ticker.Stop()
			return
		}

	}
}

func (hc *HealthChecker) Stop() {
	hc.m.Lock()
	defer hc.m.Unlock()

	hc.cancel <- struct{}{}
	close(hc.cancel)
}

func (hc *HealthChecker) Alive() bool {
	hc.m.Lock()
	defer hc.m.Unlock()

	return hc.alive
}
