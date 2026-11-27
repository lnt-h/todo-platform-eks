package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lnt-h/todo-platform-eks/apps/todo-service/internal/httpapi"
	"github.com/lnt-h/todo-platform-eks/apps/todo-service/internal/todo"
)

func main() {
	// "todo-service healthcheck" is used by Docker HEALTHCHECK: the distroless
	// image has no shell or curl, so the binary checks itself.
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(healthcheck())
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("todo-service stopped", "err", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	// Cancelled on Ctrl+C or SIGTERM (what Kubernetes sends before killing a pod).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL is required")
	}
	store, err := todo.NewPostgres(ctx, dsn)
	if err != nil {
		return err
	}
	defer store.Close()

	srv := &http.Server{
		Addr:              ":" + getenv("PORT", "8080"),
		Handler:           httpapi.NewRouter(store, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("todo-service listening", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	// Graceful shutdown: stop accepting new requests, finish in-flight ones.
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func healthcheck() int {
	client := http.Client{Timeout: 2 * time.Second}
	res, err := client.Get("http://127.0.0.1:" + getenv("PORT", "8080") + "/healthz")
	if err != nil {
		return 1
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
