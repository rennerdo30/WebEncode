# WebEncode

![WebEncode Architecture](https://via.placeholder.com/800x400?text=WebEncode+Architecture)

**Distributed, Plug-in Based Video Transcoding Platform.**

WebEncode is high-performance, distributed media processing engine designed for scale. It features a micro-kernel architecture where all major functionality (Auth, Storage, Encoding, Live Streaming, Publishing) is offloaded to a resilient, gRPC-based plugin mesh.

## 🚀 Key Features

*   **Hyper-Modular Architecture**: 5-Pillar Plugin System (Auth, Storage, Encoder, Live, Publisher).
*   **Distributed Workers**: Scalable FFmpeg execution engine powered by NATS JetStream.
*   **Modern UI**: Next.js 16 Dashboard with Real-time monitoring and glassmorphism design.
*   **Live Streaming**: Zero-config RTMP Ingest to HLS via MediaMTX integration.
*   **Production Ready**: Docker, Kubernetes, OpenTelemetry, and structured logging.
*   **Restreaming**: 1:N simulcasting to YouTube, Twitch, Kick, and Rumble with unified chat.

## 🛠️ Quick Start

### Prerequisites
*   [Docker](https://docs.docker.com/get-docker/) & Docker Compose
*   [Go 1.24+](https://go.dev/dl/) (for local development)
*   [Node.js 22+](https://nodejs.org/) (for UI development)

### Run Everything (Docker)
The easiest way to start WebEncode is using the included Make commands.

```bash
# Start the entire stack (Kernel, Workers, UI, NATS, Postgres, MediaMTX)
make up
```

*   **Dashboard**: [http://localhost:3000](http://localhost:3000)
*   **API**: [http://localhost:8080](http://localhost:8080)
*   **MediaMTX**: [http://localhost:8888](http://localhost:8888)

### Build from Source
If you want to build the binaries locally:

```bash
make build-all
```

## 📂 Project Structure

| Directory | Description |
|-----------|-------------|
| `cmd/` | Entry points for the Kernel and Worker services. |
| `pkg/` | Shared libraries (API, Bus, DB, Logger). |
| `internal/` | Core business logic (Orchestrator, Plugin Manager). |
| `plugins/` | Source code for all 5 pillars of plugins. |
| `proto/` | gRPC Service Contracts (Protobuf definitions). |
| `ui/` | Next.js Frontend Dashboard. |
| `streamhub/` | Standalone Frontend for Stream Viewing. |
| `deploy/` | Docker & Kubernetes manifests. |
| `docs/` | Detailed documentation and specifications. |

## 📚 Documentation

*   [**Contributing Guide**](CONTRIBUTING.md): How to build plugins and contribute code.
*   [**API Reference**](docs/API_REFERENCE.md): Details on the REST API and Open API Spec.
*   [**Operator Runbook**](docs/OPERATOR.md): Guide for deploying and managing WebEncode in production.
*   [**Plugin SDK**](docs/PLUGIN_SDK.md): Guide for writing new plugins in Go.
*   [**Issues & Roadmap**](ISSUES.md): Current implementation status and known issues.
*   [**Audit Log**](AUDIT.md): Gap analysis against the original specification.

## 📝 License

MIT