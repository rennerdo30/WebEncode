# V1 API Reference

WebEncode provides a comprehensive REST API for managing jobs, streams, and system configuration.

**Base URL**: `http://localhost:8080/v1`

## 📖 OpenAPI Specification
The full API specification is available in OpenAPI 3.1 format:
[docs/openapi.yaml](../docs/openapi.yaml)

You can import this file into tools like Postman or Swagger UI to explore the API interactively.

## 🔑 Key Endpoints

### Jobs
*   `GET /jobs` - List recent jobs.
*   `POST /jobs` - Submit a new transcoding job.
*   `GET /jobs/{id}` - Get job details and status.

#### Example: Submit Job
```json
POST /v1/jobs
{
  "source_url": "https://example.com/video.mp4",
  "profiles": ["1080p_h264", "720p_h264"]
}
```

### Live Streaming (Webhooks)
These endpoints are primarily used by the MediaMTX sidecar but can be manually invoked for testing.

*   `POST /webhooks/live/auth` - Authenticate a publisher.
*   `POST /webhooks/live/start` - Signal stream start.
*   `POST /webhooks/live/stop` - Signal stream end.

### System
*   `GET /health` - System health check.
*   `GET /workers` - List connected worker nodes.
*   `GET /plugins` - List active plugins.

## 📡 Real-time Updates
The API supports Server-Sent Events (SSE) for real-time updates. Check the `docs/openapi.yaml` for specific subscription endpoints.
