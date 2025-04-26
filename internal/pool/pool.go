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
	cfg    *config.Manager
	logger *slog.Logger
	iter   iter.Iterator[*Server]
}

func (s *Server) Alive() bool {
	return s.health.Alive()
}

func New(cfg *config.Manager, logger *slog.Logger) *ServerPool {
	servers := make([]*Server, 0, len(cfg.Config().Targets))

	for _, target := range cfg.Config().Targets {
		u, _ := url.Parse(target)
		servers = append(servers, &Server{
			rp:     newRP(u),
			URL:    u,
			health: health.New(u),
		})
	}

	p := &ServerPool{
		cfg:    cfg,
		iter:   newServerIterator(servers),
		logger: slog.With(slog.String("struct", "ServerPool")),
	}

	go func() {
		for range cfg.Changed() {
			servers := make([]*Server, 0, len(cfg.Config().Targets))
			for _, target := range cfg.Config().Targets {
				u, _ := url.Parse(target)
				servers = append(servers, &Server{
					rp:     newRP(u),
					URL:    u,
					health: health.New(u),
				})
			}

			p.logger.Info("updating server pool", slog.Any("servers", servers))
			p.iter.Set(servers)
		}
	}()

	return p
}

func (s *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	server, _ := s.iter.Next()
	s.logger.Info("forwarding request", slog.String("url", server.URL.String()))
	server.rp.ServeHTTP(w, r)
}

func (s *ServerPool) Run(ctx context.Context) {
	addr := fmt.Sprintf(":%d", s.cfg.Config().Port)
	s.logger.Info("starting reverse proxy", slog.String("addr", addr))
	if err := http.ListenAndServe(addr, http.HandlerFunc(s.ServeHTTP)); err != nil {
		s.logger.Error("reverse proxy failed", slog.String("err", err.Error()))
	}
}
