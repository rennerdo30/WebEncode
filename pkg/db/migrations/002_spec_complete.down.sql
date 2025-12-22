-- Migration 002 DOWN: Rollback spec complete schema

-- Drop new tables
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS encoding_profiles;
DROP TABLE IF EXISTS restream_jobs;
DROP TABLE IF EXISTS plugin_configs;

-- Drop indexes
DROP INDEX IF EXISTS idx_workers_healthy;
DROP INDEX IF EXISTS idx_streams_live;
DROP INDEX IF EXISTS idx_streams_user;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_job;
DROP INDEX IF EXISTS idx_jobs_created;
DROP INDEX IF EXISTS idx_jobs_status;
DROP INDEX IF EXISTS idx_jobs_user;

-- Remove foreign keys
ALTER TABLE streams DROP CONSTRAINT IF EXISTS fk_streams_archive_job;
ALTER TABLE jobs DROP CONSTRAINT IF EXISTS fk_jobs_source_stream;

-- Remove columns from workers
ALTER TABLE workers
    DROP COLUMN IF EXISTS ip_address,
    DROP COLUMN IF EXISTS port,
    DROP COLUMN IF EXISTS capabilities,
    DROP COLUMN IF EXISTS is_healthy,
    DROP COLUMN IF EXISTS created_at;

-- Remove columns from streams
ALTER TABLE streams
    DROP COLUMN IF EXISTS user_id,
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS description,
    DROP COLUMN IF EXISTS thumbnail_url,
    DROP COLUMN IF EXISTS ingest_server,
    DROP COLUMN IF EXISTS current_viewers,
    DROP COLUMN IF EXISTS total_viewers_lifetime,
    DROP COLUMN IF EXISTS started_at,
    DROP COLUMN IF EXISTS ended_at,
    DROP COLUMN IF EXISTS last_stats,
    DROP COLUMN IF EXISTS archive_enabled,
    DROP COLUMN IF EXISTS archive_vod_job_id,
    DROP COLUMN IF EXISTS restream_destinations;

-- Remove columns from tasks
ALTER TABLE tasks
    DROP COLUMN IF EXISTS sequence_index,
    DROP COLUMN IF EXISTS start_time_sec,
    DROP COLUMN IF EXISTS end_time_sec,
    DROP COLUMN IF EXISTS attempt_count,
    DROP COLUMN IF EXISTS max_attempts,
    DROP COLUMN IF EXISTS output_key,
    DROP COLUMN IF EXISTS output_size_bytes;

-- Remove columns from jobs
ALTER TABLE jobs
    DROP COLUMN IF EXISTS user_id,
    DROP COLUMN IF EXISTS source_type,
    DROP COLUMN IF EXISTS source_stream_id,
    DROP COLUMN IF EXISTS profile_id,
    DROP COLUMN IF EXISTS output_config,
    DROP COLUMN IF EXISTS progress_pct,
    DROP COLUMN IF EXISTS error_message,
    DROP COLUMN IF EXISTS assigned_to_worker_id,
    DROP COLUMN IF EXISTS started_at,
    DROP COLUMN IF EXISTS finished_at,
    DROP COLUMN IF EXISTS eta_seconds;

-- Note: Cannot remove enum values in PostgreSQL, so leaving job_source_type
DROP TYPE IF EXISTS job_source_type;
