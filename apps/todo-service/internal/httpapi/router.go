// Package httpapi exposes the todo REST API, health checks and metrics.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/lnt-h/todo-platform-eks/apps/todo-service/internal/todo"
)

type api struct {
	store  todo.Store
	logger *slog.Logger
}

func NewRouter(store todo.Store, logger *slog.Logger) http.Handler {
	a := &api{store: store, logger: logger}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", a.healthz)
	mux.HandleFunc("GET /readyz", a.readyz)
	mux.Handle("GET /metrics", promhttp.Handler())

	mux.HandleFunc("GET /todos", a.list)
	mux.HandleFunc("POST /todos", a.create)
	mux.HandleFunc("GET /todos/{id}", a.get)
	mux.HandleFunc("PATCH /todos/{id}", a.update)
	mux.HandleFunc("DELETE /todos/{id}", a.delete)

	return instrument(mux, logger)
}

// healthz: the process is alive (Kubernetes liveness probe).
func (a *api) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readyz: dependencies are reachable (Kubernetes readiness probe).
func (a *api) readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := a.store.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "db unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (a *api) list(w http.ResponseWriter, r *http.Request) {
	todos, err := a.store.List(r.Context())
	if err != nil {
		a.serverError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, todos)
}

func (a *api) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	t, err := a.store.Get(r.Context(), id)
	a.respond(w, http.StatusOK, t, err)
}

func (a *api) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	t, err := a.store.Create(r.Context(), title)
	a.respond(w, http.StatusCreated, t, err)
}

func (a *api) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var p todo.Patch
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if p.Title != nil && strings.TrimSpace(*p.Title) == "" {
		writeError(w, http.StatusBadRequest, "title cannot be empty")
		return
	}
	t, err := a.store.Update(r.Context(), id, p)
	a.respond(w, http.StatusOK, t, err)
}

func (a *api) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := a.store.Delete(r.Context(), id); err != nil {
		a.respond(w, 0, nil, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) respond(w http.ResponseWriter, status int, v any, err error) {
	switch {
	case errors.Is(err, todo.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case err != nil:
		a.serverError(w, err)
	default:
		writeJSON(w, status, v)
	}
}

func (a *api) serverError(w http.ResponseWriter, err error) {
	a.logger.Error("request failed", "err", err)
	writeError(w, http.StatusInternalServerError, "internal error")
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
