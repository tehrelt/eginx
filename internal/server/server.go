package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
)

type server struct {
	server *http.Server
	port   int
	logger *slog.Logger
}

func New(port int) *server {
	s := &server{
		logger: slog.Default(),
		port:   port,
	}

	return s
}

func (s *server) Run() error {
	defer func() {
		s.logger.Info("server stopped")
	}()

	addr := fmt.Sprintf(":%d", s.port)

	s.logger.Info("binding", slog.String("addr", addr))
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Error("can't bind port", slog.Int("port", s.port), slog.String("error", err.Error()))
		for {
			s.port++
			addr = fmt.Sprintf(":%d", s.port)

			s.logger.Info("binding", slog.String("addr", addr))
			listener, err = net.Listen("tcp", addr)
			if err == nil {
				break
			}

			s.logger.Warn("cannot bind port", slog.Int("port", s.port))
		}
	}

	s.logger = s.logger.With(slog.String("addr", addr))

	s.server = &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(s.handle),
	}

	s.logger.Info("start server")
	if err := s.server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.logger.Error("can't start server", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s *server) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	s.logger.Info("shutdown server")
	return s.server.Shutdown(ctx)
}

func (s *server) handle(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("incoming request")
	data := struct {
		Server int `json:"server"`
	}{
		Server: s.port,
	}

	ok(data, w)
}

func ok(data any, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Println(err)
	}
}
