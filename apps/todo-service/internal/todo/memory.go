package todo

import (
	"context"
	"sort"
	"sync"
	"time"
)

// Memory is an in-memory Store used in unit tests.
type Memory struct {
	mu     sync.Mutex
	nextID int64
	items  map[int64]Todo
}

func NewMemory() *Memory {
	return &Memory{nextID: 1, items: map[int64]Todo{}}
}

func (m *Memory) List(_ context.Context) ([]Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Todo, 0, len(m.items))
	for _, t := range m.items {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *Memory) Get(_ context.Context, id int64) (Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.items[id]
	if !ok {
		return Todo{}, ErrNotFound
	}
	return t, nil
}

func (m *Memory) Create(_ context.Context, title string) (Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t := Todo{ID: m.nextID, Title: title, CreatedAt: time.Now().UTC()}
	m.items[t.ID] = t
	m.nextID++
	return t, nil
}

func (m *Memory) Update(_ context.Context, id int64, p Patch) (Todo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.items[id]
	if !ok {
		return Todo{}, ErrNotFound
	}
	if p.Title != nil {
		t.Title = *p.Title
	}
	if p.Completed != nil {
		t.Completed = *p.Completed
	}
	m.items[id] = t
	return t, nil
}

func (m *Memory) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.items[id]; !ok {
		return ErrNotFound
	}
	delete(m.items, id)
	return nil
}

func (m *Memory) Ping(_ context.Context) error { return nil }
