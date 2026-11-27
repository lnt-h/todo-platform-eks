import { type FastifyPluginAsync } from 'fastify'

const health: FastifyPluginAsync = async (fastify): Promise<void> => {
  fastify.get('/healthz', async () => ({ status: 'ok' }))
}

export default health
