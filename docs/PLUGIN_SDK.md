# Plugin Developer Guide

> **Version**: 1.0 (WebEncode v2025)
> **SDK Package**: `github.com/rennerdo30/webencode/pkg/pluginsdk`

## Overview
WebEncode uses a **Hyper-Modular 5-Pillar Plugin Mesh** architecture. This means nearly all functionality (except the Core Kernel) is offloaded to plugins.

Plugins are **standalone binaries** that the Kernel starts and communicates with over **gRPC** (via HashiCorp's go-plugin system).

### The 5 Pillars
1. **Auth**: Identity, Session Validation, RBAC (e.g., LDAP, OIDC).
2. **Storage**: File I/O for VOD/Live assets (e.g., S3, Local FS).
3. **Encoder**: Transcoding logic (e.g., FFmpeg, Hardware).
4. **Live**: RTMP/WebRTC ingest and telemetry (e.g., MediaMTX).
5. **Publisher**: Distribution to external platforms (e.g., YouTube, Twitch).

---

## 🚀 Getting Started

To create a new plugin, create a Go `main` package.

### 1. Minimal Structure
```go
package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

func main() {
	// 1. Define your implementation
	impl := &MyStoragePlugin{}

	// 2. Serve the plugin
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"storage": &pluginsdk.Plugin{
				StorageImpl: impl, // Register the interface you implement
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// Ensure you implement the specific interface
type MyStoragePlugin struct {
	pluginsdk.UnimplementedStorageServiceServer // Embed for forward compatibility
}
```

### 2. Build Instructions
Build your plugin as a standalone binary:
```bash
go build -o ./bin/my-plugin ./cmd/my-plugin
```

### 3. Registration
Register your plugin in the WebEncode Kernel by pointing to the binary path in `plugins.toml` or via the API.

---

## 🔌 Interfaces Reference

### 1. Auth Plugin (`AuthService`)
Used for verifying tokens and managing users.

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `ValidateToken` | `TokenRequest` | `UserSession` | Validates a Bearer/API token. |
| `Authorize` | `AuthZRequest` | `AuthZResponse` | Checks RBAC permissions. |
| `GetUser` | `UserRequest` | `User` | Fetches user profile. |

### 2. Storage Plugin (`StorageService`)
Used for reading/writing media files.

| Method | Description |
|--------|-------------|
| `Upload` / `Download` | Streaming RPCs for file transfer. |
| `GetURL` / `GetUploadURL` | Generate signed URLs (if supported). |
| `BrowseRoots` / `Browse` | File browser capabilities. |

### 3. Encoder Plugin (`EncoderService`)
Used for probing and transcoding media.

| Method | Description |
|--------|-------------|
| `Probe` | Returns metadata (width, height, bitrate) for a file. |
| `Transcode` | Streaming RPC returns progress updates (0-100%). |
| `GetCapabilities` | Reports supported codecs (h264, hevc, etc.). |

### 4. Live Plugin (`LiveService`)
Used for managing RTMP/WebRTC sessions.

| Method | Description |
|--------|-------------|
| `StartIngest` | Provisions a stream key/URL. |
| `StopIngest` | Terminates a session. |
| `GetTelemetry` | Polled by Kernel for stats (Viewers, Bitrate). |

### 5. Publisher Plugin (`PublisherService`)
Used for uploading final assets to external sites.

| Method | Description |
|--------|-------------|
| `Publish` | Uploads file/stream to YouTube/Twitch. |
| `Retract` | Deletes/Unlists the content. |

---

## 🛠 Testing Your Plugin

The SDK includes a `HealthCheck` automatically. You do **not** need to implement `Check` or `Watch`.

### Unit Testing
You can import `github.com/rennerdo30/webencode/pkg/pluginsdk` in your tests to mock the gRPC server interactions.

### Integration Testing
1. Build your plugin binary.
2. Run the Kernel with your plugin registered.
3. Check Kernel logs for `[PLUGIN] <name> started`.
4. Use the WebEncode UI/API to trigger actions handled by your plugin.

## ⚠️ Common Pitfalls

1. **Stdout/Stderr**: Plugins communicate via stdout. **DO NOT** use `fmt.Println` in your plugin code. It will corrupt the handshake. Use `hclog` or the `pkg/logger` provided by the SDK which writes to stderr.
2. **Duplicate Registration**: Do not register `grpc_health_v1` manually; `go-plugin` does this for you.
3. **CGO**: Ensure your plugin is built with the same architecture (Linux/AMD64) if running inside Docker.
