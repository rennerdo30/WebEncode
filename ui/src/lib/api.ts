// API Types
export interface Job {
    id: string;
    source_url: string;
    profiles: string[];
    status: 'queued' | 'processing' | 'completed' | 'failed' | 'cancelled' | 'stitching' | 'uploading';
    created_at: string;
    updated_at: string;
    progress_pct: number | null;
    error_message: string | null;
    started_at: string | null;
    finished_at: string | null;
    eta_seconds: number | null;
}

export interface Task {
    id: string;
    job_id: string;
    type: 'probe' | 'transcode' | 'stitch' | 'upload';
    status: 'pending' | 'assigned' | 'completed' | 'failed';
    sequence_index: number;
    created_at: string;
    updated_at: string;
    start_time_sec: number | null;
    end_time_sec: number | null;
    output_key: string | null;
    result?: any;
}

export interface JobDetail {
    job: Job;
    tasks: Task[];
}

export interface Stream {
    id: string;
    stream_key: string;
    is_live: boolean;
    title: string | null;
    description: string | null;
    ingest_url: string | null;
    playback_url: string | null;
    current_viewers: number;
    created_at: string;
}

export interface Worker {
    id: string;
    is_healthy: boolean;
    current_task_id: string | null;
    last_heartbeat: string;
    capabilities: Record<string, unknown> | null;
}

export interface DashboardStats {
    activeJobs: number;
    workersOnline: number;
    completedJobs: number;
    liveStreams: number;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL || '/api/v1';
const DIRECT_UPLOAD_URL = process.env.NEXT_PUBLIC_DIRECT_UPLOAD_URL;

// Jobs API
export async function fetchJobs(limit = 20, offset = 0): Promise<Job[]> {
    const res = await fetch(`${API_BASE}/jobs?limit=${limit}&offset=${offset}`);
    if (!res.ok) throw new Error('Failed to fetch jobs');
    return res.json();
}

export async function fetchJob(id: string): Promise<JobDetail> {
    const res = await fetch(`${API_BASE}/jobs/${id}`);
    if (!res.ok) throw new Error('Failed to fetch job');
    return res.json();
}

export async function createJob(sourceUrl: string, profiles: string[]): Promise<Job> {
    const res = await fetch(`${API_BASE}/jobs`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ source_url: sourceUrl, profiles }),
    });
    if (!res.ok) throw new Error('Failed to create job');
    return res.json();
}

export async function cancelJob(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/jobs/${id}/cancel`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to cancel job');
}

export async function retryJob(id: string): Promise<Job> {
    const res = await fetch(`${API_BASE}/jobs/${id}/retry`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to retry job');
    return res.json();
}

export async function deleteJob(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/jobs/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error('Failed to delete job');
}

export interface JobLog {
    id: string;
    job_id: string;
    level: string;
    message: string;
    metadata: any;
    created_at: string;
}

export async function fetchJobLogs(id: string): Promise<JobLog[]> {
    const res = await fetch(`${API_BASE}/jobs/${id}/logs`);
    if (!res.ok) throw new Error('Failed to fetch job logs');
    return res.json();
}

// Job Outputs API
export interface JobOutput {
    name: string;
    type: 'final' | 'segment' | 'manifest';
    url: string;
    download_url?: string;
    size?: number;
    profile?: string;
}

export interface JobOutputsResponse {
    job_id: string;
    status: string;
    outputs: JobOutput[];
}

export async function fetchJobOutputs(id: string): Promise<JobOutputsResponse> {
    const res = await fetch(`${API_BASE}/jobs/${id}/outputs`);
    if (!res.ok) throw new Error('Failed to fetch job outputs');
    return res.json();
}

// Job Publishing API
export interface PublishJobRequest {
    platform: string;
    title: string;
    description?: string;
    access_token: string;
    output_key?: string;
}

export interface PublishJobResponse {
    success: boolean;
    platform_id?: string;
    platform_url?: string;
    message?: string;
}

export async function publishJob(id: string, request: PublishJobRequest): Promise<PublishJobResponse> {
    const res = await fetch(`${API_BASE}/jobs/${id}/publish`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request),
    });
    if (!res.ok) {
        const error = await res.text();
        throw new Error(error || 'Failed to publish job');
    }
    return res.json();
}

// Streams API
export async function fetchStreams(): Promise<Stream[]> {
    const res = await fetch(`${API_BASE}/streams`);
    if (!res.ok) throw new Error('Failed to fetch streams');
    return res.json();
}

export async function fetchStream(id: string): Promise<Stream> {
    const res = await fetch(`${API_BASE}/streams/${id}`);
    if (!res.ok) throw new Error('Failed to fetch stream');
    return res.json();
}

export async function createStream(title: string, description: string): Promise<Stream> {
    const res = await fetch(`${API_BASE}/streams`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title, description }),
    });
    if (!res.ok) throw new Error('Failed to create stream');
    return res.json();
}

// Stream Destinations (Restream v2)
export interface RestreamDestination {
    plugin_id: string;       // e.g., "publisher-twitch"
    access_token: string;    // OAuth token for the platform
    enabled: boolean;        // Whether this destination is active
}

export async function fetchStreamDestinations(streamId: string): Promise<RestreamDestination[]> {
    const res = await fetch(`${API_BASE}/streams/${streamId}/destinations`);
    if (!res.ok) throw new Error('Failed to fetch stream destinations');
    return res.json();
}

export async function updateStreamDestinations(streamId: string, destinations: RestreamDestination[]): Promise<void> {
    const res = await fetch(`${API_BASE}/streams/${streamId}/destinations`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(destinations),
    });
    if (!res.ok) throw new Error('Failed to update stream destinations');
}

// Workers API
export async function fetchWorkers(): Promise<Worker[]> {
    const res = await fetch(`${API_BASE}/workers`);
    if (!res.ok) throw new Error('Failed to fetch workers');
    return res.json();
}

// Dashboard Stats (aggregated client-side for now)
export async function fetchDashboardStats(): Promise<DashboardStats> {
    try {
        const [jobs, workers, streams] = await Promise.all([
            fetchJobs(100, 0).catch(() => []),
            fetchWorkers().catch(() => []),
            fetchStreams().catch(() => []),
        ]);

        const activeJobs = jobs.filter(j =>
            j.status === 'processing' || j.status === 'queued' || j.status === 'stitching'
        ).length;

        const completedJobs = jobs.filter(j => j.status === 'completed').length;
        const workersOnline = workers.filter(w => w.is_healthy).length;
        const liveStreams = streams.filter(s => s.is_live).length;

        return { activeJobs, workersOnline, completedJobs, liveStreams };
    } catch {
        return { activeJobs: 0, workersOnline: 0, completedJobs: 0, liveStreams: 0 };
    }
}

// Restreams API
export interface RestreamJob {
    id: string;
    title: string | null;
    description: string | null;
    input_type: string | null;
    input_url: string | null;
    output_destinations: { platform: string; url: string; enabled: boolean }[];
    status: string;
    created_at: string;
}

export async function fetchRestreams(): Promise<RestreamJob[]> {
    const res = await fetch(`${API_BASE}/restreams`);
    if (!res.ok) throw new Error('Failed to fetch restreams');
    return res.json();
}

export async function fetchRestream(id: string): Promise<RestreamJob> {
    const res = await fetch(`${API_BASE}/restreams/${id}`);
    if (!res.ok) throw new Error('Failed to fetch restream');
    return res.json();
}

export async function createRestream(data: {
    title: string;
    input_type: string;
    input_url: string;
    output_destinations: { platform: string; url: string; enabled: boolean }[];
}): Promise<RestreamJob> {
    const res = await fetch(`${API_BASE}/restreams`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    });
    if (!res.ok) throw new Error('Failed to create restream');
    return res.json();
}

export async function startRestream(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/restreams/${id}/start`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to start restream');
}

export async function stopRestream(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/restreams/${id}/stop`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to stop restream');
}

// Profiles API

/** Configuration for advanced encoding options stored in profile config JSON */
export interface ProfileConfig {
    // Rate Control
    rate_control_mode?: 'crf' | 'cbr' | 'vbr' | '2pass';
    crf?: number;
    max_bitrate_kbps?: number;
    min_bitrate_kbps?: number;
    buffer_size_kbps?: number;

    // Video Advanced
    frame_rate?: string;
    pixel_format?: string;
    tune?: string;
    profile?: string;
    level?: string;
    gop_size?: number;
    b_frames?: number;
    ref_frames?: number;
    scene_change_threshold?: number;
    lookahead?: number;

    // Hardware Acceleration
    hw_accel?: 'none' | 'nvenc' | 'qsv' | 'videotoolbox' | 'vaapi' | 'amf';
    hw_device?: string;

    // Audio Advanced
    audio_bitrate_kbps?: number;
    audio_sample_rate?: number;
    audio_channels?: number;
    audio_normalize?: 'off' | 'ebur128' | 'peak';

    // Filters
    deinterlace?: 'off' | 'yadif' | 'bwdif';
    denoise?: 'off' | 'hqdn3d' | 'nlmeans';
    sharpen?: 'off' | 'unsharp' | 'cas';
    scale_algorithm?: 'bilinear' | 'bicubic' | 'lanczos' | 'spline';
    crop_top?: number;
    crop_bottom?: number;
    crop_left?: number;
    crop_right?: number;
    pad_top?: number;
    pad_bottom?: number;
    pad_left?: number;
    pad_right?: number;
    custom_filters?: string;

    // Output
    fast_start?: boolean;
    metadata_title?: string;
    metadata_author?: string;
    subtitle_mode?: 'copy' | 'remove' | 'burn';

    // Legacy/custom options
    [key: string]: unknown;
}

export interface Profile {
    id: string;
    name: string;
    description?: string;
    video_codec: string;
    audio_codec?: string;
    width?: number;
    height?: number;
    bitrate_kbps?: number;
    preset?: string;
    container?: string;
    config?: ProfileConfig;
    is_system: boolean;
}

export async function fetchProfiles(): Promise<Profile[]> {
    const res = await fetch(`${API_BASE}/profiles`);
    if (!res.ok) throw new Error('Failed to fetch profiles');
    return res.json();
}

export async function fetchProfile(id: string): Promise<Profile> {
    const res = await fetch(`${API_BASE}/profiles/${id}`);
    if (!res.ok) throw new Error('Failed to fetch profile');
    return res.json();
}

export async function createProfile(data: Partial<Profile>): Promise<Profile> {
    const res = await fetch(`${API_BASE}/profiles`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    });
    if (!res.ok) throw new Error('Failed to create profile');
    return res.json();
}

export async function updateProfile(id: string, data: Partial<Profile>): Promise<void> {
    const res = await fetch(`${API_BASE}/profiles/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    });
    if (!res.ok) throw new Error('Failed to update profile');
}

export async function deleteProfile(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/profiles/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error('Failed to delete profile');
}


// Plugins API
export interface Plugin {
    id: string;
    type: string;
    config?: Record<string, unknown>;
    is_enabled: boolean;
    priority: number;
    health: 'healthy' | 'degraded' | 'failed' | 'disabled';
    version?: string;
}

export async function fetchPlugins(): Promise<Plugin[]> {
    const res = await fetch(`${API_BASE}/plugins`);
    if (!res.ok) throw new Error('Failed to fetch plugins');
    return res.json();
}

export async function enablePlugin(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/plugins/${id}/enable`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to enable plugin');
}

export async function disablePlugin(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/plugins/${id}/disable`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to disable plugin');
}

export async function fetchPlugin(id: string): Promise<Plugin> {
    const res = await fetch(`${API_BASE}/plugins/${id}`);
    if (!res.ok) throw new Error('Failed to fetch plugin');
    return res.json();
}

export interface UpdatePluginRequest {
    config: Record<string, unknown>;
    priority?: number;
}

export async function updatePluginConfig(id: string, request: UpdatePluginRequest): Promise<void> {
    const res = await fetch(`${API_BASE}/plugins/${id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request),
    });
    if (!res.ok) throw new Error('Failed to update plugin config');
}

// Audit API
export interface AuditLog {
    id: string;
    user_id: string;
    action: string;
    resource_type: string;
    resource_id: string | null;
    details: Record<string, unknown> | null;
    ip_address: string | null;
    created_at: string;
}

export async function fetchAuditLogs(limit = 50, offset = 0): Promise<AuditLog[]> {
    const res = await fetch(`${API_BASE}/audit?limit=${limit}&offset=${offset}`);
    if (!res.ok) throw new Error('Failed to fetch audit logs');
    return res.json();
}

// SSE for real-time updates
export function subscribeToJobUpdates(onEvent: (data: unknown) => void): EventSource {
    const es = new EventSource(`${API_BASE}/events/jobs`);
    es.onmessage = (e) => onEvent(JSON.parse(e.data));
    return es;
}

export function subscribeToJobDetail(jobId: string, onEvent: (data: unknown) => void): EventSource {
    const es = new EventSource(`${API_BASE}/events/jobs/${jobId}`);
    es.onmessage = (e) => onEvent(JSON.parse(e.data));
    return es;
}

export function subscribeToDashboard(onEvent: (data: unknown) => void): EventSource {
    const es = new EventSource(`${API_BASE}/events/dashboard`);
    es.onmessage = (e) => onEvent(JSON.parse(e.data));
    return es;
}

// System API
export interface SystemHealth {
    status: 'healthy' | 'degraded' | 'unhealthy';
    version: string;
    services: Record<string, string>;
}

export interface SystemStats {
    jobs: {
        total: number;
        queued: number;
        processing: number;
        completed: number;
        failed: number;
    };
    workers: {
        total: number;
        healthy: number;
    };
    streams: {
        total: number;
        live: number;
    };
    system: {
        go_version: string;
        num_goroutine: number;
        num_cpu: number;
    };
}

export async function fetchSystemHealth(): Promise<SystemHealth> {
    const res = await fetch(`${API_BASE}/system/health`);
    if (!res.ok) throw new Error('Failed to fetch system health');
    return res.json();
}

export async function fetchSystemStats(): Promise<SystemStats> {
    const res = await fetch(`${API_BASE}/system/stats`);
    if (!res.ok) throw new Error('Failed to fetch system stats');
    return res.json();
}

// Webhooks API
export interface Webhook {
    id: string;
    url: string;
    events: string[];
    is_active: boolean;
    failure_count: number;
    last_triggered_at: string | null;
    created_at: string;
}

export async function fetchWebhooks(): Promise<Webhook[]> {
    const res = await fetch(`${API_BASE}/webhooks`);
    if (!res.ok) throw new Error('Failed to fetch webhooks');
    return res.json();
}

export async function createWebhook(url: string, secret: string, events: string[]): Promise<Webhook> {
    const res = await fetch(`${API_BASE}/webhooks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url, secret, events }),
    });
    if (!res.ok) throw new Error('Failed to create webhook');
    return res.json();
}

export async function deleteWebhook(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/webhooks/${id}`, { method: 'DELETE' });
    if (!res.ok) throw new Error('Failed to delete webhook');
}

// File Browser API
export interface StoragePluginInfo {
    plugin_id: string;
    storage_type: string;
    supports_browse: boolean;
}

export interface BrowseRoot {
    plugin_id: string;
    storage_type: string;
    name: string;
    path: string;
}

export interface BrowseEntry {
    name: string;
    path: string;
    is_directory: boolean;
    size: number;
    mod_time: number;
    extension?: string;
    is_video: boolean;
    is_audio: boolean;
    is_image: boolean;
}

export interface BrowseResponse {
    plugin_id: string;
    current_path: string;
    parent_path: string;
    entries: BrowseEntry[];
}

export async function fetchStorageCapabilities(): Promise<StoragePluginInfo[]> {
    const res = await fetch(`${API_BASE}/files/capabilities`);
    if (!res.ok) throw new Error('Failed to fetch storage capabilities');
    return res.json();
}

export async function fetchBrowseRoots(): Promise<BrowseRoot[]> {
    const res = await fetch(`${API_BASE}/files/roots`);
    if (!res.ok) throw new Error('Failed to fetch browse roots');
    return res.json();
}

export interface BrowseOptions {
    plugin?: string;
    path?: string;
    mediaOnly?: boolean;
    showHidden?: boolean;
    search?: string;
}

export async function browseFiles(options: BrowseOptions = {}): Promise<BrowseResponse> {
    const params = new URLSearchParams();
    if (options.plugin) params.set('plugin', options.plugin);
    if (options.path) params.set('path', options.path);
    if (options.mediaOnly) params.set('media_only', 'true');
    if (options.showHidden) params.set('show_hidden', 'true');
    if (options.search) params.set('search', options.search);

    const res = await fetch(`${API_BASE}/files/browse?${params.toString()}`);
    if (!res.ok) throw new Error('Failed to browse files');
    return res.json();
}

// File Upload API
export interface UploadUrlRequest {
    plugin_id?: string;
    bucket?: string;
    object_key: string;
    content_type?: string;
    expiry_seconds?: number;
}

export interface UploadUrlResponse {
    url: string;
    headers?: Record<string, string>;
    expires_at: number;
    object_key: string;
    plugin_id: string;
}

export interface UploadResponse {
    url: string;
    plugin_id: string;
    object_key: string;
    size: number;
}

export interface UploadProgress {
    loaded: number;
    total: number;
    percentage: number;
}

/**
 * Get a pre-signed URL for direct upload
 * Use this for large files to upload directly to storage
 */
export async function getUploadUrl(request: UploadUrlRequest): Promise<UploadUrlResponse> {
    const res = await fetch(`${API_BASE}/files/upload-url`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request),
    });
    if (!res.ok) throw new Error('Failed to get upload URL');
    return res.json();
}

/**
 * Upload a file directly to the server
 * Returns upload response with URL that can be used as source_url for jobs
 */
export async function uploadFile(
    file: File,
    options: {
        pluginId?: string;
        bucket?: string;
        objectKey?: string;
        onProgress?: (progress: UploadProgress) => void;
    } = {}
): Promise<UploadResponse> {
    const formData = new FormData();
    formData.append('file', file);
    if (options.pluginId) formData.append('plugin_id', options.pluginId);
    if (options.bucket) formData.append('bucket', options.bucket);
    if (options.objectKey) formData.append('object_key', options.objectKey);

    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();

        xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable && options.onProgress) {
                options.onProgress({
                    loaded: e.loaded,
                    total: e.total,
                    percentage: Math.round((e.loaded / e.total) * 100),
                });
            }
        });

        xhr.addEventListener('load', () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                try {
                    resolve(JSON.parse(xhr.responseText));
                } catch {
                    reject(new Error('Invalid response from server'));
                }
            } else {
                // Try to extract detailed error message from response
                let errorMessage = xhr.statusText;
                if (xhr.responseText) {
                    try {
                        const errBody = JSON.parse(xhr.responseText);
                        // Standard error format { error: { message: ... } }
                        if (errBody.error?.message) {
                            errorMessage = errBody.error.message;
                        }
                        // Alternative format { message: ... }
                        else if (errBody.message) {
                            errorMessage = errBody.message;
                        }
                    } catch {
                        // If not JSON, use the raw text (often from http.Error)
                        errorMessage = xhr.responseText.trim(); // .trim() to remove newlines
                    }
                }
                reject(new Error(errorMessage || `Upload failed (${xhr.status})`));
            }
        });

        xhr.addEventListener('error', () => {
            reject(new Error('Network error during upload'));
        });

        xhr.addEventListener('abort', () => {
            reject(new Error('Upload cancelled'));
        });

        const uploadUrl = DIRECT_UPLOAD_URL || `${API_BASE}/files/upload`;
        xhr.open('POST', uploadUrl);
        xhr.send(formData);
    });
}

/**
 * Report a frontend error to the backend for tracking
 */
export async function reportError(
    message: string,
    source: string,
    stack?: string,
    context?: any
) {
    try {
        await fetch(`${API_BASE}/errors`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                source: "frontend",
                severity: "error",
                message: message,
                stack_trace: stack,
                context_data: {
                    ...context,
                    url: typeof window !== 'undefined' ? window.location.href : '',
                    userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : '',
                    source_module: source,
                },
            }),
        });
    } catch (err) {
        console.warn("Failed to report error to backend", err);
    }
}

/**
 * Upload a file using pre-signed URL (for direct-to-storage uploads)
 * Use this for very large files to bypass the API server
 */
export async function uploadToSignedUrl(
    file: File,
    uploadUrl: string,
    headers?: Record<string, string>,
    onProgress?: (progress: UploadProgress) => void
): Promise<void> {
    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();

        xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable && onProgress) {
                onProgress({
                    loaded: e.loaded,
                    total: e.total,
                    percentage: Math.round((e.loaded / e.total) * 100),
                });
            }
        });

        xhr.addEventListener('load', () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                resolve();
            } else {
                reject(new Error(`Upload failed: ${xhr.statusText}`));
            }
        });

        xhr.addEventListener('error', () => {
            reject(new Error('Upload failed'));
        });

        xhr.open('PUT', uploadUrl);

        // Set custom headers if provided
        if (headers) {
            for (const [key, value] of Object.entries(headers)) {
                xhr.setRequestHeader(key, value);
            }
        }
        xhr.setRequestHeader('Content-Type', file.type || 'application/octet-stream');

        xhr.send(file);
    });
}


// Chat API
export interface ChatMessage {
    id: string;
    platform: string;
    author_name: string;
    content: string;
    timestamp: number;
}

export interface SendChatMessageRequest {
    message: string;
    platform?: string; // Optional: send to specific platform only
}

export interface SendChatMessageResponse {
    success: boolean;
    sent_to: string[];
    errors?: string[];
}

export async function fetchStreamChat(streamId: string): Promise<ChatMessage[]> {
    const res = await fetch(`${API_BASE}/streams/${streamId}/chat`);
    if (!res.ok) throw new Error('Failed to fetch chat messages');
    return res.json();
}

export async function sendStreamChat(streamId: string, request: SendChatMessageRequest): Promise<SendChatMessageResponse> {
    const res = await fetch(`${API_BASE}/streams/${streamId}/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(request),
    });
    if (!res.ok) throw new Error('Failed to send chat message');
    return res.json();
}

// SSE for real-time chat updates
export function subscribeToStreamChat(streamId: string, onMessage: (msg: ChatMessage) => void): EventSource {
    const es = new EventSource(`${API_BASE}/events/streams/${streamId}/chat`);
    es.onmessage = (e) => onMessage(JSON.parse(e.data));
    return es;
}

// Notifications API
export interface Notification {
    id: string;
    user_id: string;
    title: string;
    message: string;
    link?: string;
    type: "info" | "success" | "warning" | "error";
    is_read: boolean;
    created_at: string;
}

export async function fetchNotifications(limit = 20, offset = 0): Promise<Notification[]> {
    const res = await fetch(`${API_BASE}/notifications?limit=${limit}&offset=${offset}`);
    if (!res.ok) throw new Error('Failed to fetch notifications');
    return res.json();
}

export async function markNotificationRead(id: string): Promise<void> {
    const res = await fetch(`${API_BASE}/notifications/${id}/read`, { method: 'PUT' });
    if (!res.ok) throw new Error('Failed to mark notification as read');
}

export async function clearAllNotifications(): Promise<void> {
    const res = await fetch(`${API_BASE}/notifications/clear`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to clear notifications');
}
