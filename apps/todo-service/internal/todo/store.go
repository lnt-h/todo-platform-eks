// Package todo holds the domain model and storage for todos.
package todo

import (
	"context"
	"errors"
	"time"
)

var ErrNotFound = errors.New("todo not found")

type Todo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

// Patch holds optional fields for a partial update.
type Patch struct {
	Title     *string `json:"title"`
	Completed *bool   `json:"completed"`
}

// Store is implemented by Postgres (runtime) and Memory (tests).
type Store interface {
	List(ctx context.Context) ([]Todo, error)
	Get(ctx context.Context, id int64) (Todo, error)
	Create(ctx context.Context, title string) (Todo, error)
	Update(ctx context.Context, id int64, p Patch) (Todo, error)
	Delete(ctx context.Context, id int64) error
	Ping(ctx context.Context) error
}
