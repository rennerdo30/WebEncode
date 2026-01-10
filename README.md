<div align="center">
  <img src="logo.svg" alt="WebEncode Logo" width="150" height="150">

  # WebEncode

  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
  [![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
  [![Next.js](https://img.shields.io/badge/Next.js-16-black?logo=next.js)](https://nextjs.org/)

  **Distributed, Plugin-Based Video Transcoding Platform**
</div>

WebEncode is a high-performance, distributed media processing engine designed for scale. It features a micro-kernel architecture where all major functionality (Auth, Storage, Encoding, Live Streaming, Publishing) is offloaded to a resilient, gRPC-based plugin mesh.

## Features

- **Hyper-Modular Architecture**: 5-Pillar Plugin System (Auth, Storage, Encoder, Live, Publisher)
- **Distributed Workers**: Scalable FFmpeg execution engine powered by NATS JetStream
- **Modern UI**: Next.js 16 Dashboard with real-time monitoring
- **Live Streaming**: Zero-config RTMP ingest to HLS via MediaMTX integration
- **Production Ready**: Docker, Kubernetes, OpenTelemetry, and structured logging
- **Restreaming**: 1:N simulcasting to YouTube, Twitch, Kick, and Rumble with unified chat

## Technology Stack

### Backend
- **Language**: Go 1.24+
- **Framework**: Chi/v5 (HTTP), gRPC (plugins)
- **Database**: PostgreSQL 16 (pgx/v5)
- **Message Bus**: NATS JetStream 2.10+
- **Plugin System**: HashiCorp go-plugin

### Frontend
- **Framework**: Next.js 16 with React 19
- **UI Components**: Shadcn/ui, Radix UI
- **Styling**: Tailwind CSS 4
- **State Management**: TanStack React Query

### Infrastructure
- **Containerization**: Docker, Docker Compose
- **Orchestration**: Kubernetes
- **Live Streaming**: MediaMTX (RTMP/WebRTC)
- **Storage**: S3-compatible (MinIO) + Local filesystem

## Quick Start

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- [Go 1.24+](https://go.dev/dl/) (for local development)
- [Node.js 22+](https://nodejs.org/) (for UI development)

### Run with Docker

```bash
# Start the entire stack (Kernel, Workers, UI, NATS, Postgres, MediaMTX)
make up
```

Once running:
- **Dashboard**: http://localhost:3000
- **API**: http://localhost:8080
- **MediaMTX**: http://localhost:8888

### Build from Source

```bash
# Build all binaries (kernel, worker, plugins)
make build-all

# Run tests
make test

# Generate test coverage
make test-coverage
```

## Project Structure

```
WebEncode/
├── cmd/                  # Entry points (kernel, worker)
├── internal/             # Core business logic
│   ├── api/              # REST API handlers
│   ├── orchestrator/     # Job orchestration
│   ├── plugin_manager/   # Plugin lifecycle
│   └── worker/           # Worker tasks
├── pkg/                  # Shared libraries
│   ├── api/              # API models
│   ├── bus/              # NATS messaging
│   ├── db/               # Database models (SQLC)
│   └── pluginsdk/        # Plugin SDK interfaces
├── plugins/              # 5-Pillar Plugin System
│   ├── auth-*/           # Authentication plugins
│   ├── storage-*/        # Storage plugins
│   ├── encoder-*/        # Encoder plugins
│   ├── live-*/           # Live streaming plugins
│   └── publisher-*/      # Publisher plugins
├── proto/                # gRPC Protobuf definitions
├── ui/                   # Next.js Frontend
├── deploy/               # Docker & Kubernetes configs
└── docs/                 # Documentation
```

## Plugin Architecture

WebEncode uses a 5-pillar plugin system:

| Pillar | Purpose | Included Plugins |
|--------|---------|------------------|
| **Auth** | User authentication | Basic, OIDC, LDAP, Cloudflare Access |
| **Storage** | File storage | Filesystem, S3/MinIO |
| **Encoder** | Transcoding | FFmpeg |
| **Live** | RTMP/WebRTC ingest | MediaMTX |
| **Publisher** | Distribution | YouTube, Twitch, Kick, Rumble, Generic RTMP |

## Documentation

- [**Specification**](SPECIFICATION.md): Complete system specification
- [**Contributing Guide**](CONTRIBUTING.md): Development setup and plugin development
- [**API Reference**](docs/API_REFERENCE.md): REST API documentation
- [**Plugin SDK**](docs/PLUGIN_SDK.md): Guide for writing plugins
- [**Operator Runbook**](docs/OPERATOR.md): Deployment and operations guide
- [**OpenAPI Spec**](docs/openapi.yaml): OpenAPI 3.1 specification
- [**Issues & Roadmap**](ISSUES.md): Implementation status and known issues

## Configuration

WebEncode is configured via environment variables. See the [Operator Runbook](docs/OPERATOR.md) for full configuration options.

Key environment variables:
```bash
DATABASE_URL=postgres://user:pass@localhost:5432/webencode
NATS_URL=nats://localhost:4222
AUTH_PLUGIN=auth-basic
STORAGE_PLUGIN=storage-fs
```

## Contributing

Contributions are welcome. Please read the [Contributing Guide](CONTRIBUTING.md) for development setup and guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.