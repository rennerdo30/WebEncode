-- Migration 002: Complete schema per SPECIFICATION.md
-- This adds all missing fields and tables

-- Add new enum types
CREATE TYPE job_source_type AS ENUM ('url', 'upload', 'stream', 'restream');

-- ============================================
-- Update jobs table
-- ============================================
ALTER TABLE jobs
    ADD COLUMN user_id UUID,
    ADD COLUMN source_type job_source_type DEFAULT 'url',
    ADD COLUMN source_stream_id UUID,
    ADD COLUMN profile_id VARCHAR(100) DEFAULT 'default',
    ADD COLUMN output_config JSONB DEFAULT '{}',
    ADD COLUMN progress_pct INT DEFAULT 0,
    ADD COLUMN error_message TEXT,
    ADD COLUMN assigned_to_worker_id VARCHAR(100),
    ADD COLUMN started_at TIMESTAMPTZ,
    ADD COLUMN finished_at TIMESTAMPTZ,
    ADD COLUMN eta_seconds INT;

-- Add 'stitching' and 'uploading' to job_status enum
ALTER TYPE job_status ADD VALUE IF NOT EXISTS 'stitching';
ALTER TYPE job_status ADD VALUE IF NOT EXISTS 'uploading';

-- Indexes for jobs
CREATE INDEX IF NOT EXISTS idx_jobs_user ON jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_created ON jobs(created_at DESC);

-- ============================================
-- Update tasks table
-- ============================================
ALTER TABLE tasks
    ADD COLUMN sequence_index INT DEFAULT 0,
    ADD COLUMN start_time_sec DOUBLE PRECISION DEFAULT 0,
    ADD COLUMN end_time_sec DOUBLE PRECISION DEFAULT 0,
    ADD COLUMN attempt_count INT DEFAULT 0,
    ADD COLUMN max_attempts INT DEFAULT 3,
    ADD COLUMN output_key TEXT,
    ADD COLUMN output_size_bytes BIGINT;

-- Indexes for tasks
CREATE INDEX IF NOT EXISTS idx_tasks_job ON tasks(job_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);

-- ============================================
-- Update streams table
-- ============================================
ALTER TABLE streams
    ADD COLUMN user_id UUID,
    ADD COLUMN title VARCHAR(255),
    ADD COLUMN description TEXT,
    ADD COLUMN thumbnail_url TEXT,
    ADD COLUMN ingest_server VARCHAR(100),
    ADD COLUMN current_viewers INT DEFAULT 0,
    ADD COLUMN total_viewers_lifetime BIGINT DEFAULT 0,
    ADD COLUMN started_at TIMESTAMPTZ,
    ADD COLUMN ended_at TIMESTAMPTZ,
    ADD COLUMN last_stats JSONB,
    ADD COLUMN archive_enabled BOOLEAN DEFAULT TRUE,
    ADD COLUMN archive_vod_job_id UUID,
    ADD COLUMN restream_destinations JSONB DEFAULT '[]';

-- Foreign key for archive_vod_job_id
ALTER TABLE streams
    ADD CONSTRAINT fk_streams_archive_job
    FOREIGN KEY (archive_vod_job_id) REFERENCES jobs(id);

-- Foreign key for jobs.source_stream_id
ALTER TABLE jobs
    ADD CONSTRAINT fk_jobs_source_stream
    FOREIGN KEY (source_stream_id) REFERENCES streams(id);

-- Indexes for streams
CREATE INDEX IF NOT EXISTS idx_streams_user ON streams(user_id);
CREATE INDEX IF NOT EXISTS idx_streams_live ON streams(is_live);

-- ============================================
-- Update workers table
-- ============================================
ALTER TABLE workers
    ADD COLUMN ip_address INET,
    ADD COLUMN port INT,
    ADD COLUMN capabilities JSONB,
    ADD COLUMN is_healthy BOOLEAN DEFAULT TRUE,
    ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();

-- Rename capacity to capabilities if needed (handle both cases)
-- (capacity already exists, we'll keep both for compatibility)

-- Index for workers
CREATE INDEX IF NOT EXISTS idx_workers_healthy ON workers(is_healthy);

-- ============================================
-- New table: plugin_configs
-- ============================================
CREATE TABLE IF NOT EXISTS plugin_configs (
    id VARCHAR(100) PRIMARY KEY,
    plugin_type VARCHAR(50) NOT NULL,
    config_json JSONB NOT NULL,
    is_enabled BOOLEAN DEFAULT TRUE,
    priority INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plugin_configs_type ON plugin_configs(plugin_type);

-- ============================================
-- New table: restream_jobs
-- ============================================
CREATE TABLE IF NOT EXISTS restream_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    title VARCHAR(255),
    description TEXT,
    input_type VARCHAR(20), -- 'rtmp', 'file', 'vod_job'
    input_url TEXT,
    output_destinations JSONB NOT NULL DEFAULT '[]',
    schedule_type VARCHAR(20) DEFAULT 'immediate', -- 'immediate', 'scheduled', 'recurring'
    schedule_config JSONB,
    loop_enabled BOOLEAN DEFAULT FALSE,
    simulate_live BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'stopped', -- 'streaming', 'stopped', 'error'
    current_stats JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    stopped_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_restream_user ON restream_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_restream_status ON restream_jobs(status);

-- ============================================
-- New table: encoding_profiles
-- ============================================
CREATE TABLE IF NOT EXISTS encoding_profiles (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    video_codec VARCHAR(50) NOT NULL,
    audio_codec VARCHAR(50) DEFAULT 'aac',
    width INT,
    height INT,
    bitrate_kbps INT,
    preset VARCHAR(50) DEFAULT 'fast',
    container VARCHAR(20) DEFAULT 'mp4',
    is_system BOOLEAN DEFAULT FALSE, -- System profiles can't be deleted
    config_json JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Insert default profiles
INSERT INTO encoding_profiles (id, name, description, video_codec, audio_codec, width, height, bitrate_kbps, preset, is_system) VALUES
    ('1080p_h264', '1080p H.264', 'Full HD H.264 encoding', 'libx264', 'aac', 1920, 1080, 5000, 'fast', TRUE),
    ('720p_h264', '720p H.264', 'HD H.264 encoding', 'libx264', 'aac', 1280, 720, 2500, 'fast', TRUE),
    ('480p_h264', '480p H.264', 'SD H.264 encoding', 'libx264', 'aac', 854, 480, 1000, 'fast', TRUE),
    ('4k_hevc', '4K HEVC', '4K H.265/HEVC encoding', 'libx265', 'aac', 3840, 2160, 15000, 'medium', TRUE),
    ('1080p_vp9', '1080p VP9', 'Full HD VP9 WebM encoding', 'libvpx-vp9', 'libopus', 1920, 1080, 4000, 'good', TRUE)
ON CONFLICT (id) DO NOTHING;

-- ============================================
-- New table: audit_log
-- ============================================
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id TEXT,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_log(created_at DESC);

-- ============================================
-- New table: webhooks
-- ============================================
CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    url TEXT NOT NULL,
    secret TEXT NOT NULL,
    events TEXT[] NOT NULL, -- ['job.completed', 'stream.started', etc.]
    is_active BOOLEAN DEFAULT TRUE,
    last_triggered_at TIMESTAMPTZ,
    failure_count INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_webhooks_user ON webhooks(user_id);
CREATE INDEX IF NOT EXISTS idx_webhooks_active ON webhooks(is_active);
