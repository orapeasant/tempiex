# 02 — Proxy: gRPC Proxy Specification

## Overview

**Component**: Tempiex gRPC Proxy  
**Folder**: `proxy/`  
**Language**: Go 1.24+  
**Type**: gRPC Proxy Server  
**Implementation Priority**: 2 — Depends on tempiex (needs Tempiex API types)

## Purpose

The proxy sits between SDK clients/workers/UI and one or more upstream Tempiex services. It provides a single local endpoint and handles:

- **Namespace translation** — rewrite namespace names per environment
- **TLS termination** — inbound and outbound
- **Payload encryption** — envelope encryption via pluggable KMS backends
- **Multi-upstream routing** — namespace/metadata-based rules
- **Credential injection** — API keys, mTLS certificates per upstream

Clients target `proxy:8133` with no environment-specific config. The proxy handles environment differences transparently.

## Architecture

```
┌──────────────────────────────────────────────────────┐
│                Client Applications                   │
│  Workers │ SDK Clients │ UI Server │ CLI             │
└───────────────────────┬──────────────────────────────┘
                        │ gRPC (plaintext or TLS)
                        ▼
┌──────────────────────────────────────────────────────┐
│                    proxy                             │
│                                                      │
│  ┌───────────────────────────────────────────────┐  │
│  │           Gateway  :8133                      │  │
│  │  • namespace peek from metadata               │  │
│  │  • route to upstream proxy instance           │  │
│  │  • optional inbound auth (JWT / static token) │  │
│  │  • TLS termination                            │  │
│  └───────────────────┬───────────────────────────┘  │
│                      │                               │
│      ┌───────────────┼──────────────┐               │
│  ┌───▼────┐    ┌─────▼────┐   ┌────▼─────┐         │
│  │ Ups A  │    │  Ups B   │   │  Ups C   │         │
│  │namespace│   │ namespace│   │ system   │         │
│  │ xlate  │   │  + TLS   │   │ fallback │         │
│  │ + KMS  │   │  + mTLS  │   │          │         │
│  └───┬────┘    └─────┬────┘   └────┬─────┘         │
└──────┼───────────────┼─────────────┼────────────────┘
       │               │             │
       ▼               ▼             ▼
  Tempiex        Self-hosted    Local Dev
  Cloud           Tempiex       Tempiex

Optional Extension Server (custom KMS / auth):
┌─────────────────────────────────────┐
│       Extension Server (gRPC)       │
│  • Custom key wrapping (HSM)        │
│  • Custom authentication logic      │
└─────────────────────────────────────┘
```

## Component Breakdown

### Gateway
Single inbound gRPC endpoint. Peeks namespace from request metadata. Routes to the correct upstream proxy instance. Never parses workflow payloads — codec-transparent.

### Upstream Proxy Instances
One per configured upstream Tempiex service. Responsible for:
- Namespace translation (prefix, suffix, or exact map)
- Credential attachment (API key header, mTLS client cert)
- Optional payload encryption/decryption (AES-256-GCM + KMS-wrapped DEK)
- Outbound TLS configuration

### Router
Rule-based: matches by namespace name (exact, prefix, regex) or gRPC metadata. Has a `systemUpstream` for namespace-less requests and a `default` fallback.

### Crypto Module (optional)
Envelope encryption:
1. Generate DEK per operation (AES-256-GCM)
2. Wrap DEK with KEK from KMS
3. Store `wrappedDEK + ciphertext`
4. On read: unwrap DEK → decrypt

**KMS backends:** AWS KMS, Azure Key Vault, GCP Cloud KMS, Custom extension server (gRPC)

### Extension Server Interface
`pkg/ext` — gRPC framework for custom key wrapping or custom auth. Provides: server lifecycle, TLS, health checks.

| Service | Proto | Purpose |
|---------|-------|---------|
| KeyProvider | ext.KeyProviderService | Custom key wrapping (HSM etc.) |
| Authenticator | ext.AuthenticatorService | Custom authentication logic |

## Technical Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.24+ |
| RPC | gRPC | 1.83+ |
| Config | YAML (goccy/go-yaml) | 1.19+ |
| Observability | OpenTelemetry OTLP | 1.44+ |
| Logging | zerolog | 1.35+ |
| DI | uber-go/fx | 1.24+ |
| Auth | jwkset + golang-jwt/jwt/v5 | |
| Cloud KMS | gocloud.dev | 0.46+ |

## Folder Structure

```
proxy/
├── cmd/
│   └── proxy/
│       └── main.go          # binary entry point
├── internal/
│   ├── api/                 # gRPC handler wrappers
│   ├── auth/                # JWT, static token, extension
│   ├── config/              # Config loading & validation
│   ├── crypto/              # AES-GCM envelope encryption
│   ├── kms/                 # AWS, Azure, GCP, extension KMS
│   ├── proxy/               # Core proxy logic
│   ├── router/              # Routing engine
│   ├── rpc/                 # gRPC utilities
│   ├── server/              # Lifecycle (start/stop)
│   ├── metrics/             # OTel metric definitions
│   └── transport/           # Network transport
├── pkg/
│   ├── ext/                 # Extension server framework
│   ├── match/               # Namespace match helpers
│   ├── logger/              # zerolog helpers
│   └── validation/          # Config validation
├── api/
│   └── tempiex/proxy/v1/   # Proxy-specific proto definitions
├── e2e/                     # End-to-end tests
├── examples/
│   ├── cloud/               # Tempiex Cloud example config
│   └── kms/                 # Custom KMS extension example
├── go.mod
└── go.sum
```

## Configuration

### Minimal

```yaml
gateway:
  hostPort: ":8133"

upstreams:
  - name: local
    hostPort: "localhost:8134"   # tempiex server

routing:
  default: local
```

### Production (multi-environment)

```yaml
gateway:
  hostPort: ":8133"
  tls:
    enabled: true
    certFile: /etc/proxy/tls.crt
    keyFile: /etc/proxy/tls.key
  auth:
    enabled: true
    type: jwt
    jwt:
      jwksUrl: "https://auth.example.com/.well-known/jwks.json"
      audience: "tempiex-proxy"

upstreams:
  - name: cloud
    hostPort: "my-account.tmprl.cloud:8133"
    tls:
      enabled: true
      serverName: my-account.tmprl.cloud
    apiKey:
      header: "Authorization"
      value: "${TEMPIEX_CLOUD_API_KEY}"
    namespaceTranslation:
      prefix: "prod-"
    payloadEncryption:
      enabled: true
      provider: aws
      kmsKeyId: "arn:aws:kms:us-east-1:123456789012:key/..."

  - name: local
    hostPort: "localhost:8134"

routing:
  systemUpstream: local
  default: local
  rules:
    - namespace: "prod-*"
      upstream: cloud
    - namespace: "dev-*"
      upstream: local

metrics:
  otel:
    endpoint: "localhost:4317"
    insecure: true
```

## Key Features

### Namespace Translation
- Prefix: `ns` → `prod-ns`
- Suffix: `ns` → `ns-prod`
- Exact map: `ns` → `completely-different-name`
- Applied symmetrically on both request and response.

### Payload Encryption
- Codec-transparent — the proxy intercepts and encrypts payloads without understanding Tempiex protocol.
- Per-namespace key override supported.
- Automatic key rotation via KMS.

### Rule-Based Routing
- Namespace match: exact string, glob prefix, or regex.
- Metadata match: arbitrary gRPC metadata keys.
- System upstream for namespace-less RPCs (GetSystemInfo, etc.).
- Default fallback when no rule matches.

### Inbound Authentication
| Type | Implementation |
|------|---------------|
| Static token | Bearer token in Authorization header |
| JWT | JWKS-based validation (RS256 / ES256) |
| Extension | Custom gRPC extension server |
| None | Dev mode (no auth) |

## Build & Test

```bash
cd proxy/

go build -o bin/tempiex-proxy ./cmd/proxy

go test ./...

# Integration (requires running tempiex)
go test -tags integration ./e2e/...

golangci-lint run ./...
```

## Docker

```bash
docker pull tempiex/tempiex-proxy:latest

docker run -p 8233:8233 \
  -v $(pwd)/proxy.yaml:/config.yaml \
  tempiex/tempiex-proxy:latest \
  --config /config.yaml
```

## Extension Server Example

```go
import "github.com/tempiex/proxy/pkg/ext"

type HSMKeyProvider struct{}

func (k *HSMKeyProvider) WrapKey(ctx context.Context, req *ext.WrapKeyRequest) (*ext.WrapKeyResponse, error) {
    wrapped := myHSM.Wrap(req.Plaintext)
    return &ext.WrapKeyResponse{Ciphertext: wrapped}, nil
}

func main() {
    ext.RunKeyProviderServer(":9000", &HSMKeyProvider{})
}
```

## Observability

- Metrics via OTel OTLP gRPC (no Prometheus scrape endpoint).
- Traces: inbound gRPC → upstream gRPC spans with trace context propagation.
- Logs: zerolog JSON to stdout.

Key metrics:
- `proxy.requests.total` — by upstream, namespace, method, status
- `proxy.encryption.duration` — time for DEK wrap/unwrap
- `proxy.upstream.latency` — latency per upstream

### Key Metrics (OTel)

| Metric | Type | Description |
|--------|------|-------------|
| `proxy.requests.total` | Counter | By upstream, namespace, method, status |
| `proxy.request.duration` | Histogram | Request latency by upstream |
| `proxy.encryption.duration` | Histogram | DEK wrap/unwrap time |
| `proxy.upstream.errors.total` | Counter | Upstream error count by upstream |

## Security

### Threat Model

| Threat | Mitigation |
|--------|-----------|
| Unauthorized gateway access | Inbound auth (JWT / static token / extension) |
| Man-in-the-middle | TLS / mTLS on inbound and outbound connections |
| Credential leakage | Environment variable substitution; no hardcoded secrets |
| Payload exposure | Envelope encryption with KMS-backed DEK |

### Best Practices
- Use `${ENV_VAR}` substitution in config; never hardcode API keys or TLS keys.
- Enable inbound TLS for any non-localhost deployment.
- Rotate KMS KEKs on a schedule; the proxy handles DEK re-wrap transparently.
- Restrict the extension server port to the proxy process only.

## Performance Characteristics

| Scenario | Latency overhead | Throughput |
|----------|-----------------|------------|
| Plain routing (no encryption) | ~2–5ms | 10,000+ req/s |
| With payload encryption | ~5–10ms | 5,000+ req/s |
| KMS key wrap (first call) | ~50–100ms | cached after first wrap |

### Resource Requirements

| Deployment | CPU | Memory |
|------------|-----|--------|
| Development | 0.5 cores | 256 MB |
| Production (low traffic) | 2 cores | 1 GB |
| Production (high traffic) | 4+ cores | 4+ GB |

## Dependencies (key)

```
google.golang.org/grpc v1.83+
go.tempiex.com/api v1.63+
github.com/goccy/go-yaml v1.19+
github.com/rs/zerolog v1.35+
go.opentelemetry.io/otel v1.44+
go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc
github.com/MicahParks/keyfunc/v3
github.com/golang-jwt/jwt/v5
gocloud.dev v0.46+
go.uber.org/fx v1.24+
github.com/hashicorp/golang-lru/v2
golang.org/x/sync
```
