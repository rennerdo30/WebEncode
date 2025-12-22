CREATE TYPE error_severity AS ENUM ('warning', 'error', 'critical', 'fatal');
CREATE TYPE error_source AS ENUM ('frontend', 'kernel', 'worker', 'plugin');

CREATE TABLE error_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_component TEXT NOT NULL, -- e.g. "frontend", "kernel", "plugin:s3"
    severity error_severity NOT NULL DEFAULT 'error',
    message TEXT NOT NULL,
    stack_trace TEXT,
    context_data JSONB DEFAULT '{}',
    resolved BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for quick filtering by source and time
CREATE INDEX idx_error_events_source ON error_events(source_component);
CREATE INDEX idx_error_events_created_at ON error_events(created_at DESC);
CREATE INDEX idx_error_events_severity ON error_events(severity);
