import closeWithGrace from 'close-with-grace'
import { buildApp } from './app'
import { PostgresStore } from './postgres'

async function main (): Promise<void> {
  const databaseUrl = process.env.DATABASE_URL
  if (!databaseUrl) throw new Error('DATABASE_URL is required')

  const store = await PostgresStore.connect(databaseUrl)
  const app = buildApp({ store })

  // Graceful shutdown on SIGTERM/SIGINT (what Kubernetes sends before killing a pod):
  // stop accepting requests, finish in-flight ones, then close the DB pool.
  closeWithGrace({ delay: 10_000 }, async ({ signal, err }) => {
    if (err) app.log.error({ err }, 'fatal error')
    else app.log.info({ signal }, 'shutting down')
    await app.close()
    await store.close()
  })

  // 0.0.0.0, not localhost: inside a container localhost is unreachable from outside.
  await app.listen({ host: '0.0.0.0', port: Number(process.env.PORT ?? 3000) })
}

main().catch((err) => {
  console.error(JSON.stringify({ level: 'fatal', msg: 'user-service failed to start', err: String(err) }))
  process.exit(1)
})
