import { test } from 'node:test'
import * as assert from 'node:assert'
import { buildApp } from '../src/app'
import { MemoryStore } from '../src/store'

function newApp (t: { after: (fn: () => void) => void }) {
  const app = buildApp({ store: new MemoryStore(), logger: false })
  t.after(() => void app.close())
  return app
}

test('health and readiness', async (t) => {
  const app = newApp(t)
  for (const url of ['/healthz', '/readyz']) {
    const res = await app.inject({ url })
    assert.strictEqual(res.statusCode, 200, url)
  }
})

test('user CRUD', async (t) => {
  const app = newApp(t)

  let res = await app.inject({ method: 'POST', url: '/users', payload: { name: ' Hien ', email: 'Hien@Example.com' } })
  assert.strictEqual(res.statusCode, 201)
  assert.deepStrictEqual(
    { id: res.json().id, name: res.json().name, email: res.json().email },
    { id: 1, name: 'Hien', email: 'hien@example.com' }
  )

  res = await app.inject({ method: 'PATCH', url: '/users/1', payload: { name: 'Logan' } })
  assert.strictEqual(res.statusCode, 200)
  assert.strictEqual(res.json().name, 'Logan')

  res = await app.inject({ url: '/users' })
  assert.strictEqual(res.json().length, 1)

  res = await app.inject({ method: 'DELETE', url: '/users/1' })
  assert.strictEqual(res.statusCode, 204)

  res = await app.inject({ url: '/users/1' })
  assert.strictEqual(res.statusCode, 404)
})

test('validation and conflicts', async (t) => {
  const app = newApp(t)
  await app.inject({ method: 'POST', url: '/users', payload: { name: 'A', email: 'a@example.com' } })

  const cases = [
    { method: 'POST', url: '/users', payload: { name: 'B', email: 'a@example.com' }, want: 409 },
    { method: 'POST', url: '/users', payload: { name: 'B', email: 'not-an-email' }, want: 400 },
    { method: 'POST', url: '/users', payload: { name: '   ', email: 'b@example.com' }, want: 400 },
    { method: 'GET', url: '/users/abc', want: 400 },
    { method: 'PATCH', url: '/users/1', payload: {}, want: 400 },
    { method: 'PATCH', url: '/users/99', payload: { name: 'X' }, want: 404 }
  ] as const

  for (const c of cases) {
    const res = await app.inject({ method: c.method, url: c.url, payload: 'payload' in c ? c.payload : undefined })
    assert.strictEqual(res.statusCode, c.want, `${c.method} ${c.url}`)
  }
})

test('metrics are labelled by route pattern', async (t) => {
  const app = newApp(t)
  await app.inject({ url: '/users/123' })

  const res = await app.inject({ url: '/metrics' })
  assert.match(res.body, /route="\/users\/:id"/)
})
