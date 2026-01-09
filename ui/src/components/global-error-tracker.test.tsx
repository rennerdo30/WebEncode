import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, waitFor } from '@testing-library/react'
import { GlobalErrorTracker } from './global-error-tracker'

// Polyfill PromiseRejectionEvent for jsdom
class MockPromiseRejectionEvent extends Event {
  reason: unknown
  promise: Promise<unknown>

  constructor(type: string, init: { reason?: unknown; promise?: Promise<unknown> } = {}) {
    super(type)
    this.reason = init.reason
    this.promise = init.promise ?? Promise.resolve()
  }
}

if (typeof PromiseRejectionEvent === 'undefined') {
  (global as unknown as { PromiseRejectionEvent: typeof MockPromiseRejectionEvent }).PromiseRejectionEvent = MockPromiseRejectionEvent
}

describe('GlobalErrorTracker', () => {
  let mockFetch: ReturnType<typeof vi.fn>
  let originalFetch: typeof global.fetch
  let addEventListenerSpy: ReturnType<typeof vi.spyOn>
  let removeEventListenerSpy: ReturnType<typeof vi.spyOn>

  beforeEach(() => {
    originalFetch = global.fetch
    mockFetch = vi.fn().mockResolvedValue({ ok: true })
    global.fetch = mockFetch

    addEventListenerSpy = vi.spyOn(window, 'addEventListener')
    removeEventListenerSpy = vi.spyOn(window, 'removeEventListener')
  })

  afterEach(() => {
    global.fetch = originalFetch
    vi.restoreAllMocks()
  })

  describe('rendering', () => {
    it('should render null (no visual output)', () => {
      const { container } = render(<GlobalErrorTracker />)
      expect(container.firstChild).toBeNull()
    })
  })

  describe('event listeners', () => {
    it('should register error event listener on mount', () => {
      render(<GlobalErrorTracker />)

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        'error',
        expect.any(Function)
      )
    })

    it('should register unhandledrejection event listener on mount', () => {
      render(<GlobalErrorTracker />)

      expect(addEventListenerSpy).toHaveBeenCalledWith(
        'unhandledrejection',
        expect.any(Function)
      )
    })

    it('should register error listener with capture phase for resources', () => {
      render(<GlobalErrorTracker />)

      // Find the call with capture=true
      const captureCall = addEventListenerSpy.mock.calls.find(
        (call) => call[0] === 'error' && call[2] === true
      )
      expect(captureCall).toBeDefined()
    })

    it('should remove event listeners on unmount', () => {
      const { unmount } = render(<GlobalErrorTracker />)

      unmount()

      expect(removeEventListenerSpy).toHaveBeenCalledWith(
        'error',
        expect.any(Function)
      )
      expect(removeEventListenerSpy).toHaveBeenCalledWith(
        'unhandledrejection',
        expect.any(Function)
      )
    })
  })

  describe('error reporting', () => {
    it('should report JS errors to the backend', async () => {
      render(<GlobalErrorTracker />)

      // Simulate an error event
      const errorEvent = new ErrorEvent('error', {
        message: 'Test error message',
        filename: 'test.js',
        lineno: 10,
        colno: 5,
        error: new Error('Test error'),
      })

      window.dispatchEvent(errorEvent)

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledWith('/api/v1/errors', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: expect.stringContaining('Test error message'),
        })
      })
    })

    it('should report unhandled promise rejections', async () => {
      render(<GlobalErrorTracker />)

      // Simulate a promise rejection event
      const rejectionEvent = new PromiseRejectionEvent('unhandledrejection', {
        reason: new Error('Promise rejection error'),
        promise: Promise.reject(new Error('test')).catch(() => {}),
      })

      window.dispatchEvent(rejectionEvent)

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledWith('/api/v1/errors', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: expect.stringContaining('Promise rejection error'),
        })
      })
    })

    it('should handle string rejection reasons', async () => {
      render(<GlobalErrorTracker />)

      const rejectionEvent = new PromiseRejectionEvent('unhandledrejection', {
        reason: 'String rejection reason',
        promise: Promise.reject('test').catch(() => {}),
      })

      window.dispatchEvent(rejectionEvent)

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledWith('/api/v1/errors', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: expect.stringContaining('String rejection reason'),
        })
      })
    })

    it('should include context data in error reports', async () => {
      render(<GlobalErrorTracker />)

      const errorEvent = new ErrorEvent('error', {
        message: 'Context test error',
        filename: 'context-test.js',
        lineno: 20,
        colno: 10,
        error: new Error('Context test'),
      })

      window.dispatchEvent(errorEvent)

      await waitFor(() => {
        const fetchCall = mockFetch.mock.calls[0]
        const body = JSON.parse(fetchCall[1].body)
        expect(body.context_data).toHaveProperty('url')
        expect(body.context_data).toHaveProperty('userAgent')
        expect(body.context_data.filename).toBe('context-test.js')
        expect(body.context_data.lineno).toBe(20)
        expect(body.context_data.colno).toBe(10)
      })
    })

    it('should set source to frontend:js for JS errors', async () => {
      render(<GlobalErrorTracker />)

      const errorEvent = new ErrorEvent('error', {
        message: 'JS error',
        error: new Error('JS error'),
      })

      window.dispatchEvent(errorEvent)

      await waitFor(() => {
        const fetchCall = mockFetch.mock.calls[0]
        const body = JSON.parse(fetchCall[1].body)
        expect(body.source).toBe('frontend')
      })
    })

    it('should set severity to error by default', async () => {
      render(<GlobalErrorTracker />)

      const errorEvent = new ErrorEvent('error', {
        message: 'Severity test',
        error: new Error('Severity test'),
      })

      window.dispatchEvent(errorEvent)

      await waitFor(() => {
        const fetchCall = mockFetch.mock.calls[0]
        const body = JSON.parse(fetchCall[1].body)
        expect(body.severity).toBe('error')
      })
    })

    it('should handle fetch failures gracefully', async () => {
      const consoleWarnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {})
      mockFetch.mockRejectedValueOnce(new Error('Network error'))

      render(<GlobalErrorTracker />)

      const errorEvent = new ErrorEvent('error', {
        message: 'Fetch failure test',
        error: new Error('Fetch failure test'),
      })

      window.dispatchEvent(errorEvent)

      await waitFor(() => {
        expect(consoleWarnSpy).toHaveBeenCalledWith(
          'Failed to report error to backend',
          expect.any(Error)
        )
      })

      consoleWarnSpy.mockRestore()
    })
  })

  describe('error types', () => {
    it('should report errors with stack traces', async () => {
      render(<GlobalErrorTracker />)

      const error = new Error('Stack trace test')
      const errorEvent = new ErrorEvent('error', {
        message: error.message,
        error: error,
      })

      window.dispatchEvent(errorEvent)

      await waitFor(() => {
        const fetchCall = mockFetch.mock.calls[0]
        const body = JSON.parse(fetchCall[1].body)
        expect(body.stack_trace).toBeDefined()
      })
    })

    it('should handle promise rejections without Error objects', async () => {
      render(<GlobalErrorTracker />)

      const rejectionEvent = new PromiseRejectionEvent('unhandledrejection', {
        reason: { custom: 'object' },
        promise: Promise.reject({ custom: 'object' }).catch(() => {}),
      })

      window.dispatchEvent(rejectionEvent)

      await waitFor(() => {
        const fetchCall = mockFetch.mock.calls[0]
        const body = JSON.parse(fetchCall[1].body)
        expect(body.message).toBe('Unhandled Promise Rejection')
      })
    })
  })
})
