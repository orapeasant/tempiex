# 01 - Temporal Server Core Specification

## Overview

**Component**: Temporal Server (Core Backend)  
**Repository**: `temporal/`  
**Language**: Go 1.26.4  
**Type**: Distributed Workflow Orchestration Engine  
**Implementation Priority**: 1 (Foundation - Must be implemented first)

## Purpose

Temporal Server is the core durable execution platform that enables developers to build scalable, reliable applications. It executes units of application logic called Workflows in a resilient manner, automatically handling intermittent failures and retrying failed operations.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Temporal Server                          │
│                                                             │
│  ┌────────────┐  ┌────────────┐  ┌────────────┐           │
│  │  Frontend  │  │  History   │  │  Matching  │           │
│  │  Service   │  │  Service   │  │  Service   │           │
│  │  :7233     │  │  :7234     │  │  :7235     │           │
│  └─────┬──────┘  └─────┬──────┘  └─────┬──────┘           │
│        │               │               │                    │
│        └───────────────┴───────────────┘                    │
│                        │                                     │
│                 ┌──────┴──────┐                             │
│                 │   Worker    │                             │
│                 │   Service   │                             │
│                 │   :7239     │                             │
│                 └──────┬──────┘                             │
│                        │                                     │
│        ┌───────────────┴────────────────┐                   │
│        │                                │                   │
│  ┌─────▼──────┐                  ┌──────▼──────┐           │
│  │Persistence │                  │  Membership │           │
│  │   Layer    │                  │   Manager   │           │
│  └────────────┘                  └─────────────┘           │
└─────────────────────────────────────────────────────────────┘
         │                                                      
         ▼                                                      
┌─────────────────┐                                           
│   Data Stores   │                                           
│ - PostgreSQL    │                                           
│ - MySQL         │                                           
│ - Cassandra     │                                           
│ - SQLite        │                                           
│ - Elasticsearch │                                           
└─────────────────┘                                           
```

### Core Services

#### 1. Frontend Service (Port 7233)
- **Purpose**: API Gateway for all client requests
- **Responsibilities**:
  - Exposes gRPC API to SDK clients and workers
  - Request validation and authentication
  - Rate limiting and quota management
  - Request routing to appropriate services
  - HTTP API endpoint (Port 7243)

#### 2. History Service (Port 7234)
- **Purpose**: Workflow execution engine
- **Responsibilities**:
  - Manages workflow execution state
  - Processes workflow tasks
  - Maintains event history
  - Handles workflow timers and signals
  - Implements workflow versioning
  - Manages workflow determinism

#### 3. Matching Service (Port 7235)
- **Purpose**: Task queue management and task routing
- **Responsibilities**:
  - Manages task queues (workflow and activity)
  - Routes tasks to appropriate workers
  - Implements task queue partitioning
  - Handles task queue rate limiting
  - Supports task versioning and build ID routing

#### 4. Worker Service (Port 7239)
- **Purpose**: Internal background processing
- **Responsibilities**:
  - System workflows (e.g., namespace replication)
  - Archival workflows
  - Scanner workflows
  - Visibility processing
  - Schedule workflows

## Technical Stack

### Core Technologies

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.26.4 |
| RPC Framework | gRPC | 1.80.0 |
| API Definition | Protobuf | 1.36.11 |
| Metrics | Prometheus | 1.21.0 |
| Observability | OpenTelemetry | 1.44.0 |

### Database Support

| Database | Purpose | Driver |
|----------|---------|--------|
| PostgreSQL | Primary persistence & visibility | pgx/v5 5.10.0 |
| MySQL | Primary persistence & visibility | go-sql-driver 1.9.3 |
| Cassandra | Primary persistence | gocql 1.7.0 |
| SQLite | Development & testing | modernc.org/sqlite 1.51.0 |
| Elasticsearch | Advanced visibility | olivere/elastic/v7 7.0.32 |

### Key Dependencies

```go
// Dependency Injection
go.uber.org/fx v1.24.0

// Workflow SDK
go.temporal.io/sdk v1.44.0
go.temporal.io/api v1.63.5

// Metrics & Monitoring
github.com/uber-go/tally/v4 v4.1.11
github.com/prometheus/client_golang v1.21.0
go.opentelemetry.io/otel v1.44.0

// Service Communication
google.golang.org/grpc v1.80.0
github.com/nexus-rpc/sdk-go v0.6.0

// Storage
github.com/aws/aws-sdk-go-v2 v1.41.6
cloud.google.com/go/storage v1.62.1

// Authentication
github.com/golang-jwt/jwt/v4 v4.5.2
golang.org/x/oauth2 v0.36.0
```

## Directory Structure

```
temporal/
├── api/                    # Proto definitions and generated code
├── chasm/                  # Coordinated Heterogeneous Application State Machines
├── client/                 # Inter-service client libraries
├── cmd/
│   ├── server/            # Main server binary
│   └── tools/             # Admin & migration tools
├── common/
│   ├── archiver/          # Workflow history archival
│   ├── clock/             # Time management
│   ├── config/            # Configuration management
│   ├── dynamicconfig/     # Dynamic configuration
│   ├── membership/        # Cluster membership
│   ├── metrics/           # Metrics definitions
│   ├── namespace/         # Namespace management
│   ├── nexus/             # Nexus integration
│   ├── persistence/       # Data persistence layer
│   ├── quotas/            # Rate limiting
│   └── sdk/               # Internal SDK utilities
├── components/            # Nexus components
├── config/                # Configuration files
│   ├── development.yaml   # Dev config
│   └── docker.yaml        # Docker config
├── proto/                 # Internal proto definitions
├── schema/                # Database schemas
│   ├── cassandra/
│   ├── mysql/
│   ├── postgresql/
│   └── sqlite/
├── service/
│   ├── frontend/          # Frontend service implementation
│   ├── history/           # History service implementation
│   ├── matching/          # Matching service implementation
│   └── worker/            # Worker service implementation
├── temporal/              # Main package
├── temporaltest/          # Testing utilities
└── tests/                 # Integration tests
```

## Key Features

### 1. Workflow Execution
- **Deterministic execution** with event sourcing
- **Automatic retries** with configurable policies
- **Timeouts** at multiple levels (schedule-to-start, start-to-close, etc.)
- **Signals and Queries** for external interaction
- **Child workflows** for composition
- **Continue-as-new** for long-running workflows

### 2. Task Management
- **Task queues** with sticky execution
- **Task routing** with versioning support
- **Build ID-based versioning** for gradual rollouts
- **Task priority** and rate limiting

### 3. Visibility & Search
- **Advanced search** using Elasticsearch
- **Standard visibility** using SQL databases
- **Custom search attributes**
- **Workflow filtering** and sorting

### 4. Data Management
- **Multi-region support** with XDC (Cross DC Replication)
- **Archival** to S3, GCS, or file system
- **Data encryption** at rest and in transit
- **Retention policies** per namespace

### 5. Namespace Isolation
- **Multi-tenancy** with namespace isolation
- **Per-namespace configuration**
- **Namespace migration** support
- **Global namespaces** for multi-cluster

### 6. Observability
- **Prometheus metrics** for monitoring
- **OpenTelemetry integration** for tracing
- **Structured logging** with configurable levels
- **Health checks** and readiness probes

### 7. Nexus Services
- **Cross-namespace operations**
- **Service-to-service communication**
- **Operation handlers**
- **Callback support**

## Configuration

### Minimal Development Configuration

```yaml
log:
  stdout: true
  level: info

persistence:
  defaultStore: sqlite-default
  visibilityStore: sqlite-visibility
  numHistoryShards: 1

global:
  membership:
    maxJoinDuration: 30s
    broadcastAddress: "127.0.0.1"
  metrics:
    prometheus:
      framework: "tally"
      timerType: "histogram"
      listenAddress: "127.0.0.1:8000"

services:
  frontend:
    rpc:
      grpcPort: 7233
      membershipPort: 6933
      bindOnLocalHost: true
      httpPort: 7243
  history:
    rpc:
      grpcPort: 7234
      membershipPort: 6934
      bindOnLocalHost: true
  matching:
    rpc:
      grpcPort: 7235
      membershipPort: 6935
      bindOnLocalHost: true
  worker:
    rpc:
      grpcPort: 7239
      membershipPort: 6939
      bindOnLocalHost: true
```

## Build & Deployment

### Build Commands

```bash
# Build server binary
make temporal-server

# Build all binaries
make bins

# Run all checks and tests
make all

# Generate proto files
make proto

# Run tests
make unit-test
```

### Binary Outputs

- `temporal-server` - Main server binary (all services)
- `temporal-cassandra-tool` - Cassandra schema management
- `temporal-sql-tool` - SQL schema management
- `temporal-elasticsearch-tool` - Elasticsearch index management
- `tdbg` - Debugging tool

### Environment Variables

- `TEMPORAL_ADDRESS` - Server address (default: localhost:7233)
- `TEMPORAL_NAMESPACE` - Default namespace
- `TEMPORAL_TLS_ENABLED` - Enable TLS
- `TEMPORAL_AUTH_ENABLED` - Enable authentication

## API Endpoints

### gRPC Services

| Service | Port | Proto Definition |
|---------|------|------------------|
| WorkflowService | 7233 | temporal.api.workflowservice.v1 |
| OperatorService | 7233 | temporal.api.operatorservice.v1 |
| AdminService | 7233 | temporal.server.api.adminservice.v1 |
| HealthService | 7233 | grpc.health.v1 |

### HTTP Endpoints

| Endpoint | Port | Purpose |
|----------|------|---------|
| /metrics | 8000 | Prometheus metrics |
| /health | 7243 | Health check |
| /debug/pprof | 7936 | Profiling |

## Dependencies & Integration Points

### Upstream Dependencies
- None (Foundation component)

### Downstream Dependents
- **UI Server** (consumes gRPC API)
- **Temporal Proxy** (proxies gRPC API)
- **SDK Clients** (connect via gRPC)
- **Workers** (poll tasks via gRPC)

## Security

### Authentication Methods
- **mTLS** for service-to-service
- **JWT tokens** for client authentication
- **API keys** for programmatic access
- **OAuth 2.0** integration support

### Authorization
- **Namespace-level permissions**
- **Role-based access control (RBAC)**
- **Custom authorizer plugins**

### Data Protection
- **TLS 1.2+** for transport encryption
- **At-rest encryption** for supported databases
- **Payload encryption** via data converters
- **PII handling** with codec server support

## Performance Characteristics

### Scalability
- **Horizontal scaling** of all services
- **Sharding** for workflow distribution
- **Multi-region** deployment support
- **Millions of workflows** per cluster

### Throughput
- **10,000+ workflow starts/sec** (per cluster)
- **100,000+ activity executions/sec**
- **Sub-second** task scheduling latency

### Resource Requirements

| Deployment | CPU | Memory | Storage |
|------------|-----|--------|---------|
| Development | 2 cores | 4 GB | 10 GB |
| Small Production | 8 cores | 16 GB | 100 GB |
| Large Production | 32+ cores | 64+ GB | 1+ TB |

## Testing

### Test Types
- **Unit tests** with testify
- **Integration tests** with test server
- **Functional tests** with SDK
- **Load tests** with benchmarking tools

### Test Commands

```bash
# Unit tests
make unit-test

# Integration tests (requires tags)
go test -tags integration ./...

# Coverage
make coverage
```

## Monitoring & Operations

### Key Metrics

| Metric | Type | Purpose |
|--------|------|---------|
| workflow_success_count | Counter | Successful completions |
| workflow_failed_count | Counter | Failed workflows |
| task_schedule_latency | Histogram | Task scheduling delay |
| persistence_latency | Histogram | Database operation time |
| service_requests | Counter | API request count |
| service_errors | Counter | API error count |

### Health Checks
- **Liveness**: Service process running
- **Readiness**: Database connection healthy
- **Startup**: Initialization complete

## Migration & Upgrades

### Database Migrations
- Schema versioning with migration tools
- Backward compatibility guarantees
- Rolling upgrade support

### Version Compatibility
- **Server-to-server**: Same major version
- **SDK-to-server**: N-2 minor versions
- **API compatibility**: Backward compatible

## Implementation Notes

### Dependency Injection Pattern
Uses `uber-go/fx` for dependency injection throughout the codebase.

```go
// Entry point pattern
fx.New(
    fx.Supply(config),
    service.Module,
    fx.Invoke(startServer),
)
```

### Error Handling
- Returns errors from library code
- Wraps errors with context using `fmt.Errorf`
- Uses standard error types (InvalidArgument, NotFound, etc.)
- Validates inputs early

### Concurrency
- Prefers `sync.Mutex` for synchronization
- Avoids holding locks across I/O
- Uses immutable data patterns where possible

## References

- [Architecture Documentation](./docs/architecture/README.md)
- [Workflow Lifecycle](./docs/architecture/workflow-lifecycle.md)
- [Matching Service](./docs/architecture/matching-service.md)
- [History Service](./docs/architecture/history-service.md)
- [Official Docs](https://docs.temporal.io/)
