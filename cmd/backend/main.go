package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tehrelt/eginx/internal/server"
)

var (
	port int
)

func init() {
	flag.IntVar(&port, "port", 5000, "port to bind")
}

type writer struct {
	path string
}

func newWriter(path string) *writer {
	return &writer{path: fmt.Sprintf("%d-%s", time.Now().Unix(), path)}
}

func (w *writer) Write(b []byte) (int, error) {
	f, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	if _, err := os.Stdout.Write(b); err != nil {
		return 0, err
	}

	return f.Write(b)
}

func main() {
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(newWriter("logs.json"), &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	port := port

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	server := server.New(port)
	go server.Run()

	s := <-sig
	slog.Info("signal catched: interrupting", slog.String("signal", s.String()))

	ctx := context.Background()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("server shutdown failed: %s\n", err)
	}
}
