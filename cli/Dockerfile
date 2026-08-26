# Tempiex CLI
FROM node:22-alpine AS builder

WORKDIR /app
COPY cli/package.json cli/pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile
COPY cli/ ./
RUN pnpm build

# ── runtime ──────────────────────────────────────────────────────────────────
FROM node:22-alpine

WORKDIR /app
COPY --from=builder /app/dist/tempiex.js /usr/local/bin/tempiex
RUN chmod +x /usr/local/bin/tempiex

ENTRYPOINT ["node", "/usr/local/bin/tempiex"]
CMD ["--help"]
