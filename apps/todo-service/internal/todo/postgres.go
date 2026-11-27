package todo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schema = `
CREATE TABLE IF NOT EXISTS todos (
	id         BIGSERIAL PRIMARY KEY,
	title      TEXT        NOT NULL,
	completed  BOOLEAN     NOT NULL DEFAULT FALSE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`

type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres connects to the database and makes sure the table exists.
func NewPostgres(ctx context.Context, dsn string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if _, err := pool.Exec(ctx, schema); err != nil {
		pool.Close()
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() { p.pool.Close() }

func (p *Postgres) Ping(ctx context.Context) error { return p.pool.Ping(ctx) }

func (p *Postgres) List(ctx context.Context) ([]Todo, error) {
	rows, err := p.pool.Query(ctx, `SELECT id, title, completed, created_at FROM todos ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[Todo])
}

func (p *Postgres) Get(ctx context.Context, id int64) (Todo, error) {
	return p.one(ctx, `SELECT id, title, completed, created_at FROM todos WHERE id = $1`, id)
}

func (p *Postgres) Create(ctx context.Context, title string) (Todo, error) {
	return p.one(ctx, `INSERT INTO todos (title) VALUES ($1) RETURNING id, title, completed, created_at`, title)
}

func (p *Postgres) Update(ctx context.Context, id int64, patch Patch) (Todo, error) {
	return p.one(ctx, `
		UPDATE todos
		SET title = COALESCE($2, title), completed = COALESCE($3, completed)
		WHERE id = $1
		RETURNING id, title, completed, created_at`, id, patch.Title, patch.Completed)
}

func (p *Postgres) Delete(ctx context.Context, id int64) error {
	tag, err := p.pool.Exec(ctx, `DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (p *Postgres) one(ctx context.Context, sql string, args ...any) (Todo, error) {
	rows, err := p.pool.Query(ctx, sql, args...)
	if err != nil {
		return Todo{}, err
	}
	t, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[Todo])
	if errors.Is(err, pgx.ErrNoRows) {
		return Todo{}, ErrNotFound
	}
	return t, err
}
