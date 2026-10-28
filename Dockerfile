# syntax=docker/dockerfile:1

# ---- Stage 1: build the React/Vite frontend ----
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
# install deps first so this layer caches unless the lockfile changes
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
# vite's build.outDir is ../internal/adapters/web/dist, so this writes the
# bundle to /app/internal/adapters/web/dist (where the Go embed expects it)
RUN npm run build

# ---- Stage 2: build the Go binary (embeds the frontend) ----
FROM golang:1.26-alpine AS builder
WORKDIR /src
# module deps first for layer caching
COPY go.mod go.sum ./
RUN go mod download
# the Go source
COPY . .
# bring in the freshly built frontend so //go:embed dist has files to embed
COPY --from=frontend /app/internal/adapters/web/dist ./internal/adapters/web/dist
# modernc sqlite is pure Go => CGO off => a fully static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/app ./cmd

# ---- Stage 3: minimal runtime image ----
FROM alpine:latest
# CA roots so outbound HTTPS to en.wikipedia.org actually works (scratch would fail TLS)
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /bin/app /app/app
# documents the port; Fly routes to it via internal_port in fly.toml
EXPOSE 8080
ENTRYPOINT ["/app/app"]
