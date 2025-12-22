#!/bin/bash

# Generate Go code from protobuf definitions
# Requires: protoc, protoc-gen-go, protoc-gen-go-grpc

set -e

cd "$(dirname "$0")/.."

# Install dependencies if missing
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

export PATH="$PATH:$(go env GOPATH)/bin"

# Remove old generated files in proto directory (wrong location)
rm -f proto/*.pb.go

# Generate Go code to the correct location
protoc \
  --proto_path=. \
  --go_out=. \
  --go_opt=module=github.com/rennerdo30/webencode \
  --go-grpc_out=. \
  --go-grpc_opt=module=github.com/rennerdo30/webencode \
  proto/*.proto

echo "Proto generation complete!"
