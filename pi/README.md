# pi

Tempiex Pi — Agent Harness. A lightweight Go daemon that polls Tempiex task queues and executes workflow activities as typed tools. Designed for edge nodes, Raspberry Pi, Docker, or any host that needs to run Tempiex activities.

## Architecture

```
pi daemon
    │
    ├── Worker Pool — long-polls Tempiex gRPC for activity tasks
    ├── Tool Registry — typed activity implementations
    ├── Event Bus — streams execution events (start/update/end/error)
    ├── Session Manager — SQLite-backed session tracking
    └── HTTP API :8090 — inspect sessions and stream events via SSE
```

## Quick Start

### Prerequisites

- Go 1.24+
- A running `tempiex` instance on `localhost:8133`

### 1. Configure

Edit `config/development.yaml`:

```yaml
tempiex:
  address: "localhost:8133"
  namespace: "default"

workers:
  - taskQueue: "pi-default"
    maxConcurrentActivities: 10

api:
  host: "0.0.0.0"
  port: 8090

session:
  storePath: "./pi-sessions.db"

metrics:
  otel:
    endpoint: "localhost:4317"
    insecure: true

log:
  level: "info"
  format: "json"
```

Override any value with `PI_*` environment variables:

| Env var | Default |
|---|---|
| `PI_TEMPIEX_ADDRESS` | `localhost:8133` |
| `PI_TEMPIEX_NAMESPACE` | `default` |
| `PI_API_HOST` | `0.0.0.0` |
| `PI_API_PORT` | `8090` |
| `PI_LOG_LEVEL` | `info` |

### 2. Build and run

```bash
cd pi/
go build -o bin/pi ./cmd/pi
./bin/pi --config config/development.yaml
```

The HTTP API starts on port **8090**.

## Installation

```bash
go install github.com/tempiex/pi/cmd/pi@latest
```

## Registering Tools

Tools are the activity implementations that `pi` executes. Register them before starting the agent:

```go
package main

import (
    "context"
    "github.com/tempiex/pi/internal/tool"
    "github.com/tempiex/pi/internal/agent"
)

type GreetInput struct {
    Name string `json:"name"`
}

type GreetOutput struct {
    Message string `json:"message"`
}

func main() {
    registry := tool.NewRegistry()

    tool.Register(registry, tool.Tool[GreetInput, GreetOutput]{
        Name:        "greet",
        Description: "Return a greeting",
        Execute: func(ctx context.Context, input GreetInput) (GreetOutput, error) {
            return GreetOutput{Message: "Hello, " + input.Name + "!"}, nil
        },
    })

    ag := agent.New(agent.Options{Registry: registry})
    ag.Run(context.Background())
}
```

## Built-in Tools

| Tool name | Description |
|---|---|
| `shell` | Execute a shell command; captures stdout/stderr |
| `http_request` | Make an HTTP request with configurable method, headers, body |
| `file_read` | Read file contents from disk |
| `file_write` | Write content to a file on disk |

Use them by importing the tool packages and registering:

```go
import (
    pifile    "github.com/tempiex/pi/tools/file"
    pishell   "github.com/tempiex/pi/tools/shell"
    pihttp    "github.com/tempiex/pi/tools/httptool"
)

pifile.Register(registry)
pishell.Register(registry)
pihttp.Register(registry)
```

## HTTP API

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Health check; reports tempiex connectivity |
| `GET` | `/api/sessions` | List all tracked sessions |
| `GET` | `/api/sessions/:id` | Get a single session |
| `GET` | `/api/sessions/:id/events` | SSE stream of execution events |
| `POST` | `/api/sessions/:id/cancel` | Cancel a session |
| `GET` | `/api/tools` | List registered tools |

### SSE event stream

```
GET /api/sessions/:id/events
Content-Type: text/event-stream

data: {"type":"execution_start","toolName":"shell","payload":{"cmd":"ls"}}

data: {"type":"execution_end","toolName":"shell","payload":{"exitCode":0}}
```

Historical events for the session are replayed on connect.

## Event Types

| Type | Emitted when |
|---|---|
| `execution_start` | Tool execution begins |
| `execution_update` | Streaming progress update |
| `execution_end` | Tool completed successfully |
| `execution_error` | Tool returned an error |

## Development

```bash
cd pi/

# Build
go build -o bin/pi ./cmd/pi

# Run tests
go test ./...

# Integration tests (requires tempiex running)
go test -tags integration ./...

# Lint
golangci-lint run ./...
```

## Docker

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY . .
RUN go build -o /pi ./cmd/pi

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /pi /usr/local/bin/pi
ENTRYPOINT ["/usr/local/bin/pi"]
```

```bash
docker build -t tempiex/pi:latest .
docker run \
  -e PI_TEMPIEX_ADDRESS=host.docker.internal:8133 \
  -p 8090:8090 \
  tempiex/pi:latest
```

## Observability

Metrics are exported via OTel OTLP gRPC (no Prometheus scrape endpoint):

| Metric | Type | Labels |
|---|---|---|
| `pi.activity.executions.total` | Counter | `tool_name`, `status` |
| `pi.activity.duration` | Histogram | `tool_name` |
| `pi.sessions.active` | Gauge | — |
| `pi.worker.poll.latency` | Histogram | `task_queue` |

Configure the OTel collector endpoint with `PI_OTEL_ENDPOINT` or in `config.yaml`.
