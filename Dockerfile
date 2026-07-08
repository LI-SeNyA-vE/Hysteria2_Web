# ── Stage 1: сборка фронтенда ─────────────────────────────────────────────────
FROM node:22-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ── Stage 2: сборка Go-бинаря ─────────────────────────────────────────────────
FROM golang:latest AS go-builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Вставляем собранный фронтенд, чтобы go:embed нашёл frontend/dist
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o panel ./cmd/panel

# ── Stage 3: финальный образ ───────────────────────────────────────────────────
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=go-builder /app/panel ./panel

RUN useradd -r -u 1001 -s /bin/false panel \
    && mkdir -p /app/data \
    && chown -R panel:panel /app
USER panel

EXPOSE 8080

ENTRYPOINT ["/app/panel"]
