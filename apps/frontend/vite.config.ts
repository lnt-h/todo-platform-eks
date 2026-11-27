import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Dev-only reverse proxy. The browser only talks to /api/*; the proxy forwards
// to each service and strips the /api prefix. nginx (Docker) and Ingress (K8s)
// do the same job later, so frontend code never changes between environments.
const TODO_API = process.env.TODO_API_URL ?? 'http://localhost:8080'
const USER_API = process.env.USER_API_URL ?? 'http://localhost:3000'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api/todos': { target: TODO_API, rewrite: (path) => path.replace(/^\/api/, '') },
      '/api/users': { target: USER_API, rewrite: (path) => path.replace(/^\/api/, '') },
    },
  },
})
