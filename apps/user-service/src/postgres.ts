import { Pool, DatabaseError } from 'pg'
import { ConflictError, NotFoundError, User, UserPatch, UserStore } from './store'

const SCHEMA = `
CREATE TABLE IF NOT EXISTS users (
  id         BIGSERIAL PRIMARY KEY,
  name       TEXT        NOT NULL,
  email      TEXT        NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`

const COLUMNS = 'id::int AS id, name, email, created_at'
const UNIQUE_VIOLATION = '23505'

export class PostgresStore implements UserStore {
  private constructor (private readonly pool: Pool) {}

  // Connects and makes sure the table exists.
  static async connect (connectionString: string): Promise<PostgresStore> {
    const pool = new Pool({ connectionString })
    await pool.query(SCHEMA)
    return new PostgresStore(pool)
  }

  async list (): Promise<User[]> {
    const { rows } = await this.pool.query<User>(`SELECT ${COLUMNS} FROM users ORDER BY id`)
    return rows
  }

  async get (id: number): Promise<User> {
    return this.one(`SELECT ${COLUMNS} FROM users WHERE id = $1`, [id])
  }

  async create (name: string, email: string): Promise<User> {
    return this.one(`INSERT INTO users (name, email) VALUES ($1, $2) RETURNING ${COLUMNS}`, [name, email])
  }

  async update (id: number, patch: UserPatch): Promise<User> {
    return this.one(
      `UPDATE users SET name = COALESCE($2, name), email = COALESCE($3, email)
       WHERE id = $1 RETURNING ${COLUMNS}`,
      [id, patch.name ?? null, patch.email ?? null]
    )
  }

  async delete (id: number): Promise<void> {
    const { rowCount } = await this.pool.query('DELETE FROM users WHERE id = $1', [id])
    if (rowCount === 0) throw new NotFoundError('user not found')
  }

  async ping (): Promise<void> {
    await this.pool.query('SELECT 1')
  }

  async close (): Promise<void> {
    await this.pool.end()
  }

  private async one (sql: string, params: unknown[]): Promise<User> {
    try {
      const { rows } = await this.pool.query<User>(sql, params)
      if (rows.length === 0) throw new NotFoundError('user not found')
      return rows[0]
    } catch (err) {
      if (err instanceof DatabaseError && err.code === UNIQUE_VIOLATION) {
        throw new ConflictError('email already exists')
      }
      throw err
    }
  }
}
