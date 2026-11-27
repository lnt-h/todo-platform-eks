import { FastifyPluginAsync } from 'fastify'
import { UserStore } from '../store'

// logLevel 'silent': probes hit these every few seconds, logging them is noise.
const healthRoutes: FastifyPluginAsync<{ store: UserStore }> = async (app, { store }) => {
  // Liveness: the process is alive.
  app.get('/healthz', { logLevel: 'silent' }, async () => ({ status: 'ok' }))

  // Readiness: the database is reachable.
  app.get('/readyz', { logLevel: 'silent' }, async (_request, reply) => {
    try {
      await store.ping()
      return { status: 'ready' }
    } catch {
      return reply.code(503).send({ status: 'db unavailable' })
    }
  })
}

export default healthRoutes
