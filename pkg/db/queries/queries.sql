-- ============================================
-- JOBS
-- ============================================

-- name: CreateJob :one
INSERT INTO jobs (
    source_url, profiles, status, metadata, 
    user_id, source_type, profile_id, output_config
)
VALUES ($1, $2, 'queued', $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs
WHERE id = $1 LIMIT 1;

-- name: ListJobs :many
SELECT * FROM jobs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListJobsByUser :many
SELECT * FROM jobs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListJobsByStatus :many
SELECT * FROM jobs
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateJobStatus :exec
UPDATE jobs
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateJobProgress :exec
UPDATE jobs
SET progress_pct = $2, eta_seconds = $3, updated_at = NOW()
WHERE id = $1;

-- name: UpdateJobStarted :exec
UPDATE jobs
SET status = 'processing', started_at = NOW(), assigned_to_worker_id = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateJobCompleted :exec
UPDATE jobs
SET status = 'completed', finished_at = NOW(), progress_pct = 100, updated_at = NOW()
WHERE id = $1;

-- name: UpdateJobFailed :exec
UPDATE jobs
SET status = 'failed', error_message = $2, finished_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: CancelJob :exec
UPDATE jobs
SET status = 'cancelled', finished_at = NOW(), updated_at = NOW()
WHERE id = $1 AND status NOT IN ('completed', 'failed', 'cancelled');

-- name: DeleteJob :exec
DELETE FROM jobs WHERE id = $1;

-- name: CountJobs :one
SELECT COUNT(*) FROM jobs;

-- name: CountJobsByStatus :one
SELECT COUNT(*) FROM jobs WHERE status = $1;

-- ============================================
-- TASKS
-- ============================================

-- name: CreateTask :one
INSERT INTO tasks (job_id, type, status, params, sequence_index, start_time_sec, end_time_sec)
VALUES ($1, $2, 'pending', $3, $4, $5, $6)
RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks WHERE id = $1 LIMIT 1;

-- name: GetPendingTasks :many
SELECT * FROM tasks
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1;

-- name: ListTasksByJob :many
SELECT * FROM tasks
WHERE job_id = $1
ORDER BY sequence_index ASC;

-- name: AssignTask :exec
UPDATE tasks
SET status = 'assigned', worker_id = $2, updated_at = NOW()
WHERE id = $1;

-- name: CompleteTask :exec
UPDATE tasks
SET status = 'completed', result = $2, output_key = $3, output_size_bytes = $4, updated_at = NOW()
WHERE id = $1;

-- name: FailTask :exec
UPDATE tasks
SET status = 'failed', result = $2, attempt_count = attempt_count + 1, updated_at = NOW()
WHERE id = $1;

-- name: IncrementTaskAttempt :exec
UPDATE tasks
SET attempt_count = attempt_count + 1, updated_at = NOW()
WHERE id = $1;

-- name: CountTasksByJobAndStatus :one
SELECT COUNT(*) FROM tasks WHERE job_id = $1 AND status = $2;

-- name: GetCompletedTaskOutputs :many
SELECT output_key, sequence_index FROM tasks
WHERE job_id = $1 AND status = 'completed'
ORDER BY sequence_index ASC;

-- name: UpdateTaskStatus :exec
UPDATE tasks
SET status = $2, result = $3, updated_at = NOW()
WHERE id = $1;

-- name: CountPendingTasksForJob :one
SELECT COUNT(*) FROM tasks 
WHERE job_id = $1 AND status NOT IN ('completed', 'failed');

-- ============================================
-- STREAMS
-- ============================================

-- name: CreateStream :one
INSERT INTO streams (stream_key, user_id, title, description, archive_enabled, ingest_server)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetStream :one
SELECT * FROM streams WHERE id = $1 LIMIT 1;

-- name: GetStreamByKey :one
SELECT * FROM streams WHERE stream_key = $1 LIMIT 1;

-- name: ListStreams :many
SELECT * FROM streams
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListStreamsByUser :many
SELECT * FROM streams
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListLiveStreams :many
SELECT * FROM streams
WHERE is_live = TRUE
ORDER BY started_at DESC;

-- name: UpdateStreamLive :exec
UPDATE streams
SET is_live = $2, started_at = CASE WHEN $2 THEN NOW() ELSE started_at END, 
    ended_at = CASE WHEN NOT $2 THEN NOW() ELSE ended_at END
WHERE id = $1;

-- name: UpdateStreamStats :exec
UPDATE streams
SET current_viewers = $2, last_stats = $3
WHERE id = $1;

-- name: UpdateStreamMetadata :exec
UPDATE streams
SET title = $2, description = $3, thumbnail_url = $4
WHERE id = $1;

-- name: SetStreamArchiveJob :exec
UPDATE streams
SET archive_vod_job_id = $2
WHERE id = $1;

-- name: UpdateStreamRestreamDestinations :exec
UPDATE streams
SET restream_destinations = $2
WHERE id = $1;

-- name: DeleteStream :exec
DELETE FROM streams WHERE id = $1;

-- ============================================
-- WORKERS
-- ============================================

-- name: RegisterWorker :exec
INSERT INTO workers (id, hostname, version, last_seen, status, capacity, ip_address, port, capabilities, is_healthy)
VALUES ($1, $2, $3, NOW(), $4, $5, $6, $7, $8, TRUE)
ON CONFLICT (id) DO UPDATE
SET last_seen = NOW(), status = EXCLUDED.status, capacity = EXCLUDED.capacity, 
    capabilities = EXCLUDED.capabilities, is_healthy = TRUE;

-- name: GetWorker :one
SELECT * FROM workers WHERE id = $1 LIMIT 1;

-- name: ListWorkers :many
SELECT * FROM workers
ORDER BY last_seen DESC;

-- name: ListHealthyWorkers :many
SELECT * FROM workers
WHERE is_healthy = TRUE
ORDER BY last_seen DESC;

-- name: UpdateWorkerHeartbeat :exec
UPDATE workers
SET last_seen = NOW(), is_healthy = TRUE
WHERE id = $1;

-- name: UpdateWorkerStatus :exec
UPDATE workers
SET status = $2, is_healthy = $3
WHERE id = $1;

-- name: MarkWorkersUnhealthy :many
UPDATE workers
SET is_healthy = FALSE
WHERE is_healthy = TRUE AND last_seen < NOW() - INTERVAL '30 seconds'
RETURNING *;

-- name: DeleteWorker :exec
DELETE FROM workers WHERE id = $1;

-- ============================================
-- PLUGIN CONFIGS
-- ============================================

-- name: RegisterPluginConfig :one
INSERT INTO plugin_configs (id, plugin_type, config_json, is_enabled, priority)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE
SET plugin_type = EXCLUDED.plugin_type, updated_at = NOW()
RETURNING *;

-- name: GetPluginConfig :one
SELECT * FROM plugin_configs WHERE id = $1 LIMIT 1;

-- name: ListPluginConfigs :many
SELECT * FROM plugin_configs
ORDER BY plugin_type, priority DESC;

-- name: ListPluginConfigsByType :many
SELECT * FROM plugin_configs
WHERE plugin_type = $1
ORDER BY priority DESC;

-- name: UpdatePluginConfig :exec
UPDATE plugin_configs
SET config_json = $2, updated_at = NOW()
WHERE id = $1;

-- name: EnablePlugin :exec
UPDATE plugin_configs
SET is_enabled = TRUE, updated_at = NOW()
WHERE id = $1;

-- name: DisablePlugin :exec
UPDATE plugin_configs
SET is_enabled = FALSE, updated_at = NOW()
WHERE id = $1;

-- name: DeletePluginConfig :exec
DELETE FROM plugin_configs WHERE id = $1;

-- ============================================
-- RESTREAM JOBS
-- ============================================

-- name: CreateRestreamJob :one
INSERT INTO restream_jobs (user_id, title, description, input_type, input_url, output_destinations, schedule_type, schedule_config, loop_enabled, simulate_live)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetRestreamJob :one
SELECT * FROM restream_jobs WHERE id = $1 LIMIT 1;

-- name: ListRestreamJobs :many
SELECT * FROM restream_jobs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListRestreamJobsByUser :many
SELECT * FROM restream_jobs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateRestreamJobStatus :exec
UPDATE restream_jobs
SET status = $2, current_stats = $3, updated_at = NOW(),
    started_at = CASE WHEN $2 = 'streaming' THEN NOW() ELSE started_at END,
    stopped_at = CASE WHEN $2 = 'stopped' THEN NOW() ELSE stopped_at END
WHERE id = $1;

-- name: DeleteRestreamJob :exec
DELETE FROM restream_jobs WHERE id = $1;

-- ============================================
-- ENCODING PROFILES
-- ============================================

-- name: GetEncodingProfile :one
SELECT * FROM encoding_profiles WHERE id = $1 LIMIT 1;

-- name: ListEncodingProfiles :many
SELECT * FROM encoding_profiles
ORDER BY is_system DESC, name ASC;

-- name: CreateEncodingProfile :one
INSERT INTO encoding_profiles (id, name, description, video_codec, audio_codec, width, height, bitrate_kbps, preset, container, config_json)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: UpdateEncodingProfile :exec
UPDATE encoding_profiles
SET name = $2, description = $3, video_codec = $4, audio_codec = $5, 
    width = $6, height = $7, bitrate_kbps = $8, preset = $9, container = $10, 
    config_json = $11, updated_at = NOW()
WHERE id = $1 AND is_system = FALSE;

-- name: DeleteEncodingProfile :exec
DELETE FROM encoding_profiles WHERE id = $1 AND is_system = FALSE;

-- ============================================
-- AUDIT LOG
-- ============================================

-- name: CreateAuditLog :exec
INSERT INTO audit_log (user_id, action, resource_type, resource_id, details, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListAuditLogs :many
SELECT * FROM audit_log
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAuditLogsByUser :many
SELECT * FROM audit_log
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- ============================================
-- WEBHOOKS
-- ============================================

-- name: CreateWebhook :one
INSERT INTO webhooks (user_id, url, secret, events)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWebhook :one
SELECT * FROM webhooks WHERE id = $1 LIMIT 1;

-- name: ListWebhooks :many
SELECT * FROM webhooks
ORDER BY created_at DESC;

-- name: ListWebhooksByUser :many
SELECT * FROM webhooks
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListActiveWebhooksForEvent :many
SELECT * FROM webhooks
WHERE is_active = TRUE AND @event::TEXT = ANY(events);

-- name: UpdateWebhookTriggered :exec
UPDATE webhooks
SET last_triggered_at = NOW(), failure_count = 0, updated_at = NOW()
WHERE id = $1;

-- name: IncrementWebhookFailure :exec
UPDATE webhooks
SET failure_count = failure_count + 1, updated_at = NOW()
WHERE id = $1;

-- name: DeactivateWebhook :exec
UPDATE webhooks
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1;

-- name: DeleteWebhook :exec
DELETE FROM webhooks WHERE id = $1;

-- ============================================
-- CLEANUP QUERIES
-- ============================================

-- name: DeleteOldJobsByStatus :execrows
DELETE FROM jobs 
WHERE status = $1 AND created_at < $2;

-- name: DeleteOrphanedTasks :execrows
DELETE FROM tasks 
WHERE job_id NOT IN (SELECT id FROM jobs);

-- name: DeleteOldAuditLogs :execrows
DELETE FROM audit_log 
WHERE created_at < $1;

-- name: DeleteOldWorkers :execrows
DELETE FROM workers
WHERE is_healthy = FALSE AND last_seen < $1;

-- ============================================
-- GLOBAL ERROR TRACKING
-- ============================================

-- name: CreateErrorEvent :one
INSERT INTO error_events (source_component, severity, message, stack_trace, context_data)
VALUES ($1, $2::error_severity, $3, $4, $5)
RETURNING *;

-- name: ListErrorEvents :many
SELECT * FROM error_events
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListErrorEventsBySource :many
SELECT * FROM error_events
WHERE source_component = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ResolveErrorEvent :exec
UPDATE error_events
SET resolved = TRUE
WHERE id = $1;

-- name: DeleteOldErrorEvents :execrows
DELETE FROM error_events
WHERE created_at < $1;

-- ============================================
-- NOTIFICATIONS
-- ============================================

-- name: CreateNotification :one
INSERT INTO notifications (user_id, title, message, link, type, is_read)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListNotifications :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUnreadNotificationCount :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1 AND is_read = FALSE;

-- name: MarkNotificationRead :exec
UPDATE notifications
SET is_read = TRUE
WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET is_read = TRUE
WHERE user_id = $1;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1 AND user_id = $2;

-- name: DeleteOldNotifications :exec
DELETE FROM notifications
WHERE user_id = $1 AND created_at < NOW() - INTERVAL '30 days';

-- ============================================
-- JOB LOGS
-- ============================================

-- name: CreateJobLog :exec
INSERT INTO job_logs (job_id, level, message, metadata)
VALUES ($1, $2, $3, $4);

-- name: ListJobLogs :many
SELECT * FROM job_logs
WHERE job_id = $1
ORDER BY created_at ASC;
