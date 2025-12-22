# Contributing to WebEncode

Thank you for your interest in contributing to WebEncode! This document provides guidelines for setting up your development environment and contributing to the project.

## 🛠 Development Environment

To develop on WebEncode, you will need:

*   **Go 1.24+**: For the backend Kernel, Workers, and Plugins.
*   **Node.js 22+**: For the Next.js Frontend.
*   **Docker & Docker Compose**: For running infrastructure (Postgres, NATS) and integration testing.
*   **Protoc**: Protocol Buffers compiler (optional, for modifying `.proto` files).

### Setup

1.  **Clone the repository**:
    ```bash
    git clone https://github.com/rennerdo30/webencode.git
    cd webencode
    ```

2.  **Install Go dependencies**:
    ```bash
    go mod download
    ```

3.  **Install UI dependencies**:
    ```bash
    cd ui
    npm install
    ```

## 🧪 Testing

We aim for high test coverage. Please ensure all tests pass before submitting a PR.

```bash
# Run all Go unit tests
go test ./...

# Run specific package tests
go test ./internal/orchestrator/...
```

## 🧩 Writing a New Plugin

WebEncode is built around a plugin architecture. Writing a new plugin is the most common way to extend functionality.

### 1. Create Plugin Directory
Create a directory in `plugins/`, e.g., `plugins/my-plugin`.

### 2. Implement Logic
Implement your logic in Go using `pkg/pluginsdk` for boilerplate.

### Example (Encoder)

```go
package main

import (
    "github.com/hashicorp/go-plugin"
    "github.com/rennerdo30/webencode/pkg/pluginsdk"
)

// Implement pluginsdk.EncoderServiceServer interface...
type MyEncoder struct {
    pluginsdk.UnimplementedEncoderServiceServer
}

func (e *MyEncoder) Transcode(req *pluginsdk.TranscodeRequest, stream pluginsdk.EncoderService_TranscodeServer) error {
    // Your transcoding logic here
    return nil
}

func main() {
    plugin.Serve(&plugin.ServeConfig{
        HandshakeConfig: pluginsdk.HandshakeConfig,
        Plugins: map[string]plugin.Plugin{
            "my-encoder": &pluginsdk.Plugin{
                EncoderImpl: &MyEncoder{},
            },
        },
        GRPCServer: plugin.DefaultGRPCServer,
    })
}
```

### 3. Register Plugin
*   Build the plugin: `go build -o ../../bin/my-encoder .`
*   Register it via the API or add it to the `plugins` table.

For more details, see [docs/PLUGIN_SDK.md](docs/PLUGIN_SDK.md).

## 📝 Code Standards

*   **Formatting**: run `go fmt ./...`
*   **Linting**: We follow standard Go idioms.
*   **Commits**: Use descriptive commit messages.
