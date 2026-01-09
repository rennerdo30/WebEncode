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

---

## 📈 Performance Tuning

### Database Optimization
```sql
-- Create indexes for common queries
CREATE INDEX idx_jobs_user_created ON jobs(user_id, created_at DESC) WHERE status != 'completed';
CREATE INDEX idx_tasks_job_status ON tasks(job_id, status);
CREATE INDEX idx_streams_user_live ON streams(user_id) WHERE is_live = true;

-- Analyze tables for query planner
VACUUM ANALYZE jobs;
VACUUM ANALYZE tasks;
VACUUM ANALYZE streams;

-- Tune PostgreSQL for heavy workloads
ALTER SYSTEM SET work_mem = '256MB';
ALTER SYSTEM SET shared_buffers = '1GB';
ALTER SYSTEM SET effective_cache_size = '4GB';
SELECT pg_reload_conf();
```

### NATS JetStream Tuning
```bash
# nats-server.conf
max_connections = 100000
max_subscriptions = 100000

jetstream {
  max_mem_store = 2GB
  max_file_store = 50GB
}
```

### Go Runtime Tuning
```bash
# Environment variables for Go runtime
export GOGC=75                    # Reduce GC frequency
export GOMEMLIMIT=4GiB            # Set soft memory limit
export GOMAXPROCS=0               # Use all available cores
```

---

## 🔐 Security Hardening

### Network Security
- **TLS Everywhere**: Enable TLS between all services.
- **Firewall Rules**: Only allow necessary ports (8080 for API, 4222 for NATS).
- **Network Segmentation**: Isolate workers in separate VLAN.

### Authentication
- Use **auth-oidc** plugin for production (Keycloak, Auth0, Okta).
- Set `OIDC_ISSUER_URL`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`.
- Never use dev-mode tokens in production.

### Secrets Management
- Store credentials in environment variables or secrets manager.
- Never commit `.env` files with production secrets.
- Rotate access tokens for publisher plugins regularly.

---

## 🌐 CI/CD Pipeline

### GitHub Actions Example
```yaml
name: Build & Deploy
on:
  push:
    branches: [main]
    tags: [v*]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:17
        env:
          POSTGRES_PASSWORD: test
      nats:
        image: nats:latest
        options: --js
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.24'
    - run: go test ./... -v -coverprofile=coverage.out
    - uses: codecov/codecov-action@v4

  build:
    needs: test
    runs-on: ubuntu-latest
    strategy:
      matrix:
        target: [kernel, worker]
    steps:
    - uses: actions/checkout@v4
    - run: docker build -f docker/${{ matrix.target }}.Dockerfile -t webencode/${{ matrix.target }}:${{ github.ref_name }} .
    - run: docker push webencode/${{ matrix.target }}:${{ github.ref_name }}
```

---

## 🆘 Disaster Recovery

### Recovery Objectives
- **RTO (Recovery Time Objective)**: 5 minutes (restart all services)
- **RPO (Recovery Point Objective)**: 1 minute (NATS replication lag)

### Backup Strategy
| Tier | Method | Frequency | Retention |
|------|--------|-----------|-----------|
| 1 | PostgreSQL streaming replication | Continuous | Real-time |
| 2 | pg_dump to S3 | Hourly | 30 days |
| 3 | Full snapshot | Weekly | 90 days |

### Recovery Procedures

**Scenario: Database Failure**
1. Promote PostgreSQL replica to primary.
2. Update `DATABASE_URL` in all Kernel instances.
3. Restart Kernel services.
4. Verify connectivity: `curl /v1/system/health`.

**Scenario: NATS Cluster Failure**
1. If quorum lost, restore from latest snapshot.
2. Restart NATS cluster with clean data directory.
3. Workers will automatically reconnect and resubscribe.
4. In-flight tasks will be retried (idempotent design).

**Scenario: Complete Site Failure**
1. Activate standby Kubernetes cluster in DR region.
2. Restore PostgreSQL from S3 backup.
3. Update DNS to point to DR site.
4. Start all services with restored configuration.

---

## 📊 Monitoring & Alerting

### Prometheus Metrics
WebEncode exposes metrics at `/metrics`:
- `webencode_jobs_total{status}`: Total jobs by status.
- `webencode_tasks_processing`: Currently processing tasks.
- `webencode_workers_healthy`: Number of healthy workers.
- `webencode_streams_live`: Number of live streams.

### Recommended Alerts
| Alert | Condition | Severity |
|-------|-----------|----------|
| No Healthy Workers | `webencode_workers_healthy == 0` | Critical |
| Job Queue Backlog | `webencode_jobs_total{status="queued"} > 100` | Warning |
| High Error Rate | `rate(webencode_errors_total[5m]) > 0.1` | Warning |
| Database Connection Pool | `pg_stat_activity_count > 80%` | Warning |

### Grafana Dashboard
Import the provided dashboard JSON from `docs/grafana-dashboard.json` (if available) or create custom panels using the above metrics.

