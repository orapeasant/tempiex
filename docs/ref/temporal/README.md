# Temporal Ecosystem - Specification Index

## Documentation Overview

This documentation suite provides comprehensive technical specifications for the Temporal ecosystem, covering architecture, design, implementation, and deployment across all four major components.

## Document Structure

The specifications are numbered to represent the recommended **implementation sequence**:

### 00 - System Architecture Overview
**Purpose**: High-level system architecture, component interactions, deployment patterns  
**Audience**: Architects, technical leads, operations teams  
**Topics**: 
- Complete system diagram
- Component dependencies
- Data flow
- Deployment patterns
- Scaling strategies
- Security architecture

📄 [Read: System Architecture Overview](./00-system-architecture-overview.md)

---

### 01 - Temporal Server Core Specification
**Purpose**: Foundation component - distributed workflow orchestration engine  
**Audience**: Backend developers, platform engineers, DBAs  
**Priority**: **1 (Implement First)**  
**Topics**:
- Core services (Frontend, History, Matching, Worker)
- Persistence layer and database support
- gRPC API architecture
- Workflow execution model
- Namespace management
- Observability and metrics

**Tech Stack**:
- Go 1.26.4
- gRPC + Protobuf
- PostgreSQL/MySQL/Cassandra/SQLite
- Prometheus + OpenTelemetry
- uber-go/fx dependency injection

📄 [Read: Temporal Server Specification](./01-temporal-server-core-specification.md)

---

### 02 - Temporal Proxy Specification
**Purpose**: gRPC proxy for connection abstraction and advanced features  
**Audience**: DevOps engineers, security engineers, platform teams  
**Priority**: **2 (Optional - After Temporal Server)**  
**Topics**:
- Gateway and routing architecture
- Namespace translation
- Payload encryption (envelope encryption)
- KMS integration (AWS/Azure/GCP)
- Extension server framework
- Multi-upstream routing

**Tech Stack**:
- Go 1.26.4
- gRPC + Protobuf
- gocloud.dev for KMS
- zerolog for logging
- Prometheus metrics
- uber-go/fx dependency injection

📄 [Read: Temporal Proxy Specification](./02-temporal-proxy-specification.md)

---

### 03 - UI Server Specification
**Purpose**: HTTP-to-gRPC gateway for web browser access  
**Audience**: Full-stack developers, API developers  
**Priority**: **3 (Depends on Temporal Server)**  
**Topics**:
- HTTP server architecture
- gRPC client implementation
- Authentication (OIDC/OAuth 2.0)
- CORS and CSRF protection
- API endpoint mappings
- Session management

**Tech Stack**:
- Go 1.26.3
- Echo v4 HTTP framework
- gRPC client
- go-oidc/v3 for authentication
- gorilla/securecookie for sessions

📄 [Read: UI Server Specification](./03-ui-server-specification.md)

---

### 04 - Temporal Web UI Specification
**Purpose**: Modern web application for workflow management  
**Audience**: Frontend developers, UI/UX engineers  
**Priority**: **4 (Depends on UI Server)**  
**Topics**:
- SvelteKit application architecture
- Component library (Holocene design system)
- State management with Svelte 5 runes
- API client layer
- Real-time updates and polling
- Advanced search and filtering

**Tech Stack**:
- TypeScript 6.0.3
- SvelteKit 2.57.1 + Svelte 5.55.7
- TailwindCSS 3.4.1
- CodeMirror 6 for code editing
- Vitest + Playwright for testing
- pnpm package manager

📄 [Read: Web UI Specification](./04-temporal-web-ui-specification.md)

---

## Quick Reference

### Component Dependencies

```
Temporal Server (1)
    ├─── Temporal Proxy (2) [Optional]
    │       └─── Extension Servers [Optional]
    │
    ├─── UI Server (3)
    │       └─── Web UI (4)
    │
    └─── SDK Clients / Workers
```

### Port Allocation Matrix

| Component | Default Port | Protocol | Purpose |
|-----------|-------------|----------|---------|
| **Temporal Server** | | | |
| Frontend Service | 7233 | gRPC | Main API |
| Frontend HTTP | 7243 | HTTP | Health |
| Metrics | 8000 | HTTP | Prometheus |
| **Temporal Proxy** | | | |
| Gateway | 7233 | gRPC | Proxy endpoint |
| Metrics | 9090 | HTTP | Prometheus |
| **UI Server** | | | |
| HTTP Server | 8080 | HTTP | Web UI + API |
| **Web UI** | | | |
| Dev Server | 3000 | HTTP | Development only |

### Technology Stack Summary

| Component | Language | Framework | Database | Build Tool |
|-----------|----------|-----------|----------|------------|
| Temporal Server | Go 1.26.4 | uber-go/fx | PostgreSQL/MySQL/Cassandra | Make |
| Temporal Proxy | Go 1.26.4 | uber-go/fx | - | Go build |
| UI Server | Go 1.26.3 | Echo v4 | - | Go build |
| Web UI | TypeScript 6.0.3 | SvelteKit/Svelte 5 | - | Vite |

### Implementation Checklist

#### Phase 1: Foundation
- [ ] Set up database (PostgreSQL/MySQL/Cassandra)
- [ ] Deploy Temporal Server
- [ ] Configure namespaces
- [ ] Set up monitoring (Prometheus)
- [ ] Run health checks
- [ ] Deploy test workflows

#### Phase 2: Proxy Layer (Optional)
- [ ] Configure Temporal Proxy
- [ ] Set up routing rules
- [ ] Configure namespace translation
- [ ] (Optional) Set up KMS for encryption
- [ ] (Optional) Deploy extension servers
- [ ] Test proxy connectivity

#### Phase 3: Web Gateway
- [ ] Deploy UI Server
- [ ] Configure Temporal Server connection
- [ ] Set up authentication (OIDC/OAuth)
- [ ] Configure CORS
- [ ] Test API endpoints
- [ ] Verify health checks

#### Phase 4: User Interface
- [ ] Build Web UI assets
- [ ] Deploy static assets to UI Server
- [ ] Configure environment variables
- [ ] Test UI functionality
- [ ] Verify authentication flow
- [ ] Performance testing

### Common Configuration Patterns

#### Development Setup (Minimal)

```yaml
# Temporal Server (development.yaml)
persistence:
  defaultStore: sqlite-default
  visibilityStore: sqlite-visibility

services:
  frontend:
    rpc:
      grpcPort: 7233
      bindOnLocalHost: true

# UI Server (config.yaml)
temporalGrpcAddress: "localhost:7233"
publicUrl: "http://localhost:8080"

# No Temporal Proxy needed for local dev
```

#### Production Setup (Full)

```yaml
# Temporal Server (production.yaml)
persistence:
  defaultStore: postgres-default
  visibilityStore: postgres-visibility
  datastores:
    postgres-default:
      sql:
        pluginName: "postgres"
        databaseName: "temporal"
        connectAddr: "postgres.internal:5432"
        tls:
          enabled: true
          caFile: "/certs/ca.pem"

# Temporal Proxy (config.yaml)
gateway:
  hostPort: :7233
  tls:
    enabled: true
    certFile: /certs/server.crt
    keyFile: /certs/server.key

upstreams:
  - name: production
    hostPort: temporal-frontend:7233
    tls:
      enabled: true

routing:
  default: production

# UI Server (config.yaml)
temporalGrpcAddress: "temporal-proxy:7233"
publicUrl: "https://temporal.company.com"

auth:
  enabled: true
  providers:
    - type: oidc
      issuerUrl: "https://sso.company.com"
      clientId: "temporal-ui"
```

## Getting Help

### Official Resources
- **Documentation**: https://docs.temporal.io/
- **Community Forum**: https://community.temporal.io/
- **Slack**: https://t.mp/slack
- **GitHub**: https://github.com/temporalio/

### Issue Reporting
- Temporal Server: https://github.com/temporalio/temporal/issues
- Temporal Proxy: https://github.com/temporalio/temporal-proxy/issues
- UI & UI Server: https://github.com/temporalio/ui/issues

### Component-Specific Documentation
Each component has additional documentation:
- **Temporal Server**: `temporal/docs/architecture/`
- **Temporal Proxy**: `temporal-proxy/rfc/`
- **UI Server**: `ui-server/README.md`
- **Web UI**: `ui/README.md`, `ui/docs/`

## Version Information

**Specification Version**: 1.0  
**Last Updated**: 2026-08-12  
**Compatible With**:
- Temporal Server: v1.30.x+
- Temporal Proxy: v0.x (pre-release)
- UI Server: v2.53.x+
- Web UI: v2.53.x+

## Maintenance

These specifications should be updated when:
- Major architectural changes occur
- New components are added
- Significant features are introduced
- Breaking changes are released
- Deployment patterns change

## Contributing

To update these specifications:
1. Identify the affected component(s)
2. Update the relevant specification document
3. Update this index if structure changes
4. Verify cross-references between documents
5. Update version information

## License

This documentation follows the same license as the respective components:
- Temporal Server, Proxy, UI: MIT License
- See individual repositories for details

---

**Navigation**:
- [⬆️ System Architecture Overview](./00-system-architecture-overview.md)
- [1️⃣ Temporal Server Specification](./01-temporal-server-core-specification.md)
- [2️⃣ Temporal Proxy Specification](./02-temporal-proxy-specification.md)
- [3️⃣ UI Server Specification](./03-ui-server-specification.md)
- [4️⃣ Web UI Specification](./04-temporal-web-ui-specification.md)
