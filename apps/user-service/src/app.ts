import Fastify, { FastifyError, FastifyInstance, FastifyServerOptions } from 'fastify'
import { collectDefaultMetrics, Counter, Histogram, Registry } from 'prom-client'
import { ConflictError, NotFoundError, UserStore } from './store'
import healthRoutes from './routes/health'
import userRoutes from './routes/users'

export interface AppOptions {
  store: UserStore
  logger?: FastifyServerOptions['logger']
}

export function buildApp ({ store, logger = true }: AppOptions): FastifyInstance {
  // Fastify logs with pino: one JSON line per log entry.
  const app = Fastify({ logger })

  // ---- Prometheus metrics (same names as todo-service) ----
  const registry = new Registry()
  collectDefaultMetrics({ register: registry })
  const requestsTotal = new Counter({
    name: 'http_requests_total',
    help: 'Total HTTP requests by method, route and status code.',
    labelNames: ['method', 'route', 'status'],
    registers: [registry]
  })
  const requestDuration = new Histogram({
    name: 'http_request_duration_seconds',
    help: 'HTTP request latency by method and route.',
    labelNames: ['method', 'route'],
    registers: [registry]
  })

  app.addHook('onResponse', async (request, reply) => {
    // Route pattern ("/users/:id"), not the raw URL, to keep cardinality low.
    const route = request.routeOptions.url ?? 'unmatched'
    requestsTotal.inc({ method: request.method, route, status: reply.statusCode })
    requestDuration.observe({ method: request.method, route }, reply.elapsedTime / 1000)
  })

  app.get('/metrics', { logLevel: 'silent' }, async (_request, reply) => {
    reply.type(registry.contentType)
    return registry.metrics()
  })

  // ---- Map domain errors to HTTP status codes ----
  app.setErrorHandler((err: FastifyError, request, reply) => {
    if (err instanceof NotFoundError) return reply.code(404).send({ error: err.message })
    if (err instanceof ConflictError) return reply.code(409).send({ error: err.message })
    if (err.validation) return reply.code(400).send({ error: err.message })
    request.log.error({ err }, 'request failed')
    return reply.code(500).send({ error: 'internal error' })
  })

  app.register(healthRoutes, { store })
  app.register(userRoutes, { store })

  return app
}
