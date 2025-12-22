import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import * as api from './api'

// Mock fetch globally
const mockFetch = vi.fn()
global.fetch = mockFetch

// Mock XMLHttpRequest for upload functions
const mockXhrInstances: MockXHR[] = []

class MockXHR {
  open = vi.fn()
  send = vi.fn()
  setRequestHeader = vi.fn()
  status = 200
  statusText = 'OK'
  responseText = ''
  upload = {
    addEventListener: vi.fn(),
  }
  addEventListener = vi.fn()

  constructor() {
    mockXhrInstances.push(this)
  }

  // Helper to trigger events
  triggerEvent(event: string, data?: any) {
    const listeners = this.addEventListener.mock.calls.filter(
      ([e]) => e === event
    )
    listeners.forEach(([, handler]) => handler(data))
  }

  triggerUploadEvent(event: string, data?: any) {
    const listeners = this.upload.addEventListener.mock.calls.filter(
      ([e]) => e === event
    )
    listeners.forEach(([, handler]) => handler(data))
  }
}

// Mock EventSource for SSE tests
class MockEventSource {
  url: string
  onmessage: ((event: { data: string }) => void) | null = null

  constructor(url: string) {
    this.url = url
  }
}

describe('API Module', () => {
  beforeEach(() => {
    mockFetch.mockReset()
    mockXhrInstances.length = 0
    // @ts-expect-error - mocking global
    global.XMLHttpRequest = MockXHR
    // @ts-expect-error - mocking global
    global.EventSource = MockEventSource
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  // Helper to create mock response
  const mockResponse = (data: unknown, ok = true, status = 200) => ({
    ok,
    status,
    statusText: ok ? 'OK' : 'Error',
    json: () => Promise.resolve(data),
    text: () => Promise.resolve(typeof data === 'string' ? data : JSON.stringify(data)),
  })

  describe('Jobs API', () => {
    describe('fetchJobs', () => {
      it('should fetch jobs with default pagination', async () => {
        const jobs: api.Job[] = [
          {
            id: '1',
            source_url: 'test.mp4',
            profiles: ['720p'],
            status: 'completed',
            created_at: '2024-01-01',
            updated_at: '2024-01-01',
            progress_pct: 100,
            error_message: null,
            started_at: null,
            finished_at: null,
            eta_seconds: null,
          },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(jobs))

        const result = await api.fetchJobs()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/jobs?limit=20&offset=0')
        expect(result).toEqual(jobs)
      })

      it('should fetch jobs with custom pagination', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse([]))

        await api.fetchJobs(50, 10)

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/jobs?limit=50&offset=10')
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchJobs()).rejects.toThrow('Failed to fetch jobs')
      })
    })

    describe('fetchJob', () => {
      it('should fetch a single job by id', async () => {
        const jobDetail: api.JobDetail = {
          job: {
            id: '123',
            source_url: 'test.mp4',
            profiles: ['720p'],
            status: 'processing',
            created_at: '2024-01-01',
            updated_at: '2024-01-01',
            progress_pct: 50,
            error_message: null,
            started_at: '2024-01-01',
            finished_at: null,
            eta_seconds: 120,
          },
          tasks: [],
        }
        mockFetch.mockResolvedValueOnce(mockResponse(jobDetail))

        const result = await api.fetchJob('123')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/jobs/123')
        expect(result).toEqual(jobDetail)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchJob('123')).rejects.toThrow('Failed to fetch job')
      })
    })

    describe('createJob', () => {
      it('should create a job with source URL and profiles', async () => {
        const newJob: api.Job = {
          id: 'new-1',
          source_url: 'test.mp4',
          profiles: ['720p', '1080p'],
          status: 'queued',
          created_at: '2024-01-01',
          updated_at: '2024-01-01',
          progress_pct: null,
          error_message: null,
          started_at: null,
          finished_at: null,
          eta_seconds: null,
        }
        mockFetch.mockResolvedValueOnce(mockResponse(newJob))

        const result = await api.createJob('test.mp4', ['720p', '1080p'])

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/jobs', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ source_url: 'test.mp4', profiles: ['720p', '1080p'] }),
        })
        expect(result).toEqual(newJob)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.createJob('test.mp4', ['720p'])).rejects.toThrow('Failed to create job')
      })
    })

    describe('cancelJob', () => {
      it('should cancel a job', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.cancelJob('123')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/jobs/123/cancel', { method: 'POST' })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.cancelJob('123')).rejects.toThrow('Failed to cancel job')
      })
    })

    describe('retryJob', () => {
      it('should retry a job', async () => {
        const retriedJob: api.Job = {
          id: '123',
          source_url: 'test.mp4',
          profiles: ['720p'],
          status: 'queued',
          created_at: '2024-01-01',
          updated_at: '2024-01-02',
          progress_pct: null,
          error_message: null,
          started_at: null,
          finished_at: null,
          eta_seconds: null,
        }
        mockFetch.mockResolvedValueOnce(mockResponse(retriedJob))

        const result = await api.retryJob('123')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/jobs/123/retry', { method: 'POST' })
        expect(result).toEqual(retriedJob)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.retryJob('123')).rejects.toThrow('Failed to retry job')
      })
    })

    describe('deleteJob', () => {
      it('should delete a job', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.deleteJob('123')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/jobs/123', { method: 'DELETE' })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.deleteJob('123')).rejects.toThrow('Failed to delete job')
      })
    })

    describe('fetchJobLogs', () => {
      it('should fetch job logs', async () => {
        const logs: api.JobLog[] = [
          {
            id: 'log-1',
            job_id: '123',
            level: 'info',
            message: 'Started processing',
            metadata: {},
            created_at: '2024-01-01',
          },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(logs))

        const result = await api.fetchJobLogs('123')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/jobs/123/logs')
        expect(result).toEqual(logs)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchJobLogs('123')).rejects.toThrow('Failed to fetch job logs')
      })
    })

    describe('fetchJobOutputs', () => {
      it('should fetch job outputs', async () => {
        const outputs: api.JobOutputsResponse = {
          job_id: '123',
          status: 'completed',
          outputs: [
            {
              name: 'output.mp4',
              type: 'final',
              url: 'https://storage.example.com/output.mp4',
              size: 1024000,
              profile: '720p',
            },
          ],
        }
        mockFetch.mockResolvedValueOnce(mockResponse(outputs))

        const result = await api.fetchJobOutputs('123')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/jobs/123/outputs')
        expect(result).toEqual(outputs)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchJobOutputs('123')).rejects.toThrow('Failed to fetch job outputs')
      })
    })

    describe('publishJob', () => {
      it('should publish a job', async () => {
        const request: api.PublishJobRequest = {
          platform: 'youtube',
          title: 'My Video',
          description: 'Test video',
          access_token: 'token123',
        }
        const response: api.PublishJobResponse = {
          success: true,
          platform_id: 'yt-123',
          platform_url: 'https://youtube.com/watch?v=yt-123',
        }
        mockFetch.mockResolvedValueOnce(mockResponse(response))

        const result = await api.publishJob('123', request)

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/jobs/123/publish', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(request),
        })
        expect(result).toEqual(response)
      })

      it('should throw error with message on failure', async () => {
        mockFetch.mockResolvedValueOnce({
          ok: false,
          text: () => Promise.resolve('Publishing failed: Invalid token'),
        })

        await expect(api.publishJob('123', {
          platform: 'youtube',
          title: 'Test',
          access_token: 'invalid',
        })).rejects.toThrow('Publishing failed: Invalid token')
      })
    })
  })

  describe('Streams API', () => {
    describe('fetchStreams', () => {
      it('should fetch all streams', async () => {
        const streams: api.Stream[] = [
          {
            id: 'stream-1',
            stream_key: 'key123',
            is_live: true,
            title: 'Live Stream',
            description: null,
            ingest_url: 'rtmp://ingest.example.com',
            playback_url: 'https://play.example.com/stream-1',
            current_viewers: 100,
            created_at: '2024-01-01',
          },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(streams))

        const result = await api.fetchStreams()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/streams')
        expect(result).toEqual(streams)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchStreams()).rejects.toThrow('Failed to fetch streams')
      })
    })

    describe('fetchStream', () => {
      it('should fetch a single stream', async () => {
        const stream: api.Stream = {
          id: 'stream-1',
          stream_key: 'key123',
          is_live: false,
          title: 'Test Stream',
          description: 'Test',
          ingest_url: null,
          playback_url: null,
          current_viewers: 0,
          created_at: '2024-01-01',
        }
        mockFetch.mockResolvedValueOnce(mockResponse(stream))

        const result = await api.fetchStream('stream-1')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/streams/stream-1')
        expect(result).toEqual(stream)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchStream('stream-1')).rejects.toThrow('Failed to fetch stream')
      })
    })

    describe('createStream', () => {
      it('should create a stream', async () => {
        const stream: api.Stream = {
          id: 'new-stream',
          stream_key: 'key456',
          is_live: false,
          title: 'New Stream',
          description: 'Description',
          ingest_url: null,
          playback_url: null,
          current_viewers: 0,
          created_at: '2024-01-01',
        }
        mockFetch.mockResolvedValueOnce(mockResponse(stream))

        const result = await api.createStream('New Stream', 'Description')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/streams', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ title: 'New Stream', description: 'Description' }),
        })
        expect(result).toEqual(stream)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.createStream('Test', 'Desc')).rejects.toThrow('Failed to create stream')
      })
    })

    describe('fetchStreamDestinations', () => {
      it('should fetch stream destinations', async () => {
        const destinations: api.RestreamDestination[] = [
          { plugin_id: 'publisher-twitch', access_token: 'token', enabled: true },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(destinations))

        const result = await api.fetchStreamDestinations('stream-1')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/streams/stream-1/destinations')
        expect(result).toEqual(destinations)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchStreamDestinations('stream-1')).rejects.toThrow('Failed to fetch stream destinations')
      })
    })

    describe('updateStreamDestinations', () => {
      it('should update stream destinations', async () => {
        const destinations: api.RestreamDestination[] = [
          { plugin_id: 'publisher-youtube', access_token: 'token', enabled: false },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.updateStreamDestinations('stream-1', destinations)

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/streams/stream-1/destinations', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(destinations),
        })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.updateStreamDestinations('stream-1', [])).rejects.toThrow('Failed to update stream destinations')
      })
    })
  })

  describe('Workers API', () => {
    describe('fetchWorkers', () => {
      it('should fetch all workers', async () => {
        const workers: api.Worker[] = [
          {
            id: 'worker-1',
            is_healthy: true,
            current_task_id: 'task-1',
            last_heartbeat: '2024-01-01T00:00:00Z',
            capabilities: { gpu: true },
          },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(workers))

        const result = await api.fetchWorkers()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/workers')
        expect(result).toEqual(workers)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchWorkers()).rejects.toThrow('Failed to fetch workers')
      })
    })
  })

  describe('Dashboard Stats', () => {
    describe('fetchDashboardStats', () => {
      it('should aggregate stats from multiple endpoints', async () => {
        const jobs: api.Job[] = [
          { id: '1', status: 'processing', source_url: '', profiles: [], created_at: '', updated_at: '', progress_pct: null, error_message: null, started_at: null, finished_at: null, eta_seconds: null },
          { id: '2', status: 'queued', source_url: '', profiles: [], created_at: '', updated_at: '', progress_pct: null, error_message: null, started_at: null, finished_at: null, eta_seconds: null },
          { id: '3', status: 'completed', source_url: '', profiles: [], created_at: '', updated_at: '', progress_pct: null, error_message: null, started_at: null, finished_at: null, eta_seconds: null },
          { id: '4', status: 'stitching', source_url: '', profiles: [], created_at: '', updated_at: '', progress_pct: null, error_message: null, started_at: null, finished_at: null, eta_seconds: null },
        ]
        const workers: api.Worker[] = [
          { id: 'w1', is_healthy: true, current_task_id: null, last_heartbeat: '', capabilities: null },
          { id: 'w2', is_healthy: false, current_task_id: null, last_heartbeat: '', capabilities: null },
        ]
        const streams: api.Stream[] = [
          { id: 's1', is_live: true, stream_key: '', title: null, description: null, ingest_url: null, playback_url: null, current_viewers: 0, created_at: '' },
          { id: 's2', is_live: false, stream_key: '', title: null, description: null, ingest_url: null, playback_url: null, current_viewers: 0, created_at: '' },
        ]

        mockFetch
          .mockResolvedValueOnce(mockResponse(jobs))
          .mockResolvedValueOnce(mockResponse(workers))
          .mockResolvedValueOnce(mockResponse(streams))

        const result = await api.fetchDashboardStats()

        expect(result).toEqual({
          activeJobs: 3, // processing + queued + stitching
          completedJobs: 1,
          workersOnline: 1,
          liveStreams: 1,
        })
      })

      it('should return zeros when all requests fail', async () => {
        mockFetch
          .mockRejectedValueOnce(new Error('Network error'))
          .mockRejectedValueOnce(new Error('Network error'))
          .mockRejectedValueOnce(new Error('Network error'))

        const result = await api.fetchDashboardStats()

        expect(result).toEqual({
          activeJobs: 0,
          completedJobs: 0,
          workersOnline: 0,
          liveStreams: 0,
        })
      })

      it('should handle partial failures gracefully', async () => {
        const jobs: api.Job[] = [
          { id: '1', status: 'processing', source_url: '', profiles: [], created_at: '', updated_at: '', progress_pct: null, error_message: null, started_at: null, finished_at: null, eta_seconds: null },
        ]
        mockFetch
          .mockResolvedValueOnce(mockResponse(jobs))
          .mockRejectedValueOnce(new Error('Workers error'))
          .mockRejectedValueOnce(new Error('Streams error'))

        const result = await api.fetchDashboardStats()

        expect(result).toEqual({
          activeJobs: 1,
          completedJobs: 0,
          workersOnline: 0,
          liveStreams: 0,
        })
      })
    })
  })

  describe('Restreams API', () => {
    describe('fetchRestreams', () => {
      it('should fetch all restreams', async () => {
        const restreams: api.RestreamJob[] = [
          {
            id: 'rs-1',
            title: 'Test Restream',
            description: null,
            input_type: 'rtmp',
            input_url: 'rtmp://input.example.com',
            output_destinations: [],
            status: 'running',
            created_at: '2024-01-01',
          },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(restreams))

        const result = await api.fetchRestreams()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/restreams')
        expect(result).toEqual(restreams)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchRestreams()).rejects.toThrow('Failed to fetch restreams')
      })
    })

    describe('fetchRestream', () => {
      it('should fetch a single restream', async () => {
        const restream: api.RestreamJob = {
          id: 'rs-1',
          title: 'Test',
          description: 'Desc',
          input_type: 'srt',
          input_url: 'srt://input.example.com',
          output_destinations: [{ platform: 'twitch', url: 'rtmp://twitch', enabled: true }],
          status: 'stopped',
          created_at: '2024-01-01',
        }
        mockFetch.mockResolvedValueOnce(mockResponse(restream))

        const result = await api.fetchRestream('rs-1')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/restreams/rs-1')
        expect(result).toEqual(restream)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchRestream('rs-1')).rejects.toThrow('Failed to fetch restream')
      })
    })

    describe('createRestream', () => {
      it('should create a restream', async () => {
        const data = {
          title: 'New Restream',
          input_type: 'rtmp',
          input_url: 'rtmp://input',
          output_destinations: [{ platform: 'youtube', url: 'rtmp://youtube', enabled: true }],
        }
        const restream: api.RestreamJob = { ...data, id: 'rs-new', description: null, status: 'created', created_at: '2024-01-01' }
        mockFetch.mockResolvedValueOnce(mockResponse(restream))

        const result = await api.createRestream(data)

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/restreams', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(data),
        })
        expect(result).toEqual(restream)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.createRestream({
          title: 'Test',
          input_type: 'rtmp',
          input_url: 'rtmp://test',
          output_destinations: [],
        })).rejects.toThrow('Failed to create restream')
      })
    })

    describe('startRestream', () => {
      it('should start a restream', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.startRestream('rs-1')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/restreams/rs-1/start', { method: 'POST' })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.startRestream('rs-1')).rejects.toThrow('Failed to start restream')
      })
    })

    describe('stopRestream', () => {
      it('should stop a restream', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.stopRestream('rs-1')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/restreams/rs-1/stop', { method: 'POST' })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.stopRestream('rs-1')).rejects.toThrow('Failed to stop restream')
      })
    })
  })

  describe('Profiles API', () => {
    describe('fetchProfiles', () => {
      it('should fetch all profiles', async () => {
        const profiles: api.Profile[] = [
          {
            id: 'profile-1',
            name: '720p',
            video_codec: 'h264',
            audio_codec: 'aac',
            width: 1280,
            height: 720,
            bitrate_kbps: 4000,
            is_system: true,
          },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(profiles))

        const result = await api.fetchProfiles()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/profiles')
        expect(result).toEqual(profiles)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchProfiles()).rejects.toThrow('Failed to fetch profiles')
      })
    })

    describe('fetchProfile', () => {
      it('should fetch a single profile', async () => {
        const profile: api.Profile = {
          id: 'profile-1',
          name: '1080p',
          video_codec: 'h264',
          is_system: false,
        }
        mockFetch.mockResolvedValueOnce(mockResponse(profile))

        const result = await api.fetchProfile('profile-1')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/profiles/profile-1')
        expect(result).toEqual(profile)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchProfile('profile-1')).rejects.toThrow('Failed to fetch profile')
      })
    })

    describe('createProfile', () => {
      it('should create a profile', async () => {
        const data: Partial<api.Profile> = {
          name: 'Custom 4K',
          video_codec: 'h265',
          width: 3840,
          height: 2160,
        }
        const profile: api.Profile = { ...data, id: 'profile-new', is_system: false } as api.Profile
        mockFetch.mockResolvedValueOnce(mockResponse(profile))

        const result = await api.createProfile(data)

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/profiles', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(data),
        })
        expect(result).toEqual(profile)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.createProfile({ name: 'Test' })).rejects.toThrow('Failed to create profile')
      })
    })

    describe('updateProfile', () => {
      it('should update a profile', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.updateProfile('profile-1', { name: 'Updated' })

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/profiles/profile-1', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ name: 'Updated' }),
        })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.updateProfile('profile-1', {})).rejects.toThrow('Failed to update profile')
      })
    })

    describe('deleteProfile', () => {
      it('should delete a profile', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.deleteProfile('profile-1')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/profiles/profile-1', { method: 'DELETE' })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.deleteProfile('profile-1')).rejects.toThrow('Failed to delete profile')
      })
    })
  })

  describe('Plugins API', () => {
    describe('fetchPlugins', () => {
      it('should fetch all plugins', async () => {
        const plugins: api.Plugin[] = [
          {
            id: 'storage-s3',
            type: 'storage',
            is_enabled: true,
            priority: 1,
            health: 'healthy',
            version: '1.0.0',
          },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(plugins))

        const result = await api.fetchPlugins()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/plugins')
        expect(result).toEqual(plugins)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchPlugins()).rejects.toThrow('Failed to fetch plugins')
      })
    })

    describe('enablePlugin', () => {
      it('should enable a plugin', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.enablePlugin('storage-s3')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/plugins/storage-s3/enable', { method: 'POST' })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.enablePlugin('storage-s3')).rejects.toThrow('Failed to enable plugin')
      })
    })

    describe('disablePlugin', () => {
      it('should disable a plugin', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.disablePlugin('storage-s3')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/plugins/storage-s3/disable', { method: 'POST' })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.disablePlugin('storage-s3')).rejects.toThrow('Failed to disable plugin')
      })
    })

    describe('fetchPlugin', () => {
      it('should fetch a single plugin', async () => {
        const plugin: api.Plugin = {
          id: 'auth-oidc',
          type: 'auth',
          config: { provider: 'google' },
          is_enabled: true,
          priority: 1,
          health: 'healthy',
        }
        mockFetch.mockResolvedValueOnce(mockResponse(plugin))

        const result = await api.fetchPlugin('auth-oidc')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/plugins/auth-oidc')
        expect(result).toEqual(plugin)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchPlugin('auth-oidc')).rejects.toThrow('Failed to fetch plugin')
      })
    })

    describe('updatePluginConfig', () => {
      it('should update plugin config', async () => {
        const request: api.UpdatePluginRequest = {
          config: { bucket: 'new-bucket' },
          priority: 2,
        }
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.updatePluginConfig('storage-s3', request)

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/plugins/storage-s3', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(request),
        })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.updatePluginConfig('storage-s3', { config: {} })).rejects.toThrow('Failed to update plugin config')
      })
    })
  })

  describe('Audit API', () => {
    describe('fetchAuditLogs', () => {
      it('should fetch audit logs with default pagination', async () => {
        const logs: api.AuditLog[] = [
          {
            id: 'audit-1',
            user_id: 'user-1',
            action: 'job.create',
            resource_type: 'job',
            resource_id: 'job-1',
            details: { profiles: ['720p'] },
            ip_address: '192.168.1.1',
            created_at: '2024-01-01',
          },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(logs))

        const result = await api.fetchAuditLogs()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/audit?limit=50&offset=0')
        expect(result).toEqual(logs)
      })

      it('should fetch audit logs with custom pagination', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse([]))

        await api.fetchAuditLogs(100, 50)

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/audit?limit=100&offset=50')
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchAuditLogs()).rejects.toThrow('Failed to fetch audit logs')
      })
    })
  })

  describe('System API', () => {
    describe('fetchSystemHealth', () => {
      it('should fetch system health', async () => {
        const health: api.SystemHealth = {
          status: 'healthy',
          version: '1.0.0',
          services: { database: 'healthy', storage: 'healthy' },
        }
        mockFetch.mockResolvedValueOnce(mockResponse(health))

        const result = await api.fetchSystemHealth()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/system/health')
        expect(result).toEqual(health)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchSystemHealth()).rejects.toThrow('Failed to fetch system health')
      })
    })

    describe('fetchSystemStats', () => {
      it('should fetch system stats', async () => {
        const stats: api.SystemStats = {
          jobs: { total: 100, queued: 5, processing: 3, completed: 90, failed: 2 },
          workers: { total: 4, healthy: 3 },
          streams: { total: 10, live: 2 },
          system: { go_version: 'go1.21', num_goroutine: 50, num_cpu: 8 },
        }
        mockFetch.mockResolvedValueOnce(mockResponse(stats))

        const result = await api.fetchSystemStats()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/system/stats')
        expect(result).toEqual(stats)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchSystemStats()).rejects.toThrow('Failed to fetch system stats')
      })
    })
  })

  describe('Webhooks API', () => {
    describe('fetchWebhooks', () => {
      it('should fetch all webhooks', async () => {
        const webhooks: api.Webhook[] = [
          {
            id: 'wh-1',
            url: 'https://example.com/webhook',
            events: ['job.completed', 'job.failed'],
            is_active: true,
            failure_count: 0,
            last_triggered_at: '2024-01-01',
            created_at: '2024-01-01',
          },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(webhooks))

        const result = await api.fetchWebhooks()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/webhooks')
        expect(result).toEqual(webhooks)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchWebhooks()).rejects.toThrow('Failed to fetch webhooks')
      })
    })

    describe('createWebhook', () => {
      it('should create a webhook', async () => {
        const webhook: api.Webhook = {
          id: 'wh-new',
          url: 'https://example.com/new',
          events: ['job.completed'],
          is_active: true,
          failure_count: 0,
          last_triggered_at: null,
          created_at: '2024-01-01',
        }
        mockFetch.mockResolvedValueOnce(mockResponse(webhook))

        const result = await api.createWebhook(
          'https://example.com/new',
          'secret123',
          ['job.completed']
        )

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/webhooks', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            url: 'https://example.com/new',
            secret: 'secret123',
            events: ['job.completed'],
          }),
        })
        expect(result).toEqual(webhook)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.createWebhook('url', 'secret', [])).rejects.toThrow('Failed to create webhook')
      })
    })

    describe('deleteWebhook', () => {
      it('should delete a webhook', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.deleteWebhook('wh-1')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/webhooks/wh-1', { method: 'DELETE' })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.deleteWebhook('wh-1')).rejects.toThrow('Failed to delete webhook')
      })
    })
  })

  describe('File Browser API', () => {
    describe('fetchStorageCapabilities', () => {
      it('should fetch storage capabilities', async () => {
        const capabilities: api.StoragePluginInfo[] = [
          { plugin_id: 'storage-s3', storage_type: 's3', supports_browse: true },
          { plugin_id: 'storage-local', storage_type: 'local', supports_browse: true },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(capabilities))

        const result = await api.fetchStorageCapabilities()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/files/capabilities')
        expect(result).toEqual(capabilities)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchStorageCapabilities()).rejects.toThrow('Failed to fetch storage capabilities')
      })
    })

    describe('fetchBrowseRoots', () => {
      it('should fetch browse roots', async () => {
        const roots: api.BrowseRoot[] = [
          { plugin_id: 'storage-s3', storage_type: 's3', name: 'S3 Bucket', path: '/' },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(roots))

        const result = await api.fetchBrowseRoots()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/files/roots')
        expect(result).toEqual(roots)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchBrowseRoots()).rejects.toThrow('Failed to fetch browse roots')
      })
    })

    describe('browseFiles', () => {
      it('should browse files with default options', async () => {
        const response: api.BrowseResponse = {
          plugin_id: 'storage-s3',
          current_path: '/',
          parent_path: '',
          entries: [],
        }
        mockFetch.mockResolvedValueOnce(mockResponse(response))

        const result = await api.browseFiles()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/files/browse?')
        expect(result).toEqual(response)
      })

      it('should browse files with all options', async () => {
        const response: api.BrowseResponse = {
          plugin_id: 'storage-s3',
          current_path: '/videos',
          parent_path: '/',
          entries: [
            {
              name: 'video.mp4',
              path: '/videos/video.mp4',
              is_directory: false,
              size: 1024000,
              mod_time: 1704067200,
              extension: '.mp4',
              is_video: true,
              is_audio: false,
              is_image: false,
            },
          ],
        }
        mockFetch.mockResolvedValueOnce(mockResponse(response))

        const result = await api.browseFiles({
          plugin: 'storage-s3',
          path: '/videos',
          mediaOnly: true,
          showHidden: true,
          search: 'test',
        })

        expect(mockFetch).toHaveBeenCalledWith(
          '/api/v1/files/browse?plugin=storage-s3&path=%2Fvideos&media_only=true&show_hidden=true&search=test'
        )
        expect(result).toEqual(response)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.browseFiles()).rejects.toThrow('Failed to browse files')
      })
    })
  })

  describe('File Upload API', () => {
    describe('getUploadUrl', () => {
      it('should get a pre-signed upload URL', async () => {
        const request: api.UploadUrlRequest = {
          object_key: 'uploads/video.mp4',
          content_type: 'video/mp4',
        }
        const response: api.UploadUrlResponse = {
          url: 'https://s3.example.com/presigned-url',
          expires_at: 1704070800,
          object_key: 'uploads/video.mp4',
          plugin_id: 'storage-s3',
        }
        mockFetch.mockResolvedValueOnce(mockResponse(response))

        const result = await api.getUploadUrl(request)

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/files/upload-url', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(request),
        })
        expect(result).toEqual(response)
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.getUploadUrl({ object_key: 'test' })).rejects.toThrow('Failed to get upload URL')
      })
    })

    describe('uploadFile', () => {
      it('should upload a file successfully', async () => {
        const file = new File(['test content'], 'test.mp4', { type: 'video/mp4' })
        const uploadResponse: api.UploadResponse = {
          url: 'https://storage.example.com/test.mp4',
          plugin_id: 'storage-s3',
          object_key: 'test.mp4',
          size: 12,
        }

        const uploadPromise = api.uploadFile(file)

        // Get the XHR instance that was created
        const xhr = mockXhrInstances[0]
        expect(xhr.open).toHaveBeenCalledWith('POST', '/api/v1/files/upload')

        // Simulate successful upload
        xhr.status = 200
        xhr.responseText = JSON.stringify(uploadResponse)
        xhr.triggerEvent('load')

        const result = await uploadPromise
        expect(result).toEqual(uploadResponse)
      })

      it('should handle upload progress callback', async () => {
        const file = new File(['test content'], 'test.mp4', { type: 'video/mp4' })
        const progressCallback = vi.fn()

        const uploadPromise = api.uploadFile(file, { onProgress: progressCallback })

        const xhr = mockXhrInstances[0]

        // Simulate progress event
        xhr.triggerUploadEvent('progress', { lengthComputable: true, loaded: 50, total: 100 })

        expect(progressCallback).toHaveBeenCalledWith({
          loaded: 50,
          total: 100,
          percentage: 50,
        })

        // Complete the upload
        xhr.status = 200
        xhr.responseText = JSON.stringify({ url: 'test', plugin_id: 's3', object_key: 'test', size: 12 })
        xhr.triggerEvent('load')

        await uploadPromise
      })

      it('should handle upload with options', async () => {
        const file = new File(['test'], 'test.mp4', { type: 'video/mp4' })

        const uploadPromise = api.uploadFile(file, {
          pluginId: 'storage-s3',
          bucket: 'my-bucket',
          objectKey: 'custom/path/test.mp4',
        })

        const xhr = mockXhrInstances[0]
        expect(xhr.send).toHaveBeenCalled()

        // Check that FormData was used (we can't easily check its contents)
        const sentData = xhr.send.mock.calls[0][0]
        expect(sentData).toBeInstanceOf(FormData)

        xhr.status = 200
        xhr.responseText = JSON.stringify({ url: 'test', plugin_id: 's3', object_key: 'test', size: 4 })
        xhr.triggerEvent('load')

        await uploadPromise
      })

      it('should handle upload failure with JSON error', async () => {
        const file = new File(['test'], 'test.mp4', { type: 'video/mp4' })

        const uploadPromise = api.uploadFile(file)

        const xhr = mockXhrInstances[0]
        xhr.status = 400
        xhr.responseText = JSON.stringify({ error: { message: 'File too large' } })
        xhr.triggerEvent('load')

        await expect(uploadPromise).rejects.toThrow('File too large')
      })

      it('should handle upload failure with plain text error', async () => {
        const file = new File(['test'], 'test.mp4', { type: 'video/mp4' })

        const uploadPromise = api.uploadFile(file)

        const xhr = mockXhrInstances[0]
        xhr.status = 500
        xhr.responseText = 'Internal server error\n'
        xhr.triggerEvent('load')

        await expect(uploadPromise).rejects.toThrow('Internal server error')
      })

      it('should handle network error', async () => {
        const file = new File(['test'], 'test.mp4', { type: 'video/mp4' })

        const uploadPromise = api.uploadFile(file)

        const xhr = mockXhrInstances[0]
        xhr.triggerEvent('error')

        await expect(uploadPromise).rejects.toThrow('Network error during upload')
      })

      it('should handle upload abort', async () => {
        const file = new File(['test'], 'test.mp4', { type: 'video/mp4' })

        const uploadPromise = api.uploadFile(file)

        const xhr = mockXhrInstances[0]
        xhr.triggerEvent('abort')

        await expect(uploadPromise).rejects.toThrow('Upload cancelled')
      })

      it('should handle invalid JSON response', async () => {
        const file = new File(['test'], 'test.mp4', { type: 'video/mp4' })

        const uploadPromise = api.uploadFile(file)

        const xhr = mockXhrInstances[0]
        xhr.status = 200
        xhr.responseText = 'not valid json'
        xhr.triggerEvent('load')

        await expect(uploadPromise).rejects.toThrow('Invalid response from server')
      })
    })

    describe('uploadToSignedUrl', () => {
      it('should upload to a signed URL', async () => {
        const file = new File(['test'], 'test.mp4', { type: 'video/mp4' })

        const uploadPromise = api.uploadToSignedUrl(file, 'https://s3.example.com/signed-url')

        const xhr = mockXhrInstances[0]
        expect(xhr.open).toHaveBeenCalledWith('PUT', 'https://s3.example.com/signed-url')
        expect(xhr.setRequestHeader).toHaveBeenCalledWith('Content-Type', 'video/mp4')

        xhr.status = 200
        xhr.triggerEvent('load')

        await uploadPromise
      })

      it('should set custom headers', async () => {
        const file = new File(['test'], 'test.mp4', { type: 'video/mp4' })

        const uploadPromise = api.uploadToSignedUrl(
          file,
          'https://s3.example.com/signed-url',
          { 'x-amz-acl': 'public-read', 'x-custom': 'value' }
        )

        const xhr = mockXhrInstances[0]
        expect(xhr.setRequestHeader).toHaveBeenCalledWith('x-amz-acl', 'public-read')
        expect(xhr.setRequestHeader).toHaveBeenCalledWith('x-custom', 'value')

        xhr.status = 200
        xhr.triggerEvent('load')

        await uploadPromise
      })

      it('should handle progress callback', async () => {
        const file = new File(['test'], 'test.mp4', { type: 'video/mp4' })
        const progressCallback = vi.fn()

        const uploadPromise = api.uploadToSignedUrl(
          file,
          'https://s3.example.com/signed-url',
          undefined,
          progressCallback
        )

        const xhr = mockXhrInstances[0]
        xhr.triggerUploadEvent('progress', { lengthComputable: true, loaded: 75, total: 100 })

        expect(progressCallback).toHaveBeenCalledWith({
          loaded: 75,
          total: 100,
          percentage: 75,
        })

        xhr.status = 200
        xhr.triggerEvent('load')

        await uploadPromise
      })

      it('should handle upload failure', async () => {
        const file = new File(['test'], 'test.mp4', { type: 'video/mp4' })

        const uploadPromise = api.uploadToSignedUrl(file, 'https://s3.example.com/signed-url')

        const xhr = mockXhrInstances[0]
        xhr.status = 403
        xhr.statusText = 'Forbidden'
        xhr.triggerEvent('load')

        await expect(uploadPromise).rejects.toThrow('Upload failed: Forbidden')
      })

      it('should handle network error', async () => {
        const file = new File(['test'], 'test.mp4', { type: 'video/mp4' })

        const uploadPromise = api.uploadToSignedUrl(file, 'https://s3.example.com/signed-url')

        const xhr = mockXhrInstances[0]
        xhr.triggerEvent('error')

        await expect(uploadPromise).rejects.toThrow('Upload failed')
      })

      it('should use application/octet-stream for files without type', async () => {
        const file = new File(['test'], 'test.bin')
        // Override the type getter to return empty string
        Object.defineProperty(file, 'type', { value: '' })

        const uploadPromise = api.uploadToSignedUrl(file, 'https://s3.example.com/signed-url')

        const xhr = mockXhrInstances[0]
        expect(xhr.setRequestHeader).toHaveBeenCalledWith('Content-Type', 'application/octet-stream')

        xhr.status = 200
        xhr.triggerEvent('load')

        await uploadPromise
      })
    })

    describe('reportError', () => {
      it('should report an error to the backend', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.reportError('Test error', 'test-component', 'stack trace', { extra: 'data' })

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/errors', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: expect.any(String),
        })

        const body = JSON.parse(mockFetch.mock.calls[0][1].body)
        expect(body.source).toBe('frontend')
        expect(body.severity).toBe('error')
        expect(body.message).toBe('Test error')
        expect(body.stack_trace).toBe('stack trace')
        expect(body.context_data.extra).toBe('data')
        expect(body.context_data.source_module).toBe('test-component')
      })

      it('should not throw when reporting fails', async () => {
        const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})
        mockFetch.mockRejectedValueOnce(new Error('Network error'))

        // Should not throw
        await api.reportError('Test error', 'test')

        expect(consoleSpy).toHaveBeenCalledWith('Failed to report error to backend', expect.any(Error))
        consoleSpy.mockRestore()
      })
    })
  })

  describe('SSE Subscriptions', () => {
    describe('subscribeToJobUpdates', () => {
      it('should create an EventSource for job updates', () => {
        const callback = vi.fn()

        const es = api.subscribeToJobUpdates(callback)

        expect(es).toBeInstanceOf(MockEventSource)
        expect((es as unknown as MockEventSource).url).toBe('/api/v1/events/jobs')
      })

      it('should parse and forward job update events', () => {
        const callback = vi.fn()

        const es = api.subscribeToJobUpdates(callback) as unknown as MockEventSource
        const eventData = { id: 'job-1', status: 'completed' }
        es.onmessage?.({ data: JSON.stringify(eventData) })

        expect(callback).toHaveBeenCalledWith(eventData)
      })
    })

    describe('subscribeToJobDetail', () => {
      it('should create an EventSource for specific job', () => {
        const callback = vi.fn()

        const es = api.subscribeToJobDetail('job-123', callback)

        expect(es).toBeInstanceOf(MockEventSource)
        expect((es as unknown as MockEventSource).url).toBe('/api/v1/events/jobs/job-123')
      })

      it('should parse and forward job detail events', () => {
        const callback = vi.fn()

        const es = api.subscribeToJobDetail('job-123', callback) as unknown as MockEventSource
        const eventData = { progress: 50, eta: 120 }
        es.onmessage?.({ data: JSON.stringify(eventData) })

        expect(callback).toHaveBeenCalledWith(eventData)
      })
    })

    describe('subscribeToDashboard', () => {
      it('should create an EventSource for dashboard', () => {
        const callback = vi.fn()

        const es = api.subscribeToDashboard(callback)

        expect(es).toBeInstanceOf(MockEventSource)
        expect((es as unknown as MockEventSource).url).toBe('/api/v1/events/dashboard')
      })

      it('should parse and forward dashboard events', () => {
        const callback = vi.fn()

        const es = api.subscribeToDashboard(callback) as unknown as MockEventSource
        const eventData = { activeJobs: 5, workersOnline: 3 }
        es.onmessage?.({ data: JSON.stringify(eventData) })

        expect(callback).toHaveBeenCalledWith(eventData)
      })
    })
  })

  describe('Notifications API', () => {
    describe('fetchNotifications', () => {
      it('should fetch notifications with default pagination', async () => {
        const notifications: api.Notification[] = [
          {
            id: 'notif-1',
            user_id: 'user-1',
            title: 'Job Completed',
            message: 'Your job has finished processing',
            type: 'success',
            is_read: false,
            created_at: '2024-01-01',
          },
        ]
        mockFetch.mockResolvedValueOnce(mockResponse(notifications))

        const result = await api.fetchNotifications()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/notifications?limit=20&offset=0')
        expect(result).toEqual(notifications)
      })

      it('should fetch notifications with custom pagination', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse([]))

        await api.fetchNotifications(50, 10)

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/notifications?limit=50&offset=10')
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.fetchNotifications()).rejects.toThrow('Failed to fetch notifications')
      })
    })

    describe('markNotificationRead', () => {
      it('should mark a notification as read', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.markNotificationRead('notif-1')

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/notifications/notif-1/read', { method: 'PUT' })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.markNotificationRead('notif-1')).rejects.toThrow('Failed to mark notification as read')
      })
    })

    describe('clearAllNotifications', () => {
      it('should clear all notifications', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null))

        await api.clearAllNotifications()

        expect(mockFetch).toHaveBeenCalledWith('/api/v1/notifications/clear', { method: 'POST' })
      })

      it('should throw error on failure', async () => {
        mockFetch.mockResolvedValueOnce(mockResponse(null, false))

        await expect(api.clearAllNotifications()).rejects.toThrow('Failed to clear notifications')
      })
    })
  })
})
