# Operator Runbook

> **Version**: 1.0 (WebEncode v2025)
> **Role**: System Administrator / DevOps

## 📋 Recommended Architecture
WebEncode is designed for **High Availability (HA)**.

### Minimal Deployment
- **1x Kernel Node**: Managing state and API.
- **1x Worker Node**: Processing transcoding jobs (CPU/GPU).
- **1x NATS Server**: Message bus (JetStream enabled).
- **1x PostgreSQL**: Persistent metadata storage.

### Production HA Deployment
- **3x Kernel Nodes**: Load balanced (Stateless API).
- **N+1 Worker Nodes**: Auto-scaling based on queue depth.
- **3x NATS Cluster**: Resilient messaging.
- **Postgres Primary + Replica**: Data safety.

---

## ⚙️ Configuration
Configuration is handled via Environment Variables or `.env` file.

| Variable | Default | Description |
|----------|---------|-------------|
| `NATS_URL` | `nats://localhost:4222` | Connection string for NATS JetStream. |
| `DATABASE_URL` | `postgres://...` | Postgres connection (Kernel only). |
| `PLUGIN_DIR` | `./plugins` | Directory containing compiled plugin binaries. |
| `PORT` | `8080` | Kernel API bind port. |
| `WORKER_ID` | `hostname` | Unique identifier for worker nodes (Worker only). |

---

## 🚀 Deployment Guide

### 1. Database Setup
Run migrations before starting the Kernel.
```bash
# Using golang-migrate
migrate -path pkg/db/migrations -database $DATABASE_URL up
```

### 2. Kernel Startup
Ensure `plugins` directory is populated with binaries (`make build`).
```bash
./bin/kernel
```
*Health Check*: `curl http://localhost:8080/v1/health`

### 3. Worker Startup
Workers auto-register upon connection to NATS.
```bash
export WORKER_ID=worker-gpu-01
./bin/worker
```
*Verification*: Check Kernel logs or `GET /v1/workers`.

### 4. Plugin Management
Plugins are subprocesses managed by the Kernel/Worker.
- **Install**: Place binary in `PLUGIN_DIR` and update `plugins.toml` (if using manifest) or register via API.
- **Upgrade**: Replace binary and restart Kernel/Worker (Hot reload planned).

---

## 🔍 Troubleshooting

### Logs
All components emit structured JSON logs to `stderr`.
- **Level**: Info by default.
- **Fields**: `service`, `level`, `msg`, `error`.

### Common Issues

**1. "NATS Connection Failed"**
- Ensure NATS JetStream is enabled (`nats-server -js`).
- Check firewall rules between Kernel/Worker and NATS.

**2. "Plugin Mismatch"**
- Error: `Incompatible API version`
- Cause: Kernel and Plugin built with different SDK versions.
- Fix: Rebuild both with same `pkg/api` version.

**3. "Job Stuck in Pending"**
- Cause 1: No healthy workers connected.
- Cause 2: Workers lack capabilities (e.g., job requires `nvidia`, worker has `cpu`).
- Fix: Check `GET /v1/workers` for capabilities and status.

**4. "Database Migration Failed"**
- Cause: Dirty state from failed previous migration.
- Fix: `migrate force <version>` (Use with caution).

---

## 🔄 Maintenance

### Upgrading
1. **Stop Kernel** (Users cannot submit new jobs).
2. **Upgrade DB Schema** (Run migrations).
3. **Upgrade Plugins** (Replace binaries).
4. **Start Kernel**.
5. **Rolling Upgrade Workers** (Wait for idle or drain).

### Backup
- **Postgres**: Regular `pg_dump`.
- **NATS**: JetStream limits are configured to 90 days for audit logs. Backup if critical.
- **Storage**: External (S3/FS) - Managed separately.

