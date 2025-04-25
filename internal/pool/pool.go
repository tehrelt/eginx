package pool

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/tehrelt/eginx/internal/config"
	"github.com/tehrelt/eginx/internal/pool/health"
	"github.com/tehrelt/eginx/internal/pool/iter"
)

type Server struct {
	URL    *url.URL
	rp     *reverseProxy
	health *health.HealthChecker
}

type ServerPool struct {
	cfg    config.Config
	logger *slog.Logger
	iter   iter.Iterator[*Server]
}

func (s *Server) Alive() bool {
	return s.health.Alive()
}

func New(cfg config.Config, logger *slog.Logger) *ServerPool {
	servers := make([]*Server, 0, len(cfg.Targets))

	for _, target := range cfg.Targets {
		u, _ := url.Parse(target)
		servers = append(servers, &Server{
			rp:     newRP(u),
			URL:    u,
			health: health.New(u),
		})
	}

	return &ServerPool{
		cfg:    cfg,
		iter:   newServerIterator(servers),
		logger: slog.With(slog.String("struct", "ServerPool")),
	}
}

func (s *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server, _ := s.iter.Next()
	s.logger.Info("forwarding request", slog.String("url", server.URL.String()))
	server.rp.ServeHTTP(w, r)
}

func (s *ServerPool) Run(ctx context.Context) {
	addr := fmt.Sprintf(":%d", s.cfg.Port)

	s.logger.Info("starting reverse proxy", slog.String("addr", addr))
	http.ListenAndServe(addr, http.HandlerFunc(s.ServeHTTP))
}
