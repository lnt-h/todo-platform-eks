// All calls use relative /api/* paths; a reverse proxy routes them to the right service.

export interface Todo {
  id: number
  title: string
  completed: boolean
  created_at: string
}

export interface User {
  id: number
  name: string
  email: string
  created_at: string
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/api${path}`, {
    ...init,
    headers: init?.body ? { 'Content-Type': 'application/json' } : undefined,
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error ?? `${res.status} ${res.statusText}`)
  }
  return (res.status === 204 ? undefined : await res.json()) as T
}

export const api = {
  listTodos: () => request<Todo[]>('/todos'),
  createTodo: (title: string) =>
    request<Todo>('/todos', { method: 'POST', body: JSON.stringify({ title }) }),
  toggleTodo: (todo: Todo) =>
    request<Todo>(`/todos/${todo.id}`, {
      method: 'PATCH',
      body: JSON.stringify({ completed: !todo.completed }),
    }),
  deleteTodo: (id: number) => request<void>(`/todos/${id}`, { method: 'DELETE' }),

  listUsers: () => request<User[]>('/users'),
  createUser: (name: string, email: string) =>
    request<User>('/users', { method: 'POST', body: JSON.stringify({ name, email }) }),
  deleteUser: (id: number) => request<void>(`/users/${id}`, { method: 'DELETE' }),
}
