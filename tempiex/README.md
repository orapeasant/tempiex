# tempiex

Core Tempiex workflow server. Exposes a gRPC API for workflow orchestration, task queue management, and activity execution. Backed by PostgreSQL.

## Architecture

```
tempiex (gRPC :8133)
    │
    ├── WorkflowService — start/describe/terminate workflows
    ├── Task Queue matching — poll & dispatch activities
    └── PostgreSQL — persistence layer
```

## Quick Start

### Prerequisites

- Go 1.24+
- PostgreSQL 15+

### 1. Configure

Copy and edit the development config:

```bash
cp config/development.yaml config/local.yaml
```

Key fields in `config/development.yaml`:

```yaml
server:
  grpcPort: 8133

database:
  host: localhost
  port: 5432
  user: tempiex
  password: tempiex
  dbName: tempiex
  sslMode: disable
```

Override any value with environment variables:

| Env var | Default |
|---|---|
| `TEMPIEX_GRPC_PORT` | `8133` |
| `TEMPIEX_DB_HOST` | `localhost` |
| `TEMPIEX_DB_USER` | `tempiex` |
| `TEMPIEX_DB_PASSWORD` | `tempiex` |
| `TEMPIEX_DB_NAME` | `tempiex` |

### 2. Create the database

```bash
psql -U postgres -c "CREATE USER tempiex WITH PASSWORD 'tempiex';"
psql -U postgres -c "CREATE DATABASE tempiex OWNER tempiex;"
```

### 3. Build and run

```bash
cd tempiex/
go build -o bin/tempiex ./cmd/tempiex
./bin/tempiex --config config/development.yaml
```

The server starts on gRPC port **8133**.

## Installation

```bash
go install github.com/tempiex/tempiex/cmd/tempiex@latest
```

## Development

```bash
# Run tests
go test ./...

# Lint
golangci-lint run ./...

# Build only
go build ./...
```

## gRPC API

The `WorkflowService` proto is at `api/proto/tempiex/api/workflowservice/v1/service.proto`.

Key RPCs:

| RPC | Description |
|---|---|
| `RegisterNamespace` | Create a workflow namespace |
| `StartWorkflowExecution` | Start a new workflow |
| `DescribeWorkflowExecution` | Get workflow status |
| `PollActivityTaskQueue` | Long-poll for activity tasks (used by workers) |
| `RespondActivityTaskCompleted` | Complete an activity task |
| `RespondActivityTaskFailed` | Fail an activity task |
| `RecordActivityTaskHeartbeat` | Send heartbeat from a running activity |

## Docker

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY . .
RUN go build -o /tempiex ./cmd/tempiex

FROM alpine:3.21
COPY --from=builder /tempiex /usr/local/bin/tempiex
ENTRYPOINT ["/usr/local/bin/tempiex"]
```
