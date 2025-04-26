package pool

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/tehrelt/eginx/internal/config"
	"github.com/tehrelt/eginx/internal/pool/health"
	"github.com/tehrelt/eginx/internal/pool/iter"
	"github.com/tehrelt/eginx/internal/router"
)

type ServerPool struct {
	cfg    *config.Manager
	logger *slog.Logger
	iter   iter.Iterator[*server]
}

func New(cfg *config.Manager, logger *slog.Logger) *ServerPool {
	return &ServerPool{
		cfg:    cfg,
		logger: slog.With(slog.String("struct", "ServerPool")),
	}
}

func (p *ServerPool) updateServers(ctx context.Context) {
	servers := make([]*server, 0, len(p.cfg.Config().Targets))
	for _, target := range p.cfg.Config().Targets {
		u, _ := url.Parse(target)
		servers = append(servers, &server{
			rp:     newRP(u),
			URL:    u,
			health: health.New(ctx, u),
		})
	}

	p.logger.Info("updating server pool", slog.Any("servers_len", len(servers)))

	if p.iter == nil {
		p.iter = newServerIterator(servers)
	} else {
		p.iter.Set(servers)
	}
}

func (p *ServerPool) watchConfig(ctx context.Context) {
	changed := p.cfg.Watch(ctx)

	for {
		select {
		case <-ctx.Done():
			p.logger.Debug("closing config watcher")
			return

		case <-changed:
			p.updateServers(ctx)
		}
	}
}

func (s *ServerPool) serve(w http.ResponseWriter, r *http.Request) error {
	server, ok := s.iter.Next()
	if !ok {
		return router.NewError(errNoServersAvailable, http.StatusServiceUnavailable)
	}

	s.logger.Info("forwarding request", slog.String("url", server.URL.String()))
	server.rp.ServeHTTP(w, r)
	return nil
}

func (s *ServerPool) Serve(ctx context.Context) router.HandlerFn {
	s.logger.Info("starting reverse proxy")
	s.updateServers(ctx)
	go s.watchConfig(ctx)

	return s.serve
}
