.PHONY: up down build-all proto tidy docker-build test test-coverage lint

# Docker Compose
up:
	docker compose -f deploy/compose/docker-compose.yml up -d --build

up-no-build:
	docker compose -f deploy/compose/docker-compose.yml up -d

down:
	docker compose -f deploy/compose/docker-compose.yml down

down-clean:
	docker compose -f deploy/compose/docker-compose.yml down -v

docker-build:
	docker compose -f deploy/compose/docker-compose.yml build

# Reset Database (Wipe data and restart)
reset-db:
	docker compose -f deploy/compose/docker-compose.yml down -v
	docker compose -f deploy/compose/docker-compose.yml up -d postgres kernel

# Build
build-all:
	@echo "Building kernel and worker..."
	go build -o build/kernel ./cmd/kernel
	go build -o build/worker ./cmd/worker
	@echo "Building plugins..."
	go build -o build/plugins/storage-s3.bin ./plugins/storage-s3
	go build -o build/plugins/storage-fs.bin ./plugins/storage-fs
	go build -o build/plugins/encoder-ffmpeg.bin ./plugins/encoder-ffmpeg
	go build -o build/plugins/live-mediamtx.bin ./plugins/live-mediamtx
	go build -o build/plugins/publisher-dummy.bin ./plugins/publisher-dummy
	go build -o build/plugins/publisher-youtube.bin ./plugins/publisher-youtube
	go build -o build/plugins/publisher-twitch.bin ./plugins/publisher-twitch
	go build -o build/plugins/publisher-kick.bin ./plugins/publisher-kick
	go build -o build/plugins/publisher-rumble.bin ./plugins/publisher-rumble
	go build -o build/plugins/publisher-rtmp.bin ./plugins/publisher-rtmp
	go build -o build/plugins/auth-oidc.bin ./plugins/auth-oidc
	go build -o build/plugins/auth-basic.bin ./plugins/auth-basic
	go build -o build/plugins/auth-ldap.bin ./plugins/auth-ldap
	go build -o build/plugins/auth-cloudflare-access.bin ./plugins/auth-cloudflare-access
	@echo "Build complete!"

# Testing
test:
	go test ./... -v

test-short:
	go test ./... -short

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Linting
lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# Proto Generation
proto:
	./scripts/generate_proto.sh

# SQLC Generation
gen-db:
	docker run --rm -v "$(PWD)":/src -w /src sqlc/sqlc generate

# Database Migrations
DB_URL ?= postgres://webencode:webencode@localhost:5432/webencode?sslmode=disable

migrate-up:
	@export PATH="$(PATH):$(shell go env GOPATH)/bin" && \
	migrate -path pkg/db/migrations -database "$(DB_URL)" up

migrate-down:
	@export PATH="$(PATH):$(shell go env GOPATH)/bin" && \
	migrate -path pkg/db/migrations -database "$(DB_URL)" down 1

migrate-create:
	@export PATH="$(PATH):$(shell go env GOPATH)/bin" && \
	migrate create -ext sql -dir pkg/db/migrations -seq $(name)

# Go Mod
tidy:
	go mod tidy

# Development helpers
dev: down-clean up
	@echo "Development environment started!"
	@echo "UI: http://localhost:3000"
	@echo "API: http://localhost:8090/v1/system/health"

logs:
	docker compose -f deploy/compose/docker-compose.yml logs -f

logs-kernel:
	docker compose -f deploy/compose/docker-compose.yml logs -f kernel

logs-worker:
	docker compose -f deploy/compose/docker-compose.yml logs -f worker


