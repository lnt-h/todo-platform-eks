package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/lnt-h/todo-platform-eks/apps/todo-service/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	addr := ":" + getenv("PORT", "8080")
	logger.Info("starting todo-service", "addr", addr)

	if err := http.ListenAndServe(addr, httpapi.NewRouter()); err != nil {
		logger.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
