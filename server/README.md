# server

Tempiex UI gateway server. Bridges the React web UI to the Tempiex gRPC core over HTTP/JSON. Runs on port **8080** by default.

## Architecture

```
Browser / CLI
    │  HTTP :8080
    ▼
server (HTTP gateway)
    │  gRPC :8133
    ▼
tempiex (core)
```

Provides REST-style JSON endpoints that translate to Tempiex gRPC calls, plus authentication middleware and CORS handling.

## Quick Start

### Prerequisites

- Go 1.24+
- A running `tempiex` instance on `localhost:8133`

### 1. Configure

Edit `config/development.yaml`:

```yaml
tempiexGrpcAddress: "localhost:8133"
host: "0.0.0.0"
port: 8080
dbDsn: "postgres://tempiex:tempiex@localhost:5432/tempiex"
auth:
  enabled: false
cors:
  cookieInsecure: true
```

### 2. Build and run

```bash
cd server/
go build -o bin/server ./cmd/server
./bin/server
```

The HTTP server starts on port **8080**.

## Installation

```bash
go install github.com/tempiex/server/cmd/server@latest
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

## API Endpoints

The server exposes HTTP endpoints used by the web UI and CLI. All endpoints return JSON.

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check |
| `GET` | `/api/namespaces` | List namespaces |
| `GET` | `/api/workflows` | List workflow executions |
| `POST` | `/api/workflows` | Start a workflow |
| `GET` | `/api/workflows/:id` | Describe a workflow |
| `DELETE` | `/api/workflows/:id` | Terminate a workflow |

## Docker

```bash
docker run \
  -e TEMPIEX_GRPC_ADDRESS=host.docker.internal:8133 \
  -p 8080:8080 \
  tempiex/server:latest
```
