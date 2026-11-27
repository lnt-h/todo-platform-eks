import { FastifyPluginAsync } from 'fastify'
import { UserPatch, UserStore } from '../store'

// JSON Schemas: Fastify validates requests (returns 400) before the handler runs.
const idParams = {
  type: 'object',
  properties: { id: { type: 'integer', minimum: 1 } },
  required: ['id']
} as const

const name = { type: 'string', minLength: 1, maxLength: 100, pattern: '\\S' } as const
const email = { type: 'string', format: 'email', maxLength: 254 } as const

const userRoutes: FastifyPluginAsync<{ store: UserStore }> = async (app, { store }) => {
  app.get('/users', async () => store.list())

  app.get<{ Params: { id: number } }>('/users/:id', { schema: { params: idParams } },
    async (request) => store.get(request.params.id))

  app.post<{ Body: { name: string, email: string } }>('/users', {
    schema: {
      body: {
        type: 'object',
        properties: { name, email },
        required: ['name', 'email'],
        additionalProperties: false
      }
    }
  }, async (request, reply) => {
    const user = await store.create(request.body.name.trim(), request.body.email.toLowerCase())
    return reply.code(201).send(user)
  })

  app.patch<{ Params: { id: number }, Body: UserPatch }>('/users/:id', {
    schema: {
      params: idParams,
      body: { type: 'object', properties: { name, email }, additionalProperties: false, minProperties: 1 }
    }
  }, async (request) => {
    const { name, email } = request.body
    return store.update(request.params.id, { name: name?.trim(), email: email?.toLowerCase() })
  })

  app.delete<{ Params: { id: number } }>('/users/:id', { schema: { params: idParams } },
    async (request, reply) => {
      await store.delete(request.params.id)
      return reply.code(204).send()
    })
}

export default userRoutes
