export interface User {
  id: number
  name: string
  email: string
  created_at: string
}

export interface UserPatch {
  name?: string
  email?: string
}

export class NotFoundError extends Error {}
export class ConflictError extends Error {}

// Implemented by PostgresStore (runtime) and MemoryStore (tests).
export interface UserStore {
  list(): Promise<User[]>
  get(id: number): Promise<User>
  create(name: string, email: string): Promise<User>
  update(id: number, patch: UserPatch): Promise<User>
  delete(id: number): Promise<void>
  ping(): Promise<void>
  close(): Promise<void>
}

export class MemoryStore implements UserStore {
  private nextId = 1
  private users = new Map<number, User>()

  async list (): Promise<User[]> {
    return [...this.users.values()]
  }

  async get (id: number): Promise<User> {
    const user = this.users.get(id)
    if (!user) throw new NotFoundError('user not found')
    return user
  }

  async create (name: string, email: string): Promise<User> {
    this.assertEmailFree(email)
    const user = { id: this.nextId++, name, email, created_at: new Date().toISOString() }
    this.users.set(user.id, user)
    return user
  }

  async update (id: number, patch: UserPatch): Promise<User> {
    const user = await this.get(id)
    if (patch.email !== undefined && patch.email !== user.email) this.assertEmailFree(patch.email)
    const updated = { ...user, ...patch }
    this.users.set(id, updated)
    return updated
  }

  async delete (id: number): Promise<void> {
    if (!this.users.delete(id)) throw new NotFoundError('user not found')
  }

  async ping (): Promise<void> {}

  async close (): Promise<void> {}

  private assertEmailFree (email: string): void {
    for (const u of this.users.values()) {
      if (u.email === email) throw new ConflictError('email already exists')
    }
  }
}
