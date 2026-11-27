import { useEffect, useState, type FormEvent } from 'react'
import { api, type Todo, type User } from './api'

function App() {
  return (
    <main>
      <h1>Todo Platform</h1>
      <TodoSection />
      <UserSection />
    </main>
  )
}

function TodoSection() {
  const [todos, setTodos] = useState<Todo[]>([])
  const [title, setTitle] = useState('')
  const [error, setError] = useState('')

  const run = async (fn: () => Promise<unknown>) => {
    try {
      setError('')
      await fn()
      setTodos(await api.listTodos())
    } catch (e) {
      setError(message(e))
    }
  }

  useEffect(() => {
    api.listTodos().then(setTodos).catch((e) => setError(message(e)))
  }, [])

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    if (!title.trim()) return
    run(async () => {
      await api.createTodo(title)
      setTitle('')
    })
  }

  const done = todos.filter((t) => t.completed).length

  return (
    <section>
      <h2>
        Todos <small>{done}/{todos.length} done</small>
      </h2>
      <form onSubmit={onSubmit}>
        <input value={title} onChange={(e) => setTitle(e.target.value)} placeholder="What needs doing?" />
        <button type="submit">Add</button>
      </form>
      {error && <p className="error">todo-service: {error}</p>}
      <ul>
        {todos.map((todo) => (
          <li key={todo.id}>
            <label className={todo.completed ? 'done' : ''}>
              <input type="checkbox" checked={todo.completed} onChange={() => run(() => api.toggleTodo(todo))} />
              {todo.title}
            </label>
            <button className="link" onClick={() => run(() => api.deleteTodo(todo.id))} aria-label="Delete todo">
              ✕
            </button>
          </li>
        ))}
      </ul>
    </section>
  )
}

function UserSection() {
  const [users, setUsers] = useState<User[]>([])
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')

  const run = async (fn: () => Promise<unknown>) => {
    try {
      setError('')
      await fn()
      setUsers(await api.listUsers())
    } catch (e) {
      setError(message(e))
    }
  }

  useEffect(() => {
    api.listUsers().then(setUsers).catch((e) => setError(message(e)))
  }, [])

  const onSubmit = (e: FormEvent) => {
    e.preventDefault()
    run(async () => {
      await api.createUser(name, email)
      setName('')
      setEmail('')
    })
  }

  return (
    <section>
      <h2>Users</h2>
      <form onSubmit={onSubmit}>
        <input value={name} onChange={(e) => setName(e.target.value)} placeholder="Name" required />
        <input type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="Email" required />
        <button type="submit">Add</button>
      </form>
      {error && <p className="error">user-service: {error}</p>}
      <ul>
        {users.map((user) => (
          <li key={user.id}>
            <span>
              {user.name} <small>{user.email}</small>
            </span>
            <button className="link" onClick={() => run(() => api.deleteUser(user.id))} aria-label="Delete user">
              ✕
            </button>
          </li>
        ))}
      </ul>
    </section>
  )
}

const message = (e: unknown) => (e instanceof Error ? e.message : String(e))

export default App
