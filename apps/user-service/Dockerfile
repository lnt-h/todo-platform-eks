# syntax=docker/dockerfile:1

# ---- Stage 1: install deps and compile TypeScript ----
FROM node:24-alpine AS build
WORKDIR /app

# Dependency layer: cached until package*.json change.
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

COPY tsconfig.json ./
COPY src ./src
# Compile, then drop devDependencies (typescript, c8, ...) from node_modules.
RUN npm run build && npm prune --omit=dev

# ---- Stage 2: minimal runtime (node only: no shell, no npm, non-root) ----
FROM gcr.io/distroless/nodejs24-debian13:nonroot
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/package.json ./
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/dist ./dist
USER nonroot:nonroot
EXPOSE 3000
# No curl in distroless: use node's built-in fetch.
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/nodejs/bin/node", "-e", "fetch('http://127.0.0.1:'+(process.env.PORT||3000)+'/healthz').then(r=>process.exit(r.ok?0:1),()=>process.exit(1))"]
# The distroless entrypoint is already "node".
CMD ["dist/server.js"]
