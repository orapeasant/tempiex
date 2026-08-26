# Tempiex core engine
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Cache dependencies
COPY tempiex/go.mod tempiex/go.sum ./tempiex/
RUN cd tempiex && go mod download

# Copy source
COPY tempiex/ ./tempiex/

# Build
RUN cd tempiex && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/tempiex-server ./cmd/server/

# ── runtime ──────────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/tempiex-server /usr/local/bin/tempiex-server
COPY tempiex/config/ /etc/tempiex/

ENV TEMPIEX_CONFIG=/etc/tempiex/docker.yaml

EXPOSE 8133

ENTRYPOINT ["tempiex-server"]
