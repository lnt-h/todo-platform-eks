import { test } from 'node:test'
import * as assert from 'node:assert'
import { build } from '../helper'

test('GET /healthz returns ok', async (t) => {
  const app = await build(t)

  const res = await app.inject({ url: '/healthz' })
  assert.strictEqual(res.statusCode, 200)
  assert.deepStrictEqual(res.json(), { status: 'ok' })
})
