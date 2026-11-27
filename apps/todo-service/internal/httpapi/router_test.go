package httpapi

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lnt-h/todo-platform-eks/apps/todo-service/internal/todo"
)

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	srv := httptest.NewServer(NewRouter(todo.NewMemory(), logger))
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, srv *httptest.Server, method, path, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { res.Body.Close() })
	return res
}

func TestHealthAndReady(t *testing.T) {
	srv := newServer(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		if res := do(t, srv, http.MethodGet, path, ""); res.StatusCode != http.StatusOK {
			t.Errorf("%s: want 200, got %d", path, res.StatusCode)
		}
	}
}

func TestTodoCRUD(t *testing.T) {
	srv := newServer(t)

	res := do(t, srv, http.MethodPost, "/todos", `{"title":"learn kubernetes"}`)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("create: want 201, got %d", res.StatusCode)
	}
	var created todo.Todo
	if err := json.NewDecoder(res.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID != 1 || created.Title != "learn kubernetes" || created.Completed {
		t.Fatalf("unexpected todo: %+v", created)
	}

	res = do(t, srv, http.MethodPatch, "/todos/1", `{"completed":true}`)
	var updated todo.Todo
	_ = json.NewDecoder(res.Body).Decode(&updated)
	if res.StatusCode != http.StatusOK || !updated.Completed {
		t.Fatalf("update: got %d %+v", res.StatusCode, updated)
	}

	res = do(t, srv, http.MethodGet, "/todos", "")
	var list []todo.Todo
	_ = json.NewDecoder(res.Body).Decode(&list)
	if len(list) != 1 {
		t.Fatalf("list: want 1 todo, got %d", len(list))
	}

	if res = do(t, srv, http.MethodDelete, "/todos/1", ""); res.StatusCode != http.StatusNoContent {
		t.Fatalf("delete: want 204, got %d", res.StatusCode)
	}
	if res = do(t, srv, http.MethodGet, "/todos/1", ""); res.StatusCode != http.StatusNotFound {
		t.Fatalf("get after delete: want 404, got %d", res.StatusCode)
	}
}

func TestValidation(t *testing.T) {
	srv := newServer(t)
	cases := []struct {
		method, path, body string
		want               int
	}{
		{http.MethodPost, "/todos", `{"title":"  "}`, http.StatusBadRequest},
		{http.MethodPost, "/todos", `not json`, http.StatusBadRequest},
		{http.MethodGet, "/todos/abc", "", http.StatusBadRequest},
		{http.MethodPatch, "/todos/99", `{"completed":true}`, http.StatusNotFound},
	}
	for _, c := range cases {
		if res := do(t, srv, c.method, c.path, c.body); res.StatusCode != c.want {
			t.Errorf("%s %s: want %d, got %d", c.method, c.path, c.want, res.StatusCode)
		}
	}
}

func TestMetricsUseRoutePattern(t *testing.T) {
	srv := newServer(t)
	do(t, srv, http.MethodGet, "/todos/123", "")

	res := do(t, srv, http.MethodGet, "/metrics", "")
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), `route="GET /todos/{id}"`) {
		t.Fatal("metrics should label requests by route pattern")
	}
}
