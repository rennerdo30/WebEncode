-- Migration 003: Seed default plugins
-- This seeds the default plugin configurations into the database

-- Auth Plugins
INSERT INTO plugin_configs (id, plugin_type, config_json, is_enabled, priority) VALUES
    ('auth-basic', 'auth', '{"strategy": "basic", "allow_anonymous": true}', TRUE, 100),
    ('auth-oidc', 'auth', '{"issuer": "", "client_id": "", "enabled": false}', FALSE, 90),
    ('auth-ldap', 'auth', '{"server": "", "base_dn": "", "enabled": false}', FALSE, 80)
ON CONFLICT (id) DO NOTHING;

-- Storage Plugins
INSERT INTO plugin_configs (id, plugin_type, config_json, is_enabled, priority) VALUES
    ('storage-fs', 'storage', '{"base_dir": "/data/webencode"}', TRUE, 100),
    ('storage-s3', 'storage', '{"endpoint": "seaweedfs-filer:8333", "bucket": "webencode", "region": "us-east-1"}', TRUE, 90)
ON CONFLICT (id) DO NOTHING;

-- Encoder Plugins
INSERT INTO plugin_configs (id, plugin_type, config_json, is_enabled, priority) VALUES
    ('encoder-ffmpeg', 'encoder', '{"ffmpeg_path": "/usr/bin/ffmpeg", "ffprobe_path": "/usr/bin/ffprobe", "max_parallel_tasks": 2}', TRUE, 100)
ON CONFLICT (id) DO NOTHING;

-- Live Plugins
INSERT INTO plugin_configs (id, plugin_type, config_json, is_enabled, priority) VALUES
    ('live-mediamtx', 'live', '{"api_url": "http://mediamtx:9997", "rtmp_port": 1935}', TRUE, 100)
ON CONFLICT (id) DO NOTHING;

-- Publisher Plugins
INSERT INTO plugin_configs (id, plugin_type, config_json, is_enabled, priority) VALUES
    ('publisher-youtube', 'publisher', '{"platform": "youtube", "enabled": false}', FALSE, 100),
    ('publisher-twitch', 'publisher', '{"platform": "twitch", "enabled": false}', FALSE, 90),
    ('publisher-kick', 'publisher', '{"platform": "kick", "enabled": false}', FALSE, 80),
    ('publisher-rumble', 'publisher', '{"platform": "rumble", "enabled": false}', FALSE, 70),
    ('publisher-dummy', 'publisher', '{"platform": "dummy", "dev_only": true}', TRUE, 50)
ON CONFLICT (id) DO NOTHING;
