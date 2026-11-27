# todo-platform-eks

[![todo-service](https://github.com/lnt-h/todo-platform-eks/actions/workflows/todo-service.yml/badge.svg)](https://github.com/lnt-h/todo-platform-eks/actions/workflows/todo-service.yml)
[![user-service](https://github.com/lnt-h/todo-platform-eks/actions/workflows/user-service.yml/badge.svg)](https://github.com/lnt-h/todo-platform-eks/actions/workflows/user-service.yml)
[![frontend](https://github.com/lnt-h/todo-platform-eks/actions/workflows/frontend.yml/badge.svg)](https://github.com/lnt-h/todo-platform-eks/actions/workflows/frontend.yml)
[![security](https://github.com/lnt-h/todo-platform-eks/actions/workflows/security.yml/badge.svg)](https://github.com/lnt-h/todo-platform-eks/actions/workflows/security.yml)

A small todo app that I use to practice DevOps: containers, CI/CD, Kubernetes, Terraform on AWS, GitOps and monitoring.
The app is kept simple on purpose. The focus is on how it is built, shipped and run.

## Architecture

```
browser ──► frontend (React + nginx)
              ├── /api/todos/* ──► todo-service (Go) ──────────► Postgres db "todos"
              └── /api/users/* ──► user-service (Node/Fastify) ─► Postgres db "users"
```

- Each service owns its own database.
- Every service exposes `/healthz`, `/readyz` and Prometheus `/metrics`, logs JSON and shuts down gracefully on SIGTERM.

## Run locally

Requires Docker.

```bash
docker compose up -d --build --wait
# open http://localhost:8000
docker compose down
```

## CI

| Workflow | What it does |
|---|---|
| `todo-service`, `user-service`, `frontend` | lint, test, build the image, scan it with Trivy, push to GHCR from `main` |
| `docker-image` | shared build/scan/push job used by the three workflows above |
| `security` | gitleaks on the full git history, actionlint on the workflows |

Workflows only run when their app changes. Actions are pinned to commit SHAs.

## Tech stack

| Area | Tools |
|---|---|
| App | Go, Node.js (TypeScript, Fastify), React + Vite, Postgres |
| Containers | Docker multi-stage builds, distroless images, Docker Compose |
| CI | GitHub Actions, Trivy, gitleaks, pre-commit |
| Planned | kind, Helm, Terraform, AWS EKS, ArgoCD, Prometheus, Grafana |

## Progress

See [ROADMAP.md](ROADMAP.md). Notes on what I learned: [docs/LEARNING.md](docs/LEARNING.md).
