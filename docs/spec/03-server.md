# 03 — Server: UI Server Specification

## Overview

**Component**: UI HTTP/gRPC Gateway Server  
**Folder**: `server/`  
**Language**: Go 1.24+  
**Type**: HTTP Server + gRPC Client Bridge  
**Implementation Priority**: 3 — Depends on tempiex (gRPC API)

## Purpose

`server` is a lightweight HTTP server that bridges the Tempiex React web UI (running in a browser) with the Tempiex gRPC API. It provides:

- **HTTP-to-gRPC translation** — REST endpoints backed by Tempiex gRPC calls
- **Authentication** — OIDC / OAuth 2.0 with session management
- **Static asset serving** — serves the built `web/` SPA
- **CORS & CSRF protection**
- **Markdown rendering** for workflow annotations
- **Codec server proxy** — forward payload decode requests to external codec servers

## Architecture

```
┌──────────────────────────────────────────────────────┐
│                    Web Browser                       │
│            Tempiex React UI (static assets)          │
└──────────────────────────┬───────────────────────────┘
                           │ HTTP/HTTPS
                           ▼
┌──────────────────────────────────────────────────────┐
│                  server  :8080                       │
│                                                      │
│  Middleware stack (in order):                        │
│    Logger → CORS → CSRF → Auth → SecurityHeaders     │
│    → Recover                                         │
│                                                      │
│  ┌────────────────────────────────────────────────┐  │
│  │            HTTP Route Handlers                 │  │
│  │  /api/v1/namespaces/...                        │  │
│  │  /api/v1/workflows/...                         │  │
│  │  /api/v1/cluster                               │  │
│  │  /api/v1/auth/*                                │  │
│  │  /api/v1/settings                              │  │
│  └────────────────────┬───────────────────────────┘  │
│                       │                               │
│  ┌────────────────────▼───────────────────────────┐  │
│  │     HTTP → gRPC Transformation Layer           │  │
│  │  • JSON ↔ protobuf marshal/unmarshal           │  │
│  │  • pagination token pass-through               │  │
│  │  • timestamp formatting                        │  │
│  └────────────────────┬───────────────────────────┘  │
│                       │ gRPC                          │
│  ┌────────────────────▼───────────────────────────┐  │
│  │         Tempiex gRPC Client                   │  │
│  │  WorkflowService + OperatorService             │  │
│  │  Connection pool, TLS, reconnection            │  │
│  └────────────────────┬───────────────────────────┘  │
└───────────────────────┼──────────────────────────────┘
                        │ gRPC :8133
                        ▼
              ┌────────────────────┐
              │  tempiex / proxy   │
              └────────────────────┘

Optional:
  OIDC Provider (Auth0, Okta, etc.)
  Codec Server (external payload decoder)
```

## Technical Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.24+ |
| HTTP Framework | Echo v4 | 4.13+ |
| gRPC client | google.golang.org/grpc | 1.79+ |
| Auth | go-oidc/v3 + oauth2 | 3.11+, 0.34+ |
| Session | gorilla/securecookie | 1.1+ |
| Observability | OpenTelemetry OTLP | 1.44+ |
| Logging | zerolog | 1.35+ |
| DI | uber-go/fx | 1.24+ |
| Config | gopkg.in/yaml.v3 | |

## Folder Structure

```
server/
├── cmd/
│   └── server/
│       └── main.go          # binary entry point
├── internal/
│   ├── api/                 # HTTP route handlers
│   │   ├── namespace.go
│   │   ├── workflow.go
│   │   ├── schedule.go
│   │   ├── cluster.go
│   │   ├── auth.go
│   │   └── settings.go
│   ├── auth/                # OIDC, OAuth 2.0, sessions
│   │   ├── oidc.go
│   │   └── session.go
│   ├── codec/               # Codec server proxy client
│   ├── config/              # YAML config loader
│   ├── cors/
│   ├── csrf/
│   ├── headers/             # Security headers middleware
│   ├── metrics/             # OTel metric definitions
│   ├── rpc/                 # gRPC client setup
│   └── server.go            # Echo server wiring
├── ui/                      # Embedded built web/ assets (go:embed)
├── config/
│   ├── development.yaml
│   └── docker.yaml
├── plugins/                 # Plugin system (codec server, custom auth)
│   ├── codec/
│   └── auth/
├── go.mod
└── go.sum
```

## API Routes

### Namespace Routes
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/namespaces` | List all namespaces |
| GET | `/api/v1/namespaces/:ns` | Get namespace detail |
| POST | `/api/v1/namespaces` | Register namespace |
| PATCH | `/api/v1/namespaces/:ns` | Update namespace |

### Workflow Routes
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/namespaces/:ns/workflows` | List workflows |
| GET | `/api/v1/namespaces/:ns/workflows/:wfId` | Get workflow |
| GET | `/api/v1/namespaces/:ns/workflows/:wfId/history` | Event history |
| POST | `/api/v1/namespaces/:ns/workflows` | Start workflow |
| POST | `/api/v1/namespaces/:ns/workflows/:wfId/terminate` | Terminate |
| POST | `/api/v1/namespaces/:ns/workflows/:wfId/signal` | Send signal |
| POST | `/api/v1/namespaces/:ns/workflows/:wfId/query` | Query |
| POST | `/api/v1/namespaces/:ns/workflows/:wfId/update` | Update |
| POST | `/api/v1/namespaces/:ns/workflows/:wfId/cancel` | Cancel |
| POST | `/api/v1/namespaces/:ns/workflows/:wfId/reset` | Reset |
| GET  | `/api/v1/namespaces/:ns/workflows/:wfId/query` | Query workflow state |
| GET | `/api/v1/namespaces/:ns/workflows/:wfId/stack-trace` | Stack trace |

### Schedule Routes
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/namespaces/:ns/schedules` | List schedules |
| GET | `/api/v1/namespaces/:ns/schedules/:id` | Get schedule |
| POST | `/api/v1/namespaces/:ns/schedules` | Create schedule |
| PATCH | `/api/v1/namespaces/:ns/schedules/:id` | Update schedule |
| DELETE | `/api/v1/namespaces/:ns/schedules/:id` | Delete schedule |

### Cluster Routes
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/cluster` | Cluster info |
| GET | `/api/v1/cluster/health` | Health check |
| GET | `/api/v1/cluster/nexus/endpoints` | Nexus endpoints |
| GET | `/api/v1/cluster/search-attributes` | List custom search attributes |

### Auth Routes (when enabled)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/auth/sso/login` | Initiate OAuth/OIDC |
| GET | `/api/v1/auth/sso/callback` | OAuth callback |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/auth/sso` | SSO info |
| GET | `/api/v1/auth/user` | Get current authenticated user |

### Settings
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/settings` | UI runtime settings |

### Static Asset Routes

| Route | Description |
|-------|-------------|
| `GET /` | Serve SPA `index.html` |
| `GET /static/*` | Serve JS, CSS, images |
| `GET /*` | SPA fallback → `index.html` |

### Health
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Liveness |
| GET | `/ready` | Readiness (checks gRPC connection) |

## Configuration

```yaml
# server/config/development.yaml
tempiexGrpcAddress: "localhost:8133"
host: "0.0.0.0"
port: 8080
publicPath: "/"

tls:
  enabled: false

auth:
  enabled: false
  # providers:
  #   - label: "Okta"
  #     type: oidc
  #     issuerUrl: "https://dev-123456.okta.com"
  #     clientId: "your-client-id"
  #     clientSecret: "${OIDC_CLIENT_SECRET}"
  #     callbackUrl: "http://localhost:8080/api/v1/auth/sso/callback"
  #     scopes: [openid, profile, email]

cors:
  cookieInsecure: true   # dev only

csrf:
  enabled: true
  secure: false          # dev only

codecEndpoint: ""        # optional external codec server URL

metrics:
  otel:
    endpoint: "localhost:4317"
    insecure: true
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TEMPIEX_GRPC_ADDRESS` | `localhost:8133` | Tempiex gRPC address |
| `TEMPIEX_UI_PORT` | `8080` | HTTP listen port |
| `TEMPIEX_TLS_CA_PATH` | — | CA certificate path |
| `TEMPIEX_TLS_CERT_PATH` | — | Client cert path (mTLS) |
| `TEMPIEX_TLS_KEY_PATH` | — | Client key path (mTLS) |
| `OIDC_CLIENT_SECRET` | — | OIDC client secret |
| `TEMPIEX_CORS_ORIGINS` | — | Allowed CORS origins (comma-separated) |

## Static Asset Serving

The `web/` build output is embedded into the `server` binary using `//go:embed ui/*`.

In development:
- Run `pnpm dev` in `web/` (Vite proxies API calls to `server`).
- `server` does not need to serve static assets during web development.

In production:
- `pnpm build` in `web/` outputs to `web/dist/`.
- CI copies `web/dist/` → `server/ui/` before building the `server` binary.
- The binary serves the SPA at `/*` and API at `/api/v1/*`.

## Middleware Stack

```
Request
  │
  ▼ Logger (zerolog request log)
  ▼ Recover (panic → 500)
  ▼ CORS (allow origins from config)
  ▼ CSRF (token validation, set cookie)
  ▼ Auth (OIDC session check, redirect to login if needed)
  ▼ SecurityHeaders (CSP, X-Frame-Options, X-Content-Type-Options)
  ▼ Route Handler
```

The CSRF middleware exposes `GET /api/v1/csrf-token`, which returns a fresh CSRF token and sets the `_csrf` cookie. The web UI calls this on load and includes the token in the `X-CSRF-Token` header for all state-changing requests.

## Authentication Flow (OIDC)

```
Browser → GET /api/v1/auth/sso/login
  → server redirects to OIDC provider /authorize
Browser → OIDC provider callback → GET /api/v1/auth/sso/callback?code=...
  → server exchanges code for tokens
  → server stores session in secure cookie
  → redirect to /
```

### Security Headers

```yaml
headers:
  contentSecurityPolicy: "default-src 'self'"
  strictTransportSecurity: "max-age=31536000; includeSubDomains"
  xFrameOptions: "DENY"
  xContentTypeOptions: "nosniff"
  referrerPolicy: "strict-origin-when-cross-origin"
```

## Build & Test

```bash
cd server/

# Build (assumes web/dist copied to server/ui/)
go build -o bin/tempiex-server ./cmd/server

# Run (dev mode, no UI assets needed)
./bin/tempiex-server --config config/development.yaml

# Run with env override
TEMPIEX_SERVER_PORT=9090 ./bin/tempiex-server

# Tests
go test ./...

# Lint
golangci-lint run ./...
```

## Performance Characteristics

| Metric | Value |
|--------|-------|
| Static asset latency | <10ms (served from memory) |
| API request overhead | 10–50ms over direct gRPC |
| Auth session validation | +5–10ms |
| Concurrent connections | 1,000+ (Echo framework) |
| Requests per second | 5,000+ (cached gRPC connections) |

### Resource Requirements

| Deployment | CPU | Memory |
|------------|-----|--------|
| Development | 0.25 cores | 128 MB |
| Production (low traffic) | 1 core | 512 MB |
| Production (high traffic) | 2 cores | 1 GB |

## Observability

- OTel OTLP gRPC for metrics (request count, latency, gRPC errors).
- Trace context propagated from HTTP headers → gRPC metadata.
- zerolog JSON to stdout.

Key metrics:
- `server.http.requests.total` — by route, method, status
- `server.http.request.duration` — histogram
- `server.grpc.errors.total` — by service, method
- `server.auth.sessions.active` — gauge

## Dependencies (key)

```
github.com/labstack/echo/v4 v4.13+
google.golang.org/grpc v1.79+
go.tempiex.com/api v1.63+
github.com/coreos/go-oidc/v3 v3.11+
golang.org/x/oauth2 v0.34+
github.com/gorilla/securecookie v1.1+
gopkg.in/yaml.v3
go.opentelemetry.io/otel v1.44+
go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc
go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc
github.com/rs/zerolog v1.35+
go.uber.org/fx v1.24+
github.com/gomarkdown/markdown
```
