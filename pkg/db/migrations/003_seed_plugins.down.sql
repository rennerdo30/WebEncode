-- Migration 003 down: Remove seeded plugins
DELETE FROM plugin_configs WHERE id IN (
    'auth-basic', 'auth-oidc', 'auth-ldap',
    'storage-fs', 'storage-s3',
    'encoder-ffmpeg',
    'live-mediamtx',
    'publisher-youtube', 'publisher-twitch', 'publisher-kick', 'publisher-rumble', 'publisher-dummy'
);
