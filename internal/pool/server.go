package pool

import (
	"net/url"

	"github.com/tehrelt/eginx/internal/pool/health"
)

type server struct {
	URL    *url.URL
	rp     *reverseProxy
	health *health.HealthChecker
}

func (s *server) Alive() bool {
	return s.health.Alive()
}
