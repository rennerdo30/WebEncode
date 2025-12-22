CREATE TYPE job_status AS ENUM ('queued', 'processing', 'completed', 'failed', 'cancelled');
CREATE TYPE task_type AS ENUM ('probe', 'transcode', 'stitch', 'manifest', 'restream');

CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_url TEXT NOT NULL,
    profiles TEXT[] NOT NULL, -- JSON array ["1080p", "720p"]
    status job_status NOT NULL DEFAULT 'queued',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metadata JSONB -- Flexible metadata
);

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    type task_type NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending, assigned, completed, failed
    params JSONB NOT NULL, -- CLI args or config
    worker_id TEXT, -- ID of the worker processing this
    result JSONB, -- Output from the worker
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE streams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_key TEXT NOT NULL UNIQUE,
    is_live BOOLEAN NOT NULL DEFAULT FALSE,
    ingest_url TEXT,
    playback_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE workers (
    id TEXT PRIMARY KEY,
    hostname TEXT NOT NULL,
    version TEXT NOT NULL,
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'idle',
    capacity JSONB -- CPU/GPU capabilities
);
