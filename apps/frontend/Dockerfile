# syntax=docker/dockerfile:1

# ---- Stage 1: build static files ----
FROM node:24-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY . .
RUN npm run build

# ---- Stage 2: serve with nginx as a non-root user (listens on 8080, not 80) ----
FROM nginxinc/nginx-unprivileged:1.31-alpine
# Pull security patches newer than the upstream image, and drop curl (unused:
# the healthcheck uses busybox wget). Then switch back to the non-root user.
USER root
RUN apk upgrade --no-cache && apk del --no-cache curl
USER 101
ENV TODO_API_URL=http://todo-service:8080 \
    USER_API_URL=http://user-service:3000
COPY nginx/default.conf.template /etc/nginx/templates/default.conf.template
COPY nginx/proxy_params.inc /etc/nginx/conf.d/proxy_params.inc
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
