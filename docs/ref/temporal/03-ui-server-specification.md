# 03 - UI Server Specification

## Overview

**Component**: UI Server  
**Repository**: `ui-server/`  
**Language**: Go 1.26.3  
**Type**: HTTP/gRPC Gateway Server  
**Implementation Priority**: 3 (Depends on Temporal Server)

## Purpose

UI Server is a lightweight HTTP server that acts as a bridge between the Temporal Web UI (frontend) and the Temporal Server gRPC API. It provides:

- **HTTP-to-gRPC translation** for API calls
- **Authentication integration** (OIDC, OAuth 2.0)
- **CORS handling** for cross-origin requests
- **CSRF protection** for secure session management
- **Static asset serving** for the UI
- **Markdown rendering** for workflow documentation
- **Request/response transformation** and formatting

The UI Server enables web browsers to interact with Temporal Server without requiring gRPC-Web support.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Web Browser                            │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │          Temporal Web UI (JavaScript)                │  │
│  │          (Served as static assets)                   │  │
│  └──────────────────┬───────────────────────────────────┘  │
│                     │ HTTP/HTTPS                            │
└─────────────────────┼─────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                    UI Server (Go)                           │
│                                                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │              HTTP Server (Port 8080)               │    │
│  │  • Serve static UI assets                          │    │
│  │  • CORS middleware                                 │    │
│  │  • CSRF protection                                 │    │
│  │  • Authentication (OIDC/OAuth)                     │    │
│  │  • Session management                              │    │
│  └────────────────────┬───────────────────────────────┘    │
│                       │                                     │
│  ┌────────────────────▼───────────────────────────────┐    │
│  │          API Route Handlers                        │    │
│  │  • /api/v1/namespaces/...                         │    │
│  │  • /api/v1/workflows/...                          │    │
│  │  • /api/v1/cluster                                │    │
│  │  • /api/v1/auth (login, logout, callback)        │    │
│  └────────────────────┬───────────────────────────────┘    │
│                       │                                     │
│  ┌────────────────────▼───────────────────────────────┐    │
│  │       HTTP-to-gRPC Transformation                  │    │
│  │  • Convert HTTP requests to gRPC                   │    │
│  │  • Marshal/unmarshal protobuf                      │    │
│  │  • Handle pagination                               │    │
│  │  • Format timestamps                               │    │
│  └────────────────────┬───────────────────────────────┘    │
│                       │                                     │
│  ┌────────────────────▼───────────────────────────────┐    │
│  │         gRPC Client (Temporal API)                 │    │
│  │  • WorkflowService client                          │    │
│  │  • OperatorService client                          │    │
│  │  • Connection pooling                              │    │
│  │  • TLS support                                     │    │
│  └────────────────────┬───────────────────────────────┘    │
│                       │ gRPC                                │
└───────────────────────┼─────────────────────────────────────┘
                        │
                        ▼
            ┌───────────────────────┐
            │   Temporal Server     │
            │   (gRPC API :7233)    │
            └───────────────────────┘

Optional Components:
┌──────────────────────┐          ┌────────────────────┐
│  OIDC Provider       │          │  Codec Server      │
│  (Auth0, Okta, etc.) │          │  (Payload decode)  │
└──────────────────────┘          └────────────────────┘
```

### Component Breakdown

#### 1. HTTP Server
- **Framework**: Echo v4.13.4
- **Port**: Configurable (default 8080)
- **Responsibilities**:
  - Serve static UI assets from `/ui/` directory
  - Handle API requests at `/api/v1/...`
  - CORS middleware for cross-origin requests
  - CSRF token generation and validation
  - Request logging and metrics
  - Health checks

#### 2. Authentication Module
- **Methods**:
  - **OIDC (OpenID Connect)**: Integration with Auth0, Okta, Google, etc.
  - **OAuth 2.0**: Standard OAuth flow
  - **No auth**: Development mode
- **Libraries**:
  - `github.com/coreos/go-oidc/v3` for OIDC
  - `golang.org/x/oauth2` for OAuth 2.0
- **Session Management**:
  - Secure cookies with `gorilla/securecookie`
  - CSRF tokens per session

#### 3. API Route Handlers
**Namespace Routes**:
- `GET /api/v1/namespaces` - List namespaces
- `GET /api/v1/namespaces/{namespace}` - Get namespace details
- `POST /api/v1/namespaces` - Register namespace

**Workflow Routes**:
- `GET /api/v1/namespaces/{namespace}/workflows` - List workflows
- `GET /api/v1/namespaces/{namespace}/workflows/{workflowId}` - Get workflow details
- `GET /api/v1/namespaces/{namespace}/workflows/{workflowId}/history` - Get event history
- `POST /api/v1/namespaces/{namespace}/workflows/{workflowId}/terminate` - Terminate workflow
- `POST /api/v1/namespaces/{namespace}/workflows/{workflowId}/signal` - Send signal

**Cluster Routes**:
- `GET /api/v1/cluster` - Cluster info
- `GET /api/v1/cluster/health` - Health check

**Auth Routes** (when enabled):
- `GET /api/v1/auth/login` - Initiate OAuth login
- `GET /api/v1/auth/callback` - OAuth callback handler
- `POST /api/v1/auth/logout` - Logout

#### 4. gRPC Client
- **Services**:
  - `WorkflowService` - Workflow operations
  - `OperatorService` - Namespace and cluster operations
- **Features**:
  - Connection pooling
  - Automatic reconnection
  - TLS support
  - Request timeout handling
  - Metadata propagation

#### 5. Middleware Stack
1. **Logger** - Request logging
2. **CORS** - Cross-origin resource sharing
3. **CSRF** - Cross-site request forgery protection
4. **Auth** - Authentication check (optional)
5. **Headers** - Security headers (CSP, X-Frame-Options, etc.)
6. **Recover** - Panic recovery

## Technical Stack

### Core Technologies

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.26.3 |
| HTTP Framework | Echo v4 | 4.13.4 |
| RPC Client | gRPC | 1.79.3 |
| API Definition | Protobuf | 1.36.10 |
| Config Format | YAML | gopkg.in/yaml.v3 |

### Key Dependencies

```go
// HTTP Framework
github.com/labstack/echo/v4 v4.13.4

// gRPC & Temporal API
google.golang.org/grpc v1.79.3
go.temporal.io/api v1.63.4

// Authentication
github.com/coreos/go-oidc/v3 v3.11.0
golang.org/x/oauth2 v0.34.0

// Session & CSRF
github.com/gorilla/securecookie v1.1.1

// Configuration
gopkg.in/yaml.v3 v3.0.1
gopkg.in/validator.v2 v2.0.0-20210331031555-b37d688a7fb0

// Utilities
github.com/Masterminds/sprig/v3 v3.3.0
github.com/gomarkdown/markdown v0.0.0-20260411013819-759bbc3e3207
github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0
github.com/urfave/cli/v2 v2.3.0
```

## Directory Structure

```
ui-server/
├── cmd/
│   └── server/             # Main server binary
│       └── main.go
├── config/                 # Configuration files
├── docker/                 # Docker build files
├── plugins/                # Plugin system (codec, auth)
├── proto/                  # Proto definitions
├── server/
│   ├── api/               # API route handlers
│   │   ├── namespace.go   # Namespace routes
│   │   ├── workflow.go    # Workflow routes
│   │   └── cluster.go     # Cluster routes
│   ├── auth/              # Authentication logic
│   │   ├── oidc.go        # OIDC provider
│   │   └── session.go     # Session management
│   ├── config/            # Configuration loading
│   │   └── config.go
│   ├── cors/              # CORS middleware
│   │   └── cors.go
│   ├── csrf/              # CSRF protection
│   │   └── csrf.go
│   ├── headers/           # Security headers
│   │   └── headers.go
│   ├── route/             # Route registration
│   │   └── route.go
│   ├── rpc/               # gRPC client setup
│   │   └── client.go
│   ├── server_options/    # Server configuration options
│   │   └── options.go
│   ├── version/           # Version info
│   │   └── version.go
│   └── server.go          # Main server setup
└── ui/                    # Static UI assets (from ui build)
```

## Key Features

### 1. HTTP-to-gRPC Translation

**Request Flow**:
```
HTTP Request → Echo Handler → gRPC Call → Temporal Server
HTTP Response ← JSON Marshal ← gRPC Response ← Result
```

**Example**:
```go
// GET /api/v1/namespaces/default/workflows?pageSize=10
// Translates to:
req := &workflowservice.ListWorkflowExecutionsRequest{
    Namespace: "default",
    PageSize: 10,
}
resp, err := client.ListWorkflowExecutions(ctx, req)
```

### 2. Authentication & Authorization

#### OIDC Configuration

```yaml
auth:
  enabled: true
  providers:
    - label: "Okta"
      type: oidc
      issuerUrl: "https://dev-123456.okta.com"
      clientId: "your-client-id"
      clientSecret: "${OIDC_CLIENT_SECRET}"
      scopes:
        - openid
        - profile
        - email
      callbackUrl: "http://localhost:8080/auth/sso/callback"
```

#### OAuth 2.0 Configuration

```yaml
auth:
  enabled: true
  providers:
    - label: "Google"
      type: oauth
      issuerUrl: "https://accounts.google.com"
      clientId: "your-client-id"
      clientSecret: "${OAUTH_CLIENT_SECRET}"
      scopes:
        - openid
        - email
```

### 3. CORS Configuration

**Flexible CORS support** for different deployment scenarios:

```yaml
cors:
  allowOrigins:
    - "https://app.example.com"
    - "https://staging.example.com"
  allowMethods:
    - GET
    - POST
    - PUT
    - DELETE
  allowHeaders:
    - Content-Type
    - Authorization
  exposeHeaders:
    - X-Request-Id
  maxAge: 3600
  cookieInsecure: false  # Allow cookies over HTTP (dev only)
  unsafeAllowAllOrigins: false  # ⚠️ UNSAFE - dev only
```

**Security Note**: `unsafeAllowAllOrigins` should **never** be used in production.

### 4. CSRF Protection

**Token-based CSRF** protection for state-changing operations:

```yaml
csrf:
  enabled: true
  tokenLength: 32
  cookieName: "_csrf"
  headerName: "X-CSRF-Token"
```

**Flow**:
1. UI requests CSRF token from `/api/v1/csrf-token`
2. Server generates token, stores in secure cookie
3. UI includes token in `X-CSRF-Token` header for POST/PUT/DELETE
4. Server validates token matches cookie

### 5. Static Asset Serving

**UI Assets** built from `ui/` repository and embedded:

```go
// Serve UI assets from /ui directory
e.Static("/", "ui/")

// Fallback to index.html for SPA routing
e.File("/*", "ui/index.html")
```

### 6. Markdown Rendering

**Workflow documentation** can include markdown:

```go
import "github.com/gomarkdown/markdown"

// Render workflow memo as HTML
html := markdown.ToHTML([]byte(memo), nil, nil)
```

### 7. Health Checks

```bash
# Liveness check
curl http://localhost:8080/health

# Response:
{"status": "ok"}
```

## Configuration

### Minimal Configuration

```yaml
# UI Server listens on this address
publicUrl: "http://localhost:8080"

# Temporal Server gRPC address
temporalGrpcAddress: "localhost:7233"

# Optional TLS for Temporal connection
tls:
  caPath: ""
  certPath: ""
  keyPath: ""
  serverName: ""
  enableHostVerification: false

# CORS (optional)
cors:
  allowOrigins:
    - "http://localhost:3000"
```

### Production Configuration

```yaml
publicUrl: "https://temporal-ui.company.com"

temporalGrpcAddress: "temporal-frontend:7233"

# TLS for Temporal Server connection
tls:
  caPath: "/certs/ca.pem"
  certPath: "/certs/client.pem"
  keyPath: "/certs/client-key.pem"
  serverName: "temporal-frontend"
  enableHostVerification: true

# Authentication
auth:
  enabled: true
  providers:
    - label: "Company SSO"
      type: oidc
      issuerUrl: "https://sso.company.com"
      clientId: "temporal-ui"
      clientSecret: "${OIDC_CLIENT_SECRET}"
      scopes:
        - openid
        - profile
        - email
      callbackUrl: "https://temporal-ui.company.com/auth/sso/callback"

# CORS
cors:
  allowOrigins:
    - "https://temporal-ui.company.com"
  allowMethods:
    - GET
    - POST
    - PUT
    - DELETE
  allowCredentials: true
  cookieInsecure: false

# CSRF
csrf:
  enabled: true

# Security Headers
headers:
  contentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"
  xFrameOptions: "DENY"
  xContentTypeOptions: "nosniff"
```

## Build & Deployment

### Build Commands

```bash
# Build binary
go build -o ui-server ./cmd/server

# Run locally
./ui-server --config config.yaml

# Run with environment variables
export TEMPORAL_GRPC_ADDRESS=localhost:7233
./ui-server
```

### Docker Image

```bash
# Pull from Docker Hub
docker pull temporalio/ui-server:latest

# Run container
docker run -p 8080:8080 \
  -e TEMPORAL_GRPC_ADDRESS=host.docker.internal:7233 \
  temporalio/ui-server:latest
```

### Docker Compose

```yaml
version: '3.8'
services:
  ui-server:
    image: temporalio/ui-server:latest
    ports:
      - "8080:8080"
    environment:
      - TEMPORAL_GRPC_ADDRESS=temporal:7233
    depends_on:
      - temporal
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TEMPORAL_GRPC_ADDRESS` | Temporal Server address | localhost:7233 |
| `TEMPORAL_UI_PORT` | UI Server listen port | 8080 |
| `TEMPORAL_TLS_CA_PATH` | CA certificate path | "" |
| `TEMPORAL_TLS_CERT_PATH` | Client certificate path | "" |
| `TEMPORAL_TLS_KEY_PATH` | Client key path | "" |
| `OIDC_CLIENT_SECRET` | OIDC client secret | "" |
| `TEMPORAL_CORS_ORIGINS` | Allowed CORS origins (comma-separated) | "" |

## API Endpoints

### Static Assets

| Route | Description |
|-------|-------------|
| `GET /` | Serve UI index.html |
| `GET /static/*` | Serve static assets (JS, CSS, images) |

### API Routes

#### Cluster & Health

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/cluster` | Get cluster info |
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/cluster/search-attributes` | Get search attributes |

#### Namespaces

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/namespaces` | List namespaces |
| GET | `/api/v1/namespaces/:namespace` | Get namespace details |
| POST | `/api/v1/namespaces` | Register namespace |
| PUT | `/api/v1/namespaces/:namespace` | Update namespace |

#### Workflows

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/namespaces/:namespace/workflows` | List workflows |
| GET | `/api/v1/namespaces/:namespace/workflows/:workflowId` | Get workflow |
| GET | `/api/v1/namespaces/:namespace/workflows/:workflowId/history` | Get history |
| POST | `/api/v1/namespaces/:namespace/workflows` | Start workflow |
| POST | `/api/v1/namespaces/:namespace/workflows/:workflowId/terminate` | Terminate |
| POST | `/api/v1/namespaces/:namespace/workflows/:workflowId/cancel` | Cancel |
| POST | `/api/v1/namespaces/:namespace/workflows/:workflowId/signal` | Signal |
| POST | `/api/v1/namespaces/:namespace/workflows/:workflowId/query` | Query |
| POST | `/api/v1/namespaces/:namespace/workflows/:workflowId/reset` | Reset |

#### Schedules

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/namespaces/:namespace/schedules` | List schedules |
| GET | `/api/v1/namespaces/:namespace/schedules/:scheduleId` | Get schedule |
| POST | `/api/v1/namespaces/:namespace/schedules` | Create schedule |
| PUT | `/api/v1/namespaces/:namespace/schedules/:scheduleId` | Update schedule |
| DELETE | `/api/v1/namespaces/:namespace/schedules/:scheduleId` | Delete schedule |

#### Authentication (when enabled)

| Method | Route | Description |
|--------|-------|-------------|
| GET | `/api/v1/auth/sso/login` | Initiate SSO login |
| GET | `/api/v1/auth/sso/callback` | OAuth callback |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/auth/user` | Get current user |

## Dependencies & Integration Points

### Upstream Dependencies
- **Temporal Server** (required) - gRPC API backend
- **OIDC Provider** (optional) - Authentication
- **Codec Server** (optional) - Payload decoding

### Downstream Dependents
- **Web Browser** - Users access UI via HTTP
- **Temporal Web UI** (static assets) - SvelteKit application

## Security

### Transport Security
- **HTTPS**: Recommended for production
- **TLS to Temporal Server**: mTLS support
- **Secure cookies**: HttpOnly, Secure, SameSite flags

### Authentication Security
- **OIDC/OAuth 2.0**: Industry-standard authentication
- **Session tokens**: Secure random tokens
- **CSRF protection**: Token-based validation
- **Password handling**: Never stores passwords (delegated to OIDC)

### Security Headers

```yaml
headers:
  contentSecurityPolicy: "default-src 'self'"
  strictTransportSecurity: "max-age=31536000; includeSubDomains"
  xFrameOptions: "DENY"
  xContentTypeOptions: "nosniff"
  referrerPolicy: "strict-origin-when-cross-origin"
```

## Performance Characteristics

### Latency
- **Static assets**: <10ms (served from memory)
- **API requests**: 10-50ms overhead over direct gRPC call
- **With auth**: +5-10ms for session validation

### Throughput
- **Concurrent connections**: 1000+ (Echo framework)
- **Requests per second**: 5000+ (cached connections)

### Resource Requirements

| Deployment | CPU | Memory |
|------------|-----|--------|
| Development | 0.25 cores | 128 MB |
| Production (low traffic) | 1 core | 512 MB |
| Production (high traffic) | 2 cores | 1 GB |

## Testing

### Development Mode

```bash
# Run with hot reload (requires Air)
go run ./cmd/server --config config.yaml
```

### Integration with UI

```bash
# UI server expects UI assets at ./ui/
# Build UI first:
cd ../ui
pnpm build:server  # Outputs to ../ui-server/ui/

# Then run UI server:
cd ../ui-server
go run ./cmd/server
```

## Monitoring & Operations

### Logs
- **Structured logging** with request ID
- **Log levels**: DEBUG, INFO, WARN, ERROR
- **Request logging**: Method, path, status, duration

### Metrics
- **HTTP request count** by route and status
- **Request duration** histogram
- **Active connections** gauge
- **Auth success/failure** counter

### Health Checks

```bash
# Liveness
curl http://localhost:8080/health

# Readiness (checks Temporal connection)
curl http://localhost:8080/api/v1/cluster
```

## Implementation Notes

### Echo Framework
Uses Echo v4 for HTTP routing and middleware:

```go
e := echo.New()
e.Use(middleware.Logger())
e.Use(middleware.Recover())
e.Use(corsMiddleware)
e.Use(csrfMiddleware)
```

### gRPC Client Lifecycle
- Connections created at startup
- Automatic reconnection on failure
- Graceful shutdown on server stop

### Error Handling
- HTTP status codes mapped from gRPC codes
- JSON error responses with details
- Request ID for tracing

## References

- [UI Repository](https://github.com/temporalio/ui)
- [gRPC-Gateway Documentation](https://grpc-ecosystem.github.io/grpc-gateway/)
- [Echo Framework](https://echo.labstack.com/)
