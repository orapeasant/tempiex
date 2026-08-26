# 00 - System Architecture Overview

## System Overview

This document provides a comprehensive overview of the Temporal ecosystem architecture, including all four major components and their interactions.

## Component Summary

| Component | Priority | Type | Language | Purpose |
|-----------|----------|------|----------|---------|
| **Temporal Server** | 1 | Core Backend | Go 1.26.4 | Durable execution platform, workflow orchestration |
| **Temporal Proxy** | 2 | gRPC Proxy | Go 1.26.4 | Connection abstraction, namespace translation, encryption |
| **UI Server** | 3 | HTTP Gateway | Go 1.26.3 | HTTP-to-gRPC bridge for web UI |
| **Web UI** | 4 | Frontend SPA | TypeScript/Svelte | User interface for managing workflows |

## Implementation Sequence

### Phase 1: Foundation (Priority 1)
**Temporal Server** must be implemented first as it provides the core API and functionality that all other components depend on.

**Deliverables**:
- Temporal Server binary
- Database schemas (PostgreSQL/MySQL/Cassandra/SQLite)
- gRPC API endpoints
- Core services (Frontend, History, Matching, Worker)

**Why First?**:
- Foundation for all other components
- No dependencies on other components
- Provides the core Temporal API

### Phase 2: Proxy Layer (Priority 2)
**Temporal Proxy** can be implemented next to provide connection abstraction and advanced features.

**Deliverables**:
- Proxy binary
- Routing engine
- Namespace translation
- Optional: Payload encryption

**Why Second?**:
- Depends only on Temporal Server
- Optional component (can be skipped if not needed)
- Enables advanced deployment scenarios

### Phase 3: API Gateway (Priority 3)
**UI Server** implements the HTTP-to-gRPC gateway for web browser access.

**Deliverables**:
- UI Server binary
- HTTP API endpoints
- Authentication integration
- CORS and CSRF protection

**Why Third?**:
- Depends on Temporal Server (or Temporal Proxy)
- Required for Web UI
- Relatively simple component

### Phase 4: User Interface (Priority 4)
**Web UI** provides the graphical interface for end users.

**Deliverables**:
- Built UI assets (JavaScript, CSS, HTML)
- Component library
- Routing and state management

**Why Last?**:
- Depends on UI Server for API
- End-user facing component
- Most complex frontend development

## System Architecture

### Full System Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│                         End Users & Applications                     │
│                                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │ Web Browser  │  │ SDK Client   │  │   Worker     │             │
│  │  (Web UI)    │  │              │  │              │             │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘             │
│         │                 │                  │                      │
└─────────┼─────────────────┼──────────────────┼──────────────────────┘
          │ HTTP            │ gRPC             │ gRPC
          │                 │                  │
┌─────────▼─────────────────┼──────────────────┼──────────────────────┐
│         │                 │                  │                      │
│    ┌────▼──────────┐      │                  │                      │
│    │  UI Server    │      │                  │                      │
│    │  (Port 8080)  │      │                  │                      │
│    │  • HTTP API   │      │                  │                      │
│    │  • Auth       │      │                  │                      │
│    │  • CORS       │      │                  │                      │
│    └────┬──────────┘      │                  │                      │
│         │ gRPC            │                  │                      │
│         └─────────────────┼──────────────────┘                      │
│                           │                                         │
│                    Optional: Temporal Proxy                         │
│         ┌─────────────────▼─────────────────────────┐               │
│         │      Temporal Proxy (Port 7233)           │               │
│         │      • Namespace translation               │               │
│         │      • Payload encryption                  │               │
│         │      • TLS termination                     │               │
│         │      • Multi-upstream routing              │               │
│         └─────────────────┬─────────────────────────┘               │
│                           │ gRPC                                    │
└───────────────────────────┼─────────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────────┐
│                      Temporal Server                                │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │   Frontend   │  │   History    │  │   Matching   │             │
│  │   Service    │  │   Service    │  │   Service    │             │
│  │   :7233      │  │   :7234      │  │   :7235      │             │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘             │
│         │                 │                  │                      │
│         │          ┌──────▼──────┐           │                      │
│         │          │   Worker    │           │                      │
│         │          │   Service   │           │                      │
│         │          │   :7239     │           │                      │
│         │          └──────┬──────┘           │                      │
│         │                 │                  │                      │
│         └─────────────────┴──────────────────┘                      │
│                           │                                         │
│                  ┌────────▼────────┐                                │
│                  │  Persistence    │                                │
│                  │     Layer       │                                │
│                  └────────┬────────┘                                │
└───────────────────────────┼─────────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────────┐
│                       Data Stores                                   │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │ PostgreSQL   │  │    MySQL     │  │  Cassandra   │             │
│  │ (Primary)    │  │  (Primary)   │  │  (Primary)   │             │
│  └──────────────┘  └──────────────┘  └──────────────┘             │
│                                                                     │
│  ┌──────────────┐  ┌──────────────┐                                │
│  │    SQLite    │  │Elasticsearch │                                │
│  │   (Dev)      │  │ (Visibility) │                                │
│  └──────────────┘  └──────────────┘                                │
└─────────────────────────────────────────────────────────────────────┘
```

## Data Flow

### Workflow Execution Flow

```
1. SDK Client starts workflow
   ↓ gRPC: StartWorkflowExecution
2. Temporal Server (Frontend Service)
   ↓ Validates request
3. Temporal Server (History Service)
   ↓ Creates workflow execution
   ↓ Stores in persistence layer
4. Temporal Server (Matching Service)
   ↓ Enqueues workflow task
5. Worker polls for tasks
   ↓ gRPC: PollWorkflowTaskQueue
6. Worker executes workflow code
   ↓ Returns completed task
7. Temporal Server (History Service)
   ↓ Appends events to history
   ↓ Schedules activity tasks
8. Repeat steps 5-7 for activities
9. Workflow completes
   ↓ Results returned to client
```

### UI Data Flow

```
1. User opens Web UI in browser
   ↓ HTTP: GET /
2. UI Server serves static assets
   ↓ Returns HTML/JS/CSS
3. Web UI loads in browser
   ↓ Renders UI
4. User requests workflow list
   ↓ HTTP: GET /api/v1/namespaces/default/workflows
5. UI Server
   ↓ Converts HTTP to gRPC
   ↓ gRPC: ListWorkflowExecutions
6. Temporal Server (Frontend Service)
   ↓ Queries persistence layer
   ↓ Returns workflow list
7. UI Server
   ↓ Converts gRPC to JSON
   ↓ Returns HTTP response
8. Web UI
   ↓ Renders workflow list
```

### Proxy-Enabled Flow

```
1. Worker connects to Proxy
   ↓ gRPC (plaintext): localhost:7233
2. Temporal Proxy (Gateway)
   ↓ Peeks namespace from metadata
   ↓ Routes to upstream proxy
3. Temporal Proxy (Upstream Proxy)
   ↓ Translates namespace name
   ↓ Encrypts payloads (optional)
   ↓ Adds credentials (API key/mTLS)
   ↓ gRPC (TLS): Temporal Cloud
4. Temporal Cloud
   ↓ Processes request
   ↓ Returns response
5. Temporal Proxy (Upstream Proxy)
   ↓ Decrypts payloads (optional)
   ↓ Translates namespace back
6. Temporal Proxy (Gateway)
   ↓ Returns to worker
7. Worker receives response
```

## Port Allocation

| Component | Port | Protocol | Purpose |
|-----------|------|----------|---------|
| **Temporal Server** | | | |
| Frontend Service | 7233 | gRPC | Main API |
| Frontend HTTP | 7243 | HTTP | Health checks |
| History Service | 7234 | gRPC | Internal |
| Matching Service | 7235 | gRPC | Internal |
| Worker Service | 7239 | gRPC | Internal |
| Metrics | 8000 | HTTP | Prometheus |
| pprof | 7936 | HTTP | Profiling |
| **Temporal Proxy** | | | |
| Gateway | 7233 | gRPC | Main proxy endpoint |
| Metrics | 9090 | HTTP | Prometheus |
| **UI Server** | | | |
| HTTP Server | 8080 | HTTP | Web UI & API |
| **Extension Servers** | | | |
| Custom KMS | 9000 | gRPC | Key management |
| Custom Auth | 9001 | gRPC | Authentication |

## Technology Stack Comparison

### Languages

| Component | Language | Version | Rationale |
|-----------|----------|---------|-----------|
| Temporal Server | Go | 1.26.4 | High performance, excellent concurrency, strong ecosystem |
| Temporal Proxy | Go | 1.26.4 | Consistent with server, efficient gRPC handling |
| UI Server | Go | 1.26.3 | Lightweight, simple HTTP server |
| Web UI | TypeScript | 6.0.3 | Type safety, excellent tooling, modern frontend |

### Frameworks

| Component | Framework | Purpose |
|-----------|-----------|---------|
| Temporal Server | uber-go/fx | Dependency injection |
| Temporal Proxy | uber-go/fx | Dependency injection |
| UI Server | Echo v4 | HTTP routing & middleware |
| Web UI | SvelteKit | Full-stack web framework |
| Web UI | Svelte 5 | Reactive UI components |

### RPC & APIs

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Temporal Server | gRPC, Protobuf | Service-to-service, SDK communication |
| Temporal Proxy | gRPC, Protobuf | Proxy passthrough |
| UI Server | gRPC (client), HTTP (server) | Bridge gRPC ↔ HTTP |
| Web UI | HTTP REST | Browser-based API calls |

### Data Stores

| Database | Usage | Components |
|----------|-------|------------|
| PostgreSQL | Primary persistence & visibility | Temporal Server |
| MySQL | Primary persistence & visibility | Temporal Server |
| Cassandra | Primary persistence (large scale) | Temporal Server |
| SQLite | Development & testing | Temporal Server |
| Elasticsearch | Advanced visibility search | Temporal Server |

### Observability

| Component | Metrics | Logging | Tracing |
|-----------|---------|---------|---------|
| Temporal Server | Prometheus, OpenTelemetry | Structured (zap) | OpenTelemetry |
| Temporal Proxy | Prometheus | Structured (zerolog) | - |
| UI Server | - | Echo middleware | - |
| Web UI | - | Console | - |

## Security Architecture

### Authentication

| Component | Methods |
|-----------|---------|
| Temporal Server | mTLS, JWT, OAuth 2.0, API keys |
| Temporal Proxy | JWT, Static tokens, Custom (extension) |
| UI Server | OIDC, OAuth 2.0 |
| Web UI | Session cookies (via UI Server) |

### Encryption

| Layer | Technology | Components |
|-------|-----------|------------|
| Transport | TLS 1.2+ | All components |
| Payload | Envelope encryption (AES-256-GCM) | Temporal Proxy |
| At Rest | Database encryption (native) | Data stores |

### Secure Communication Paths

```
Web UI ──[HTTPS + Session Cookie]──> UI Server
         ┌──────────────────────────────┘
         │
         └──[gRPC + mTLS]──> Temporal Server
                              or
                             Temporal Proxy ──[gRPC + API Key/mTLS]──> Temporal Server
```

## Deployment Patterns

### Pattern 1: Simple Development

```
┌─────────────────┐
│  Developer PC   │
│                 │
│  ┌───────────┐  │
│  │ Temporal  │  │
│  │  Server   │  │
│  │  (SQLite) │  │
│  └─────┬─────┘  │
│        │        │
│  ┌─────▼─────┐  │
│  │UI Server  │  │
│  └─────┬─────┘  │
│        │        │
│  ┌─────▼─────┐  │
│  │  Browser  │  │
│  └───────────┘  │
└─────────────────┘
```

**Use Case**: Local development, testing  
**Components**: Temporal Server + UI Server + Web UI  
**Database**: SQLite in-memory  

### Pattern 2: Production Self-Hosted

```
┌────────────────────────────────────────────────┐
│                  Production Cluster            │
│                                                │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐  │
│  │ Frontend  │  │  History  │  │ Matching  │  │
│  │  Service  │  │  Service  │  │  Service  │  │
│  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘  │
│        └───────────────┴───────────────┘        │
│                        │                        │
│                 ┌──────▼──────┐                 │
│                 │ PostgreSQL  │                 │
│                 │  (Primary)  │                 │
│                 └─────────────┘                 │
└────────────────────┬───────────────────────────┘
                     │
┌────────────────────▼───────────────────────────┐
│                 DMZ / Load Balancer            │
│                                                │
│  ┌───────────┐           ┌───────────┐         │
│  │UI Server  │           │  Proxy    │         │
│  └─────┬─────┘           └─────┬─────┘         │
│        │                       │                │
└────────┼───────────────────────┼────────────────┘
         │                       │
         │ HTTPS                 │ gRPC/TLS
         │                       │
    ┌────▼────┐            ┌─────▼──────┐
    │ Browser │            │ SDK Clients│
    └─────────┘            └────────────┘
```

**Use Case**: Enterprise deployment  
**Components**: All components  
**Database**: PostgreSQL with replication  
**Features**: Load balancing, TLS, authentication

### Pattern 3: Temporal Cloud with Proxy

```
┌──────────────────────┐
│   Your Infrastructure│
│                      │
│  ┌────────────────┐  │
│  │ Temporal Proxy │  │
│  │  • Namespace   │  │
│  │    translation │  │
│  │  • Encryption  │  │
│  └────────┬───────┘  │
│           │          │
└───────────┼──────────┘
            │ gRPC/TLS + API Key
            │
┌───────────▼──────────┐
│   Temporal Cloud     │
│                      │
│  ┌────────────────┐  │
│  │ Temporal Server│  │
│  │  (Managed)     │  │
│  └────────────────┘  │
└──────────────────────┘
```

**Use Case**: Using Temporal Cloud with abstraction  
**Components**: Temporal Proxy + Temporal Cloud  
**Features**: Namespace abstraction, optional encryption, credential management

## Scaling Considerations

### Temporal Server

| Dimension | Strategy |
|-----------|----------|
| Horizontal | Scale each service independently |
| Vertical | Increase CPU/memory per instance |
| Sharding | History shards distribute load |
| Database | Read replicas for visibility |

**Limits**:
- 10,000+ workflow starts/sec
- 100,000+ activity executions/sec
- Millions of concurrent workflows

### Temporal Proxy

| Dimension | Strategy |
|-----------|----------|
| Horizontal | Multiple proxy instances behind load balancer |
| Vertical | Increase CPU for encryption workload |
| Caching | Cache KMS-wrapped keys |

**Limits**:
- 10,000+ req/sec without encryption
- 5,000+ req/sec with encryption

### UI Server

| Dimension | Strategy |
|-----------|----------|
| Horizontal | Multiple instances behind load balancer |
| Vertical | Increase memory for connection pooling |
| Caching | Cache static assets with CDN |

**Limits**:
- 5,000+ req/sec
- 1000+ concurrent connections

### Web UI

| Dimension | Strategy |
|-----------|----------|
| CDN | Serve static assets from CDN |
| Code Splitting | Lazy load routes and components |
| Caching | Browser caching for assets |

## Monitoring & Observability

### Metrics Collection

| Component | Exporter | Port | Metrics |
|-----------|----------|------|---------|
| Temporal Server | Prometheus | 8000 | Workflow counts, latencies, errors |
| Temporal Proxy | Prometheus | 9090 | Request counts, encryption latency |
| UI Server | Built-in | - | HTTP metrics |

### Log Aggregation

All components use structured logging:
- **Temporal Server**: uber-go/zap
- **Temporal Proxy**: rs/zerolog
- **UI Server**: Echo middleware
- **Web UI**: Browser console

### Tracing

- **Temporal Server**: OpenTelemetry
- **Temporal Proxy**: (future)
- **UI Server**: (future)

### Health Checks

| Component | Endpoint | Port |
|-----------|----------|------|
| Temporal Server | `/health` | 7243 |
| Temporal Proxy | `/health` (gRPC Health) | 7233 |
| UI Server | `/health` | 8080 |

## Development Workflow

### Local Development Setup

```bash
# 1. Start Temporal Server
cd temporal/
make temporal-server
./temporal-server --config config/development.yaml

# 2. (Optional) Start Temporal Proxy
cd temporal-proxy/
go run ./cmd/proxy --config examples/local/config.yaml

# 3. Start UI Server
cd ui-server/
go run ./cmd/server --temporal-grpc-address localhost:7233

# 4. Start Web UI dev server
cd ui/
pnpm install
pnpm dev
```

### Testing Strategy

| Component | Unit Tests | Integration Tests | E2E Tests |
|-----------|-----------|-------------------|-----------|
| Temporal Server | Go test | Go test with test cluster | - |
| Temporal Proxy | Go test | Go test with mock server | - |
| UI Server | Go test | - | - |
| Web UI | Vitest | Vitest | Playwright |

## Configuration Management

### Environment Variables

All components support environment variable substitution in config files:

```yaml
# Example
temporalGrpcAddress: "${TEMPORAL_ADDRESS}"
auth:
  clientSecret: "${OAUTH_CLIENT_SECRET}"
```

### Configuration Files

| Component | Format | Location |
|-----------|--------|----------|
| Temporal Server | YAML | `config/development.yaml` |
| Temporal Proxy | YAML | User-provided |
| UI Server | YAML | User-provided |
| Web UI | Env vars | Build-time only |

## Migration & Upgrades

### Version Compatibility

| Component Pair | Compatibility |
|----------------|---------------|
| Server ↔ SDK | Server: N, SDK: N-2 to N |
| Server ↔ Proxy | Same major version |
| Server ↔ UI Server | Same major version recommended |
| UI Server ↔ Web UI | Bundled together |

### Database Migrations

Temporal Server includes migration tools:
- `temporal-sql-tool` - PostgreSQL, MySQL, SQLite
- `temporal-cassandra-tool` - Cassandra

### Rolling Upgrades

- **Temporal Server**: Supported (upgrade history → matching → frontend)
- **Temporal Proxy**: Restart required
- **UI Server**: Restart required
- **Web UI**: Deploy new assets

## Cost Optimization

### Development
- Use SQLite for local development (zero infrastructure cost)
- Single Temporal Server instance
- No proxy needed

### Production
- **Self-hosted**:
  - 3+ node Temporal Server cluster
  - Managed PostgreSQL
  - Optional Elasticsearch
  - Proxy for multi-environment abstraction
  
- **Temporal Cloud**:
  - Pay per action (workflow/activity execution)
  - No infrastructure management
  - Proxy for encryption/namespace abstraction

## Troubleshooting

### Common Issues

| Issue | Component | Solution |
|-------|-----------|----------|
| Connection refused | All | Check port bindings, firewall rules |
| TLS handshake failed | Server/Proxy | Verify certificates, server names |
| Namespace not found | Server | Register namespace first |
| Auth failed | UI Server | Check OIDC config, client ID/secret |
| CORS error | UI Server | Add origin to allowOrigins |

### Diagnostic Commands

```bash
# Temporal Server health
curl http://localhost:7243/health

# Temporal Proxy health (gRPC)
grpcurl -plaintext localhost:7233 grpc.health.v1.Health/Check

# UI Server health
curl http://localhost:8080/health

# Check Temporal Server connectivity
temporal operator cluster health
```

## References

### Detailed Specifications
1. [01 - Temporal Server Core Specification](./01-temporal-server-core-specification.md)
2. [02 - Temporal Proxy Specification](./02-temporal-proxy-specification.md)
3. [03 - UI Server Specification](./03-ui-server-specification.md)
4. [04 - Temporal Web UI Specification](./04-temporal-web-ui-specification.md)

### External Documentation
- [Temporal Documentation](https://docs.temporal.io/)
- [Temporal Server Repository](https://github.com/temporalio/temporal)
- [Temporal Proxy Repository](https://github.com/temporalio/temporal-proxy)
- [Temporal UI Repository](https://github.com/temporalio/ui)
- [gRPC Documentation](https://grpc.io/)
- [SvelteKit Documentation](https://kit.svelte.dev/)
