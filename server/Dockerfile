# Tempiex UI server (gateway + embedded web UI)
FROM node:22-alpine AS web-builder

WORKDIR /web
COPY web/package.json web/pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

# ── Go build ──────────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS go-builder

WORKDIR /build

# Shared proto-gen stubs (server uses replace directive to ../tempiex)
COPY tempiex/ ./tempiex/
COPY server/go.mod server/go.sum ./server/
RUN cd server && go mod download

COPY server/ ./server/

# Embed the web dist — must match the //go:embed ui/* path in internal/server/server.go
COPY --from=web-builder /web/dist ./server/internal/server/ui/

RUN cd server && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/tempiex-ui-server ./cmd/server/

# ── runtime ──────────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=go-builder /out/tempiex-ui-server /usr/local/bin/tempiex-ui-server
COPY server/config/ /etc/tempiex-server/

ENV TEMPIEX_CONFIG=/etc/tempiex-server/docker.yaml

EXPOSE 8080

ENTRYPOINT ["tempiex-ui-server"]
