package pool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/tehrelt/eginx/internal/config"
	"github.com/tehrelt/eginx/internal/pool/health"
	"github.com/tehrelt/eginx/internal/pool/iter"
)

type ServerPool struct {
	cfg    *config.Manager
	logger *slog.Logger
	iter   iter.Iterator[*server]
	server *http.Server
}

type serverError struct {
	error
	code int
}

func newServerError(err error, code int) *serverError {
	return &serverError{
		error: err,
		code:  code,
	}
}

func (e *serverError) Error() string {
	return fmt.Sprintf("%s [%d]", e.error.Error(), e.code)
}

func (e *serverError) write(w http.ResponseWriter) {
	resp := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    e.code,
		Message: e.error.Error(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode response json", slog.Any("error", err), slog.Any("resp", resp))
	}
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

func (p *ServerPool) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	var errResp *serverError
	if !errors.As(err, &errResp) {
		errResp = newServerError(err, http.StatusInternalServerError)
	}

	errResp.write(w)

	p.logger.Error("proxy error", slog.String("error", errResp.Error()))
}

func (s *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := s.Serve(w, r); err != nil {
		s.logger.Error("proxy error", slog.Any("error", err))
		s.errorHandler(w, r, err)
	}
}

func (s *ServerPool) Serve(w http.ResponseWriter, r *http.Request) error {
	server, ok := s.iter.Next()
	if !ok {
		return newServerError(errNoServersAvailable, http.StatusServiceUnavailable)
	}

	s.logger.Info("forwarding request", slog.String("url", server.URL.String()))
	server.rp.ServeHTTP(w, r)
	return nil
}

func (s *ServerPool) Run(ctx context.Context) {
	addr := fmt.Sprintf(":%d", s.cfg.Config().Port)

	s.server = &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(s.ServeHTTP),
	}
	s.logger.Info("starting reverse proxy", slog.String("addr", addr))

	s.updateServers(ctx)

	go s.watchConfig(ctx)

	defer s.logger.Info("server stopped")

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("reverse proxy failed", slog.String("err", err.Error()))
	}
}

func (s *ServerPool) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
