package health

import (
	"log/slog"
	"time"
)

type HealthCheckerOption func(*HealthChecker)

func WithTimeout(timeout time.Duration) HealthCheckerOption {
	return func(hc *HealthChecker) {
		hc.timeout = timeout
	}
}

func WithPeriod(period time.Duration) HealthCheckerOption {
	return func(hc *HealthChecker) {
		hc.period = period
	}
}

func WithLogger(l *slog.Logger) HealthCheckerOption {
	return func(hc *HealthChecker) {
		hc.logger = l
	}
}
