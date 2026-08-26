# 02 - Temporal Proxy Specification

## Overview

**Component**: Temporal Proxy  
**Repository**: `temporal-proxy/`  
**Language**: Go 1.26.4  
**Type**: gRPC Proxy Server  
**Implementation Priority**: 2 (Depends on Temporal Server)

## Purpose

Temporal Proxy is a gRPC proxy that sits between Temporal SDK Clients, Workers, and the Temporal UI on one side, and one or more upstream Temporal Services on the other. It abstracts away connection details, enabling applications to target a single local endpoint while the proxy handles:

- **Namespace translation** (local → upstream names)
- **TLS termination** (inbound and outbound)
- **Payload encryption** (optional envelope encryption)
- **Multi-upstream routing** (based on namespace/metadata)
- **Credential management** (API keys, mTLS certificates)

This eliminates the need for applications to embed environment-specific configuration and enables seamless migration between development, staging, and production environments.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Client Applications                        │
│  ┌─────────┐  ┌──────────┐  ┌─────────┐  ┌─────────────┐      │
│  │ Workers │  │ Clients  │  │   UI    │  │  Temporal   │      │
│  │         │  │   (SDK)  │  │ Server  │  │     CLI     │      │
│  └────┬────┘  └─────┬────┘  └────┬────┘  └──────┬──────┘      │
│       │            │             │               │              │
│       └────────────┴─────────────┴───────────────┘              │
│                            │                                     │
│                   plaintext/TLS (optional)                      │
│                            │                                     │
└────────────────────────────┼─────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Temporal Proxy                              │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                   Gateway (Port 7233)                     │  │
│  │  • Namespace-based routing (codec-transparent)           │  │
│  │  • Optional inbound auth (JWT, static token, extension) │  │
│  │  • TLS termination                                       │  │
│  └───────────────────────┬──────────────────────────────────┘  │
│                          │                                      │
│     ┌────────────────────┼────────────────────┐                │
│     │                    │                    │                 │
│  ┌──▼──────────┐  ┌──────▼──────┐  ┌─────────▼────┐           │
│  │  Upstream   │  │  Upstream   │  │   Upstream   │           │
│  │  Proxy A    │  │  Proxy B    │  │   Proxy C    │           │
│  │             │  │             │  │   (System)   │           │
│  │ • Namespace │  │ • Namespace │  │              │           │
│  │   translate │  │   translate │  │              │           │
│  │ • Payload   │  │ • TLS       │  │              │           │
│  │   encrypt   │  │ • mTLS cert │  │              │           │
│  │ • API key   │  │             │  │              │           │
│  └──────┬──────┘  └──────┬──────┘  └──────┬───────┘           │
│         │                │                │                     │
└─────────┼────────────────┼────────────────┼─────────────────────┘
          │                │                │
          │ TLS + Auth     │ mTLS           │ plaintext
          │                │                │
          ▼                ▼                ▼
┌───────────────┐  ┌──────────────┐  ┌────────────────┐
│ Temporal      │  │ Self-hosted  │  │ Local Dev      │
│ Cloud         │  │ Temporal     │  │ Temporal       │
└───────────────┘  └──────────────┘  └────────────────┘

Optional Extension Server (for custom KMS or auth):
┌─────────────────────────────────────┐
│      Extension Server (gRPC)        │
│  • Custom key wrapping              │
│  • Custom authentication            │
└─────────────────────────────────────┘
          ▲
          │ gRPC
          │
    (called by proxy when needed)
```

### Component Breakdown

#### 1. Gateway
- **Purpose**: Single inbound gRPC endpoint for all clients
- **Port**: Configurable (default 7233)
- **Responsibilities**:
  - Accept gRPC requests from SDKs, workers, UI
  - Peek namespace from request metadata
  - Route to appropriate upstream proxy
  - Optional inbound authentication (JWT/static token/extension)
  - TLS termination (optional)
  - **Codec-transparent**: Never parses payloads

#### 2. Upstream Proxies
- **Purpose**: One instance per configured upstream Temporal Service
- **Responsibilities**:
  - Namespace translation (prefix, suffix, exact mapping)
  - Payload encryption/decryption (optional)
  - Attach upstream credentials (API key, mTLS cert)
  - Outbound TLS configuration
  - Connection pooling to upstream

#### 3. Router
- **Purpose**: Rule-based routing logic
- **Routing Criteria**:
  - Namespace name (primary)
  - Request metadata (secondary)
  - System upstream for namespace-less requests
  - Default fallback upstream

#### 4. Crypto Module (Optional)
- **Purpose**: Envelope encryption for payloads
- **Features**:
  - DEK (Data Encryption Key) generation
  - KEK (Key Encryption Key) wrapping via KMS
  - Automatic key rotation
  - Per-namespace key override
  - **Supported KMS**:
    - AWS KMS
    - Azure Key Vault
    - GCP KMS
    - Custom extension server

#### 5. Extension Server Interface
- **Purpose**: Pluggable backend for custom implementations
- **Use Cases**:
  - Custom key management (HSM, internal key service)
  - Custom authentication logic
- **Protocol**: gRPC (defined in `pkg/ext`)

## Technical Stack

### Core Technologies

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.26.4 |
| RPC Framework | gRPC | 1.83.0 |
| API Definition | Protobuf | 1.36.11 |
| Config Format | YAML | goccy/go-yaml 1.19.2 |
| Metrics | Prometheus | 1.24.1 |
| Logging | zerolog | 1.35.1 |
| Dependency Injection | uber-go/fx | 1.24.0 |

### Key Dependencies

```go
// gRPC & API
google.golang.org/grpc v1.83.0
go.temporal.io/api v1.63.4

// Configuration
github.com/goccy/go-yaml v1.19.2
github.com/urfave/cli/v3 v3.10.1

// Authentication
github.com/MicahParks/jwkset v0.11.3
github.com/MicahParks/keyfunc/v3 v3.8.1
github.com/golang-jwt/jwt/v5 v5.3.1

// Metrics & Observability
github.com/prometheus/client_golang v1.24.1
github.com/rs/zerolog v1.35.1

// Cloud KMS Integration
gocloud.dev v0.46.0
// - AWS KMS: github.com/aws/aws-sdk-go-v2
// - Azure KV: github.com/Azure/azure-sdk-for-go
// - GCP KMS: cloud.google.com/go/kms

// Dependency Injection
go.uber.org/fx v1.24.0

// Utilities
github.com/hashicorp/golang-lru/v2 v2.0.7
golang.org/x/sync v0.22.0
```

## Directory Structure

```
temporal-proxy/
├── api/                      # Proto definitions
│   └── temporal/proxy/v1/    # Proxy-specific protos
├── buf.gen.yaml              # Buf code generation config
├── buf.yaml                  # Buf configuration
├── cmd/
│   └── proxy/                # Main proxy binary
│       └── main.go
├── dev/                      # Development utilities
├── e2e/                      # End-to-end tests
├── examples/
│   ├── cloud/                # Temporal Cloud example
│   └── kms/                  # KMS extension example
├── internal/
│   ├── api/                  # API handlers
│   ├── auth/                 # Authentication logic
│   ├── config/               # Configuration loading & validation
│   ├── dataplane/            # Data plane (payload handling)
│   ├── kms/                  # KMS integrations
│   ├── metrics/              # Prometheus metrics
│   ├── protoutil/            # Proto utilities
│   ├── proxy/                # Core proxy logic
│   ├── router/               # Routing engine
│   ├── rpc/                  # gRPC utilities
│   ├── server/               # Server lifecycle
│   ├── services/             # Service implementations
│   ├── template/             # Template processing
│   └── transport/            # Network transport
├── pkg/
│   ├── api/                  # Public API types
│   ├── crypto/               # Encryption utilities
│   ├── ext/                  # Extension server framework
│   ├── logger/               # Logging utilities
│   ├── match/                # Routing match logic
│   ├── testutil/             # Testing utilities
│   └── validation/           # Configuration validation
├── rfc/                      # Design documents
│   ├── 01-overview.md
│   ├── encryption.md
│   └── routing.md
└── temporal-proxy.png        # Architecture diagram
```

## Key Features

### 1. Namespace Translation

**Problem**: Applications need to use different namespace names across environments.

**Solution**: Proxy rewrites namespace names in both requests and responses.

**Translation Modes**:
- **Prefix**: `local-ns` → `prod-local-ns`
- **Suffix**: `local-ns` → `local-ns-prod`
- **Exact mapping**: `local-ns` → `completely-different-name`

**Configuration Example**:
```yaml
upstreams:
  - name: prod-cloud
    hostPort: my-namespace.tmprl.cloud:7233
    namespaceTranslation:
      prefix: "prod-"
```

### 2. Payload Encryption

**Problem**: Upstream Temporal Service stores workflow payloads in plaintext.

**Solution**: Envelope encryption with KMS-backed key wrapping.

**Flow**:
1. Generate DEK (Data Encryption Key) per encryption operation
2. Encrypt payload with DEK (AES-256-GCM)
3. Wrap DEK with KEK (Key Encryption Key) from KMS
4. Store wrapped DEK + ciphertext
5. On read: unwrap DEK, decrypt payload

**Supported KMS Backends**:
- AWS KMS
- Azure Key Vault
- GCP Cloud KMS
- Custom extension server

**Configuration Example**:
```yaml
upstreams:
  - name: cloud-encrypted
    hostPort: my-namespace.tmprl.cloud:7233
    payloadEncryption:
      enabled: true
      kmsKeyId: "arn:aws:kms:us-east-1:123456789012:key/..."
      provider: aws
```

### 3. Rule-Based Routing

**Problem**: Route different namespaces to different Temporal deployments.

**Solution**: Flexible routing rules based on namespace and metadata.

**Routing Types**:
- **Namespace match**: Route by namespace name (exact, prefix, regex)
- **Metadata match**: Route by gRPC metadata
- **System upstream**: For namespace-less requests (GetSystemInfo, etc.)
- **Default upstream**: Fallback for unmatched requests

**Configuration Example**:
```yaml
routing:
  systemUpstream: local-dev
  default: self-hosted
  rules:
    - namespace: "prod-*"
      upstream: cloud
    - namespace: "staging-*"
      upstream: self-hosted
    - namespace: "dev-*"
      upstream: local-dev
```

### 4. Inbound Authentication

**Problem**: Secure the proxy gateway from unauthorized access.

**Solutions**:
- **Static token**: Simple bearer token in Authorization header
- **JWT validation**: JWKS-based JWT validation
- **Extension server**: Custom authentication logic

**Configuration Example**:
```yaml
gateway:
  auth:
    enabled: true
    type: jwt
    jwt:
      jwksUrl: "https://auth.example.com/.well-known/jwks.json"
      audience: "temporal-proxy"
```

### 5. TLS Termination

**Inbound TLS** (Gateway):
```yaml
gateway:
  tls:
    enabled: true
    certFile: /path/to/cert.pem
    keyFile: /path/to/key.pem
    clientAuth: require # optional mTLS
```

**Outbound TLS** (Upstream):
```yaml
upstreams:
  - name: cloud
    tls:
      enabled: true
      serverName: my-namespace.tmprl.cloud
      rootCAs: /path/to/ca.pem
```

### 6. Extension Server Support

**Purpose**: Implement custom backends for KMS or authentication.

**Provided by `pkg/ext`**:
- gRPC server framework
- TLS support
- Graceful shutdown
- Health checks

**Example Extension Server**:
```go
import "github.com/temporalio/temporal-proxy/pkg/ext"

type MyKeyProvider struct{}

func (k *MyKeyProvider) WrapKey(ctx context.Context, req *ext.WrapKeyRequest) (*ext.WrapKeyResponse, error) {
    // Call your HSM or key service
    wrappedKey := myHSM.Wrap(req.Plaintext)
    return &ext.WrapKeyResponse{Ciphertext: wrappedKey}, nil
}

func main() {
    ext.RunKeyProviderServer(":9000", &MyKeyProvider{})
}
```

## Configuration

### Minimal Configuration

```yaml
# Gateway configuration
gateway:
  hostPort: :7233

# Upstream Temporal Services
upstreams:
  - name: local
    hostPort: localhost:7234

# Routing
routing:
  default: local
```

### Production Configuration

```yaml
gateway:
  hostPort: :7233
  tls:
    enabled: true
    certFile: /certs/server.crt
    keyFile: /certs/server.key
  auth:
    enabled: true
    type: jwt
    jwt:
      jwksUrl: "https://auth.company.com/.well-known/jwks.json"

upstreams:
  - name: cloud-prod
    hostPort: prod.a2dd6.tmprl.cloud:7233
    tls:
      enabled: true
      serverName: prod.a2dd6.tmprl.cloud
    credentials:
      apiKey: "${TEMPORAL_CLOUD_API_KEY}"
    namespaceTranslation:
      prefix: "prod-"
    payloadEncryption:
      enabled: true
      kmsKeyId: "arn:aws:kms:us-east-1:123:key/abc123"
      provider: aws

  - name: self-hosted
    hostPort: temporal.internal:7233
    tls:
      enabled: true
      clientCert: /certs/client.crt
      clientKey: /certs/client.key
      rootCAs: /certs/ca.crt

  - name: local-dev
    hostPort: localhost:7234

routing:
  systemUpstream: local-dev
  default: self-hosted
  rules:
    - namespace: "prod-*"
      upstream: cloud-prod
    - namespace: "dev-*"
      upstream: local-dev
```

## Build & Deployment

### Build Commands

```bash
# Build binary
go build -o proxy ./cmd/proxy

# Using mise
mise run build

# Run tests
mise run test

# Lint
mise run lint

# Format code
mise run format
```

### Docker Image

```bash
# Pull from Docker Hub
docker pull temporalio/temporal-proxy:latest

# Run container
docker run -p 7233:7233 \
  -v $(pwd)/config.yaml:/config.yaml \
  temporalio/temporal-proxy:latest \
  --config /config.yaml
```

### Helm Deployment

```bash
# Install with Helm
helm repo add temporal https://go.temporal.io/helm-charts
helm install temporal-proxy temporal/temporal-proxy \
  -f values.yaml
```

**values.yaml**:
```yaml
config:
  gateway:
    hostPort: :7233
  upstreams:
    - name: cloud
      hostPort: my-namespace.tmprl.cloud:7233
  routing:
    default: cloud
```

## API Endpoints

### gRPC Services (Gateway)

| Service | Port | Proto |
|---------|------|-------|
| Proxied Temporal API | 7233 | temporal.api.workflowservice.v1 |
| Health | 7233 | grpc.health.v1 |

### Metrics Endpoint

| Endpoint | Port | Purpose |
|----------|------|---------|
| /metrics | 9090 | Prometheus metrics |

### Extension Server API (pkg/ext)

| Service | Proto | Purpose |
|---------|-------|---------|
| KeyProvider | ext.KeyProviderService | Custom key wrapping |
| Authenticator | ext.AuthenticatorService | Custom authentication |

## Dependencies & Integration Points

### Upstream Dependencies
- **Temporal Server** (required) - Proxied destination
- **KMS Providers** (optional) - AWS KMS, Azure KV, GCP KMS
- **Auth Provider** (optional) - JWKS endpoint
- **Extension Server** (optional) - Custom KMS/auth

### Downstream Dependents
- **SDK Clients** - Connect to proxy instead of Temporal Server
- **Workers** - Poll tasks via proxy
- **UI Server** - Can use proxy for multi-cluster access

## Security

### Threat Model
1. **Unauthorized access to gateway** → Inbound auth (JWT/token/extension)
2. **Man-in-the-middle attacks** → TLS/mTLS
3. **Credential leakage** → Environment variable injection, no hardcoded secrets
4. **Payload exposure** → Envelope encryption with KMS

### Security Best Practices
- Never hardcode credentials in config files
- Use environment variable substitution: `${ENV_VAR}`
- Enable TLS for production deployments
- Use mTLS for inbound auth when possible
- Rotate KMS keys regularly
- Audit logging for authentication failures

## Performance Characteristics

### Latency
- **Overhead**: ~2-5ms added latency for routing
- **Encryption overhead**: ~5-10ms for payload encryption/decryption
- **KMS call overhead**: ~50-100ms for key wrapping (cached after first call)

### Throughput
- **Request throughput**: 10,000+ req/s (without encryption)
- **With encryption**: 5,000+ req/s

### Resource Requirements

| Deployment | CPU | Memory |
|------------|-----|--------|
| Development | 0.5 cores | 256 MB |
| Production (low traffic) | 2 cores | 1 GB |
| Production (high traffic) | 4+ cores | 4+ GB |

## Testing

### Test Types
- **Unit tests**: `internal/` and `pkg/` packages
- **Integration tests**: `e2e/` directory
- **Example tests**: `examples/` directory

### Test Commands

```bash
# Run all tests
mise run test

# Run with coverage
go test -cover ./...

# Run e2e tests
make test-e2e
```

## Monitoring & Operations

### Key Metrics

| Metric | Type | Description |
|--------|------|-------------|
| proxy_requests_total | Counter | Total requests by upstream |
| proxy_request_duration_seconds | Histogram | Request latency by upstream |
| proxy_payload_encryption_duration_seconds | Histogram | Encryption operation time |
| proxy_kms_operations_total | Counter | KMS operations by type |
| proxy_auth_failures_total | Counter | Authentication failures |
| proxy_routing_errors_total | Counter | Routing failures |

### Health Checks
- **Liveness**: Process running
- **Readiness**: Gateway listening, upstreams reachable

## Implementation Notes

### Codec-Transparent Design
- Gateway never parses payloads, only peeks namespace from metadata
- Encryption happens in upstream proxy layer, not gateway
- Enables zero-copy proxying for non-encrypted paths

### fx Dependency Injection
Uses `uber-go/fx` for wiring:

```go
fx.New(
    config.Module,
    server.Module,
    router.Module,
    metrics.Module,
    fx.Invoke(startGateway),
)
```

### Error Handling
- Wraps errors with context using `fmt.Errorf("...: %w", err)`
- Validates config early at startup
- Returns gRPC status codes to clients

## References

- [RFC: Overview](./rfc/01-overview.md)
- [RFC: Encryption](./rfc/encryption.md)
- [RFC: Routing](./rfc/routing.md)
- [Official Docs](https://docs.temporal.io/production-deployment/temporal-proxy/)
