# syntax=docker/dockerfile:1

# ---- Stage 1: build a static binary ----
FROM golang:1.27-alpine AS build
WORKDIR /src

# Copy only dependency files first: this layer stays cached until go.mod/go.sum change.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
# CGO_ENABLED=0: fully static binary, runs without libc.
# -trimpath -ldflags="-s -w": strip local paths and debug symbols (smaller binary).
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/todo-service ./cmd/server

# ---- Stage 2: minimal runtime (no shell, no package manager, non-root) ----
FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/todo-service /todo-service
USER nonroot:nonroot
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/todo-service", "healthcheck"]
ENTRYPOINT ["/todo-service"]
